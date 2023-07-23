package registrar

import (
    "registry/server"
    "registry/epp/dbreg"
    . "registry/epp/eppcom"
    "github.com/jackc/pgx/v5"
)

type RegistrarInfo struct {
    RegistrarPair
    System bool
}

func GetRegistrarByHandle(db *server.DBConn, handle string) (*RegistrarInfo, error) {
    row := db.QueryRow("SELECT id, system FROM registrar WHERE handle = $1::text FOR SHARE",
                      handle)

    reg_info := RegistrarInfo{}
    reg_info.Handle.Set(handle)

    err := row.Scan(&reg_info.Id, &reg_info.System)
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

func SetNewPassword(db *server.DBConn, registrar_id uint, passwd string, new_passwd string) error {
    row := db.QueryRow("UPDATE registraracl SET password = $1::text WHERE password = $2::text and registrarid = $3::integer returning id",
                       new_passwd, passwd, registrar_id)

    var changed_reg int
    err := row.Scan(&changed_reg)
    if err != nil {
        if err == pgx.ErrNoRows {
            return dbreg.ObjectNotFound
        }
        return err
    }

    return nil
}

func AuthenticateRegistrar(db *server.DBConn, regid uint, fingerprint string, passwd string) error {
    row := db.QueryRow("SELECT registrarid FROM registraracl WHERE registrarid = $1::integer and cert = $2::text and password = $3::text",
                       regid, fingerprint, passwd)

    var reg int
    err := row.Scan(&reg)

    if err != nil {
        if err == pgx.ErrNoRows {
            return dbreg.ObjectNotFound
        }
        return err
    }

    return nil
}
