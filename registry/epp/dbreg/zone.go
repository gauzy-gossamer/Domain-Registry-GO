package dbreg

import (
    "strings"
    "registry/server"
)

type Zone struct {
    Id int
    Fqdn string
}

func GetDomainZone(db *server.DBConn, domain string) *Zone {
    parts := strings.Split(domain, ".")
    zone := strings.Join(parts[1:], ".")

    row := db.QueryRow("SELECT id, fqdn FROM zone " +
                       "WHERE fqdn = $1::text", zone)

    zone_obj := Zone{}
    err := row.Scan(&zone_obj.Id, &zone_obj.Fqdn)
    if err != nil {
        return nil
    }
    return &zone_obj
}

func GetRegistrarZones(db *server.DBConn, regid uint) ([]string, error) {
    rows, err := db.Query("SELECT fqdn FROM registrarinvoice r JOIN zone z on r.zone=z.id " +
                       "WHERE registrarid = $1::integer and " +
                       "todate is null and fromdate <= now()" , regid)
    var zones []string
    if err != nil {
        return zones, err
    }

    for rows.Next() {
        var zone string
        err := rows.Scan(&zone)
        if err != nil {
            return zones, err
        }
        zones = append(zones, zone)
    }

    return zones, nil
}

func TestRegistrarZoneAccess(db *server.DBConn, regid uint, zoneid int) (bool, error) {
    row := db.QueryRow("SELECT id FROM registrarinvoice " +
                       "WHERE registrarid = $1::integer and zone = $2::integer and " +
                       "todate is null and fromdate <= now()" , regid, zoneid)

    var invoiceid int
    err := row.Scan(&invoiceid)
    return err == nil, err
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
