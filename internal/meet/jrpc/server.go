package jrpc

import (
	"net/http"

	log "github.com/pion/ion-log"
	"github.com/randsoy/ct-sfu/internal/meet"
	"github.com/randsoy/ct-sfu/internal/meet/conf"

	"github.com/gorilla/websocket"
)

// Server a
type Server struct {
	meet   *meet.Meet
	closed chan bool
}

// New create jsonRPC
func New(c *conf.Config, m *meet.Meet) *Server {
	s := &Server{
		meet: m,
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	http.Handle("/ws", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get user id
		parms := r.URL.Query()
		fields := parms["uid"]
		if fields == nil || len(fields) == 0 {
			log.Errorf("invalid uid")
			http.Error(w, "invalid uid", http.StatusForbidden)
			return
		}
		uid := string(fields[0])
		log.Infof("peer connected, uid=%s", uid)

		// upgrade to websocket connection
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}
		defer ws.Close()

		// create a peer
		ch := NewChannel(r.Context(), ws, uid, m)
		defer ch.Close()
		log.Infof("new peer: %s, %s", r.RemoteAddr, ch.UID())

		// wait the peer disconnecting
		select {
		case <-ch.Conn().DisconnectNotify():
			log.Infof("peer disconnected: %s, %s", r.RemoteAddr, ch.UID())
			break
		case <-s.closed:
			log.Infof("server closed: disconnect peer, %s, %s", r.RemoteAddr, ch.UID())
			break
		}
	}))
	go func() {
		// start web server
		var err error
		if c.CertFile == "" || c.PrivateFile == "" {
			log.Infof("non-TLS WebSocketServer listening on: %s", c.Addr)
			err = http.ListenAndServe(c.Addr, nil)
		} else {
			log.Infof("TLS WebSocketServer listening on:  %s", c.Addr)
			err = http.ListenAndServeTLS(c.Addr, c.CertFile, c.PrivateFile, nil)
		}
		if err != nil {
			log.Errorf("http serve error: %v", err)
		}
	}()
	return s
}

// Close jsonrpc
func (s *Server) Close() {
	close(s.closed)
}
