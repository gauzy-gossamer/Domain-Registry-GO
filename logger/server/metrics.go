package server

import (
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
        Queries = promauto.NewCounter(prometheus.CounterOpts{
                Name: "logger_queries_total",
                Help: "The total number of processed queries",
        })
)
