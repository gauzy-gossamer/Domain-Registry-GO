package contact

import (
    "registry/epp/dbreg"
    "registry/server"
)

func TransferContact(db *server.DBConn, contactid uint64, new_clid uint) error {
    err := dbreg.LockObjectById(db, contactid, "contact")
    if err != nil {
        return err
    }

    row := db.QueryRow("UPDATE object SET clid = $1::integer, trdate = now() AT TIME ZONE 'UTC' " +
                       "WHERE id = $2::integer returning id", new_clid, contactid)
    var transfered_id uint64
    err = row.Scan(&transfered_id)
    if err != nil {
        return err
    }

    err = dbreg.UpdateObject(db, contactid, new_clid)

    return err
}
