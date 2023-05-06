package server

import (
    "os"
    "bufio"

    "github.com/go-ini/ini"
    "github.com/kpango/glg"
)

type DBConfig struct {
    host     string
    port     int
    user     string
    password string
    dbname   string
}

type HTTPConfig struct {
    Host string
    Port uint
    CertFile string
    KeyFile string
    UseProxy bool
}

type RegConfig struct {
    DBconf DBConfig
    HTTPConf HTTPConfig
    MaxRegistrarSessions uint
    MaxQueriesPerMinute uint
    SessionTimeout uint
    DomainMinHosts int
    DomainMaxHosts int
    SchemaPath string
    GrpcPort int
    ChargeOperations bool
}

var LogLevelMap = map[string]glg.LEVEL {
    "TRACE":glg.TRACE,
    "LOG":glg.LOG,
    "INFO":glg.INFO,
    "WARN":glg.WARN,
    "ERR":glg.ERR,
    "FATAL":glg.FATAL,
}

func (r *RegConfig) LoadConfig(config_path string)  {
    file, err := os.Open(config_path)
    if err != nil {
        glg.Fatal(err)
    }
    defer func() {
        if err = file.Close(); err != nil {
            glg.Fatal(err)
        }
    }()

    var config string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        config += scanner.Text() + "\n"
    }

    cfg, err := ini.Load([]byte(config))
    if err != nil {
        glg.Fatal(err)
    }
    section, _ := cfg.GetSection("database")

    params := []string {"host", "port", "user", "password", "name"}
    for _,  val := range params {
        key, err := section.GetKey(val)
        if err != nil {
            glg.Fatal(err)
        }
        switch val {
            case "host":
                r.DBconf.host = key.String()
            case "port":
                r.DBconf.port, err = key.Int()
                if err != nil {
                    glg.Fatal(err)
                }
            case "password":
                r.DBconf.password = key.String()
            case "name":
                r.DBconf.dbname = key.String()
            case "user":
                r.DBconf.user = key.String()
        }
    }
    section, _ = cfg.GetSection("reg_server")

    params = []string {"session_registrar_max", "domain_min_hosts", "domain_max_hosts", "schema_path", "epp_operations_charging", "session_timeout", "query_limit"}
    for _,  val := range params {
        key, err := section.GetKey(val)
        if err != nil {
            glg.Fatal(err)
        }
        switch val {
            case "session_registrar_max":
                r.MaxRegistrarSessions, err = key.Uint()
                if err != nil {
                    glg.Fatal(err)
                }
            case "session_timeout":
                r.SessionTimeout, err = key.Uint()
                if err != nil {
                    glg.Fatal(err)
                }
            case "query_limit":
                r.MaxQueriesPerMinute, err = key.Uint()
                if err != nil {
                    glg.Fatal(err)
                }
            case "domain_min_hosts":
                r.DomainMinHosts, err = key.Int()
                if err != nil {
                    glg.Fatal(err)
                }
            case "domain_max_hosts":
                r.DomainMaxHosts, err = key.Int()
                if err != nil {
                    glg.Fatal(err)
                }
            case "schema_path":
                r.SchemaPath = key.String()
            case "epp_operations_charging":
                r.ChargeOperations, err = key.Bool()
                if err != nil {
                    glg.Fatal(err)
                }
        }
    }

    section, _ = cfg.GetSection("http")
    params = []string {"host", "port", "key_file", "cert_file", "nginx_proxy"}
    for _,  val := range params {
        key, err := section.GetKey(val)
        if err != nil {
            glg.Error(err)
            continue
        }
        switch val {
            case "host":
                r.HTTPConf.Host = key.String()
            case "port":
                r.HTTPConf.Port, err = key.Uint()
                if err != nil {
                    glg.Fatal(err)
                }
            case "cert_file":
                r.HTTPConf.CertFile = key.String()
            case "key_file":
                r.HTTPConf.KeyFile = key.String()
            case "nginx_proxy":
                r.HTTPConf.UseProxy, err = key.Bool()
                if err != nil {
                    glg.Fatal(err)
                }
        }
    }

    /* needs support for syslog */
    section, _ = cfg.GetSection("log")
    params = []string {"file", "level"}
    for _,  val := range params {
        key, err := section.GetKey(val)
        if err != nil {
//           check required fields 
            glg.Error(err)
            continue
        }
        switch val {
            case "file":
                logfile := key.String()
                if logfile == "" {
                    break
                }
                logwriter := glg.FileWriter(logfile, 0666)
                glg.Get().SetMode(glg.WRITER).SetWriter(logwriter)
            case "level":
                loglevel := key.String()
                if loglevel_, ok := LogLevelMap[loglevel]; !ok {
                    glg.Fatal("unknown log level", loglevel)
                } else {
                    glg.Get().SetLevel(loglevel_)
                }
        }
    }

    section, _ = cfg.GetSection("grpc")
    params = []string {"port"}
    for _,  val := range params {
        key, err := section.GetKey(val)
        if err != nil {
            glg.Error(err)
            continue
        }
        switch val {
            case "port":
                r.GrpcPort, err = key.Int()
                if err != nil {
                    glg.Fatal(err)
                }
        }
    }
}
