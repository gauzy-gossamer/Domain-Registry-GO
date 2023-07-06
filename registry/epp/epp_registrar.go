package epp

import (
    "strings"

    "registry/xml"
    "registry/epp/dbreg/registrar"
    "registry/epp/dbreg"
    . "registry/epp/eppcom"
    "github.com/jackc/pgx/v5"
)

func get_registrar_object(ctx *EPPContext, registrar_handle string, for_update bool) (*InfoRegistrarData, *EPPResult) {
    info_db := registrar.NewInfoRegistrarDB()
    registrar_data, err := info_db.SetLock(for_update).SetHandle(registrar_handle).Exec(ctx.dbconn)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
        }
        ctx.logger.Error(err)
        return nil, &EPPResult{RetCode:EPP_FAILED}
    }

    return registrar_data, nil
}

func epp_registrar_info_impl(ctx *EPPContext, v *xml.InfoObject) (*EPPResult) {
    ctx.logger.Info("Info registrar", v.Name)
    registrar_handle := strings.ToUpper(v.Name)

    registrar_data, cmd := get_registrar_object(ctx, registrar_handle, false)
    if cmd != nil {
        return cmd
    }

    var err error
    registrar_data.Addrs, err = registrar.GetRegistrarIPAddrs(ctx.dbconn, registrar_data.Id)
    ctx.logger.Error(registrar_data.Addrs)
 
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = registrar_data
    return &res
}

func epp_registrar_update_impl(ctx *EPPContext, v *xml.UpdateRegistrar) *EPPResult {
    registrar_handle := strings.ToUpper(v.Name)
    ctx.logger.Info("Update registrar", registrar_handle)
    registrar_data, cmd := get_registrar_object(ctx, registrar_handle, true)
    if cmd != nil {
        return cmd
    }

    if !ctx.session.System {
        if registrar_data.Id != uint64(ctx.session.Regid) {
            return &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
        }
    }

    var err error
    if len(v.AddAddrs) > 0 {
        if err = checkIPAddresses(v.AddAddrs) ; err != nil {
            if perr, ok := err.(*dbreg.ParamError); ok {
                return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
            }
        }
    }
    if len(v.RemAddrs) > 0 {
        if err = checkIPAddresses(v.RemAddrs) ; err != nil {
            if perr, ok := err.(*dbreg.ParamError); ok {
                return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
            }
        }
    }

    // TODO check rem and add addresses
    //host_addrs := getHostIPAddresses(ctx.dbconn, host_data.Id)

    err = ctx.dbconn.Begin()
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }
    defer ctx.dbconn.Rollback()

    update_reg := registrar.NewUpdateRegistrar()

    if len(v.WWW) > 0 {
        update_reg.SetWWW(v.WWW)
    }

    if len(v.Whois) > 0 {
        update_reg.SetWhois(v.Whois)
    }
    if len(v.Voice) > 0 {
        update_reg.SetVoice(v.Voice)
    }
    if len(v.Fax) > 0 {
        update_reg.SetFax(v.Fax)
    }

    err = update_reg.Exec(ctx.dbconn, registrar_data.Id, ctx.session.Regid, v.AddAddrs, v.RemAddrs)
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    err = ctx.dbconn.Commit()
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    return &EPPResult{RetCode:EPP_OK}
}
