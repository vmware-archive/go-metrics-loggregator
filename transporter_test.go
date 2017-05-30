// Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package loggregator

import (
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
)

var _ = Describe("metronTransporter", func() {

	It("sends metrics to metron over UDP", func() {
		tc := setup("127.0.0.1:9863")
		defer tc.teardown()

		tags := make(map[string]string)
		tags["serviceGuid"] = "abc-123"

		transporter := newMetronTransporter(&Options{
			MetronAddress: "127.0.0.1:9863",
			Origin:        "some origin",
			Tags:          tags,
		})

		transporter.send([]*dataPoint{
			{
				Name:      "test-counter",
				Type:      "COUNTER",
				Value:     123,
				Timestamp: 872828732,
				Unit:      "counts",
			},
		})

		Expect(<-tc.envelopes).To(Equal(&events.Envelope{
			Origin:     proto.String("some origin"),
			EventType:  events.Envelope_ValueMetric.Enum(),
			Timestamp:  proto.Int64(872828732),
			ValueMetric: &events.ValueMetric{
				Name:  proto.String("test-counter"),
				Value: proto.Float64(123),
				Unit:  proto.String("counts"),
			},
			Tags: map[string]string{
				"serviceGuid": "abc-123",
				"type":        "COUNTER",
			},
		}))

	})
})

type testContext struct {
	envelopes chan *events.Envelope
	teardown  func()
}

func setup(address string) *testContext {
	stop := make(chan interface{})
	envelopes := make(chan *events.Envelope)

	serverConn, err := net.ListenPacket("udp", address)
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()

		buf := make([]byte, 10240000)

		for {
			if shouldStop(stop) {
				return
			}

			n, _, err := serverConn.ReadFrom(buf)
			if err != nil {
				continue
			}

			env := new(events.Envelope)
			err = proto.Unmarshal(buf[0:n], env)
			Expect(err).NotTo(HaveOccurred())

			envelopes <- env
		}
	}()

	return &testContext{
		envelopes: envelopes,
		teardown: func() {
			close(stop)
			serverConn.Close()
		},
	}
}

func shouldStop(c chan interface{}) bool {
	select {
	case <-c:
		return true
	default:
		return false
	}
}
