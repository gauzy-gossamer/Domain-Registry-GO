package xml

import (
    "fmt"
    "time"
    "bytes"
    . "registry/epp/eppcom"
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

type PersonFields struct {
    XMLName  xml.Name `xml:"person"`
    IntPostal struct {
        Name string `xml:"name,omitempty"`
        Address []string `xml:"address,omitempty"`
    } `xml:"intPostalInfo"`

    Birthday string `xml:"birthday,omitempty"`
    Voice []string `xml:"voice,omitempty"`
    Email []string `xml:"email,omitempty"`

}

type OrgFields struct {
    XMLName  xml.Name `xml:"organization"`
    IntPostal struct {
        Org string `xml:"org,omitempty"`
        Address []string `xml:"address,omitempty"`
    } `xml:"intPostalInfo"`

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

type CDObj struct {
    V struct {
        Name   string `xml:",chardata"`
        Avail  string `xml:"avail,attr,omitempty"`
    } `xml:"name"`
    Reason string `xml:"reason,omitempty"`
}

type DomainCheck struct {
    XMLName  xml.Name `xml:"domain:chkData"`
    XMLNSDom string   `xml:"xmlns:domain,attr"`
    XMLNS    string   `xml:"xmlns,attr,omitempty"`

    CD []CDObj `xml:"cd"`
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

func getCheckState(check_result *CheckResult) (string, string) {
    var avail string
    var reason string

    switch check_result.Result {
        case CD_NOT_APPLICABLE:
            avail = "0"
            reason = "Domain name not applicable."
        case CD_REGISTERED:
            avail = "0"
            reason = "already registered."
        case CD_AVAILABLE:
            avail = "1"
    }
    return avail, reason
}

func CheckDomainResponse(response *EPPResult) *ResDataS {
    check_results := response.Content.([]CheckResult)

    domain_check := &DomainCheck{
        XMLNSDom:DOMAIN_NS,
        XMLNS:DOMAIN_NS,
    }

    for _, res := range check_results {
        avail, reason := getCheckState(&res)
        cd_obj := CDObj{Reason:reason}
        cd_obj.V.Name = res.Name
        cd_obj.V.Avail = avail
        domain_check.CD = append(domain_check.CD, cd_obj)
    }

    return &ResDataS{Obj:domain_check}
}

func DomainResponse(response *EPPResult) *ResDataS {
    domain_data := response.Content.(*InfoDomainData)
    domain := &Domain{
        XMLNSDom:DOMAIN_NS,
        XMLNS:DOMAIN_NS,
        Name:domain_data.Fqdn,
        Roid:domain_data.Roid,
        Registrant:domain_data.Registrant.Handle,
        ClID:domain_data.Sponsoring_registrar.Handle.String,
        CrID:domain_data.Create_registrar.Handle.String,
        UpID:domain_data.Update_registrar.Handle.String,
        UpDate:FormatDatePG(domain_data.Update_time),
        CrDate:FormatDatePG(domain_data.Creation_time),
        ExDate:FormatDatePG(domain_data.Expiration_date),
        Description:domain_data.Description,
    }
    if len(domain_data.Hosts) > 0 {
        host_ns := Hosts{}
        for _,v := range domain_data.Hosts {
            host_ns.Hosts = append(host_ns.Hosts, v)
        }
        domain.NS = host_ns
    }
    for _,v := range domain_data.States {
        domain.States = append(domain.States, ObjectState{Val:v})
    }

    return &ResDataS{Obj:domain}
}

func CreateDomainResponse(response *EPPResult) *ResDataS {
    cre_data := response.Content.(*CreateDomainResult)
    domain := &CreateDomainRes{
        XMLNSDom:DOMAIN_NS,
        XMLNS:DOMAIN_NS,
        Name:cre_data.Name,
        CrDate:FormatDatePG(cre_data.Crdate),
        ExDate:FormatDatePG(cre_data.Exdate),
    }
    return &ResDataS{Obj:domain}
}

func TransferDomainResponse(response *EPPResult) *ResDataS {
    tr_data := response.Content.(*TransferRequestObject)
    trn := &TransferResponse{
        XMLNSDom:DOMAIN_NS,
        XMLNS:DOMAIN_NS,
        Name:tr_data.Domain,
        AcID:tr_data.AcID.Handle.String,
        AcDate:FormatDatePG(tr_data.AcDate),
        ReID:tr_data.ReID.Handle.String,
        ReDate:FormatDatePG(tr_data.ReDate),
        TrStatus:tr_data.Status,
    }

    return &ResDataS{Obj:trn}
}

func HostResponse(response *EPPResult) *ResDataS {
    host_data := response.Content.(*InfoHostData)
    host := &Host{
        XMLNSDom:HOST_NS,
        XMLNS:HOST_NS,
        Name:host_data.Fqdn,
        Roid:host_data.Roid,
        ClID:host_data.Sponsoring_registrar.Handle.String,
        CrID:host_data.Create_registrar.Handle.String,
        UpID:host_data.Update_registrar.Handle.String,
        UpDate:FormatDatePG(host_data.Update_time),
        CrDate:FormatDatePG(host_data.Creation_time),
    }
    for _,v := range host_data.States {
        host.States = append(host.States, ObjectState{Val:v})
    }
    for _, ipaddr := range host_data.Addrs {
        host.Addrs = append(host.Addrs, ipaddr)
    }

    return &ResDataS{Obj:host}
}

func CreateHostResponse(response *EPPResult) *ResDataS {
    cre_data := response.Content.(*CreateObjectResult)
    host := &CreateHostRes{
        XMLNSDom:HOST_NS,
        XMLNS:HOST_NS,
        Name:cre_data.Name,
        CrDate:FormatDatePG(cre_data.Crdate),
    }
    return &ResDataS{Obj:host}
}

func ContactResponse(response *EPPResult) *ResDataS {
    contact_data := response.Content.(*InfoContactData)
    contact := &Contact{
        XMLNSDom:CONTACT_NS,
        XMLNS:CONTACT_NS,
        Name:contact_data.Name,
        Roid:contact_data.Roid,
        ClID:contact_data.Sponsoring_registrar.Handle.String,
        CrID:contact_data.Create_registrar.Handle.String,
        UpID:contact_data.Update_registrar.Handle.String,
        UpDate:FormatDatePG(contact_data.Update_time),
        CrDate:FormatDatePG(contact_data.Creation_time),
    }

    if contact_data.ContactType == CONTACT_ORG {
        org_data := OrgFields{}
        org_data.IntPostal.Org = contact_data.IntPostal
        org_data.Email = contact_data.Emails
        org_data.Fax = contact_data.Fax
        org_data.Voice = contact_data.Voice

        contact.ContactData = org_data
    } else {
        person_data := PersonFields{}
        person_data.IntPostal.Name = contact_data.IntPostal
        person_data.Birthday = contact_data.Birthday
        person_data.Email = contact_data.Emails
        person_data.Voice = contact_data.Voice

        contact.ContactData = person_data
    }

    if contact_data.Verified {
        contact.Verified = &VerifiedField{}
    } else {
        contact.Verified = &UnverifiedField{}
    }

    for _,v := range contact_data.States {
        contact.States = append(contact.States, ObjectState{Val:v})
    }

    return &ResDataS{Obj:contact}
}

func CreateContactResponse(response *EPPResult) *ResDataS {
    cre_data := response.Content.(*CreateObjectResult)
    contact := &CreateContactRes{
        XMLNSDom:CONTACT_NS,
        XMLNS:CONTACT_NS,
        Name:cre_data.Name,
        CrDate:FormatDatePG(cre_data.Crdate),
    }
    return &ResDataS{Obj:contact}
}

func PollReqResponse(response *EPPResult) *MsgQ {
    poll_msg := response.Content.(*PollMessage)
    msg_q := &MsgQ{
        Count:poll_msg.Count,
        MsgId:poll_msg.Msgid,
        Msg:poll_msg.Msg,
        QDate:FormatDatePG(poll_msg.QDate),
    }
    return msg_q
}

func GenerateResponse(response *EPPResult, clTRID string, svTRID string) string {
    w := &bytes.Buffer{}

    v := &EPP{XMLns:EPP_NS, XSI:XSI, Loc:schemaLoc}
    resp := &Response{}
    resp.Result.Code = response.RetCode
    if resp.Result.Code == 2500 {
        resp.Result.Msg = "Command failed; server closing connection"
    } else {
        resp.Result.Msg = response.Msg
    }
    if resp.Result.Code != 1000 {
        if len(response.Errors) > 0 {
            for _,v := range response.Errors {
                resp.Result.ExtValue = append(resp.Result.ExtValue, ExtValueS{Reason:v})
            }
        }
    } else {
        if response.Content != nil {
            switch response.CmdType {
                case EPP_CHECK_DOMAIN:
                    resp.ResData = CheckDomainResponse(response)

                case EPP_INFO_DOMAIN:
                    resp.ResData = DomainResponse(response)
                case EPP_INFO_HOST:
                    resp.ResData = HostResponse(response)
                case EPP_INFO_CONTACT:
                    resp.ResData = ContactResponse(response)

                case EPP_CREATE_DOMAIN:
                    resp.ResData = CreateDomainResponse(response)
                case EPP_CREATE_HOST:
                    resp.ResData = CreateHostResponse(response)
                case EPP_CREATE_CONTACT:
                    resp.ResData = CreateContactResponse(response)

                case EPP_TRANSFER_DOMAIN:
                    resp.ResData = TransferDomainResponse(response)

            }
        }
    }

    if response.CmdType == EPP_POLL_REQ {
        if response.Content != nil {
            resp.MsgQ = PollReqResponse(response)
        }
    }
    resp.TrID.ClTRID = clTRID
    resp.TrID.SvTRID = svTRID
    v.Content = resp

    enc := xml.NewEncoder(w)
    enc.Indent(" ", " ")
    if err := enc.Encode(v); err != nil {
        fmt.Println("error", err)
    }
    return w.String()
}

func GenerateGreeting() string {
    w := &bytes.Buffer{}

    v := &EPP{XMLns:EPP_NS, XSI:XSI, Loc:schemaLoc}
    greeting := &Greeting{}
    greeting.SvID = "RIPN-EPP Server"
    greeting.SvDate = time.Now().UTC().Format(time.RFC3339)
    greeting.SvcMenu.Version = "1.0"
    greeting.SvcMenu.Lang = []string{"en"}

    var objURI []string
    for _, val := range namespaces {
        objURI = append(objURI, val)
    }
    greeting.SvcMenu.ObjURI = objURI
    v.Content = greeting

    enc := xml.NewEncoder(w)
    enc.Indent("", " ")
    if err := enc.Encode(v); err != nil {
        fmt.Println("error", err)
    }
    return w.String()
}
