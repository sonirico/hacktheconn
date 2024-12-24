package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

func ProxyHTTPTransport(proxyURL string) (*http.Transport, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}
	return &http.Transport{
		Proxy:           http.ProxyURL(u),
		MaxConnsPerHost: 1,
	}, nil
}

func ProxySocks5Transport(socksAddr string) (*http.Transport, error) {
	dialer, err := proxy.SOCKS5("tcp", socksAddr, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}
	dialContext := func(ctx context.Context, network, address string) (net.Conn, error) {
		// TODO: logger debug
		return dialer.Dial(network, address)
	}
	return &http.Transport{
		DialContext: dialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxConnsPerHost: 1,
	}, nil
}
