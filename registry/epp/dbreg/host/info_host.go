package host

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
    var query strings.Builder
    query.WriteString("SELECT obr.id AS id ")
    query.WriteString(" , obr.roid AS roid , obr.name AS fqdn ")
    query.WriteString(" , obj.clid AS registrar_id ")
    query.WriteString(" , clr.handle AS registrar_handle, obr.crid AS cr_registrar_id ")
    query.WriteString(" , crr.handle AS cr_registrar_handle, obj.upid AS upd_registrar_id ")
    query.WriteString(" , upr.handle AS upd_registrar_handle ")
    query.WriteString(" , (obr.crdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS created ")
    query.WriteString(" , (obj.trdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS transfer_time ")
    query.WriteString(" , (obj.update AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS update_time ") /*+
        " , obj.authinfopw " +
        " , (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::timestamp AS utc_timestamp " +
        " , (CURRENT_TIMESTAMP AT TIME ZONE '" + q.p_local_zone  +"')::timestamp AS local_timestamp " */

    query.WriteString(" FROM object_registry obr ")
    query.WriteString(" JOIN object obj ON obj.id = obr.id ")
    query.WriteString(" JOIN registrar clr ON clr.id = obj.clid JOIN registrar crr ON crr.id = obr.crid ")

    query.WriteString(" LEFT JOIN registrar upr ON upr.id = obj.upid ")

    query.WriteString(" WHERE obr.type = get_object_type_id('nsset'::text) ")

    if _, ok := q.params["fqdn"]; ok {
        query.WriteString("and obr.name = $1")
    } else {
        query.WriteString("and obr.id = $1")
    }

    if q.lock_ {
        query.WriteString(" FOR UPDATE of obr ")
    } else {
        query.WriteString(" FOR SHARE of obr ")
    }

    return query.String()
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

func GetNumberOfLinkedDomains(db *server.DBConn, hostid uint64) (int, error) {
    query := "SELECT count(distinct domainid) FROM domain_host_map WHERE hostid = $1::bigint"

    row := db.QueryRow(query, hostid)

    var domains int 
    err := row.Scan(&domains)
    if err != nil {
        return 0, err 
    }   

    return domains, err 
}
