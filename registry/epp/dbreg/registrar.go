package dbreg

import (
    "registry/server"
    . "registry/epp/eppcom"
    "github.com/jackc/pgx/v5"
)

func GetRegistrarByHandle(db *server.DBConn, handle string) (*RegistrarPair, error) {
    row := db.QueryRow("SELECT id, system " +
                       " FROM registrar WHERE handle = $1::text FOR SHARE", handle)

    reg_info := RegistrarPair{}
    reg_info.Handle.Set(handle)

    var system bool
    err := row.Scan(&reg_info.Id, &system)
    if err != nil {
        if err == pgx.ErrNoRows {
            return &reg_info, &ParamError{Val:"registrar " + handle + " doesn't exist"}
        }
    }

    return &reg_info, err
}

