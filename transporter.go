package loggregator

import (
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"log"
	"net"
)

type metronTransporter struct {
	options       *Options
	udpConn net.Conn
}

func newMetronTransporter(options *Options) *metronTransporter {
	udpConn, err := net.Dial("udp", options.MetronAddress)
	if err != nil {
		log.Printf("Cannot resolve Metron's address %q: %s\n", options.MetronAddress, err.Error())
	}

	return &metronTransporter{
		udpConn:       udpConn,
		options:       options,
	}
}

func (t *metronTransporter) send(points []*dataPoint) {
	for _, d := range points {
		err := t.writeMessage(d)
		if err != nil {
			log.Printf("Cannot write envelope to metron agent: %s\n", err.Error())
		}
	}
}

func (t *metronTransporter) writeMessage(dataPoint *dataPoint) error {
	tags := make(map[string]string)
	tags["type"] = dataPoint.Type

	for k, v := range t.options.Tags {
		tags[k] = v
	}

	envelope := &events.Envelope{
		Origin:     proto.String(t.options.Origin),

		Timestamp: proto.Int64(dataPoint.Timestamp),
		EventType: events.Envelope_ValueMetric.Enum(),

		ValueMetric: &events.ValueMetric{
			Name:  proto.String(dataPoint.Name),
			Value: proto.Float64(dataPoint.Value),
			Unit:  proto.String(dataPoint.Unit),
		},

		Tags: tags,
	}

	bytes, err := proto.Marshal(envelope)
	if err != nil {
		return err
	}

	_, err = t.udpConn.Write(bytes)
	return err
}
