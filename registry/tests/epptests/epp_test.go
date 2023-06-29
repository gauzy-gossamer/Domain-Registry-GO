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

func getRegistrar(db *server.DBConn) (uint, string, string, string, error) {
    row := db.QueryRow("SELECT r.id, handle, cert, password FROM registrar r " +
                       "JOIN registraracl rc on r.id=rc.registrarid WHERE r.system = 'f' LIMIT 1;")
    var regid uint
    var reg_handle string
    var cert string
    var password string
    err := row.Scan(&regid, &reg_handle, &cert, &password)

    return regid, reg_handle, cert, password, err
}

func TestEPPLogin(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    serv.Sessions.InitSessions(dbconn)

    _, handle, cert, password, err := getRegistrar(dbconn)
    if err != nil {
        t.Error("failed to get registrar")
        return
    }
    login_cmd := xml.XMLCommand{CmdType:EPP_LOGIN} 

    eppc := epp.NewEPPContext(serv)

    login_cmd.Content = &xml.EPPLogin{Clid:handle, PW:password + "err", Lang:LANG_EN, Fingerprint:cert}
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &login_cmd)
    if epp_res.RetCode != EPP_AUTHENTICATION_ERR {
        t.Error(epp_res.Msg)
    }

    login_cmd.Content = &xml.EPPLogin{Clid:handle, PW:password, Lang:LANG_EN, Fingerprint:cert + "err"}
    epp_res = eppc.ExecuteEPPCommand(context.Background(), &login_cmd)
    if epp_res.RetCode != EPP_AUTHENTICATION_ERR {
        t.Error(epp_res.Msg)
    }

    login_cmd.Content = &xml.EPPLogin{Clid:handle + "err", PW:password, Lang:LANG_EN, Fingerprint:cert}
    epp_res = eppc.ExecuteEPPCommand(context.Background(), &login_cmd)
    if epp_res.RetCode != EPP_AUTHENTICATION_ERR {
        t.Error(epp_res.Msg)
    }

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
    sessionid := login_obj.Sessionid

    logout_cmd := xml.XMLCommand{CmdType:EPP_LOGOUT, Sessionid:sessionid}

    epp_res = eppc.ExecuteEPPCommand(context.Background(), &logout_cmd)
    if epp_res.RetCode != EPP_CLOSING_LOGOUT {
        t.Error(epp_res.Msg)
    }
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
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, _, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

    eppc := epp.NewEPPContext(serv)

    poll_cmd := xml.XMLCommand{CmdType:EPP_POLL_REQ, Sessionid:sessionid}
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &poll_cmd)
    /* if no message, create it */
    if epp_res.RetCode == EPP_POLL_NO_MSG {
        _, err = dbreg.CreatePollMessage(dbconn, regid, POLL_LOW_CREDIT)
        if err != nil {
            t.Error(err)
        }
        epp_res = eppc.ExecuteEPPCommand(context.Background(), &poll_cmd)
    }
    if epp_res.RetCode != EPP_POLL_ACK_MSG {
        t.Error("should be ", EPP_POLL_ACK_MSG, epp_res.RetCode)
    }
    poll_msg, ok := epp_res.Content.(*PollMessage)
    if !ok {
        t.Error("should be ok")
    }
    pollAck(t, eppc, poll_msg.Msgid, EPP_OK, sessionid) 

    logoutSession(t, serv, dbconn, sessionid)
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
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, _, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

    eppc := epp.NewEPPContext(serv)

    test_contact := getExistingContact(t, eppc, dbconn, regid, sessionid)

    info_contact := xml.InfoContact{Name:test_contact}
    cmd := xml.XMLCommand{CmdType:EPP_INFO_CONTACT, Sessionid:sessionid}
    cmd.Content = &info_contact
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ok", epp_res.RetCode)
    }

    org_handle := "TEST-" + server.GenerateRandString(8)
    create_org := getCreateContact(org_handle, CONTACT_ORG)
    createContact(t, eppc, create_org, EPP_OK, sessionid)

    person_handle := "TEST-" + server.GenerateRandString(8)
    create_contact := getCreateContact(person_handle, CONTACT_PERSON)
    createContact(t, eppc, create_contact, EPP_OK, sessionid)

    info_contact = xml.InfoContact{Name:person_handle}
    cmd = xml.XMLCommand{CmdType:EPP_INFO_CONTACT, Sessionid:sessionid}
    cmd.Content = &info_contact
    epp_res = eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ok", epp_res.RetCode)
    }
    create_contact.Fields.ContactId = ""
    info_return := epp_res.Content.(*InfoContactData).ContactFields
    if !reflect.DeepEqual(create_contact.Fields, info_return) {
        t.Error("create and info are not equal ", info_return, create_contact.Fields)
    }

    fields := ContactFields{ContactId:person_handle}
    fields.Verified.Set(true)
    updateContact(t, eppc, fields, EPP_OK, sessionid)

    deleteObject(t, eppc, org_handle, EPP_DELETE_CONTACT, EPP_OK, sessionid)
    deleteObject(t, eppc, person_handle, EPP_DELETE_CONTACT, EPP_OK, sessionid)

    logoutSession(t, serv, dbconn, sessionid)
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
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, _, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

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

    logoutSession(t, serv, dbconn, sessionid)
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

func TestEPPHost(t *testing.T) {
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

    test_host := "ns1." + generateRandomDomain(zone) 
    non_subordinate_host := "ns1." + generateRandomDomain("nonexistant.ru")

    info_host := xml.InfoHost{Name:test_host}
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
//    updateHost(t, serv, test_host, []string{}, []string{"127.0.0.1"}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateHost(t, eppc, test_host, []string{"127.0.0.1"}, []string{}, EPP_OK, sessionid)
    updateHost(t, eppc, non_subordinate_host, []string{"127.0.0.1"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateHost(t, eppc, test_host, []string{}, []string{"127.0.0.1"}, EPP_OK, sessionid)

    deleteObject(t, eppc, test_host, EPP_DELETE_HOST, EPP_OK, sessionid)
    deleteObject(t, eppc, non_subordinate_host, EPP_DELETE_HOST, EPP_OK, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
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

    eppc := epp.NewEPPContext(serv)

    sessionid := fakeSession(t, serv, dbconn, regid)

    test_host := "ns1." + generateRandomDomain(zone)

    createHost(t, eppc, test_host, []string{}, EPP_OK, sessionid)

    updateHostStates(t, eppc, test_host, []string{"clientUpdateProhibited","nonexistant"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    /* state allowed only for domains */
    updateHostStates(t, eppc, test_host, []string{"clientHold"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateHostStates(t, eppc, test_host, []string{"clientUpdateProhibited"}, []string{}, EPP_OK, sessionid)

    updateHost(t, eppc, test_host, []string{"10.10.0.1"}, []string{}, EPP_STATUS_PROHIBITS_OPERATION, sessionid)

    updateHostStates(t, eppc, test_host, []string{}, []string{"clientUpdateProhibited"}, EPP_OK, sessionid)

    deleteObject(t, eppc, test_host, EPP_DELETE_HOST, EPP_OK, sessionid)

    logoutSession(t, serv, dbconn, sessionid)
}
