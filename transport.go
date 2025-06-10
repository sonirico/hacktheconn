package hacktheconn

import (
	"net/http"
	"net/url"
)

func DefaultTransportFactory(proxyURL string) (*http.Transport, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http", "https":
		return ProxyHTTPTransport(proxyURL)
	default:
		return ProxySocks5Transport(proxyURL)
	}
}
