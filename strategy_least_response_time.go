package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sonirico/gozo/slices"
)

// ResponseTimeCalculator defines how response time is calculated.
type ResponseTimeCalculator func(lastRequestDuration time.Duration) time.Duration

// leastResponseTimeRoundTripper tracks and calculates response times for a transport.
type leastResponseTimeRoundTripper struct {
	roundTripper           http.RoundTripper
	clock                  func() time.Time
	responseTime           time.Duration
	responseTimeCalculator ResponseTimeCalculator
}

// RoundTrip measures the time taken for a request and updates the response time.
func (l *leastResponseTimeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := l.clock()

	// Execute the actual request
	res, err := l.roundTripper.RoundTrip(req)

	// Update response time using the calculator
	l.responseTime = l.responseTimeCalculator(time.Since(start))

	return res, err
}

func (l *leastResponseTimeRoundTripper) Unwrap() http.RoundTripper {
	return l.roundTripper
}

// LeastResponseTimeStrategy selects the transport with the least response time.
type LeastResponseTimeStrategy struct {
	transports []*leastResponseTimeRoundTripper
	mutex      sync.Mutex
}

// NewLeastResponseTimeStrategy initializes the least response time strategy.
func NewLeastResponseTimeStrategy(
	transports []http.RoundTripper,
	clock func() time.Time,
	calculator ResponseTimeCalculator,
) *LeastResponseTimeStrategy {
	return &LeastResponseTimeStrategy{
		transports: slices.Map(transports, func(rt http.RoundTripper) *leastResponseTimeRoundTripper {
			return &leastResponseTimeRoundTripper{
				roundTripper:           rt,
				clock:                  clock,
				responseTimeCalculator: calculator,
			}
		}),
	}
}

// Acquire picks the transport with the least response time.
func (lr *LeastResponseTimeStrategy) Acquire() (http.RoundTripper, error) {
	if len(lr.transports) == 0 {
		return nil, ErrNoTransports
	}

	lr.mutex.Lock()
	defer lr.mutex.Unlock()

	minIndex := 0
	minTime := lr.transports[0].responseTime

	for i, rt := range lr.transports {
		if rt.responseTime < minTime {
			minIndex = i
			minTime = rt.responseTime
		}
	}

	return lr.transports[minIndex], nil
}

// Release is a no-op for this strategy as response times are tracked automatically.
func (lr *LeastResponseTimeStrategy) Release(_ http.RoundTripper) {}

type (
	// OptLeastResponseTime configures the LeastResponseTime strategy.
	OptLeastResponseTime = Option[LeastResponseTimeConfig]

	LeastResponseTimeConfig struct {
		baseStrategyConfig

		clock                  func() time.Time
		responseTimeCalculator ResponseTimeCalculator
	}
)

// TransportLeastResponseTime creates a StrategyTransport with configurable options.
func TransportLeastResponseTime(proxies []string, opts ...OptLeastResponseTime) http.RoundTripper {
	cfg := &LeastResponseTimeConfig{
		baseStrategyConfig: baseStrategyConfig{
			Proxies:          proxies,
			TransportFactory: DefaultTransportFactory,
		},
		responseTimeCalculator: LeastResponseTimeWeightedAverageCalculator(0.75),
		clock:                  time.Now,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	var transports []http.RoundTripper
	for _, proxy := range cfg.Proxies {
		transport, err := cfg.TransportFactory(proxy)
		if err != nil {
			fmt.Printf("Error creating transport for proxy %s: %v\n", proxy, err)
			continue
		}
		transports = append(transports, transport)
	}

	return Transport(NewLeastResponseTimeStrategy(transports, cfg.clock, cfg.responseTimeCalculator))
}

// OptLeastResponseTimeWithClock configures a custom clock function.
func OptLeastResponseTimeWithClock(fn func() time.Time) OptLeastResponseTime {
	return func(cfg *LeastResponseTimeConfig) {
		cfg.clock = fn
	}
}

// OptLeastResponseTimeWithCalculator configures a custom response time calculator.
func OptLeastResponseTimeWithCalculator(calculator ResponseTimeCalculator) OptLeastResponseTime {
	return func(cfg *LeastResponseTimeConfig) {
		cfg.responseTimeCalculator = calculator
	}
}

// Predefined response time calculators.

// LeastResponseTimeLastResponseTimeCalculator uses the most recent response time.
func LeastResponseTimeLastResponseTimeCalculator(lastRequestDuration time.Duration) time.Duration {
	return lastRequestDuration
}

// LeastResponseTimeMovingAverageCalculator applies a moving average for response times.
// Moving Average
//
// A Moving Average calculates the average of the last n response times. This method gives equal weight to all recent measurements within the window size.
// Example:
//
// Let's say we have a window size of 3 and these response times:
//
//	Request 1: 100ms
//	Request 2: 150ms
//	Request 3: 200ms
//	Request 4: 120ms
//
// The moving average would work like this:
//
//	After Request 1: Average = 100ms (only one data point so far).
//	After Request 2: Average = (100 + 150) / 2 = 125ms.
//	After Request 3: Average = (100 + 150 + 200) / 3 = 150ms.
//	After Request 4: The oldest value (100ms) is dropped. Average = (150 + 200 + 120) / 3 = 156.67ms.
//
// This smooths out spikes but only considers the last n values.
func LeastResponseTimeMovingAverageCalculator(windowSize int) ResponseTimeCalculator {
	if windowSize <= 0 {
		panic("windowSize must be greater than 0")
	}

	buffer := make([]time.Duration, windowSize)
	index := 0
	count := 0
	sum := time.Duration(0)

	return func(lastRequestDuration time.Duration) time.Duration {
		sum -= buffer[index]

		buffer[index] = lastRequestDuration
		sum += lastRequestDuration

		index = (index + 1) % windowSize

		if count < windowSize {
			count++
		}

		return sum / time.Duration(count)
	}
}

// LeastResponseTimeWeightedAverageCalculator applies a weighted average for response times.
// A Weighted Average gives more importance to recent response times while still accounting for the historical average. The formula is:
//
// New Average=(Weight×Latest Duration)+((1−Weight)×Previous Average)New Average=(Weight×Latest Duration)+((1−Weight)×Previous Average)
//
//	Weight determines how much influence the latest response has. A higher weight (e.g., 0.9) makes the system more responsive to recent changes.
//	Historical response times decay over time. Weight should be between 0 and 1.
//
// Example:
//
// Assume weight = 0.8 and:
//
//	Initial response time = 100ms (this is the first value, so it's the average initially).
//	New response times = 200ms, 150ms, 50ms.
//
// Calculations:
//
//	After Request 1 (100ms): Average = 100ms.
//	After Request 2 (200ms):
//	    New Average = (0.8 × 200) + (0.2 × 100) = 160ms.
//	After Request 3 (150ms):
//	    New Average = (0.8 × 150) + (0.2 × 160) = 152ms.
//	After Request 4 (50ms):
//	    New Average = (0.8 × 50) + (0.2 × 152) = 70.4ms.
//
// This method adapts faster to recent changes but never completely discards the historical influence.
func LeastResponseTimeWeightedAverageCalculator(weight float64) ResponseTimeCalculator {
	if weight < 0 || weight > 1 {
		panic("weight must be between 0 and 1")
	}
	previousResponseTime := time.Duration(0)
	return func(lastRequestDuration time.Duration) time.Duration {
		defer func() {
			println(previousResponseTime.String())
		}()
		if previousResponseTime == 0 {
			previousResponseTime = lastRequestDuration
			return previousResponseTime
		}

		previousResponseTime = time.Duration(
			weight*float64(lastRequestDuration) + (1-weight)*float64(previousResponseTime),
		)
		return previousResponseTime
	}
}
