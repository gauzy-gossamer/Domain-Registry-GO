/*
go test -coverpkg=./... -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
*/
package tests

import (
    "testing"
    "registry/server"
    . "registry/epp/eppcom"
    "registry/xml"
)

func prepareParser() *xml.XMLParser {
    conf := server.RegConfig{}
    conf.LoadConfig("../server.conf")
    var parser xml.XMLParser
    parser.ReadSchema(conf.SchemaPath)
    return &parser
}

func TestSimple(t *testing.T) {
    parser := prepareParser()
    _, err := parser.ParseMessage("incorrect xml")
    if err == nil {
        t.Error("must be an error")
    }
    xml_msg := `<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
    <hello />
  </epp>`
    cmd, err := parser.ParseMessage(xml_msg)
    if err != nil || cmd.CmdType != EPP_HELLO {
        t.Error("must be hello")
    }

    xml_msg = `<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
  <command>
    <login>
      <clID>NEWREG-3LVL</clID>
      <pw>pawd</pw>
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
    cmd, err = parser.ParseMessage(xml_msg)
    if err == nil {
        t.Error("must be error")
    }
    if _, ok := err.(*xml.CommandError); !ok {
        t.Error("must be error")
    }
}

func TestDomain(t *testing.T) {
    parser := prepareParser()
    xml_msg := `<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
    <command>
      <info>
        <domain:info xmlns:domain="http://www.ripn.net/epp/ripn-domain-1.0">
          <domain:name>domain.net.ru</domain:name>
        </domain:info>
      </info>
    </command>
  </epp>`
    cmd, err := parser.ParseMessage(xml_msg)
    if err != nil || cmd.CmdType != EPP_INFO_DOMAIN {
        t.Error("failed domain info")
    }
    xml_msg = `<epp xmlns="http://www.ripn.net/epp/ripn-epp-1.0">
    <command>
      <update>
        <domain:update xmlns:domain="http://www.ripn.net/epp/ripn-domain-1.0" xmlns="http://www.ripn.net/epp/ripn-domain-1.0">
          <name>domain.ru</name>
          <add>
            <ns>
              <hostObj>ns1.domain</hostObj>
            </ns>
          </add>
          <chg>
            <registrant>TEST-CONTACT6</registrant>
          </chg>
        </domain:update>
      </update>
    </command>
  </epp>`
    cmd, err = parser.ParseMessage(xml_msg)
    if err != nil || cmd.CmdType != EPP_UPDATE_DOMAIN {
        t.Error("failed domain update")
    }
}