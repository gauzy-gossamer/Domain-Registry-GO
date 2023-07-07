package dbreg

import (
    "registry/server"
)

func UpdateHost(db *server.DBConn, hostid uint64, regid uint, add_addrs []string, rem_addrs []string) error {
    err := LockObjectById(db, hostid, "nsset")
    if err != nil {
        return err
    }

    for _, ipaddr := range add_addrs {
        query := "INSERT INTO host_ipaddr_map(hostid, ipaddr) " +
                 "VALUES($1::integer, $2::inet)"
        _, err = db.Exec(query, hostid, ipaddr)
        if err != nil {
            return err
        }
    }

    for _, ipaddr := range rem_addrs {
        query := "DELETE FROM host_ipaddr_map " +
                 "WHERE hostid = $1::integer and ipaddr = $2::inet"
        _, err = db.Exec(query, hostid, ipaddr)
        if err != nil {
            return err
        }
    }

    err = UpdateObject(db, hostid, regid)

    return err
}
