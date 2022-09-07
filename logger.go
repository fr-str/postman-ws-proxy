package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var log logS = initLogger()

type logS struct {
	zerolog.Logger
}

func initLogger() logS {
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "15:04:05.999",
	}
	// time.
	output.FormatTimestamp = func(i interface{}) string {
		return fmt.Sprintf("\033[38:5:241m%v", time.UnixMicro(time.Now().UnixMicro()).Format("Jan 2 15:04:05.000"))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("| %s", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}
	zerolog.SetGlobalLevel(zerolog.Level(config.LogLevel))

	return logS{zerolog.New(output).With().Timestamp().Logger()}
}

func (log *logS) PrintJSON(msg string, v interface{}) {
	log.Logger.Debug().MsgFunc(func() string {
		b, _ := json.MarshalIndent(v, " ", "  ")
		return msg + "\n" + string(b)
	})
}

func (l *logS) Error(p *proxy, v any) {
	log.Logger.Error().Msgf("%v", v)
	p.writeOrigin([]byte(fmt.Sprintf("Proxy: %v", v)))
}
