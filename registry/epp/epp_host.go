package epp

import (
    "fmt"
    "strings"
    "registry/xml"
    "registry/epp/dbreg"
    "registry/epp/dbreg/host"
    . "registry/epp/eppcom"
    "github.com/jackc/pgx/v5"
)

func hostRegistrarHandle(handle string, regid uint) string {
    return fmt.Sprintf("%s:%d", handle, regid)
}

func epp_host_check_impl(ctx *EPPContext, v *xml.CheckObject) (*EPPResult) {
    ctx.logger.Info("Host check", v.Names)

    var check_results []CheckResult

    for _, host := range v.Names {
        host_name := normalizeDomainUpper(host)
        host_handle := hostRegistrarHandle(host_name, ctx.session.Regid)
        if !checkDomainValidity(host_name) {
            check_results = append(check_results, CheckResult{Name:host, Result:CD_NOT_APPLICABLE})
            continue
        }

        _, err := dbreg.GetHostObject(ctx.dbconn, host_handle, ctx.session.Regid)
        if err == nil {
            check_results = append(check_results, CheckResult{Name:host, Result:CD_REGISTERED})
            continue
        } else if err != pgx.ErrNoRows {
            ctx.logger.Error(err)
            continue
        }

        check_results = append(check_results, CheckResult{Name:host, Result:CD_AVAILABLE})
    }   

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = check_results
    return &res
}

func get_host_object(ctx *EPPContext, host_handle string, for_update bool) (*InfoHostData, *ObjectStates, *EPPResult) {
    info_db := host.NewInfoHostDB()
    host_data, err := info_db.SetLock(for_update).Set_fqdn(host_handle).Exec(ctx.dbconn)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil, &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
        }
        ctx.logger.Error(err)
        return nil, nil, &EPPResult{RetCode:EPP_FAILED}
    }

    if !ctx.session.System {
        if host_data.Sponsoring_registrar.Id.Get() != ctx.session.Regid {
            return nil, nil, &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
        }
    }

    if for_update {
        if err := UpdateObjectStates(ctx.dbconn, host_data.Id); err != nil {
            ctx.logger.Error(err)
            return nil,nil, &EPPResult{RetCode:EPP_FAILED}
        }
    }

    object_states, err := getObjectStates(ctx.dbconn, host_data.Id)
    if err != nil {
        ctx.logger.Error(err)
        return nil,nil, &EPPResult{RetCode:EPP_FAILED}
    }

    return host_data, object_states, nil
}

func normalizeHostHandlePair(ctx *EPPContext, name string) (string, string, *EPPResult) {
    host_name := normalizeDomainUpper(name)
    host_handle := hostRegistrarHandle(host_name, ctx.session.Regid)
    if !checkDomainValidity(host_name) {
        return host_name, host_handle, &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{"incorrect host name"}}
    }
    _, err := dbreg.GetHostObject(ctx.dbconn, host_handle, ctx.session.Regid)
    if err == nil {
        return host_name, host_handle, &EPPResult{RetCode:EPP_OBJECT_EXISTS}
    } else if err != pgx.ErrNoRows {
        ctx.logger.Error(err)
        return host_name, host_handle, &EPPResult{RetCode:EPP_FAILED}
    }

    return host_name, host_handle, nil
}

func epp_host_info_impl(ctx *EPPContext, v *xml.InfoObject) (*EPPResult) {
    ctx.logger.Info("Info host", v.Name)
    host_name := normalizeDomainUpper(v.Name)
    host_handle := hostRegistrarHandle(host_name, ctx.session.Regid)
    host_data, object_states, cmd := get_host_object(ctx, host_handle, false)
    if cmd != nil {
        return cmd
    }

    host_data.States = object_states.copyObjectStates()

    var err error
    host_data.Addrs, err = host.GetHostIPAddrs(ctx.dbconn, host_data.Id)
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = host_data
    return &res
}

func epp_host_create_impl(ctx *EPPContext, v *xml.CreateHost) (*EPPResult) {
    ctx.logger.Info("Create host", v.Name, v.Addr)

    host_name, host_handle, err_res := normalizeHostHandlePair(ctx, v.Name)
    if err_res != nil {
        return err_res
    }

    var err error
    if len(v.Addr) > 0 {
        if ok, err := isHostSubordinate(ctx.dbconn, host_name, ctx.session.Regid); !ok || err != nil  {
            if err != nil {
                ctx.logger.Error(err)
                return &EPPResult{RetCode:EPP_FAILED}
            }
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{"host is not subordinate to any registrar zones"}}
        }
        if err = checkIPAddresses(v.Addr) ; err != nil {
            if perr, ok := err.(*dbreg.ParamError); ok {
                return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
            }
        }
    }

    if err = ctx.dbconn.Begin(); err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode: EPP_FAILED}
    }
    defer ctx.dbconn.Rollback()

    create_host := host.NewCreateHostDB()
    create_result, err := create_host.SetParams(host_handle, ctx.session.Regid, strings.ToLower(host_name), v.Addr).Exec(ctx.dbconn)
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode: EPP_FAILED}
    }

    if err = ctx.dbconn.Commit(); err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode: EPP_FAILED}
    }

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = create_result
    return &res
}

func epp_host_update_impl(ctx *EPPContext, v *xml.UpdateHost) *EPPResult {
    host_name := normalizeDomainUpper(v.Name)
    host_handle := hostRegistrarHandle(host_name, ctx.session.Regid)
    ctx.logger.Info("Update host", host_name)
    host_data, object_states, cmd := get_host_object(ctx, host_handle, true)
    if cmd != nil {
        return cmd
    }

    var err error
    if len(v.AddAddrs) > 0 {
        if ok, err := isHostSubordinate(ctx.dbconn, host_name, ctx.session.Regid); !ok || err != nil  {
            if err != nil {
                ctx.logger.Error(err)
                return &EPPResult{RetCode:EPP_FAILED}
            }
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{"host is not subordinate to any registrar zones"}}
        }
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

    update_host := host.NewUpdateHostDB()

    if v.NewName != "" {
        new_host_name, new_host_handle, res := normalizeHostHandlePair(ctx, v.NewName)
        if res != nil {
            return res
        }
        update_host.SetNewName(new_host_name, new_host_handle)
    }

    if err = ctx.dbconn.Begin(); err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }
    defer ctx.dbconn.Rollback()

    if len(v.AddStatus) > 0 || len(v.RemStatus) > 0 {
        err := updateObjectClientStates(ctx, host_data.Id, object_states, v.AddStatus, v.RemStatus, "nsset")
        if err != nil {
            if perr, ok := err.(*dbreg.ParamError); ok {
                return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
            }
            ctx.logger.Error(err)
            return &EPPResult{RetCode:EPP_FAILED}
        }
    }

    if !ctx.session.System {
        if object_states.hasState(serverUpdateProhibited) ||
           object_states.hasState(clientUpdateProhibited) ||
           object_states.hasState(pendingDelete) {
            return &EPPResult{RetCode:EPP_STATUS_PROHIBITS_OPERATION}
        }
    }

    if len(v.RemAddrs) > 0 || len(v.AddAddrs) > 0 {
        host_addrs, err := host.GetHostIPAddrs(ctx.dbconn, host_data.Id)
        if err != nil {
            ctx.logger.Error(err)
            return &EPPResult{RetCode:EPP_FAILED}
        }

        /* TODO normalize ipv6 addresses */
        err_addr := allAddrsPresent(v.RemAddrs, host_addrs)
        if err_addr != "" {
            err_msg := err_addr + " isn't present for this host"
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{err_msg}}
        }
        err_addr = anyAddrsPresent(host_addrs, v.AddAddrs)
        if err_addr != "" {
            err_msg := err_addr + " already present for this host"
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{err_msg}}
        }

        if len(host_addrs) + len(v.AddAddrs) - len(v.RemAddrs) > ctx.serv.RGconf.MaxValueList {
            err_val := "Maximum number of IPs exceeded"
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{err_val}}
        }
    }

    if err = update_host.Exec(ctx.dbconn, host_data.Id, ctx.session.Regid, v.AddAddrs, v.RemAddrs); err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if err = ctx.dbconn.Commit(); err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    return &EPPResult{RetCode:EPP_OK}
}

func epp_host_delete_impl(ctx *EPPContext, v *xml.DeleteObject) *EPPResult {
    host_name := normalizeDomainUpper(v.Name)
    host_handle := hostRegistrarHandle(host_name, ctx.session.Regid)
    ctx.logger.Info("Delete host", host_name)
    host_data, object_states, cmd := get_host_object(ctx, host_handle, true)
    if cmd != nil {
        return cmd
    }

    if !ctx.session.System {
        if object_states.hasState(serverDeleteProhibited) ||
           object_states.hasState(clientDeleteProhibited) ||
           object_states.hasState(pendingDelete) {
            return &EPPResult{RetCode:EPP_STATUS_PROHIBITS_OPERATION}
        }
    }
    if object_states.hasState(stateLinked) {
        return &EPPResult{RetCode:EPP_LINKED_PROHIBITS_OPERATION}
    }

    err := ctx.dbconn.Begin()
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()

    err = host.DeleteHost(ctx.dbconn, host_data.Id)
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    err = ctx.dbconn.Commit()
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }

    return &EPPResult{RetCode:EPP_OK}
}
