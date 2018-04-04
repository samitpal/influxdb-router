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

import (
	"fmt"
	"os"
)

//AuthFromENV implements Authenticator and enables environment variables based auth
/*
The environment variable for the username should be of the format username_<api key>
The environment variable for the password should be of the format password_<api key>
Example
export username_abcd = "user1"
export password_abcd = "password1"
*/
type authFromENV struct{}

//newAuthFromENV provides auth from environment
func newAuthFromENV() authFromENV {
	return authFromENV{}
}

func (a authFromENV) Creds(apiKey string) (string, string) {
	u := os.Getenv(fmt.Sprintf("username_%s", apiKey))
	p := os.Getenv(fmt.Sprintf("password_%s", apiKey))
	return u, p
}
