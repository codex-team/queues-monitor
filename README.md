# queues-monitor
Tool for monitoring RabbitMQ queues with Prometheus

### Parameters
```shell
  -grafana string
        Link to Grafana dashboard
  -notify string
        Notifications endpoint
  -p8s string
        Prometheus address (default "http://localhost:9090")
  -server string
        Server name (default "prod")
  -t duration
        Scrape interval (default 24h0m0s)
```

### Build and run
```shell
$ make
$ ./notify [parameters]
```
