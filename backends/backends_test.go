package backends

import (
	"reflect"
	"testing"
)

var url = "http://localhost:8086"
var outgoingQueueCap = 1000
var retryQueueCap = 10
var expBackend = BackendDest{
	URL:        "http://localhost:8086",
	Queue:      make(chan *Payload, 1000),
	RetryQueue: make(chan *Payload, 10),
	Health: &health{
		url:                "http://localhost:8086/ping",
		timeout:            3,
		interval:           5,
		unhealthyThreshold: 2,
		healthyThreshold:   1,
		healthStatus:       false,
	},
}

var gotBackend = NewBackendDest(url, outgoingQueueCap, retryQueueCap)

func TestBackendDest(t *testing.T) {
	if expBackend.URL != gotBackend.URL {
		t.Errorf("Backend URL does not match, Got: %v, Expected: %v", gotBackend.URL, expBackend.URL)
	}
	if !reflect.DeepEqual(expBackend.Health, gotBackend.Health) {
		t.Errorf("Backend health does not match, Got: %v, Expected: %v", gotBackend.Health, expBackend.Health)
	}
}

func TestGetHealth(t *testing.T) {
	if gotBackend.GetHealth() {
		t.Error("Health should be false")
	}
}

func TestSetHealth(t *testing.T) {
	gotBackend.SetHealth(true)
	if !gotBackend.GetHealth() {
		t.Error("Health should be true")
	}

	gotBackend.SetHealth(false)
	if gotBackend.GetHealth() {
		t.Error("Health should be false")
	}
}
