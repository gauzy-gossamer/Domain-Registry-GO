package xml 

import (
    "strconv"
    "encoding/xml"

    "registry/epp/eppcom"

    "github.com/lestrrat-go/libxml2/types"
    "github.com/lestrrat-go/libxml2/xpath"
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

func parseDSData(ctx *xpath.Context, node types.Node) (eppcom.DSRecord, error) {
    var err error
    ds := eppcom.DSRecord{}
    if err = ctx.SetContextNode(node); err != nil {
        return ds, err
    }

    keytag := xpath.String(ctx.Find("secDNS:keyTag"))
    ds.KeyTag, err = strconv.Atoi(keytag)
    if err != nil {
        return ds, err
    }
    digest_alg := xpath.String(ctx.Find("secDNS:alg"))
    ds.Alg, err = strconv.Atoi(digest_alg)
    if err != nil {
        return ds, err
    }
    digest_type := xpath.String(ctx.Find("secDNS:digestType"))
    ds.DigestType, err = strconv.Atoi(digest_type)
    if err != nil {
        return ds, err
    }
    ds.Digest = xpath.String(ctx.Find("secDNS:digest"))

    flags := xpath.String(ctx.Find("secDNS:keyData/secDNS:flags"))
    ds.Key.Flags, err = strconv.Atoi(flags)
    if err != nil {
        return ds, err
    }
    protocol := xpath.String(ctx.Find("secDNS:keyData/secDNS:protocol"))
    ds.Key.Protocol, err = strconv.Atoi(protocol)
    if err != nil {
        return ds, err
    }
    alg := xpath.String(ctx.Find("secDNS:keyData/secDNS:alg"))
    ds.Key.Alg, err = strconv.Atoi(alg)
    if err != nil {
        return ds, err
    }
    ds.Key.Key = xpath.String(ctx.Find("secDNS:keyData/secDNS:pubKey"))

    return ds, nil
}

func parseDsRecs(ctx *xpath.Context) ([]eppcom.DSRecord, error) {
    nodes := xpath.NodeList(ctx.Find("secDNS:dsData"))

    dsrecs := []eppcom.DSRecord{}

    for _, node := range nodes {
        dsrec, err := parseDSData(ctx, node)
        if err != nil {
            return dsrecs, err
        }
        dsrecs = append(dsrecs, dsrec)
    }

    return dsrecs, nil
}

func parseCreateSecDNS(ctx *xpath.Context) (eppcom.EPPExt, error) {
    ext := eppcom.EPPExt{ExtType:eppcom.EPP_EXT_SECDNS}
    dsrecs, err := parseDsRecs(ctx)
    if err != nil {
        return ext, err
    }
    ext.Content = dsrecs

    return ext, nil
}
