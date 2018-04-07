# NOT PRODUCTION READY YET.
### Metric flow with influxdb-router
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

### Influxdb-router Usage
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
### Sample config.toml
```
[[customers]]
  name = "servicex"
  email = "user1@email.com"
  # api key should not have !, *, or - character. This is the value of the telegraf header (default header name is 'Service-API-Key')
  api_key = "7ba4e75a"
  # Name of the influxdb database where the metrics for this customer should be stored.
  influx_db_name = "telegraf1"
  # Max number of batches that will be kept in-memory for each of the 'influx_hosts'. Beyind that batches with be dropped
  outgoing_queue_cap = 7000
  # Influxdb-routed maintains a retry queue for batches that it fails to send to InfluxDB backends.
  # retry_queue_cap is the max number of batches that can be kept in the retry queue (in-memory).
  retry_queue_cap = 10
  # list of InfluxDB hosts.
  influx_hosts = ["http://127.0.0.1:9086", "http://127.0.0.1:8086"]
  # The auth section needs to come at the end. This should be populated only if you enabled auth in influx-router
  # and set auth-mode to 'from-config'
  [customers.auth]
      # influxdb user
      username = "user1"
      # influxdb password for user user1
      password = "password1"
```

### Example telegraf configuration
![alt text](images/telegraf.png "Telegraf configuration")

