package dbreg

import (
    "strings"
    "strconv"
    "registry/server"
    . "registry/epp/eppcom"
)

type UpdateContactDB struct {
    p_local_zone string
    regid uint
    verified NullableVal
    emails NullableVal
    voice NullableVal
}

func NewUpdateContactDB() UpdateContactDB {
    obj := UpdateContactDB{}
    obj.verified.Set(nil)
    obj.emails.Set(nil)
    obj.voice.Set(nil)
    return obj
}

func (up *UpdateContactDB) SetVerified(verified bool) *UpdateContactDB {
    up.verified.Set(verified)
    return up
}

func (up *UpdateContactDB) SetEmails(emails []string) *UpdateContactDB {
    up.emails.Set(emails)
    return up
}

func (up *UpdateContactDB) SetVoice(voice []string) *UpdateContactDB {
    up.voice.Set(voice)
    return up
}

func (up *UpdateContactDB) Exec(db *server.DBConn, contactid uint64, regid uint) error {
    err := lockObjectById(db, contactid, "contact")
    if err != nil {
        return err
    }

    if !up.voice.IsNull() || !up.emails.IsNull() || !up.verified.IsNull() {
        var params []any
        var fields []string

        if !up.verified.IsNull() {
            params = append(params, up.verified.Get())
            fields = append(fields, "verified = $" + strconv.Itoa(len(params)) + "::boolean")
        }
        if !up.emails.IsNull() {
            params = append(params, packJson(up.emails.Get().([]string)))
            fields = append(fields, "email = $" + strconv.Itoa(len(params)) + "::jsonb")
        }
        if !up.voice.IsNull() {
            params = append(params, packJson(up.voice.Get().([]string)))
            fields = append(fields, "telephone = $" + strconv.Itoa(len(params)) + "::jsonb")
        }
        fields_str := strings.Join(fields, ", ")
        params = append(params, contactid)

        _, err = db.Exec("UPDATE contact SET " + fields_str + " WHERE id = $"+strconv.Itoa(len(params))+"::bigint", params...)
        if err != nil {
            return err
        }
    }

    err = updateObject(db, contactid, regid)

    return err
}
