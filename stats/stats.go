// Package stats exports various metrics
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
package stats

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/samitpal/influxdb-router/backends"
	"github.com/samitpal/influxdb-router/config"
	"github.com/samitpal/influxdb-router/logging"
)

var (
	log       = logging.For("stats")
	startTime = time.Now()
)

// Statsd struct holds the statsd data.
type Statsd struct {
	Interval int
	Conn     net.Conn
}

// ConnectStatsd connects to a statsd server and returns the connection.
func ConnectStatsd(s string, p string) (net.Conn, error) {
	conn, err := net.DialTimeout(p, s, time.Duration(3*time.Second))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// SendStatsdCounterMetric sends a counter metric to statsd
func (s *Statsd) SendStatsdCounterMetric(m string, v int) error {
	_, err := fmt.Fprintf(s.Conn, "%s:%d|c", m, v)
	if err != nil {
		return err
	}
	return nil
}

// SendStatsdMetrics sends a bunch of metrics already in statsd format to statsd.
func (s *Statsd) SendStatsdMetrics(metrics []string) {
	for _, m := range metrics {
		io.WriteString(s.Conn, m)
	}
}

// ExportMetrics exports metrics in statsd format
func ExportMetrics(s *Statsd, incomingQueueCap int, incomingQueue chan *backends.Payload, ac config.APIKeyMap) {
	interval := time.Tick(time.Duration(s.Interval) * time.Second)
	for {
		<-interval
		var metrics []string
		// collect metrics
		incomingQueue := fmt.Sprintf("influx_router.incoming_queue.current_size:%d|g", len(incomingQueue))
		incomingQueueCap := fmt.Sprintf("influx_router.incoming_queue.limit:%d|g", incomingQueueCap)
		metrics = append(metrics, incomingQueue, incomingQueueCap)

		for _, v := range ac {
			svcName := strings.Replace(v.Name, "-", "_", -1)
			for _, vd := range v.Dests {
				bURL := strings.TrimPrefix(strings.Replace(vd.URL, ".", "_", -1), "http://")
				//replace the colon chracter also.
				bURL = strings.Replace(bURL, ":", "_", -1)

				destQueueSize := fmt.Sprintf("influx_router.%s.outgoing_queue.%s.current_size:%d|g", svcName, bURL, len(vd.Queue))
				destQueueLimit := fmt.Sprintf("influx_router.%s.outgoing_queue.%s.limit:%d|g", svcName, bURL, v.OutgoingQueueCap)

				destRetryQueueSize := fmt.Sprintf("influx_router.%s.outgoing_retry_queue.%s.current_size:%d|g", svcName, bURL, len(vd.RetryQueue))
				destRetryQueueLimit := fmt.Sprintf("influx_router.%s.outgoing_retry_queue.%s.limit:%d|g", svcName, bURL, v.RetryQueueCap)
				metrics = append(metrics, destQueueSize, destQueueLimit, destRetryQueueSize, destRetryQueueLimit)

				vd.RLock()
				h := vd.GetHealth()
				vd.RUnlock()

				var ih int
				if h {
					ih = 1
				} else {
					ih = 0
				}
				backendHealth := fmt.Sprintf("influx_router.%s.backend_health.%s:%d|g", svcName, bURL, ih)
				metrics = append(metrics, backendHealth)
			}
		}

		// internal metrics
		uptimeMetric := uptimeStats()
		metrics = append(metrics, uptimeMetric...)

		cpuMetrics := internalCPUStats()
		metrics = append(metrics, cpuMetrics...)

		gcMetrics := internalGCStats()
		metrics = append(metrics, gcMetrics...)

		memMetrics := internalMemStats()
		metrics = append(metrics, memMetrics...)

		go s.SendStatsdMetrics(metrics)
	}
}

func gaugeFunc(m string, v interface{}) string {
	return fmt.Sprintf("influx_router.internal_stats.%s:%v|g", m, v)
}

func uptimeStats() []string {
	var m []string
	m = append(m, gaugeFunc("influx_router.uptime", int64(time.Since(startTime).Seconds())))
	return m
}

func internalCPUStats() []string {
	var m []string
	m = append(m, gaugeFunc("influx_router.cpu.goroutines", uint64(runtime.NumGoroutine())))
	m = append(m, gaugeFunc("influx_router.cpu.cgo_calls", uint64(runtime.NumCgoCall())))
	return m
}

func internalGCStats() []string {
	var m []string
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)
	m = append(m, gaugeFunc("influx_router.mem.gc.sys", ms.GCSys))
	m = append(m, gaugeFunc("influx_router.mem.gc.next", ms.NextGC))
	m = append(m, gaugeFunc("influx_router.mem.gc.last", ms.LastGC))
	m = append(m, gaugeFunc("influx_router.mem.gc.pause_total", ms.PauseTotalNs))
	m = append(m, gaugeFunc("influx_router.mem.gc.pause", ms.PauseNs[(ms.NumGC+255)%256]))
	m = append(m, gaugeFunc("influx_router.mem.gc.count", uint64(ms.NumGC)))
	return m
}

func internalMemStats() []string {
	var m []string
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)
	m = append(m, gaugeFunc("influx_router.mem.alloc", ms.Alloc))
	m = append(m, gaugeFunc("influx_router.mem.total", ms.TotalAlloc))
	m = append(m, gaugeFunc("influx_router.mem.sys", ms.Sys))
	m = append(m, gaugeFunc("influx_router.mem.lookups", ms.Lookups))
	m = append(m, gaugeFunc("influx_router.mem.malloc", ms.Mallocs))
	m = append(m, gaugeFunc("influx_router.mem.frees", ms.Frees))

	// Heap
	m = append(m, gaugeFunc("influx_router.mem.heap.alloc", ms.HeapAlloc))
	m = append(m, gaugeFunc("influx_router.mem.heap.sys", ms.HeapSys))
	m = append(m, gaugeFunc("influx_router.mem.heap.idle", ms.HeapIdle))
	m = append(m, gaugeFunc("influx_router.mem.heap.inuse", ms.HeapInuse))
	m = append(m, gaugeFunc("influx_router.mem.heap.released", ms.HeapReleased))
	m = append(m, gaugeFunc("influx_router.mem.heap.objects", ms.HeapObjects))

	// Stack
	m = append(m, gaugeFunc("influx_router.mem.stack.inuse", ms.StackInuse))
	m = append(m, gaugeFunc("influx_router.mem.stack.sys", ms.StackSys))
	m = append(m, gaugeFunc("influx_router.mem.stack.mspan_inuse", ms.MSpanInuse))
	m = append(m, gaugeFunc("influx_router.mem.stack.mspan_sys", ms.MSpanSys))
	m = append(m, gaugeFunc("influx_router.mem.stack.mcache_inuse", ms.MCacheInuse))
	m = append(m, gaugeFunc("influx_router.mem.stack.mcache_sys", ms.MCacheSys))

	m = append(m, gaugeFunc("influx_router.mem.othersys", ms.OtherSys))

	return m
}
