This is the `go-metrics` exporter for Loggregator.


## 1. Adding the exporter to your program

Add the 2 required packages:

```
go get github.com/rcrowley/go-metrics
git clone git@github.com/pivotal-cf/go-metrics-loggregator $GOPATH/src/github.com/pivotal-cf/go-metrics-loggregator
```

Add to your main.go:

```
import (
	"github.com/pivotal-cf/go-metrics-loggregator"
	"github.com/rcrowley/go-metrics"
)

func main() {
    go loggregator.Loggregator(metrics.DefaultRegistry, &loggregator.Options{ Origin: "RedisAgent" })
    ...
}
```

If you'd like to include extra tags in the dropsonde envelope, for instance a serviceGuid:

```
func main() {
    tags := make(map[string]string)
    tags["serviceGuid"] = "Abc-123"
    go loggregator.Loggregator(metrics.DefaultRegistry, &loggregator.Options{
        Tags: tags,
        Origin: "RedisAgent",
    })
    ...
}
```

## 2. Adding a metron agent to your bosh vm

```
jobs:
- name: dedicated-node
  templates:
  ...
  - name: metron
    release: loggregator
  - name: consul_agent
    release: consul
  properties:
    metron_endpoint:
      shared_secret:  (( fill me in ))
    metron_agent:
      listening_port: 3457
      deployment: (( fill me in ))
      etcd:
        client_cert: (( fill me in ))
        client_key: (( fill me in ))
    loggregator:
      etcd:
        require_ssl: true
        machines: ['cf-etcd.service.cf.internal']
        ca_cert: (( fill me in ))
    consul:
      encrypt_keys:
      - (( fill me in ))
      ca_cert: (( fill me in ))
      agent_cert: (( fill me in ))
      agent_key: (( fill me in ))
      server_cert: (( fill me in ))
      server_key: (( fill me in ))
      agent:
        domain: cf.internal
        servers:
          lan: (( fill me in ))
    nats:
      machines: (( fill me in ))
      password: (( fill me in ))
      port: 4222
      user: (( fill me in ))

```

or with bosh links:

```
TODO
```

Please see the spec files for more details:
- https://github.com/cloudfoundry/loggregator/blob/develop/jobs/metron_agent/spec


## 3. Instrumenting your program

Any of the go-metrics metrics will work. If you'd like to, for instance, time a potentially lengthy operation:

```
import (
	"github.com/rcrowley/go-metrics"
)

func instrumentLengthyOperation() {
    timer := metrics.GetOrRegisterTimer("lengthy", metrics.DefaultRegistry)
    timer.Time(func() {
        lengthyOperation()
    })
}
```

## 4. Testing metrics

Each go-metric (Gauge, Counter, Timer, Meter, Histogram) is an interface, so they're quite simple to create fakes for and register with the registry before your code calls `GetOrRegister*`. See exporter_test.go for examples.