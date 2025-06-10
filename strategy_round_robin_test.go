package hacktheconn

import (
	"fmt"
	"net/http"
	"testing"
)

// mockLeastResponseTimeTransport is a mock http.RoundTripper for testing.
type MockTransport struct {
	ID string
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("mock transport %s called", m.ID)
}

func TestRoundRobinStrategy(t *testing.T) {
	transports := []http.RoundTripper{
		&MockTransport{ID: "A"},
		&MockTransport{ID: "B"},
		&MockTransport{ID: "C"},
	}

	strategy := NewRoundRobinStrategy(transports)

	tests := []struct {
		expectedID string
	}{
		{expectedID: "A"},
		{expectedID: "B"},
		{expectedID: "C"},
		{expectedID: "A"},
		{expectedID: "B"},
		{expectedID: "C"},
	}

	for i, test := range tests {
		transport, err := strategy.Acquire()
		if err != nil {
			t.Fatalf("unexpected error in test %d: %v", i, err)
		}

		mockTransport, ok := transport.(*MockTransport)
		if !ok {
			t.Fatalf("unexpected type for transport in test %d", i)
		}

		if mockTransport.ID != test.expectedID {
			t.Errorf("test %d: expected ID %s, got %s", i, test.expectedID, mockTransport.ID)
		}
	}
}
