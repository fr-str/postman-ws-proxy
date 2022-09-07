package main

import (
	"os"
	"path/filepath"

	"github.com/main-kube/util/env"
)

type configS struct {
	ProxyAddr        string
	LogLevel         int
	ProxyLogFilePath string
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
	home, _ = os.UserHomeDir()
	config  = configS{
		ProxyAddr:        env.Get("PP_ADDRESS", ":8008"),
		LogLevel:         env.Get("PP_LOG_LEVEL", 1),
		ProxyLogFilePath: env.Get("PP_LOG_FILE_PATH", filepath.Join(home, ".proxylog/")),
	}
	// got home directory
)
