package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/main-kube/util/safe"
)

type proxy struct {
	origin   *websocket.Conn
	target   *websocket.Conn
	fileName string
}

var (
	conns = safe.Map[string, *proxy]{}
)

// just to make sure map is not holding closed conns
func cleanConns() {
	for I := range conns.Iter() {
		if !I.Value.connected(I.Value.origin) {
			I.Value.target.Close()
			conns.Delete(I.Key)
		}
	}
}

func (p *proxy) connected(conn *websocket.Conn) bool {
	return conn != nil && conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)) == nil
}

func (p *proxy) originReader() {
	// Reading from the connection
	m := map[string]interface{}{}
	for {
		err := p.origin.ReadJSON(&m)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Info().Msgf("Connection closed: %v", err)
				conns.Delete(p.origin.RemoteAddr().String())
				return
			}
		}

		if v, ok := m["ProxyTarget"].(string); ok && !p.connected(p.target) {
			if !p.connectTarget(v) {
				continue
			}

		} else if !p.connected(p.target) {

			log.Error(p, fmt.Sprintf("Error during target address read %v", m["ProxyTarget"]))

		}
		if v, ok := m["ProxyFileName"].(string); ok {
			p.fileName = filepath.Join("/app/log-files", v)
		} else {
			p.fileName = "/app/log-files/logs"

		}

		delete(m, "ProxyTarget")
		delete(m, "ProxyFilePath")
		log.PrintJSON(m)
		p.writeTarget(m)
	}
}

func (p *proxy) writeOrigin(v []byte) {
	if v == nil {
		return
	}

	err := p.origin.WriteMessage(websocket.TextMessage, v)
	if err != nil {
		log.Logger.Error().Msgf("Writing Origin error: %v", err)
	}
}

func (p *proxy) connectTarget(addr string) bool {
	if addr == "" {
		log.Error(p, "Target address is empty")

		return false
	}

	addrL := strings.Split(addr, "?")

	if !strings.HasSuffix(addrL[0], "/") {
		addrL[0] += "/"
	}

	if p.target != nil && p.target.RemoteAddr().String() == addrL[0] && p.connected(p.target) {
		return true
	}

	var header http.Header
	if len(addrL) > 1 {
		header = http.Header{}
		for _, element := range strings.Split(addrL[1], "&") {
			kv := strings.Split(element, "=")
			header.Add(kv[0], kv[1])
		}
	}

	log.Info().Msgf("Starting new connection: %s", addr)
	conn, _, err := websocket.DefaultDialer.Dial(addr, header)
	if err != nil {
		log.Error(p, fmt.Sprintf("Target dial error: %v", err))
	}

	p.target = conn
	go p.readTarget()
	return true
}

func (p *proxy) readTarget() {
	if p.target == nil {
		return
	}
	log.Info().Msgf("Reading target: %s", p.target.RemoteAddr().String())
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

func (p *proxy) writeTarget(v any) {
	if v == nil || p.target == nil {
		return
	}
	err := p.target.WriteJSON(v)
	if err != nil {
		log.Error(p, fmt.Sprintf("Writing target error: %v", err))
	}

}

func (p *proxy) writeToFile(data []byte) {
	if len(data) == 0 {
		return
	}
	log.Debug().Msgf("logFile %s", p.fileName)
	f, err := os.OpenFile(p.fileName, os.O_WRONLY|os.O_CREATE /* |os.O_APPEND */, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	// black magic
	// --------------------------------------------------
	file := map[string]interface{}{}
	m := map[string]interface{}{}
	// doing f.Read() returns 'bad file descriptor' and it's late, will fix it later
	// TODO fix
	b, err := os.ReadFile(p.fileName)
	if err != nil {
		log.Error(p, fmt.Sprintf("Writing to file: %v", err))
	}

	json.Unmarshal(b, &file)
	json.Unmarshal(data, &m)
	file[strconv.Itoa(int(time.Now().UnixMicro()))] = m
	data, _ = json.MarshalIndent(file, " ", "  ")
	// --------------------------------------------------

	_, err = f.Write(data)
	if err != nil {
		return
	}

}
