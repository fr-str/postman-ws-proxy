package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	http.HandleFunc("/", Handler)
	log.Info().Msgf("Listening on %d", 8008)
	http.ListenAndServe(":8008", nil)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Upgrade raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Msgf("Error during connection upgradation: %v", err)
		return
	}
	log.Debug().Msgf("Incoming websocket connection: %v", conn.RemoteAddr().String())
	p := &proxy{
		origin: conn,
	}
	conns.Set(conn.RemoteAddr().String(), p)
	go p.originReader()
}
