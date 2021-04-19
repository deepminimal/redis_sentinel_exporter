package main

import (
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
}
