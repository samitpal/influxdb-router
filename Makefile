# To build a new binary just run
# $ make clean
# $ make influxdb-router

INFLUXDB-ROUTER_PKG := github.com/samitpal/influxdb-router

# Change the version with the git tag
INFLUXDB-ROUTER_VERSION := 0.1.3

BUILDTIME := $(shell date +%FT%T%z)
GIT_COMMIT_ID := $(shell git rev-parse HEAD)
GOPATH := $(GOPATH)

all: influxdb-router

influxdb-router:
	go install  -ldflags="-X main.version=${INFLUXDB-ROUTER_VERSION} -X main.date=${BUILDTIME} -X main.commit=${GIT_COMMIT_ID}" ${INFLUXDB-ROUTER_PKG}

.PHONY: clean

clean:
	rm -f $(GOPATH)/bin/influxdb-router
