# API Aggregator Bridge 
Bridges implementation that can generically aggregate numerical values for any given amount of APIs.

**Supported Aggregation Methods:**
- Mode
- Median
- Mean

### Setup Instructions
#### Local Install
Make sure [Golang](https://golang.org/pkg/) is installed.

Build (in the root of the bridges repository):
```
GO111MODULE=on go build examples/apiaggregator/main -o apiaggregator
```

Then run the bridge:
```
./apiaggregator
```

#### Docker
To run the container:
```
docker run -it -p 8080:8080 linkpool/apiaggregator-bridge:latest
```

#### AWS Lambda

```bash
zip api_aggregator.zip ./apiaggregator
```

Upload the the zip file into AWS and then use `apiaggregator` as the
handler.

**Important:** Set the `LAMBDA` environment variable to `true` in AWS for
the adaptor to be compatible with Lambda.

### Solidity Usage

Example: https://github.com/linkpoolio/example-chainlinks/blob/master/contracts/APIAggregatorConsumer.sol

### Testing

To call the API, you need to send a POST request to `http://localhost:<port>/` with the request body being of the ChainLink `RunResult` type.

For example:
```bash
curl -X POST http://localhost:8080/ \
-H 'Content-Type: application/json' \
-d @- << EOF
{
	"jobId": "1234",
	"data": {
		"api": ["https://www.bitstamp.net/api/v2/ticker/btcusd/", "https://api.pro.coinbase.com/products/btc-usd/ticker"],
		"paths": ["$.last", "$.price"],
		"aggregationType": "median"
	}
}
EOF
```
Should return something similar to:
```json
{
    "jobRunId": "1234",
    "status": "completed",
    "error": null,
    "pending": false,
    "data": {
        "EUR": 141.65,
        "JPY": 17864.71,
        "USD": 160.11,
        "aggregationType": "median",
        "api": [
            "https://www.bitstamp.net/api/v2/ticker/btcusd/",
            "https://api.pro.coinbase.com/products/btc-usd/ticker"
        ],
        "paths": [
            "$.last",
            "$.price"
        ]
    }
}
```
