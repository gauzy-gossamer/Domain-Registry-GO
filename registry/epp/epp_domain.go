package epp

import (
    "time"
    "fmt"
    "registry/xml"
    "registry/epp/dbreg"
    . "registry/epp/eppcom"
    "github.com/jackc/pgx/v5"
    "github.com/kpango/glg"
)

func epp_domain_check_impl(ctx *EPPContext, v *xml.CheckObject) (*EPPResult) {
    glg.Info("Domain check", v.Names)

    var check_results []CheckResult
    domain_checker := NewDomainChecker()

    for _, domain := range v.Names {
        domain = normalizeDomain(domain)
        if !checkDomainValidity(domain) {
            check_results = append(check_results, CheckResult{Name:domain, Result:CD_NOT_APPLICABLE})
            continue
        }

        zone := getDomainZone(ctx.dbconn, domain)

        if zone == nil {
            check_results = append(check_results, CheckResult{Name:domain, Result:CD_NOT_APPLICABLE})
            continue
        }
        if ok, err := isDomainAvailable(ctx.dbconn, domain); !ok {
            if err != nil {
                glg.Error(err)
            }
            check_results = append(check_results, CheckResult{Name:domain, Result:CD_REGISTERED})
            continue
        }
        if ok, err := domain_checker.IsDomainRegistrable(ctx.dbconn, domain, zone.id); !ok {
            if err != nil {
                glg.Error(err)
            }
            check_results = append(check_results, CheckResult{Name:domain, Result:CD_NOT_APPLICABLE})
            continue
        }
        check_results = append(check_results, CheckResult{Name:domain, Result:CD_AVAILABLE})
    }

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = check_results
    return &res
}

func epp_domain_info_impl(ctx *EPPContext, v *xml.InfoDomain) (*EPPResult) {
    glg.Info("Domain info", v.Name)
    domain := normalizeDomain(v.Name)
    info_db := dbreg.NewInfoDomainDB()
    domain_data, err := info_db.Set_fqdn(domain).Exec(ctx.dbconn)
    if err != nil {
        if err == pgx.ErrNoRows {
            return &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
        }
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if !ctx.session.System {
        if domain_data.Sponsoring_registrar.Id.Get() != ctx.session.Regid {
            return &EPPResult{RetCode:EPP_AUTHENTICATION_ERR}
        }
    }

    object_states, err := getObjectStates(ctx.dbconn, domain_data.Id)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    domain_data.States = object_states.copyObjectStates()

    domain_hosts, err := dbreg.GetDomainHosts(ctx.dbconn, domain_data.Id)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }
    for _, host := range domain_hosts {
        domain_data.Hosts = append(domain_data.Hosts, host.Fqdn)
    }

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = domain_data
    return &res
}

func epp_domain_create_impl(ctx *EPPContext, v *xml.CreateDomain) (*EPPResult) {
    glg.Info("Domain create", v.Name)
    domain := normalizeDomain(v.Name)

    if !checkDomainValidity(domain) {
        return &EPPResult{RetCode:EPP_PARAM_ERR}
    }

    zone := getDomainZone(ctx.dbconn, domain)

    if zone == nil {
        return &EPPResult{RetCode:2306}
    }

    if !ctx.session.System {
        if !testRegistrarZoneAccess(ctx.dbconn, ctx.session.Regid, zone.id) {
            return &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
        }
    }

    if ok, err := isDomainAvailable(ctx.dbconn, domain); !ok {
        if err != nil {
            glg.Error(err)
        }
        return &EPPResult{RetCode:EPP_OBJECT_EXISTS}
    }
    if ok, err := NewDomainChecker().IsDomainRegistrable(ctx.dbconn, domain, zone.id); !ok {
        if err != nil {
            glg.Error(err)
        }
        return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY}
    }

    registrant, err := dbreg.GetContactIdByHandle(ctx.dbconn, v.Registrant, ctx.session.Regid)
    if err != nil {
        if perr, ok := err.(*dbreg.ParamError); ok {
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
        }
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    var host_objects []dbreg.HostObj
    if len(v.Hosts) > 0 {
        host_objects, err = dbreg.GetHostObjects(ctx.dbconn, normalizeHosts(v.Hosts), ctx.session.Regid)
        if err != nil {
            if perr, ok := err.(*dbreg.ParamError); ok {
                return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
            }
            glg.Error(err)
            return &EPPResult{RetCode:EPP_FAILED}
        }
        cmd := testNumberOfHosts(ctx, len(host_objects))
        if cmd != nil {
            return cmd
        }
    }

    err = ctx.dbconn.Begin()
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()

    create_domain := dbreg.NewCreateDomainDB()
    result, err := create_domain.SetParams(domain, zone.id, registrant, ctx.session.Regid, v.Description, host_objects).Exec(ctx.dbconn)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    if ctx.serv.RGconf.ChargeOperations {
        err = dbreg.ChargeCreateOp(ctx.dbconn, result.Id, ctx.session.Regid, zone.id, result.Crdate)
        if err != nil {
            if _, ok := err.(*dbreg.BillingFailure); ok {
                return &EPPResult{RetCode:EPP_BILLING_FAILURE}
            }
            glg.Error(err)
            return &EPPResult{RetCode:2500}
        }
    }

    if err := UpdateObjectStates(ctx.dbconn, result.Id); err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    err = updateHostStates(ctx.dbconn, host_objects)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    err = ctx.dbconn.Commit()
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = result
    return &res
}

func get_domain_obj(ctx *EPPContext, domain string, for_update bool) (*InfoDomainData, *ObjectStates, *EPPResult) {
    info_db := dbreg.NewInfoDomainDB()
    domain_data, err := info_db.Set_lock(for_update).Set_fqdn(domain).Exec(ctx.dbconn)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil, &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
        }
        glg.Error(err)
        return nil, nil, &EPPResult{RetCode:EPP_FAILED}
    }

    if !ctx.session.System {
        if domain_data.Sponsoring_registrar.Id.Get() != ctx.session.Regid {
            return nil, nil, &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
        }
    }

    if for_update {
        err = UpdateObjectStates(ctx.dbconn, domain_data.Id)
        if err != nil {
            glg.Error(err)
            return nil, nil, &EPPResult{RetCode:EPP_FAILED}
        }
    }

    object_states, err := getObjectStates(ctx.dbconn, domain_data.Id)
    if err != nil {
        glg.Error(err)
        return nil, nil, &EPPResult{RetCode:EPP_FAILED}
    }

    return domain_data, object_states, nil
}

func testNumberOfHosts(ctx *EPPContext, hosts_n int) *EPPResult {
    /* zero hosts are allowed */
    if hosts_n == 0 {
        return nil
    }
    if hosts_n < ctx.serv.RGconf.DomainMinHosts {
        err_msg := fmt.Sprint("minimum number of hosts is ", ctx.serv.RGconf.DomainMinHosts)
        return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{err_msg}}
    }
    if hosts_n > ctx.serv.RGconf.DomainMaxHosts {
        err_msg := fmt.Sprint("maximum number of hosts is ", ctx.serv.RGconf.DomainMaxHosts)
        return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{err_msg}}
    }
    return nil
}

func updateObjectClientStates(ctx *EPPContext, object_id uint64, cur_states *ObjectStates, add_states_ []string, rem_states_ []string) error {
    add_states, err := getClientObjectStates(ctx.dbconn, add_states_, "domain")
    if err != nil {
        return err
    }

    rem_states, err := getClientObjectStates(ctx.dbconn, rem_states_, "domain")
    if err != nil {
        return err
    }

    /* some states prohibit update operations, so we can remove from cur_states,
    so that we can proceed after state checks
    on the other hand if we are setting these states, we allow updating other fields within the same operation
    */
    for state_name, state_id := range add_states {
        if cur_states.hasState(state_id) {
            err_msg := "state " + state_name + " already set up"
            return &dbreg.ParamError{Val:err_msg}
        }
        if _, err := dbreg.CreateObjectStateRequest(ctx.dbconn, object_id, uint(state_id)) ; err != nil {
            return err
        }
    }

    for _, state_id := range rem_states {
        if _, err := dbreg.CancelObjectStateRequest(ctx.dbconn, object_id, uint(state_id)) ; err != nil {
            return err
        }
        cur_states.deleteState(state_id)
    }

    return nil
}

func epp_domain_update_impl(ctx *EPPContext, v *xml.UpdateDomain) (*EPPResult) {
    glg.Info("Domain update", v.Name)
    domain := normalizeDomain(v.Name)

    domain_data, object_states, cmd := get_domain_obj(ctx, domain, true)
    if cmd != nil {
        return cmd
    }

    err := ctx.dbconn.Begin()
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()

    if len(v.AddStatus) > 0 || len(v.RemStatus) > 0 {
        err := updateObjectClientStates(ctx, domain_data.Id, object_states, v.AddStatus, v.RemStatus)
        if err != nil {
            if perr, ok := err.(*dbreg.ParamError); ok {
                return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
            }
            glg.Error(err)
            return &EPPResult{RetCode:EPP_FAILED}
        }
    }

    if !ctx.session.System {
        if object_states.hasState(serverUpdateProhibited) ||
           object_states.hasState(clientUpdateProhibited) ||
           object_states.hasState(changeProhibited) ||
           object_states.hasState(pendingDelete) {
            return &EPPResult{RetCode:EPP_STATUS_PROHIBITS_OPERATION}
        }
    }

    update_domain := dbreg.NewUpdateDomainDB()
    if len(v.AddHosts) > 0 || len(v.RemHosts) > 0 {
        add_hosts, err := dbreg.GetHostObjects(ctx.dbconn, normalizeHosts(v.AddHosts), ctx.session.Regid)
        if err != nil {
            if perr, ok := err.(*dbreg.ParamError); ok {
                return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
            }
            glg.Error(err)
            return &EPPResult{RetCode:EPP_FAILED}
        }

        rem_hosts, err := dbreg.GetHostObjects(ctx.dbconn, normalizeHosts(v.RemHosts), ctx.session.Regid)
        if err != nil {
            if perr, ok := err.(*dbreg.ParamError); ok {
                return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
            }
            glg.Error(err)
            return &EPPResult{RetCode:EPP_FAILED}
        }

        domain_hosts, err := dbreg.GetDomainHosts(ctx.dbconn, domain_data.Id)
        if err != nil {
            glg.Error(err)
            return &EPPResult{RetCode:EPP_FAILED}
        }
        err_host := allHostsPresent(rem_hosts, domain_hosts)
        if err_host != "" {
            err_msg := err_host + " isn't present for this domain"
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{err_msg}}
        }
        err_host = anyHostsPresent(domain_hosts, add_hosts)
        if err_host != "" {
            err_msg := err_host + " already present for this domain"
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{err_msg}}
        }

        total_hosts:= len(domain_hosts) - len(rem_hosts) + len(add_hosts)
        cmd := testNumberOfHosts(ctx, total_hosts)
        if cmd != nil {
            return cmd
        }
        update_domain.SetAddHosts(add_hosts).SetRemHosts(rem_hosts)
    }
    if v.Registrant != "" {
        registrant, err := dbreg.GetContactIdByHandle(ctx.dbconn, v.Registrant, ctx.session.Regid)
        if err != nil {
            if perr, ok := err.(*dbreg.ParamError); ok {
                return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
            }
            glg.Error(err)
            return &EPPResult{RetCode:EPP_FAILED}
        }
        glg.Error("registrant", v.Registrant, registrant)
        update_domain.SetRegistrant(registrant)
    }
    if len(v.Description) > 0 {
        update_domain.SetDescription(v.Description)
    }

    err = update_domain.Exec(ctx.dbconn, domain_data.Id, ctx.session.Regid)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    err = UpdateObjectStates(ctx.dbconn, domain_data.Id)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }
    /* if we updated registrant , the previous registrant could've become unlinked*/
    if v.Registrant != "" {
        err = deleteUnlinkedContacts(ctx, domain_data.Registrant.Id)
        if err != nil {
            glg.Error(err)
            return &EPPResult{RetCode:2500}
        }
    }

    err = ctx.dbconn.Commit()
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    return &EPPResult{RetCode:EPP_OK}
}

func testExpDate(userExpDate string, domainExpDate time.Time) bool {
    if !testDateValidity(userExpDate) {
        return false
    }

    return userExpDate == domainExpDate.UTC().Format("2006-01-02")
}

func epp_domain_renew_impl(ctx *EPPContext, v *xml.RenewDomain) (*EPPResult) {
    domain := normalizeDomain(v.Name)
    glg.Info("Renew domain", domain)

    domain_data, object_states, cmd := get_domain_obj(ctx, domain, true)
    if cmd != nil {
        return cmd
    }

    if !testExpDate(v.CurExpDate, domain_data.Expiration_date.Time) {
        return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{"incorrect exdate"}}
    }

    if !ctx.session.System {
        if object_states.hasState(serverRenewProhibited) ||
           object_states.hasState(clientRenewProhibited) ||
           object_states.hasState(changeProhibited) ||
           object_states.hasState(pendingDelete) {
            return &EPPResult{RetCode:EPP_STATUS_PROHIBITS_OPERATION}
        }
    }

    /* period is ignored for now, always prolong for one year */
    new_exdate, err := dbreg.GetNewExdate(ctx.dbconn, domain_data.Expiration_date, 12)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    err = ctx.dbconn.Begin()
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()

    err = dbreg.RenewDomain(ctx.dbconn, domain_data.Id, domain_data.Expiration_date, new_exdate)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if ctx.serv.RGconf.ChargeOperations {
        err = dbreg.ChargeRenewOp(ctx.dbconn, domain_data.Id, ctx.session.Regid, domain_data.ZoneId, domain_data.Cur_time)
        if err != nil {
            if _, ok := err.(*dbreg.BillingFailure); ok {
                return &EPPResult{RetCode:EPP_BILLING_FAILURE}
            }
            glg.Error(err)
            return &EPPResult{RetCode:2500}
        }
    }

    if err := UpdateObjectStates(ctx.dbconn, domain_data.Id); err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    err = ctx.dbconn.Commit()
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    return &EPPResult{RetCode:EPP_OK}
}

func deleteUnlinkedContacts(ctx *EPPContext, registrant uint64) error {
    if err := UpdateObjectStates(ctx.dbconn, registrant); err != nil {
        return err
    }

    object_states, err := getObjectStates(ctx.dbconn, registrant)
    if err != nil {
        return err
    }
    if !object_states.hasState(stateLinked) {
        err = dbreg.DeleteContact(ctx.dbconn, registrant)
        if err != nil {
            return err
        }
    }

    return nil
}

func deleteUnlinkedHosts(ctx *EPPContext, hosts []dbreg.HostObj) error {
    for _, host := range hosts {
        if err := UpdateObjectStates(ctx.dbconn, host.Id); err != nil {
            return err
        }

        object_states, err := getObjectStates(ctx.dbconn, host.Id)
        if err != nil {
            return err
        }
        if !object_states.hasState(stateLinked) {
            err = dbreg.DeleteHost(ctx.dbconn, host.Id)
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func epp_domain_delete_impl(ctx *EPPContext, v *xml.DeleteObject) (*EPPResult) {
    domain := normalizeDomain(v.Name)
    glg.Info("Delete domain", domain)

    domain_data, object_states, cmd := get_domain_obj(ctx, domain, true)
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

    domain_hosts, err := dbreg.GetDomainHosts(ctx.dbconn, domain_data.Id)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    err = ctx.dbconn.Begin()
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()

    err = dbreg.DeleteDomain(ctx.dbconn, domain_data.Id)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }
    err = deleteUnlinkedContacts(ctx, domain_data.Registrant.Id)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }
    err = deleteUnlinkedHosts(ctx, domain_hosts)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    err = ctx.dbconn.Commit()
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    return &EPPResult{RetCode:EPP_OK}
}