package epp

import (
    "fmt"
    "registry/xml"
    "registry/server"
    . "registry/epp/eppcom"
    "github.com/jackc/pgx/v5"
    "github.com/kpango/glg"
)

type EPPContext struct {
    serv *server.Server
    session *server.EPPSession
    dbconn *server.DBConn
}

/* fill the text message by return code */
func ResolveErrorMsg(db *server.DBConn, epp_result *EPPResult, lang uint) {
    status_field := "status"
    if lang == LANG_RU {
        status_field = "status_ru"
    }
    row := db.QueryRow("SELECT " + status_field + " FROM enum_error " +
                       "WHERE id = $1::integer ", epp_result.RetCode)

    err := row.Scan(&epp_result.Msg)
    if err != nil {
        glg.Error(err)
    }
}

func ExecuteEPPCommand(serv *server.Server, cmd *xml.XMLCommand) (*EPPResult) {
    dbconn, err := server.AcquireConn(serv.Pool)
    if err != nil {
        return &EPPResult{CmdType:EPP_UNKNOWN_CMD, RetCode:EPP_FAILED}
    }
    defer dbconn.Close()
    ctx := EPPContext{serv:serv, dbconn:dbconn}

    /* default to english */
    Lang := uint(LANG_EN)

    glg.Info(cmd.SvTRID)
    if cmd.CmdType != EPP_LOGIN {
        // need to set prefix for each coroutine, not globally
//        glg.SetPrefix(glg.TRACE, fmt.Sprint(cmd.SvTRID))
//        glg.Trace(cmd.SvTID)
        ctx.session = serv.Sessions.CheckSession(dbconn, cmd.Sessionid)
        if ctx.session == nil {
            epp_result := &EPPResult{CmdType:EPP_UNKNOWN_CMD, RetCode:EPP_AUTHENTICATION_ERR}
            ResolveErrorMsg(ctx.dbconn, epp_result, Lang)
            return epp_result
        }
        Lang = ctx.session.Lang
        if cmd.CmdType != EPP_LOGOUT && serv.Sessions.QueryLimitExceeded(ctx.session.Regid) {
            glg.Info(ctx.session.Regid, " exceeded limit on queries")
            epp_result := &EPPResult{CmdType:cmd.CmdType, RetCode:EPP_SESSION_LIMIT, Msg:"exceeded number of queries per minute"}
            ResolveErrorMsg(ctx.dbconn, epp_result, Lang)
            return epp_result
        }
    }

    var epp_result *EPPResult

    switch cmd.CmdType {
        case EPP_LOGIN:
            if v, ok := cmd.Content.(*xml.EPPLogin) ; ok {
                Lang = v.Lang
                epp_result = epp_login_impl(&ctx, v)
            }

        case EPP_LOGOUT:
            glg.Info("Logout", cmd.Sessionid)
            serv.Sessions.LogoutSession(dbconn, cmd.Sessionid)
            epp_result = &EPPResult{CmdType:EPP_LOGOUT, RetCode:EPP_CLOSING_LOGOUT}
        case EPP_CHECK_DOMAIN:
            if v, ok := cmd.Content.(*xml.CheckObject) ; ok {
                epp_result = epp_domain_check_impl(&ctx, v)
            }
        case EPP_INFO_DOMAIN:
            if v, ok := cmd.Content.(*xml.InfoDomain) ; ok {
                epp_result = epp_domain_info_impl(&ctx, v)
            }
        case EPP_CREATE_DOMAIN:
            if v, ok := cmd.Content.(*xml.CreateDomain) ; ok {
                epp_result = epp_domain_create_impl(&ctx, v)
            }
        case EPP_UPDATE_DOMAIN:
            if v, ok := cmd.Content.(*xml.UpdateDomain) ; ok {
                epp_result = epp_domain_update_impl(&ctx, v)
            }
        case EPP_RENEW_DOMAIN:
            if v, ok := cmd.Content.(*xml.RenewDomain) ; ok {
                epp_result = epp_domain_renew_impl(&ctx, v)
            }
        case EPP_TRANSFER_DOMAIN:
            if v, ok := cmd.Content.(*xml.TransferDomain) ; ok {
                epp_result = epp_domain_transfer_impl(&ctx, v)
            }
        case EPP_DELETE_DOMAIN:
            if v, ok := cmd.Content.(*xml.DeleteObject) ; ok {
                epp_result = epp_domain_delete_impl(&ctx, v)
            }
        case EPP_INFO_CONTACT:
            if v, ok := cmd.Content.(*xml.InfoContact) ; ok {
                epp_result = epp_contact_info_impl(&ctx, v)
            }
        case EPP_CREATE_CONTACT:
            if v, ok := cmd.Content.(*xml.CreateContact) ; ok {
                epp_result = epp_contact_create_impl(&ctx, v)
            }
        case EPP_UPDATE_CONTACT:
            if v, ok := cmd.Content.(*xml.UpdateContact) ; ok {
                epp_result = epp_contact_update_impl(&ctx, v)
            }
        case EPP_DELETE_CONTACT:
            if v, ok := cmd.Content.(*xml.DeleteObject) ; ok {
                epp_result = epp_contact_delete_impl(&ctx, v)
            }
        case EPP_INFO_HOST:
            if v, ok := cmd.Content.(*xml.InfoHost) ; ok {
                epp_result = epp_host_info_impl(&ctx, v)
            }
        case EPP_CREATE_HOST:
            if v, ok := cmd.Content.(*xml.CreateHost) ; ok {
                epp_result = epp_host_create_impl(&ctx, v)
            }
        case EPP_UPDATE_HOST:
            if v, ok := cmd.Content.(*xml.UpdateHost) ; ok {
                epp_result = epp_host_update_impl(&ctx, v)
            }
        case EPP_DELETE_HOST:
            if v, ok := cmd.Content.(*xml.DeleteObject) ; ok {
                epp_result = epp_host_delete_impl(&ctx, v)
            }
        case EPP_POLL_REQ:
            epp_result = epp_poll_req_impl(&ctx)
        case EPP_POLL_ACK:
            if v, ok := cmd.Content.(string) ; ok  {
                epp_result = epp_poll_ack_impl(&ctx, v)
            }
        default:
            epp_result = &EPPResult{CmdType:EPP_UNKNOWN_CMD, RetCode:EPP_UNKNOWN_ERR}
    }
    if epp_result == nil {
        epp_result = &EPPResult{CmdType:EPP_UNKNOWN_CMD, RetCode:EPP_UNKNOWN_ERR}
    }
    epp_result.CmdType = cmd.CmdType

    ResolveErrorMsg(ctx.dbconn, epp_result, Lang)
    return epp_result
}

func authenticateRegistrar(db *server.DBConn, regid uint, v *xml.EPPLogin) bool {
    var cert string
    glg.Info("authenticate", regid, v.Fingerprint, v.PW)
    row := db.QueryRow("SELECT cert, password FROM registraracl " +
                       "WHERE registrarid = $1::integer and cert = $2::text and password = $3::text", regid, v.Fingerprint, v.PW)
    err := row.Scan(&cert)

    if err == pgx.ErrNoRows {
        return false
    }
    return true
}

func epp_login_impl(ctx *EPPContext, v *xml.EPPLogin) (*EPPResult) {
    glg.Info("Login", v.Clid)
    var id uint
    var system bool
    var requests int

    row := ctx.dbconn.QueryRow("SELECT id, system, epp_requests_limit" +
                               " FROM registrar WHERE handle = $1::text", v.Clid)
    err := row.Scan(&id, &system, &requests)
    if err == pgx.ErrNoRows {
        return &EPPResult{RetCode:EPP_AUTHENTICATION_ERR}
    }
    if !authenticateRegistrar(ctx.dbconn, id, v) {
        return &EPPResult{RetCode:EPP_AUTHENTICATION_ERR}
    }

    sessionid, err := ctx.serv.Sessions.LoginSession(ctx.dbconn, id, v.Lang)
    if err != nil {
        var res = EPPResult{RetCode:EPP_SESSION_LIMIT, Msg:fmt.Sprint(err)}
        return &res
    }

    var res = EPPResult{RetCode:EPP_OK}
    var loginResult = LoginResult{Sessionid:sessionid}
    res.Content = &loginResult
    return &res
}
