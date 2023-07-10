package contact

import (
    "strings"

    "registry/server"
    "registry/epp/dbreg"
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
    var  query strings.Builder
    query.WriteString("SELECT obr.id AS id ")
    query.WriteString(" , obr.roid AS roid , obr.name AS contact_name ")
    query.WriteString(" , obj.clid AS registrar_id ")
    query.WriteString(" , clr.handle AS registrar_handle, obr.crid AS cr_registrar_id ")
    query.WriteString(" , crr.handle AS cr_registrar_handle, obj.upid AS upd_registrar_id ")
    query.WriteString(" , upr.handle AS upd_registrar_handle ")
    query.WriteString(" , c.contact_type, c.email, c.telephone, c.fax, c.verified ")
    query.WriteString(" , c.birthday::text, c.passport, c.vat, c.intpostal, c.locpostal ")
    query.WriteString(" , c.locaddress, c.intaddress, c.legaladdress ")
    query.WriteString(" , (obr.crdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS created ")
    query.WriteString(" , (obj.trdate AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS transfer_time ")
    query.WriteString(" , (obj.update AT TIME ZONE 'UTC') AT TIME ZONE '" + q.p_local_zone + "' AS update_time ") /*+
        " , obj.authinfopw " +
        " , (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::timestamp AS utc_timestamp " +
        " , (CURRENT_TIMESTAMP AT TIME ZONE '" + q.p_local_zone  +"')::timestamp AS local_timestamp " */

    query.WriteString(" FROM object_registry obr ")
    query.WriteString(" JOIN object obj ON obj.id = obr.id ")
    query.WriteString(" JOIN contact c ON obr.id = c.id ")
    query.WriteString(" JOIN registrar clr ON clr.id = obj.clid JOIN registrar crr ON crr.id = obr.crid ")

    query.WriteString(" LEFT JOIN registrar upr ON upr.id = obj.upid ")

    query.WriteString(" WHERE obr.type = get_object_type_id('contact'::text) ")

    if _, ok := q.params["name"]; ok {
        query.WriteString("and obr.name = lower($1)")
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

    var email, telephone, fax, birthday, passport, taxnumbers pgtype.Text
    var locaddress, intaddress, legaladdress pgtype.Text

    err := row.Scan(&data.Id, &data.Roid, &data.Name,
                    &data.Sponsoring_registrar.Id, &data.Sponsoring_registrar.Handle,
                    &data.Create_registrar.Id, &data.Create_registrar.Handle, &data.Update_registrar.Id, &data.Update_registrar.Handle,
                    &data.ContactType, &email, &telephone, &fax, &data.Verified, &birthday, &passport, &taxnumbers, &data.IntPostal, &data.LocPostal,
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

    data.Passport = dbreg.UnpackJson(passport)
    data.LocAddress = dbreg.UnpackJson(locaddress)
    data.IntAddress = dbreg.UnpackJson(intaddress)
    data.LegalAddress = dbreg.UnpackJson(legaladdress)

    data.Emails = dbreg.UnpackJson(email)
    data.Voice = dbreg.UnpackJson(telephone)
    data.Fax = dbreg.UnpackJson(fax)

    return &data, nil
}

func GetContactIdByHandle(db *server.DBConn, handle string, regid... uint) (uint64, error) {
    contact_id, err := dbreg.GetObjectIdByName(db, handle, "contact", regid...)

    if err != nil {
        if err == pgx.ErrNoRows {
            return 0, &dbreg.ParamError{Val:"contact " + handle + " doesn't exist"}
        }
    }

    return contact_id, err
}

func GetNumberOfLinkedDomains(db *server.DBConn, contactid uint64) (int, error) {
    query := "SELECT count(*) FROM domain WHERE registrant = $1::bigint"

    row := db.QueryRow(query, contactid)

    var domains int
    err := row.Scan(&domains)
    if err != nil {
        return 0, err 
    }   

    return domains, err 
}
