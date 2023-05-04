package main

import (
    "time"
    "fmt"
    "flag"
    "net"
    "net/http"
    "log"
    "io"
    "io/ioutil"
    "strings"
    "bufio"
    "whois/server"
    "github.com/jackc/pgx/v5"
    "github.com/kpango/glg"
)

var serv server.Server

func process_query(query string) (string, error) {
    domain_name := strings.TrimSpace(query)

    whois_resp := WhoisResponse{Header:serv.RGconf.Header}

    domain, err := getDomain(domain_name)
    if err != nil {
        if err == pgx.ErrNoRows {
            return whois_resp.EmptyResponse(), nil
        }
        return "", err
    }
    return whois_resp.FormatDomain(domain), nil
}

func handle_root(w http.ResponseWriter, req *http.Request) {
    if req.URL.Path != "/" {
        http.NotFound(w, req)
        return
    }

    query, err := ioutil.ReadAll(req.Body)
    if err != nil {
        io.WriteString(w, "error")
        return
    }

    glg.Info(query)

    start := time.Now()

    whois_response, err := process_query(string(query))
    if err != nil {
        glg.Error(err)
        io.WriteString(w, "error")
        return
    }

    elapsed := time.Since(start)

    glg.Trace("exec took ", elapsed)

    io.WriteString(w, whois_response)
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

        /* TODO validate input */

        whois_response, err := process_query(request)
        if err != nil {
            glg.Error(err)
            whois_response = "error"
        }

        if _, err := conn.Write([]byte(whois_response)); err != nil {
            glg.Error(err)
        }
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
    serv.Pool, err = server.CreatePool(&serv.RGconf.DBconf)
    if err != nil {
        log.Fatal(err)
    }

    whois_addr := fmt.Sprintf("%s:%v", serv.RGconf.HTTPConf.Host, serv.RGconf.HTTPConf.Port)

    go start_tcp_server(whois_addr)

    /* metrics */
    host_addr := fmt.Sprintf("%s:%v", serv.RGconf.HTTPConf.Host, 8083) //serv.RGconf.HTTPConf.Port)

    httpserver := &http.Server{
        Addr: host_addr,
    }

    http.HandleFunc("/", handle_root)

    fmt.Println("server is running at", host_addr)

    httpserver.ListenAndServe()
}
