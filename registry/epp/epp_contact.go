package epp

import (
    "strings"
    "registry/xml"
    "registry/epp/dbreg"
    . "registry/epp/eppcom"
    "github.com/jackc/pgx/v5"
    "github.com/kpango/glg"
)

func get_contact_object(ctx *EPPContext, contact_handle string) (*InfoContactData, *ObjectStates, *EPPResult) {
    info_db := dbreg.NewInfoContactDB()
    contact_data, err := info_db.SetName(contact_handle).Exec(ctx.dbconn)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil, &EPPResult{RetCode:EPP_OBJECT_NOT_EXISTS}
        }
        glg.Error(err)
        return nil, nil, &EPPResult{RetCode:EPP_FAILED}
    }

    if !ctx.session.System {
        if contact_data.Sponsoring_registrar.Id.Get() != ctx.session.Regid {
            return nil, nil, &EPPResult{RetCode:EPP_AUTHORIZATION_ERR}
        }
    }

    object_states, err := getObjectStates(ctx.dbconn, contact_data.Id)
    if err != nil {
        glg.Error(err)
        return nil,nil, &EPPResult{RetCode:EPP_FAILED}
    }

    return contact_data, object_states, nil
}

func epp_contact_info_impl(ctx *EPPContext, v *xml.InfoContact) (*EPPResult) {
    glg.Info("Info contact", v.Name)
    contact_handle := strings.ToLower(v.Name)
    contact_data, object_states, cmd := get_contact_object(ctx, contact_handle)
    if cmd != nil {
        return cmd
    }

    contact_data.States = object_states.copyObjectStates()

    var res = EPPResult{RetCode:EPP_OK}
    res.Content = contact_data
    return &res
}

func epp_contact_create_impl(ctx *EPPContext, v *xml.CreateContact) (*EPPResult) {
    contact_handle := strings.ToLower(v.Fields.ContactId)
    glg.Info("Create contact", contact_handle)

    if !checkContactHandleValidity(contact_handle) {
        return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY}
    }

    if ok, err := isContactAvailable(ctx.dbconn, contact_handle); !ok {
        if err != nil {
            glg.Error(err)
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
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()

    create_contact := dbreg.NewCreateContactDB()
    create_contact.SetEmails(v.Fields.Emails)
    create_contact.SetVoice(v.Fields.Voice)
    create_contact.SetFax(v.Fields.Fax)
    create_contact.SetBirthday(v.Fields.Birthday)
    create_contact.SetIntPostal(v.Fields.IntPostal)
    create_contact.SetVerified(v.Fields.Verified)
    create_result, err := create_contact.SetParams(contact_handle, ctx.session.Regid, v.Fields.ContactType).Exec(ctx.dbconn)
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
    res.Content = create_result
    return &res
}

func epp_contact_update_impl(ctx *EPPContext, v *xml.UpdateContact) (*EPPResult) {
    contact_handle := strings.ToLower(v.Fields.ContactId)
    glg.Info("Update contact", contact_handle)
    contact_data, object_states, cmd := get_contact_object(ctx, contact_handle)
    if cmd != nil {
        return cmd
    }

    if !ctx.session.System {
        if object_states.hasState(serverUpdateProhibited) ||
           object_states.hasState(clientUpdateProhibited) ||
           object_states.hasState(pendingDelete) {
            return &EPPResult{RetCode:EPP_STATUS_PROHIBITS_OPERATION}
        }
    }

    if v.Fields.ContactType != contact_data.ContactType {
        return &EPPResult{RetCode:EPP_PARAM_VALUE_POLICY, Errors:[]string{"incorrect contact type"}}
    }

    update_contact := dbreg.NewUpdateContactDB()
    if len(v.Fields.Emails) > 0 {
        update_contact.SetEmails(v.Fields.Emails)
    }
    if len(v.Fields.Voice) > 0 {
        update_contact.SetVoice(v.Fields.Voice)
    }
/*
    if len(v.Fields.Fax) > 0 {
        update_contact.SetFax(v.Fields.Fax)
    }
*/
//  should be nullable
//    create_contact.SetVerified(v.Fields.Verified)

    err := ctx.dbconn.Begin()
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()

    err = update_contact.Exec(ctx.dbconn, contact_data.Id, ctx.session.Regid)
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
    return &res
}

func epp_contact_delete_impl(ctx *EPPContext, v *xml.DeleteObject) *EPPResult {
    contact_handle := strings.ToLower(v.Name)
    glg.Info("Delete contact", contact_handle)
    contact_data, object_states, cmd := get_contact_object(ctx, contact_handle)
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
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }
    defer ctx.dbconn.Rollback()

    err = dbreg.DeleteContact(ctx.dbconn, contact_data.Id)
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:EPP_FAILED}
    }

    err = ctx.dbconn.Commit()
    if err != nil {
        glg.Error(err)
        return &EPPResult{RetCode:2500}
    }

    return &EPPResult{RetCode:EPP_OK}
}
