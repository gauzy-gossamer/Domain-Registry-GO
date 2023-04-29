package dbreg

import (
    "registry/server"
)

func CreateObjectStateRequest(db *server.DBConn, objectid uint64, stateid uint) (uint, error) {
    query := "INSERT INTO object_state_request(object_id, state_id, " +
             "crdate,valid_from,valid_to) VALUES " +
             "($1::bigint,$2::bigint,now(), now(), NULL)" +
             " RETURNING id"

    row := db.QueryRow(query, objectid, stateid)

    var object_state_id uint
    err := row.Scan(&object_state_id)

    return object_state_id, err
}

func CancelObjectStateRequest(db *server.DBConn, objectid uint64, stateid uint) (uint, error) {
    query := "UPDATE object_state_request SET valid_to = now(), canceled=now() " +
             "WHERE object_id = $1::bigint and state_id = $2::bigint and valid_to is null " +
             " RETURNING id"

    row := db.QueryRow(query, objectid, stateid)

    var object_state_id uint
    err := row.Scan(&object_state_id)

    return object_state_id, err
}
