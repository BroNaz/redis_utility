package logger

import (
	"log/syslog"
	"os"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LogConfig struct {
	Output string
	Level  string
}

//пока что тут просто заглушки, а не конфиг логера
func SetupLogging(cfg *LogConfig) {
	zerolog.DurationFieldUnit = time.Microsecond
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000000-07:00"

	level := logLevel(cfg.Level)
	log.Logger = log.Logger.
		Level(level)

	switch cfg.Output {
	case "console", "":
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			NoColor:    !isatty.IsTerminal(os.Stdout.Fd()),
			TimeFormat: "2006-01-02 15:04:05.000",
		})
	case "syslog":
		syswr, err := syslog.New(syslog.LOG_DEBUG|syslog.LOG_LOCAL0, "mad")
		if err != nil {
			log.Fatal().Err(err).Msg("не удалось подключиться к syslog")
		}
		log.Logger = log.Output(zerolog.SyslogLevelWriter(syswr))
	default:
		log.Fatal().Str("name", cfg.Output).Msg("неизвестный Log.Output в конфиге")
	}
}
