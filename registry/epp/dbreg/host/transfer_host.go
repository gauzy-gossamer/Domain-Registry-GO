package host

import (
    "registry/epp/dbreg"
    "registry/server"
)

func TransferHost(db *server.DBConn, hostid uint64, new_handle string, new_clid uint) error {
    err := dbreg.LockObjectById(db, hostid, "nsset")
    if err != nil {
        return err
    }

    row := db.QueryRow("UPDATE object SET clid = $1::integer, trdate = now() AT TIME ZONE 'UTC' " +
                       "WHERE id = $2::bigint returning id", new_clid, hostid)
    var transfered_id uint64
    err = row.Scan(&transfered_id)
    if err != nil {
        return err
    }

    row = db.QueryRow("UPDATE object_registry SET name = $1::text WHERE id = $2::bigint returning id", new_handle, hostid)
    err = row.Scan(&transfered_id)
    if err != nil {
        return err
    }

    err = dbreg.UpdateObject(db, hostid, new_clid)

    return err
}
