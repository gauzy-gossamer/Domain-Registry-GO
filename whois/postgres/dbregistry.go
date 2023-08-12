package postgres

import (
    "strings"
    "time"

    "whois/whois_resp"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type WhoisStorage struct {
    pool *pgxpool.Pool
}

func NewWhoisStorage(dbconf DBConfig) (*WhoisStorage, error) {
    pool, err := CreatePool(&dbconf)
    return &WhoisStorage{pool:pool}, err
}

func getHosts(db *DBConn, domainid uint64) ([]string, error) {
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

func (w *WhoisStorage) GetDomain(domainname string) (whois_resp.Domain, error) {
    var domain whois_resp.Domain
    dbconn, err := AcquireConn(w.pool)
    if err != nil {
        return domain, err
    }
    defer dbconn.Close()

    var query strings.Builder

    query.WriteString("SELECT obr.id, obr.name, crdate, exdate, ")
    query.WriteString("(((d.exdate + (SELECT val || ' day' FROM enum_parameters WHERE id = 6)::interval)::timestamp ")
    query.WriteString("+ (SELECT val || ' hours' FROM enum_parameters WHERE name = 'regular_day_procedure_period')::interval) AT TIME ZONE (SELECT val FROM enum_parameters WHERE name = 'regular_day_procedure_zone'))::timestamp as deletedate, ")
    query.WriteString("c.intpostal, c.verified, c.contact_type, r.handle, r.url, ")
    query.WriteString("exists(SELECT * FROM object_state WHERE object_id=d.id and valid_to is null and state_id = 17) as pendingdelete,")
    query.WriteString("exists(SELECT * FROM object_state WHERE object_id=d.id and valid_to is null and state_id = 15) as inactive,")
    query.WriteString("tr_reg.handle as tr_handle ")
    query.WriteString("FROM object_registry obr ")
    query.WriteString(" INNER JOIN domain d ON obr.id=d.id ")
    query.WriteString(" INNER JOIN contact c ON d.registrant=c.id ")
    query.WriteString(" INNER JOIN object o ON obr.id=o.id ")
    query.WriteString(" INNER JOIN registrar r ON o.clid=r.id ")
    query.WriteString(" LEFT JOIN epp_transfer_request etr ON d.id = etr.domain_id AND status = 0 ")
    query.WriteString(" LEFT JOIN registrar tr_reg ON tr_reg.id = etr.acquirer_id ")
    query.WriteString("WHERE obr.name = LOWER($1::text) and type = 3 and erdate is null")

    row := dbconn.QueryRow(query.String(), domainname)

    err = row.Scan(&domain.Id, &domain.Name, &domain.CrDate, &domain.ExDate, &domain.DeleteDate, &domain.IntPostal,
                   &domain.Verified, &domain.CType, &domain.Registrar, &domain.Url, &domain.PendingDelete,
                   &domain.Inactive, &domain.TrRegistrar)
    if err != nil {
        if err == pgx.ErrNoRows {
            return domain, whois_resp.ObjectNotFound
        }
        return domain, err
    }
    domain.Hosts, err = getHosts(dbconn, domain.Id)
    if err != nil {
        return domain, err
    }

    domain.Retrieved = time.Now()

    return domain, nil
}

func (w *WhoisStorage) GetRegistrar(reg_handle string) (whois_resp.Registrar, error) {
    var reg whois_resp.Registrar
    dbconn, err := AcquireConn(w.pool)
    if err != nil {
        return reg, err
    }
    defer dbconn.Close()

    var query strings.Builder

    query.WriteString("SELECT handle, intpostal, telephone->>0, fax->>0, email->>0, www, whois FROM registrar ")
    query.WriteString("WHERE handle = UPPER($1::text)")

    row := dbconn.QueryRow(query.String(), reg_handle)

    err = row.Scan(&reg.Handle, &reg.Org, &reg.Phone, &reg.Fax, &reg.Email, &reg.WWW, &reg.Whois)
    if err != nil {
        if err == pgx.ErrNoRows {
            return reg, whois_resp.ObjectNotFound
        }
        return reg, err
    }

    reg.Retrieved = time.Now()

    return reg, nil
}
