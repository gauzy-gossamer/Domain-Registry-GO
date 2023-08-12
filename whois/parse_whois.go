package main

import (
    "strings"
    "errors"
)

var supported_types = map[string]struct{} {
    "domain":struct{}{},
}

/* returns options map, index end of options, error|nil */
func extractOpts(query string) (map[string]string, int, error) {
    opts := make(map[string]string)
 
    i := 0
    for i + 2 < len(query) && query[i] == '-' {
        i += 1
        if query[i] != 'T' || query[i+1] != ' ' {
            return opts, i, errors.New("unsupported option")
        }
        /* copy option value to val */
        i += 2
        start_i := i
        /* skip until the next space */
        for ; i < len(query) && query[i] != ' '; i++ { }
        val := query[start_i:i]

        if _, ok := supported_types[val]; !ok {
            return opts, i, errors.New("unsupported type")
        }

        opts["type"] = val
    }

    return opts, i, nil
}

func parseWhoisQuery(query string) (map[string]string, string, error) {
    query = strings.TrimSpace(query)

    opts, i, err := extractOpts(query)
    if err != nil {
        return opts, query, err
    }

    query = strings.TrimSpace(query[i:])

    if _, ok := opts["type"]; !ok {
        /* if there are no dots, assume registrar is queried */
        if !strings.Contains(query, ".") {
            opts["type"] = "registrar"
        } else {
            opts["type"] = "domain"
        }
    }

    return opts, query, nil
}
