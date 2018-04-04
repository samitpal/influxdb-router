// Package writer provides code for wiring metrics to influxdb
// The MIT License (MIT)
//
// Copyright (c) 2017 Samit Pal
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package writer

import (
	"github.com/samitpal/influxdb-router/backends"
	"github.com/samitpal/influxdb-router/config"
	"github.com/samitpal/influxdb-router/logging"
)

var log = logging.For("writer")

//OutQueueWriter starts some goroutines and writes the metric streams to the out going queues.
func OutQueueWriter(apiConf config.APIKeyMap, incomingQueue chan *backends.Payload, ready chan bool) {
	for _, c := range apiConf {
		// start a goroutine for each of the out going queues.
		for _, d := range c.Dests {
			go InfluxWriter(d, c.InfluxDBName, c.InfluxDBUserName, c.InfluxDBPassword)
			go RetryQueueHandler(d, c.InfluxDBName, c.InfluxDBUserName, c.InfluxDBPassword)
			// Note that this will start multiple health check goroutines for diff customer even if the URL is same.
			go d.HealthCheck()
		}
	}

	ready <- true

	// pop messages from the incoming queue and distribute to the relevant out going queues.
	for messages := range incomingQueue {
		conf := apiConf[messages.APIKey]
		for _, v := range conf.Dests {
			go func(m *backends.Payload, d *backends.BackendDest) {
				select {
				case d.Queue <- m:
				default:
					log.Errorf("Error copying messages to outgoing queue of dest %s", d.URL)
				}
			}(messages, v)
		}
	}
}
