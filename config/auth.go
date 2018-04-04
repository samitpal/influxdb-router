//Package config for managing configurations.
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
	"fmt"

	"github.com/samitpal/influxdb-router/logging"
)

var log = logging.For("auth")

//Authenticator is the interface that needs to be implemented for providing auth service.
type Authenticator interface {
	// Creds returns username, password given and api key string
	Creds(string) (user string, password string)
}

//AuthMode returns the appropriate auth mode
func AuthMode(m string, c Config) (Authenticator, error) {
	if m == "from-config" {
		return newAuthFromConfig(c), nil
	} else if m == "from-env" {
		return newAuthFromENV(), nil
	}
	return nil, fmt.Errorf("Auth Mode: %v invalid", m)
}
