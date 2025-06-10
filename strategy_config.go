package hacktheconn

import "net/http"

type baseStrategyConfig struct {
	Proxies          []string
	TransportFactory func(string) (*http.Transport, error)
}

type Option[T any] func(*T)
