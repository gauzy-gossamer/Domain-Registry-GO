package whois_resp

import (
    "errors"
    "time"
    "github.com/jackc/pgtype"
)

var ObjectNotFound = errors.New("object not found")

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

type Registrar struct {
    Handle     string
    Org        string
    Phone      pgtype.Text
    Fax        pgtype.Text
    Email      pgtype.Text
    WWW        pgtype.Text
    Whois      pgtype.Text

    Retrieved time.Time
}
