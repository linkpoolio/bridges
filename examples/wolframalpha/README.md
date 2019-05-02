# WolframAlpha Short Answers Bidge
Bridges implementation that uses the WolframAlpha Short Answers API to pass in natural language queries.

Docs: http://products.wolframalpha.com/short-answers-api/documentation/

### Setup Instructions
#### Local Install
Make sure [Golang](https://golang.org/pkg/) is installed.

Build (in the root of the bridges repository):
```
GO111MODULE=on go build examples/wolframalpha/main -o wolframalpha
```

Then run the bridge:
```
./wolframalpha
```

#### Docker
To run the container:
```
docker run -it -p 8080:8080 linkpool/wolframalpha-bridge:latest
```

#### AWS Lambda

```bash
zip wolframalpha.zip ./wolframalpha
```

Upload the the zip file into AWS and then use `wolframalpha` as the
handler.

**Important:** Set the `LAMBDA` environment variable to `true` in AWS for
the adaptor to be compatible with Lambda.

### Solidity Usage

Example: https://github.com/linkpoolio/example-chainlinks/blob/master/contracts/WolframAlphaConsumer.sol

### Testing

To call the API, you need to send a POST request to `http://localhost:<port>/` with the request body being of the ChainLink `RunResult` type.

For example:
```bash
curl -X POST http://localhost:8080/fetch \
-H 'Content-Type: application/json' \
-d @- << EOF
{
	"jobRunId": "1234",
	"data": {
		"query": "What's the distance between Los Angeles and New York?",
		"index": 1
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
        "full": "about 2464 miles",
        "index": 1,
        "query": "What's the distance between Los Angeles and New York?",
        "result": "2464",
        "unit": "miles"
    }
}
```