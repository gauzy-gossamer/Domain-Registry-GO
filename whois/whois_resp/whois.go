package whois_resp

import (
    "fmt"
    "time"
    "strings"
    "github.com/jackc/pgtype"
)

type WhoisResponse struct {
    Header string
    Source string
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

func getStates(domain Domain) string {
    var b strings.Builder
    b.WriteString("REGISTERED, ")
    if len(domain.Hosts) == 0 {
        b.WriteString("NOT ")
    }
    b.WriteString("DELEGATED, ")
    if domain.Verified {
        b.WriteString("VERIFIED")
    } else {
        b.WriteString("NOT VERIFIED")
    }

    if domain.PendingDelete {
        b.WriteString(", pendingDelete")
    }

    return b.String()
}

func (w *WhoisResponse) FormatResponse(whois_fields []*WhoisField, retrieved time.Time) string {
    var b strings.Builder
    b.WriteString(w.Header)
    for _, whois_field := range whois_fields {
        b.WriteString(fmt.Sprintf("%-15s%s\n", whois_field.field, whois_field.value))
    }
    b.WriteString(fmt.Sprintf("source:        %s\n", w.Source))
    b.WriteString("\nLast updated on ")
    b.WriteString(retrieved.UTC().Format(time.RFC3339))
    b.WriteByte('\n')

    b.WriteString("\n")
    return b.String()
}

func (w *WhoisResponse) EmptyResponse() string {
    response := w.Header
    response += "No entries found for the selected source(s).\n\n"
    return response
}

func (w *WhoisResponse) FormatDomain(domain Domain) string {
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

    return w.FormatResponse(whois_fields, domain.Retrieved)
}
