package dbreg

import (
    "registry/server"
    . "registry/epp/eppcom"
)

type CreateHostDB struct {
    p_local_zone string
    regid uint
    handle string
    fqdn string
    addrs []string
}

func NewCreateHostDB() CreateHostDB {
    obj := CreateHostDB{}
    obj.p_local_zone = "UTC"
    return obj
}

func (q *CreateHostDB) SetParams(handle string, regid uint, fqdn string, addrs []string) *CreateHostDB {
    q.fqdn = fqdn
    q.handle = handle
    q.regid = regid
    q.addrs = addrs
    return q
}

func (q *CreateHostDB) Exec(db *server.DBConn) (*CreateObjectResult, error) {
    createObj := NewCreateObjectDB("nsset")
    create_result, err := createObj.exec(db, q.handle, q.regid)

    if err != nil {
        return nil, err
    }

    row := db.QueryRow("SELECT crdate::timestamp AT TIME ZONE 'UTC' AT TIME ZONE $1::text " +
                    "  FROM object_registry " +
                    " WHERE id = $2::bigint FOR UPDATE OF object_registry", q.p_local_zone, create_result.Id)
    err = row.Scan(&create_result.Crdate)
    if err != nil {
        return nil, err
    }

    var params []any
    cols := "INSERT INTO host(hostid, fqdn) "
    vals := "VALUES($1::integer, lower($2::text))"
    params = append(params, create_result.Id)
    params = append(params, q.fqdn)

    _, err = db.Exec(cols + vals, params...)
    if err != nil {
        return nil, err
    }

    for _, ipaddr := range q.addrs {
        query := "INSERT INTO host_ipaddr_map(hostid, ipaddr) " +
                 "VALUES($1::integer, $2::inet)"
        _, err = db.Exec(query, create_result.Id, ipaddr)
        if err != nil {
            return nil, err
        }
    }

    return create_result, nil
}
