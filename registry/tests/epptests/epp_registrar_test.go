package epptests

import (
    "testing"
    "context"
    "registry/server"
    "registry/epp"
    . "registry/epp/eppcom"
    "registry/xml"
)

func infoRegistrar(t *testing.T, eppc *epp.EPPContext, name string, retcode int, sessionid uint64) *InfoRegistrarData {
    info_registrar := xml.InfoObject{Name:name}
    cmd := xml.XMLCommand{CmdType:EPP_INFO_REGISTRAR, Sessionid:sessionid}
    cmd.Content = &info_registrar
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
    if retcode == EPP_OK {
	reg_info := epp_res.Content.(*InfoRegistrarData)
	return reg_info
    }
    return nil
}

func TestEPPRegistrar(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, reg_handle, _, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }
    sessionid := fakeSession(t, serv, dbconn, regid)
    eppc := epp.NewEPPContext(serv)

    /* no session */
    _ = infoRegistrar(t, eppc, reg_handle, EPP_AUTHENTICATION_ERR, 0)

    nonexistant_registrar := server.GenerateRandString(8)
    _ = infoRegistrar(t, eppc, nonexistant_registrar, EPP_OBJECT_NOT_EXISTS, sessionid)

    _ = infoRegistrar(t, eppc, reg_handle, EPP_OK, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func updateRegistrar(t *testing.T, eppc *epp.EPPContext, name string, update_registrar *xml.UpdateRegistrar, add_ips []string, rem_ips []string, retcode int, sessionid uint64) {
    if update_registrar == nil {
        update_registrar = &xml.UpdateRegistrar{Name:name, AddAddrs:add_ips, RemAddrs:rem_ips}
    }
    update_cmd := xml.XMLCommand{CmdType:EPP_UPDATE_REGISTRAR, Sessionid:sessionid}
    update_cmd.Content = update_registrar
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }
}

func findAddress(ipaddrs []string, addr string) bool {
    for _, addr_ := range ipaddrs {
        if addr_ == addr {
            return true
        }
    }
    return false
}

func TestEPPUpdateRegistrar(t *testing.T) {
    serv := prepareServer()

    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }
    defer dbconn.Close()

    regid, reg_handle, _, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        t.Error("failed to get registrar")
    }
    _, reg_handle2, _, err := getRegistrarAndZone(dbconn, regid)
    if err != nil {
        t.Error("failed to get registrar")
    }
    sessionid := fakeSession(t, serv, dbconn, regid)

    eppc := epp.NewEPPContext(serv)

    insert_addr := "127.1.0.1"

    reg_info := infoRegistrar(t, eppc, reg_handle, EPP_OK, sessionid)
    /* delete first if it already exists */
    if findAddress(reg_info.Addrs, insert_addr) {
        updateRegistrar(t, eppc, reg_handle, nil, []string{}, []string{insert_addr}, EPP_OK, sessionid)
    }

    updateRegistrar(t, eppc, reg_handle, nil, []string{"127.1122.0.1"}, []string{}, EPP_PARAM_VALUE_POLICY, sessionid)
    /* try to change info on another registrar */
    updateRegistrar(t, eppc, reg_handle2, nil, []string{"127.1.0.1"}, []string{}, EPP_AUTHORIZATION_ERR, sessionid)

    updateRegistrar(t, eppc, reg_handle, nil, []string{insert_addr}, []string{}, EPP_OK, sessionid)

    rand_string := server.GenerateRandString(8)

    set_www := "http://"+rand_string
    set_whois := "whois"+rand_string

    update_registrar := xml.UpdateRegistrar{Name:reg_handle, WWW:set_www, Whois:set_whois}
    updateRegistrar(t, eppc, reg_handle, &update_registrar, []string{}, []string{}, EPP_OK, sessionid)

    reg_info = infoRegistrar(t, eppc, reg_handle, EPP_OK, sessionid)
    if !findAddress(reg_info.Addrs, insert_addr) {
        t.Error("should be present", insert_addr, reg_info.Addrs)
    }
    if reg_info.WWW.String != set_www {
        t.Error("expected www ", set_www, reg_info.WWW)
    }
    updateRegistrar(t, eppc, reg_handle, nil, []string{}, []string{insert_addr}, EPP_OK, sessionid)

    err = serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}
