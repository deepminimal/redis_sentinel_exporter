package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRedisSentinelExporter(t *testing.T) {
	e := NewRedisSentinelExporter(&Options{
		Addr:             "127.0.0.1:8080",
		MetricsNamespace: "ns",
	})
	assert.NotNil(t, e)

	_, err := e.scrapeInfo()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "connection refused")
	}

	e.setMetrics(&SentinelInfo{
		Metrics: map[string]interface{}{
			"redis_version":  "1.2.3",
			"redis_build_id": "10",
			"redis_mode":     "abc",
			"expired_keys":   0.0,
		},
		Masters: []*Master{
			{
				Metrics: map[string]interface{}{
					"status":    0.0,
					"slaves":    0.0,
					"sentinels": 0.0,
					"name":      "abc",
					"address":   "127.0.0.1:8081",
				},
			},
		},
	})

	e.resetMetrics()
}

func TestExporter_IndexHandler(t *testing.T) {
	e := NewRedisSentinelExporter(&Options{
		Addr:        "127.0.0.1:8080",
		MetricsPath: "/foobar",
	})
	assert.NotNil(t, e)

	ts := httptest.NewServer(http.HandlerFunc(e.IndexHandler))
	defer ts.Close()
	resp, err := http.Get(ts.URL)
	if !assert.NoError(t, err) {
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(b), "Redis Sentinel Exporter")
	assert.Contains(t, string(b), e.o.MetricsPath)
}

func TestExporter_HealthyHandler(t *testing.T) {
	e := NewRedisSentinelExporter(&Options{
		Addr: "127.0.0.1:8080",
	})
	assert.NotNil(t, e)

	ts := httptest.NewServer(http.HandlerFunc(e.HealthyHandler))
	defer ts.Close()
	resp, err := http.Get(ts.URL)
	if !assert.NoError(t, err) {
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(b), "ok")
}
