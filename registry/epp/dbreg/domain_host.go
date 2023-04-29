package dbreg

import (
    "fmt"
    "strings"
    "registry/server"
    "github.com/jackc/pgx/v5"
)

type HostObj struct {
    Id uint64
    Fqdn string
    handle string
}

func hostRegistrarHandle(handle string, regid uint) string {
    return fmt.Sprintf("%s:%d", handle, regid)
}

func GetHostObject(db *server.DBConn, host_handle string, regid uint) (HostObj, error) {
    query := "SELECT obr.id, obr.name, h.fqdn FROM object_registry obr " +
             "INNER JOIN object obj ON obj.id=obr.id " +
             "INNER JOIN host h ON obj.id = h.hostid " +
             "WHERE obr.type = get_object_type_id('nsset'::text) and obr.name = $1 and obj.clid = $2 and obr.erdate is null " +
             "FOR SHARE of obr"

    var host_object HostObj
    row := db.QueryRow(query, host_handle, regid)
    err := row.Scan(&host_object.Id, &host_object.handle, &host_object.Fqdn)

    return host_object, err
}

func GetHostObjects(db *server.DBConn, hosts []string, regid uint) ([]HostObj, error) {
    host_objects := []HostObj{}
    check_duplicates := make(map[string]bool)
    for _, host := range hosts {
        host_handle := strings.ToUpper(hostRegistrarHandle(host, regid))
        host_obj, err := GetHostObject(db, host_handle, regid)
        if err != nil {
            if err == pgx.ErrNoRows {
                return nil, &ParamError{Val:host + " doesn't exist"}
            }
            return nil, err
        }
        if _, ok := check_duplicates[host_obj.handle]; ok {
            return nil, &ParamError{Val:"duplicate " + host}
        }
        check_duplicates[host_obj.handle] = true

        host_objects = append(host_objects, host_obj)
    }

    return host_objects, nil
}

