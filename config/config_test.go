package config

import (
	"os"
	"testing"
)

//"reflect"

var customerAPIKey = []string{"7ba4e75a", "97dafb09"}
var customerInfluxDBName = []string{"telegraf1", "telegraf2"}
var customerOutgoingQueueCap = []int{4096, 5000}
var customerRetryQueueCap = []int{10, 4096}

var gotConf, _ = NewConfigs("./test_config.toml")

func TestNewConfigs(t *testing.T) {
	for i, c := range customerAPIKey {
		if c != *gotConf.Customers[i].APIKey {
			t.Errorf("APIKey does not match for customer%d. Got: %s, Expected: %s", i, *gotConf.Customers[i].APIKey, c)
		}
	}
	for i, c := range customerInfluxDBName {
		if c != *gotConf.Customers[i].InfluxDBName {
			t.Errorf("InfluxDBName does not match for customer%d. Got: %s, Expected: %s", i, *gotConf.Customers[i].InfluxDBName, c)
		}
	}
	for i, c := range customerOutgoingQueueCap {
		if c != *gotConf.Customers[i].OutgoingQueueCap {
			t.Errorf("OutgoingQueueCap does not match for customer%d. Got: %d, Expected: %d", i, *gotConf.Customers[i].OutgoingQueueCap, c)
		}
	}
	for i, c := range customerRetryQueueCap {
		if c != *gotConf.Customers[i].RetryQueueCap {
			t.Errorf("RetryQueueCap does not match for customer%d. Got: %d, Expected: %d", i, *gotConf.Customers[i].RetryQueueCap, c)
		}
	}
}

var gotBackEndDests = genBackends([]string{"http://127.0.0.1:8086", "http://1.2.3.4:8086"}, 400, 10)
var BackEndDestURLMap = map[string]string{
	"4bb0eb0ea4d5dfb784579db3b840a84b": "http://127.0.0.1:8086",
	"7a64eda9e403e9c434b0f3dbbca80e73": "http://1.2.3.4:8086",
}

func TestGenBackends(t *testing.T) {
	for k, v := range BackEndDestURLMap {
		_, ok := gotBackEndDests[k]
		if !ok {
			t.Errorf("Backend Hash:%s does not exist.", v)
		}
		if gotBackEndDests[k].URL != BackEndDestURLMap[k] {
			t.Errorf("Backend URL does not match, Got: %s, expected: %s", gotBackEndDests[k].URL, BackEndDestURLMap[k])
		}
	}
}

func TestNewAPIKeyMap(t *testing.T) {
	var expAPIKeyMap = map[string]APIKeyConfig{
		"7ba4e75a": APIKeyConfig{
			Dests:            genBackends([]string{"http://127.0.0.1:9086", "http://127.0.0.1:8086"}, 4096, 10),
			InfluxDBName:     "telegraf1",
			InfluxDBUserName: "user1",
			InfluxDBPassword: "password1",
			OutgoingQueueCap: 4096,
			RetryQueueCap:    10,
		},
		"97dafb09": APIKeyConfig{
			Dests:            genBackends([]string{"http://127.0.0.1:9086", "http://1.2.3.4:8086"}, 5000, 4096),
			InfluxDBName:     "telegraf2",
			InfluxDBUserName: "user2",
			InfluxDBPassword: "password2",
			OutgoingQueueCap: 5000,
			RetryQueueCap:    4096,
		},
	}
	// Auth from config
	var gotAPIKeyMap, _ = NewAPIKeyMap(gotConf.Customers, true, "from-config")
	for _, v := range []string{"7ba4e75a", "97dafb09"} {
		if expAPIKeyMap[v].InfluxDBName != gotAPIKeyMap[v].InfluxDBName {
			t.Errorf("Influxdbname does not match. Got: %s, Expected: %s", gotAPIKeyMap[v].InfluxDBName, expAPIKeyMap[v].InfluxDBName)
		}
		if expAPIKeyMap[v].InfluxDBUserName != gotAPIKeyMap[v].InfluxDBUserName {
			t.Errorf("Influxdb user does not match. Got: %s, Expected: %s", gotAPIKeyMap[v].InfluxDBUserName, expAPIKeyMap[v].InfluxDBUserName)
		}
		if expAPIKeyMap[v].InfluxDBPassword != gotAPIKeyMap[v].InfluxDBPassword {
			t.Errorf("Influxdb password does not match. Got: %s, Expected: %s", gotAPIKeyMap[v].InfluxDBPassword, expAPIKeyMap[v].InfluxDBPassword)
		}
		if expAPIKeyMap[v].OutgoingQueueCap != gotAPIKeyMap[v].OutgoingQueueCap {
			t.Errorf("Influxdb OutgoingQueueCap does not match. Got: %d, Expected: %d", gotAPIKeyMap[v].OutgoingQueueCap, expAPIKeyMap[v].OutgoingQueueCap)
		}
		if expAPIKeyMap[v].RetryQueueCap != gotAPIKeyMap[v].RetryQueueCap {
			t.Errorf("Influxdb RetryQueueCap does not match. Got: %d, Expected: %d", gotAPIKeyMap[v].RetryQueueCap, expAPIKeyMap[v].RetryQueueCap)
		}
	}

	// Auth from env
	os.Setenv("username_7ba4e75a", "user1-shell")
	os.Setenv("password_7ba4e75a", "password1-shell")

	os.Setenv("username_97dafb09", "user2-shell")
	os.Setenv("password_97dafb09", "password2-shell")

	expAPIKeyMap = map[string]APIKeyConfig{
		"7ba4e75a": APIKeyConfig{
			Dests:            genBackends([]string{"http://127.0.0.1:9086", "http://127.0.0.1:8086"}, 4096, 10),
			InfluxDBName:     "telegraf1",
			InfluxDBUserName: "user1-shell",
			InfluxDBPassword: "password1-shell",
			OutgoingQueueCap: 4096,
			RetryQueueCap:    10,
		},
		"97dafb09": APIKeyConfig{
			Dests:            genBackends([]string{"http://127.0.0.1:9086", "http://1.2.3.4:8086"}, 5000, 4096),
			InfluxDBName:     "telegraf2",
			InfluxDBUserName: "user2-shell",
			InfluxDBPassword: "password2-shell",
			OutgoingQueueCap: 5000,
			RetryQueueCap:    4096,
		},
	}

	gotAPIKeyMap, _ = NewAPIKeyMap(gotConf.Customers, true, "from-env")
	for _, v := range []string{"7ba4e75a", "97dafb09"} {
		if expAPIKeyMap[v].InfluxDBName != gotAPIKeyMap[v].InfluxDBName {
			t.Errorf("Influxdbname does not match. Got: %s, Expected: %s", gotAPIKeyMap[v].InfluxDBName, expAPIKeyMap[v].InfluxDBName)
		}
		if expAPIKeyMap[v].InfluxDBUserName != gotAPIKeyMap[v].InfluxDBUserName {
			t.Errorf("Influxdb user does not match. Got: %s, Expected: %s", gotAPIKeyMap[v].InfluxDBUserName, expAPIKeyMap[v].InfluxDBUserName)
		}
		if expAPIKeyMap[v].InfluxDBPassword != gotAPIKeyMap[v].InfluxDBPassword {
			t.Errorf("Influxdb password does not match. Got: %s, Expected: %s", gotAPIKeyMap[v].InfluxDBPassword, expAPIKeyMap[v].InfluxDBPassword)
		}
		if expAPIKeyMap[v].OutgoingQueueCap != gotAPIKeyMap[v].OutgoingQueueCap {
			t.Errorf("Influxdb OutgoingQueueCap does not match. Got: %d, Expected: %d", gotAPIKeyMap[v].OutgoingQueueCap, expAPIKeyMap[v].OutgoingQueueCap)
		}
		if expAPIKeyMap[v].RetryQueueCap != gotAPIKeyMap[v].RetryQueueCap {
			t.Errorf("Influxdb RetryQueueCap does not match. Got: %d, Expected: %d", gotAPIKeyMap[v].RetryQueueCap, expAPIKeyMap[v].RetryQueueCap)
		}
	}
}

func TestMask(t *testing.T) {
	s := "Hello World"
	mString := "*******orld"

	m := Mask(s, 4)
	if m != mString {
		t.Errorf("Returned masked String does not match. Got: %s, Expected: %s", m, mString)
	}
}
