# queues-monitor
Tool for monitoring RabbitMQ queues with Prometheus

### Features
* Gets current number of messages in RabbitMQ queues from Prometheus
* Sends collected data in readable representation to provided notification endpoint
* Can attach link to related Grafana dashboard in notification

### Parameters
```shell
  -grafana string
        Link to Grafana dashboard
  -notify string
        Notifications endpoint
  -p8s string
        Prometheus address (default "http://localhost:9090")
  -server string
        Server name (default "Hawk (prod)")
  -t duration
        Scrape interval (default 24h0m0s)
```

### Build and run
```shell
$ make
$ ./notify [parameters]
```
