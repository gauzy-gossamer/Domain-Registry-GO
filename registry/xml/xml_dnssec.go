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

type DsData struct {
    KeyTag     int   `xml:"secDNS:keyTag"`
    Alg        int   `xml:"secDNS:alg"`
    DigestType int   `xml:"secDNS:digestType"`
    Digest     string   `xml:"secDNS:digest"`
    KeyData struct {
        Flags    int   `xml:"secDNS:flags"`
        Protocol int   `xml:"secDNS:protocol"`
        Alg      int   `xml:"secDNS:alg"`
        PubKey   string   `xml:"secDNS:pubKey"`
    } `xml:"secDNS:keyData"`
}

type SecDNS struct {
    XMLName  xml.Name `xml:"secDNS:infData"`
    XMLNS    string   `xml:"xmlns:secDNS,attr"`
    Loc      string   `xml:"xsi:schemaLocation,attr,omitempty"`

    DsData []DsData `xml:"secDNS:dsData"`
}

func SecDNSResponse(content interface{}) interface{} {
    dsrecs, ok := content.([]eppcom.DSRecord)
    if !ok {
        return nil
    }

    response := SecDNS{
        XMLNS:secDNSNS,
        Loc:secDNSLoc,
    }
    for _, dsrec := range dsrecs {
        dsdata := DsData{}
        dsdata.KeyTag = dsrec.KeyTag
        dsdata.Alg = dsrec.Alg
        dsdata.DigestType = dsrec.DigestType
        dsdata.Digest = dsrec.Digest
        dsdata.KeyData.Flags = dsrec.Key.Flags
        dsdata.KeyData.Protocol = dsrec.Key.Protocol
        dsdata.KeyData.Alg = dsrec.Key.Alg
        dsdata.KeyData.PubKey = dsrec.Key.Key

        response.DsData = append(response.DsData, dsdata)
    }

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

func parseUpdateSecDNS(ctx *xpath.Context) (eppcom.EPPExt, error) {
    var err error
    ext := eppcom.EPPExt{ExtType:eppcom.EPP_EXT_SECDNS}
    secupdate := eppcom.SecDNSUpdate{}

    rem_nodes := xpath.NodeList(ctx.Find("secDNS:rem"))
    add_nodes := xpath.NodeList(ctx.Find("secDNS:add"))

    if len(rem_nodes) > 0 {
        if err = ctx.SetContextNode(rem_nodes[0]); err != nil {
            return ext, err
        }
        all := xpath.String(ctx.Find("secDNS:all"))
        if all == "true" {
            secupdate.RemAll = true
        } else {
            secupdate.RemDS, err = parseDsRecs(ctx)
            if err != nil {
                return ext, err
            }
        }
    }

    if len(add_nodes) > 0 {
        if err = ctx.SetContextNode(add_nodes[0]); err != nil {
            return ext, err
        }
        secupdate.AddDS, err = parseDsRecs(ctx)
        if err != nil {
            return ext, err
        }
    }

    ext.Content = secupdate

    return ext, nil
}
