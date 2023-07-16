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
        return &dbreg.ParamError{Val:"failed to validate KeyData"}
    }

    if dsrecord.KeyTag != int(ds.KeyTag) || dsrecord.DigestType != int(ds.DigestType) || dsrecord.Digest != strings.ToUpper(ds.Digest) {
        return &dbreg.ParamError{Val:"failed to validate DSRecord"}
    }

    return nil
}

func getKeysetHandle(domain string) string {
    return "K-" + domain
}

func createKeyset(ctx *EPPContext, domain string, dsrecord *eppcom.DSRecord) (uint64, error) {
    create_keyset := dnssec.NewCreateKeysetDB(getKeysetHandle(domain), ctx.session.Regid)

    err := validateDSRecord(domain, dsrecord)
    if err != nil {
        return 0, err
    }

    keyset_id, err := create_keyset.
                      SetDSRecord(dsrecord.KeyTag, dsrecord.Alg, dsrecord.DigestType, dsrecord.Digest, dsrecord.MaxSigLife).
                      SetDNSKey(dsrecord.Key.Flags, dsrecord.Key.Alg, dsrecord.Key.Protocol, dsrecord.Key.Key).
                      Exec(ctx.dbconn)

    return keyset_id, err
}

func updateKeyset(ctx *EPPContext, keyset_id uint64, domain string, dsrecord *eppcom.DSRecord) error {
    create_keyset := dnssec.NewCreateKeysetDB(getKeysetHandle(domain), ctx.session.Regid)

    err := validateDSRecord(domain, dsrecord)
    if err != nil {
        return err
    }

    _, err = create_keyset.
                      SetDSRecord(dsrecord.KeyTag, dsrecord.Alg, dsrecord.DigestType, dsrecord.Digest, dsrecord.MaxSigLife).
                      SetDNSKey(dsrecord.Key.Flags, dsrecord.Key.Alg, dsrecord.Key.Protocol, dsrecord.Key.Key).
                      Exec(ctx.dbconn)

    return err
}

func createDomainSecDNS(ctx *EPPContext, domain string, content interface{}) (uint64, error) {
    dsrecs, ok := content.([]eppcom.DSRecord)
    if !ok || len(dsrecs) == 0 {
        return 0, errors.New("failed to convert DSRecord")

    }   
    return createKeyset(ctx, domain, &dsrecs[0])
}
