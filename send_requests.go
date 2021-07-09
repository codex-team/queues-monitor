package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/valyala/fastjson"
)

// sendRequest sends HTTP request to provided address.
func sendRequest(ctx context.Context, method, addr string, reqBody string) ([]byte, error) {
	client := http.Client{}
	req := &http.Request{}
	var err error

	if method == "POST" {
		form := url.Values{}
		form.Set("message", reqBody)
		form.Set("parse_mode", "HTML")
		reqReader := strings.NewReader(form.Encode())
		req, err = http.NewRequestWithContext(ctx, method, addr, reqReader)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, addr, nil)
	}
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	var body []byte
	done := make(chan struct{})
	go func() {
		body, err = ioutil.ReadAll(resp.Body)
		close(done)
	}()

	select {
	case <-ctx.Done():
		<-done
		err = resp.Body.Close()
		if err == nil {
			err = ctx.Err()
		}
	case <-done:
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received code %d; body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// getMetricValues collects metric values.
func (m *Metric) getMetricValues(wg *sync.WaitGroup, ctx context.Context) error {
	defer wg.Done()
	var parser fastjson.Parser

	respBytes, err := sendRequest(ctx, "GET", p8sAddr+"/api/v1/query?query="+m.query, "")
	if err != nil {
		return err
	}
	data, err := parser.ParseBytes(respBytes)
	if err != nil {
		return err
	}

	var qVal, mtVal *fastjson.Value
	var queue, val []byte
	result := data.GetArray("data", "result")
	if len(m.values) == 0 {
		m.values = make([]metricValue, len(result))
	}
	cnt := 0
	for _, res := range result {
		qVal = res.Get("metric", "queue")
		if qVal == nil {
			continue
		}
		mtVal = res.Get("value", "1")
		if mtVal == nil {
			continue
		}

		queue = qVal.GetStringBytes()
		val = mtVal.GetStringBytes()

		if string(val) == "0" {
			continue
		}
		m.values[cnt] = metricValue{
			value: string(val),
			queue: string(queue),
		}
		cnt++
	}

	return nil
}

// notify sends notification with collected data.
func notify(ctx context.Context, report string) error {
	_, err := sendRequest(ctx, "POST", notifyAddr, report)
	if err != nil {
		return err
	}

	return nil
}
