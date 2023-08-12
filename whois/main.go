package main

import (
    "fmt"
    "flag"
    "time"
    "net"
    "errors"
    "net/http"
    "log"
    "io"
    "bufio"
    "whois/cache"
    "whois/server"
    "whois/whois_resp"
    "whois/postgres"
    "github.com/kpango/glg"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var serv server.Server

func process_query(query string) (string, error) {
    server.Queries.Inc()

    opts, query, err := parseWhoisQuery(query)
    if err != nil {
        return "", err
    }

    glg.Trace(opts["type"], query)

    whois_res := whois_resp.WhoisResponse{
        Header:serv.RGconf.Header,
        Source:serv.RGconf.Source,
    }

    if opts["type"] == "domain" {
        domain, ok := serv.Cache.Get(query)
        if !ok || time.Since(domain.Retrieved).Seconds() > 60 {
            domain, err = serv.Storage.GetDomain(query)
            if err != nil {
                if errors.Is(err, whois_resp.ObjectNotFound) {
                    return whois_res.EmptyResponse(), nil
                }
                return "", err
            }
            serv.Cache.Put(query, domain)
        }
        return whois_res.FormatDomain(domain), nil
    } else if opts["type"] == "registrar" {
        reg, err := serv.Storage.GetRegistrar(query)
        if err != nil {
            if errors.Is(err, whois_resp.ObjectNotFound) {
                return whois_res.EmptyResponse(), nil
            }
            return "", err
        }
        return whois_res.FormatRegistrar(reg), nil

    } 
    return whois_res.EmptyResponse(), nil
}

func handleTCPRequest(conn net.Conn) {
    defer conn.Close()
    reader := bufio.NewReader(conn)

    for {
        request, err := reader.ReadString('\n')
        if err != nil {
            if err != io.EOF {
                glg.Error("error: %v\n", err)
            }
            return
        }

        whois_response, err := process_query(request)
        if err != nil {
            glg.Error(err)
            whois_response = "error"
        }

        if _, err := conn.Write([]byte(whois_response)); err != nil {
            glg.Error(err)
        }
        break
    }
}

func start_tcp_server(whois_addr string) {
    listen, err := net.Listen("tcp", whois_addr)
    if err != nil {
        glg.Fatal(err)
    }
    defer listen.Close()

    for {
        conn, err := listen.Accept()
        if err != nil {
            glg.Error(err)
            continue
        }
        go handleTCPRequest(conn)
    }
}

func main() {
    config_file := flag.String("config", "server.conf", "filename with config")
    port := flag.Uint("port", 0, "port")

    flag.Parse()

    serv.RGconf.LoadConfig(*config_file)
    if *port > 0 {
        serv.RGconf.HTTPConf.Port = *port
    }
    var err error
    serv.Storage, err = postgres.NewWhoisStorage(serv.RGconf.DBconf)
    if err != nil {
        log.Fatal(err)
    }

    serv.Cache = cache.NewLRUCache[string, whois_resp.Domain](100)

    whois_addr := fmt.Sprintf("%s:%v", serv.RGconf.HTTPConf.Host, serv.RGconf.HTTPConf.Port)

    go start_tcp_server(whois_addr)

    /* metrics */
    host_addr := fmt.Sprintf("%s:%v", serv.RGconf.HTTPConf.Host, 8083) //serv.RGconf.HTTPConf.Port)

    httpserver := &http.Server{
        Addr: host_addr,
    }

    http.Handle("/metrics", promhttp.Handler())

    fmt.Println("server is running at", host_addr)

    httpserver.ListenAndServe()
}
