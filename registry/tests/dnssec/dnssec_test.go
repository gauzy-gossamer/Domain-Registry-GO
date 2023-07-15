package dnssec

import (
    "testing"

    "registry/server"
    "registry/epp/dbreg/dnssec"
    "registry/epp/dbreg"
    "registry/tests/epptests"
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

    keytag := 10
    alg := 5
    digestType := 1
    digest := "7CE5D830A8194AA98DBF3A32223F2A4C79DF5E578CB39D0217878810F28C880DB11398DF565FD1C7555F786CC1A22B53"
    flags := 10
    protocol := 3
    pubkey := "3ab1GX"

    dbconn.Begin()
    keyset_id, err := create_keyset.SetDSRecord(keytag, alg, digestType, digest, 4).SetDNSKey(flags, alg, protocol, pubkey).Exec(dbconn)
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
