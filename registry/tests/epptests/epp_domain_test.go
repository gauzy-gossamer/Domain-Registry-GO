package epptests

import (
    "testing"
    "context"
    "registry/server"
    "registry/epp"
    . "registry/epp/eppcom"
    "registry/xml"
    "github.com/jackc/pgtype"
)

func infoDomain(t *testing.T, eppc *epp.EPPContext, name string, retcode int, sessionid uint64) *InfoDomainData {
    info_domain := xml.InfoDomain{Name:name}
    cmd := xml.XMLCommand{CmdType:EPP_INFO_DOMAIN, Sessionid:sessionid}
    cmd.Content = &info_domain
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
    if retcode == EPP_OK {
        info := epp_res.Content.(*InfoDomainData)
        return info
    }
    return nil
}

func TestEPPDomain(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }
    test_domain := generateRandomDomain(zone)
    sessionid := fakeSession(t, serv, dbconn, regid)
    eppc := epp.NewEPPContext(serv)
    test_contact := getExistingContact(t, eppc, dbconn, regid, sessionid)

    _ = infoDomain(t, eppc, test_domain, EPP_AUTHENTICATION_ERR, 0)
    _ = infoDomain(t, eppc, test_domain, EPP_OBJECT_NOT_EXISTS, sessionid)

    createDomain(t, eppc, test_domain, test_contact + "?err", EPP_PARAM_VALUE_POLICY, sessionid)
    createDomain(t, eppc, test_domain, test_contact, EPP_OK, sessionid)

    createDomain(t, eppc, test_domain, test_contact, EPP_OBJECT_EXISTS, sessionid)

    _ = infoDomain(t, eppc, test_domain, EPP_OK, sessionid)

    deleteObject(t, eppc, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func TestEPPCheckDomain(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }
    test_domain := generateRandomDomain(zone)

    sessionid := fakeSession(t, serv, dbconn, regid)

    check_domain := xml.CheckObject{Names:[]string{test_domain, "a." + zone, "a-domain.nonexistant", "?domain." + zone}}
    cmd := xml.XMLCommand{CmdType:EPP_CHECK_DOMAIN, Sessionid:sessionid}
    cmd.Content = &check_domain

    eppc := epp.NewEPPContext(serv)
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ok")
    }

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func updateDomain(t *testing.T, eppc *epp.EPPContext, name string, registrant string, description []string,  retcode int, sessionid uint64) {
    update_domain := xml.UpdateDomain{Name:name, Registrant:registrant, Description:description}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_DOMAIN, Sessionid:sessionid}
    update_cmd.Content = &update_domain
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
}

func TestEPPUpdateDomain(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }
    sessionid := fakeSession(t, serv, dbconn, regid)

    eppc := epp.NewEPPContext(serv)

    test_contact := getExistingContact(t, eppc, dbconn, regid, sessionid)
    test_domain := generateRandomDomain(zone)

    createDomain(t, eppc, test_domain, test_contact, EPP_OK, sessionid)

    updateDomain(t, eppc, test_domain, "nonexistant-contact", []string{"description"}, EPP_PARAM_VALUE_POLICY, sessionid)

    org_handle := "TEST-" + server.GenerateRandString(8)
    create_org := getCreateContact(org_handle, CONTACT_ORG)
    createContact(t, eppc, create_org, EPP_OK, sessionid)

    updateDomain(t, eppc, test_domain, org_handle, []string{"description"}, EPP_OK, sessionid)

    deleteObject(t, eppc, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid)
    deleteObject(t, eppc, org_handle, EPP_DELETE_DOMAIN, EPP_OBJECT_NOT_EXISTS, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func updateDomainHosts(t *testing.T, eppc *epp.EPPContext, name string, add_hosts []string, rem_hosts []string, retcode int, sessionid uint64) {
    update_domain := xml.UpdateDomain{Name:name, AddHosts:add_hosts, RemHosts:rem_hosts}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_DOMAIN, Sessionid:sessionid}
    update_cmd.Content = &update_domain
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
}

func TestEPPDomainHosts(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

    eppc := epp.NewEPPContext(serv)

    test_contact := getExistingContact(t, eppc, dbconn, regid, sessionid)
    test_domain := generateRandomDomain(zone)

    createDomain(t, eppc, test_domain, test_contact, EPP_OK, sessionid)

    test_host := "ns1." + generateRandomDomain(zone)
    serv.RGconf.DomainMinHosts = 2

    updateDomainHosts(t, eppc, test_domain, []string{test_host}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)

    createHost(t, eppc, test_host, []string{}, EPP_OK, sessionid)
    updateDomainHosts(t, eppc, test_domain, []string{test_host}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)

    test_host2 := "ns2." + generateRandomDomain(zone)
    test_host3 := "ns3." + generateRandomDomain(zone)
    createHost(t, eppc, test_host2, []string{}, EPP_OK, sessionid)
    createHost(t, eppc, test_host3, []string{}, EPP_OK, sessionid)
    updateDomainHosts(t, eppc, test_domain, []string{test_host, test_host2, test_host3}, []string{}, EPP_OK, sessionid)
    updateDomainHosts(t, eppc, test_domain, []string{}, []string{"ns1.nonexistant.ru"}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateDomainHosts(t, eppc, test_domain, []string{}, []string{test_host3}, EPP_OK, sessionid)

    deleteObject(t, eppc, test_host, EPP_DELETE_HOST, EPP_LINKED_PROHIBITS_OPERATION, sessionid)
    deleteObject(t, eppc, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid)
    deleteObject(t, eppc, test_host3, EPP_DELETE_HOST, EPP_OK, sessionid)

    deleteObject(t, eppc, test_host, EPP_DELETE_HOST, EPP_OBJECT_NOT_EXISTS, sessionid)
    deleteObject(t, eppc, test_host2, EPP_DELETE_HOST, EPP_OBJECT_NOT_EXISTS, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func updateDomainStates(t *testing.T, eppc *epp.EPPContext, name string, add_states []string, rem_states []string, retcode int, sessionid uint64) {
    update_domain := xml.UpdateDomain{Name:name, AddStatus:add_states, RemStatus:rem_states}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_DOMAIN, Sessionid:sessionid}
    update_cmd.Content = &update_domain
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
}

func TestEPPDomainStatus(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

    eppc := epp.NewEPPContext(serv)
    test_contact := getExistingContact(t, eppc, dbconn, regid, sessionid)
    test_domain := generateRandomDomain(zone)

    createDomain(t, eppc, test_domain, test_contact, EPP_OK, sessionid)

    updateDomainStates(t, eppc, test_domain, []string{"clientUpdateProhibited","nonexistant"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateDomainStates(t, eppc, test_domain, []string{"clientUpdateProhibited"}, []string{}, EPP_OK, sessionid)

    /* already present */
    updateDomainStates(t, eppc, test_domain, []string{"clientUpdateProhibited"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)

    update_domain := xml.UpdateDomain{Name:test_domain, Description:[]string{"hello"}}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_DOMAIN, Sessionid:sessionid}
    update_cmd.Content = &update_domain
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &update_cmd)
    if epp_res.RetCode != EPP_STATUS_PROHIBITS_OPERATION {
        t.Error("should be ", EPP_STATUS_PROHIBITS_OPERATION, epp_res.Msg, epp_res.Errors)
    }

    updateDomainStates(t, eppc, test_domain, []string{}, []string{"clientUpdateProhibited"}, EPP_OK, sessionid)

    deleteObject(t, eppc, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func fakeExpdate(db *server.DBConn, domainid uint64) (string, error) {
    row := db.QueryRow("SELECT now() AT TIME ZONE 'UTC' + interval '30 day' ")
    var new_exdate pgtype.Timestamp
    err := row.Scan(&new_exdate)
    if err != nil {
        return "", err
    }

    _, err = db.Exec("UPDATE domain SET exdate = $1::timestamp WHERE id = $2::integer", new_exdate.Time, domainid)
    if err != nil {
        return "", err
    }
    err = epp.UpdateObjectStates(db, domainid)
    if err != nil {
        return "", err
    }
    return new_exdate.Time.UTC().Format("2006-01-02"), nil
}

func TestEPPDomainRenew(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

    eppc := epp.NewEPPContext(serv)

    test_contact := getExistingContact(t, eppc, dbconn, regid, sessionid)
    test_domain := generateRandomDomain(zone)

    createDomain(t, eppc, test_domain, test_contact, EPP_OK, sessionid)

    renew_domain := xml.RenewDomain{Name:test_domain, CurExpDate:"2020-01-01"}
    renew_cmd := xml.XMLCommand{CmdType:EPP_RENEW_DOMAIN, Sessionid:sessionid}
    renew_cmd.Content = &renew_domain
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &renew_cmd)
    if epp_res.RetCode != EPP_PARAM_VALUE_POLICY {
        t.Error("should be ", EPP_PARAM_VALUE_POLICY, epp_res.Msg, epp_res.Errors)
    }

    info_domain := xml.InfoDomain{Name:test_domain}
    cmd := xml.XMLCommand{CmdType:EPP_INFO_DOMAIN, Sessionid:sessionid}
    cmd.Content = &info_domain
    epp_res = eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ", EPP_OK, epp_res.Msg)
    }
    domain_data := epp_res.Content.(*InfoDomainData)
    cur_exdate := domain_data.Expiration_date.Time.UTC().Format("2006-01-02")

    renew_domain = xml.RenewDomain{Name:test_domain, CurExpDate:cur_exdate}
    renew_cmd.Content = &renew_domain
    epp_res = eppc.ExecuteEPPCommand(context.Background(), &renew_cmd)
    if epp_res.RetCode != EPP_STATUS_PROHIBITS_OPERATION {
        t.Error("should be ", EPP_STATUS_PROHIBITS_OPERATION, epp_res.Msg, epp_res.Errors)
    }

    cur_exdate, err = fakeExpdate(dbconn, domain_data.Id)
    if err != nil {
        t.Error("fake exdate failed", err)
    }

    renew_domain = xml.RenewDomain{Name:test_domain, CurExpDate:cur_exdate}
    renew_cmd.Content = &renew_domain
    epp_res = eppc.ExecuteEPPCommand(context.Background(), &renew_cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ", EPP_OK, epp_res.Msg, epp_res.Errors, cur_exdate)
    }

    deleteObject(t, eppc, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func transferDomain(t *testing.T, eppc *epp.EPPContext, name string, acid string, optype int, retcode int, sessionid uint64) {
    transfer_domain := xml.TransferDomain{Name:name, OP:optype, AcID:acid}
    cmd := xml.XMLCommand{CmdType:EPP_TRANSFER_DOMAIN, Sessionid:sessionid}
    cmd.Content = &transfer_domain
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg)
    }
}

func TestEPPDomainTransfer(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    regid2, reg_handle2, _, err := getRegistrarAndZone(dbconn, regid)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)
    sessionid2 := fakeSession(t, serv, dbconn, regid2)

    eppc := epp.NewEPPContext(serv)
    test_contact := getExistingContact(t, eppc, dbconn, regid, sessionid)
    test_domain := generateRandomDomain(zone)

    createDomain(t, eppc, test_domain, test_contact, EPP_OK, sessionid)

    _ = infoDomain(t, eppc, test_domain, EPP_AUTHORIZATION_ERR, sessionid2)
    _ = infoContact(t, eppc, test_contact, EPP_AUTHORIZATION_ERR, sessionid2)

    transferDomain(t, eppc, test_domain, reg_handle2, TR_REQUEST, EPP_OK, sessionid)
    transferDomain(t, eppc, test_domain, reg_handle2, TR_QUERY, EPP_OK, sessionid)
    transferDomain(t, eppc, test_domain, reg_handle2, TR_CANCEL, EPP_OK, sessionid)

    transferDomain(t, eppc, test_domain, reg_handle2, TR_REQUEST, EPP_OK, sessionid)

    /* should be allowed with a pending transfer */
    _ = infoDomain(t, eppc, test_domain, EPP_OK, sessionid2)
    _ = infoContact(t, eppc, test_contact, EPP_OK, sessionid2)

    poll_cmd := xml.XMLCommand{CmdType:EPP_POLL_REQ, Sessionid:sessionid2}
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &poll_cmd)
    /* should definitely exist */
    if epp_res.RetCode != EPP_POLL_ACK_MSG {
        t.Error("should be ", EPP_POLL_ACK_MSG, epp_res.RetCode)
    }
    poll_msg, ok := epp_res.Content.(*PollMessage)
    if !ok {
        t.Error("should be ok")
    }
    pollAck(t, eppc, poll_msg.Msgid, EPP_OK, sessionid2) 

    transferDomain(t, eppc, test_domain, "", TR_REJECT, EPP_OK, sessionid2)

    transferDomain(t, eppc, test_domain, reg_handle2, TR_REQUEST, EPP_OK, sessionid)
    transferDomain(t, eppc, test_domain, "", TR_APPROVE, EPP_OK, sessionid2)

    deleteObject(t, eppc, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid2)

    logoutSession(t, serv, dbconn, sessionid)
    logoutSession(t, serv, dbconn, sessionid2)
}
