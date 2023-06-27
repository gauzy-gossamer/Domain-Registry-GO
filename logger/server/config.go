package server

import (
    "os"
    "bufio"
    "reflect"
    "errors"

    "logger/postgres"
    "logger/file"
    "logger/logging"

    "github.com/go-ini/ini"
    "github.com/kpango/glg"
)

type RegConfig struct {
    GrpcPort uint
    GrpcHost string
    MetricsPath string
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

func (r *RegConfig) LoadConfig(config_path string, serv *Server)  {
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

    logger_conf := struct { StorageType string }{}

    parseSection(cfg, "logger", &logger_conf, []ConfigVal {
        {"StorageType", "storage_type", "", true},
    })

    if logger_conf.StorageType == "postgres" {
        postgres_storage := postgres.NewPostgresStorage()
        dbconf := postgres.DBConfig{}
        parseSection(cfg, "database", &dbconf, []ConfigVal {
            {"Host", "host", "", true},
            {"Port", "port", 0, true},
            {"Password", "password", "", true},
            {"DBname", "name", "", true},
            {"User", "user", "", true},
        })

        err = postgres_storage.InitModule(&dbconf)
        if err != nil {
            glg.Fatal(err)
        }
        serv.Storage = &postgres_storage
    } else if logger_conf.StorageType == "file" {
        file_storage := filestorage.NewFileStorage()

        parseSection(cfg, "file", &file_storage.Conf, []ConfigVal {
            {"Directory", "directory", "", true},
        })

        err = file_storage.InitModule()
        if err != nil {
            glg.Fatal(err)
        }

        serv.Storage = file_storage
    } else {
        glg.Fatal("unknown storage", logger_conf.StorageType)
    }

    parseSection(cfg, "logger", r, []ConfigVal {
        {"MetricsPath", "metrics", "", true},
        {"GrpcPort", "port", uint(0), true},
        {"GrpcHost", "host", "", true},
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
                logging.SetLogWriter(logfile)
            case "level":
                loglevel := key.String()
                logging.SetLogLevel(loglevel)
        }
    }
}
