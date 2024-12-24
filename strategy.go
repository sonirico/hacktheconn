package main

import (
	"net/http"
)

// strategy interface for roundtripper selection.
type strategy interface {
	Acquire() (http.RoundTripper, error)
	Release(http.RoundTripper)
}

// StrategyTransport wraps a strategy for dynamic transport selection.
type StrategyTransport struct {
	strategy strategy
}

// Transport creates a new StrategyTransport with the given strategy.
func Transport(strategy strategy) *StrategyTransport {
	return &StrategyTransport{
		strategy: strategy,
	}
}

// RoundTrip selects a transport dynamically and executes the request.
func (t *StrategyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport, err := t.strategy.Acquire()
	if err != nil {
		return nil, err
	}
	defer t.strategy.Release(transport)

	return transport.RoundTrip(req)
}
