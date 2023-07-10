package contact

import (
    "registry/server"
    "registry/epp/dbreg"
)

func DeleteContact(db *server.DBConn, contactid uint64) error {
    err := dbreg.LockObjectById(db, contactid, "contact")
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

    err = dbreg.DeleteObject(db, contactid)

    return err
}
