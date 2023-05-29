package dbreg

import (
    . "registry/epp/eppcom"
    "registry/server"
)

type CreateObjectDB struct {
    p_local_zone string
    object_type string
    authinfo string
}

func NewCreateObjectDB(object_type string) CreateObjectDB {
    obj := CreateObjectDB{}
    obj.p_local_zone = "UTC"
    obj.object_type = object_type
    return obj
}

func (c *CreateObjectDB) setAuthInfo(authinfo string) *CreateObjectDB {
    c.authinfo = authinfo
    return c
}

func (q *CreateObjectDB) exec(db *server.DBConn, handle string, regid uint) (*CreateObjectResult, error) {
    object_type_id, err := getObjectTypeId(db, q.object_type)
    if err != nil {
        return nil, err
    }

    row := db.QueryRow("SELECT create_object($1::integer " +//registrar
                    " , $2::text " + //object handle
                    " , $3::integer )", regid, handle, object_type_id)

    var result CreateObjectResult
    err = row.Scan(&result.Id)

    if err != nil {
        return nil, err
    }
    result.Name = handle

    var params []any
    params = append(params, result.Id)
    params = append(params, regid)

    cols := "INSERT INTO object(id, clid"
    vals := " VALUES($1::bigint, $2::integer"

    if q.authinfo != "" {
        cols += ", authinfo, authinfoupdate"
        vals += ", $3::text, now()"
        params = append(params, q.authinfo)
    }

    cols += ")"
    vals += ")"

    _, err = db.Exec(cols + vals, params...)

    if err != nil {
        return nil, err
    }

    return &result, nil
}

func getObjectTypeId(db *server.DBConn, object_type string) (int, error) {
    var object_type_id int
    err := db.QueryRow("SELECT get_object_type_id($1::text)", object_type).Scan(&object_type_id)
    return object_type_id, err
}

func deleteObject(db *server.DBConn, object_id uint64) error {
    row := db.QueryRow("UPDATE object_registry SET erdate = now() " +
            "WHERE id = $1::integer RETURNING id", object_id)

    var deleted_id uint64
    err := row.Scan(&deleted_id)
    if err != nil {
        return err
    }

    row = db.QueryRow("DELETE FROM object WHERE id = $1::integer RETURNING id", object_id)
    err = row.Scan(&deleted_id)

    return err
}

func updateObject(db *server.DBConn, object_id uint64, regid uint) error {
    row := db.QueryRow("UPDATE object SET update = now(), upid = $1::integer " +
            "WHERE id = $2::integer RETURNING id", regid, object_id)

    var updated_id uint64
    err := row.Scan(&updated_id)
    return err
}

func lockObjectById(db *server.DBConn, object_id uint64, object_type string) error {
    row := db.QueryRow("SELECT id FROM object_registry WHERE id = $1::integer and erdate is null " +
                      "and type = get_object_type_id($2::text) FOR UPDATE", object_id, object_type)
    var locked_id uint64
    return row.Scan(&locked_id)
}

func getObjectIdByName(db *server.DBConn, handle string, object_type string, regid... uint) (uint64, error) {
    query := "SELECT obr.id FROM object_registry obr " +
             "INNER JOIN object obj ON obj.id=obr.id " +
             "WHERE obr.type = get_object_type_id($1::text) and obr.name = lower($2::text) and obr.erdate is null "

    var params []any
    params = append(params, object_type)
    params = append(params, handle)
    if regid != nil {
        params = append(params, regid[0])
        query += " and obj.clid = $3 "
    }

    query += "FOR SHARE of obr"

    row := db.QueryRow(query, params...)

    var object_id uint64
    err := row.Scan(&object_id)
    if err != nil {
        return 0, err
    }

    return object_id, err
}
