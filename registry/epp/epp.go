package epp

import (
    "fmt"
    "context"

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
    logger server.Logger
}

func NewEPPContext(serv *server.Server) *EPPContext {
    return &EPPContext{serv:serv}
}

/* fill the text message by return code */
func (e *EPPContext) ResolveErrorMsg(db *server.DBConn, epp_result *EPPResult, lang uint) {
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

func (e *EPPContext) SetLogger(logger server.Logger) {
    e.logger = logger
}

func (e *EPPContext) GetLogger() server.Logger {
    return e.logger
}

func (e *EPPContext) GetReqContext(ctx context.Context) server.ReqContext {
    if reqctx, ok := ctx.Value("meta").(server.ReqContext); ok {
        return reqctx
    }   
    return server.ReqContext{}
}

func (ctx *EPPContext) ExecuteEPPCommand(ctx_ context.Context, cmd *xml.XMLCommand) (*EPPResult) {
    dbconn, err := server.AcquireConn(ctx.serv.Pool, ctx.logger)
    if err != nil {
        return &EPPResult{CmdType:EPP_UNKNOWN_CMD, RetCode:EPP_FAILED}
    }
    defer dbconn.Close()
    ctx.dbconn = dbconn

    /* default to english */
    Lang := uint(LANG_EN)

    if cmd.CmdType != EPP_LOGIN {
        ctx.session = ctx.serv.Sessions.CheckSession(dbconn, cmd.Sessionid)
        if ctx.session == nil {
            epp_result := &EPPResult{CmdType:EPP_UNKNOWN_CMD, RetCode:EPP_AUTHENTICATION_ERR}
            ctx.ResolveErrorMsg(ctx.dbconn, epp_result, Lang)
            return epp_result
        }
        Lang = ctx.session.Lang
        if cmd.CmdType != EPP_LOGOUT && ctx.serv.Sessions.QueryLimitExceeded(ctx.session.Regid) {
            ctx.logger.Info(ctx.session.Regid, " exceeded limit on queries")
            epp_result := &EPPResult{CmdType:cmd.CmdType, RetCode:EPP_SESSION_LIMIT, Msg:"exceeded number of queries per minute"}
            ctx.ResolveErrorMsg(ctx.dbconn, epp_result, Lang)
            return epp_result
        }
    }

    var epp_result *EPPResult

    switch cmd.CmdType {
        case EPP_LOGIN:
            if v, ok := cmd.Content.(*xml.EPPLogin) ; ok {
                Lang = v.Lang
                epp_result = epp_login_impl(ctx, v)
            }

        case EPP_LOGOUT:
            ctx.logger.Info("Logout", cmd.Sessionid)
            err = ctx.serv.Sessions.LogoutSession(dbconn, cmd.Sessionid)
            if err != nil {
                return &EPPResult{RetCode:EPP_FAILED}
            }
            epp_result = &EPPResult{CmdType:EPP_LOGOUT, RetCode:EPP_CLOSING_LOGOUT}
        case EPP_CHECK_DOMAIN:
            if v, ok := cmd.Content.(*xml.CheckObject) ; ok {
                epp_result = epp_domain_check_impl(ctx, v)
            }
        case EPP_INFO_DOMAIN:
            if v, ok := cmd.Content.(*xml.InfoDomain) ; ok {
                epp_result = epp_domain_info_impl(ctx, v)
            }
        case EPP_CREATE_DOMAIN:
            if v, ok := cmd.Content.(*xml.CreateDomain) ; ok {
                epp_result = epp_domain_create_impl(ctx, v)
            }
        case EPP_UPDATE_DOMAIN:
            if v, ok := cmd.Content.(*xml.UpdateDomain) ; ok {
                epp_result = epp_domain_update_impl(ctx, v)
            }
        case EPP_RENEW_DOMAIN:
            if v, ok := cmd.Content.(*xml.RenewDomain) ; ok {
                epp_result = epp_domain_renew_impl(ctx, v)
            }
        case EPP_TRANSFER_DOMAIN:
            if v, ok := cmd.Content.(*xml.TransferDomain) ; ok {
                epp_result = epp_domain_transfer_impl(ctx, v)
            }
        case EPP_DELETE_DOMAIN:
            if v, ok := cmd.Content.(*xml.DeleteObject) ; ok {
                epp_result = epp_domain_delete_impl(ctx, v)
            }
        case EPP_CHECK_CONTACT:
            if v, ok := cmd.Content.(*xml.CheckObject) ; ok {
                epp_result = epp_contact_check_impl(ctx, v)
            }
        case EPP_INFO_CONTACT:
            if v, ok := cmd.Content.(*xml.InfoObject) ; ok {
                epp_result = epp_contact_info_impl(ctx, v)
            }
        case EPP_CREATE_CONTACT:
            if v, ok := cmd.Content.(*xml.CreateContact) ; ok {
                epp_result = epp_contact_create_impl(ctx, v)
            }
        case EPP_UPDATE_CONTACT:
            if v, ok := cmd.Content.(*xml.UpdateContact) ; ok {
                epp_result = epp_contact_update_impl(ctx, v)
            }
        case EPP_DELETE_CONTACT:
            if v, ok := cmd.Content.(*xml.DeleteObject) ; ok {
                epp_result = epp_contact_delete_impl(ctx, v)
            }
        case EPP_CHECK_HOST:
            if v, ok := cmd.Content.(*xml.CheckObject) ; ok {
                epp_result = epp_host_check_impl(ctx, v)
            }
        case EPP_INFO_HOST:
            if v, ok := cmd.Content.(*xml.InfoObject) ; ok {
                epp_result = epp_host_info_impl(ctx, v)
            }
        case EPP_CREATE_HOST:
            if v, ok := cmd.Content.(*xml.CreateHost) ; ok {
                epp_result = epp_host_create_impl(ctx, v)
            }
        case EPP_UPDATE_HOST:
            if v, ok := cmd.Content.(*xml.UpdateHost) ; ok {
                epp_result = epp_host_update_impl(ctx, v)
            }
        case EPP_DELETE_HOST:
            if v, ok := cmd.Content.(*xml.DeleteObject) ; ok {
                epp_result = epp_host_delete_impl(ctx, v)
            }
        case EPP_POLL_REQ:
            epp_result = epp_poll_req_impl(ctx)
        case EPP_POLL_ACK:
            if v, ok := cmd.Content.(string) ; ok  {
                epp_result = epp_poll_ack_impl(ctx, v)
            }
        case EPP_INFO_REGISTRAR:
            if v, ok := cmd.Content.(*xml.InfoObject) ; ok {
                epp_result = epp_registrar_info_impl(ctx, v)
            }
        case EPP_UPDATE_REGISTRAR:
            if v, ok := cmd.Content.(*xml.UpdateRegistrar) ; ok {
                epp_result = epp_registrar_update_impl(ctx, v)
            }
        default:
            epp_result = &EPPResult{CmdType:EPP_UNKNOWN_CMD, RetCode:EPP_UNKNOWN_ERR}
    }
    if epp_result == nil {
        epp_result = &EPPResult{CmdType:EPP_UNKNOWN_CMD, RetCode:EPP_UNKNOWN_ERR}
    }
    epp_result.CmdType = cmd.CmdType

    ctx.ResolveErrorMsg(ctx.dbconn, epp_result, Lang)
    return epp_result
}

func authenticateRegistrar(db *server.DBConn, regid uint, v *xml.EPPLogin) (bool, error) {
    var cert string
    glg.Info("authenticate", regid, v.Fingerprint, v.PW)
    row := db.QueryRow("SELECT cert FROM registraracl " +
                       "WHERE registrarid = $1::integer and cert = $2::text and password = $3::text", regid, v.Fingerprint, v.PW)
    err := row.Scan(&cert)

    if err != nil {
        if err == pgx.ErrNoRows {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

func epp_login_impl(ctx *EPPContext, v *xml.EPPLogin) (*EPPResult) {
    ctx.logger.Info("Login", v.Clid)
    var id uint
    var system bool
    var requests int

    row := ctx.dbconn.QueryRow("SELECT id, system, epp_requests_limit" +
                               " FROM registrar WHERE handle = $1::text", v.Clid)
    err := row.Scan(&id, &system, &requests)
    if err != nil {
        if err == pgx.ErrNoRows {
            return &EPPResult{RetCode:EPP_AUTHENTICATION_ERR}
        }
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }
    if ok, err := authenticateRegistrar(ctx.dbconn, id, v); !ok || err != nil {
        if !ok {
            return &EPPResult{RetCode:EPP_AUTHENTICATION_ERR}
        }
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
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
