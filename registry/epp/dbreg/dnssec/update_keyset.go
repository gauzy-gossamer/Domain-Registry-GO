package dnssec

import (
    "registry/epp/eppcom"
    "registry/epp/dbreg"
    "registry/server"
)

func UpdateKeyset(db *server.DBConn, keysetid uint64, add_ds []eppcom.DSRecord, rem_ds []uint64) error {
    err := dbreg.LockObjectById(db, keysetid, "keyset")
    if err != nil {
        return err
    }

    for _, ds := range add_ds {
        err = insertDSRecord(db, keysetid, ds)
        if err != nil {
            return err
        }
    }

    for _, dsid := range rem_ds {
        _, err := db.Exec("DELETE FROM dnskey WHERE id in (SELECT id FROM dsrecord WHERE id = $1::integer)", dsid)
        if err != nil {
            return err
        }

        _, err = db.Exec("DELETE FROM dsrecord WHERE id = $1::integer", dsid)
        if err != nil {
            return err
        }
    }

    return nil
}
