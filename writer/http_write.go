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
	"bytes"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/samitpal/influxdb-router/backends"
	"github.com/samitpal/influxdb-router/writer/client"
)

// random generates a random number between min and max
func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

//InfluxWriter reads from dest queue and writes to the dest.
func InfluxWriter(b *backends.BackendDest, db string, user string, password string) {
	c := client.HTTPConfig{URL: b.URL, ContentEncoding: "gzip", Username: user, Password: password}
	p := client.WriteParams{Database: db}
	httpClient, err := client.NewHTTP(c, p)
	if err != nil {
		log.Info("Error in creating http client. Returning out of the go-routine.")
		return
	}

	// Keep popping messages from the channel and write the same to influxdb in a for loop
	for message := range b.Queue {
		body := ioutil.NopCloser(bytes.NewBuffer(message.Body))

		if b.GetHealth() {
			go httpClient.WriteInflux(body, db, message.MessageID, b.URL)
		} else {
			log.Infof("Backend:%s is unhealthy. Can't push metrics.", b.URL)
			select {
			case b.RetryQueue <- message:
			default:
				log.Info("Retry queue for backend:%s might be at capacity.", b.URL)
			}
		}
	}
}

// RetryQueueHandler retries messages from the retry queue.
func RetryQueueHandler(b *backends.BackendDest, db string, user string, password string) {
	c := client.HTTPConfig{URL: b.URL, ContentEncoding: "gzip", Username: user, Password: password}
	p := client.WriteParams{Database: db}
	httpClient, err := client.NewHTTP(c, p)
	if err != nil {
		log.Info("Error in creating http client. Returning out of the goroutine.")
		return
	}

	for {
		if len(b.RetryQueue) > 0 {
			if b.GetHealth() {
				select {
				case message := <-b.RetryQueue:
					body := ioutil.NopCloser(bytes.NewBuffer(message.Body))
					go httpClient.WriteInflux(body, db, message.MessageID, b.URL)
				}
			} else {
				time.Sleep(time.Duration(random(1, 3)) * time.Second)
			}
		} else {
			time.Sleep(time.Duration(random(1, 3)) * time.Second)
		}
	}
}
