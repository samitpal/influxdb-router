// Package client provides code to build influxdb clients
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
// Lot of the code here is from https://github.com/influxdata/telegraf/blob/master/plugins/outputs/influxdb/client/http.go
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/samitpal/influxdb-router/logging"
)

var (
	defaultRequestTimeout      = time.Second * 5
	defaultMaxIdleConnsPerHost = 10
	log                        = logging.For("client")
)

//NewHTTP returns an httpClient.
func NewHTTP(config HTTPConfig, defaultWP WriteParams) (*httpClient, error) {
	// validate required parameters:
	if len(config.URL) == 0 {
		return nil, fmt.Errorf("config.URL is required to create an HTTP client")
	}
	if len(defaultWP.Database) == 0 {
		return nil, fmt.Errorf("A default database is required to create an HTTP client")
	}

	// set defaults:
	if config.Timeout == 0 {
		config.Timeout = defaultRequestTimeout
	}
	if config.MaxIdleConnsPerHost == 0 {
		config.MaxIdleConnsPerHost = defaultMaxIdleConnsPerHost
	}
	// parse URL:
	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, fmt.Errorf("error parsing config.URL: %s", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("config.URL scheme must be http(s), got %s", u.Scheme)
	}

	var transport http.Transport
	if len(config.HTTPProxy) > 0 {
		proxyURL, err := url.Parse(config.HTTPProxy)
		if err != nil {
			return nil, fmt.Errorf("error parsing config.HTTPProxy: %s", err)
		}

		transport = http.Transport{
			Proxy:               http.ProxyURL(proxyURL),
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		}
	} else {
		transport = http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		}
	}

	return &httpClient{
		writeURL: writeURL(u, defaultWP),
		config:   config,
		url:      u,
		client: &http.Client{
			Timeout:   config.Timeout,
			Transport: &transport,
		},
	}, nil
}

//WriteParams sets up the params sent to the http api call
type WriteParams struct {
	Database        string
	RetentionPolicy string
	Consistency     string
}

//HTTPHeaders to append to HTTP requests.
type HTTPHeaders map[string]string

//HTTPConfig for api call to the backend.
type HTTPConfig struct {
	// URL should be of the form "http://host:port" (REQUIRED)
	URL string

	// UserAgent sets the User-Agent header.
	UserAgent string

	// Timeout specifies a time limit for requests made by this
	// Client. The timeout includes connection time, any
	// redirects, and reading the response body. The timer remains
	// running after Get, Head, Post, or Do return and will
	// interrupt reading of the Response.Body.
	//
	// A Timeout of zero means no timeout.
	Timeout time.Duration

	// Username is the basic auth username for the server.
	Username string
	// Password is the basic auth password for the server.
	Password string

	// Proxy URL should be of the form "http://host:port"
	HTTPProxy string

	// HTTP headers to append to HTTP requests.
	HTTPHeaders HTTPHeaders

	// The content encoding mechanism to use for each request.
	ContentEncoding string

	//Max idle connections per hosts
	MaxIdleConnsPerHost int
}

// Response represents a list of statement results.
type Response struct {
	// ignore Results:
	Results []interface{} `json:"-"`
	Err     string        `json:"error,omitempty"`
}

// Error returns the first error from any statement.
// Returns nil if no errors occurred on any statements.
func (r *Response) Error() error {
	if r.Err != "" {
		return fmt.Errorf(r.Err)
	}
	return nil
}

type httpClient struct {
	writeURL string
	config   HTTPConfig
	client   *http.Client
	url      *url.URL
}

func (c *httpClient) WriteInflux(r io.Reader, db string, id string, url string) {
	if e := c.WriteStream(r); e != nil {
		// If the database was not found
		if strings.Contains(e.Error(), "database not found") {
			log.Errorf("E! Error: Database %s not found\n", db)
			return
		}

		if strings.Contains(e.Error(), "field type conflict") {
			log.Errorf("E! Field type conflict, dropping conflicted points: %s", e)
			return
		}

		if strings.Contains(e.Error(), "points beyond retention policy") {
			log.Errorf("W! Points beyond retention policy: %s", e)
			return
		}

		if strings.Contains(e.Error(), "unable to parse") {
			log.Errorf("E! Parse error; dropping points: %s", e)
			return
		}

		if strings.Contains(e.Error(), "hinted handoff queue not empty") {
			return
		}

		// Log any other write failure
		log.Errorf("E! InfluxDB Output Error: %v", e)
		return
	}
	log.Infof("Successfully sent message-id: %s, db: %s, backend: %s", id, db, url)
}

func (c *httpClient) WriteStream(r io.Reader) error {
	req, err := c.makeWriteRequest(r, c.writeURL)
	if err != nil {
		return err
	}
	return c.doRequest(req, http.StatusNoContent)
}

func (c *httpClient) doRequest(req *http.Request, expectedCode int) error {
	resp, err := c.client.Do(req)
	if err != nil {
		log.Info("http req failed.")
		return err
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	// If it's a "no content" response, then release and return nil
	if code == http.StatusNoContent {
		return nil
	}

	// not a "no content" response, so parse the result:
	var response Response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Fatal error reading body: %s", err)
	}

	decErr := json.Unmarshal(body, &response)

	// If we got a JSON decode error, send that back
	if decErr != nil {
		err = fmt.Errorf("Unable to decode json: received status code %d err: %s", code, decErr)
	}
	// Unexpected response code OR error in JSON response body overrides
	// a JSON decode error:
	if code != expectedCode || response.Error() != nil {
		err = fmt.Errorf("Response Error: Status Code [%d], expected [%d], [%v]",
			code, expectedCode, response.Error())
	}

	return err
}

func (c *httpClient) makeWriteRequest(
	body io.Reader,
	writeURL string,
) (*http.Request, error) {
	req, err := c.makeRequest(writeURL, body)
	if err != nil {
		return nil, err
	}

	if c.config.ContentEncoding == "gzip" {
		req.Header.Set("Content-Encoding", "gzip")
	}

	return req, nil
}

func (c *httpClient) makeRequest(uri string, body io.Reader) (*http.Request, error) {
	var req *http.Request
	var err error

	req, err = http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}

	for header, value := range c.config.HTTPHeaders {
		req.Header.Set(header, value)
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("User-Agent", c.config.UserAgent)
	if c.config.Username != "" && c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}
	return req, nil
}

func (c *httpClient) Close() error {
	// Nothing to do.
	return nil
}

func writeURL(u *url.URL, wp WriteParams) string {
	params := url.Values{}
	params.Set("db", wp.Database)
	if wp.RetentionPolicy != "" {
		params.Set("rp", wp.RetentionPolicy)
	}
	if wp.Consistency != "one" && wp.Consistency != "" {
		params.Set("consistency", wp.Consistency)
	}

	u.RawQuery = params.Encode()
	p := u.Path
	u.Path = path.Join(p, "write")
	s := u.String()
	u.Path = p
	return s
}
