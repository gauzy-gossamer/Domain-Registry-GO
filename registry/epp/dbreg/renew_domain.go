package dbreg

import (
    "registry/server"
    "github.com/jackc/pgtype"
)

func RenewDomain(db *server.DBConn, domainid uint64, exdate pgtype.Timestamp, new_exdate pgtype.Timestamp) error {
    err := lockObjectById(db, domainid, "domain")
    if err != nil {
        return err
    }

    row := db.QueryRow("UPDATE domain SET exdate = $1::timestamp WHERE id = $2::integer and exdate = $3::timestamp RETURNING id",
                       new_exdate, domainid, exdate)
    var updated_id uint64
    err = row.Scan(&updated_id)

    return err
}

func GetNewExdate(db *server.DBConn, exdate pgtype.Timestamp, period int) (pgtype.Timestamp, error) {
    row := db.QueryRow("SELECT $1::timestamp + $2::integer*('1 month'::interval)", exdate, period)

    var newexdate pgtype.Timestamp

    err := row.Scan(&newexdate)

    return newexdate, err
}
