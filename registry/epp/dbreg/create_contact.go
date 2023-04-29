package dbreg

import (
    "registry/server"
    . "registry/epp/eppcom"
)

type CreateContactDB struct {
    p_local_zone string
    regid uint
    handle string
    contact_type int
    emails []string
    voice []string
    fax []string
    intpostal string
    birthday string
    verified bool
}

func NewCreateContactDB() CreateContactDB {
    obj := CreateContactDB{}
    obj.p_local_zone = "UTC"
    return obj
}

func (q *CreateContactDB) SetVoice(voice []string) *CreateContactDB {
    q.voice = voice
    return q
}

func (q *CreateContactDB) SetFax(fax []string) *CreateContactDB {
    q.fax = fax
    return q
}

func (q *CreateContactDB) SetIntPostal(intpostal string) *CreateContactDB {
    q.intpostal = intpostal
    return q
}

func (q *CreateContactDB) SetBirthday(birthday string) *CreateContactDB {
    q.birthday = birthday
    return q
}

func (q *CreateContactDB) SetEmails(emails []string) *CreateContactDB {
    q.emails = emails
    return q
}

func (q *CreateContactDB) SetVerified(verified bool) *CreateContactDB {
    q.verified = verified
    return q
}

func (q *CreateContactDB) SetParams(handle string, regid uint, contact_type int) *CreateContactDB {
    q.handle = handle
    q.regid = regid
    q.contact_type = contact_type
    return q
}

func (q *CreateContactDB) Exec(db *server.DBConn) (*CreateObjectResult, error) {
    createObj := NewCreateObjectDB("contact")
    create_result, err := createObj.exec(db, q.handle, q.regid)

    if err != nil {
        return nil, err
    }

    row := db.QueryRow("SELECT crdate::timestamp AT TIME ZONE 'UTC' AT TIME ZONE $1::text " +
                    "  FROM object_registry " +
                    " WHERE id = $2::bigint FOR UPDATE OF object_registry", q.p_local_zone, create_result.Id)
    err = row.Scan(&create_result.Crdate)
    if err != nil {
        return nil, err
    }

    var params []any
    cols := "INSERT INTO contact(id, contact_type, email, telephone, intpostal "
    vals := "VALUES($1::integer, $2::integer, $3::jsonb, $4::jsonb, $5::text "
    params = append(params, create_result.Id)
    params = append(params, q.contact_type)
    params = append(params, packJson(q.emails))
    params = append(params, packJson(q.voice))
    params = append(params, q.intpostal)

    if q.contact_type == CONTACT_ORG {
        params = append(params, packJson(q.fax))
        cols += ", fax)"
        vals += ", $6::jsonb)"
    } else {
        params = append(params, q.birthday)
        cols += ", birthday)"
        vals += ", $6::date)"
    }

    _, err = db.Exec(cols + vals, params...)
    if err != nil {
        return nil, err
    }

    return create_result, nil
}
