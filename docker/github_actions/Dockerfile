FROM golang:1.24

RUN mkdir -p /go/src/go.uber.org/cadence
WORKDIR /go/src/go.uber.org/cadence

ADD go.mod go.sum /go/src/go.uber.org/cadence/

# allow git-status and similar to work
RUN git config --global --add safe.directory /go/src/go.uber.org/cadence
RUN GO111MODULE=on go mod download
