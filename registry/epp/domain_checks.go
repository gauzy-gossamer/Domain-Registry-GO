package epp 

import (
    "strings"
    "regexp"
    "registry/server"
    "github.com/kpango/glg"
    "golang.org/x/net/idna"
)

var zone_checks = map[string]func(string) bool {
    "dncheck_idna":checkIDNAValidity,
    "dncheck_no_consecutive_hyphens":checkDoubleHyphens,
}

func ConvertIDNA(domain string) (string, error) {
    p := idna.New()
    return p.ToUnicode(domain)
}

func checkIDNAValidity(domain string) bool {
    if !strings.HasPrefix(domain, "xn--") {
        return true
    }
    val, err := ConvertIDNA(domain)
    glg.Error(val)
    if err != nil {
        return false
    }
    return true
}

func checkDoubleHyphens(domain string) bool {
    return !strings.Contains(domain, "--")
}

func iterZoneSpecificChecks(domain string, tests []string) bool {
    for _, funcname := range tests {
        if _, ok := zone_checks[funcname]; !ok {
            glg.Error("incorrect check " +funcname)
            continue
        }
        if !zone_checks[funcname](domain) {
            return false
        }
    }
    return true
}

/* cached values can be used for check domain calls */
type DomainChecker struct {
    init_regexps bool
    regexps []*regexp.Regexp
    zone_tests map[int][]string
}

func NewDomainChecker() *DomainChecker {
    dc := DomainChecker{init_regexps:false}
    dc.zone_tests = make(map[int][]string)
    return &dc
}

func (c *DomainChecker) SetRegexps(regexps []*regexp.Regexp) *DomainChecker {
    c.regexps = regexps
    return c
}

func (c *DomainChecker) SetZoneTests(zoneid int, tests []string) *DomainChecker {
    c.zone_tests[zoneid] = tests
    return c
}

func (c *DomainChecker) CheckZoneSpecificTests(db *server.DBConn, domain string, zoneid int) (bool, error) {
    tests, ok := c.zone_tests[zoneid]
    if ok {
        return iterZoneSpecificChecks(domain, tests), nil
    }
    c.zone_tests[zoneid] = []string{}
    tests = c.zone_tests[zoneid]
    rows, err := db.Query("SELECT name FROM zone_domain_name_validation_checker_map m " +
                    "INNER JOIN enum_domain_name_validation_checker en ON m.checker_id=en.id " +
                    "WHERE zone_id = $1", zoneid)

    if err != nil {
        return false, err
    }
    for rows.Next() {
        var zone_check string
        err := rows.Scan(&zone_check)
        if err != nil {
            return false, err
        }
        tests = append(tests, zone_check)
    }

    return iterZoneSpecificChecks(domain, tests), nil
}

func (c *DomainChecker) GetBlacklistRegexps(db *server.DBConn) (error) {
    rows, err := db.Query("SELECT regexp FROM domain_blacklist")
    if err != nil {
        return nil
    }
    defer rows.Close()

    for rows.Next() {
        var regex string
        err := rows.Scan(&regex)
        if err != nil {
            return err
        }
        compiled_regexp, err := regexp.Compile(regex)
        if err != nil {
            glg.Error("failed to compile regexp ", regex, err)
            continue
        }
        c.regexps = append(c.regexps, compiled_regexp)
    }
    c.init_regexps = true
    return nil
}

func (c *DomainChecker) TestBlacklistRegexps(domain string) (bool, error) {
    if strings.HasPrefix(domain, "xn--") {
        idn_domain, err := ConvertIDNA(domain)
        if err != nil {
            return true, err
        }
        domain = idn_domain
    }
    for _, regex := range c.regexps {
        if regex.MatchString(domain) {
            return true, nil
        }
    }
    return false, nil
}

func (c *DomainChecker) IsDomainRegistrable(db *server.DBConn, domain string, zoneid int) (bool, error) {
    test, err := c.CheckZoneSpecificTests(db, domain, zoneid)
    if err != nil || !test {
        return false, err
    }

    /* test for blocked domains */
    if !c.init_regexps {
        err = c.GetBlacklistRegexps(db)
        if err != nil {
            return false, err
        }
    }
    test, err = c.TestBlacklistRegexps(domain)
    test = !test

    return test, err
}
