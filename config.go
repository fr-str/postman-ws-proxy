package main

import (
	"github.com/main-kube/util/env"
	"github.com/rs/zerolog"
)

type configS struct {
	Port     string
	LogLevel zerolog.Level
}

// DebugLevel defines debug log level.
// - DebugLevel 0
// InfoLevel defines info log level.
// - InfoLevel 1
// WarnLevel defines warn log level.
// - WarnLevel 2
// ErrorLevel defines error log level.
// - ErrorLevel 3
// FatalLevel defines fatal log level.
// - FatalLevel 4
// PanicLevel defines panic log level.
// - PanicLevel 5
// NoLevel defines an absent log level.
// - NoLevel 6
// Disabled disables the logger.
// - Disabled 7

var (
	config = configS{
		Port:     env.Get("PP_PORT", "8008"),
		LogLevel: env.Get[zerolog.Level]("PP_LOG_LEVEL", 1),
	}
)
