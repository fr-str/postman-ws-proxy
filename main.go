package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	http.HandleFunc("/", Handler)
	log.Info().Msgf("Listening on %s", config.Port)
	go cleanConns()
	http.ListenAndServe(fmt.Sprintf(":%v", config.Port), nil)
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
		origin: conn,
	}
	conns.Set(conn.RemoteAddr().String(), p)
	go p.originReader()
}
