package xml 

import (
    "registry/epp/eppcom"

    "encoding/xml"
)

var secDNSLoc = "urn:ietf:params:xml:ns:secDNS-1.1 secDNS-1.1.xsd"
var secDNSNS = "urn:ietf:params:xml:ns:secDNS-1.1"

type SecDNS struct {
    XMLName  xml.Name `xml:"secDNS:infData"`
    XMLNS    string   `xml:"xmlns:secDNS,attr"`
    Loc      string   `xml:"xsi:schemaLocation,attr,omitempty"`

    DsData struct {
        KeyTag     int   `xml:"keyTag"`
        Alg        int   `xml:"alg"`
        DigestType int   `xml:"digestType"`
        Digest     string   `xml:"digest"`
        KeyData struct {
            Flags    int   `xml:"flags"`
            Protocol int   `xml:"protocol"`
            Alg      int   `xml:"alg"`
            PubKey   string   `xml:"pubKey"`
        } `xml:"keyData"`
    } `xml:"dsData"`
}


func SecDNSResponse(content interface{}) interface{} {
    dsrec, ok := content.(*eppcom.DSRecord)
    if !ok {
        return nil
    }

    response := SecDNS{
        XMLNS:secDNSNS,
        Loc:secDNSLoc,
    }
    response.DsData.KeyTag = dsrec.KeyTag
    response.DsData.Alg = dsrec.Alg
    response.DsData.DigestType = dsrec.DigestType
    response.DsData.Digest = dsrec.Digest
    response.DsData.KeyData.Flags = dsrec.Key.Flags
    response.DsData.KeyData.Protocol = dsrec.Key.Protocol
    response.DsData.KeyData.Alg = dsrec.Key.Alg
    response.DsData.KeyData.PubKey = dsrec.Key.Key

    return response
}
