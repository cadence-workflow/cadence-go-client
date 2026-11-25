# Developing cadence-client

This doc is intended for contributors to `cadence-client` (hopefully that's you!)

> 📚 **New to contributing to Cadence?** Check out our [Contributing Guide](https://cadenceworkflow.io/community/how-to-contribute/getting-started) for an overview of the contribution process across all Cadence repositories. This document contains cadence-client specific setup and development instructions.

Once you go through the rest of this doc and get familiar with local development setup, feel free to start contributing!

Join our community on the CNCF Slack workspace at [cloud-native.slack.com](https://communityinviter.com/apps/cloud-native/cncf) in the **#cadence-users** channel to reach out and discuss issues with the team.

**Note:** All contributors also need to fill out the [Uber Contributor License Agreement](http://t.uber.com/cla) before we can merge in any of your changes

## Development Environment

* Go. Install on OS X with `brew install go`. The minimum required Go version is visible in the go.mod file.
* `make build` should download and build anything else necessary.

## Checking out the code

Make sure the repository is cloned to the correct location:
(Note: the path is `go.uber.org/cadence/` instead of github repo)

```bash
go get -d go.uber.org/cadence
cd $GOPATH/src/go.uber.org/cadence
```

## Dependency management

Dependencies are tracked via go modules (tool-only dependencies in `tools.go`), and we no longer support dep/glide/etc
at all.  If you wish to check for dependencies that may be good to update, try `make deps`.

Dependencies should generally be as _low_ of a version as possible, not whatever is "current",
unless the project has a proven history of API stability and reliability.  go.uber.org/yarpc is a good example of this.  
Dependency upgrades may force our users to upgrade as well, so only upgrade them when necessary,
like on a new needed feature or an important bugfix.

## API stability

While we are pre-1.0 we do our best to maintain backwards compatibility at a binary *and semantic* level, but there will
be some gaps where this is not practical.  Any known breakages should be mentioned in the associated Github releases
(soon: also changelog.md).  
Generally speaking though, all obviously-intentionally-exposed APIs must be highly stable, take care to maintain it.

As a concrete example of gaps that may occur, any use of our generated Thrift or gRPC APIs is quite fragile.  High level
APIs that accidentally expose this will be _relatively_ stable at the top level of fields/methods, but e.g. use of the
client.GetWorkflowHistory iterator (which exposes _many_ raw RPC types) is effectively stable only by accident.

In the future we intend to have a clearer split between long-term-stable APIs and ones that are only best-effort.
Ideally the best-effort APIs will be for special cases only, e.g. RPC clients, and users that access them should be made
aware of the instability (and how to update when breakages occur) where possible.

## Licence headers

This project is Open Source Software, and requires a header at the beginning of
all source files. `make build` should enforce this for you.

To re-generate copyright headers, e.g. to add missing ones, run `make copyright`.

## Testing

To run all the tests with coverage and race detector enabled, start a cadence server (e.g. via `docker compose up`) and:
```bash
make test
```

You can run only unit tests, which is quicker and likely sufficient while in the middle of development, with:
```bash
make unit_test
```