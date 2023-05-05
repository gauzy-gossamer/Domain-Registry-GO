package main

import (
    "whois/server"
    "github.com/jackc/pgtype"
)

type Domain struct {
    id         uint64
    Name       string
    CrDate     pgtype.Timestamp
    ExDate     pgtype.Timestamp
    DeleteDate pgtype.Timestamp
    IntPostal  string
    Verified   bool
    CType      int
    Registrar  string
    Url        pgtype.Text
    PendingDelete bool
    Inactive      bool
    TrRegistrar pgtype.Text
    Hosts []string
}

func getHosts(db *server.DBConn, domainid uint64) ([]string, error) {
    var hosts []string

    rows, err := db.Query("SELECT fqdn FROM domain_host_map dh INNER JOIN host h ON dh.hostid=h.hostid WHERE domainid=$1;", domainid)
    defer rows.Close()
    if err != nil {
        return nil, err
    }

    for rows.Next() {
        var host string
        err := rows.Scan(&host)
        if err != nil {
            return nil, err
        }

        hosts = append(hosts, host)
    }

    return hosts, nil
}

func getDomain(domainname string) (*Domain, error) {
    dbconn, err := server.AcquireConn(serv.Pool)
    if err != nil {
        return nil, err
    }
    defer dbconn.Close()

    query := "SELECT obr.id, obr.name, crdate, exdate, " +
            "(((d.exdate + (SELECT val || ' day' FROM enum_parameters WHERE id = 6)::interval)::timestamp + (SELECT val || ' hours' FROM enum_parameters WHERE name = 'regular_day_procedure_period')::interval) AT TIME ZONE (SELECT val FROM enum_parameters WHERE name = 'regular_day_procedure_zone'))::timestamp as deletedate, " +
            "c.intpostal, c.verified, c.contact_type, r.handle, r.url, " +
            "exists(SELECT * FROM object_state WHERE object_id=d.id and valid_to is null and state_id = 17) as pendingdelete," +
            "exists(SELECT * FROM object_state WHERE object_id=d.id and valid_to is null and state_id = 15) as inactive," +
            "tr_reg.handle as tr_handle " +
            "FROM object_registry obr " +
            "  INNER JOIN domain d ON obr.id=d.id " +
            "  INNER JOIN contact c ON d.registrant=c.id " +
            "  INNER JOIN object o ON obr.id=o.id " +
            "  INNER JOIN registrar r ON o.clid=r.id " +
            "  LEFT JOIN epp_transfer_request etr ON d.id = etr.domain_id AND status = 0 " +
            "  LEFT JOIN registrar tr_reg ON tr_reg.id = etr.acquirer_id "  +
            "WHERE obr.name = LOWER($1::text) and type = 3 and erdate is null"
    var domain Domain

    row := dbconn.QueryRow(query, domainname)

    err = row.Scan(&domain.id, &domain.Name, &domain.CrDate, &domain.ExDate, &domain.DeleteDate, &domain.IntPostal,
                   &domain.Verified, &domain.CType, &domain.Registrar, &domain.Url, &domain.PendingDelete,
                   &domain.Inactive, &domain.TrRegistrar)
    if err != nil {
        return nil, err
    }
    domain.Hosts, err = getHosts(dbconn, domain.id)
    if err != nil {
        return nil, err
    }

    return &domain, nil
}
