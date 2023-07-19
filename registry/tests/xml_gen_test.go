package tests

import (
    "testing"
    "registry/server"
    "registry/xml"
    . "registry/epp/eppcom"
)

func TestSimpleResponse(t *testing.T) {
    xmlparser := prepareParser()
    response := xmlparser.GenerateGreeting()
    if response == "" {
        t.Error(response)
    }

    TrID := server.GenerateTRID(10)

    /* incorrect content */
    info_domain := &EPPResult{CmdType:EPP_INFO_DOMAIN, RetCode:EPP_OK}
    response = xml.GenerateResponse(info_domain, "", TrID)
    if response == "" {
        t.Error(response)
    }
}
