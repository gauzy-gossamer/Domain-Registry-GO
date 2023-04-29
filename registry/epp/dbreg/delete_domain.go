package dbreg

import (
    "registry/server"
)

func DeleteDomain(db *server.DBConn, domainid uint64) error {
    err := lockObjectById(db, domainid, "domain")
    if err != nil {
        return err
    }

    _, err = db.Exec("DELETE FROM domain_host_map WHERE domainid = $1::integer ", domainid)
    if err != nil {
        return err
    }

    row := db.QueryRow("DELETE FROM domain WHERE id = $1::integer " +
                       "returning id", domainid)
    var deleted_id uint64
    err = row.Scan(&deleted_id)
    if err != nil {
        return err
    }

    err = deleteObject(db, domainid)

    return err
}
