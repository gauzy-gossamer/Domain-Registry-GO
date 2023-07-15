package dnssec

import (
    "registry/epp/dbreg"
    "registry/server"
)

func UpdateKeyset(db *server.DBConn, keysetid uint64, add_addrs []string, rem_addrs []string) error {
    err := dbreg.LockObjectById(db, keysetid, "keyset")
    if err != nil {
        return err
    }

    return nil
}
