package epp

import (
    "strings"
    "registry/server"
)

type Zone struct {
    id int
    fqdn string
}

func getDomainZone(db *server.DBConn, domain string) *Zone {
    parts := strings.Split(domain, ".")
    zone := strings.Join(parts[1:], ".")

    row := db.QueryRow("SELECT id, fqdn FROM zone " +
                       "WHERE fqdn = $1::text", zone)

    zone_obj := Zone{}
    err := row.Scan(&zone_obj.id, &zone_obj.fqdn)
    if err != nil {
        return nil
    }
    return &zone_obj
}

func testRegistrarZoneAccess(db *server.DBConn, regid uint, zoneid int) bool {
    row := db.QueryRow("SELECT id FROM registrarinvoice " +
                       "WHERE registrarid = $1::integer and zone = $2::integer and " +
                       "todate is null and fromdate <= now()" , regid, zoneid)

    var invoiceid int
    err := row.Scan(&invoiceid)
    if err != nil {
        return false
    }
    return true
}

func zoneSupported(db *server.DBConn, domain string) bool {
    parts := strings.Split(domain, ".")
    zone := strings.Join(parts[1:], ".")

    row := db.QueryRow("SELECT id FROM zone " +
                       "WHERE fqdn = $1::text ", zone)

    var zoneid int
    err := row.Scan(&zoneid)
    if err != nil {
        return false
    }
    return true
}
