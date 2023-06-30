package main

import (
    "time"
    "fmt"
    "flag"
    "crypto/tls"
    "net/http"
    _ "net/http/pprof"
    "log"
    "io"
    "strconv"
    "context"

    "registry/server"
    "registry/epp"
    . "registry/epp/eppcom"
    "registry/xml"
    "registry/regrpc"
    "registry/maintenance"

    "github.com/kpango/glg"
)

var serv server.Server

func process_command(w http.ResponseWriter, req *http.Request, serv *server.Server, XML string) string {
    cmd, errv := serv.Xml_parser.ParseMessage(XML)

    /* generate server transaction id before main procedure */
    SvTRID := server.GenerateTRID(10)
    logger := server.NewLogger(SvTRID)
    ctx := context.WithValue(context.Background(), "logger", logger)

    if errv != nil {
        epp_res := EPPResult{}
        if cmd_err, ok := errv.(*xml.CommandError); ok {
            epp_res.RetCode = cmd_err.RetCode
            if cmd_err.Msg != "" {
                epp_res.Errors = []string{cmd_err.Msg}
            }
        } else {
            epp_res.RetCode = 2500
            epp_res.Errors = []string{fmt.Sprint(errv)}
        }
        dbconn, err := server.AcquireConn(serv.Pool, logger)
        if err != nil {
            logger.Error(err)
        } else {
            defer dbconn.Close()
            epp.ResolveErrorMsg(dbconn, &epp_res, LANG_EN)
        }
        return xml.GenerateResponse(&epp_res, "", "")
    } else {
        cmd.SvTRID = SvTRID
        if cmd.CmdType == EPP_HELLO {
            return xml.GenerateGreeting()
        }

        /* either get an ssl certificate fingerprint on login or session id otherwise */
        if cmd.CmdType == EPP_LOGIN {
            cert_fingerprint, err := server.GetCertificateFingerprint(serv, req)
            if err != nil {
                logger.Error(err)
            }
            if login_obj, ok := cmd.Content.(*xml.EPPLogin); ok {
                login_obj.Fingerprint = cert_fingerprint
            }
        } else {
            sessionid, err := req.Cookie("EPPSESSIONID")
            if err == nil {
                cmd.Sessionid, _ = strconv.ParseUint(sessionid.Value, 10, 64)
            }
        }

        epp_res := epp.ExecuteEPPCommand(ctx, serv, cmd)

        if epp_res.CmdType == EPP_LOGIN {
            if login_obj, ok := epp_res.Content.(*LoginResult); ok {
                logger.Trace("set session", login_obj.Sessionid)
                cookie := http.Cookie{Name: "EPPSESSIONID", Value:strconv.FormatUint(login_obj.Sessionid,10)}
                http.SetCookie(w, &cookie)
            }
        }

        return xml.GenerateResponse(epp_res, cmd.ClTRID, cmd.SvTRID)
    }
}

func handle_root(w http.ResponseWriter, req *http.Request) {
    XML, err := io.ReadAll(req.Body)
    if err != nil {
        epp_res := EPPResult{RetCode:2500}
        io.WriteString(w, xml.GenerateResponse(&epp_res, "", ""))
        return
    }

    start := time.Now()

    response := process_command(w, req, &serv, string(XML))

    elapsed := time.Since(start)

    glg.Trace("exec took ", elapsed)
    glg.Trace(response)

    io.WriteString(w, response)
}

func main() {
    config_file := flag.String("config", "server.conf", "filename with config")
    port := flag.Uint("port", 0, "port")

    flag.Parse()

    serv.RGconf.LoadConfig(*config_file)
    if *port > 0 {
        serv.RGconf.HTTPConf.Port = *port
    }

    err := serv.XmlParser.SetNamespaces(serv.RGconf.SchemaNs)
    if err != nil {
        log.Fatal(err)
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
