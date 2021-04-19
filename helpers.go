package main

import (
	"os"
	"strconv"
	"time"
)

func GetStringEnv(name string, def string) string {
	if val, ok := os.LookupEnv(name); ok {
		return val
	}
	return def
}

func GetBoolEnv(name string, def bool) bool {
	if val, ok := os.LookupEnv(name); ok {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return def
}

func GetDurationEnv(name string, def time.Duration) time.Duration {
	if val, ok := os.LookupEnv(name); ok {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return def
}
