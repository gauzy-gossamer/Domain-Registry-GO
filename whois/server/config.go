package server

import (
    "os"
    "bufio"
    "strings"

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
}

type RegConfig struct {
    DBconf DBConfig
    HTTPConf HTTPConfig
    Header string
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

    section, _ = cfg.GetSection("whois")
    params = []string {"host", "port", "whois_header"}
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
