package dbreg

import (
    "registry/server"
    . "registry/epp/eppcom"
)

type CreateDomainDB struct {
    p_local_zone string
    domain string
    regid uint
    zoneid int
    host_objects []HostObj
    registrant uint64
    description []string
}

func NewCreateDomainDB() CreateDomainDB {
    obj := CreateDomainDB{}
    obj.p_local_zone = "UTC"
    return obj
}

func (q *CreateDomainDB) SetParams(domain string, zoneid int, registrant uint64, regid uint, description []string, host_objects []HostObj) *CreateDomainDB {
    q.domain = domain
    q.regid = regid
    q.zoneid = zoneid
    q.registrant = registrant
    q.host_objects = host_objects
    q.description = description
    return q
}

func (q *CreateDomainDB) Exec(db *server.DBConn) (*CreateDomainResult, error) {
    createObj := NewCreateObjectDB("domain")
    object, err := createObj.exec(db, q.domain, q.regid)

    if err != nil {
        return nil, err
    }

    create_result := CreateDomainResult{}
    create_result.Id = object.Id
    create_result.Name = object.Name
    row := db.QueryRow("SELECT crdate::timestamp AT TIME ZONE 'UTC' AT TIME ZONE $1::text " +
                    " , (crdate::timestamp AT TIME ZONE 'UTC' AT TIME ZONE $1::text + ( $3::integer * interval '1 month') )::timestamp " +
                    "  FROM object_registry " +
                    " WHERE id = $2::bigint FOR UPDATE OF object_registry", q.p_local_zone, object.Id, 12)
    err = row.Scan(&create_result.Crdate, &create_result.Exdate)
    if err != nil {
        return nil, err
    }

    var params []any
    cols := "INSERT INTO domain(id, zone, registrant, exdate, description)"
    vals := "VALUES($1::integer, $2::integer, $3::integer, $4::timestamp, $5::jsonb)"
    params = append(params, object.Id)
    params = append(params, q.zoneid)
    params = append(params, q.registrant)
    params = append(params, create_result.Exdate)
    params = append(params, PackJson(q.description))

    _, err = db.Exec(cols + vals, params...)
    if err != nil {
        return nil, err
    }

    for _, host := range q.host_objects {
        _, err = db.Exec("INSERT INTO domain_host_map(domainid, hostid) " +
                         "VALUES($1::integer, $2::integer)", object.Id, host.Id)
        if err != nil {
            return nil, err
        }
    }

    return &create_result, nil
}
