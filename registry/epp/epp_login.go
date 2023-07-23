package epp 

import (
    "fmt"
    "errors"

    "registry/xml"
    "registry/epp/dbreg"
    "registry/epp/dbreg/registrar"
    . "registry/epp/eppcom"
)

func authenticateRegistrar(ctx *EPPContext, regid uint, v *xml.EPPLogin) error {
    ctx.logger.Info("authenticate", regid, v.Fingerprint, v.PW)

    if err := registrar.AuthenticateRegistrar(ctx.dbconn, regid, v.Fingerprint, v.PW); err != nil {
        return err
    }

    if v.NewPW != "" {
        if err := registrar.SetNewPassword(ctx.dbconn, regid, v.PW, v.NewPW); err != nil {
            return err
        }
    }

    return nil
}

func epp_login_impl(ctx *EPPContext, v *xml.EPPLogin) (*EPPResult) {
    ctx.logger.Info("Login", v.Clid)

    if err := ctx.dbconn.Begin(); err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }
    defer ctx.dbconn.Rollback()

    reg_info, err := registrar.GetRegistrarByHandle(ctx.dbconn, v.Clid)
    if err != nil {
        perr := &dbreg.ParamError{}
        if errors.As(err, &perr) {
            return &EPPResult{RetCode:EPP_AUTHENTICATION_ERR}
        }
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }
    if err := authenticateRegistrar(ctx, reg_info.Id.Get(), v); err != nil {
        if errors.Is(err, dbreg.ObjectNotFound) {
            return &EPPResult{RetCode:EPP_AUTHENTICATION_ERR}
        }
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    sessionid, err := ctx.serv.Sessions.LoginSession(ctx.dbconn, reg_info.Id.Get(), v.Lang)
    if err != nil {
        var res = EPPResult{RetCode:EPP_SESSION_LIMIT, Msg:fmt.Sprint(err)}
        return &res
    }

    if err = ctx.dbconn.Commit(); err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    var res = EPPResult{RetCode:EPP_OK}
    var loginResult = LoginResult{Sessionid:sessionid}
    res.Content = &loginResult
    return &res
}
