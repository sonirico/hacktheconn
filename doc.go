/*
package hacktheconn provides HTTP client strategies for dynamic transport selection based on various criteria.
It includes configurable strategies for efficient connection pooling and request routing, tailored for
real-time, high-throughput applications.

# Overview

The package introduces pluggable transport selection strategies to dynamically choose the most suitable
HTTP transport based on custom metrics, such as response time or load. It also supports advanced proxy
handling for distributed systems.

# Features

- **Round-Robin Strategy**: Distributes requests evenly across all available transports.
- **Fill Holes Strategy**: Selects the transport with the fewest concurrent requests.
- **Least Response Time Strategy**: Dynamically picks the transport with the lowest response time using:
  - Last Response Time
  - Moving Average
  - Weighted Average

- **Customizable Strategies**: Extendable with user-defined selection algorithms.
- **Proxy Support**: Fully compatible with both HTTP and SOCKS5 proxies.

# Strategies

## Round-Robin Strategy

Distributes requests evenly across all available transports in a cyclic order.

Example:

	transports := []http.RoundTripper{
		&http.Transport{},
		&http.Transport{},
	}
	strategy := NewRoundRobinStrategy(transports)

	client := &http.Client{
		Transport: Transport(strategy),
	}

	resp, err := client.Get("https://example.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

## Fill Holes Strategy

Chooses the transport with the fewest concurrent requests to ensure fair resource utilization.

Example:

	transports := []http.RoundTripper{
		&http.Transport{},
		&http.Transport{},
	}
	strategy := NewFillHolesStrategy(transports)

	client := &http.Client{
		Transport: Transport(strategy),
	}

	resp, err := client.Get("https://example.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

## Least Response Time Strategy

Selects the transport with the lowest response time. Response time calculation can be configured using
predefined calculators or custom ones.

### Predefined Calculators:
- `LeastResponseTimeLastResponseTimeCalculator`: Uses the most recent response time.
- `LeastResponseTimeMovingAverageCalculator(windowSize int)`: Averages the response times of the last `windowSize` requests.
- `LeastResponseTimeWeightedAverageCalculator(weight float64)`: Applies a weighted average with higher importance to recent response times.

Example with Weighted Average:

	transports := []http.RoundTripper{
		&http.Transport{},
		&http.Transport{},
	}

	strategy := NewLeastResponseTimeStrategy(
		transports,
		time.Now,
		LeastResponseTimeWeightedAverageCalculator(0.8),
	)

	client := &http.Client{
		Transport: Transport(strategy),
	}

	resp, err := client.Get("https://example.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

## Custom Strategies

To define a custom strategy, implement the `strategy` interface:

	type strategy interface {
		Acquire() (http.RoundTripper, error)
		Release(http.RoundTripper)
	}

Example:

	type CustomStrategy struct { ... }

	func (cs *CustomStrategy) Acquire() (http.RoundTripper, error) {
		// Custom logic
	}

	func (cs *CustomStrategy) Release(rt http.RoundTripper) {
		// Custom cleanup logic
	}

# Contributing

Contributions are welcome! Please ensure your code is documented and tested.

1. Fork the repository.
2. Create a feature branch.
3. Submit a pull request.

# License

This package is licensed under the MIT License.
*/
package hacktheconn
