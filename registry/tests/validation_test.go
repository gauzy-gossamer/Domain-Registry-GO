/*
go test -coverpkg=./... -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
*/
package tests

import (
    "testing"
    "regexp"
    "registry/server"
    "registry/epp"
    "github.com/kpango/glg"
)

func prepareDB() *server.DBConn {
    conf := server.RegConfig{}
    conf.LoadConfig("../server.conf")

    pool, err := server.CreatePool(&conf.DBconf)
    if err != nil {
        glg.Fatal(err)
    }

    dbconn, err := server.AcquireConn(pool, server.NewLogger(""))
    if err != nil {
        glg.Fatal(err)
    }

    return dbconn
}

func TestRegexpMatch(t *testing.T) {
    domains := map[string]bool{
        "domain.net.ru":false,
        "xn--b1agh1afp.ru":true,
        "domain..ru":false,
        "-domain.net.ru":false,
        "-d?omain.net.ru":false,
        "d=omain.net.ru":false,
        "domain.net.ru=":false,
        "DOMAIN.NET.RU":false,
        "astro.def.com":true,
        "metropolis.def.com":false,
    }

    rg := regexp.MustCompile(`привет`)
    var regexps []*regexp.Regexp
    regexps = append(regexps, rg)
    regexps = append(regexps, regexp.MustCompile(`^astr`))
    regexps = append(regexps, regexp.MustCompile(`^polis`))

    domain_checker := epp.NewDomainChecker().SetRegexps(regexps)

    for domain, expected := range domains {
        t.Run(domain, func(t *testing.T) {
            result, err := domain_checker.TestBlacklistRegexps(domain)
            if err != nil {
                t.Error("regex match failed ", err)
            }
            if result != expected {
                t.Error("regex match incorrect ", domain)
            }
        })
    }
}

func TestIDNZoneSpecific(t *testing.T) {
    dbconn := prepareDB()
    domain_checker := epp.NewDomainChecker()

    domains := map[string]bool{
        "domain.net.ru":true,
        "xn--b1agh1afp.ru":true,
        "xn--b1a+agh1afp.ru":false,
    }
    domain_checker.SetZoneTests(1, []string{"dncheck_idna"})

    for domain, expected := range domains {
        t.Run(domain, func(t *testing.T) {
            result, err := domain_checker.CheckZoneSpecificTests(dbconn, domain, 1)
            if err != nil {
                t.Error("zone test failed ", err)
            }
            if result != expected {
                t.Error("zone test incorrect ", domain)
            }
        })
    }

    domains = map[string]bool{
        "domain.ex.com":true,
        "xn--b1agh1afp.ex.com":false,
        "xn--b1a+agh1afp.ex.com":false,
    }
    domain_checker.SetZoneTests(1, []string{"dncheck_no_idn_punycode", "nonexistant"})

    for domain, expected := range domains {
        t.Run(domain, func(t *testing.T) {
            result, err := domain_checker.CheckZoneSpecificTests(dbconn, domain, 1)
            if err != nil {
                t.Error("zone test failed ", err)
            }
            if result != expected {
                t.Error("zone test incorrect ", domain)
            }
        })
    }
}

func TestSizeZoneSpecific(t *testing.T) {
    dbconn := prepareDB()
    domain_checker := epp.NewDomainChecker()
    domains := map[string]bool{
        "domain.ex.com":true,
        "p.com":false,
        "1.com":false,
        "12.com":true,
    }
    domain_checker.SetZoneTests(1, []string{"dncheck_no_single_character"})

    for domain, expected := range domains {
        t.Run(domain, func(t *testing.T) {
            result, err := domain_checker.CheckZoneSpecificTests(dbconn, domain, 1)
            if err != nil {
                t.Error("zone test failed ", err)
            }
            if result != expected {
                t.Error("zone test incorrect ", domain)
            }
        })
    }
}

func TestHyphensZoneSpecific(t *testing.T) {
    dbconn := prepareDB()
    domain_checker := epp.NewDomainChecker()
    domains := map[string]bool{
        "domain.ex.com":true,
        "p.com":true,
        "pp--1.com":false,
        "pp-1.com":true,
        "ppp--1.com":true,
        "ppp-1.com":true,
    }
    domain_checker.SetZoneTests(1, []string{"dncheck_no_34_hyphens"})

    for domain, expected := range domains {
        t.Run(domain, func(t *testing.T) {
            result, err := domain_checker.CheckZoneSpecificTests(dbconn, domain, 1)
            if err != nil {
                t.Error("zone test failed ", err)
            }
            if result != expected {
                t.Error("zone test incorrect ", domain)
            }
        })
    }

    domains = map[string]bool{
        "domain.ex.com":true,
        "p.com":true,
        "pp--1.com":false,
        "ppp--1.com":false,
        "ppp-1.com":true,
    }
    domain_checker.SetZoneTests(1, []string{"dncheck_no_consecutive_hyphens"})

    for domain, expected := range domains {
        t.Run(domain, func(t *testing.T) {
            result, err := domain_checker.CheckZoneSpecificTests(dbconn, domain, 1)
            if err != nil {
                t.Error("zone test failed ", err)
            }
            if result != expected {
                t.Error("zone test incorrect ", domain)
            }
        })
    }
}

/*
func TestDomainValidity(t *testing.T) {
    domains := map[string]bool{
        "domain.net.ru":true,
        "dOOain.net.ru.":true,
        "domain..ru":false,
        "-domain.net.ru":false,
        "-d?omain.net.ru":false,
        "d=omain.net.ru":false,
        "domain.net.ru=":false,
        "DOMAIN.NET.RU":true,
    }

    for domain, expected := range domains {
        t.Run(domain, func(t *testing.T) {
            result := epp.checkDomainValidity(domain)
            if result != expected {
                t.Errorf("test on %s failed", domain)
            }
        })
    }
}

func TestAllowedLetters(t *testing.T) {
    vals := map[string]bool{
        "contact_ab0-10":true,
        "contact*9":false,
        " contact9":false,
    }
    for val, expected := range vals {
        t.Run(val, func(t *testing.T) {
            result := testAllowedLetters(val, "abcdefghijklmnopqrstuvwxyz0123456789-_")
            if result != expected {
                t.Errorf("test on %s failed", val)
            }
        })
    }
}
*/
