package server

import (
    "os"
    "bufio"
    "reflect"
    "errors"

    "github.com/go-ini/ini"
    "github.com/kpango/glg"
)

type DBConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBname   string
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

type ConfigVal struct {
    Field string
    Name string
    Default any
    Required bool
}

func setField(v interface{}, name string, value any) error {
    fv := reflect.ValueOf(v).Elem().FieldByName(name)

    switch value := value.(type) {
        case string:
            fv.SetString(value)
        case uint:
            fv.SetUint(uint64(value))
        case int:
            fv.SetInt(int64(value))
        case bool:
            fv.SetBool(value)
        default:
            return errors.New("unknown type")
    }
    return nil
}

func parseSection(cfg *ini.File, section_name string, set_to interface{}, params []ConfigVal) {
    section, _ := cfg.GetSection(section_name)

    for _,  val := range params {
        key, err := section.GetKey(val.Name)
        if err != nil && val.Required {
            glg.Fatal(err)
        }
        switch val.Default.(type) {
            case int:
                kval, err := key.Int()
                if err != nil {
                    glg.Fatal(err)
                }
                err = setField(set_to, val.Field, kval)
                if err != nil {
                    glg.Fatal(err)
                }
            case uint:
                kval, err := key.Uint()
                if err != nil {
                    glg.Fatal(err)
                }
                err = setField(set_to, val.Field, kval)
                if err != nil {
                    glg.Fatal(err)
                }
            case bool:
                kval, err := key.Bool()
                if err != nil {
                    glg.Fatal(err)
                }
                err = setField(set_to, val.Field, kval)
                if err != nil {
                    glg.Fatal(err)
                }
            case string:
                err = setField(set_to, val.Field, key.String())
                if err != nil {
                    glg.Fatal(err)
                }
        }
    }
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

    parseSection(cfg, "database", &r.DBconf, []ConfigVal {
        {"Host", "host", "", true},
        {"Port", "port", 0, true},
        {"Password", "password", "", true},
        {"DBname", "name", "", true},
        {"User", "user", "", true},
    })

    parseSection(cfg, "reg_server", r, []ConfigVal {
        {"MaxRegistrarSessions", "session_registrar_max", uint(0), true},
        {"SessionTimeout", "session_timeout", uint(0), true},
        {"MaxQueriesPerMinute", "query_limit", uint(0), false},
        {"DomainMinHosts", "domain_min_hosts", 0, true},
        {"DomainMaxHosts", "domain_max_hosts", 0, true},
        {"SchemaPath", "schema_path", "", true},
        {"ChargeOperations", "epp_operations_charging", false, true},
    })

    parseSection(cfg, "http", &r.HTTPConf, []ConfigVal {
        {"Host", "host", "", true},
        {"Port", "port", uint(0), true},
        {"CertFile", "cert_file", "", false},
        {"KeyFile", "key_file", "", false},
        {"UseProxy", "nginx_proxy", false, false},
    })

    parseSection(cfg, "grpc", r, []ConfigVal {
        {"GrpcPort", "port", 0, true},
    })

    /* needs support for syslog */
    section, _ := cfg.GetSection("log")
    params := []string {"file", "level"}
    for _,  val := range params {
        key, err := section.GetKey(val)
        if err != nil {
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
}
