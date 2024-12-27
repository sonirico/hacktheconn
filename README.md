# HackTheConn

**HackTheConn** is a Go package designed to overcome the limitations of the default HTTP connection management in `net/http`. While Go’s HTTP client optimizes for performance by reusing connections based on `host:port`, this can result in uneven load distribution, rigid connection reuse policies, and difficulty in managing routing strategies. HackTheConn provides dynamic, pluggable strategies for smarter connection management.

## Table of Contents

1. [Why HackTheConn?](#why-hacktheconn)
2. [Features](#features)
3. [Installation](#installation)
4. [Quick Start](#quick-start)
   - [Round-Robin Strategy Example](#round-robin-strategy-example)
   - [Fill Holes Strategy Example](#fill-holes-strategy-example)
   - [Least Response Time Strategy Example](#least-response-time-strategy-example)
5. [Strategies](#strategies)
   - [Round-Robin](#round-robin)
   - [Fill Holes](#fill-holes)
   - [Least Response Time](#least-response-time)
   - [Custom Strategies](#custom-strategies)
6. [Contributing](#contributing)
7. [License](#license)
8. [Acknowledgments](#acknowledgments)

## Why HackTheConn?

By default, Go’s `net/http` pools and reuses connections aggressively:

- **Host Affinity**: Connections are tied to `host:port`, leading to imbalanced loads.
- **Shared Connection Pools**: Multiple `http.Client` instances often share the same underlying connection pool, making it hard to control routing.
- **Lack of Flexibility**: Fine-grained control over connection management and custom routing (e.g., proxy selection) is not natively supported.

### **HackTheConn Fixes These Problems**

- **Smarter Strategies**: Implements advanced algorithms like Round-Robin and Fill Holes to distribute requests more effectively.
- **Dynamic Proxy Management**: Integrates seamlessly with HTTP and SOCKS5 proxies.
- **Real-Time Optimization**: Designed for high-throughput, low-latency scenarios where fairness and control are critical.

## Features

- **Round-Robin Strategy**: Distributes requests evenly across connections.
- **Fill Holes Strategy**: Routes requests to connections with the fewest concurrent requests.
- **Least Response Time Strategy**: Dynamically selects the transport with the lowest response time, supporting customizable calculators (e.g., moving average, weighted average).
- **Customizable Strategies**: Extendable with your own connection balancing algorithms.
- **Proxy-Aware**: Supports both HTTP and SOCKS5 proxies.
- **Optimized for Real-Time Applications**: Ensures fairness and low latency in high-throughput environments.

## Installation

To install HackTheConn, use:

```bash
go get github.com/sonirico/hacktheconn
```

## Quick Start

### Round-Robin Strategy Example

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/sonirico/hacktheconn"
)

func main() {
	proxies := []string{
		"http://proxy1.example.com",
		"http://proxy2.example.com",
	}

	// Create a round-robin transport
	transport := hacktheconn.TransportRoundRobin(proxies)

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Get("https://example.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)
}
```

### Fill Holes Strategy Example

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/sonirico/hacktheconn"
)

func main() {
	proxies := []string{
		"http://proxy1.example.com",
		"http://proxy2.example.com",
		"http://proxy3.example.com",
	}

	// Create a fill holes transport
	transport := hacktheconn.TransportFillHoles(proxies)

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Get("https://example.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)
}
```

### Least Response Time Strategy Example

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sonirico/hacktheconn"
)

func main() {
	proxies := []string{
		"http://proxy1.example.com",
		"http://proxy2.example.com",
	}

	// Create a least response time transport with weighted average calculator
	transport := hacktheconn.TransportLeastResponseTime(
		proxies,
		hacktheconn.OptLeastResponseTimeWithCalculator(
			hacktheconn.LeastResponseTimeWeightedAverageCalculator(0.75),
		),
	)

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Get("https://example.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)
}
```

## Strategies

### Round-Robin

Distributes requests evenly across available connections. Ensures that each transport gets its fair share of requests, avoiding overloading any single connection.

### Fill Holes

Routes requests to the transport with the fewest concurrent requests. Ideal for environments with uneven workloads, ensuring efficient utilization of resources.

### Least Response Time

Selects the transport with the lowest response time. This strategy supports multiple calculators:

- **Last Response Time Calculator**: Uses the most recent response time.
- **Moving Average Calculator**: Averages the response times of the last `n` requests.
- **Weighted Average Calculator**: Applies a weighted average, giving more importance to recent response times.

### Custom Strategies

You can implement your own strategy by following the `Strategy` interface:

```go
type Strategy interface {
	Acquire() (http.RoundTripper, error)
	Release(http.RoundTripper)
}
```

## Contributing

We welcome contributions! Feel free to submit issues or pull requests.

1. Fork the repository
2. Create a feature branch
3. Submit a pull request

## License

HackTheConn is licensed under the MIT License. See `LICENSE` for details.

## Acknowledgments

HackTheConn was inspired by the challenges faced in managing high-load environments, like those encountered at [Atani Labs](https://github.com/atani-labs), and the hacker spirit of tweaking systems to achieve optimal performance.

