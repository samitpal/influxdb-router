### Metric flow with influxdb-router (Typical use case)
![alt text](images/influx-router.png "Metric flow with influx-router")

Installation
-------------------
##### To install the binary:
Go to the releases page and select the appropriate binary for your platform.

##### To build from source follow the steps below:

```sh
Install glide from http://glide.sh/.

$ go get github.com/samitpal/influxdb-router/...

$ glide install

Switch to the git tag, e.g if you want to build version 1.0

$ git checkout 1.0

$ make influxdb-router
```

### Sample config.toml
```
# You can configure multiple customers.
[[customers]]
  # name must be unique across customers
  name = "servicex"
  email = "user1@email.com"
  # api key. Should be a long random string. This is the value of the telegraf header (default header name is 'Service-API-Key')
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
  # and set auth-mode to 'from-config'. Additionally you need to enable authentication by setting the 'auth-enabled' option
  # to the in the [http] section of the InfluxDB config. 
  [customers.auth]
      # influxdb user. Yo need to grant privilege to user 'user1' to database 'telegraf1'
      username = "user1"
      # influxdb password for user user1
      password = "password1"
```

### Influxdb-router Usage
1. **Without InfluxDB authentication enabled.**
```
$./influxdb-router -config_file config.toml -api-listen-http-port 8080 -listen-http-port 8090
```

2. **With the InfluxDB creds in config.toml and with InfluxDB authentication enabled in the backends.**
```
$./influxdb-router -auth-enabled -auth-mode from-config -config_file config.toml -api-listen-http-port 8080 -listen-http-port 8090
```

3. **config.toml with the InfluxDB creds not in config but in env and with InfluxDB authentication enabled in the backends. You can set the creds in environment in the following format.**

```
$ export username_<name> = "user1"
$ export password_<name> = "password1"
<name> is the 'Name' attribute in the toml.


$ ./influxdb-router -auth-enabled -auth-mode from-env -config_file config.toml -api-listen-http-port 8090 -listen-http-port 8080
```

4. **With SSL**

```
$ ./influxdb-router -secure -config_file config.toml -api-listen-http-port 8080 -listen-https-port 8443 -ssl-server-cert <path to server cert> -ssl-server-key <path to server key>
```
5. **With SSL and with ssl client cert authentication enabled (telegraf needs to be configured with client certs)**

```
$ ./influxdb-router -secure -ssl-client-cert-auth -config_file config.toml -api-listen-http-port 8080 -listen-https-port 8443 -ssl-ca-server-cert <path to CA cert> -ssl-server-cert <path to server cert> -ssl-server-key <path to server key>
```

### Example client side config (telegraf configuration)
![alt text](images/telegraf.png "Telegraf configuration")
