# NOT PRODUCTION READY YET.
##### Metric flow with influxdb-router
![alt text](images/influx-router.png "Metric flow with influx-router")

Installation
-------------------
##### To build from source follow the steps below: 

```sh
Install glide from http://glide.sh/.

$ go get -u github.com/samitpal/influxdb-router/...

$ glide install

$ go install
```

##### Influxdb-router Usage
```
Usage of influxdb-router:
  -api-key-header-name string
    	Name of the API key header. [INFLUX_API_KEY_HEADER_NAME] (default "Service-API-Key")
  -api-listen-addr string
    	Influx proxy api listen address [INFLUX_API_LISTEN_ADDR] (default "127.0.0.1")
  -api-listen-http-port string
    	Influx proxy api listen port [INFLUX_API_LISTEN_HTTP_PORT] (default "8080")
  -auth-enabled
    	Whether to enable authentication when communicating with InfluxDB [INFLUX_AUTH_ENABLED]
  -auth-mode string
    	Can be either 'from-config or 'from-env' presently. 'auth-enabled' flag needs to be turned on. [INFLUX_AUTH_MODE] (default "from-config")
  -config_file string
    	Configuration options. [INFLUX_CONFIG_FILE] (default "config.toml")
  -incoming-queue-cap int
    	In-flight incoming message queue capacity [INFLUX_INCOMING_QUEUE_CAP] (default 500000)
  -listen-addr string
    	Influx proxy listen address [INFLUX_LISTEN_ADDR] (default "0.0.0.0")
  -listen-http-port string
    	Influx proxy listen port (http) [INFLUX_LISTEN_HTTP_PORT] (default "80")
  -listen-https-port string
    	Influx proxy listen port (https) [INFLUX_LISTEN_HTTPS_PORT] (default "443")
  -ssl-cert string
    	TLS Certificate [INFLUX_SSL_CERT]
  -ssl-key string
    	TLS Key [INFLUX_SSL_KEY]
  -stats-interval int
    	Interval in seconds for sending statsd metrics. [INFLUX_STATS_INTERVAL] (default 30)
  -statsd-server string
    	statsd server:port for sending metrics [INFLUX_STATSD_SERVER] (default "localhost:8125")
  -wait-before-shutdown int
    	Number of seconds to wait before the process shuts down. Health checks will be failed during this time. [INFLUX_WAIT_BEFORE_SHUTDOWN] (default 1)
```

### Example telegraf configuration
![alt text](images/telegraf.png "Telegraf configuration")

