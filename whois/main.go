package main

import (
    "time"
    "fmt"
    "flag"
    "net/http"
    "log"
    "io"
    "io/ioutil"
    "strings"
    "whois/server"
    "github.com/jackc/pgx/v5"
    "github.com/kpango/glg"
)

var serv server.Server


// {'handle': '153584275_DOMAIN_COM-VRSN', 'parent_handle': '', 'name': 'REDDIT.COM', 'whois_server': '', 'type': 'domain', 
//'terms_of_service_url': 'https://www.verisign.com/domain-names/registration-data-access-protocol/terms-service/index.xhtml', 'copyright_notice': '', 
///'description': [], 'last_changed_date': datetime.datetime(2022, 3, 28, 9, 30, 6, tzinfo=tzutc()), 'registration_date': datetime.datetime(2005, 4, 29, 17, 59, 19, tzinfo=tzutc()), 
//'expiration_date': datetime.datetime(2024, 4, 29, 17, 59, 19, tzinfo=tzutc()), 'url': 'https://rdap.verisign.com/com/v1/domain/REDDIT.COM', 'rir': '', 
//'entities': {'registrar': [{'handle': '292', 'type': 'entity', 'name': 'MarkMonitor Inc.'}], 'abuse': [{'type': 'entity', 'name': '', 'email': 'abusecomplaints@markmonitor.com'}]}, 
//'nameservers': ['NS-1029.AWSDNS-00.ORG', 'NS-1887.AWSDNS-43.CO.UK', 'NS-378.AWSDNS-47.COM', 'NS-557.AWSDNS-05.NET'], 'status': ['client delete prohibited', 'client transfer prohibited', 'client update prohibited', 'server delete prohibited', 'server transfer prohibited', 'server update prohibited'], 'dnssec': False}

func process_query(query string) (string, error) {
    domain_name := strings.TrimSpace(query)

    whois_resp := WhoisResponse{Header:serv.RGconf.Header}

    domain, err := getDomain(domain_name)
    if err != nil {
        if err == pgx.ErrNoRows {
            return whois_resp.EmptyResponse(), nil
        }
        return "", err
    }
    return whois_resp.FormatDomain(domain), nil
}

func handle_root(w http.ResponseWriter, req *http.Request) {
    if req.URL.Path != "/" {
        http.NotFound(w, req)
        return
    }

    query, err := ioutil.ReadAll(req.Body)
    if err != nil {
        io.WriteString(w, "error")
        return
    }

    glg.Info(query)

    start := time.Now()

    whois_response, err := process_query(string(query))
    if err != nil {
        glg.Error(err)
        io.WriteString(w, "error")
        return
    }

    elapsed := time.Since(start)

    glg.Trace("exec took ", elapsed)

    io.WriteString(w, whois_response)
}

func main() {
    config_file := flag.String("config", "server.conf", "filename with config")
    port := flag.Uint("port", 0, "port")

    flag.Parse()

    serv.RGconf.LoadConfig(*config_file)
    if *port > 0 {
        serv.RGconf.HTTPConf.Port = *port
    }
    var err error
    serv.Pool, err = server.CreatePool(&serv.RGconf.DBconf)
    if err != nil {
        log.Fatal(err)
    }

    host_addr := fmt.Sprintf("%s:%v", serv.RGconf.HTTPConf.Host, serv.RGconf.HTTPConf.Port)

    httpserver := &http.Server{
        Addr: host_addr,
    }

    http.HandleFunc("/", handle_root)

    fmt.Println("server is running at", host_addr)

    httpserver.ListenAndServe()
}
