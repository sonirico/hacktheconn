package main

import (
	"fmt"
	"net/http"
	"sync"
)

// FillHolesStrategy selects the transport with the least ongoing requests.
type FillHolesStrategy struct {
	transports    []http.RoundTripper
	requestCounts []int
	mutex         sync.Mutex
}

// NewFillHolesStrategy initializes the fill-holes strategy.
func NewFillHolesStrategy(transports []http.RoundTripper) *FillHolesStrategy {
	return &FillHolesStrategy{
		transports:    transports,
		requestCounts: make([]int, len(transports)),
	}
}

// Acquire picks the transport with the least ongoing requests.
func (fh *FillHolesStrategy) Acquire() (http.RoundTripper, error) {
	if len(fh.transports) == 0 {
		return nil, fmt.Errorf("no transports available")
	}

	fh.mutex.Lock()
	defer fh.mutex.Unlock()

	minIndex := 0
	minCount := fh.requestCounts[0]

	for i, count := range fh.requestCounts {
		if count < minCount {
			minIndex = i
			minCount = count
		}
	}

	fh.requestCounts[minIndex]++
	return fh.transports[minIndex], nil
}

// Release decrements the request count for a transport.
func (fh *FillHolesStrategy) Release(transport http.RoundTripper) {
	fh.mutex.Lock()
	defer fh.mutex.Unlock()

	for i, t := range fh.transports {
		if t == transport {
			fh.requestCounts[i]--
			if fh.requestCounts[i] < 0 {
				fh.requestCounts[i] = 0
			}
			break
		}
	}
}

type FillHolesConfig struct {
	Proxies          []string
	TransportFactory func(string) (*http.Transport, error)
}

type FillHolesOption func(*FillHolesConfig)

// TransportFillHoles creates a round-robin StrategyTransport with configurable options.
func TransportFillHoles(proxies []string, opts ...FillHolesOption) http.RoundTripper {
	cfg := &FillHolesConfig{
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

	return Transport(NewFillHolesStrategy(transports))
}

func FillHolesWithTransportFactory(factory func(string) (*http.Transport, error)) FillHolesOption {
	return func(cfg *FillHolesConfig) {
		cfg.TransportFactory = factory
	}
}
