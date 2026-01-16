package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	errorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shellhook_errors_total",
		Help: "The total number of errors found",
	})
	execsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "shellhook_execs_total",
		Help: "The total number of calls to exec",
	})
	execDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "shellhook_exec_duration_seconds",
		Help:    "Script execution duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"script"})
)
