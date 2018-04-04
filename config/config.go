// Package config handles the configurations etc.
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
package config

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/url"

	"github.com/samitpal/influxdb-router/backends"
	"github.com/BurntSushi/toml"
)

type errMandatoryField struct {
	field string
}

func (e *errMandatoryField) Error() string {
	return fmt.Sprintf("Mandatory field, %s is not defined", e.field)
}

func newErr(f string) *errMandatoryField {
	return &errMandatoryField{field: f}
}

//Config is the toml config
type Config struct {
	APIKey           *string   `toml:"api_key"`
	InfluxHosts      *[]string `toml:"influx_hosts"`
	InfluxDBName     *string   `toml:"influx_db_name"`
	OutgoingQueueCap *int      `toml:"outgoing_queue_cap"`
	Auth             *Authentication
	RetryQueueCap    *int `toml:"retry_queue_cap"`
}

// Authentication for influxdb.
type Authentication struct {
	UserName string
	Password string
}

//Configs is a slice of Config
type Configs struct {
	Customers []Config
}

// NewConfigs returns the routes derived from the toml config.
func NewConfigs(f string) (*Configs, error) {
	var config Configs
	if _, err := toml.DecodeFile(f, &config); err != nil {
		return nil, err
	}

	conf, err := config.checkConfig()

	if err != nil {
		return nil, err
	}

	return conf, nil
}

//LogConfig prints the config during startup
func (c *Configs) LogConfig() string {
	buff := bytes.NewBufferString("Starting up with the following configs\n")
	buff.WriteString("==========================\n")
	buff.WriteString("          CONFIGURATION           \n")
	for _, r := range c.Customers {
		buff.WriteString(fmt.Sprintf(
			`ApiKey = %s
InfluxHosts = %s
InfluxDB = %v
OutgoingQueueCap = %v
RetryQueueCap = %v
Auth = %v`,
			*r.APIKey,
			*r.InfluxHosts,
			*r.InfluxDBName,
			*r.OutgoingQueueCap,
			*r.RetryQueueCap,
			*r.Auth))
		buff.WriteString("\n-----------------------\n")
	}
	buff.WriteString("==========================\n")
	return buff.String()
}

// function checkConfig validates the toml config and also sets some defaults if any.
func (c *Configs) checkConfig() (*Configs, error) {

	mroutes := []Config{}
	for _, v := range c.Customers {
		if v.APIKey == nil {
			return nil, newErr("ApiKey")
		}
		if v.InfluxHosts == nil {
			return nil, newErr("InfluxHosts")
		}

		if v.InfluxDBName == nil {
			return nil, newErr("InfluxDBName")
		}

		if v.OutgoingQueueCap == nil {
			o := 4096
			v.OutgoingQueueCap = &o
		}
		if v.RetryQueueCap == nil {
			r := 4096
			v.RetryQueueCap = &r
		}
		if v.Auth == nil {
			user := ""
			password := ""
			v.Auth.UserName = user
			v.Auth.Password = password
		}
		mroutes = append(mroutes, v)
	}
	c.Customers = mroutes
	return c, nil
}

// APIKeyConfig contains the backend pool.
type APIKeyConfig struct {
	Dests            map[string]*backends.BackendDest
	InfluxDBName     string // database name in the backends
	InfluxDBUserName string // db user name
	InfluxDBPassword string // db password
	OutgoingQueueCap int    // Max in-memory outgoing queue size
	RetryQueueCap    int    // Max in-memory retry queue size
}

// APIKeyMap is a mapping of the customer api key to Apiconfig
type APIKeyMap map[string]APIKeyConfig

// NewAPIKeyMap returns APIKey map from the toml configs.
func NewAPIKeyMap(r []Config, authEnabled bool, authMode string) (APIKeyMap, error) {
	rp := make(APIKeyMap)
	for _, v := range r {
		s := APIKeyConfig{}
		s.InfluxDBName = *v.InfluxDBName
		s.OutgoingQueueCap = *v.OutgoingQueueCap
		s.RetryQueueCap = *v.RetryQueueCap

		err := checkURLS(*v.InfluxHosts)
		if err != nil {
			return nil, err
		}

		if authEnabled {
			authenticator, err := AuthMode(authMode, v)
			if err != nil {
				return nil, err
			}
			s.InfluxDBUserName, s.InfluxDBPassword = authenticator.Creds(*v.APIKey)
		}

		s.Dests = genBackends(*v.InfluxHosts, *v.OutgoingQueueCap, *v.RetryQueueCap)
		rp[*v.APIKey] = s
	}
	return rp, nil
}

func checkURLS(us []string) error {
	for _, u := range us {
		_, err := url.Parse(u)
		if err != nil {
			return err
		}
	}
	return nil
}

func genBackends(hosts []string, outgoingQueueCap int, retryQueueCap int) map[string]*backends.BackendDest {
	bs := make(map[string]*backends.BackendDest)
	for _, v := range hosts {
		h := md5.New()
		io.WriteString(h, v)
		hs := fmt.Sprintf("%x", h.Sum(nil))
		b := backends.NewBackendDest(v, outgoingQueueCap, retryQueueCap)
		bs[hs] = b
	}
	return bs
}
