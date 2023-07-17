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

type EPPTester struct {
    serv *server.Server
    sessionid uint64
    Regid uint
    RegHandle string
    Zone string
}

func NewEPPTester() *EPPTester {
    epp := &EPPTester{}
    epp.serv = PrepareServer("../../server.conf")
    return epp
}

func NewEPPTesterConfig(config string) *EPPTester {
    epp := &EPPTester{}
    epp.serv = PrepareServer(config)
    return epp
}

func (e *EPPTester) SetupSession() error {
    dbconn, err := server.AcquireConn(e.serv.Pool, server.NewLogger(""))
    if err != nil {
        return err   
    }

    defer dbconn.Close()

    regid, reg_handle, zone, err := getRegistrarAndZone(dbconn, 0)
    if err != nil {
        return err
    }
    e.Regid = regid
    e.RegHandle = reg_handle
    e.Zone = zone

    e.serv.Sessions.InitSessions(dbconn)
    sessionid, err := e.serv.Sessions.LoginSession(dbconn, regid, LANG_EN)
    e.sessionid = sessionid
    return err
}

func (e *EPPTester) CloseSession() error {
    dbconn, err := server.AcquireConn(e.serv.Pool, server.NewLogger(""))
    if err != nil {
        return err   
    }
    return e.serv.Sessions.LogoutSession(dbconn, e.sessionid)
}

func (e *EPPTester) GetServer() *server.Server {
    return e.serv
}

func (e *EPPTester) GetSessionid() uint64 {
    return e.sessionid
}

func (e *EPPTester) GetExistingContact(t *testing.T, eppc *epp.EPPContext, db *server.DBConn) string {
    return getExistingContact(t, eppc, db, e.Regid, e.sessionid)
}

func (e *EPPTester) InfoDomain(t *testing.T, domain string) *InfoDomainData {
    dbconn, err := server.AcquireConn(e.serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }   
    defer dbconn.Close()

    eppc := epp.NewEPPContext(e.serv)
    info_domain := infoDomain(t, eppc, domain, EPP_OK, e.sessionid)

    return info_domain
}

func (e *EPPTester) DeleteDomain(t *testing.T, eppc *epp.EPPContext, domain string) {
    deleteObject(t, eppc, domain, EPP_DELETE_DOMAIN, EPP_OK, e.sessionid)
}

func (e *EPPTester) CreateDomain(t *testing.T) (string, uint64) {
    dbconn, err := server.AcquireConn(e.serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }   
    defer dbconn.Close()
    test_domain := generateRandomDomain(e.Zone)

    eppc := epp.NewEPPContext(e.serv)
    contact_name := e.GetExistingContact(t, eppc, dbconn)
    createDomain(t, eppc, test_domain, contact_name, EPP_OK, e.sessionid)

    domain_data := infoDomain(t, eppc, test_domain, EPP_OK, e.sessionid)

    return test_domain, domain_data.Id
}

type Logger struct {
}

func (l Logger) StartRequest(SourceIP string, RequestType uint32, SessionID uint64, UserID uint64) uint64 {
    return 0
}

func (l Logger) EndRequest(LogID uint64, ResponseCode uint32) {

}

func PrepareServer(config string) *server.Server {
    serv := server.Server{}
    serv.Logger = Logger{}
    serv.RGconf.LoadConfig(config)
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
    rand_id := server.GenerateRandString(8)
    return rand_id + "." + zone
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

func fakeExpdate(db *server.DBConn, domainid uint64, interval string) (string, error) {
    row := db.QueryRow("SELECT now() AT TIME ZONE 'UTC' " + interval)
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

func SetExpiredExpdate(db *server.DBConn, domainid uint64) (string, error) {
    return fakeExpdate(db, domainid, "- interval '40 day'")
}

func setProlongExpdate(db *server.DBConn, domainid uint64) (string, error) {
    return fakeExpdate(db, domainid, "+ interval '30 day'")
}

func createDomain(t *testing.T, eppc *epp.EPPContext, name string, contact_name string, retcode int, sessionid uint64) {
    create_domain := xml.CreateDomain{Name:name, Registrant:contact_name}
    create_cmd := xml.XMLCommand{CmdType:EPP_CREATE_DOMAIN, Sessionid:sessionid}
    create_cmd.Content = &create_domain
    
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &create_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.RetCode, epp_res.Msg)
    }
}

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

func SetExpiredDomain(t *testing.T, serv *server.Server, domain_id uint64) {
    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        t.Error("failed acquire conn")
    }   
    defer dbconn.Close()
    _, err = SetExpiredExpdate(dbconn, domain_id)
    if err != nil {
        t.Errorf("set expired failed: %v", err)
    }   
}

func getExistingContact(t *testing.T, eppc *epp.EPPContext, db *server.DBConn, regid uint, sessionid uint64) string {
    row := db.QueryRow("SELECT name FROM object_registry obr " +
                       "JOIN object o on obr.id=o.id and o.clid = $1::integer " +
                       "WHERE erdate is null and type = get_object_type_id('contact'::text)", regid)
    var contact_name string
    err := row.Scan(&contact_name)
    if err != nil {
        contact_name = "TEST-CONTACT1"
        create_org := getCreateContact(contact_name, CONTACT_ORG)
        createContact(t, eppc, create_org, EPP_OK, sessionid)

        //t.Error(err)
    }
    return contact_name
}

func deleteObject(t *testing.T, eppc *epp.EPPContext, name string, cmdtype int, retcode int, sessionid uint64) {
    delete_obj := xml.DeleteObject{Name:name}
    delete_cmd := xml.XMLCommand{CmdType:cmdtype, Sessionid:sessionid}
    delete_cmd.Content = &delete_obj
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &delete_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.RetCode, epp_res.Msg)
    }
}

func createHost(t *testing.T, eppc *epp.EPPContext, name string, ips []string, retcode int, sessionid uint64) {
    create_host := xml.CreateHost{Name:name, Addr:ips}
    create_cmd := xml.XMLCommand{CmdType:EPP_CREATE_HOST, Sessionid:sessionid}
    create_cmd.Content = &create_host
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &create_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.RetCode, epp_res.Msg)
    }
}

func createContact(t *testing.T, eppc *epp.EPPContext, create_contact *xml.CreateContact, retcode int, sessionid uint64) {
    create_cmd := xml.XMLCommand{CmdType:EPP_CREATE_CONTACT, Sessionid:sessionid}
    create_cmd.Content = create_contact
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &create_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg)
    }
}
