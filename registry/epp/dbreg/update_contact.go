package dbreg

import (
    "strings"
    "strconv"
    "registry/server"
    . "registry/epp/eppcom"
)

type UpdateContactDB struct {
    intpostal NullableVal
    intaddress NullableVal
    locpostal NullableVal
    locaddress NullableVal

    legaladdress NullableVal
    taxnumbers NullableVal

    passport NullableVal
    birthday NullableVal

    emails NullableVal
    voice NullableVal
    fax NullableVal

    verified NullableVal
}

func NewUpdateContactDB() UpdateContactDB {
    obj := UpdateContactDB{}

    obj.intpostal.Set(nil)
    obj.intaddress.Set(nil)
    obj.locpostal.Set(nil)
    obj.locaddress.Set(nil)

    obj.legaladdress.Set(nil)
    obj.taxnumbers.Set(nil)

    obj.birthday.Set(nil)
    obj.passport.Set(nil)

    obj.emails.Set(nil)
    obj.voice.Set(nil)
    obj.fax.Set(nil)

    obj.verified.Set(nil)

    return obj
}

func (up *UpdateContactDB) SetIntPostal(intpostal string) *UpdateContactDB {
    up.intpostal.Set(intpostal)
    return up
}

func (up *UpdateContactDB) SetIntAddress(intaddress []string) *UpdateContactDB {
    up.intaddress.Set(intaddress)
    return up
}

func (up *UpdateContactDB) SetLocPostal(locpostal string) *UpdateContactDB {
    up.locpostal.Set(locpostal)
    return up
}

func (up *UpdateContactDB) SetLocAddress(locaddress []string) *UpdateContactDB {
    up.locaddress.Set(locaddress)
    return up
}

func (up *UpdateContactDB) SetLegalAddress(legaladdress []string) *UpdateContactDB {
    up.legaladdress.Set(legaladdress)
    return up
}

func (up *UpdateContactDB) SetTaxNumbers(taxnumbers string) *UpdateContactDB {
    up.taxnumbers.Set(taxnumbers)
    return up
}

func (up *UpdateContactDB) SetBirthday(birthday string) *UpdateContactDB {
    up.birthday.Set(birthday)
    return up
}

func (up *UpdateContactDB) SetPassport(passport []string) *UpdateContactDB {
    up.passport.Set(passport)
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

func (up *UpdateContactDB) SetFax(fax []string) *UpdateContactDB {
    up.fax.Set(fax)
    return up
}

func (up *UpdateContactDB) SetVerified(verified bool) *UpdateContactDB {
    up.verified.Set(verified)
    return up
}

func (up *UpdateContactDB) Exec(db *server.DBConn, contactid uint64, regid uint) error {
    err := LockObjectById(db, contactid, "contact")
    if err != nil {
        return err
    }

    var params []any
    var fields []string

    if !up.intpostal.IsNull() {
        params = append(params, up.intpostal.Get())
        fields = append(fields, "intpostal = $" + strconv.Itoa(len(params)) + "::text")
    }

    if !up.intaddress.IsNull() {
        params = append(params, PackJson(up.intaddress.Get().([]string)))
        fields = append(fields, "intaddress = $" + strconv.Itoa(len(params)) + "::jsonb")
    }

    if !up.locpostal.IsNull() {
        params = append(params, up.locpostal.Get())
        fields = append(fields, "locpostal = $" + strconv.Itoa(len(params)) + "::text")
    }

    if !up.locaddress.IsNull() {
        params = append(params, PackJson(up.locaddress.Get().([]string)))
        fields = append(fields, "locaddress = $" + strconv.Itoa(len(params)) + "::jsonb")
    }

    if !up.legaladdress.IsNull() {
        params = append(params, PackJson(up.legaladdress.Get().([]string)))
        fields = append(fields, "legaladdress = $" + strconv.Itoa(len(params)) + "::jsonb")
    }

    if !up.taxnumbers.IsNull() {
        params = append(params, up.taxnumbers.Get())
        fields = append(fields, "vat = $" + strconv.Itoa(len(params)) + "::text")
    }

    if !up.passport.IsNull() {
        params = append(params, PackJson(up.passport.Get().([]string)))
        fields = append(fields, "passport = $" + strconv.Itoa(len(params)) + "::jsonb")
    }

    if !up.birthday.IsNull() {
        params = append(params, up.birthday.Get())
        fields = append(fields, "birthday = $" + strconv.Itoa(len(params)) + "::date")
    }

    if !up.emails.IsNull() {
        params = append(params, PackJson(up.emails.Get().([]string)))
        fields = append(fields, "email = $" + strconv.Itoa(len(params)) + "::jsonb")
    }
    if !up.voice.IsNull() {
        params = append(params, PackJson(up.voice.Get().([]string)))
        fields = append(fields, "telephone = $" + strconv.Itoa(len(params)) + "::jsonb")
    }
    if !up.fax.IsNull() {
        params = append(params, PackJson(up.fax.Get().([]string)))
        fields = append(fields, "fax = $" + strconv.Itoa(len(params)) + "::jsonb")
    }
    if !up.verified.IsNull() {
        params = append(params, up.verified.Get())
        fields = append(fields, "verified = $" + strconv.Itoa(len(params)) + "::boolean")
    }

    if len(params) > 0 {
        fields_str := strings.Join(fields, ", ")
        params = append(params, contactid)

        _, err = db.Exec("UPDATE contact SET " + fields_str + " WHERE id = $"+strconv.Itoa(len(params))+"::bigint", params...)
        if err != nil {
            return err
        }
    }

    err = UpdateObject(db, contactid, regid)

    return err
}
