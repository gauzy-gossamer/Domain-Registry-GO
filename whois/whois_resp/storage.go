package whois_resp

import (
    "time"
    "github.com/jackc/pgtype"
)

type Domain struct {
    Id         uint64
    Name       string
    CrDate     pgtype.Timestamp
    ExDate     pgtype.Timestamp
    DeleteDate pgtype.Timestamp
    IntPostal  string
    Verified   bool
    CType      int 
    Registrar  string
    Url        pgtype.Text
    PendingDelete bool
    Inactive      bool
    TrRegistrar pgtype.Text
    Hosts []string

    Retrieved time.Time
}
