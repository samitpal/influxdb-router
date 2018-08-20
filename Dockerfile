FROM scratch
COPY influxdb-router /
ENTRYPOINT ["/influxdb-router"]
