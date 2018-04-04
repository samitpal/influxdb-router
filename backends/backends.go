// Package backends provides code for influxdb backends.
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
package backends

import (
	//"io"
	"net/http"
	"sync"
	"time"

	"github.com/samitpal/influxdb-router/logging"
)

var log = logging.For("backends")

// Payload is what is sent to the incoming queue.
type Payload struct {
	MessageID string
	Body      []byte
	APIKey    string
}

// BackendDest struct holds properties of an influxdb backend destination.
type BackendDest struct {
	sync.RWMutex
	URL        string
	Queue      chan *Payload
	Registered time.Time
	RetryQueue chan *Payload
	Health     *health
}

type health struct {
	url                string
	timeout            int
	interval           int
	unhealthyThreshold int
	healthyThreshold   int
	healthStatus       bool
}

//HealthCheck function does the influxdb health checks.
func (b *BackendDest) HealthCheck() {
	var unhealthyCount, healthyCount int
	client := &http.Client{
		Timeout: (time.Duration(b.Health.timeout) * time.Second),
	}
	log.Infof("Starting health check for url %s", b.Health.url)
	interval := time.Tick(time.Duration(b.Health.interval) * time.Second)
	for {
		<-interval
		resp, err := client.Head(b.Health.url)

		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 204 {
				unhealthyCount = 0

				if !b.GetHealth() {
					if healthyCount >= b.Health.healthyThreshold {
						b.SetHealth(true)
					}
					healthyCount++
				}
				continue
			} else {
				healthyCount = 0
				if b.GetHealth() {
					if unhealthyCount >= b.Health.unhealthyThreshold {
						b.SetHealth(false)
					}
					unhealthyCount++
				}
				continue
			}
		} else {
			healthyCount = 0
			if b.GetHealth() {
				if unhealthyCount >= b.Health.unhealthyThreshold {
					b.SetHealth(false)
				}
				unhealthyCount++
			}
			continue
		}
	}
}

// GetHealth returns the health of a backend
func (b *BackendDest) GetHealth() bool {
	b.RLock()
	defer b.RUnlock()
	return b.Health.healthStatus
}

// SetHealth sets the health of a backend
func (b *BackendDest) SetHealth(s bool) {
	b.Lock()
	defer b.Unlock()
	b.Health.healthStatus = s
	if s {
		log.Infof("Backend: %s status is now healthy", b.URL)
	} else {
		log.Infof("Backend: %s status is now unhealthy", b.URL)
	}
}

// NewBackendDest initializes a *BackendDest.
func NewBackendDest(url string, outgoingQueueCap int, retryQueueCap int) *BackendDest {
	// To-Do: Make the healthcheck url configurable.
	healthCheckURL := url + "/ping"
	backend := &BackendDest{
		URL:        url,
		Queue:      make(chan *Payload, outgoingQueueCap),
		RetryQueue: make(chan *Payload, retryQueueCap),
		Health: &health{
			url:                healthCheckURL,
			timeout:            3,
			interval:           5,
			unhealthyThreshold: 2,
			healthyThreshold:   1,
			healthStatus:       false,
		},
	}
	return backend
}
