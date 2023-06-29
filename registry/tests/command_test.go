package tests

import (
    "testing"
    "net/http/httptest"
    "strings"

    "registry/server"
    "registry/epp"
    "registry/tests/epptests"
)

func TestGetUserIPAddr(t *testing.T) {
    test_ipaddr := "10.10.10.10"
    serv := epptests.PrepareServer("../server.conf")
    req := httptest.NewRequest("GET", "http://localhost", nil)

    req.RemoteAddr = test_ipaddr

    ipaddr := server.GetUserIPAddr(serv, req)
    if ipaddr != test_ipaddr {
        t.Error(ipaddr)
    }

    req.Header.Set("X-Forwarded-For", "10.10.10.10")
}

/* test full HandleRequest */
func TestHttpRequest(t *testing.T) {
    serv := epptests.PrepareServer("../server.conf")
    serv.XmlParser.ReadSchema(serv.RGconf.SchemaPath)
    req := httptest.NewRequest("GET", "https://localhost", nil)
    w := httptest.NewRecorder()
    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        panic(err)
    }

    serv.Sessions.InitSessions(dbconn)

    req.RemoteAddr = "10.10.10.10"

    // test xml parse error
    eppc := epp.NewEPPContext(serv)
    response := server.HandleRequest(serv, eppc, w, req, "<incorrect></incorrect>")

    if !strings.Contains(response, "<msg>Command syntax error") {
        t.Error("expected error", response)
    }

    // test incorrect authentication
    xml_msg := `<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
  <command>
    <login>
      <clID>NEWREG-3LVL</clID>
      <pw>password</pw>
      <options>
        <version>1.0</version>
        <lang>en</lang>
      </options>
      <svcs>
        <objURI>urn:ietf:params:xml:ns:domain-1.0</objURI>
        <objURI>urn:ietf:params:xml:ns:host-1.0</objURI>
        <objURI>urn:ietf:params:xml:ns:contact-1.0</objURI>
      </svcs>
    </login>
    <clTRID>XzpZDcPrN7je</clTRID>
  </command>
</epp>`

    response = server.HandleRequest(serv, eppc, w, req, xml_msg)

    if !strings.Contains(response, "<msg>Authentication error") {
        t.Error("expected auth error", response)
    }

    req.Header.Set("X-Forwarded-For", "10.10.10.10")
    req.Header.Set("X-SSL-CERT", test_registrar_cert) // from cert_test.go

    serv.RGconf.HTTPConf.UseProxy = true
    xml_msg = `<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
  <command>
    <login>
      <clID>TEST-REG</clID>
      <pw>password</pw>
      <options>
        <version>1.0</version>
        <lang>en</lang>
      </options>
      <svcs>
        <objURI>urn:ietf:params:xml:ns:domain-1.0</objURI>
        <objURI>urn:ietf:params:xml:ns:host-1.0</objURI>
        <objURI>urn:ietf:params:xml:ns:contact-1.0</objURI>
      </svcs>
    </login>
    <clTRID>XzpZDcPrN7je</clTRID>
  </command>
</epp>`

    response = server.HandleRequest(serv, eppc, w, req, xml_msg)

    if !strings.Contains(response, "<msg>Command completed successfully</msg>") {
        t.Error("expected success", response)
    }

    // EPPSESSIONID should be set 
    req.AddCookie(w.Result().Cookies()[0])
    xml_msg = `<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
  <command>
    <logout />
  </command>
</epp>`
    response = server.HandleRequest(serv, eppc, w, req, xml_msg)

    if !strings.Contains(response, "<result code=\"1500\">") {
        t.Error("expected success", response)
    }
}
