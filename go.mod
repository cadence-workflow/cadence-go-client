module go.uber.org/cadence

go 1.23

toolchain go1.24.5

require (
	github.com/apache/thrift v0.16.0
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/golang/mock v1.7.0-rc.1
	github.com/google/gofuzz v1.0.0
	github.com/jonboulle/clockwork v0.4.0
	github.com/marusama/semaphore/v2 v2.5.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pborman/uuid v0.0.0-20160209185913-a97ce2ca70fa
	github.com/robfig/cron v1.2.0
	github.com/stretchr/testify v1.9.0
	github.com/uber-go/tally v3.5.8+incompatible
	github.com/uber/cadence-idl v0.0.0-20260424161124-35996664fb49
	github.com/uber/jaeger-client-go v2.30.0+incompatible
	github.com/uber/tchannel-go v1.34.4
	go.uber.org/atomic v1.11.0
	go.uber.org/goleak v1.3.0
	go.uber.org/multierr v1.11.0
	go.uber.org/thriftrw v1.32.0
	go.uber.org/yarpc v1.88.0
	go.uber.org/zap v1.27.0
	golang.org/x/net v0.28.0
	golang.org/x/oauth2 v0.22.0
	golang.org/x/time v0.0.0-20170927054726-6dc17368e09b
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/status v1.1.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.1 // indirect
	github.com/prometheus/client_model v0.6.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	github.com/uber-go/mapdecode v1.0.0 // indirect
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	go.uber.org/dig v1.17.1 // indirect
	go.uber.org/fx v1.22.0 // indirect
	go.uber.org/net/metrics v1.4.0 // indirect
	golang.org/x/exp/typeparams v0.0.0-20221208152030-732eee02a75a // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/grpc v1.67.3 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	honnef.co/go/tools v0.4.3 // indirect
)

retract (
	v1.2.8 // Contains bad commits
	v1.2.5 // Published accidentally
	v1.2.4 // Published accidentally
	v1.2.3 // Published accidentally
	v1.2.2 // Published accidentally
	v1.2.1 // Published accidentally
	v0.25.0 // Published accidentally
	v0.24.0 // Published accidentally
	v0.23.2 // Published accidentally
	v0.23.1 // Published accidentally
	v0.22.4 // Published accidentally
	v0.22.3 // Published accidentally
	v0.22.2 // Published accidentally
	v0.22.1 // Published accidentally
	v0.22.0 // Published accidentally
	v0.21.3 // Published accidentally
	v0.21.2 // Published accidentally
	v0.21.0 // Published accidentally
	v0.20.0 // Published accidentally
	v0.19.2 // Published accidentally
	v0.18.0 // Published accidentally
	v0.16.1 // Published accidentally
	v0.15.1 // Published accidentally
	v0.14.2 // Published accidentally
	v0.10.3 // Published accidentally
	v0.10.2 // Published accidentally
	v0.5.5 // Published accidentally
	v0.5.4 // Published accidentally
	v0.5.3 // Published accidentally
	v0.3.15 // Published accidentally
	v0.3.14 // Published accidentally
	v0.3.13 // Published accidentally
	v0.3.12 // Published accidentally
	v0.3.11 // Published accidentally
	v0.3.9 // Published accidentally
	v0.3.8 // Published accidentally
	v0.3.7 // Published accidentally
	v0.3.6 // Published accidentally
	v0.3.5 // Published accidentally
	v0.3.4 // Published accidentally
	v0.3.3 // Published accidentally
	v0.3.0 // Published accidentally
	v0.2.0 // Published accidentally
)
