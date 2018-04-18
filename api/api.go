// Package api provides code to expose the running configs.
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
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/samitpal/influxdb-router/config"
	"github.com/samitpal/influxdb-router/logging"
)

var log = logging.For("api")

// HTTPListenerConfig holds configs for the http daemon
type HTTPListenerConfig struct {
	Addr     string
	Port     string
	TomlConf config.Configs
	APIConf  config.APIKeyMap
}

// httpHandlers has all the routes defined.
func httpHandlers(h *http.ServeMux, conf *HTTPListenerConfig) *http.ServeMux {
	h.Handle("/api/v1/config", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { displayConfig(w, conf) }))
	return h
}

// HTTPListener exposes the http listener for api access (read only as of now).
func HTTPListener(conf *HTTPListenerConfig) {
	h := http.NewServeMux()
	h = httpHandlers(h, conf)

	go func() {
		log.Infof("InfluxDB Router http rest API service listening on %s:%s\n", conf.Addr, conf.Port)
		err := http.ListenAndServe(conf.Addr+":"+conf.Port, h)
		if err != nil {
			log.Fatalf("ListenAndServe: %s\n", err)
		}
	}()
}

func displayConfig(w http.ResponseWriter, conf *HTTPListenerConfig) {
	data, err := json.Marshal(conf.TomlConf)
	if err != nil {
		log.Errorf("Error while json marshal: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error while decoding struct to json")
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(data))
}
