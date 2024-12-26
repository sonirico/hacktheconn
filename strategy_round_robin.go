package main

import (
	"fmt"
	"net/http"
	"sync"
)

// RoundRobinStrategy manages round-robin selection.
type RoundRobinStrategy struct {
	transports   []http.RoundTripper
	lastSelected int
	mutex        sync.Mutex
}

// NewRoundRobinStrategy initializes the round-robin strategy.
func NewRoundRobinStrategy(transports []http.RoundTripper) *RoundRobinStrategy {
	return &RoundRobinStrategy{
		transports:   transports,
		lastSelected: -1,
	}
}

// Acquire picks the next transport in a round-robin manner.
func (rr *RoundRobinStrategy) Acquire() (http.RoundTripper, error) {
	if len(rr.transports) == 0 {
		return nil, ErrNoTransports
	}

	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	rr.lastSelected = (rr.lastSelected + 1) % len(rr.transports)
	return rr.transports[rr.lastSelected], nil
}

func (rr *RoundRobinStrategy) Release(http.RoundTripper) {}

type RoundRobinConfig struct {
	Proxies          []string
	TransportFactory func(string) (*http.Transport, error)
}

type RoundRobinOption func(*RoundRobinConfig)

// TransportRoundRobin creates a round-robin StrategyTransport with configurable options.
func TransportRoundRobin(proxies []string, opts ...RoundRobinOption) http.RoundTripper {
	cfg := &RoundRobinConfig{
		Proxies:          proxies,
		TransportFactory: DefaultTransportFactory,
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

	return Transport(NewRoundRobinStrategy(transports))
}

func RoundRobinWithTransportFactory(factory func(string) (*http.Transport, error)) RoundRobinOption {
	return func(cfg *RoundRobinConfig) {
		cfg.TransportFactory = factory
	}
}
