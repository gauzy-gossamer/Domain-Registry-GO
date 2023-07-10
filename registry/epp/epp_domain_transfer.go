package epp

import (
    "registry/xml"
    "registry/server"
    "registry/epp/dbreg"
    "registry/epp/dbreg/contact"
    "registry/epp/dbreg/registrar"
    . "registry/epp/eppcom"
    "github.com/jackc/pgx/v5"
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
        ctx.logger.Error(err)
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
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }

    if transfer_obj.ReID.Id.Get() != ctx.session.Regid {
        return &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
    }

    err = ctx.dbconn.Begin()
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()
    err = dbreg.ChangeTransferRequestState(ctx.dbconn, transfer_obj.Id, dbreg.TrClientCancelled, ctx.session.Regid, transfer_obj.AcID.Id.Get())
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }

    if err = cancel_pending_transfer(ctx, domain_data.Id); err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if err = ctx.dbconn.Commit(); err !=nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    find_transfer = dbreg.FindTransferRequest{Domainid:domain_data.Id, TrID:transfer_obj.Id}
    transfer_obj, err = find_transfer.Exec(ctx.dbconn)
    if err != nil {
        ctx.logger.Error(err)
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
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    find_transfer := dbreg.FindTransferRequest{Domainid:domain_id, ActiveOnly:true}
    transfer_obj, err := find_transfer.SetLock(true).Exec(ctx.dbconn)
    if err != nil {
        if err == pgx.ErrNoRows {
            return &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
        }
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }

    if transfer_obj.AcID.Id.Get() != ctx.session.Regid {
        return &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
    }

    if err = ctx.dbconn.Begin(); err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()
    err = dbreg.ChangeTransferRequestState(ctx.dbconn, transfer_obj.Id, dbreg.TrClientRejected, ctx.session.Regid, transfer_obj.ReID.Id.Get())
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }

    if err = cancel_pending_transfer(ctx, domain_id); err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if err = ctx.dbconn.Commit(); err !=nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    find_transfer = dbreg.FindTransferRequest{Domainid:domain_id, TrID:transfer_obj.Id}
    transfer_obj, err = find_transfer.Exec(ctx.dbconn)
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }
    transfer_obj.Domain = domain

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = transfer_obj
    return &res
}

/* generate new handle when copying contacts */
func generateNewContactHandle(ctx *EPPContext) (string, error) {
    for {
        contact_handle := "cid-" + server.GenerateRandString(8)
        avail, err := isContactAvailable(ctx.dbconn, contact_handle)
        if err != nil {
            return "", err
        }
        if avail {
            return contact_handle, nil
        }
    }
}

/* transfer or copy contacts & hosts linked to domain */
func transferDependableObjects(ctx *EPPContext, domain_data *InfoDomainData) error {
    new_contact, err := generateNewContactHandle(ctx)
    if err != nil {
        return err
    }

    domains_n, err := contact.GetNumberOfLinkedDomains(ctx.dbconn, domain_data.Registrant.Id)
    if err != nil {
        return err
    }

    ctx.logger.Error("number of linked domains", domains_n)
    /* if it's the only linked domain, then we can transfer contact to new registrar */
    if domains_n > 1 {
        err = contact.CopyContact(ctx.dbconn, domain_data.Registrant.Id, new_contact, ctx.session.Regid)
        if err != nil {
            return err
        }
    } else {
        err = contact.TransferContact(ctx.dbconn, domain_data.Registrant.Id, ctx.session.Regid)
        if err != nil {
            return err
        }
    }

    return nil
}

func approve_transfer_request(ctx *EPPContext, domain string, v *xml.TransferDomain) *EPPResult {
    var err error
    var res *EPPResult
    /* use serializable transaction to manage possible collisions with other transactions */
    ctx.dbconn.RetryTx(func() error {
        err = nil
        if err = ctx.dbconn.BeginSerializable(); err != nil {
            ctx.logger.Error(err)
            res = &EPPResult{RetCode:EPP_FAILED}
            return nil
        }
        defer ctx.dbconn.Rollback()

        info_db := dbreg.NewInfoDomainDB()
        domain_data, err := info_db.Set_fqdn(domain).Exec(ctx.dbconn)
        if err != nil {
            if err == pgx.ErrNoRows {
                res = &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
                return nil
            }
            return err
        }   
        domain_id := domain_data.Id

        find_transfer := dbreg.FindTransferRequest{Domainid:domain_id, ActiveOnly:true}
        transfer_obj, err := find_transfer.SetLock(true).Exec(ctx.dbconn)
        if err != nil {
            if err == pgx.ErrNoRows {
                res = &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
                return nil
            }
            return err
        }

        if transfer_obj.AcID.Id.Get() != ctx.session.Regid {
            res = &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
            return nil
        }

        err = dbreg.TransferDomain(ctx.dbconn, domain_id, ctx.session.Regid)
        if err != nil {
            return err
        }

        err = transferDependableObjects(ctx, domain_data)
        if err != nil {
            return err
        }

        err = dbreg.ChangeTransferRequestState(ctx.dbconn, transfer_obj.Id, dbreg.TrClientApproved, ctx.session.Regid, transfer_obj.ReID.Id.Get())
        if err != nil {
            return err
        }

        if err = cancel_pending_transfer(ctx, domain_id); err != nil {
            return err
        }

        if err = ctx.dbconn.Commit(); err != nil {
            return err
        }

        return nil
    })
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }
    if res != nil {
        return res
    }

    res = &EPPResult{RetCode:EPP_OK}
//    res.Content = transfer_obj
    return res
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

    acquirer, err := registrar.GetRegistrarByHandle(ctx.dbconn, v.AcID)
    if err != nil {
        if perr, ok := err.(*dbreg.ParamError); ok {
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
        }
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    /* acquirer registrar doesn't have permissions for the zone in which domain is registered */
    if ok, err := dbreg.TestRegistrarZoneAccess(ctx.dbconn, acquirer.Id.Get(), domain_data.ZoneId); !ok || err != nil {
        if err != nil {
            ctx.logger.Error(err)
            return &EPPResult{RetCode:EPP_FAILED}
        }
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
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()

    tr_request_id, err := cr_transfer.Exec(ctx.dbconn, acquirer.Id.Get())
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if err = set_pending_transfer(ctx, domain_data.Id); err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    if err = ctx.dbconn.Commit(); err !=nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    find_transfer = dbreg.FindTransferRequest{Domainid:domain_data.Id, TrID:tr_request_id}
    transfer_obj, err := find_transfer.Exec(ctx.dbconn)
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }

    transfer_obj.Domain = domain

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = transfer_obj
    return &res
}

func epp_domain_transfer_impl(ctx *EPPContext, v *xml.TransferDomain) (*EPPResult) {
    ctx.logger.Info("Domain transfer", v.Name)
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
