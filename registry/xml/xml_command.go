package xml

import (
   "registry/epp/eppcom"
)

type EPPLogin struct {
    Clid string
    PW string
    Lang uint
    Fingerprint string
}

type XMLCommand struct {
    CmdType int
    Sessionid uint64
    ClTRID string
    SvTRID string
    Content interface{}
}

type CheckObject struct {
    Names []string
}

type InfoObject struct {
    Name string
}

type InfoDomain struct {
    Name string
    AuthInfo string
}

type InfoContact struct {
    Name string
}

type CreateDomain struct {
    Name string
    Registrant string
    Hosts []string
    Description []string
}

type CreateHost struct {
    Name string
    Addr []string
}

type CreateContact struct {
    Fields eppcom.ContactFields
}

type UpdateDomain struct {
    Name string
    Registrant string
    AddHosts []string
    RemHosts []string
    AddStatus []string
    RemStatus []string
    Description []string
}

type UpdateHost struct {
    Name string
    AddAddrs []string
    RemAddrs []string
    AddStatus []string
    RemStatus []string
}

type UpdateContact struct {
    Fields eppcom.ContactFields
    AddStatus []string
    RemStatus []string
}

type RenewDomain struct {
    Name string
    CurExpDate string
    Period string
}

type DeleteObject struct {
    Name string
}

type TransferDomain struct {
    Name string
    AcID string
    ReID string
    OP   int
}
