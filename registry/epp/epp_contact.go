package epp

import (
    "strings"
    "registry/xml"
    "registry/epp/dbreg"
    "registry/epp/dbreg/contact"
    . "registry/epp/eppcom"
    "github.com/jackc/pgx/v5"
)

func epp_contact_check_impl(ctx *EPPContext, v *xml.CheckObject) (*EPPResult) {
    ctx.logger.Info("Contact check", v.Names)

    var check_results []CheckResult

    for _, contact := range v.Names {
        contact_handle := strings.ToLower(contact)

        if !checkContactHandleValidity(contact_handle) {
            check_results = append(check_results, CheckResult{Name:contact, Result:CD_NOT_APPLICABLE})
            continue
        }

        if ok, err := isContactAvailable(ctx.dbconn, contact_handle); !ok {
            if err != nil {
                ctx.logger.Error(err)
                continue
            }
            check_results = append(check_results, CheckResult{Name:contact, Result:CD_REGISTERED})
            continue
        }

        check_results = append(check_results, CheckResult{Name:contact, Result:CD_AVAILABLE})
    }   

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = check_results
    return &res
}

func get_contact_object(ctx *EPPContext, contact_handle string, for_update bool) (*InfoContactData, *ObjectStates, *EPPResult) {
    info_db := contact.NewInfoContactDB()
    contact_data, err := info_db.SetLock(for_update).SetName(contact_handle).Exec(ctx.dbconn)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil, &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
        }
        ctx.logger.Error(err)
        return nil, nil, &EPPResult{RetCode:EPP_FAILED}
    }

    if for_update {
        if err := UpdateObjectStates(ctx.dbconn, contact_data.Id); err != nil {
            ctx.logger.Error(err)
            return nil,nil, &EPPResult{RetCode:EPP_FAILED}
        }
    }

    object_states, err := getObjectStates(ctx.dbconn, contact_data.Id)
    if err != nil {
        ctx.logger.Error(err)
        return nil,nil, &EPPResult{RetCode:EPP_FAILED}
    }

    return contact_data, object_states, nil
}

func allowContactInfoAccess(ctx *EPPContext, contact_data *InfoContactData) (bool, error) {
    if ctx.session.System {
        return true, nil
    }
    if contact_data.Sponsoring_registrar.Id.Get() == ctx.session.Regid {
        return true, nil
    }
    if exists, err := dbreg.CheckExistingTransferByContact(ctx.dbconn, contact_data.Id, ctx.session.Regid); exists || err != nil {
        if err != nil {
            return false, err
        }
        return true, nil
    }

    return false, nil
}

func epp_contact_info_impl(ctx *EPPContext, v *xml.InfoObject) (*EPPResult) {
    ctx.logger.Info("Info contact", v.Name)
    contact_handle := strings.ToLower(v.Name)
    contact_data, object_states, cmd := get_contact_object(ctx, contact_handle, false)
    if cmd != nil {
        return cmd
    }

    if allow, err := allowContactInfoAccess(ctx, contact_data); !allow || err != nil {
	if err != nil {
            return &EPPResult{RetCode:EPP_FAILED}
	}
	return &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
    }

    contact_data.States = object_states.copyObjectStates()

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = contact_data
    return &res
}

func epp_contact_create_impl(ctx *EPPContext, v *xml.CreateContact) (*EPPResult) {
    contact_handle := strings.ToLower(v.Fields.ContactId)
    ctx.logger.Info("Create contact", contact_handle)

    if !checkContactHandleValidity(contact_handle) {
        return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY}
    }

    if ok, err := isContactAvailable(ctx.dbconn, contact_handle); !ok {
        if err != nil {
            ctx.logger.Error(err)
            return &EPPResult{RetCode:EPP_FAILED}
        }
        return &EPPResult{RetCode:EPP_OBJECT_EXISTS}
    }

    if v.Fields.ContactType == CONTACT_PERSON {
        if !testDateValidity(v.Fields.Birthday) {
            return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{"incorrect birthday"}}
        }
    }

    err := ctx.dbconn.Begin()
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()

    create_contact := contact.NewCreateContactDB()
    create_contact.SetIntPostal(v.Fields.IntPostal)
    create_contact.SetIntAddress(v.Fields.IntAddress)
    create_contact.SetLocPostal(v.Fields.LocPostal)
    create_contact.SetLocAddress(v.Fields.LocAddress)

    create_contact.SetEmails(v.Fields.Emails)
    create_contact.SetVoice(v.Fields.Voice)

    if v.Fields.ContactType == CONTACT_ORG {
        create_contact.SetFax(v.Fields.Fax)
        create_contact.SetLegalAddress(v.Fields.LegalAddress)
        create_contact.SetTaxNumbers(v.Fields.TaxNumbers)

    } else {
        create_contact.SetBirthday(v.Fields.Birthday)
        create_contact.SetPassport(v.Fields.Passport)

    }
    create_contact.SetVerified(v.Fields.Verified.Get())
    create_result, err := create_contact.SetParams(contact_handle, ctx.session.Regid, v.Fields.ContactType).Exec(ctx.dbconn)
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }

    err = ctx.dbconn.Commit()
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = create_result
    return &res
}

func epp_contact_update_impl(ctx *EPPContext, v *xml.UpdateContact) (*EPPResult) {
    contact_handle := strings.ToLower(v.Fields.ContactId)
    ctx.logger.Info("Update contact", contact_handle)
    contact_data, object_states, cmd := get_contact_object(ctx, contact_handle, true)
    if cmd != nil {
        return cmd
    }

    err := ctx.dbconn.Begin()
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }
    defer ctx.dbconn.Rollback()

    if len(v.AddStatus) > 0 || len(v.RemStatus) > 0 {
        err := updateObjectClientStates(ctx, contact_data.Id, object_states, v.AddStatus, v.RemStatus, "contact")
        if err != nil {
            if perr, ok := err.(*dbreg.ParamError); ok {
                return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{perr.Val}}
            }
            ctx.logger.Error(err)
            return &EPPResult{RetCode:EPP_FAILED}
        }
    }

    if !ctx.session.System {
        if contact_data.Sponsoring_registrar.Id.Get() != ctx.session.Regid {
            return &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
        }
        if object_states.hasState(serverUpdateProhibited) ||
           object_states.hasState(clientUpdateProhibited) ||
           object_states.hasState(pendingDelete) {
            return &EPPResult{RetCode:EPP_STATUS_PROHIBITS_OPERATION}
        }
    }

    if v.Fields.ContactType != contact_data.ContactType {
        return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{"incorrect contact type"}}
    }

    update_contact := contact.NewUpdateContactDB()
    if len(v.Fields.IntPostal) > 0 {
        update_contact.SetIntPostal(v.Fields.IntPostal)
    }
    if len(v.Fields.IntAddress) > 0 {
        update_contact.SetIntAddress(v.Fields.IntAddress)
    }
    if len(v.Fields.LocPostal) > 0 {
        update_contact.SetLocPostal(v.Fields.LocPostal)
    }
    if len(v.Fields.LocAddress) > 0 {
        update_contact.SetLocAddress(v.Fields.LocAddress)
    }

    if len(v.Fields.LegalAddress) > 0 {
        update_contact.SetLegalAddress(v.Fields.LegalAddress)
    }
    if len(v.Fields.TaxNumbers) > 0 {
        update_contact.SetTaxNumbers(v.Fields.TaxNumbers)
    }

    if len(v.Fields.Birthday) > 0 {
        update_contact.SetBirthday(v.Fields.Birthday)
    }
    if len(v.Fields.Passport) > 0 {
        update_contact.SetPassport(v.Fields.Passport)
    }

    if len(v.Fields.Emails) > 0 {
        update_contact.SetEmails(v.Fields.Emails)
    }
    if len(v.Fields.Voice) > 0 {
        update_contact.SetVoice(v.Fields.Voice)
    }
    if len(v.Fields.Fax) > 0 {
        update_contact.SetFax(v.Fields.Fax)
    }
    if !v.Fields.Verified.IsNull() {
        update_contact.SetVerified(v.Fields.Verified.Get())
    }

    err = update_contact.Exec(ctx.dbconn, contact_data.Id, ctx.session.Regid)
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:2500}
    }

    err = ctx.dbconn.Commit()
    if err != nil {
        ctx.logger.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    var res = EPPResult{RetCode:EPP_OK}
    return &res
}

func epp_contact_delete_impl(ctx *EPPContext, v *xml.DeleteObject) *EPPResult {
    contact_handle := strings.ToLower(v.Name)
    ctx.logger.Info("Delete contact", contact_handle)
    contact_data, object_states, cmd := get_contact_object(ctx, contact_handle, true)
    if cmd != nil {
        return cmd
    }

    if !ctx.session.System {
        if contact_data.Sponsoring_registrar.Id.Get() != ctx.session.Regid {
            return &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
        }
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

    err = contact.DeleteContact(ctx.dbconn, contact_data.Id)
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
