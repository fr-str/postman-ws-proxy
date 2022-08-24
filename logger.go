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
		fmt.Print("\033[H\033[2J")
		return fmt.Sprintf("\033[38:5:241m%v", time.UnixMicro(time.Now().UnixMicro()).Format("15:04:05.000"))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("| %s", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s:", i)
	}

	return logS{zerolog.New(output).With().Timestamp().Logger()}
}

func (log *logS) PrintJSON(v interface{}) {
	log.Info().MsgFunc(func() string {
		b, _ := json.MarshalIndent(v, " ", "  ")
		return "\n" + string(b)
	})

}
