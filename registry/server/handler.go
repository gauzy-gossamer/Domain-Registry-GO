package server

import (
    "time"
    "fmt"
    "net/http"
    "strconv"
    "context"

    . "registry/epp/eppcom"
    "registry/xml"
)

type EPPContext interface {
    ResolveErrorMsg(db *DBConn, epp_result *EPPResult, lang uint)
    ExecuteEPPCommand(ctx_ context.Context, cmd *xml.XMLCommand) (*EPPResult)
    GetLogger() Logger
    SetLogger(logger Logger)
    GetReqContext(ctx context.Context) ReqContext
}

type ReqContext struct {
    SvTRID string
    IPAddr string
}

/* either get from X-Forwarded-For or from Request */
func GetUserIPAddr(serv *Server, req *http.Request) string {
    if serv.RGconf.HTTPConf.UseProxy {
        return req.Header.Get("X-Forwarded-For")
    } else {
        return req.RemoteAddr
    }   
}

func ProcessCommand(ctx context.Context, epp EPPContext, w http.ResponseWriter, req *http.Request, serv *Server, XML string) string {
    logger := epp.GetLogger()
    cmd, errv := serv.XmlParser.ParseMessage(XML)
    reqctx := epp.GetReqContext(ctx)

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
        dbconn, err := AcquireConn(serv.Pool, logger)
        if err != nil {
            logger.Error(err)
        } else {
            defer dbconn.Close()
            epp.ResolveErrorMsg(dbconn, &epp_res, LANG_EN)
        }
        return xml.GenerateResponse(&epp_res, "", "")
    } else {
        if cmd.CmdType == EPP_HELLO {
            return xml.GenerateGreeting()
        }

        /* either get an ssl certificate fingerprint on login or session id otherwise */
        if cmd.CmdType == EPP_LOGIN {
            cert_fingerprint, err := GetCertificateFingerprint(serv, req)
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

        epp_res := epp.ExecuteEPPCommand(ctx, cmd)

        if epp_res.CmdType == EPP_LOGIN {
            if login_obj, ok := epp_res.Content.(*LoginResult); ok {
                logger.Trace("set session", login_obj.Sessionid)
                cookie := http.Cookie{Name: "EPPSESSIONID", Value:strconv.FormatUint(login_obj.Sessionid,10)}
                http.SetCookie(w, &cookie)
            }
        }

        return xml.GenerateResponse(epp_res, cmd.ClTRID, reqctx.SvTRID)
    }
}

func HandleRequest(serv *Server, epp EPPContext, w http.ResponseWriter, req *http.Request, XML string) string {
    start := time.Now()

    /* generate server transaction id before main procedure */
    ipaddr := GetUserIPAddr(serv, req)
    SvTRID := GenerateTRID(10)
    logger := NewLogger(SvTRID)
    epp.SetLogger(logger)
    ctx := context.WithValue(context.Background(), "meta", ReqContext{IPAddr:ipaddr, SvTRID:SvTRID})

    logger.Trace("query from ", ipaddr)

    response := ProcessCommand(ctx, epp, w, req, serv, XML)

    elapsed := time.Since(start)

    logger.Trace("exec took ", elapsed)
    logger.Trace(response)

    return response
}
