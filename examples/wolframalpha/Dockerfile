FROM golang:1.13-alpine as builder

ENV GO111MODULE=on

RUN apk add git

ADD . /go/src/github.com/linkpoolio/bridges
RUN cd /go/src/github.com/linkpoolio/bridges/examples/wolframalpha \
    && go get \
    && go build -o wolframalpha

# Copy into a second stage container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/linkpoolio/bridges/examples/wolframalpha/wolframalpha /usr/local/bin/

ENTRYPOINT ["wolframalpha"]