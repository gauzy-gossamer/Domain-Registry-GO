package epptests

import (
    "testing"
    "context"
    "registry/server"
    "registry/epp"
    "registry/maintenance"
    . "registry/epp/eppcom"
    "registry/xml"
)

func TestEPPDomain(t *testing.T) {
    tester := NewEPPTester()
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }   
    defer tester.CloseSession()
    zone := tester.Zone
    sessionid := tester.GetSessionid()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    test_domain := generateRandomDomain(zone)
    eppc := epp.NewEPPContext(serv)
    test_contact := tester.GetExistingContact(t, eppc, dbconn)

    _ = infoDomain(t, eppc, test_domain, EPP_AUTHENTICATION_ERR, 0)
    _ = infoDomain(t, eppc, test_domain, EPP_OBJECT_NOT_EXISTS, sessionid)

    createDomain(t, eppc, test_domain, test_contact + "?err", EPP_PARAM_VALUE_POLICY, sessionid)
    createDomain(t, eppc, test_domain, test_contact, EPP_OK, sessionid)

    createDomain(t, eppc, test_domain, test_contact, EPP_OBJECT_EXISTS, sessionid)

    _ = infoDomain(t, eppc, test_domain, EPP_OK, sessionid)

    deleteObject(t, eppc, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid)
}

func TestEPPCheckDomain(t *testing.T) {
    tester := NewEPPTester()
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }   
    defer tester.CloseSession()
    zone := tester.Zone
    sessionid := tester.GetSessionid()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    test_domain := generateRandomDomain(zone)

    check_domain := xml.CheckObject{Names:[]string{test_domain, "a." + zone, "a-domain.nonexistant", "?domain." + zone}}
    cmd := xml.XMLCommand{CmdType:EPP_CHECK_DOMAIN, Sessionid:sessionid}
    cmd.Content = &check_domain

    eppc := epp.NewEPPContext(serv)
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ok")
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
    tester := NewEPPTester()
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }   
    defer tester.CloseSession()
    zone := tester.Zone
    sessionid := tester.GetSessionid()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    eppc := epp.NewEPPContext(serv)

    test_contact := tester.GetExistingContact(t, eppc, dbconn)
    test_domain := generateRandomDomain(zone)

    createDomain(t, eppc, test_domain, test_contact, EPP_OK, sessionid)

    updateDomain(t, eppc, test_domain, "nonexistant-contact", []string{"description"}, EPP_PARAM_VALUE_POLICY, sessionid)

    org_handle := "TEST-" + server.GenerateRandString(8)
    create_org := getCreateContact(org_handle, CONTACT_ORG)
    createContact(t, eppc, create_org, EPP_OK, sessionid)

    updateDomain(t, eppc, test_domain, org_handle, []string{"description"}, EPP_OK, sessionid)

    deleteObject(t, eppc, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid)
    deleteObject(t, eppc, org_handle, EPP_DELETE_DOMAIN, EPP_OBJECT_NOT_EXISTS, sessionid)
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
    tester := NewEPPTester()
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }   
    defer tester.CloseSession()
    zone := tester.Zone
    sessionid := tester.GetSessionid()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    eppc := epp.NewEPPContext(serv)

    test_contact := tester.GetExistingContact(t, eppc, dbconn)
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
    tester := NewEPPTester()
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }   
    defer tester.CloseSession()
    zone := tester.Zone
    sessionid := tester.GetSessionid()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    eppc := epp.NewEPPContext(serv)
    test_contact := tester.GetExistingContact(t, eppc, dbconn)
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
}

func TestEPPDomainRenew(t *testing.T) {
    tester := NewEPPTester()
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }   
    defer tester.CloseSession()
    zone := tester.Zone
    sessionid := tester.GetSessionid()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    eppc := epp.NewEPPContext(serv)

    test_contact := tester.GetExistingContact(t, eppc, dbconn)
    test_domain := generateRandomDomain(zone)

    createDomain(t, eppc, test_domain, test_contact, EPP_OK, sessionid)

    renew_domain := xml.RenewDomain{Name:test_domain, CurExpDate:"2020-01-01"}
    renew_cmd := xml.XMLCommand{CmdType:EPP_RENEW_DOMAIN, Sessionid:sessionid}
    renew_cmd.Content = &renew_domain
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &renew_cmd)
    if epp_res.RetCode != EPP_PARAM_VALUE_POLICY {
        t.Error("should be ", EPP_PARAM_VALUE_POLICY, epp_res.Msg, epp_res.Errors)
    }

    domain_data := infoDomain(t, eppc, test_domain, EPP_OK, sessionid)
    cur_exdate := domain_data.Expiration_date.Time.UTC().Format("2006-01-02")

    renew_domain = xml.RenewDomain{Name:test_domain, CurExpDate:cur_exdate}
    renew_cmd.Content = &renew_domain
    epp_res = eppc.ExecuteEPPCommand(context.Background(), &renew_cmd)
    if epp_res.RetCode != EPP_STATUS_PROHIBITS_OPERATION {
        t.Error("should be ", EPP_STATUS_PROHIBITS_OPERATION, epp_res.Msg, epp_res.Errors)
    }

    cur_exdate, err = setProlongExpdate(dbconn, domain_data.Id)
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

/* test transfer with copying dependant objects */
func TestEPPDomainTransfer(t *testing.T) {
    tester := NewEPPTester()
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }   
    defer tester.CloseSession()
    regid := tester.Regid
    zone := tester.Zone
    sessionid := tester.GetSessionid()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid2, reg_handle2, _, err := getRegistrarAndZone(dbconn, regid)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid2 := fakeSession(t, serv, dbconn, regid2)

    eppc := epp.NewEPPContext(serv)
    test_contact := tester.GetExistingContact(t, eppc, dbconn)
    test_domain := generateRandomDomain(zone)
    test_domain2 := generateRandomDomain(zone)

    createDomain(t, eppc, test_domain, test_contact, EPP_OK, sessionid)
    createDomain(t, eppc, test_domain2, test_contact, EPP_OK, sessionid)

    /* access to a second registrar should not be allowed without a transfer */
    _ = infoDomain(t, eppc, test_domain, EPP_AUTHORIZATION_ERR, sessionid2)

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
    if epp_res.MsgQ == nil {
        t.Error("should be ok")
    }
    pollAck(t, eppc, epp_res.MsgQ.Msgid, EPP_OK, sessionid2) 

    transferDomain(t, eppc, test_domain, "", TR_REJECT, EPP_OK, sessionid2)

    transferDomain(t, eppc, test_domain, reg_handle2, TR_REQUEST, EPP_OK, sessionid)
    transferDomain(t, eppc, test_domain, "", TR_APPROVE, EPP_OK, sessionid2)

    deleteObject(t, eppc, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid2)
    deleteObject(t, eppc, test_domain2, EPP_DELETE_DOMAIN, EPP_OK, sessionid)

    logoutSession(t, serv, dbconn, sessionid2)
}

/* test transfer with transfering dependant objects */
func TestEPPDomainTransfer2(t *testing.T) {
    tester := NewEPPTester()
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }   
    defer tester.CloseSession()
    regid := tester.Regid
    zone := tester.Zone
    sessionid := tester.GetSessionid()

    logger := server.NewLogger("")
    dbconn, err := server.AcquireConn(serv.Pool, logger)
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid2, reg_handle2, _, err := getRegistrarAndZone(dbconn, regid)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid2 := fakeSession(t, serv, dbconn, regid2)

    eppc := epp.NewEPPContext(serv)

    /* create contact that will be transfered */
    test_contact := "TEST-" + server.GenerateRandString(8)
    create_org := getCreateContact(test_contact, CONTACT_ORG)
    createContact(t, eppc, create_org, EPP_OK, sessionid)
    test_domain := generateRandomDomain(zone)

    host_domain := generateRandomDomain("nonexistant.ru")
    test_host1 := "ns1." + host_domain
    test_host2 := "ns2." + host_domain
    createHost(t, eppc, test_host1, []string{}, EPP_OK, sessionid)
    createHost(t, eppc, test_host2, []string{}, EPP_OK, sessionid)

    /* access to a second registrar should not be allowed without a transfer */
    _ = infoContact(t, eppc, test_contact, EPP_AUTHORIZATION_ERR, sessionid2)

    createDomain(t, eppc, test_domain, test_contact, EPP_OK, sessionid)
    updateDomainHosts(t, eppc, test_domain, []string{test_host1, test_host2}, []string{}, EPP_OK, sessionid)

    transferDomain(t, eppc, test_domain, reg_handle2, TR_REQUEST, EPP_OK, sessionid)

    _, err = dbconn.Exec("UPDATE epp_transfer_request SET acdate = now() - interval '1 day' WHERE registrar_id=$1::integer and acquirer_id=$2::integer " +
                "and status = 0 and domain_id = (select id from object_registry where name = lower($3::text) and erdate is null)", regid, regid2, test_domain)
    if err != nil {
        t.Error("failed to update transfer request", err)
    }

    err = maintenance.FinishExpiredTransferRequests(serv, logger, dbconn)
    if err != nil {
        t.Error("failed close expired transfer request", err)
    }

    transferDomain(t, eppc, test_domain, reg_handle2, TR_QUERY, EPP_OBJECT_NOT_EXISTS, sessionid)

    transferDomain(t, eppc, test_domain, reg_handle2, TR_REQUEST, EPP_OK, sessionid)
    transferDomain(t, eppc, test_domain, "", TR_APPROVE, EPP_OK, sessionid2)

    deleteObject(t, eppc, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid2)

    logoutSession(t, serv, dbconn, sessionid2)
}
