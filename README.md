# Go framework for Cadence
[![CI Checks](https://github.com/cadence-workflow/cadence-go-client/actions/workflows/ci-checks.yml/badge.svg)](https://github.com/cadence-workflow/cadence-go-client/actions/workflows/ci-checks.yml)
[![Coverage](https://codecov.io/gh/uber-go/cadence-client/graph/badge.svg?token=iEpqo5HbDe)](https://codecov.io/gh/uber-go/cadence-client)
[![GoDoc](https://godoc.org/go.uber.org/cadence?status.svg)](https://godoc.org/go.uber.org/cadence)

[Cadence](https://github.com/uber/cadence) is a distributed, scalable, durable, and highly available orchestration engine we developed at Uber Engineering to execute asynchronous long-running business logic in a scalable and resilient way.

`cadence-client` is the framework for authoring workflows and activities.

## How to use

Make sure you clone this repo into the correct location.

```bash
git clone git@github.com:uber-go/cadence-client.git $GOPATH/src/go.uber.org/cadence
```

or

```bash
go get go.uber.org/cadence
```

See [samples](https://github.com/uber-common/cadence-samples) to get started.

Documentation is available [here](https://cadenceworkflow.io/docs/go-client/).
You can also find the API documentation [here](https://godoc.org/go.uber.org/cadence).

## Contributing
We'd love your help in making the Cadence Go client great. Please review our [contribution guidelines](CONTRIBUTING.md).

## License
Apache 2.0 License, please see [LICENSE](LICENSE) for details.
