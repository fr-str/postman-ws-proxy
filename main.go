package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	http.HandleFunc("/api/", Handler)
	log.Info().Msgf("Listening on %s", config.ProxyAddr)
	go cleanConns()
	if err := os.Mkdir(config.ProxyLogFilePath, 0755); err != nil && !strings.Contains(err.Error(), "file exists") {
		panic(err)
	}

	log.Fatal().Msgf("%v", http.ListenAndServe(config.ProxyAddr, nil))
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Upgrade raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Logger.Error().Msgf("Error during connection upgrade: %v", err)
		return
	}
	log.Info().Msgf("Incoming websocket connection: %v", conn.RemoteAddr().String())
	p := &proxy{
		origin:   conn,
		fileName: filepath.Join(config.ProxyLogFilePath, "logs"),
	}
	conns.Set(conn.RemoteAddr().String(), p)
	go p.originReader()
}
