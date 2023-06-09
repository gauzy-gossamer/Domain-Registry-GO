package host

import (
    "registry/epp/dbreg"
    "registry/server"
)

func DeleteHost(db *server.DBConn, hostid uint64) error {
    err := dbreg.LockObjectById(db, hostid, "nsset")
    if err != nil {
        return err
    }

    _, err = db.Exec("DELETE FROM host_ipaddr_map WHERE hostid = $1::integer", hostid)
    if err != nil {
        return err
    }

    row := db.QueryRow("DELETE FROM host WHERE hostid = $1::integer " +
                       "returning hostid", hostid)
    var deleted_id uint64
    err = row.Scan(&deleted_id)
    if err != nil {
        return err
    }

    err = dbreg.DeleteObject(db, hostid)

    return err
}
