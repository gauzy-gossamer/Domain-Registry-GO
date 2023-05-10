package epptests

import (
    "testing"
    "registry/server"
    "registry/epp"
    . "registry/epp/eppcom"
    "registry/xml"
)

func prepareServer() *server.Server {
    serv := server.Server{}
    serv.RGconf.LoadConfig("../../server.conf")
    var err error
    serv.Pool, err = server.CreatePool(&serv.RGconf.DBconf)
    if err != nil {
        panic(err)
    }
    serv.Sessions.SessionTimeoutSec = serv.RGconf.SessionTimeout
    serv.Sessions.MaxRegistrarSessions = serv.RGconf.MaxRegistrarSessions

    return &serv
}

func getRegistrarAndZone(db *server.DBConn, exclude_reg uint) (uint, string, string, error) {
    row := db.QueryRow("SELECT r.id, r.handle, z.fqdn FROM registrar r JOIN registrarinvoice ri on r.id=ri.registrarid "+
                       "JOIN zone z on ri.zone=z.id WHERE r.system = 'f' and r.id != $1::integer LIMIT 1;", exclude_reg)
    var regid uint
    var reg_handle string
    var zone string
    err := row.Scan(&regid, &reg_handle, &zone)

    return regid, reg_handle, zone, err
}

func fakeSession(t *testing.T, serv *server.Server, db *server.DBConn, regid uint) uint64 {
    serv.Sessions.InitSessions(db)
    sessionid, err := serv.Sessions.LoginSession(db, regid, LANG_EN)
    if err != nil {
        t.Error("failed login session")
    }

    return sessionid
}

func logoutSession(t *testing.T, serv *server.Server, dbconn *server.DBConn, sessionid uint64) {
    err := serv.Sessions.LogoutSession(dbconn, sessionid)
    if err != nil {
        t.Error("logout failed")
    }
}

func generateRandomDomain(zone string) string {
    rand_id := server.GenerateRandString(6)
    return rand_id + "." + zone
}

func createDomain(t *testing.T, serv *server.Server, name string, contact_name string, retcode int, sessionid uint64) {
    create_domain := xml.CreateDomain{Name:name, Registrant:contact_name}
    create_cmd := xml.XMLCommand{CmdType:EPP_CREATE_DOMAIN, Sessionid:sessionid}
    create_cmd.Content = &create_domain
    epp_res := epp.ExecuteEPPCommand(serv, &create_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.RetCode, epp_res.Msg)
    }
}

func getExistingContact(t *testing.T, db *server.DBConn, regid uint) string {
    row := db.QueryRow("SELECT name FROM object_registry obr " +
                       "JOIN object o on obr.id=o.id and o.clid = $1::integer " +
                       "WHERE erdate is null and type = get_object_type_id('contact'::text)", regid)
    var contact_name string
    err := row.Scan(&contact_name)
    if err != nil {
        t.Error(err)
    }
    return contact_name
}

func deleteObject(t *testing.T, serv *server.Server, name string, cmdtype int, retcode int, sessionid uint64) {
    delete_obj := xml.DeleteObject{Name:name}
    delete_cmd := xml.XMLCommand{CmdType:cmdtype, Sessionid:sessionid}
    delete_cmd.Content = &delete_obj
    epp_res := epp.ExecuteEPPCommand(serv, &delete_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.RetCode, epp_res.Msg)
    }
}

func createHost(t *testing.T, serv *server.Server, name string, ips []string, retcode int, sessionid uint64) {
    create_host := xml.CreateHost{Name:name, Addr:ips}
    create_cmd := xml.XMLCommand{CmdType:EPP_CREATE_HOST, Sessionid:sessionid}
    create_cmd.Content = &create_host
    epp_res := epp.ExecuteEPPCommand(serv, &create_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.RetCode, epp_res.Msg)
    }
}

func createContact(t *testing.T, serv *server.Server, create_contact *xml.CreateContact, retcode int, sessionid uint64) {
    create_cmd := xml.XMLCommand{CmdType:EPP_CREATE_CONTACT, Sessionid:sessionid}
    create_cmd.Content = create_contact
    epp_res := epp.ExecuteEPPCommand(serv, &create_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg)
    }
}
