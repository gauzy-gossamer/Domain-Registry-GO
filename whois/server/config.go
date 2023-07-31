package server

import (
    "os"
    "bufio"
    "strings"
    "reflect"
    "errors"

    "whois/postgres"

    "github.com/go-ini/ini"
    "github.com/kpango/glg"
)

type HTTPConfig struct {
    Host string
    Port uint
}

type RegConfig struct {
    DBconf postgres.DBConfig
    HTTPConf HTTPConfig
    Header string
    Source string
    CacheCap int
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
    section, err := cfg.GetSection(section_name)
    if err != nil {
        glg.Fatal("no section", section_name, "in config")
    }

    for _,  val := range params {
        key, err := section.GetKey(val.Name)
        var kval any
        if err != nil { 
            if val.Required {
                glg.Fatal(err)
            }
            /* set default value */
            kval = val.Default
        } else {
            switch val.Default.(type) {
                case int:
                kval, err = key.Int()
                if err != nil {
                    glg.Fatal(err)
                }
                case uint:
                kval, err = key.Uint()
                if err != nil {
                    glg.Fatal(err)
                }
                case bool:
                kval, err = key.Bool()
                if err != nil {
                    glg.Fatal(err)
                }
                case string:
                kval = key.String()
            }
        }

        if err = setField(set_to, val.Field, kval); err != nil {
            glg.Fatal(err)
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

    parseSection(cfg, "whois", &r.HTTPConf, []ConfigVal {
        {"Host", "host", "", true},
        {"Port", "port", uint(0), true},
    })

    parseSection(cfg, "whois", r, []ConfigVal {
        {"Source", "source", "", false},
        {"CacheCap", "cache_capacity", 1000, false},
    })

    section, _ := cfg.GetSection("whois")
    params := []string {"whois_header"}
    for _,  val := range params {
        key, err := section.GetKey(val)
        if err != nil {
            glg.Error(err)
            continue
        }
        switch val {
            case "whois_header":
                parts := strings.Split(key.String(), "\\n")
                r.Header = strings.Join(parts, "\r\n")
                r.Header += "\r\n"
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
}
