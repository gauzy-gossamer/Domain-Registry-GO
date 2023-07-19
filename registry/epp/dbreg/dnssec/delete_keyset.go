package dnssec

import (
    "registry/epp/dbreg"
    "registry/server"
)

func DeleteKeyset(db *server.DBConn, dnskey_id uint64) error {
    err := dbreg.LockObjectById(db, dnskey_id, "keyset")
    if err != nil {
        return err 
    }   

    _, err = db.Exec("DELETE FROM dsrecord WHERE keysetid = $1::bigint ", dnskey_id)
    if err != nil {
        return err 
    }   

    _, err = db.Exec("DELETE FROM dnskey WHERE keysetid = $1::bigint ", dnskey_id)
    if err != nil {
        return err 
    }   

    row := db.QueryRow("DELETE FROM keyset WHERE id = $1::bigint returning id",
                       dnskey_id)
    var deleted_id uint64
    err = row.Scan(&deleted_id)
    if err != nil {
        return err 
    }   

    err = dbreg.DeleteObject(db, dnskey_id)

    return err 
}

