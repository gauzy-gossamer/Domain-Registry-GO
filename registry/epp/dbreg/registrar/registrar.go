package registrar

import (
    "registry/server"
    "registry/epp/dbreg"
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
            return &reg_info, &dbreg.ParamError{Val:"registrar " + handle + " doesn't exist"}
        }
    }

    return &reg_info, err
}

func GetRegistrarIPAddrs(db *server.DBConn, registrar_id uint64) ([]string, error) {
    addrs := []string{}

    rows, err := db.Query("SELECT host(ipaddr) " +
                       " FROM registrar_ipaddr_map WHERE registrarid = $1::integer FOR SHARE", registrar_id)
    if err != nil {
        return addrs, err
    }
    defer rows.Close()

    for rows.Next() {
        var ipaddr string
        err = rows.Scan(&ipaddr)
        if err != nil {
            return addrs, nil
        }
        addrs = append(addrs, ipaddr)
    }

    return addrs, nil
}
