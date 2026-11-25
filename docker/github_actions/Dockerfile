FROM golang:1.26.10-alpine3.23

# install the tools missing the in the alpine base image
RUN apk upgrade --no-cache && apk add --no-cache bash build-base git grep

RUN mkdir -p /go/src/go.uber.org/cadence
WORKDIR /go/src/go.uber.org/cadence

# allow git-status and similar to work
RUN git config --global --add safe.directory /go/src/go.uber.org/cadence
