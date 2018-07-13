# To build a new binary just run
# $ make clean
# $ make influxdb-router

INFLUXDB-ROUTER_PKG := github.com/samitpal/influxdb-router

# Change the version with the git tag
INFLUXDB-ROUTER_VERSION := 1.0

BUILDTIME := $(shell date +%FT%T%z)
GOPATH := $(GOPATH)

all: influxdb-router

influxdb-router:
	go install  -ldflags="-X main.Version=${INFLUXDB-ROUTER_VERSION} -X main.BuildTime=${BUILDTIME}" ${INFLUXDB-ROUTER_PKG}

.PHONY: clean

clean:
	rm -f $(GOPATH)/bin/influxdb-router