package server

import (
    "github.com/kpango/glg"
)

var logLevelMap = map[string]glg.LEVEL {
    "TRACE":glg.TRACE,
    "LOG":glg.LOG,
    "INFO":glg.INFO,
    "WARN":glg.WARN,
    "ERR":glg.ERR,
    "FATAL":glg.FATAL,
}

func SetLogWriter(logfile string) {
    logwriter := glg.FileWriter(logfile, 0666)
	glg.Get().SetMode(glg.WRITER).SetWriter(logwriter)
}

/* only done on startup */
func SetLogLevel(loglevel string) {
	if loglevel_, ok := logLevelMap[loglevel]; !ok {
		glg.Fatal("unknown log level", loglevel)
	} else {
		glg.Get().SetLevel(loglevel_)
	}
}

type Logger struct {
	prefix string
}

func NewLogger(prefix string) Logger {
	return Logger{prefix:prefix}
}

func (l *Logger) SetPrefix(prefix string) {
	l.prefix = "["+prefix+"]"
}

func (l *Logger) Info(params... any) {
	t_params := []any{l.prefix}
	t_params = append(t_params, params...)
    glg.Info(t_params...)
}

func (l *Logger) Error(params... any) {
	t_params := []any{l.prefix}
	t_params = append(t_params, params...)
    glg.Error(t_params...)
}

func (l *Logger) Trace(params... any) {
	t_params := []any{l.prefix}
	t_params = append(t_params, params...)
    glg.Trace(t_params...)
}

func (l *Logger) Fatal(params... any) {
	t_params := []any{l.prefix}
	t_params = append(t_params, params...)
    glg.Fatal(t_params...)
}
