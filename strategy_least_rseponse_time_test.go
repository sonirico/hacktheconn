package hacktheconn

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLeastResponseTimeMovingAverageCalculator(t *testing.T) {
	tests := []struct {
		name       string
		windowSize int
		inputs     []time.Duration
		expected   time.Duration
	}{
		{
			name:       "moving average with 3 inputs",
			windowSize: 3,
			inputs:     []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 300 * time.Millisecond},
			expected:   200 * time.Millisecond,
		},
		{
			name:       "moving average with more inputs than window",
			windowSize: 3,
			inputs: []time.Duration{
				100 * time.Millisecond,
				200 * time.Millisecond,
				300 * time.Millisecond,
				400 * time.Millisecond,
			},
			expected: 300 * time.Millisecond, // Last 3 inputs: (200+300+400)/3
		},
		{
			name:       "moving average with single input",
			windowSize: 1,
			inputs:     []time.Duration{150 * time.Millisecond},
			expected:   150 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculator := LeastResponseTimeMovingAverageCalculator(tt.windowSize)
			var result time.Duration
			for _, input := range tt.inputs {
				result = calculator(input)
			}

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLeastResponseTimeWeightedAverageCalculator(t *testing.T) {
	tests := []struct {
		name     string
		weight   float64
		inputs   []time.Duration
		expected string
	}{
		{
			name:   "weighted average with multiple inputs",
			weight: 0.8,
			inputs: []time.Duration{
				100 * time.Millisecond, // Initial value
				200 * time.Millisecond, // New average: (0.8*200) + (0.2*100) = 180
				150 * time.Millisecond, // New average: (0.8*150) + (0.2*180) = 156
				50 * time.Millisecond,  // New average: (0.8*50) + (0.2*156) = 71.2
			},
			expected: "71.2ms",
		},
		{
			name:     "weighted average with single input",
			weight:   0.5,
			inputs:   []time.Duration{120 * time.Millisecond},
			expected: "120ms",
		},
		{
			name:   "high weight favors recent inputs",
			weight: 0.9,
			inputs: []time.Duration{
				100 * time.Millisecond, // Initial value
				200 * time.Millisecond, // New average: (0.9*200) + (0.1*100) = 190
				300 * time.Millisecond, // New average: (0.9*300) + (0.1*190) = 289
			},
			expected: "289ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculator := LeastResponseTimeWeightedAverageCalculator(tt.weight)
			var result time.Duration
			for _, input := range tt.inputs {
				result = calculator(input)
			}

			if result.String() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// mockLeastResponseTimeTransport simulates a RoundTripper with configurable delay.
type mockLeastResponseTimeTransport struct {
	delay time.Duration
	ID    int
}

func (m *mockLeastResponseTimeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	time.Sleep(m.delay)
	return &http.Response{StatusCode: 200}, nil
}

func TestLeastResponseTimeStrategy(t *testing.T) {
	tests := []struct {
		name       string
		transports []http.RoundTripper
		calculator ResponseTimeCalculator
		mockDelays []time.Duration
		expectedID int
	}{
		{
			name: "select fastest transport with least response time",
			transports: []http.RoundTripper{
				&mockLeastResponseTimeTransport{ID: 1, delay: 100 * time.Millisecond},
				&mockLeastResponseTimeTransport{ID: 2, delay: 50 * time.Millisecond},
				&mockLeastResponseTimeTransport{ID: 3, delay: 150 * time.Millisecond},
			},
			calculator: LeastResponseTimeLastResponseTimeCalculator,
			mockDelays: []time.Duration{
				100 * time.Millisecond,
				50 * time.Millisecond,
				150 * time.Millisecond,
			},
			expectedID: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strat := NewLeastResponseTimeStrategy(tt.transports, time.Now, tt.calculator)

			for _, transport := range strat.transports {
				// Simulate a request to calculate response times
				req, _ := http.NewRequest("GET", "https://example.com", nil)
				_, err := transport.RoundTrip(req)
				if err != nil {
					t.Fatalf("unexpected error during RoundTrip: %v", err)
				}
			}

			// Acquire the best transport
			selectedTransport, err := strat.Acquire()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tp, ok := selectedTransport.(*leastResponseTimeRoundTripper)
			if !ok {
				t.Fatalf("unexpected type for transport")
			}

			mockTp, ok := tp.Unwrap().(*mockLeastResponseTimeTransport)
			if !ok {
				t.Fatalf("unexpected type for transport")
			}

			assert.Equal(t, tt.expectedID, mockTp.ID)
		})
	}
}
