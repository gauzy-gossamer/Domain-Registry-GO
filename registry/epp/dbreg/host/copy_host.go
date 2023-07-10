package host

import (
    "registry/epp/dbreg"
    "registry/server"
)

func CopyHost(db *server.DBConn, hostid uint64, host_handle string, target_regid uint) error {
    err := dbreg.LockObjectById(db, hostid, "nsset")
    if err != nil {
        return err
    }

    createObj := dbreg.NewCreateObjectDB("nsset")
    create_result, err := createObj.Exec(db, host_handle, target_regid)

    if err != nil {
        return err 
    }   

    query := "INSERT INTO host(hostid, fqdn) SELECT $1::bigint, fqdn FROM host WHERE hostid = $2::bigint"

    _, err = db.Exec(query, create_result.Id, hostid)
    if err != nil {
        return err
    }

    query = "INSERT INTO host_ipaddr_map(hostid, ipaddr) SELECT $1::bigint, ipaddr FROM host_ipaddr_map WHERE hostid=$2::bigint"
    _, err = db.Exec(query, create_result.Id, hostid)
    if err != nil {
        return err
    }

    return nil 
}
