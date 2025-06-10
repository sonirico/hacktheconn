# HackTheConn

**HackTheConn** is a Go package designed to overcome the limitations of the default HTTP connection management in `net/http`. While Go's HTTP client optimizes for performance by reusing connections based on `host:port`, this can result in uneven load distribution, rigid connection reuse policies, and difficulty in managing routing strategies. HackTheConn provides dynamic, pluggable strategies for smarter connection management.

## Table of Contents

- [HackTheConn](#hacktheconn)
	- [Table of Contents](#table-of-contents)
	- [Why HackTheConn?](#why-hacktheconn)
		- [**HackTheConn Fixes These Problems**](#hacktheconn-fixes-these-problems)
	- [Features](#features)
	- [Installation](#installation)
	- [Quick Start](#quick-start)
		- [Round-Robin Strategy Example](#round-robin-strategy-example)
		- [Fill Holes Strategy Example](#fill-holes-strategy-example)
		- [Least Response Time Strategy Example](#least-response-time-strategy-example)
		- [Direct Connections Example](#direct-connections-example)
		- [Mixed Connections Example](#mixed-connections-example)
	- [Strategies](#strategies)
		- [Round-Robin](#round-robin)
		- [Fill Holes](#fill-holes)
		- [Least Response Time](#least-response-time)
		- [Direct Connections](#direct-connections)
		- [Custom Strategies](#custom-strategies)
	- [Contributing](#contributing)
	- [License](#license)
	- [Acknowledgments](#acknowledgments)

## Why HackTheConn?

By default, Go's `net/http` pools and reuses connections aggressively:

- **Host Affinity**: Connections are tied to `host:port`, leading to imbalanced loads.
- **Shared Connection Pools**: Multiple `http.Client` instances often share the same underlying connection pool, making it hard to control routing.
- **Lack of Flexibility**: Fine-grained control over connection management and custom routing (e.g., proxy selection) is not natively supported.

### **HackTheConn Fixes These Problems**

- **Smarter Strategies**: Implements advanced algorithms like Round-Robin and Fill Holes to distribute requests more effectively.
- **Dynamic Proxy Management**: Integrates seamlessly with HTTP and SOCKS5 proxies.
- **Direct Connection Support**: Creates multiple direct connections to leverage upstream load balancers.
- **Real-Time Optimization**: Designed for high-throughput, low-latency scenarios where fairness and control are critical.

## Features

- **Round-Robin Strategy**: Distributes requests evenly across connections.
- **Fill Holes Strategy**: Routes requests to connections with the fewest concurrent requests.
- **Least Response Time Strategy**: Dynamically selects the transport with the lowest response time, supporting customizable calculators (e.g., moving average, weighted average).
- **Direct Connections**: Creates multiple direct connections for upstream load balancing scenarios.
- **Customizable Strategies**: Extendable with your own connection balancing algorithms.
- **Proxy-Aware**: Supports both HTTP and SOCKS5 proxies.
- **Mixed Connection Types**: Combine proxies with direct connections in the same strategy.
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

### Direct Connections Example

When you're behind a load balancer or proxy that handles upstream distribution, you can create multiple direct connections:

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/sonirico/hacktheconn"
)

func main() {
    // Create 5 direct connections using round-robin
    transport := hacktheconn.TransportDirectRoundRobin(5)

    client := &http.Client{
        Transport: transport,
    }

    resp, err := client.Get("https://api.example.com/data")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    fmt.Println("Response status:", resp.Status)
}
```

### Mixed Connections Example

You can mix proxy URLs with direct connections by using `"direct://"` in your proxy list:

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/sonirico/hacktheconn"
)

func main() {
    connections := []string{
        "http://proxy1.example.com",
        "direct://", // Direct connection
        "http://proxy2.example.com",
        "direct://", // Another direct connection
    }

    transport := hacktheconn.TransportFillHoles(connections)

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

**Available functions:**

- `TransportRoundRobin(proxies []string, opts ...OptRoundRobin)` - With proxy configuration
- `TransportDirectRoundRobin(connectionCount int, opts ...OptRoundRobin)` - Direct connections only

### Fill Holes

Routes requests to the transport with the fewest concurrent requests. Ideal for environments with uneven workloads, ensuring efficient utilization of resources.

**Available functions:**

- `TransportFillHoles(proxies []string, opts ...OptFillHoles)` - With proxy configuration
- `TransportDirectFillHoles(connectionCount int, opts ...OptFillHoles)` - Direct connections only

### Least Response Time

Selects the transport with the lowest response time. This strategy supports multiple calculators:

- **Last Response Time Calculator**: Uses the most recent response time.
- **Moving Average Calculator**: Averages the response times of the last `n` requests.
- **Weighted Average Calculator**: Applies a weighted average, giving more importance to recent response times.

**Available functions:**

- `TransportLeastResponseTime(proxies []string, opts ...OptLeastResponseTime)` - With proxy configuration
- `TransportDirectLeastResponseTime(connectionCount int, opts ...OptLeastResponseTime)` - Direct connections only

### Direct Connections

Creates multiple direct connections (no proxy) to the same destination. This is useful when:

- You're behind a load balancer that distributes connections across multiple backends
- You want to leverage upstream rebalancing capabilities
- You need to bypass connection pooling limitations for specific scenarios

**Use cases:**

- **Load Balancers Upstream**: ALB, HAProxy, Nginx upstream balancing
- **CDNs**: Leverage edge caching with multiple connections
- **Microservices**: When service mesh handles balancing
- **Rate Limited APIs**: Distribute load across multiple TCP connections

### Custom Strategies

You can implement your own strategy by following the `Strategy` interface:

```go
type Strategy interface {
    Acquire() (http.RoundTripper, error)
    Release(http.RoundTripper)
}
```

Example custom strategy implementation:

```go
type MyCustomStrategy struct {
    transports []http.RoundTripper
    // your custom fields
}

func (s *MyCustomStrategy) Acquire() (http.RoundTripper, error) {
    // Your custom logic here
    return s.transports[0], nil
}

func (s *MyCustomStrategy) Release(transport http.RoundTripper) {
    // Your custom cleanup logic here
}

// Use it with StrategyTransport
transport := hacktheconn.Transport(&MyCustomStrategy{
    transports: myTransports,
})
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
