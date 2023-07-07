package xml

import (
    "encoding/xml"
)

var XSI = "http://www.w3.org/2001/XMLSchema-instance"
var schemaLoc = "http://www.ripn.net/epp/ripn-epp-1.0 ripn-epp-1.0.xsd"

type EPP struct {
    XMLName  xml.Name `xml:"epp"`
    XMLns    string   `xml:"xmlns,attr"`
    XSI      string   `xml:"xmlns:xsi,attr,omitempty"`
    Loc      string   `xml:"xsi:schemaLocation,attr,omitempty"`

    Content interface{}
}

type Greeting struct {
    XMLName   xml.Name `xml:"greeting"`
    SvID string `xml:"svID"`
    SvDate string `xml:"svDate"`
    SvcMenu struct {
        Version string `xml:"version"`
        Lang []string `xml:"lang"`
        ObjURI []string `xml:"objURI"`
    } `xml:"svcMenu"`
}

type ExtValueS struct {
    Reason string   `xml:"reason"`
}

type ObjectState struct {
    Val    string   `xml:"s,attr"`
}

type Hosts struct {
    XMLName xml.Name `xml:"ns"`
    Hosts []string `xml:"hostObj"`
}

type Domain struct {
    XMLName  xml.Name `xml:"domain:infData"`
    XMLNSDom string   `xml:"xmlns:domain,attr"`
    XMLNS    string   `xml:"xmlns,attr,omitempty"`
    Name     string   `xml:"name"`
    Roid     string   `xml:"roid"`
    States   []ObjectState `xml:"status"`
    Registrant string `xml:"registrant"`
    NS       interface{}
    Description []string `xml:"description,omitempty"`
    ClID     string   `xml:"clID"`
    CrID     string   `xml:"crID"`
    CrDate   string   `xml:"crDate"`
    UpID     string   `xml:"upID,omitempty"`
    UpDate   string   `xml:"upDate,omitempty"`
    ExDate   string   `xml:"exDate"`
}

type Host struct {
    XMLName  xml.Name `xml:"host:infData"`
    XMLNSDom string   `xml:"xmlns:host,attr"`
    XMLNS    string   `xml:"xmlns,attr,omitempty"`
    Name     string   `xml:"name"`
    Roid     string   `xml:"roid"`
    States   []ObjectState `xml:"status"`
    Addrs    []string `xml:"addr,omitempty"`
    ClID     string   `xml:"clID"`
    CrID     string   `xml:"crID"`
    CrDate   string   `xml:"crDate"`
    UpID     string   `xml:"upID,omitempty"`
    UpDate   string   `xml:"upDate,omitempty"`
}

type VerifiedField struct {
    XMLName  xml.Name `xml:"verified"`
}

type UnverifiedField struct {
    XMLName  xml.Name `xml:"unverified"`
}

type PersonPostalInfo struct {
    Name string `xml:"name,omitempty"`
    Address []string `xml:"address,omitempty"`
}

type PostalInfo struct {
    Org string `xml:"org,omitempty"`
    Address []string `xml:"address,omitempty"`
}

type LegalInfo struct {
    Address []string `xml:"address,omitempty"`
}

type PersonFields struct {
    XMLName  xml.Name `xml:"person"`
    IntPostal PersonPostalInfo `xml:"intPostalInfo"`
    LocPostal PersonPostalInfo `xml:"locPostalInfo"`

    Birthday string `xml:"birthday,omitempty"`
    Voice []string `xml:"voice,omitempty"`
    Email []string `xml:"email,omitempty"`
}

type OrgFields struct {
    XMLName  xml.Name `xml:"organization"`

    IntPostal PostalInfo `xml:"intPostalInfo"`
    LocPostal PostalInfo `xml:"locPostalInfo"`
    LegalInfo LegalInfo `xml:"legalInfo,omitempty"`

    Voice []string `xml:"voice,omitempty"`
    Email []string `xml:"email,omitempty"`
    Fax []string `xml:"fax,omitempty"`

    Verified interface {}
}

type Contact struct {
    XMLName  xml.Name `xml:"contact:infData"`
    XMLNSDom string   `xml:"xmlns:contact,attr"`
    XMLNS    string   `xml:"xmlns,attr,omitempty"`
    Name     string   `xml:"id"`
    Roid     string   `xml:"roid"`
    States   []ObjectState `xml:"status"`
    ContactData interface {}
    ClID     string   `xml:"clID"`
    CrID     string   `xml:"crID"`
    CrDate   string   `xml:"crDate"`
    UpID     string   `xml:"upID,omitempty"`
    UpDate   string   `xml:"upDate,omitempty"`
    Verified interface {}
}

type Registrar struct {
    XMLName  xml.Name `xml:"registrar:infData"`
    XMLNSDom string   `xml:"xmlns:registrar,attr"`
    XMLNS    string   `xml:"xmlns,attr,omitempty"`

    Handle     string   `xml:"id"`

    IntPostal PostalInfo `xml:"intPostalInfo,omitempty"`
    LocPostal PostalInfo `xml:"locPostalInfo,omitempty"`
    LegalInfo LegalInfo `xml:"legalInfo,omitempty"`
    Voice []string `xml:"voice,omitempty"`
    Fax []string `xml:"fax,omitempty"`
    Email []string `xml:"email,omitempty"`
    WWW string `xml:"www,omitempty"`
    Whois string `xml:"whois,omitempty"`
    Addrs    []string `xml:"addr,omitempty"`

    UpDate   string   `xml:"upDate,omitempty"`
}

type CDObj struct {
    V struct {
        Name   string `xml:",chardata"`
        Avail  string `xml:"avail,attr,omitempty"`
    } `xml:"name"`
    Reason string `xml:"reason,omitempty"`
}

type CDIDObj struct {
    V struct {
        Name   string `xml:",chardata"`
        Avail  string `xml:"avail,attr,omitempty"`
    } `xml:"id"`
    Reason string `xml:"reason,omitempty"`
}

type DomainCheck struct {
    XMLName  xml.Name `xml:"domain:chkData"`
    XMLNSDom string   `xml:"xmlns:domain,attr"`
    XMLNS    string   `xml:"xmlns,attr,omitempty"`

    CD []CDObj `xml:"cd"`
}

type HostCheck struct {
    XMLName  xml.Name `xml:"host:chkData"`
    XMLNSDom string   `xml:"xmlns:host,attr"`
    XMLNS    string   `xml:"xmlns,attr,omitempty"`

    CD []CDObj `xml:"cd"`
}

type ContactCheck struct {
    XMLName  xml.Name `xml:"contact:chkData"`
    XMLNSDom string   `xml:"xmlns:contact,attr"`
    XMLNS    string   `xml:"xmlns,attr,omitempty"`

    CD []CDIDObj `xml:"cd"`
}

type ResDataS struct {
    XMLName  xml.Name `xml:"resData"`
    Obj interface{}
}

type CreateDomainRes struct {
    XMLName  xml.Name `xml:"domain:creData"`
    XMLNSDom string   `xml:"xmlns:domain,attr"`
    XMLNS    string   `xml:"xmlns,attr,omitempty"`

    Name     string   `xml:"name"`
    CrDate   string   `xml:"crDate"`
    ExDate   string   `xml:"exDate"`
}

type CreateContactRes struct {
    XMLName  xml.Name `xml:"contact:creData"`
    XMLNSDom string   `xml:"xmlns:contact,attr"`
    XMLNS    string   `xml:"xmlns,attr,omitempty"`

    Name     string   `xml:"id"`
    CrDate   string   `xml:"crDate"`
}

type CreateHostRes struct {
    XMLName  xml.Name `xml:"host:creData"`
    XMLNSDom string   `xml:"xmlns:host,attr"`
    XMLNS    string   `xml:"xmlns,attr,omitempty"`

    Name     string   `xml:"name"`
    CrDate   string   `xml:"crDate"`
}

type TransferResponse struct {
    XMLName  xml.Name `xml:"domain:trnData"`
    XMLNSDom string   `xml:"xmlns:domain,attr"`
    XMLNS    string   `xml:"xmlns,attr,omitempty"`

    Name     string   `xml:"name"`
    TrStatus string   `xml:"trStatus"`
    ReID     string   `xml:"reID"`
    ReDate   string   `xml:"reDate"`
    AcID     string   `xml:"acID"`
    AcDate   string   `xml:"acDate"`
}

type MsgQ struct {
    XMLName xml.Name `xml:"msgQ"`
    Count   uint     `xml:"count,attr"`
    MsgId   uint     `xml:"id,attr"`
    QDate   string  `xml:"qDate"`
    Msg     string  `xml:"msg"`
}

type Response struct {
    XMLName  xml.Name `xml:"response"`
    Result struct {
        Code     int   `xml:"code,attr"`
        Msg      string   `xml:"msg"`
        ExtValue []ExtValueS `xml:"extValue"`
    } `xml:"result"`
    MsgQ interface{}
    ResData interface{}
    TrID struct {
        ClTRID   string   `xml:"clTRID,omitempty"`
        SvTRID   string   `xml:"svTRID"`
    } `xml:"trID"`
}
