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
    "fmt"
    "reflect"
    "registry/server"
    "registry/epp"
    "registry/epp/dbreg"
    . "registry/epp/eppcom"
    "registry/xml"
    "github.com/jackc/pgtype"
)

func TestEPPDomain(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool)
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }
    test_domain := generateRandomDomain(zone)
    test_contact := getExistingContact(t, dbconn, regid)

    info_domain := xml.InfoDomain{Name:test_domain}
    cmd := xml.XMLCommand{CmdType:EPP_INFO_DOMAIN}
    cmd.Content = &info_domain

    epp_res := epp.ExecuteEPPCommand(serv, &cmd)
    if epp_res.RetCode != EPP_AUTHENTICATION_ERR {
        t.Error("should be auth error")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

    cmd.Sessionid = sessionid
    epp_res = epp.ExecuteEPPCommand(serv, &cmd)
    if epp_res.RetCode != EPP_OBJECT_NOT_EXISTS {
        t.Error("should be ok")
    }

    createDomain(t, serv, test_domain, test_contact, EPP_OK, sessionid)

    createDomain(t, serv, test_domain, test_contact, EPP_OBJECT_EXISTS, sessionid)

    epp_res = epp.ExecuteEPPCommand(serv, &cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ok", epp_res.RetCode)
    }

    deleteObject(t, serv, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func pollAck(t *testing.T, serv *server.Server, msgid uint, retcode int, sessionid uint64) {
    poll_ack_cmd := xml.XMLCommand{CmdType:EPP_POLL_ACK, Sessionid:sessionid}
    poll_ack_cmd.Content = fmt.Sprint(msgid)
    epp_res := epp.ExecuteEPPCommand(serv, &poll_ack_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.RetCode, epp_res.Msg)
    }
}

func TestEPPPoll(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool)
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, _, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

    poll_cmd := xml.XMLCommand{CmdType:EPP_POLL_REQ, Sessionid:sessionid}
    epp_res := epp.ExecuteEPPCommand(serv, &poll_cmd)
    /* if no message, create it */
    if epp_res.RetCode == EPP_POLL_NO_MSG {
        _, err = dbreg.CreatePollMessage(dbconn, regid, POLL_LOW_CREDIT)
        if err != nil {
            t.Error(err)
        }
        epp_res = epp.ExecuteEPPCommand(serv, &poll_cmd)
    }
    if epp_res.RetCode != EPP_POLL_ACK_MSG {
        t.Error("should be ", EPP_POLL_ACK_MSG, epp_res.RetCode)
    }
    poll_msg, ok := epp_res.Content.(*PollMessage)
    if !ok {
        t.Error("should be ok")
    }
    pollAck(t, serv, poll_msg.Msgid, EPP_OK, sessionid) 

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func updateDomainHosts(t *testing.T, serv *server.Server, name string, add_hosts []string, rem_hosts []string, retcode int, sessionid uint64) {
    update_domain := xml.UpdateDomain{Name:name, AddHosts:add_hosts, RemHosts:rem_hosts}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_DOMAIN, Sessionid:sessionid}
    update_cmd.Content = &update_domain
    epp_res := epp.ExecuteEPPCommand(serv, &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
}

func TestEPPDomainHosts(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool)
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

    test_contact := getExistingContact(t, dbconn, regid)
    test_domain := generateRandomDomain(zone)

    createDomain(t, serv, test_domain, test_contact, EPP_OK, sessionid)

    test_host := "ns1." + generateRandomDomain(zone)
    serv.RGconf.DomainMinHosts = 2

    updateDomainHosts(t, serv, test_domain, []string{test_host}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)

    createHost(t, serv, test_host, sessionid)
    updateDomainHosts(t, serv, test_domain, []string{test_host}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)

    test_host2 := "ns2." + generateRandomDomain(zone)
    test_host3 := "ns3." + generateRandomDomain(zone)
    createHost(t, serv, test_host2, sessionid)
    createHost(t, serv, test_host3, sessionid)
    updateDomainHosts(t, serv, test_domain, []string{test_host, test_host2, test_host3}, []string{}, EPP_OK, sessionid)
    updateDomainHosts(t, serv, test_domain, []string{}, []string{"ns1.nonexistant.ru"}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateDomainHosts(t, serv, test_domain, []string{}, []string{test_host3}, EPP_OK, sessionid)

    deleteObject(t, serv, test_host, EPP_DELETE_HOST, EPP_LINKED_PROHIBITS_OPERATION, sessionid)
    deleteObject(t, serv, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid)
    deleteObject(t, serv, test_host3, EPP_DELETE_HOST, EPP_OK, sessionid)

    deleteObject(t, serv, test_host, EPP_DELETE_HOST, EPP_OBJECT_NOT_EXISTS, sessionid)
    deleteObject(t, serv, test_host2, EPP_DELETE_HOST, EPP_OBJECT_NOT_EXISTS, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func updateDomainStates(t *testing.T, serv *server.Server, name string, add_states []string, rem_states []string, retcode int, sessionid uint64) {
    update_domain := xml.UpdateDomain{Name:name, AddStatus:add_states, RemStatus:rem_states}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_DOMAIN, Sessionid:sessionid}
    update_cmd.Content = &update_domain
    epp_res := epp.ExecuteEPPCommand(serv, &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
}

func TestEPPDomainStatus(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool)
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

    test_contact := getExistingContact(t, dbconn, regid)
    test_domain := generateRandomDomain(zone)

    createDomain(t, serv, test_domain, test_contact, EPP_OK, sessionid)

    updateDomainStates(t, serv, test_domain, []string{"clientUpdateProhibited","nonexistant"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateDomainStates(t, serv, test_domain, []string{"clientUpdateProhibited"}, []string{}, EPP_OK, sessionid)

    /* already present */
    updateDomainStates(t, serv, test_domain, []string{"clientUpdateProhibited"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)

    update_domain := xml.UpdateDomain{Name:test_domain, Description:[]string{"hello"}}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_DOMAIN, Sessionid:sessionid}
    update_cmd.Content = &update_domain
    epp_res := epp.ExecuteEPPCommand(serv, &update_cmd)
    if epp_res.RetCode != EPP_STATUS_PROHIBITS_OPERATION {
        t.Error("should be ", EPP_STATUS_PROHIBITS_OPERATION, epp_res.Msg, epp_res.Errors)
    }

    updateDomainStates(t, serv, test_domain, []string{}, []string{"clientUpdateProhibited"}, EPP_OK, sessionid)

    deleteObject(t, serv, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func fakeExpdate(db *server.DBConn, domainid uint64) (string, error) {
    row := db.QueryRow("SELECT now() AT TIME ZONE 'UTC' + interval '30 day' ")
    var new_exdate pgtype.Timestamp
    err := row.Scan(&new_exdate)

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

    dbconn, err := server.AcquireConn(serv.Pool)
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

    test_contact := getExistingContact(t, dbconn, regid)
    test_domain := generateRandomDomain(zone)

    createDomain(t, serv, test_domain, test_contact, EPP_OK, sessionid)

    renew_domain := xml.RenewDomain{Name:test_domain, CurExpDate:"2020-01-01"}
    renew_cmd := xml.XMLCommand{CmdType:EPP_RENEW_DOMAIN, Sessionid:sessionid}
    renew_cmd.Content = &renew_domain
    epp_res := epp.ExecuteEPPCommand(serv, &renew_cmd)
    if epp_res.RetCode != EPP_PARAM_VALUE_POLICY {
        t.Error("should be ", EPP_PARAM_VALUE_POLICY, epp_res.Msg, epp_res.Errors)
    }

    info_domain := xml.InfoDomain{Name:test_domain}
    cmd := xml.XMLCommand{CmdType:EPP_INFO_DOMAIN, Sessionid:sessionid}
    cmd.Content = &info_domain
    epp_res = epp.ExecuteEPPCommand(serv, &cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ", EPP_OK, epp_res.Msg)
    }
    domain_data := epp_res.Content.(*InfoDomainData)
    cur_exdate := domain_data.Expiration_date.Time.UTC().Format("2006-01-02")

    renew_domain = xml.RenewDomain{Name:test_domain, CurExpDate:cur_exdate}
    renew_cmd.Content = &renew_domain
    epp_res = epp.ExecuteEPPCommand(serv, &renew_cmd)
    if epp_res.RetCode != EPP_STATUS_PROHIBITS_OPERATION {
        t.Error("should be ", EPP_STATUS_PROHIBITS_OPERATION, epp_res.Msg, epp_res.Errors)
    }

    cur_exdate, err = fakeExpdate(dbconn, domain_data.Id)
    if err != nil {
        t.Error("fake exdate failed", err)
    }

    renew_domain = xml.RenewDomain{Name:test_domain, CurExpDate:cur_exdate}
    renew_cmd.Content = &renew_domain
    epp_res = epp.ExecuteEPPCommand(serv, &renew_cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ", EPP_OK, epp_res.Msg, epp_res.Errors, cur_exdate)
    }

    deleteObject(t, serv, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func transferDomain(t *testing.T, serv *server.Server, name string, acid string, optype int, retcode int, sessionid uint64) {
    transfer_domain := xml.TransferDomain{Name:name, OP:optype, AcID:acid}
    cmd := xml.XMLCommand{CmdType:EPP_TRANSFER_DOMAIN, Sessionid:sessionid}
    cmd.Content = &transfer_domain
    epp_res := epp.ExecuteEPPCommand(serv, &cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg)
    }
}

func TestEPPDomainTransfer(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool)
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

    test_contact := getExistingContact(t, dbconn, regid)
    test_domain := generateRandomDomain(zone)

    createDomain(t, serv, test_domain, test_contact, EPP_OK, sessionid)

    transferDomain(t, serv, test_domain, reg_handle2, TR_REQUEST, EPP_OK, sessionid)
    transferDomain(t, serv, test_domain, reg_handle2, TR_QUERY, EPP_OK, sessionid)
    transferDomain(t, serv, test_domain, reg_handle2, TR_CANCEL, EPP_OK, sessionid)

    transferDomain(t, serv, test_domain, reg_handle2, TR_REQUEST, EPP_OK, sessionid)

    sessionid2 := fakeSession(t, serv, dbconn, regid2)
    poll_cmd := xml.XMLCommand{CmdType:EPP_POLL_REQ, Sessionid:sessionid2}
    epp_res := epp.ExecuteEPPCommand(serv, &poll_cmd)
    /* should definitely exist */
    if epp_res.RetCode != EPP_POLL_ACK_MSG {
        t.Error("should be ", EPP_POLL_ACK_MSG, epp_res.RetCode)
    }
    poll_msg, ok := epp_res.Content.(*PollMessage)
    if !ok {
        t.Error("should be ok")
    }
    pollAck(t, serv, poll_msg.Msgid, EPP_OK, sessionid2) 

    transferDomain(t, serv, test_domain, "", TR_REJECT, EPP_OK, sessionid2)

    transferDomain(t, serv, test_domain, reg_handle2, TR_REQUEST, EPP_OK, sessionid)
    transferDomain(t, serv, test_domain, "", TR_APPROVE, EPP_OK, sessionid2)

    deleteObject(t, serv, test_domain, EPP_DELETE_DOMAIN, EPP_OK, sessionid2)

    logoutSession(t, serv, dbconn, sessionid)
    logoutSession(t, serv, dbconn, sessionid2)
}

func getCreateContact(contact_id string, contact_type int) *xml.CreateContact {
    var fields ContactFields
    if contact_type == CONTACT_ORG {
        fields = ContactFields{
            ContactId:contact_id,
            IntPostal:"Company Inc",
            LocPostal:"Company Inc",
            LegalAddress:[]string{"address"},
            Emails:[]string{"first@company.com"},
            Voice:[]string{"+9 000 99999"},
            Fax:[]string{},
            ContactType:CONTACT_ORG,
        }
    } else {
        fields = ContactFields{
            ContactId:contact_id,
            IntPostal:"Person",
            LocPostal:"Person",
            Emails:[]string{"first@company.com"},
            Voice:[]string{"+9 000 99999"},
            Birthday:"1998-01-01",
            ContactType:CONTACT_PERSON,
        }
    }

    fields.Verified.Set(false)
    return &xml.CreateContact{Fields: fields}
}

func TestEPPContact(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool)
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, _, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

    test_contact := getExistingContact(t, dbconn, regid)

    info_contact := xml.InfoContact{Name:test_contact}
    cmd := xml.XMLCommand{CmdType:EPP_INFO_CONTACT, Sessionid:sessionid}
    cmd.Content = &info_contact
    epp_res := epp.ExecuteEPPCommand(serv, &cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ok", epp_res.RetCode)
    }

    org_handle := "TEST-" + server.GenerateRandString(8)
    create_org := getCreateContact(org_handle, CONTACT_ORG)
    createContact(t, serv, create_org, EPP_OK, sessionid)

    person_handle := "TEST-" + server.GenerateRandString(8)
    create_contact := getCreateContact(person_handle, CONTACT_PERSON)
    createContact(t, serv, create_contact, EPP_OK, sessionid)

    info_contact = xml.InfoContact{Name:person_handle}
    cmd = xml.XMLCommand{CmdType:EPP_INFO_CONTACT, Sessionid:sessionid}
    cmd.Content = &info_contact
    epp_res = epp.ExecuteEPPCommand(serv, &cmd)
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
    update_contact := &xml.UpdateContact{Fields: fields}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_CONTACT, Sessionid:sessionid}
    update_cmd.Content = update_contact
    epp_res = epp.ExecuteEPPCommand(serv, &update_cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ", EPP_OK, epp_res.Msg)
    }

    deleteObject(t, serv, org_handle, EPP_DELETE_CONTACT, EPP_OK, sessionid)
    deleteObject(t, serv, person_handle, EPP_DELETE_CONTACT, EPP_OK, sessionid)

    logoutSession(t, serv, dbconn, sessionid)
}

func updateHost(t *testing.T, serv *server.Server, name string, add_ips []string, rem_ips []string, retcode int, sessionid uint64) {
    update_host := xml.UpdateHost{Name:name, AddAddrs:add_ips, RemAddrs:rem_ips}
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_HOST, Sessionid:sessionid}
    update_cmd.Content = &update_host
    epp_res := epp.ExecuteEPPCommand(serv, &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
}

func TestEPPHost(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool)
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, _, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }

    sessionid := fakeSession(t, serv, dbconn, regid)

    test_host := "ns1." + generateRandomDomain(zone) 
    non_subordinate_host := "ns1." + generateRandomDomain("nonexistant.ru")

    info_host := xml.InfoHost{Name:test_host}
    cmd := xml.XMLCommand{CmdType:EPP_INFO_HOST, Sessionid:sessionid}
    cmd.Content = &info_host

    epp_res := epp.ExecuteEPPCommand(serv, &cmd)
    if epp_res.RetCode != EPP_OBJECT_NOT_EXISTS {
        t.Error("should be ok", epp_res.RetCode)
    }

    createHost(t, serv, test_host, sessionid)
    createHost(t, serv, non_subordinate_host, sessionid)

    epp_res = epp.ExecuteEPPCommand(serv, &cmd)
    if epp_res.RetCode != EPP_OK {
        t.Error("should be ok", epp_res.RetCode)
    }

    updateHost(t, serv, test_host, []string{"127.1110.0.1"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
//    updateHost(t, serv, test_host, []string{}, []string{"127.0.0.1"}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateHost(t, serv, test_host, []string{"127.0.0.1"}, []string{}, EPP_OK, sessionid)
    updateHost(t, serv, non_subordinate_host, []string{"127.0.0.1"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    updateHost(t, serv, test_host, []string{}, []string{"127.0.0.1"}, EPP_OK, sessionid)

    deleteObject(t, serv, test_host, EPP_DELETE_HOST, EPP_OK, sessionid)
    deleteObject(t, serv, non_subordinate_host, EPP_DELETE_HOST, EPP_OK, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}
