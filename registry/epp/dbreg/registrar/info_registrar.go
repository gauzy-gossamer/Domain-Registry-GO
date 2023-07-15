package registrar

import (
    "strings"
    "registry/epp/dbreg"
    . "registry/epp/eppcom"
    "registry/server"

    "github.com/jackc/pgtype"
)

type InfoRegistrarDB struct {
    p_local_zone string
    lock_ bool
    params map[string]string
}

func NewInfoRegistrarDB() InfoRegistrarDB {
    obj := InfoRegistrarDB{}
    obj.p_local_zone = "UTC"
    obj.lock_ = false
    obj.params = map[string]string{}
    return obj
}

func (q *InfoRegistrarDB) create_info_query() string {
    var query strings.Builder
    query.WriteString("SELECT r.id AS id, r.object_id as oid ")
    query.WriteString(" , obr.roid AS roid , r.handle AS handle ")
    query.WriteString(" , r.intpostal, r.intaddress, r.locpostal, r.locaddress ")
    query.WriteString(" , r.legaladdress, r.fax, r.telephone, r.email, r.www, r.whois ")
    query.WriteString(" , obj.clid AS registrar_id ")
    query.WriteString(" , r.handle AS registrar_handle, obr.crid AS cr_registrar_id ")
    query.WriteString(" , crr.handle AS cr_registrar_handle, obj.upid AS upd_registrar_id ")
    query.WriteString(" , upr.handle AS upd_registrar_handle ")
    query.WriteString(" , (obr.crdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS created ")
    query.WriteString(" , (obj.update AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS update_time ")

    query.WriteString(" FROM object_registry obr ")
    query.WriteString(" JOIN object obj ON obj.id = obr.id ")
    query.WriteString(" JOIN registrar r ON r.object_id = obj.id JOIN registrar crr ON crr.id = obr.crid ")

    query.WriteString(" INNER JOIN registraracl acl ON r.id = acl.registrarid ")
    query.WriteString(" LEFT JOIN registrar upr ON upr.id = obj.upid ")

    query.WriteString(" WHERE obr.type = get_object_type_id('registrar'::text) ")

    if _, ok := q.params["handle"]; ok {
        query.WriteString("and r.handle = $1")
    } else {
        query.WriteString("and r.id = $1")
    }

    if q.lock_ {
        query.WriteString(" FOR UPDATE of obr ")
    } else {
        query.WriteString(" FOR SHARE of obr ")
    }

    return query.String()
}

func (q *InfoRegistrarDB) SetLock(lock bool) *InfoRegistrarDB {
    q.lock_ = lock
    return q
}

func (q *InfoRegistrarDB) SetHandle(handle string) *InfoRegistrarDB {
    q.params["handle"] = handle
    return q
}

func (q *InfoRegistrarDB) Exec(db *server.DBConn) (*InfoRegistrarData, error) {
    info_query := q.create_info_query()

    row := db.QueryRow(info_query, q.params["handle"])
    var data InfoRegistrarData

    var intaddress, locaddress, legaladdress, fax, telephone, email pgtype.Text

    err := row.Scan(&data.Id, &data.ObjectID, &data.Roid, &data.Handle,
                    &data.IntPostal, &intaddress, &data.LocPostal, &locaddress,
                    &legaladdress, &fax, &telephone, &email, &data.WWW, &data.Whois,
                    &data.Sponsoring_registrar.Id, &data.Sponsoring_registrar.Handle,
                    &data.Create_registrar.Id, &data.Create_registrar.Handle, &data.Update_registrar.Id, &data.Update_registrar.Handle, &data.Creation_time,
                    &data.Update_time)
    if err != nil {
        return nil, err
    }

    data.LocAddress = dbreg.UnpackJson(locaddress)
    data.IntAddress = dbreg.UnpackJson(intaddress)
    data.LegalAddress = dbreg.UnpackJson(legaladdress)

    data.Emails = dbreg.UnpackJson(email)
    data.Voice = dbreg.UnpackJson(telephone)
    data.Fax = dbreg.UnpackJson(fax)

    return &data, nil
}
