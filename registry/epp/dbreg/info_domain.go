package dbreg

import (
    "strings"
    "registry/server"
    . "registry/epp/eppcom"
    "github.com/jackc/pgtype"
    "github.com/jackc/pgx/v5"
)

type InfoDomainDB struct {
    p_local_zone string
    lock_ bool
    params map[string]string
}

func NewInfoDomainDB() InfoDomainDB {
    obj := InfoDomainDB{}
    obj.p_local_zone = "UTC"
    obj.lock_ = false
    obj.params = map[string]string{}
    return obj
}

func (q *InfoDomainDB) create_info_query() string {
    var query strings.Builder
    query.WriteString("SELECT dobr.id AS id ")
    query.WriteString(" , dobr.roid AS roid , dobr.name AS fqdn ")
    query.WriteString(" , (dobr.erdate AT TIME ZONE 'UTC' ) AT TIME ZONE '" + q.p_local_zone + "' AS delete_time ")
    query.WriteString(" , cor.id AS registrant_id , cor.name  AS registrant_handle ")
    query.WriteString(" , dt.description, obj.clid AS registrar_id ")
    query.WriteString(" , clr.handle AS registrar_handle, dobr.crid AS cr_registrar_id ")
    query.WriteString(" , crr.handle AS cr_registrar_handle, obj.upid AS upd_registrar_id ")
    query.WriteString(" , upr.handle AS upd_registrar_handle ")
    query.WriteString(" , (dobr.crdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS created ")
    query.WriteString(" , (obj.trdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS transfer_time ")
    query.WriteString(" , (obj.update AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS update_time ")
    query.WriteString(" , dt.exdate , obj.authinfopw ")
    query.WriteString(" , (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::timestamp AS utc_timestamp, z.id as zone_id ")/*
        " , obj.authinfoupdate >= current_timestamp + (SELECT val || ' days' FROM enum_parameters WHERE id = 21)::interval AS authinfo_valid " +
        " , z.id AS zone_id, z.fqdn AS zone_fqdn" */

    query.WriteString(" FROM object_registry dobr ")
    query.WriteString(" JOIN object obj ON obj.id = dobr.id JOIN domain dt ON dt.id = obj.id ")
    query.WriteString(" JOIN object_registry cor ON dt.registrant=cor.id ")
    query.WriteString(" JOIN registrar clr ON clr.id = obj.clid JOIN registrar crr ON crr.id = dobr.crid ")
    query.WriteString(" JOIN zone z ON dt.zone = z.id ")

    query.WriteString(" LEFT JOIN registrar upr ON upr.id = obj.upid ")

    query.WriteString(" WHERE dobr.type = get_object_type_id('domain'::text) ")

    if _, ok := q.params["fqdn"]; ok {
        query.WriteString("and dobr.name = $1")
    } else {
        query.WriteString("and dobr.id = $1")
    }

    if q.lock_ {
        query.WriteString(" FOR UPDATE of dobr ")
    } else {
        query.WriteString(" FOR SHARE of dobr ")
    }

    return query.String()
}

func (q *InfoDomainDB) Set_lock(lock bool) *InfoDomainDB {
    q.lock_ = lock
    return q
}

func (q *InfoDomainDB) Set_fqdn(fqdn string) *InfoDomainDB {
    q.params["fqdn"] = fqdn
    return q
}

func (q *InfoDomainDB) Exec(db *server.DBConn) (*InfoDomainData, error) {
    info_query := q.create_info_query()

    row := db.QueryRow(info_query, q.params["fqdn"])
    var data InfoDomainData

    var description pgtype.Text

    err := row.Scan(&data.Id, &data.Roid, &data.Fqdn, &data.Expiration_date, &data.Registrant.Id, &data.Registrant.Handle,
                    &description, &data.Sponsoring_registrar.Id, &data.Sponsoring_registrar.Handle,
                    &data.Create_registrar.Id, &data.Create_registrar.Handle, &data.Update_registrar.Id, &data.Update_registrar.Handle, &data.Creation_time,
                    &data.Transfer_time, &data.Update_time, &data.Expiration_date, &data.Authinfopw, &data.Cur_time, &data.ZoneId)
    if err != nil {
        return nil, err
    }
    data.Description = UnpackJson(description)

    return &data, nil
}

func GetDomainHosts(db *server.DBConn, domainid uint64) ([]HostObj, error) {
    var hosts []HostObj

    rows, err := db.Query("SELECT h.hostid, fqdn FROM domain_host_map dh INNER JOIN host h ON dh.hostid=h.hostid WHERE domainid=$1;", domainid)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var host HostObj
        err := rows.Scan(&host.Id, &host.Fqdn)
        if err != nil {
            return nil, err
        }

        hosts = append(hosts, host)
    }

    return hosts, nil
}

func GetDomainIdByName(db *server.DBConn, handle string, regid... uint) (uint64, error) {
    domain_id, err := GetObjectIdByName(db, handle, "domain", regid...)

    if err != nil {
        if err == pgx.ErrNoRows {
            return 0, &ParamError{Val:"domain " + handle + " doesn't exist"}
        }
    }

    return domain_id, err
}
