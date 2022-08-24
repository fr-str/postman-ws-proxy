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
			msg := fmt.Sprintf("Error during target address read %v", m["Target"])
			log.Error().Msg(msg)
			p.writeOrigin([]byte(msg), true)

		}
		if v, ok := m["ProxyFileName"].(string); ok {
			p.fileName = filepath.Join("/app/log-files", v)
		}

		delete(m, "Target")
		delete(m, "ProxyFilePath")
		log.PrintJSON(m)
		p.writeTarget(m)
	}

}

func (p *proxy) connectTarget(addr string) bool {
	if addr == "" {
		msg := "Target address is empty"
		log.Error().Msg(msg)
		p.writeOrigin([]byte(msg), true)

		return false
	}

	addrL := strings.Split(addr, "?")

	if !strings.HasSuffix(addrL[0], "/") {
		addrL[0] += "/"
	}
	if p.target != nil && p.target.RemoteAddr().String() == addrL[0] && p.Connected(p.target) {
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

	log.Debug().Msgf("Starting new connection: %s", addr)
	conn, _, err := websocket.DefaultDialer.Dial(addr, header)
	if err != nil {
		msg := fmt.Sprintf("Target dial error: %v", err)
		log.Error().Msg(msg)
		p.writeOrigin([]byte(msg), true)
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
		p.writeOrigin([]byte(fmt.Sprintf("Writing target error: %v", err)), true)

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
		p.writeOrigin(b, false)
	}
}

func (p *proxy) writeOrigin(v []byte, err bool) {
	if v == nil {
		return
	}
	if err {
		v = append([]byte("Proxy: "), v...)
	}
	errx := p.origin.WriteMessage(websocket.TextMessage, v)
	if errx != nil {
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
		msg := fmt.Sprintf("Writing to file: %v", err)
		log.Error().Msgf(msg)
		p.writeOrigin([]byte(msg), true)
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
