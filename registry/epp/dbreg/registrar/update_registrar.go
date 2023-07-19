package registrar

import (
    "strings"
    "strconv"

    "registry/epp/dbreg"
    "registry/server"

    . "registry/epp/eppcom"
)

type UpdateRegistrar struct {
    www NullableVal
    whois NullableVal
    fax NullableVal
    voice NullableVal
    emails NullableVal
}

func NewUpdateRegistrar() UpdateRegistrar {
    obj := UpdateRegistrar{}

    obj.www.Set(nil)
    obj.whois.Set(nil)
    obj.fax.Set(nil)
    obj.voice.Set(nil)
    obj.emails.Set(nil)

    return obj
}

func (u *UpdateRegistrar) SetWWW(www string) *UpdateRegistrar {
    u.www.Set(www)
    return u
}

func (u *UpdateRegistrar) SetWhois(whois string) *UpdateRegistrar {
    u.whois.Set(whois)
    return u
}

func (u *UpdateRegistrar) SetVoice(voice []string) *UpdateRegistrar {
    u.voice.Set(voice)
    return u
}

func (u *UpdateRegistrar) SetFax(fax []string) *UpdateRegistrar {
    u.fax.Set(fax)
    return u
}

func (u *UpdateRegistrar) SetEmails(emails []string) *UpdateRegistrar {
    u.emails.Set(emails)
    return u
}

func (u *UpdateRegistrar) Exec(db *server.DBConn, objectid uint64, regid uint, add_addrs []string, rem_addrs []string) error {
    err := dbreg.LockObjectById(db, objectid, "registrar")
    if err != nil {
        return err
    }

    var params []any
    var fields []string

    if !u.www.IsNull() {
        params = append(params, u.www.Get())
        fields = append(fields, "www = $" + strconv.Itoa(len(params)) + "::text")
    }
    if !u.whois.IsNull() {
        params = append(params, u.whois.Get())
        fields = append(fields, "whois = $" + strconv.Itoa(len(params)) + "::text")
    }
    if !u.fax.IsNull() {
        params = append(params, dbreg.PackJson(u.fax.Get().([]string)))
        fields = append(fields, "fax = $" + strconv.Itoa(len(params)) + "::jsonb")
    }
    if !u.voice.IsNull() {
        params = append(params, dbreg.PackJson(u.voice.Get().([]string)))
        fields = append(fields, "telephone = $" + strconv.Itoa(len(params)) + "::jsonb")
    }
    if !u.emails.IsNull() {
        params = append(params, dbreg.PackJson(u.emails.Get().([]string)))
        fields = append(fields, "email = $" + strconv.Itoa(len(params)) + "::jsonb")
    }

    if len(params) > 0 {
        fields_str := strings.Join(fields, ", ")
        params = append(params, regid)

        _, err = db.Exec("UPDATE registrar SET " + fields_str + " WHERE id = $"+strconv.Itoa(len(params))+"::integer", params...)
        if err != nil {
            return err
        }
    }

    for _, ipaddr := range add_addrs {
        query := "INSERT INTO registrar_ipaddr_map(registrarid, ipaddr) " +
                 "VALUES($1::integer, $2::inet)"
        _, err = db.Exec(query, regid, ipaddr)
        if err != nil {
            return err
        }
    }

    for _, ipaddr := range rem_addrs {
        query := "DELETE FROM registrar_ipaddr_map " +
                 "WHERE registrarid = $1::integer and ipaddr = $2::inet"
        _, err = db.Exec(query, regid, ipaddr)
        if err != nil {
            return err
        }
    }

    err = dbreg.UpdateObject(db, objectid, regid)

    return err
}
