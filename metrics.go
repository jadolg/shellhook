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
)
