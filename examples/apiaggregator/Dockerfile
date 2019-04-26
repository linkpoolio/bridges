FROM golang:1.12-alpine as builder

ENV GO111MODULE=on

RUN apk add --no-cache make curl git gcc musl-dev linux-headers
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

ADD . /go/src/github.com/linkpoolio/bridges
RUN cd /go/src/github.com/linkpoolio/bridges/examples/apiaggregator && go get && go build -o apiaggregator

# Copy into a second stage container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/linkpoolio/bridges/examples/apiaggregator/apiaggregator /usr/local/bin/

ENTRYPOINT ["apiaggregator"]