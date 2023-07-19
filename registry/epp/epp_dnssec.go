package epp

import (
    "strings"
    "errors"

    "registry/epp/dbreg"
    "registry/epp/dbreg/dnssec"
    "registry/epp/eppcom"

    "github.com/miekg/dns"
)

func validateDSRecord(domain string, dsrecord *eppcom.DSRecord) error {
    dnskey := dns.DNSKEY{
        Hdr:       dns.RR_Header{Name:domain, Rrtype:dns.TypeDNSKEY, Class:dns.ClassINET}, 
        Flags:     uint16(dsrecord.Key.Flags), 
        Protocol:  uint8(dsrecord.Key.Protocol), 
        Algorithm: uint8(dsrecord.Key.Alg), 
        PublicKey: dsrecord.Key.Key,
    }
    ds_alg := uint8(dsrecord.DigestType)
    ds := dnskey.ToDS(ds_alg)
    if ds == nil {
        return &dbreg.ParamError{Val:"failed to validate KeyData" + dsrecord.Key.Key}
    }

    if dsrecord.KeyTag != int(ds.KeyTag) || dsrecord.DigestType != int(ds.DigestType) || dsrecord.Digest != strings.ToUpper(ds.Digest) {
        return &dbreg.ParamError{Val:"failed to validate DSRecord " + strings.ToUpper(ds.Digest)}
    }

    return nil
}

func getKeysetHandle(domain string) string {
    return "K-" + domain
}

func createKeyset(ctx *EPPContext, domain string, dsrecords []eppcom.DSRecord) (uint64, error) {
    create_keyset := dnssec.NewCreateKeysetDB(getKeysetHandle(domain), ctx.session.Regid)

    for _, dsrec := range dsrecords {
        err := validateDSRecord(domain, &dsrec)
        if err != nil {
            return 0, err
        }
        create_keyset.SetDSRecord(dsrec)
    }

    keyset_id, err := create_keyset.Exec(ctx.dbconn)

    return keyset_id, err
}

func updateKeyset(ctx *EPPContext, keyset_id uint64, domain string, add_ds []eppcom.DSRecord, rem_ds []uint64) error {
    for _, dsrec := range add_ds {
        err := validateDSRecord(domain, &dsrec)
        if err != nil {
            return err
        }
    }

    return dnssec.UpdateKeyset(ctx.dbconn, keyset_id, add_ds, rem_ds) 
}

func createDomainSecDNS(ctx *EPPContext, domain string, content interface{}) (uint64, error) {
    dsrecs, ok := content.([]eppcom.DSRecord)
    if !ok || len(dsrecs) == 0 {
        return 0, errors.New("failed to convert DSRecord")
    }
    return createKeyset(ctx, domain, dsrecs)
}

func deleteKeyset(ctx *EPPContext, domainid uint64, keysetid uint64) error {
    /* we need to set keyset to null before removing keyset, otherwise contraints are violated */
    if _, err := ctx.dbconn.Exec("UPDATE domain SET keyset = null WHERE id = $1::bigint", domainid) ; err != nil {
        return err
    }
    return dnssec.DeleteKeyset(ctx.dbconn, keysetid)
}

func updateDomainSecDNS(ctx *EPPContext, domain_data *eppcom.InfoDomainData, domain string, content interface{}) (uint64, error) {
    secdns_update, ok := content.(eppcom.SecDNSUpdate)
    if !ok {
        return 0, errors.New("failed to convert secDNSUpdate")
    }

    dsrecs_map := make(map[eppcom.DSRecord]uint64)
    rem_ids := []uint64{}

    if !domain_data.Keysetid.IsNull() {
        dsrecs, err := dnssec.GetDSRecord(ctx.dbconn, domain_data.Keysetid.Get())
        if err != nil {
            return 0, err
        }
        for _, dsrec := range dsrecs {
            dsrec_id := dsrec.Id
            dsrec.Id = 0
            dsrec.Key.Id = 0
            dsrecs_map[dsrec] = dsrec_id
        }
    }

    if secdns_update.RemAll {
        if len(dsrecs_map) == 0 {
            return 0, &dbreg.ParamError{Val:"no dsrecords to remove"}
        }
        for _, v := range dsrecs_map {
            rem_ids = append(rem_ids, v)
        }
        dsrecs_map = make(map[eppcom.DSRecord]uint64)
    } else {
        /* remove individual records */
        for _, dsrec := range secdns_update.RemDS {
            if _, found := dsrecs_map[dsrec]; !found {
                return 0, &dbreg.ParamError{Val:"dsrecord not found"}
            }

            rem_ids = append(rem_ids, dsrecs_map[dsrec])
            delete(dsrecs_map, dsrec)
        }
    }

    for _, dsrec := range secdns_update.AddDS {
        if _, found := dsrecs_map[dsrec]; found {
            return 0, &dbreg.ParamError{Val:"dsrecord already present"}
        }
        dsrecs_map[dsrec] = 0
    }

    if len(dsrecs_map) > 0 {
        if domain_data.Keysetid.IsNull() {
            /* get current dsrecords after removing and adding new records */
            current_ds := []eppcom.DSRecord{}
            for dsrec, _ := range dsrecs_map {
                current_ds = append(current_ds, dsrec)
            }
            keysetid, err := createKeyset(ctx, domain, current_ds)
            return keysetid, err
        } 
        /* update existing keyset */
        err := updateKeyset(ctx, domain_data.Keysetid.Get(), domain, secdns_update.AddDS, rem_ids)
        return 0, err
    }

    if len(dsrecs_map) == 0 && !domain_data.Keysetid.IsNull() {
        if err := deleteKeyset(ctx, domain_data.Id, domain_data.Keysetid.Get()); err != nil {
            return 0, err
        }
        return 0, nil
    }

    return 0, nil
}
