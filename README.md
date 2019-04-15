<p align="center">
  <img src="https://s3.linkpool.io/images/bridgestype.png">
</p>

[![Build Status](https://travis-ci.org/linkpoolio/bridges.svg?branch=master)](https://travis-ci.org/linkpoolio/bridges)
[![codecov](https://codecov.io/gh/linkpoolio/bridges/branch/master/graph/badge.svg)](https://codecov.io/gh/linkpoolio/bridges)
[![Go Report Card](https://goreportcard.com/badge/github.com/linkpoolio/bridges)](https://goreportcard.com/report/github.com/linkpoolio/bridges)
-----------------------

Bridges is a Chainlink adaptor framework, lowering the barrier of entry for anyone to create their own:

- Bridges CLI application, allowing you to quickly run an adaptor.
- Create adaptors with an easy to interpret JSON schema.
- Simple interface to implement your own custom adaptors that can do anything.
- Supports running in serverless environments such as AWS Lambda & GCP functions.

### Install

View the [releases page](https://github.com/linkpoolio/bridges/releases) and download the latest version for your
operating system, then add it to PATH.

### Quick Usage

For the simplest adaptor, run the following:
```
bridges -b https://s3.linkpool.io/bridges/cryptocompare.json
```
Once running, the adaptor will be started on port 8080.

### Usage
```
Usage of bridges:
  -b, --bridge string   Filepath/URL of bridge JSON file
  -p, --port int        Server port (default 8080)
```

With the `-b` flag, either URLs or relative file paths can be specified, for example:
```
bridges -b https://s3.linkpool.io/bridges/rapidapi.json
```
is equal to
```
bridges -b json/rapidapi.json
```

### Lambda Usage

View the [releases page](https://github.com/linkpoolio/bridges/releases) and download the Linux x86-64 zip. Upload the zip 
into Lambda and set the handler as `bridges`.

Then set the following environment variables:

 - `LAMBDA=true`
 - `BRIDGE=<your bridge url>`
 
Since bridges queries the bridge URL each call, it's recommend to host your own JSON files in S3 for latency and 
your own redundancy. This is not the case when running locally or using docker.
 
### Docker Usage

Run by either appending arguments or setting environment variables:
```
docker run -it linkpool/bridges:latest -b https://s3.linkpool.io/bridges/rapidapi.json
```

### Examples

JSON:

- [CryptoCompare](json/cryptocompare.json): Simplest example.
- [AlphaVantage](json/alphavantage.json): Uses GET param authentication, param passthrough.
- [RapidAPI](json/rapidapi.json): Two adaptors specified, header authentication and param & form passthrough.

Interface implementations:

- [CryptoCompare](examples/cryptocompare): Simplest example.
- [API Aggregator](examples/apiaggregator): Aggregates multiple endpoints using mean/median/mode. 
- [Wolfram Alpha](examples/wolframalpha): Short answers API, non-JSON, uses string splitting.

### Implement your own

```go
package main

import (
	"github.com/linkpoolio/bridges/bridge"
)

type MyAdaptor struct{}

func (ma *MyAdaptor) Run(h *bridge.Helper) (interface{}, error) {
	return map[string]string{"hello": "world"}, nil
}

func (ma *MyAdaptor) Opts() *bridge.Opts {
	return &bridge.Opts{
		Name:   "MyAdaptor",
		Lambda: true,
	}
}

func main() {
	bridge.NewServer(&MyAdaptor{}).Start(8080)
}
```

### TODO

- [ ] Increase test coverage
- [ ] Support S3 urls for adaptor fetching
- [ ] Look at the validity of doing a Docker Hub style adaptor repository

### Contributing

We welcome all contributors, please raise any issues for any feature request, issue or suggestion you may have.