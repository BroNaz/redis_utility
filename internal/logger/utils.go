package logger

import "github.com/rs/zerolog"

func logLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

type Addrs struct {
	value []string
}

func (a *Addrs) String() string {
	return ""
}

func (a *Addrs) Set(s string) error {
	a.value = append(a.value, s)
	return nil
}
func (a *Addrs) Get() []string {
	return a.value
}
