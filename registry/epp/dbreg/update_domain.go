package dbreg

import (
    "strings"
    "strconv"
    "registry/server"
    . "registry/epp/eppcom"
)

type UpdateDomainDB struct {
    add_hosts []HostObj
    rem_hosts []HostObj
    registrant NullableVal
    description NullableVal
}

func NewUpdateDomainDB() UpdateDomainDB {
    obj := UpdateDomainDB{}
    obj.registrant.Set(nil)
    obj.description.Set(nil)
    return obj
}

func (up *UpdateDomainDB) SetRegistrant(registrant uint64) *UpdateDomainDB {
    up.registrant.Set(registrant)
    return up
}

func (up *UpdateDomainDB) SetDescription(description []string) *UpdateDomainDB {
    up.description.Set(description)
    return up
}

func (up *UpdateDomainDB) SetAddHosts(add_hosts []HostObj) *UpdateDomainDB {
    up.add_hosts = add_hosts
    return up
}

func (up *UpdateDomainDB) SetRemHosts(rem_hosts []HostObj) *UpdateDomainDB {
    up.rem_hosts = rem_hosts
    return up
}

func (up *UpdateDomainDB) Exec(db *server.DBConn, domainid uint64, regid uint) error {
    err := lockObjectById(db, domainid, "domain")
    if err != nil {
        return err
    }

    for _, host := range up.add_hosts {
        _, err = db.Exec("INSERT INTO domain_host_map(domainid, hostid) VALUES($1::bigint, $2::integer)", domainid, host.Id)
        if err != nil {
            return err
        }
    }

    for _, host := range up.rem_hosts {
        _, err = db.Exec("DELETE FROM domain_host_map WHERE domainid = $1::bigint and hostid = $2::integer", domainid, host.Id)
        if err != nil {
            return err
        }
    }
    if !up.registrant.IsNull() || !up.description.IsNull() {
        var params []any
        var fields []string

        if !up.registrant.IsNull() {
            params = append(params, up.registrant.Get())
            fields = append(fields, "registrant = $" + strconv.Itoa(len(params)) + "::bigint")
        }
        if !up.description.IsNull() {
            params = append(params, PackJson(up.description.Get().([]string)))
            fields = append(fields, "description = $" + strconv.Itoa(len(params)) + "::jsonb")
        }
        fields_str := strings.Join(fields, ", ")
        params = append(params, domainid)

        _, err = db.Exec("UPDATE domain SET " + fields_str + " WHERE id = $"+strconv.Itoa(len(params))+"::bigint", params...)
        if err != nil {
            return err
        }
    }

    err = updateObject(db, domainid, regid)

    return err
}
