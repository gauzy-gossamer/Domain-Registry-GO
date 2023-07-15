package dnssec

import (
    "strings"

    "registry/server"
    "registry/epp/eppcom"
)

func GetDSRecord(db *server.DBConn, keyset_id uint64) (*eppcom.DSRecord, error) {
    var query strings.Builder

    query.WriteString("SELECT ds.id as dsid, ds.keytag, ds.alg AS ds_alg, ds.digesttype, ds.digest, ds.maxsiglife, ");
    query.WriteString("d.id, d.flags, d.protocol, d.alg AS alg, d.key");
    query.WriteString(" FROM dsrecord ds INNER JOIN dnskey d ON ds.dnskey_id=d.id ");
    query.WriteString("WHERE ds.keysetid = $1::bigint")

    row := db.QueryRow(query.String(), keyset_id)
    var data eppcom.DSRecord

    err := row.Scan(&data.Id, &data.KeyTag, &data.Alg, &data.DigestType, &data.Digest, &data.MaxSigLife,
                    &data.Key.Id, &data.Key.Flags, &data.Key.Protocol, &data.Key.Alg, &data.Key.Key)
    if err != nil {
        return nil, err 
    }   

    return &data, nil 
}
