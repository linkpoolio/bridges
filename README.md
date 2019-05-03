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

## Contents
1. [Install](#install)
2. [Usage](#usage)
3. [Managing API Keys](#managing-api-keys)
4. [Testing your Bridge](#testing-your-bridge)
5. [Chainlink Integration](#chainlink-integration)
6. [Bridge JSON](#bridge-json)
7. [Examples](#examples)

## Install

View the [releases page](https://github.com/linkpoolio/bridges/releases) and download the latest version for your
operating system, then add it to PATH.

## Usage

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

## Managing API Keys

Bridges supports passing in API keys on the bridge http calls. These api keys are fed in as environment variables
on running bridges. For example:

```
API_KEY=my-api-key bridges -b https://s3.linkpool.io/bridges/rapidapi.json
```

It's recommended to use secret managers for storing API keys.

#### Considerations
**API key environment variables may not be named `API_KEY`.** Refer to the JSON file for each variable name,
 in `auth.env`.

Custom implementations of the bridges interface may also completely differ and not use environment variables.

### Lambda Usage

View the [releases page](https://github.com/linkpoolio/bridges/releases) and download the Linux x86-64 zip. Upload the 
zip into Lambda and set the handler as `bridges`.

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

## Testing your Bridge

To test a bridge, you need to send a `POST` request to it in the Chainlink `RunResult` type. For example:

Start your bridge:
```
bridges -b https://s3.linkpool.io/bridges/cryptocompare.json
```

Call it:
```
curl -X POST -d "{\"jobRunId\":\"1234\",\"data\":{\"key\":\"value\"}}" http://localhost:8080
```

Result:
```json
{
   "jobRunId":"1234",
   "status":"completed",
   "error":null,
   "pending":false,
   "data":{
      "EUR":140.88,
      "JPY":17717.05,
      "USD":159.77,
      "key":"value"
   }
}
```

## Chainlink Integration

Once you have a running bridge, you can then add the URL of the running bridge to your Chainlink node in the UI.

1. Login to your Chainlink node
2. Click "Bridges"
3. Add a new bridge
4. Enter your bridges URL, for example: `http://localhost:8080/`

If your bridge has multiple paths or specifies a path other than `/`, you'll need to take that into account when adding 
your bridge in Chainlink. For example, with the [RapidAPI](json/rapidapi.json) example, you'd have two URLs:

- `http://localhost:8080/get`
- `http://localhost:8080/post`

## Bridge JSON

Example JSON file below with all the fields set:
```json
[
  {
    "name": "Example",
    "method": "POST",
    "url": "http://exampleapi.com/endpoint",
    "path": "/",
    "auth": {
      "type": "header",
      "key": "X-API-KEY",
      "env": "API_KEY"
    },
    "opts": {
      "queryPassthrough": false,
      "query": {
        "key": "value"
      },
      "body": "{\"message\":\"Hello\"}",
      "expectedCode": 200
    }
  }
]
```

To then use this save it to file, for example `bridge.json`, then run:
```
bridges -b bridge.json
```

The resulting adaptor will perform the following API call, when called on `POST http://localhost:8080/`:

- HTTP Method: `POST`
- Header X-API-KEY: From environment variable `API_KEY`
- URL: `http://exampleapi.com/endpoint?key=value`
- Body: `{"message":"Hello"}`

It will then check to see if the status code returned was 200.

## Examples

JSON:

- [CryptoCompare](json/cryptocompare.json): Simplest example.
- [AlphaVantage](json/alphavantage.json): Uses GET param authentication, param passthrough.
- [RapidAPI](json/rapidapi.json): Two adaptors specified, header authentication and param & form passthrough.

Interface implementations:

- [CryptoCompare](examples/cryptocompare): Simplest example.
- [API Aggregator](examples/apiaggregator): Aggregates multiple endpoints using mean/median/mode. 
- [Wolfram Alpha](examples/wolframalpha): Short answers API, non-JSON, uses string splitting.

## Implement your own

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