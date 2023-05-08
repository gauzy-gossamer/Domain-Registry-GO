package dbreg

import (
    "strings"
    . "registry/epp/eppcom"
    "registry/server"
)

type InfoHostDB struct {
    p_local_zone string
    lock_ bool
    params map[string]string
}

func NewInfoHostDB() InfoHostDB {
    obj := InfoHostDB{}
    obj.p_local_zone = "UTC"
    obj.lock_ = false
    obj.params = map[string]string{}
    return obj
}

func (q *InfoHostDB) create_info_query() string {
    info_host_query := "SELECT obr.id AS id " +
        " , obr.roid AS roid , obr.name AS fqdn " +
        " , obj.clid AS registrar_id " +
        " , clr.handle AS registrar_handle, obr.crid AS cr_registrar_id " +
        " , crr.handle AS cr_registrar_handle, obj.upid AS upd_registrar_id " +
        " , upr.handle AS upd_registrar_handle " +
        " , (obr.crdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS created " +
        " , (obj.trdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS transfer_time " +
        " , (obj.update AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS update_time " /*+
        " , obj.authinfopw " +
        " , (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::timestamp AS utc_timestamp " +
        " , (CURRENT_TIMESTAMP AT TIME ZONE '" + q.p_local_zone  +"')::timestamp AS local_timestamp " */

    info_host_query += " FROM object_registry obr " +
                         " JOIN object obj ON obj.id = obr.id " +
                         " JOIN registrar clr ON clr.id = obj.clid JOIN registrar crr ON crr.id = obr.crid "

    info_host_query += " LEFT JOIN registrar upr ON upr.id = obj.upid "

    info_host_query += " WHERE obr.type = get_object_type_id('nsset'::text) "

    if _, ok := q.params["fqdn"]; ok {
        info_host_query += "and obr.name = $1"
    } else {
        info_host_query += "and obr.id = $1"
    }

    if q.lock_ {
        info_host_query += " FOR UPDATE of obr "
    } else {
        info_host_query += " FOR SHARE of obr "
    }

    return info_host_query
}

func (q *InfoHostDB) SetLock(lock bool) *InfoHostDB {
    q.lock_ = lock
    return q
}

func (q *InfoHostDB) Set_fqdn(fqdn string) *InfoHostDB {
    q.params["fqdn"] = fqdn
    return q
}

func (q *InfoHostDB) Exec(db *server.DBConn) (*InfoHostData, error) {
    info_query := q.create_info_query()

    row := db.QueryRow(info_query, q.params["fqdn"])
    var data InfoHostData

    err := row.Scan(&data.Id, &data.Roid, &data.Fqdn,
                    &data.Sponsoring_registrar.Id, &data.Sponsoring_registrar.Handle,
                    &data.Create_registrar.Id, &data.Create_registrar.Handle, &data.Update_registrar.Id, &data.Update_registrar.Handle, &data.Creation_time,
                    &data.Transfer_time, &data.Update_time)
    parts := strings.Split(data.Fqdn, ":")
    data.Fqdn = parts[0]
    if err != nil {
        return nil, err
    }

    return &data, nil
}

func GetHostIPAddrs(db *server.DBConn, hostid uint64) ([]string, error) {
    ipaddresses := []string{}
    rows, err := db.Query("SELECT host(ipaddr) FROM host_ipaddr_map WHERE hostid = $1::bigint", hostid)
    if err != nil {
        return ipaddresses, err
    }
    defer rows.Close()

    for rows.Next() {
        var ipaddr string
        err = rows.Scan(&ipaddr)
        if err != nil {
            return ipaddresses, err
        }
        ipaddresses = append(ipaddresses, ipaddr)
    }

    return ipaddresses, nil
}
