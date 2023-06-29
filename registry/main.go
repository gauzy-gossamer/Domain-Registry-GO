package main

import (
    "fmt"
    "flag"
    "crypto/tls"
    "net/http"
    _ "net/http/pprof"
    "log"
    "io"

    "registry/server"
    "registry/epp"
    . "registry/epp/eppcom"
    "registry/xml"
    "registry/regrpc"
    "registry/maintenance"
)

var serv server.Server

func handle_root(w http.ResponseWriter, req *http.Request) {
    eppc := epp.NewEPPContext(&serv)
    logger := eppc.GetLogger()

    XML, err := io.ReadAll(req.Body)
    if err != nil {
        epp_res := EPPResult{RetCode:2500}
        _, err = io.WriteString(w, xml.GenerateResponse(&epp_res, "", ""))
        if err != nil {
            logger.Error(err)
        }
        return
    }

    response := server.HandleRequest(&serv, eppc, w, req, string(XML))

    _, err = io.WriteString(w, response)
    if err != nil {
        logger.Error(err)
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

    serv.XmlParser.ReadSchema(serv.RGconf.SchemaPath)
    var err error
    serv.Pool, err = server.CreatePool(&serv.RGconf.DBconf)
    if err != nil {
        log.Fatal(err)
    }
    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        log.Fatal(err)
    }
    serv.Sessions.MaxRegistrarSessions = serv.RGconf.MaxRegistrarSessions
    serv.Sessions.MaxQueriesPerMinute = serv.RGconf.MaxQueriesPerMinute
    serv.Sessions.SessionTimeoutSec = serv.RGconf.SessionTimeout
    serv.Sessions.InitSessions(dbconn)
    dbconn.Close()

    go regrpc.StartgRPCServer(&serv)

    go maintenance.Maintenance(&serv)

    host_addr := fmt.Sprintf("%s:%v", serv.RGconf.HTTPConf.Host, serv.RGconf.HTTPConf.Port)

    httpserver := &http.Server{
        TLSConfig: &tls.Config{
            ClientAuth: tls.RequireAnyClientCert,
        },
        Addr: host_addr,
    }

    http.HandleFunc("/", handle_root)

    fmt.Println("server is running at", host_addr)

    if serv.RGconf.HTTPConf.UseProxy {
        if err := httpserver.ListenAndServe(); err != nil {
            log.Fatal(err)
        }
    } else {
        if err := httpserver.ListenAndServeTLS(serv.RGconf.HTTPConf.CertFile, serv.RGconf.HTTPConf.KeyFile); err != nil {
            log.Fatal(err)
        }
    }
}
