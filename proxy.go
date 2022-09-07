package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	jsp "github.com/buger/jsonparser"
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
	for range time.Tick(time.Second) {
		for I := range conns.Iter() {
			if !I.Value.connected(I.Value.origin) {
				if I.Value.target != nil {
					I.Value.target.Close()
				}
				conns.Delete(I.Key)
			}
		}
	}
}

func (p *proxy) connected(conn *websocket.Conn) bool {
	return conn != nil && conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)) == nil
}

func (p *proxy) originReader() {
	// Reading from the connection
	for {
		_, msg, err := p.origin.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Info().Msgf("Connection closed: %v", err)
				if p.target != nil {
					p.target.Close()
				}
				conns.Delete(p.origin.RemoteAddr().String())
				return
			}
		}
		proxyTarget, _ := jsp.GetString(msg, "ProxyTarget")
		proxyFileName, _ := jsp.GetString(msg, "ProxyFileName")
		msg = jsp.Delete(msg, "ProxyFileName")
		msg = jsp.Delete(msg, "ProxyTarget")

		if proxyTarget != "" && !p.connected(p.target) {
			if !p.connectTarget(proxyTarget) {
				continue
			}

		} else if !p.connected(p.target) {

			log.Error(p, fmt.Sprintf("Error during target address read , ProxyTarget = %v", proxyTarget))

		}
		if proxyFileName != "" {
			p.fileName = filepath.Join(config.ProxyLogFilePath, proxyFileName)
		}

		if config.LogLevel == 0 {
			m := map[string]interface{}{}
			json.Unmarshal(msg, &m)
			log.PrintJSON("Message sent to target:", m)
		}

		p.writeTarget(msg)
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
		if !p.connected(p.target) {
			return
		}
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

func (p *proxy) writeTarget(v []byte) {
	if v == nil || p.target == nil {
		return
	}
	log.Info().Msgf("Writing target, msg length: %d", len(v))
	err := p.target.WriteMessage(websocket.TextMessage, v)
	if err != nil {
		log.Error(p, fmt.Sprintf("Writing target error: %v", err))
	}

}

func (p *proxy) writeToFile(data []byte) {
	if len(data) == 0 {
		return
	}
	log.Debug().Msgf("logFile %s", p.fileName)
	f, err := os.OpenFile(p.fileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	// black magic
	// --------------------------------------------------
	file := map[string]interface{}{}
	m := map[string]interface{}{}
	b := make([]byte, 5<<20)
	var i int
	i, err = f.Read(b)
	if err != nil && err != io.EOF {
		log.Error(p, fmt.Sprintf("Reading file: %v", err))
	}
	f.Truncate(0)
	f.Seek(0, 0)

	log.Debug().Msgf("Read from file: %d", i)
	if i > 0 {
		if err = json.Unmarshal(b[:i], &file); err != nil {
			log.Logger.Error().Msgf("error: %v", err)
		}
	}

	json.Unmarshal(data, &m)
	file[time.UnixMicro(time.Now().UnixMicro()).Format("15:04:05.000")] = m
	data, _ = json.MarshalIndent(file, " ", "  ")
	// --------------------------------------------------
	log.Debug().Msgf("Write to file: %v", len(data))
	_, err = f.Write(data)
	if err != nil {
		log.Error(p, fmt.Sprintf("Writing to file: %v", err))
		return
	}

}
