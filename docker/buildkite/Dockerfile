FROM golang:1.24

RUN mkdir -p /go/src/go.uber.org/cadence
WORKDIR /go/src/go.uber.org/cadence

ADD go.mod go.sum /go/src/go.uber.org/cadence/

# allow git-status and similar to work
RUN git config --global --add safe.directory /go/src/go.uber.org/cadence
RUN GO111MODULE=on go mod download

# Install Buildkite agent
# https://buildkite.com/docs/agent/v3/ubuntu
RUN apt-get install -y apt-transport-https dirmngr curl
RUN curl -fsSL https://keys.openpgp.org/vks/v1/by-fingerprint/32A37959C2FA5C3C99EFBC32A79206696452D198 | \
  gpg --dearmor -o /usr/share/keyrings/buildkite-agent-archive-keyring.gpg
RUN echo \
  "deb [signed-by=/usr/share/keyrings/buildkite-agent-archive-keyring.gpg] https://apt.buildkite.com/buildkite-agent stable main" | \
  tee /etc/apt/sources.list.d/buildkite-agent.list
RUN apt-get update && apt-get install -yy --no-install-recommends buildkite-agent

# TODO: delete this file after migration to github actions