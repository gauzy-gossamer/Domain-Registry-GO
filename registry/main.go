package main

import (
//  "os"
    "time"
    "fmt"
    "flag"
    "crypto/tls"
    "net/http"
    "log"
    "io"
    "strconv"
    "io/ioutil"
    "registry/server"
    "registry/epp"
    . "registry/epp/eppcom"
    "registry/xml"
    "registry/regrpc"
    "github.com/kpango/glg"
)

var serv server.Server

func process_command(w http.ResponseWriter, req *http.Request, serv *server.Server, XML string) string {
    var cmd *xml.XMLCommand
    var errv error

    cmd, errv = serv.Xml_parser.ParseMessage(XML)

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
        dbconn, err := server.AcquireConn(serv.Pool)
        defer dbconn.Close()
        if err == nil {
            epp.ResolveErrorMsg(dbconn, &epp_res, LANG_EN)
        }
        return xml.GenerateResponse(&epp_res, "", "")
    } else {
        if cmd.CmdType == EPP_HELLO {
            return xml.GenerateGreeting()
        }

        /* either get an ssl certificate fingerprint on login or session id otherwise */
        if cmd.CmdType == EPP_LOGIN {
            cert_fingerprint, err := server.GetCertificateFingerprint(serv, req)
            if err != nil {
                glg.Error(err)
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

        /* generate server transaction id before main procedure */
        cmd.SvTRID = server.GenerateTRID(10)
        epp_res := epp.ExecuteEPPCommand(serv, cmd)

        if epp_res.CmdType == EPP_LOGIN {
            if login_obj, ok := epp_res.Content.(*LoginResult); ok {
                glg.Trace("set session", login_obj.Sessionid)
                cookie := http.Cookie{Name: "EPPSESSIONID", Value:strconv.FormatUint(login_obj.Sessionid,10)}
                http.SetCookie(w, &cookie)
            }
        }

        return xml.GenerateResponse(epp_res, cmd.ClTRID, cmd.SvTRID)
    }
}

func handle_root(w http.ResponseWriter, req *http.Request) {
    XML, err := ioutil.ReadAll(req.Body)
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

    serv.Xml_parser.ReadSchema(serv.RGconf.SchemaPath)
    var err error
    serv.Pool, err = server.CreatePool(&serv.RGconf.DBconf)
    if err != nil {
        log.Fatal(err)
    }
    dbconn, err := server.AcquireConn(serv.Pool)
    if err != nil {
        log.Fatal(err)
    }
    defer dbconn.Close()
    serv.Sessions.MaxRegistrarSessions = serv.RGconf.MaxRegistrarSessions
    serv.Sessions.SessionTimeoutSec = serv.RGconf.SessionTimeout
    serv.Sessions.InitSessions(dbconn)

    go regrpc.StartgRPCServer(&serv)

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
        httpserver.ListenAndServe()
    } else {
        httpserver.ListenAndServeTLS(serv.RGconf.HTTPConf.CertFile, serv.RGconf.HTTPConf.KeyFile)
    }
}
