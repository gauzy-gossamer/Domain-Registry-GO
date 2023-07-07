package dbreg

import (
    "registry/server"
)

func DeleteContact(db *server.DBConn, contactid uint64) error {
    err := LockObjectById(db, contactid, "contact")
    if err != nil {
        return err
    }

    row := db.QueryRow("DELETE FROM contact WHERE id = $1::integer " +
                       "returning id", contactid)
    var deleted_id uint64
    err = row.Scan(&deleted_id)
    if err != nil {
        return err
    }

    err = deleteObject(db, contactid)

    return err
}
