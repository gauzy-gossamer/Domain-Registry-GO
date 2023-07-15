package dnssec

import (
    "testing"

    "registry/server"
    "registry/epp/dbreg/dnssec"
    "registry/epp/dbreg"
    "registry/tests/epptests"

    "github.com/miekg/dns"
)

func TestRegistryServer(t *testing.T) {
    serv := epptests.PrepareServer("../../server.conf")
    dbconn, err := server.AcquireConn(serv.Pool, server.NewLogger(""))
    if err != nil {
        panic(err)
    }

    regid := uint(2)

    domain, domainid := epptests.CreateDomain(t, serv)

    create_keyset := dnssec.NewCreateKeysetDB("K-" + domain, regid)

    alg := dns.RSASHA256
    flags := 256
    protocol := 3
    pubkey := "AwEAAcNEU67LJI5GEgF9QLNqLO1SMq1EdoQ6E9f85ha0k0ewQGCblyW2836GiVsm6k8Kr5ECIoMJ6fZWf3CQSQ9ycWfTyOHfmI3eQ/1Covhb2y4bAmL/07PhrL7ozWBW3wBfM335Ft9xjtXHPy7ztCbV9qZ4TVDTW/Iyg0PiwgoXVesz"

    dnskey := dns.DNSKEY{
        Hdr:       dns.RR_Header{Name:domain, Rrtype:dns.TypeDNSKEY, Class:dns.ClassINET}, 
        Flags:     uint16(flags), 
        Protocol:  uint8(protocol), 
        Algorithm: alg, 
        PublicKey: pubkey,
    }
    ds_alg := dns.SHA256
    ds := dnskey.ToDS(ds_alg)

    dbconn.Begin()
    keyset_id, err := create_keyset.SetDSRecord(int(ds.KeyTag), int(ds_alg), int(ds.DigestType), ds.Digest, 4).SetDNSKey(flags, int(alg), protocol, pubkey).Exec(dbconn)
    if err != nil {
        dbconn.Rollback()
        t.Error(err)
    } else {
        dbconn.Commit()
    }

    update_domain := dbreg.NewUpdateDomainDB()
    err = update_domain.SetKeyset(keyset_id).Exec(dbconn, domainid, regid)
    if err != nil {
        t.Error(err)
    }

    domain_data := epptests.InfoDomain(t, serv, domain)

    if domain_data.Keysetid.IsNull() {
        t.Error("expected dsrecord")
    }

/*
    err = dnssec.DeleteKeyset(dbconn, keyset_id)
    if err != nil {
        t.Error(err)
    }
*/
}
