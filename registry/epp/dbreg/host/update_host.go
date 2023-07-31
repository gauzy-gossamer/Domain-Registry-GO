package host

import (
    "registry/epp/dbreg"
    "registry/server"
    . "registry/epp/eppcom"
)

type UpdateHostDB struct {
    new_name NullableVal
    new_handle NullableVal
}

func NewUpdateHostDB() UpdateHostDB {
    obj := UpdateHostDB{}
    obj.new_name.Set(nil)
    obj.new_handle.Set(nil)
    return obj 
}

func (u *UpdateHostDB) SetNewName(new_name string, new_handle string) {
    u.new_name.Set(new_name)
    u.new_handle.Set(new_handle)
}

func (u *UpdateHostDB) Exec(db *server.DBConn, hostid uint64, regid uint, add_addrs []string, rem_addrs []string) error {
    err := dbreg.LockObjectById(db, hostid, "nsset")
    if err != nil {
        return err
    }

    if !u.new_name.IsNull() {
        _, err = db.Exec("UPDATE host SET fqdn = $1::varchar WHERE hostid = $2::bigint",
                         u.new_name.Get(), hostid)
        if err != nil {
            return err
        }
        _, err = db.Exec("UPDATE object_registry SET name = $1::varchar WHERE id = $2::bigint",
                         u.new_handle.Get(), hostid)
        if err != nil {
            return err
        }
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

    err = dbreg.UpdateObject(db, hostid, regid)

    return err
}
