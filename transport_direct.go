package hacktheconn

import (
	"net/http"
)

// DirectTransport creates a transport that connects directly (no proxy).
// Each transport instance will create its own connection pool, allowing
// multiple connections to the same host for load balancing scenarios.
func DirectTransport() (*http.Transport, error) {
	return &http.Transport{
		MaxConnsPerHost: 1,
		MaxIdleConns:    1,
	}, nil
}
