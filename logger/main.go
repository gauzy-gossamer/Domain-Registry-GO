package main

import (
    "log"
    "flag"
    "net/http"
    "logger/server"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var serv server.Server

func main() {
    config_file := flag.String("config", "server.conf", "filename with config")
    port := flag.Uint("port", 0, "port")

    flag.Parse()

    serv.RGconf.LoadConfig(*config_file, &serv)
    if *port > 0 {
        serv.RGconf.GrpcPort = *port
    }
    host_addr := serv.RGconf.MetricsPath

    go StartgRPCServer(&serv)

    httpserver := &http.Server{
        Addr: host_addr,
    }

    http.Handle("/metrics", promhttp.Handler())

    log.Println("server is running at", host_addr)

    if err := httpserver.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
