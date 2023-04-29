package dbreg

import (
     "encoding/json"
     "github.com/jackc/pgtype"
)

func packJson(vals []string) string {
    bytes, _ := json.Marshal(vals)
    return string(bytes)
}

func unpackJson(val pgtype.Text) []string {
    var result []string
    if val.Status != pgtype.Null {
        _ = json.Unmarshal([]byte(val.String), &result)
    }

    return result
}

