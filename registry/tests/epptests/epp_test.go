/*
go test -coverpkg=./... -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
*/
/*
these tests assume a working postgres database and at least two registrars with a shared zone
*/
package epptests

import (
    "testing"
    "context"
    "fmt"
    "reflect"
    "registry/server"
    "registry/epp"
    "registry/epp/dbreg"
    . "registry/epp/eppcom"
    "registry/xml"
)

func getRegistrar(db *server.DBConn, system bool) (uint, string, string, string, error) {
    row := db.QueryRow("SELECT r.id, handle, cert, password FROM registrar r " +
                       "JOIN registraracl rc on r.id=rc.registrarid WHERE r.system = $1 LIMIT 1;", system)
    var regid uint
    var reg_handle string
    var cert string
    var password string
    err := row.Scan(&regid, &reg_handle, &cert, &password)

    return regid, reg_handle, cert, password, err
}

func logoutCommand(t *testing.T, eppc *epp.EPPContext, sessionid uint64) {
    logout_cmd := xml.XMLCommand{CmdType:EPP_LOGOUT, Sessionid:sessionid}

    epp_res := eppc.ExecuteEPPCommand(context.Background(), &logout_cmd)
    if epp_res.RetCode != EPP_CLOSING_LOGOUT {
        t.Error(epp_res.Msg)
    }
}

func TestEPPLogin(t *testing.T) {
    tester := NewEPPTester()
    serv := tester.GetServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    serv.Sessions.InitSessions(dbconn)

    _, handle, cert, password, err := getRegistrar(dbconn, false)
    if err != nil {
        t.Error("failed to get registrar")
        return
    }
    login_cmd := xml.XMLCommand{CmdType:EPP_LOGIN} 

    eppc := epp.NewEPPContext(serv)

    /* incorrect password */
    login_cmd.Content = &xml.EPPLogin{Clid:handle, PW:password + "err", Lang:LANG_EN, Fingerprint:cert}
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &login_cmd)
    if epp_res.RetCode != EPP_AUTHENTICATION_ERR {
        t.Error(epp_res.Msg)
    }

    /* incorrect certificate */
    login_cmd.Content = &xml.EPPLogin{Clid:handle, PW:password, Lang:LANG_EN, Fingerprint:cert + "err"}
    epp_res = eppc.ExecuteEPPCommand(context.Background(), &login_cmd)
    if epp_res.RetCode != EPP_AUTHENTICATION_ERR {
        t.Error(epp_res.Msg)
    }

    /* incorrect registrar handle */
    login_cmd.Content = &xml.EPPLogin{Clid:handle + "err", PW:password, Lang:LANG_EN, Fingerprint:cert}
    epp_res = eppc.ExecuteEPPCommand(context.Background(), &login_cmd)
    if epp_res.RetCode != EPP_AUTHENTICATION_ERR {
        t.Error(epp_res.Msg)
    }

    /* must pass */
    login_cmd.Content = &xml.EPPLogin{Clid:handle, PW:password, Lang:LANG_EN, Fingerprint:cert}
    epp_res = eppc.ExecuteEPPCommand(context.Background(), &login_cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error(epp_res.Msg)
    }

    login_obj, ok := epp_res.Content.(*LoginResult)
    if !ok {
        t.Error("conversion error")
        return
    }

    logoutCommand(t, eppc, login_obj.Sessionid)
}

func TestEPPChangePassword(t *testing.T) {
    tester := NewEPPTester()
    serv := tester.GetServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    serv.Sessions.InitSessions(dbconn)

    new_password := server.GenerateRandString(8)

    _, handle, cert, password, err := getRegistrar(dbconn, true)
    if err != nil {
        t.Error("failed to get registrar")
        return
    }

    login_cmd := xml.XMLCommand{CmdType:EPP_LOGIN} 
    eppc := epp.NewEPPContext(serv)

    /* set new password */
    login_cmd.Content = &xml.EPPLogin{Clid:handle, PW:password, NewPW:new_password, Fingerprint:cert}
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &login_cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error(epp_res.Msg)
    }

    login_obj, ok := epp_res.Content.(*LoginResult)
    if !ok {
        t.Error("conversion error")
        return
    }

    logoutCommand(t, eppc, login_obj.Sessionid)

    /* change back */
    login_cmd.Content = &xml.EPPLogin{Clid:handle, PW:new_password, NewPW:password, Fingerprint:cert}
    epp_res = eppc.ExecuteEPPCommand(context.Background(), &login_cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error(epp_res.Msg)
    }

    login_obj, ok = epp_res.Content.(*LoginResult)
    if !ok {
        t.Error("conversion error")
        return
    }

    logoutCommand(t, eppc, login_obj.Sessionid)
}

func pollAck(t *testing.T, eppc *epp.EPPContext, msgid uint, retcode int, sessionid uint64) {
    poll_ack_cmd := xml.XMLCommand{CmdType:EPP_POLL_ACK, Sessionid:sessionid}
    poll_ack_cmd.Content = fmt.Sprint(msgid)
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &poll_ack_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.RetCode, epp_res.Msg)
    }
}

func TestEPPPoll(t *testing.T) {
    tester := NewEPPTester()
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }
    defer tester.CloseSession()
    regid := tester.Regid
    sessionid := tester.GetSessionid()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    eppc := epp.NewEPPContext(serv)

    poll_cmd := xml.XMLCommand{CmdType:EPP_POLL_REQ, Sessionid:sessionid}
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &poll_cmd)
    /* if no message, create it */
    if epp_res.RetCode == EPP_POLL_NO_MSG {
        _, err = dbreg.CreatePollMessage(dbconn, regid, dbreg.POLL_LOW_CREDIT)
        if err != nil {
            t.Error(err)
        }
        epp_res = eppc.ExecuteEPPCommand(context.Background(), &poll_cmd)
    }
    if epp_res.RetCode != EPP_POLL_ACK_MSG {
        t.Error("should be ", EPP_POLL_ACK_MSG, epp_res.RetCode)
    }
    if epp_res.MsgQ == nil {
        t.Error("should be ok")
    }
    pollAck(t, eppc, epp_res.MsgQ.Msgid, EPP_OK, sessionid) 
}

func infoContact(t *testing.T, eppc *epp.EPPContext, name string, retcode int, sessionid uint64) *InfoContactData {
    info_contact := xml.InfoObject{Name:name}
    cmd := xml.XMLCommand{CmdType:EPP_INFO_CONTACT, Sessionid:sessionid}
    cmd.Content = &info_contact
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
    if retcode == EPP_OK {
        info := epp_res.Content.(*InfoContactData)
        return info
    }
    return nil
}

func updateContact(t *testing.T, eppc *epp.EPPContext, fields ContactFields, retcode int, sessionid uint64) {
    update_contact := &xml.UpdateContact{Fields: fields}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_CONTACT, Sessionid:sessionid}
    update_cmd.Content = update_contact
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg)
    }
}

func TestEPPContact(t *testing.T) {
    tester := NewEPPTester()
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }
    defer tester.CloseSession()
    sessionid := tester.GetSessionid()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    eppc := epp.NewEPPContext(serv)

    test_contact := tester.GetExistingContact(t, eppc, dbconn)

    _ = infoContact(t, eppc, test_contact, EPP_OK, sessionid)

    org_handle := "TEST-" + server.GenerateRandString(8)
    create_org := getCreateContact(org_handle, CONTACT_ORG)
    createContact(t, eppc, create_org, EPP_OK, sessionid)

    person_handle := "TEST-" + server.GenerateRandString(8)
    create_contact := getCreateContact(person_handle, CONTACT_PERSON)
    createContact(t, eppc, create_contact, EPP_OK, sessionid)

    info_contact := infoContact(t, eppc, person_handle, EPP_OK, sessionid)
    info_return := info_contact.ContactFields

    create_contact.Fields.ContactId = ""
    if !reflect.DeepEqual(create_contact.Fields, info_return) {
        t.Error("create and info are not equal ", info_return, create_contact.Fields)
    }

    /* wrong contact type */
    updateContact(t, eppc, ContactFields{ContactId:person_handle, ContactType:CONTACT_ORG}, EPP_PARAM_VALUE_POLICY, sessionid)

    fields := ContactFields{ContactId:person_handle}
    fields.Verified.Set(true)
    fields.IntPostal = "new name"
    fields.IntAddress = []string{"new addr", "addr2"}
    fields.LocPostal = "new loc name"
    fields.LocAddress = []string{"new loc addr", "addr2"}
    fields.Passport = []string{"new passport"}
    fields.Birthday = "2000-01-04"

    updateContact(t, eppc, fields, EPP_OK, sessionid)

    create_contact.Fields.Verified.Set(true)
    create_contact.Fields.IntPostal = fields.IntPostal
    create_contact.Fields.IntAddress = fields.IntAddress
    create_contact.Fields.LocPostal = fields.LocPostal
    create_contact.Fields.LocAddress = fields.LocAddress
    create_contact.Fields.Passport = fields.Passport
    create_contact.Fields.Birthday = fields.Birthday

    info_contact = infoContact(t, eppc, person_handle, EPP_OK, sessionid)
    info_return = info_contact.ContactFields
    if !reflect.DeepEqual(create_contact.Fields, info_return) {
        t.Error("fields dont match after contact update ", info_return, create_contact.Fields)
    }

    deleteObject(t, eppc, org_handle, EPP_DELETE_CONTACT, EPP_OK, sessionid)
    deleteObject(t, eppc, person_handle, EPP_DELETE_CONTACT, EPP_OK, sessionid)
}

func TestEPPCheckContact(t *testing.T) {
    tester := NewEPPTester()
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }
    defer tester.CloseSession()
    sessionid := tester.GetSessionid()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }   
    defer dbconn.Close()

    eppc := epp.NewEPPContext(serv)

    test_contact := tester.GetExistingContact(t, eppc, dbconn)
    test_contact2 := "TEST-" + server.GenerateRandString(8)

    tests := map[string]int{test_contact:CD_REGISTERED, test_contact2:CD_AVAILABLE, "?domain.":CD_NOT_APPLICABLE}
    names := []string{}
    for name := range tests {
        names = append(names, name)
    }

    check_host := xml.CheckObject{Names:names}
    cmd := xml.XMLCommand{CmdType:EPP_CHECK_CONTACT, Sessionid:sessionid}
    cmd.Content = &check_host

    epp_res := eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ok")
    }
    testCheckResults(t, epp_res.Content, tests)
}

func updateContactStates(t *testing.T, eppc *epp.EPPContext, name string, add_states []string, rem_states []string, retcode int, sessionid uint64) {
    update_contact := xml.UpdateContact{Fields:ContactFields{ContactId:name}, AddStatus:add_states, RemStatus:rem_states}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_CONTACT, Sessionid:sessionid}
    update_cmd.Content = &update_contact
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
}

func TestEPPContactStates(t *testing.T) {
    tester := NewEPPTester()
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }
    defer tester.CloseSession()
    sessionid := tester.GetSessionid()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    eppc := epp.NewEPPContext(serv)

    test_contact := "TEST-" + server.GenerateRandString(8)
    create_org := getCreateContact(test_contact, CONTACT_PERSON)
    createContact(t, eppc, create_org, EPP_OK, sessionid)

    updateContactStates(t, eppc, test_contact, []string{"clientUpdateProhibited","nonexistant"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    /* state allowed only for domains */
    updateContactStates(t, eppc, test_contact, []string{"clientHold"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateContactStates(t, eppc, test_contact, []string{"clientUpdateProhibited"}, []string{}, EPP_OK, sessionid)

    fields := ContactFields{ContactId:test_contact}
    fields.Verified.Set(true)
    updateContact(t, eppc, fields, EPP_STATUS_PROHIBITS_OPERATION, sessionid)

    updateContactStates(t, eppc, test_contact, []string{}, []string{"clientUpdateProhibited"}, EPP_OK, sessionid)

    deleteObject(t, eppc, test_contact, EPP_DELETE_CONTACT, EPP_OK, sessionid)
}

func updateHost(t *testing.T, eppc *epp.EPPContext, name string, add_ips []string, rem_ips []string, retcode int, sessionid uint64) {
    update_host := xml.UpdateHost{Name:name, AddAddrs:add_ips, RemAddrs:rem_ips}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_HOST, Sessionid:sessionid}
    update_cmd.Content = &update_host
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
}

func renameHost(t *testing.T, eppc *epp.EPPContext, name string, new_name string, retcode int, sessionid uint64) {
    update_host := xml.UpdateHost{Name:name, NewName:new_name}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_HOST, Sessionid:sessionid}
    update_cmd.Content = &update_host
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
}

func TestEPPHost(t *testing.T) {
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

    test_host := "ns1." + generateRandomDomain(zone) 
    non_subordinate_host := "ns1." + generateRandomDomain("nonexistant.ru")

    info_host := xml.InfoObject{Name:test_host}
    cmd := xml.XMLCommand{CmdType:EPP_INFO_HOST, Sessionid:sessionid}
    cmd.Content = &info_host

    epp_res := eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != EPP_OBJECT_NOT_EXISTS {
        t.Error("should be ok", epp_res.RetCode)
    }

    createHost(t, eppc, test_host, []string{"127.88.88.88"}, EPP_OK, sessionid)
    createHost(t, eppc, non_subordinate_host, []string{"127.88.88.88"}, EPP_PARAM_VALUE_POLICY, sessionid)
    createHost(t, eppc, non_subordinate_host, []string{}, EPP_OK, sessionid)

    epp_res = eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ok", epp_res.RetCode)
    }

    updateHost(t, eppc, test_host, []string{"127.1110.0.1"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    /* nonexistant ip addr */
    updateHost(t, eppc, test_host, []string{}, []string{"127.0.0.1"}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateHost(t, eppc, test_host, []string{"127.0.0.1"}, []string{}, EPP_OK, sessionid)
    updateHost(t, eppc, non_subordinate_host, []string{"127.0.0.1"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateHost(t, eppc, test_host, []string{}, []string{"127.0.0.1"}, EPP_OK, sessionid)

    /* exceed number of allowed ips */
    add_ips := make([]string, serv.RGconf.MaxValueList + 1)
    for i := 0; i <= serv.RGconf.MaxValueList; i++ {
        add_ips = append(add_ips, fmt.Sprintf("127.0.1.%v", i + 1))
    }
    updateHost(t, eppc, test_host, add_ips, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)

    new_name := "ns1." + generateRandomDomain("nonexistant.ru")

    renameHost(t, eppc, non_subordinate_host, "ns1-p?", EPP_PARAM_VALUE_POLICY, sessionid)
    renameHost(t, eppc, non_subordinate_host, new_name, EPP_OK, sessionid)

    deleteObject(t, eppc, test_host, EPP_DELETE_HOST, EPP_OK, sessionid)
    deleteObject(t, eppc, non_subordinate_host, EPP_DELETE_HOST, EPP_OBJECT_NOT_EXISTS, sessionid)
    deleteObject(t, eppc, new_name, EPP_DELETE_HOST, EPP_OK, sessionid)
}

func testCheckResults(t *testing.T, content interface{}, tests map[string]int) {
    check_results, ok := content.([]CheckResult)
    if !ok {
        t.Error("conversion error")
        return
    }

    for _, check_result := range check_results {
        res, ok := tests[check_result.Name]
        if !ok {
            t.Error(check_result.Name, " not found")
            continue
        }
        if res != check_result.Result {
            t.Error("dont match ", res, check_result.Result)
        }
    }
}

func TestEPPCheckHost(t *testing.T) {
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

    test_host := generateRandomDomain(zone)

    tests := map[string]int{"ns1."+test_host:CD_AVAILABLE, "a." + zone:CD_AVAILABLE, "a-domain.nonexistant":CD_AVAILABLE, "?domain." + zone:CD_NOT_APPLICABLE}

    names := []string{}
    for name := range tests {
        names = append(names, name)
    }

    check_host := xml.CheckObject{Names:names}
    cmd := xml.XMLCommand{CmdType:EPP_CHECK_HOST, Sessionid:sessionid}
    cmd.Content = &check_host

    eppc := epp.NewEPPContext(serv)
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ok")
    }
    testCheckResults(t, epp_res.Content, tests)
}

func updateHostStates(t *testing.T, eppc *epp.EPPContext, name string, add_states []string, rem_states []string, retcode int, sessionid uint64) {
    update_host := xml.UpdateHost{Name:name, AddStatus:add_states, RemStatus:rem_states}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_HOST, Sessionid:sessionid}
    update_cmd.Content = &update_host
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
}

func TestEPPHostStates(t *testing.T) {
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

    test_host := "ns1." + generateRandomDomain(zone)

    createHost(t, eppc, test_host, []string{}, EPP_OK, sessionid)

    updateHostStates(t, eppc, test_host, []string{"clientUpdateProhibited","nonexistant"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    /* state allowed only for domains */
    updateHostStates(t, eppc, test_host, []string{"clientHold"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateHostStates(t, eppc, test_host, []string{"clientUpdateProhibited"}, []string{}, EPP_OK, sessionid)

    updateHost(t, eppc, test_host, []string{"10.10.0.1"}, []string{}, EPP_STATUS_PROHIBITS_OPERATION, sessionid)

    updateHostStates(t, eppc, test_host, []string{}, []string{"clientUpdateProhibited"}, EPP_OK, sessionid)

    deleteObject(t, eppc, test_host, EPP_DELETE_HOST, EPP_OK, sessionid)
}
