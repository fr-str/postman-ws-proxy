package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/main-kube/util/safe"
)

type proxy struct {
	origin   *websocket.Conn
	target   *websocket.Conn
	filePath string
}

var (
	conns = safe.Map[string, *proxy]{}
)

func (p *proxy) originReader() {
	// Reading from the connection
	m := map[string]interface{}{}
	for {
		err := p.origin.ReadJSON(&m)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Info().Msgf("Connection closed: %v", err)
				return
			}
		}

		if v, ok := m["Target"].(string); ok && !p.Connected(p.target) {
			if !p.connectTarget(v) {
				continue
			}
		} else if !p.Connected(p.target) {
			log.Error().Msgf("Error during target address read %v", m["Target"])
		}
		if v, ok := m["ProxyFilePath"].(string); ok {
			p.filePath = v
		}
		delete(m, "Target")
		delete(m, "ProxyFilePath")
		log.PrintJSON(m)
		p.writeTarget(m)
	}

}

func (p *proxy) connectTarget(addr string) bool {
	if addr == "" {
		log.Error().Msg("Proxy address is empty")
		return false
	}
	if !strings.HasSuffix(addr, "/") {
		addr += "/"
	}
	if p.target != nil && p.target.RemoteAddr().String() == addr && p.Connected(p.target) {
		return true
	}
	log.Debug().Msgf("Starting new connection: %s", addr)
	conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Error().Msgf("Target dial error: %v", err)
		p.writeOrigin([]byte(fmt.Sprintf("Target dial error: %v", err)))
	}

	p.target = conn
	go p.readTarget()
	return true
}

func (p *proxy) writeTarget(v any) {
	if v == nil || p.target == nil {
		return
	}
	err := p.target.WriteJSON(v)
	if err != nil {
		log.Error().Msgf("Writing target error: %v", err)
	}

}

func (p *proxy) readTarget() {
	if p.target == nil {
		return
	}
	log.Debug().Msgf("Reading target: %s", p.target.RemoteAddr().String())
	for {
		_, b, err := p.target.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Info().Msgf("Connection closed: %v", err)
				p.connectTarget(p.target.RemoteAddr().String())
				return
			}
		}
		p.writeToFile(b)
		p.writeOrigin(b)
	}
}

func (p *proxy) writeOrigin(v []byte) {
	if v == nil {
		return
	}
	err := p.origin.WriteMessage(websocket.TextMessage, v)
	if err != nil {
		log.Error().Msgf("Writing Origin error: %v", err)
	}
}

func (p *proxy) Connected(conn *websocket.Conn) bool {
	return conn != nil && conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)) == nil
}

func (p *proxy) writeToFile(data []byte) {
	if len(data) == 0 {
		return
	}
	f, err := os.OpenFile(p.filePath, os.O_WRONLY|os.O_CREATE /* |os.O_APPEND */, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	file := map[string]interface{}{}
	m := map[string]interface{}{}
	b := make([]byte, 10<<20)
	i, err := f.Read(b)
	log.Info().Msgf("len: %d,%v", i, err)
	json.Unmarshal(b, &file)
	json.Unmarshal(data, &m)
	file[strconv.Itoa(int(time.Now().UnixMicro()))] = m
	data, _ = json.MarshalIndent(file, " ", "  ")
	_, err = f.Write(data)
	if err != nil {
		return
	}

}
