package dbreg

import (
    "registry/server"
)

func TransferDomain(db *server.DBConn, domainid uint64, new_clid uint) error {
    err := LockObjectById(db, domainid, "domain")
    if err != nil {
        return err
    }

    row := db.QueryRow("UPDATE object SET clid = $1::integer, trdate = now() AT TIME ZONE 'UTC' " +
                       "WHERE id = $2::integer returning id", new_clid, domainid)
    var transfered_id uint64
    err = row.Scan(&transfered_id)
    if err != nil {
        return err
    }

    err = UpdateObject(db, domainid, new_clid)

    return err
}
