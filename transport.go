package hacktheconn

import (
	"net/http"
	"net/url"
)

// DefaultTransportFactory creates a transport based on the provided proxy URL.
// It supports "http", "https", "direct", and SOCKS5 protocols.
// If the scheme is "http" or "https", it uses ProxyHTTPTransport.
// If the scheme is "direct", it uses DirectTransport.
// For any other scheme, it defaults to ProxySocks5Transport.
// This is useful for dynamically selecting transports based on the proxy URL.
// It returns an error if the proxy URL is invalid.
func DefaultTransportFactory(proxyURL string) (*http.Transport, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http", "https":
		return ProxyHTTPTransport(proxyURL)
	case "direct":
		return DirectTransport()
	default:
		return ProxySocks5Transport(proxyURL)
	}
}

// DirectTransportFactory creates multiple direct transports (no proxy).
// This is useful when you're behind a load balancer and want multiple connections
// to take advantage of upstream rebalancing.
func DirectTransportFactory(_ string) (*http.Transport, error) {
	return DirectTransport()
}

// MultiDirectTransportFactory creates a slice of direct transports for strategies.
func MultiDirectTransportFactory(count int) []string {
	transports := make([]string, count)
	for i := range count {
		transports[i] = "direct://"
	}
	return transports
}
