package epp

import (
    "registry/xml"
    "registry/epp/dbreg"
    . "registry/epp/eppcom"
    "github.com/jackc/pgx/v5"
    "github.com/kpango/glg"
)

func set_pending_transfer(ctx *EPPContext, domainid uint64) error {
    if _, err := dbreg.CreateObjectStateRequest(ctx.dbconn, domainid, pendingTransfer) ; err != nil {
        return err
    }
    if err := UpdateObjectStates(ctx.dbconn, domainid); err != nil {
        return err
    }
    return nil
}

func cancel_pending_transfer(ctx *EPPContext, domainid uint64) error {
    if _, err := dbreg.CancelObjectStateRequest(ctx.dbconn, domainid, pendingTransfer) ; err != nil {
        return err
    }
    if err := UpdateObjectStates(ctx.dbconn, domainid); err != nil {
        return err
    }
    return nil
}

func query_transfer_object(ctx *EPPContext, domain string, v *xml.TransferDomain) *EPPResult {
    domain_data, _, cmd := get_domain_obj(ctx, domain, false)
    if cmd != nil {
        return cmd
    }

    find_transfer := dbreg.FindTransferRequest{Domainid:domain_data.Id, ActiveOnly:true}
    transfer_obj, err := find_transfer.Exec(ctx.dbconn)
    if err != nil {
        if err == pgx.ErrNoRows {
            return &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
        }
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    if !ctx.session.System {
        if transfer_obj.AcID.Id.Get() != ctx.session.Regid && transfer_obj.ReID.Id.Get() != ctx.session.Regid {
            return &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
        }
    }
    transfer_obj.Domain = domain

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = transfer_obj
    return &res
}

/* cancel is called by the initiator */
func cancel_transfer_request(ctx *EPPContext, domain string, v *xml.TransferDomain) *EPPResult {
    domain_data, _, cmd := get_domain_obj(ctx, domain, false)
    if cmd != nil {
        return cmd
    }

    find_transfer := dbreg.FindTransferRequest{Domainid:domain_data.Id, ActiveOnly:true}
    transfer_obj, err := find_transfer.SetLock(true).Exec(ctx.dbconn)
    if err != nil {
        if err == pgx.ErrNoRows {
            return &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
        }
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    if transfer_obj.ReID.Id.Get() != ctx.session.Regid {
        return &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
    }

    err = ctx.dbconn.Begin()
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()
    err = dbreg.ChangeTransferRequestState(ctx.dbconn, transfer_obj.Id, dbreg.TrClientCancelled, ctx.session.Regid, transfer_obj.AcID.Id.Get())
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    if err = cancel_pending_transfer(ctx, domain_data.Id); err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if err = ctx.dbconn.Commit(); err !=nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    find_transfer = dbreg.FindTransferRequest{Domainid:domain_data.Id, TrID:transfer_obj.Id}
    transfer_obj, err = find_transfer.Exec(ctx.dbconn)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    transfer_obj.Domain = domain

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = transfer_obj
    return &res
}

/* reject is called by the recipient */
func reject_transfer_request(ctx *EPPContext, domain string, v *xml.TransferDomain) *EPPResult {
    domain_id, err := dbreg.GetDomainIdByName(ctx.dbconn, domain)
    if err != nil {
        if perr, ok := err.(*dbreg.ParamError); ok {
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
        }
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    find_transfer := dbreg.FindTransferRequest{Domainid:domain_id, ActiveOnly:true}
    transfer_obj, err := find_transfer.SetLock(true).Exec(ctx.dbconn)
    if err != nil {
        if err == pgx.ErrNoRows {
            return &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
        }
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    if transfer_obj.AcID.Id.Get() != ctx.session.Regid {
        return &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
    }

    if err = ctx.dbconn.Begin(); err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()
    err = dbreg.ChangeTransferRequestState(ctx.dbconn, transfer_obj.Id, dbreg.TrClientRejected, ctx.session.Regid, transfer_obj.ReID.Id.Get())
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    if err = cancel_pending_transfer(ctx, domain_id); err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if err = ctx.dbconn.Commit(); err !=nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    find_transfer = dbreg.FindTransferRequest{Domainid:domain_id, TrID:transfer_obj.Id}
    transfer_obj, err = find_transfer.Exec(ctx.dbconn)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }
    transfer_obj.Domain = domain

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = transfer_obj
    return &res
}

func approve_transfer_request(ctx *EPPContext, domain string, v *xml.TransferDomain) *EPPResult {
    /* call info domain with a for update lock instead */
    domain_id, err := dbreg.GetDomainIdByName(ctx.dbconn, domain)
    if err != nil {
        if perr, ok := err.(*dbreg.ParamError); ok {
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
        }
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    find_transfer := dbreg.FindTransferRequest{Domainid:domain_id, ActiveOnly:true}
    transfer_obj, err := find_transfer.SetLock(true).Exec(ctx.dbconn)
    if err != nil {
        if err == pgx.ErrNoRows {
            return &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
        }
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if transfer_obj.AcID.Id.Get() != ctx.session.Regid {
        return &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
    }

    if err = ctx.dbconn.Begin(); err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }
    defer ctx.dbconn.Rollback()

    /* this is an incomplete transfer, we also need to transfer or copy linked objects (contact, hosts) */
    err = dbreg.TransferDomain(ctx.dbconn, domain_id, ctx.session.Regid)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    err = dbreg.ChangeTransferRequestState(ctx.dbconn, transfer_obj.Id, dbreg.TrClientApproved, ctx.session.Regid, transfer_obj.ReID.Id.Get())
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if err = cancel_pending_transfer(ctx, domain_id); err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if err = ctx.dbconn.Commit(); err !=nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    var res = EPPResult{RetCode:EPP_OK}
//    res.Content = transfer_obj
    return &res
}

func create_transfer_request(ctx *EPPContext, domain string, v *xml.TransferDomain) *EPPResult {
    domain_data, object_states, cmd := get_domain_obj(ctx, domain, false)
    if cmd != nil {
        return cmd
    }

    if !ctx.session.System {
        if object_states.hasState(serverTransferProhibited) ||
           object_states.hasState(clientTransferProhibited) ||
           object_states.hasState(changeProhibited) ||
           object_states.hasState(pendingDelete) {
            return &EPPResult{RetCode:EPP_STATUS_PROHIBITS_OPERATION}
        }
    }

    acquirer, err := dbreg.GetRegistrarByHandle(ctx.dbconn, v.AcID)
    if err != nil {
        if perr, ok := err.(*dbreg.ParamError); ok {
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
        }
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    /* acquirer registrar doesn't have permissions for the zone in which domain is registered */
    if !testRegistrarZoneAccess(ctx.dbconn, acquirer.Id.Get(), domain_data.ZoneId) {
        return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{"acID doesn't have permissions to use this zone"}}
    }

    /* transfer request already exists */
    /* we can probably test domain status (pendingTransfer) as well */
    find_transfer := dbreg.FindTransferRequest{Domainid:domain_data.Id, ActiveOnly:true}
    _, err = find_transfer.Exec(ctx.dbconn)
    if err == nil {
        return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{"transfer request already exists"}}
    }

    cr_transfer := dbreg.CreateTransferRequest{}

    cr_transfer.SetParams(ctx.session.Regid, acquirer.Id.Get(), domain_data.Id)

    if err = ctx.dbconn.Begin() ; err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()

    tr_request_id, err := cr_transfer.Exec(ctx.dbconn, acquirer.Id.Get())
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if err = set_pending_transfer(ctx, domain_data.Id); err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if err = ctx.dbconn.Commit(); err !=nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    find_transfer = dbreg.FindTransferRequest{Domainid:domain_data.Id, TrID:tr_request_id}
    transfer_obj, err := find_transfer.Exec(ctx.dbconn)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    transfer_obj.Domain = domain

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = transfer_obj
    return &res
}

func epp_domain_transfer_impl(ctx *EPPContext, v *xml.TransferDomain) (*EPPResult) {
    glg.Info("Domain transfer", v.Name)
    domain := normalizeDomain(v.Name)

    switch v.OP {
        case TR_REQUEST:
            return create_transfer_request(ctx, domain, v)
        case TR_QUERY:
            return query_transfer_object(ctx, domain, v)
        case TR_CANCEL:
            return cancel_transfer_request(ctx, domain, v)
        case TR_REJECT:
            return reject_transfer_request(ctx, domain, v)
        case TR_APPROVE:
            return approve_transfer_request(ctx, domain, v)
        default:
            return &EPPResult{CmdType:EPP_UNKNOWN_CMD, RetCode:EPP_UNKNOWN_ERR}
    }
}
