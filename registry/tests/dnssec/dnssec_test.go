package dnssec

import (
    "strings"
    "testing"
    "context"

    "registry/xml"
    "registry/epp"
    "registry/epp/eppcom"
    "registry/tests/epptests"

    "github.com/miekg/dns"
)

func generateDSRecord(domain string, pubkey string, alg uint8, ds_alg uint8) eppcom.DSRecord {
    flags := 256
    protocol := 3

    dnskey := dns.DNSKEY{
        Hdr:       dns.RR_Header{Name:domain, Rrtype:dns.TypeDNSKEY, Class:dns.ClassINET}, 
        Flags:     uint16(flags), 
        Protocol:  uint8(protocol), 
        Algorithm: alg, 
        PublicKey: pubkey,
    }
    ds := dnskey.ToDS(ds_alg)

    dsrecord := eppcom.DSRecord{}
    dsrecord.KeyTag = int(ds.KeyTag)
    dsrecord.Alg = int(ds_alg)
    dsrecord.DigestType = int(ds.DigestType)
    dsrecord.Digest = strings.ToUpper(ds.Digest)
    dsrecord.Key.Flags = flags
    dsrecord.Key.Alg = int(alg)
    dsrecord.Key.Protocol = protocol
    dsrecord.Key.Key = pubkey

    return dsrecord
}

func updateDomain(t *testing.T, eppc *epp.EPPContext, name string, retcode int, secupdate eppcom.SecDNSUpdate, sessionid uint64) {
    ext := eppcom.EPPExt{ExtType:eppcom.EPP_EXT_SECDNS, Content:secupdate}
    update_domain := xml.UpdateDomain{Name:name}
    update_cmd := xml.XMLCommand{CmdType:eppcom.EPP_UPDATE_DOMAIN, Sessionid:sessionid}
    update_cmd.Exts = append(update_cmd.Exts, ext)
    update_cmd.Content = &update_domain
    epp_res := eppc.ExecuteEPPCommand(context.Background(), &update_cmd)
    if epp_res.RetCode != retcode {
        t.Error("should be ", retcode, epp_res.Msg, epp_res.Errors)
    }   
}

func TestRegistryServer(t *testing.T) {
    tester := epptests.NewEPPTesterConfig("../../server.conf")
    serv := tester.GetServer()

    if err := tester.SetupSession(); err != nil {
        t.Error("failed to setup ", err)
    }   
    defer tester.CloseSession()
    sessionid := tester.GetSessionid()
    eppc := epp.NewEPPContext(serv)

    domain, _ := tester.CreateDomain(t)
    pubkey := "AwEAAcNEU67LJI5GEgF9QLNqLO1SMq1EdoQ6E9f85ha0k0ewQGCblyW2836GiVsm6k8Kr5ECIoMJ6fZWf3CQSQ9ycWfTyOHfmI3eQ/1Covhb2y4bAmL/07PhrL7ozWBW3wBfM335Ft9xjtXHPy7ztCbV9qZ4TVDTW/Iyg0PiwgoXVesz"

    incorrect_dsrecord := generateDSRecord("domain.com", pubkey, dns.RSASHA256, dns.SHA256)
    dsrecord := generateDSRecord(domain, pubkey, dns.RSASHA256, dns.SHA256)
    dsrecord2 := generateDSRecord(domain, pubkey, dns.RSASHA256, dns.SHA1)

    updateDomain(t, eppc, domain, eppcom.EPP_PARAM_VALUE_POLICY, eppcom.SecDNSUpdate{AddDS:[]eppcom.DSRecord{incorrect_dsrecord}}, sessionid)
    updateDomain(t, eppc, domain, eppcom.EPP_OK, eppcom.SecDNSUpdate{AddDS:[]eppcom.DSRecord{dsrecord, dsrecord2}}, sessionid)
    updateDomain(t, eppc, domain, eppcom.EPP_OK, eppcom.SecDNSUpdate{RemAll:true}, sessionid)
    updateDomain(t, eppc, domain, eppcom.EPP_OK, eppcom.SecDNSUpdate{AddDS:[]eppcom.DSRecord{dsrecord, dsrecord2}}, sessionid)
    updateDomain(t, eppc, domain, eppcom.EPP_OK, eppcom.SecDNSUpdate{RemDS:[]eppcom.DSRecord{dsrecord2}}, sessionid)
    /* already present */
    updateDomain(t, eppc, domain, eppcom.EPP_PARAM_VALUE_POLICY, eppcom.SecDNSUpdate{AddDS:[]eppcom.DSRecord{dsrecord}}, sessionid)
    updateDomain(t, eppc, domain, eppcom.EPP_OK, eppcom.SecDNSUpdate{AddDS:[]eppcom.DSRecord{dsrecord2}}, sessionid)

    domain_data := tester.InfoDomain(t, domain)

    if domain_data.Keysetid.IsNull() {
        t.Error("expected dsrecord")
    }
    tester.DeleteDomain(t, eppc, domain)
}
