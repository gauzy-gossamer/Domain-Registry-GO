package dnssec

import (
    "registry/epp/dbreg"
    "registry/server"
    "registry/epp/eppcom"
)

type CreateKeysetDB struct {
    regid uint
    handle string

    ds []eppcom.DSRecord
}

func NewCreateKeysetDB(handle string, regid uint) CreateKeysetDB {
    obj := CreateKeysetDB{}
    obj.handle = handle
    obj.regid = regid
    return obj 
}

func (q *CreateKeysetDB) SetDSRecord(dsrec eppcom.DSRecord) *CreateKeysetDB {
    q.ds = append(q.ds, dsrec)
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

    for _, ds := range q.ds {
        err = insertDSRecord(db, object.Id, ds)
        if err != nil {
            return 0, err
        }
    }

    return object.Id, nil
}

func insertDSRecord(db *server.DBConn, object_id uint64, ds eppcom.DSRecord) error {
    row := db.QueryRow("INSERT INTO dnskey(keysetid, flags, protocol, alg, key) " +
                       "VALUES($1::bigint, $2::integer, $3::integer, $4::integer, $5::text) returning id", 
                       object_id, ds.Key.Flags, ds.Key.Protocol, ds.Key.Alg, ds.Key.Key)
    var dnskey_id int
    err := row.Scan(&dnskey_id)
    if err != nil {
        return err
    }

    _, err = db.Exec("INSERT INTO dsrecord(keysetid, dnskey_id, keytag, alg, digesttype, digest, maxsiglife) " +
                     "VALUES($1::bigint, $2::integer, $3::integer, $4::integer, $5::integer, $6::text, $7::integer)", 
                     object_id, dnskey_id, ds.KeyTag, ds.Alg, ds.DigestType, ds.Digest, ds.MaxSigLife)
    return err
}
