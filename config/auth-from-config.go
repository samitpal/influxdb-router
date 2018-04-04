//Package config for managing configurations.
// The MIT License (MIT)
//
// Copyright (c) 2017 Samit Pal
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

//authFromConfig implements Authenticator and enables auth based on the toml config
/*
Example auth config in toml follows
[[customers]]
  name = "servicex"
  email = "user1@email.com"
  api_key = "7ba4e75a-30a4-476f-8b95-cf26b7f4b70a"
  influx_db_name = "telegraf1"
  outgoing_queue_cap = 7000
  retry_queue_cap = 10
  influx_hosts = ["http://127.0.0.1:9086", "http://127.0.0.1:8086"]
  # The auth section needs to come at the end. This should be populated only if you enabled auth in influx-router
  # and set auth-mode to 'from-config'
  [customers.auth]
      username = "user1"
      password = "password1"

*/
type authFromConfig struct {
	conf Config
}

//newAuthFromConfig provides auth from config
func newAuthFromConfig(c Config) authFromConfig {
	return authFromConfig{conf: c}
}

func (a authFromConfig) Creds(apiKey string) (u string, p string) {
	return a.conf.Auth.UserName, a.conf.Auth.Password
}
