package jrpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/pion/ion-log"
	"github.com/randsoy/ct-sfu/internal/meet"
	"github.com/sourcegraph/jsonrpc2"
	websocketjsonrpc2 "github.com/sourcegraph/jsonrpc2/websocket"
)

// Channel 通道
type Channel struct {
	uid  string
	mid  string
	conn *jsonrpc2.Conn
	ctx  context.Context

	meet *meet.Meet
}

// NewChannel create a peer instance
func NewChannel(ctx context.Context, c *websocket.Conn, uid string, m *meet.Meet) *Channel {
	ch := &Channel{
		uid:  uid,
		ctx:  ctx,
		meet: m,
		mid:  uuid.New().String(),
	}
	ch.conn = jsonrpc2.NewConn(ctx, websocketjsonrpc2.NewObjectStream(c), ch)
	return ch
}

// Handle incoming RPC call events, implement jsonrpc2.Handler
func (ch *Channel) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {

	replyError := func(err error) {
		_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    500,
			Message: fmt.Sprintf("%s", err),
		})
	}

	switch req.Method {
	case "join":
		var join meet.Join
		err := json.Unmarshal(*req.Params, &join)
		if err != nil {
			log.Errorf("connect: error parsing Join: %v", err)
			replyError(err)
			break
		}
		answer, err := ch.meet.Join(ch, join)
		if err != nil {
			replyError(err)
			break
		}
		_ = conn.Reply(ctx, req.ID, answer)
	case "offer":
		var negotiation meet.Negotiation
		err := json.Unmarshal(*req.Params, &negotiation)
		if err != nil {
			log.Errorf("connect: error parsing Negotiation: %v", err)
			replyError(err)
			break
		}
		answer, err := ch.meet.Offer(ch, negotiation)
		if err != nil {
			replyError(err)
			break
		}
		_ = conn.Reply(ctx, req.ID, answer)
	case "answer":
		var negotiation meet.Negotiation
		err := json.Unmarshal(*req.Params, &negotiation)
		if err != nil {
			log.Errorf("connect: error parsing Negotiation: %v", err)
			replyError(err)
			break
		}
		_, err = ch.meet.Answer(ch, negotiation)
		if err != nil {
			replyError(err)
			break
		}
	case "trickle": // 其他ice相关消息
		var trickle meet.Trickle
		err := json.Unmarshal(*req.Params, &trickle)
		if err != nil {
			log.Errorf("connect: error parsing Trickle: %v", err)
			replyError(err)
			break
		}
		_, err = ch.meet.Trickle(ch, trickle)
		if err != nil {
			replyError(err)
			break
		}
	case "leave":
		var leave meet.Leave
		err := json.Unmarshal(*req.Params, &leave)
		if err != nil {
			log.Errorf("connect: error parsing Leave: %v", err)
			replyError(err)
			break
		}
		_, err = ch.meet.Leave(ch, leave)
		if err != nil {
			replyError(err)
			break
		}
	}
}

// Close peer
func (ch *Channel) Close() {
}

// Notify peer
func (ch *Channel) Notify(method string, params interface{}) (err error) {
	if ch.conn == nil {
		log.Errorf("mid %s ch.conn == nil", ch.mid)
		return nil
	}
	if err = ch.conn.Notify(ch.ctx, method, params); err != nil {
		log.Errorf("error sending offer %s", err)
	}
	return err
}

// UID return peer user id
func (ch *Channel) UID() string {
	return ch.uid
}

// MID return peer media id
func (ch *Channel) MID() string {
	return ch.mid
}

// Conn return peer uid
func (ch *Channel) Conn() *jsonrpc2.Conn {
	return ch.conn
}
