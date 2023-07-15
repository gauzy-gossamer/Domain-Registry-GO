package dnssec

import (
    "registry/epp/dbreg"
    "registry/server"
    "registry/epp/eppcom"
)

type CreateKeysetDB struct {
    regid uint
    handle string

    ds eppcom.DSRecord
}

func NewCreateKeysetDB(handle string, regid uint) CreateKeysetDB {
    obj := CreateKeysetDB{}
    obj.handle = handle
    obj.regid = regid
    return obj 
}

func (q *CreateKeysetDB) SetDSRecord(keytag int, alg int, digest_type int, digest string, maxsiglife int) *CreateKeysetDB {
    q.ds.KeyTag = keytag
    q.ds.Alg = alg
    q.ds.DigestType = digest_type
    q.ds.Digest = digest
    q.ds.MaxSigLife = maxsiglife
    return q
}

func (q *CreateKeysetDB) SetDNSKey(flags int, alg int, protocol int, key string) *CreateKeysetDB {
    q.ds.Key.Key = key
    q.ds.Key.Alg = alg
    q.ds.Key.Protocol = protocol
    q.ds.Key.Flags = flags
    return q
}

func (q *CreateKeysetDB) Exec(db *server.DBConn) (uint64, error) {
    createObj := dbreg.NewCreateObjectDB("keyset")
    object, err := createObj.Exec(db, q.handle, q.regid)

    if err != nil {
        return 0, err
    }

    _, err = db.Exec("INSERT INTO keyset(id) VALUES($1::bigint)", object.Id)
    if err != nil {
        return 0, err
    }

    row := db.QueryRow("INSERT INTO dnskey(keysetid, flags, protocol, alg, key) " +
                       "VALUES($1::bigint, $2::integer, $3::integer, $4::integer, $5::text) returning id", 
                       object.Id, q.ds.Key.Flags, q.ds.Key.Protocol, q.ds.Key.Alg, q.ds.Key.Key)
    var dnskey_id int
    err = row.Scan(&dnskey_id)
    if err != nil {
        return 0, err
    }

    _, err = db.Exec("INSERT INTO dsrecord(keysetid, dnskey_id, keytag, alg, digesttype, digest, maxsiglife) " +
                     "VALUES($1::bigint, $2::integer, $3::integer, $4::integer, $5::integer, $6::text, $7::integer)", 
                     object.Id, dnskey_id, q.ds.KeyTag, q.ds.Alg, q.ds.DigestType, q.ds.Digest, q.ds.MaxSigLife)
    if err != nil {
        return 0, err
    }

    return object.Id, nil
}
