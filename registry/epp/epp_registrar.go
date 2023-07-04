package epp

import (
    "strings"

    "registry/xml"
    "registry/epp/dbreg/registrar"
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

    if !ctx.session.System {
        if registrar_data.Id != uint64(ctx.session.Regid) {
            return nil, &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
        }
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

