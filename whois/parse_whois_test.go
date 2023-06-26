package main

import (
    "testing"
    "reflect"
)

type Expected struct {
    err bool
    opts map[string]string
    query string
}

func TestParser(t *testing.T) {
    queries := map[string]Expected {
        "-T domain domain.net.ru":Expected{false, map[string]string{"type":"domain"}, "domain.net.ru"},
        "domain.ex.com":Expected{false, map[string]string{"type":"domain"}, "domain.ex.com"},
        "-T domaindomain.ex.com":Expected{true, map[string]string{}, "-T domaindomain.ex.com"},
        "-T domain ":Expected{false, map[string]string{"type":"domain"}, ""},
        "-T domain":Expected{false, map[string]string{"type":"domain"}, ""},
        "-Tn":Expected{true, map[string]string{}, "-Tn"},
        "-T":Expected{false, map[string]string{"type":"domain"}, "-T"},
        "-":Expected{false, map[string]string{"type":"domain"}, "-"},
        "-q domain.com":Expected{true, map[string]string{}, "-q domain.com"},
        "-T domain domain.com ":Expected{false, map[string]string{"type":"domain"}, "domain.com"},
        " -T domain domain.com ":Expected{false, map[string]string{"type":"domain"}, "domain.com"},
    }

    for query, expected := range queries {
        t.Run(query, func(t *testing.T) {
            opts, parsed_query, err := parseWhoisQuery(query)
            has_err := err != nil 
            if has_err != expected.err || !reflect.DeepEqual(opts, expected.opts) || parsed_query != expected.query {
                t.Error("expected ", expected, err, opts, parsed_query, reflect.DeepEqual(opts, expected.opts))
            }
        })
    }
}

