package hacktheconn

import (
	"net/http"
	"testing"
)

func TestFillHolesStrategy(t *testing.T) {
	transports := []http.RoundTripper{
		&MockTransport{ID: "A"},
		&MockTransport{ID: "B"},
		&MockTransport{ID: "C"},
	}

	s := NewFillHolesStrategy(transports)

	// Simulate request selection with uneven usage
	_, _ = s.Acquire() // A: 1, B: 0, C: 0
	_, _ = s.Acquire() // A: 1, B: 1, C: 0
	_, _ = s.Acquire() // A: 1, B: 1, C: 1
	_, _ = s.Acquire() // A: 2, B: 1, C: 1

	transport, err := s.Acquire()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mockTransport, ok := transport.(*MockTransport)
	if !ok {
		t.Fatalf("unexpected type for transport")
	}

	if mockTransport.ID != "B" {
		t.Errorf("expected transport A to be acquired, got %s", mockTransport.ID)
	}

	// Verify counts after selection
	expectedCounts := []int{2, 2, 1}
	for i, count := range s.requestCounts {
		if count != expectedCounts[i] {
			t.Errorf("expected request count for transport %s to be %d, got %d",
				transports[i].(*MockTransport).ID, expectedCounts[i], count)
		}
	}
}
