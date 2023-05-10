package dbreg

import (
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
    info_domain_query := "SELECT dobr.id AS id " +
        " , dobr.roid AS roid , dobr.name AS fqdn " +
        " , (dobr.erdate AT TIME ZONE 'UTC' ) AT TIME ZONE '" + q.p_local_zone + "' AS delete_time " +
        " , cor.id AS registrant_id , cor.name  AS registrant_handle " +
        " , dt.description, obj.clid AS registrar_id " +
        " , clr.handle AS registrar_handle, dobr.crid AS cr_registrar_id " +
        " , crr.handle AS cr_registrar_handle, obj.upid AS upd_registrar_id " +
        " , upr.handle AS upd_registrar_handle " +
        " , (dobr.crdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS created " +
        " , (obj.trdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS transfer_time " +
        " , (obj.update AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS update_time " +
        " , dt.exdate , obj.authinfopw " +
        " , (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::timestamp AS utc_timestamp, z.id as zone_id "/*
        " , obj.authinfoupdate >= current_timestamp + (SELECT val || ' days' FROM enum_parameters WHERE id = 21)::interval AS authinfo_valid " +
        " , z.id AS zone_id, z.fqdn AS zone_fqdn" */

    info_domain_query += " FROM object_registry dobr " +
                         " JOIN object obj ON obj.id = dobr.id JOIN domain dt ON dt.id = obj.id " +
                         " JOIN object_registry cor ON dt.registrant=cor.id " +
                         " JOIN registrar clr ON clr.id = obj.clid JOIN registrar crr ON crr.id = dobr.crid " +
                         " JOIN zone z ON dt.zone = z.id "

    info_domain_query += " LEFT JOIN registrar upr ON upr.id = obj.upid "

    info_domain_query += " WHERE dobr.type = get_object_type_id('domain'::text) "

    if _, ok := q.params["fqdn"]; ok {
        info_domain_query += "and dobr.name = $1"
    } else {
        info_domain_query += "and dobr.id = $1"
    }

    if q.lock_ {
        info_domain_query += " FOR UPDATE of dobr "
    } else {
        info_domain_query += " FOR SHARE of dobr "
    }

    return info_domain_query
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
    data.Description = unpackJson(description)

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
    domain_id, err := getObjectIdByName(db, handle, "domain", regid...)

    if err != nil {
        if err == pgx.ErrNoRows {
            return 0, &ParamError{Val:"domain " + handle + " doesn't exist"}
        }
    }

    return domain_id, err
}
