package dbreg

import (
    "registry/server"
    . "registry/epp/eppcom"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgtype"
)

type InfoContactDB struct {
    p_local_zone string
    lock_ bool
    params map[string]string
}

func NewInfoContactDB() InfoContactDB {
    obj := InfoContactDB{}
    obj.p_local_zone = "UTC"
    obj.lock_ = false
    obj.params = map[string]string{}
    return obj
}

func (q *InfoContactDB) create_info_query() string {
    info_host_query := "SELECT obr.id AS id " +
        " , obr.roid AS roid , obr.name AS contact_name " +
        " , obj.clid AS registrar_id " +
        " , clr.handle AS registrar_handle, obr.crid AS cr_registrar_id " +
        " , crr.handle AS cr_registrar_handle, obj.upid AS upd_registrar_id " +
        " , upr.handle AS upd_registrar_handle " +
        " , c.contact_type, c.email, c.telephone, c.fax, c.verified " +
        " , c.birthday::text, c.vat, c.intpostal, c.locpostal " +
        " , c.locaddress, c.intaddress, c.legaladdress " +
        " , (obr.crdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS created " +
        " , (obj.trdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS transfer_time " +
        " , (obj.update AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS update_time " /*+
        " , obj.authinfopw " +
        " , (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::timestamp AS utc_timestamp " +
        " , (CURRENT_TIMESTAMP AT TIME ZONE '" + q.p_local_zone  +"')::timestamp AS local_timestamp " */

    info_host_query += " FROM object_registry obr " +
                       " JOIN object obj ON obj.id = obr.id " +
                       " JOIN contact c ON obr.id = c.id " +
                       " JOIN registrar clr ON clr.id = obj.clid JOIN registrar crr ON crr.id = obr.crid "

    info_host_query += " LEFT JOIN registrar upr ON upr.id = obj.upid "

    info_host_query += " WHERE obr.type = get_object_type_id('contact'::text) "

    if _, ok := q.params["name"]; ok {
        info_host_query += "and obr.name = lower($1)"
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

func (q *InfoContactDB) SetLock(lock bool) *InfoContactDB {
    q.lock_ = lock
    return q
}

func (q *InfoContactDB) SetName(name string) *InfoContactDB {
    q.params["name"] = name
    return q
}

func (q *InfoContactDB) Exec(db *server.DBConn) (*InfoContactData, error) {
    info_query := q.create_info_query()

    row := db.QueryRow(info_query, q.params["name"])
    var data InfoContactData

    var email, telephone, fax, birthday, taxnumbers pgtype.Text
    var locaddress, intaddress, legaladdress pgtype.Text

    err := row.Scan(&data.Id, &data.Roid, &data.Name,
                    &data.Sponsoring_registrar.Id, &data.Sponsoring_registrar.Handle,
                    &data.Create_registrar.Id, &data.Create_registrar.Handle, &data.Update_registrar.Id, &data.Update_registrar.Handle,
                    &data.ContactType, &email, &telephone, &fax, &data.Verified, &birthday, &taxnumbers, &data.IntPostal, &data.LocPostal,
                    &locaddress, &intaddress, &legaladdress,
                    &data.Creation_time, &data.Transfer_time, &data.Update_time)

    if err != nil {
        return nil, err
    }
    if birthday.Status != pgtype.Null {
        data.Birthday = birthday.String
    }
    if taxnumbers.Status != pgtype.Null {
        data.TaxNumbers = taxnumbers.String
    }

    data.LocAddress = unpackJson(locaddress)
    data.IntAddress = unpackJson(intaddress)
    data.LegalAddress = unpackJson(legaladdress)

    data.Emails = unpackJson(email)
    data.Voice = unpackJson(telephone)
    data.Fax = unpackJson(fax)

    return &data, nil
}

func GetContactIdByHandle(db *server.DBConn, handle string, regid... uint) (uint64, error) {
    contact_id, err := getObjectIdByName(db, handle, "contact", regid...)

    if err != nil {
        if err == pgx.ErrNoRows {
            return 0, &ParamError{Val:"contact " + handle + " doesn't exist"}
        }
    }

    return contact_id, err
}
