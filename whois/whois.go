package main

import (
    "fmt"
    "time"
    "github.com/jackc/pgtype"
)

type WhoisResponse struct {
    Header string
}

type WhoisField struct {
    field string
    value string
}

func FormatDatetimePG(date pgtype.Timestamp) string {
    if date.Status == pgtype.Null {
        return ""
    }
    return date.Time.Format(time.RFC3339)
}

func FormatDatePG(date pgtype.Timestamp) string {
    if date.Status == pgtype.Null {
        return ""
    }
    return date.Time.Format("2006-01-01")
}

func getStates(domain *Domain) string {
    states := "REGISTERED, "
    if len(domain.Hosts) == 0 {
        states += "NOT "
    }
    states += "DELEGATED, "
    if domain.Verified {
        states += "VERIFIED"
    } else {
        states += "NOT VERIFIED"
    }

    if domain.PendingDelete {
        states += ", pendingDelete"
    }

    return states
}

func (w *WhoisResponse) FormatResponse(whois_fields []*WhoisField) string {
    response := w.Header
    for _, whois_field := range whois_fields {
        response += fmt.Sprintf("%-15s%s\r\n", whois_field.field, whois_field.value)
    }
    return response
}

func (w *WhoisResponse) EmptyResponse() string {
    response := w.Header
    response += "No entries found for the selected source(s).\r\n"
    return response
}

func (w *WhoisResponse) FormatDomain(domain *Domain) string {
    whois_fields := []*WhoisField{}
    whois_fields = append(whois_fields, &WhoisField{field:"domain:", value:domain.Name})
    for _, host := range domain.Hosts {
        whois_fields = append(whois_fields, &WhoisField{field:"nserver:", value:host})
    }

    whois_fields = append(whois_fields, &WhoisField{field:"state:", value:getStates(domain)})
    /* CONTACT_ORG*/
    if domain.CType == 1 {
        whois_fields = append(whois_fields, &WhoisField{field:"org:", value:domain.IntPostal})
    } else {
        whois_fields = append(whois_fields, &WhoisField{field:"person:", value:"Private Person"})
    }

    whois_fields = append(whois_fields, &WhoisField{field:"registrar:", value:domain.Registrar})
    if domain.Url.Status != pgtype.Null {    
        whois_fields = append(whois_fields, &WhoisField{field:"admin-contact:", value:domain.Url.String})
    }

    whois_fields = append(whois_fields, &WhoisField{field:"created:", value:FormatDatetimePG(domain.CrDate)})
    whois_fields = append(whois_fields, &WhoisField{field:"paid-till:", value:FormatDatetimePG(domain.ExDate)})
    whois_fields = append(whois_fields, &WhoisField{field:"free-date:", value:FormatDatePG(domain.DeleteDate)})

    return w.FormatResponse(whois_fields)
}
