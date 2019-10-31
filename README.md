<p align="center">
  <img src="https://s3.linkpool.io/images/bridgestype.png">
</p>

[![Build Status](https://travis-ci.org/linkpoolio/bridges.svg?branch=master)](https://travis-ci.org/linkpoolio/bridges)
[![codecov](https://codecov.io/gh/linkpoolio/bridges/branch/master/graph/badge.svg)](https://codecov.io/gh/linkpoolio/bridges)
[![Go Report Card](https://goreportcard.com/badge/github.com/linkpoolio/bridges)](https://goreportcard.com/report/github.com/linkpoolio/bridges)
-----------------------

Bridges is a Chainlink adaptor framework, lowering the barrier of entry for anyone to create their own:

- A tested hardened library that removes the need to build your own HTTP server, allowing you to just focus on 
adapter requirements.
- Simple interface to allow you to build an adapter that confides to Chainlink schema.
- Kept up to date with any changes, meaning no extra work for existing adapters to support new schema changes or 
features.
- Supports running in serverless environments such as AWS Lambda & GCP functions with minimal effort.

## Contents
1. [Code Examples](#code-examples)
2. [Running in AWS Lambda](#running-in-aws-lambda)
3. [Running in GCP Functions](#running-in-gcp-functions)
4. [Example Implementations](#example-implementations)
    - [Basic](#basic)
    - [Unauthenticated HTTP Calls](#unauthenticated-http-calls)
    - [Authenticated HTTP Calls](#authenticated-http-calls)

## Code Examples

- [CryptoCompare](examples/cryptocompare): Simplest example.
- [API Aggregator](examples/apiaggregator): Aggregates multiple endpoints using mean/median/mode. 
- [Wolfram Alpha](examples/wolframalpha): Short answers API, non-JSON, uses string splitting.
- [Gas Station](examples/gasstation): Single answer response, no authentication.
- [Asset Price](https://github.com/linkpoolio/asset-price-cl-ea): A more complex example that aggregates crypto asset 
prices from multiple exchanges by weighted volume. 

## Running in Docker
After implementing your bridge, if you'd like to run it in Docker, you can reference the Dockerfiles in 
[examples](examples/cryptocompare/Dockerfile) to then use as a template for your own Dockerfile.

## Running in AWS Lambda
After you've completed implementing your bridge, you can then test it in AWS Lambda. To do so:

1. Build the executable:
    ```bash
    GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o bridge
    ```
2. Add the file to a ZIP archive:
    ```bash
    zip bridge.zip ./bridge
    ```
3. Upload the the zip file into AWS and then use `bridge` as the
handler.
4. Set the `LAMBDA` environment variable to `true` in AWS for
the adaptor to be compatible with Lambda.

## Running in GCP Functions
Due to the difference in running Go within GCP Functions, it requires specific considerations for it be supported 
within your bridge:
- Bridge implementation cannot be within the `main` package
- An extra `Handler` function within your implementation:
    ```go
    func Handler(w http.ResponseWriter, r *http.Request) {
        bridges.NewServer(&Example{}).Handler(w, r)
    }
    ```
- A `go.mod` and `go.sum` within the sub-package that contains the `Handler` function

For an example implementation for GCP Functions, view the 
[asset price adapter](https://github.com/linkpoolio/asset-price-cl-ea).

You can then use the gcloud CLI tool to deploy it, for example:
```bash
gcloud functions deploy bridge --runtime go111 --entry-point Handler --trigger-http
```

## Example Implementations

### Basic
Bridges works by providing a simple interface to confide to. The interface contains two functions, `Run` and `Opts`. 
The `Run` function is called on each HTTP request, `Opts` is called on start-up. Below is a very basic implementation 
that returns the `value` as passed in by Chainlink, set back as `newValue` in the response:

```go
package main

import (
	"github.com/linkpoolio/bridges"
)

type MyAdapter struct{}

func (ma *MyAdapter) Run(h *bridge.Helper) (interface{}, error) {
	return map[string]string{"newValue": h.GetParam("value")}, nil
}

func (ma *MyAdapter) Opts() *bridge.Opts {
	return &bridge.Opts{
		Name:   "MyAdapter",
		Lambda: true,
	}
}

func main() {
	bridge.NewServer(&MyAdaptor{}).Start(8080)
}
```

### Unauthenticated HTTP Calls
The bridges library provides a helper object that intends to make actions like performing HTTP calls simpler, removing 
the need to write extensive error handling or the need to have the knowledge of Go's in-built http libraries.

For example, this below implementation uses the `HTTPCall` function to make a simple unauthenticated call to ETH Gas 
Station:
```go
package main

import (
	"github.com/linkpoolio/bridges"
)

type GasStation struct{}

func (gs *GasStation) Run(h *bridges.Helper) (interface{}, error) {
	obj := make(map[string]interface{})
	err := h.HTTPCall(
		http.MethodGet,
		"https://ethgasstation.info/json/ethgasAPI.json",
		&obj,
	)
	return obj, err
}

func (gs *GasStation) Opts() *bridges.Opts {
	return &bridges.Opts{
		Name:   "GasStation",
		Lambda: true,
	}
}

func main() {
	bridges.NewServer(&GasStation{}).Start(8080)
}
```

### Authenticated HTTP Calls
Bridges also provides an interface to support authentication methods when making HTTP requests to external sources. By 
default, bridges supports authentication via HTTP headers or GET parameters.

Below is a modified version of the WolframAlpha adapter, showing authentication setting the `appid` header from the 
`APP_ID` environment variable:
```go
package main

import (
	"errors"
    "fmt"
	"github.com/linkpoolio/bridges"
	"net/http"
	"os"
	"strings"
)

type WolframAlpha struct{}

func (cc *WolframAlpha) Run(h *bridges.Helper) (interface{}, error) {
	b, err := h.HTTPCallRawWithOpts(
		http.MethodGet,
		"https://api.wolframalpha.com/v1/result",
		bridges.CallOpts{
			Auth: bridges.NewAuth(bridges.AuthParam, "appid", os.Getenv("APP_ID")),
			Query: map[string]interface{}{
				"i": h.GetParam("query"),
			},
		},
	)
	return fmt.Sprint(b), err
}

func (cc *WolframAlpha) Opts() *bridges.Opts {
	return &bridges.Opts{
		Name:   "WolframAlpha",
		Lambda: true,
	}
}

func main() {
	bridges.NewServer(&WolframAlpha{}).Start(8080)
}
```

### Contributing

We welcome all contributors, please raise any issues for any feature request, issue or suggestion you may have.
