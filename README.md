This is the `go-metrics` exporter for Loggregator. The goal of this library is to make it easy for CF/bosh component developers to define and emit metrics to the Loggregator firehose using a standard metrics library ([go-metrics](https://github.com/rcrowley/go-metrics)) as opposed to Loggregator's [dropsonde](https://github.com/cloudfoundry/dropsonde).

go-metrics aggregates metrics in memory and this exporter emits them on a specified time interval, defaulting to once per minute. The advantage to using go-metrics over dropsonde is that it provides many more metric types. The additional types, such as Timer, expose more information than a standard counter or gauge, such as percentiles (75th, 90th, 95th, 99th by default) and rates (1 minute, 5, 10, 15, and mean). The documentation for dropwizard (the java library that go-metrics is based on) can be found [here](http://metrics.dropwizard.io/3.2.2/getting-started.html).

## 1. Adding the exporter to your program

Add the 4 required packages:

```
go get github.com/rcrowley/go-metrics
go get github.com/pivotal-cf/go-metrics-loggregator
go get github.com/cloudfoundry/sonde-go
go get github.com/gogo/protobuf
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
Since metrics are aggregated in go-metrics, each benchmark won't individually emit to the firehose, but they'll be rolled up into derived metrics (https://github.com/rcrowley/go-metrics/blob/master/timer.go#L9)

Emitting the health of the service could look like:

```
metrics.NewRegisteredFunctionalGauge("health", metrics.DefaultRegistry, func() int64 {
 c := redisPool.Get()
 defer c.Close()

 err := c.Ping()
 if err != nil {
  return 0
 }

 return 1
})
```

## 4. Testing metrics

Each go-metric (Gauge, Counter, Timer, Meter, Histogram) is an interface, so they're quite simple to create fakes for and register with the registry before your code calls `GetOrRegister*`. See exporter_test.go for examples.


# Metrics best practices

TODO

# Dropsonde format

All go-metrics are converted to dropsonde's "Envelope_ValueMetric" and they are sent through the firehose like so:

```
events.Envelope{
 Origin:     proto.String("RedisAgent"),
 EventType:  events.Envelope_ValueMetric.Enum(),
 Timestamp:  proto.Int64(1493138903202394827),
 ValueMetric: &events.ValueMetric{
  Name:  proto.String("cpu.usage"),
  Value: proto.Float64(46.45),
  Unit:  proto.String("percentage"),
 },
 Tags: map[string]string{
  "serviceGuid": "abc-123",
  "type":        "gauge",
 },
}
```
