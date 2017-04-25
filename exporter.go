package loggregator

import (
	"github.com/rcrowley/go-metrics"
	"time"
)

type dataPoint struct {
	Name      string
	Type      string
	Value     float64
	Timestamp int64
	Unit      string
}

type transporter interface {
	send([]*dataPoint)
}

type timeHelper interface {
	currentTimeInMillis() int64
}

type Options struct {
	Frequency time.Duration
	TimeUnit  time.Duration

	MetronAddress string

	Origin string
	Tags   map[string]string
}

func Loggregator(registry metrics.Registry, options *Options) {

	if options.MetronAddress == "" {
		options.MetronAddress = "127.0.0.1:3457"
	}

	if options.Origin == "" {
		options.Origin
	}

	if int64(options.TimeUnit) == 0 {
		options.TimeUnit = time.Millisecond
	}

	if int64(options.Frequency) == 0 {
		options.Frequency = time.Minute
	}

	timer := time.NewTimer(options.Frequency)
	transport := newMetronTransporter(options)
	exporter := newExporter(transport, &realTimeHelper{}, options.TimeUnit)

	for {
		<-timer.C
		timer.Reset(options.Frequency)

		exporter.exportMetrics(registry)
	}
}
func (e *exporter) exportMetrics(registry metrics.Registry) {
	dataPoints := e.assembleDataPoints(registry)
	e.transport.send(dataPoints)
}

type exporter struct {
	transport  transporter
	timeHelper timeHelper
	timeUnit   time.Duration
}

func newExporter(transport transporter, timeHelper timeHelper, timeUnit time.Duration) *exporter {
	return &exporter{
		transport:  transport,
		timeHelper: timeHelper,
		timeUnit:   timeUnit,
	}
}

func (e *exporter) assembleDataPoints(registry metrics.Registry) []*dataPoint {
	data := make([]*dataPoint, 0)
	currentTime := e.timeHelper.currentTimeInMillis()

	registry.Each(func(name string, metric interface{}) {
		switch m := metric.(type) {
		case metrics.Counter:
			data = append(data, convertCounter(m.Snapshot(), name, currentTime))
		case metrics.Gauge:
			data = append(data, convertGauge(m.Snapshot(), name, currentTime))
		case metrics.GaugeFloat64:
			data = append(data, convertGaugeFloat64(m.Snapshot(), name, currentTime))
		case metrics.Meter:
			data = append(data, convertMeter(m.Snapshot(), name, currentTime)...)
		case metrics.Timer:
			data = append(data, convertTimer(m.Snapshot(), name, currentTime, e.timeUnit)...)
		case metrics.Histogram:
			data = append(data, convertHistogram(m.Snapshot(), name, currentTime)...)
		}
	})

	return data
}
