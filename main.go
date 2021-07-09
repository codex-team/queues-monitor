package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// metricValue is a struct containing values and queue names respectively.
type metricValue struct {
	value string
	queue string
}

// Metric contains information about a metric.
type Metric struct {
	// query is a PromQL query.
	query string
	// values contains values of this metric for different queues.
	values []metricValue
	// description is title of notification.
	description string
}

var (
	// interval is default scrape interval.
	interval time.Duration
	// p8sAddr is Prometheus address.
	p8sAddr string
	// notifyAddr is notifications endpoint
	notifyAddr string
	// grafanaLink is link to Grafana dashboard
	grafanaLink string
	// server name.
	server string
)

// init initializes command line flags.
func init() {
	flag.DurationVar(&interval, "t", 24*time.Hour, "Scrape interval")
	flag.StringVar(&p8sAddr, "p8s", "http://localhost:9090", "Prometheus address")
	flag.StringVar(&notifyAddr, "notify", "", "Notifications endpoint")
	flag.StringVar(&grafanaLink, "grafana", "", "Link to Grafana dashboard")
	flag.StringVar(&server, "server", "prod", "Server name")
}

func main() {
	log.Printf("starting")
	flag.Parse()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	ticker := time.NewTicker(interval)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if grafanaLink != "" {
		grafanaLink = fmt.Sprintf("<a href='%s'>See Details</a>", grafanaLink)
	}

	metrics := &[]Metric{
		{
			query:       "rabbitmq_queue_messages",
			values:      []metricValue{},
			description: "Queues on Hawk (" + server + ") 🌀",
		},
	}

	for {
		select {
		case <-done:
			signal.Stop(done)
			log.Printf("stopping")
			return
		case <-ticker.C:
			err := collectData(ctx, metrics)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// collectData gets data from Prometheus and reports it.
func collectData(ctx context.Context, metrics *[]Metric) error {
	errs := make(chan error)
	wg := sync.WaitGroup{}
	for i := range *metrics {
		wg.Add(1)
		go func(mt *Metric, ctx context.Context) {
			err := mt.getMetricValues(&wg, ctx)
			if err != nil {
				errs <- err
				return
			}
			errs <- notify(ctx, mt.toString())
			mt.values = mt.values[:0]
		}(&(*metrics)[i], ctx)
	}
	wg.Wait()

	if len(errs) != 0 {
		err := <-errs
		if err != nil {
			return <-errs
		}
	}

	return nil
}

// toString converts Metric data to readable representation.
func (mt *Metric) toString() (strVal string) {
	strVal = mt.description + "  \n\n"
	sort.Slice(mt.values[:], func(i, j int) bool {
		first, _ := strconv.Atoi(mt.values[i].value)
		second, _ := strconv.Atoi(mt.values[j].value)
		return first > second
	})
	for _, data := range mt.values {
		if (data.value == "") || (data.queue == "") {
			continue
		}
		strVal += fmt.Sprintf("%s: %s \n", data.queue, data.value)
	}
	if grafanaLink != "" {
		strVal += fmt.Sprintf("\n%s", grafanaLink)
	}

	return
}
