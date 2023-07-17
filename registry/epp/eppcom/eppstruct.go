/* common structures used by other modules */
package eppcom

import (
    "time"
    "github.com/jackc/pgtype"
)

const (
    LANG_EN = iota + 1
    LANG_RU
)

var LanguageMap = map[string]int{
    "en":LANG_EN,
    "ru":LANG_RU,
}

/* transfer operation codes */
const (
    TR_REQUEST = iota
    TR_APPROVE
    TR_CANCEL
    TR_QUERY
    TR_REJECT
)

var TransferOPMap = map[string]int{
    "request":TR_REQUEST,
    "approve":TR_APPROVE,
    "cancel":TR_CANCEL,
    "query":TR_QUERY,
    "reject":TR_REJECT,
}

type TransferRequestObject struct {
    Id      uint
    Domain  string
    StatusId uint
    Status  string
    AcID    RegistrarPair /* acquirer */
    ReID    RegistrarPair /* initiator */
    UpID    RegistrarPair /* who updated it last */
    ReDate  pgtype.Timestamp /* when transfer was created */
    AcDate  pgtype.Timestamp /* when transfer expired */
}

/* check domain states */
const (
    CD_AVAILABLE  = iota // ""
    CD_NOT_APPLICABLE // "Domain name not applicable."
    CD_REGISTERED // "already registered."
)

func FormatDatePG(date pgtype.Timestamp) string {
    if date.Status == pgtype.Null {
        return ""
    }
    return date.Time.Format(time.RFC3339)
}

const (
    EPP_EXT_SECDNS = iota + 1
)

type EPPExt struct {
    ExtType int
    Content interface{}
}

/* SecDNS extension secDNS:update section */
type SecDNSUpdate struct {
    AddDS []DSRecord
    RemDS []DSRecord
    RemAll bool
}

/* main structure returned by ExecuteEPPCommand */
type EPPResult struct {
    CmdType int
    RetCode int
    Msg string
    Errors []string
    Reason string
    // main EPP content
    Content interface{}
    // EPP Extensions
    Ext []EPPExt
}

type CheckResult struct {
    Name string
    Result int
}

type LoginResult struct {
    Sessionid uint64
}

/* contact id & handle */
type RegistrantPair struct {
    Id uint64
    Handle string
}

/* registrar id & handle */
type RegistrarPair struct {
    Id NullableUint
    Handle pgtype.Text
}

type ObjectData struct {
    States []string
    Id uint64   /* id of the object*/
    Roid string /* registry object identifier  */
    Sponsoring_registrar RegistrarPair
    Create_registrar RegistrarPair /* handle of registrar which created the object */
    Update_registrar RegistrarPair /* handle of registrar that changed the object last time */
    Creation_time pgtype.Timestamp /* time of object creation set in local zone*/
    Update_time pgtype.Timestamp   /* time of last update time set in local zone*/
    Transfer_time pgtype.Timestamp /* time of last transfer set in local zone*/
}

/* ContactType */
const (
    CONTACT_PERSON = iota
    CONTACT_ORG
)

type ContactFields struct {
    ContactId string
    IntPostal string
    IntAddress []string
    LocPostal string
    LocAddress []string

    LegalAddress []string
    TaxNumbers string

    Passport []string
    Birthday string

    Emails []string
    Voice []string
    Fax []string

    ContactType int
    Verified NullableBool
}

type InfoContactData struct {
    ObjectData
    ContactFields
    Name string
}

type InfoHostData struct {
    ObjectData

    Fqdn string /* fully qualified domain name */
    Addrs []string /* ip-addresses for subordinate hosts */
}

type InfoRegistrarData struct {
    ObjectData

    ObjectID uint64
    Handle string

    IntPostal pgtype.Text
    IntAddress []string
    LocPostal pgtype.Text
    LocAddress []string

    LegalAddress []string

    Emails []string
    Voice []string
    Fax []string

    WWW pgtype.Text
    Whois pgtype.Text

    Addrs []string
}

type DNSKey struct {
    Id uint64
    Flags int 
    Protocol int 
    Alg int 
    Key string
}

type DSRecord struct {
    Id uint64
    KeyTag int 
    Alg int 
    DigestType int 
    Digest string
    MaxSigLife int 
    Key DNSKey
}

type InfoDomainData struct {
    ObjectData

    Fqdn string /* fully qualified domain name */
    Registrant RegistrantPair /**< registrant contact id and handle, owner of the domain*/

    Expiration_date pgtype.Timestamp /* domain expiration local date */
    Authinfopw pgtype.Text  /* password for domain transfer (not used by current implementation of transfers) */
    Authinfo_valid bool     /* whether authinfo is still valid */
    Cur_time pgtype.Timestamp /* current db timestamp */
    ZoneId int
    Keysetid NullableUint64 /* id of DNSSEC object records */

    Description []string
    Hosts []string
}

type CreateObjectResult struct {
    Id uint64
    Name string
    Crdate pgtype.Timestamp
}

type CreateDomainResult struct {
    CreateObjectResult
    Exdate pgtype.Timestamp
}

type PollMessage struct {
    MsgType uint
    Msgid uint
    Msg   string
    Count uint
    QDate  pgtype.Timestamp /* when poll message was created */
    Content interface{}
}
