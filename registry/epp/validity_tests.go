package epp

import (
    "time"
    "strings"
    "net"
    "registry/server"
    "registry/epp/dbreg"
)

func testDateValidity(date string) bool {
     _, err := time.Parse("2006-01-02", date)

    return err == nil
}

/*
    if !regexp.MustCompile(`^[A-Za-z0-9\-\.]+$`).MatchString(domain) {
        return false
    }
 only these letters are allowed. this is faster than regexp */
func testAllowedLetters(val string, letters string) bool {
    for i := 0; i < len(val); i ++ {
        if strings.Count(letters, string(val[i])) != 1 {
            return false
        }
    }
    return true
}

func checkContactHandleValidity(id string) bool {
    l := len(id)
    if l > 32 || l < 3 {
        return false
    }

    if !testAllowedLetters(id, "abcdefghijklmnopqrstuvwxyz0123456789-_") {
        return false
    }

    return true
}

func testDomainAllowedCharacters(domain string) bool {
    return testAllowedLetters(domain, "abcdefghijklmnopqrstuvwxyz0123456789-.")
}

func checkDomainValidity(domain string) bool {
    domain = normalizeDomain(domain)
    l := len(domain)
    if l == 0 || l > 253 {
        return false
    }

    if !testDomainAllowedCharacters(domain) {
        return false
    }

    parts := strings.Split(domain, ".")

    for _, v := range parts {
        part_l := len(v)
        if part_l < 1 || v[0] == '-' || v[part_l-1] == '-' || part_l > 63 {
            return false
        }
    }

    return true
}

func normalizeDomainUpper(domain string) string {
    l := len(domain)
    if l > 0 {
        if domain[l-1] == '.' {
            domain = domain[:l-1]
        }
    }
    return strings.ToUpper(domain)
}

func normalizeDomain(domain string) string {
    l := len(domain)
    if l > 0 {
        if domain[l-1] == '.' {
            domain = domain[:l-1]
        }
    }
    return strings.ToLower(domain)
}

func normalizeHosts(hosts []string) []string {
    normalized_hosts := []string{}
    for _, host := range hosts {
        normalized_hosts = append(normalized_hosts, normalizeDomain(host))
    }

    return normalized_hosts
}

func isHandleAvailable(db *server.DBConn, handle string, object_type string) (bool, error) {
    if object_type == "domain" || object_type == "contact" {
        handle = strings.ToLower(handle)
    } else {
        handle = strings.ToUpper(handle)
    }
    row := db.QueryRow("SELECT count(*) FROM object_registry o WHERE o.type=get_object_type_id($1::text) " +
                       "AND o.erdate ISNULL AND o.name=$2::text", object_type, handle)
    var cnt int
    err := row.Scan(&cnt)
    return cnt == 0, err
}

func isContactAvailable(db *server.DBConn, contact string) (bool, error) {
    return isHandleAvailable(db, contact, "contact")
}

func isDomainAvailable(db *server.DBConn, domain string) (bool, error) {
    return isHandleAvailable(db, domain, "domain")
}

/* check if the host is subordinate to any of the zones available to registrar */
func isHostSubordinate(db *server.DBConn, host string, regid uint) (bool, error) {
    host = strings.ToLower(host)
    zones, err := getRegistrarZones(db, regid)
    if err != nil {
        return false, err
    }

    for _, zone := range zones {
        zone_parts := strings.Split(zone, ".")
        parts := strings.Split(host, ".")
        if len(zone_parts) > len(parts) {
            continue
        }
        host_zone := strings.Join(parts[len(parts)-len(zone_parts):], ".") 
        if zone == host_zone {
            return true, nil
        }

    }
    return false, nil
}

func checkIPAddresses(addrs []string) error {
    check_duplicates := make(map[string]bool)
    for _, ipaddr := range addrs {
        if net.ParseIP(ipaddr) == nil {
            return &dbreg.ParamError{Val:"incorrect ip " + ipaddr}
        }
        if _, ok := check_duplicates[ipaddr]; ok {
            return &dbreg.ParamError{Val:"duplicate " + ipaddr}
        }
        check_duplicates[ipaddr] = true
    }
    return nil
}

/* compare two arrays, 
   if stop_condition is false, then the source array should be a subset of the target  
   if stop_condition is true, then the source array shouldn't contain any values from the target
*/
func compareVals[T dbreg.HostObj | string](source_vals []T, target_vals []T, get_val func(T) string, stop_condition bool) string {
    for _, source_val := range source_vals {
        found := false
        for _, target_val := range target_vals {
            if get_val(target_val) ==  get_val(source_val) {
                found = true
                break
            }
        }
        if found == stop_condition {
            return get_val(source_val)
        }
    }
    return ""
}

func getVal(v1 dbreg.HostObj) string {
    return v1.Fqdn
}

func allHostsPresent(source_hosts []dbreg.HostObj, target_hosts []dbreg.HostObj) string {
    return compareVals[dbreg.HostObj](source_hosts, target_hosts, getVal, false)
}

func anyHostsPresent(source_hosts []dbreg.HostObj, target_hosts []dbreg.HostObj) string {
    return compareVals[dbreg.HostObj](source_hosts, target_hosts, getVal, true)
}

func allAddrsPresent(source_ips []string, target_ips []string) string {
    return compareVals[string](source_ips, target_ips, func(v string) string {return v}, false)
}

func anyAddrsPresent(source_ips []string, target_ips []string) string {
    return compareVals[string](source_ips, target_ips, func(v string) string {return v}, true)
}
