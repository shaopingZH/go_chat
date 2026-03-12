package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8192
)

type HeartbeatHandler func(userID uint64)

type Client struct {
	UserID uint64

	conn        *websocket.Conn
	sendQueue   chan []byte
	done        chan struct{}
	closeOnce   sync.Once
	onHeartbeat HeartbeatHandler
}

func NewClient(userID uint64, conn *websocket.Conn, onHeartbeat HeartbeatHandler) *Client {
	return &Client{
		UserID:      userID,
		conn:        conn,
		sendQueue:   make(chan []byte, 128),
		done:        make(chan struct{}),
		onHeartbeat: onHeartbeat,
	}
}

func (c *Client) ReadPump(handle func(IncomingEnvelope)) {
	defer c.Close()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		if c.onHeartbeat != nil {
			c.onHeartbeat(c.UserID)
		}
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		if c.onHeartbeat != nil {
			c.onHeartbeat(c.UserID)
		}

		var incoming IncomingEnvelope
		if err := json.Unmarshal(message, &incoming); err != nil {
			c.SendError("invalid message format")
			continue
		}

		if handle != nil {
			handle(incoming)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer c.Close()

	for {
		select {
		case <-c.done:
			return
		case payload := <-c.sendQueue:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) Enqueue(payload []byte) bool {
	select {
	case <-c.done:
		return false
	default:
	}

	select {
	case <-c.done:
		return false
	case c.sendQueue <- payload:
		return true
	default:
		return false
	}
}

func (c *Client) SendJSON(v any) bool {
	payload, err := json.Marshal(v)
	if err != nil {
		return false
	}

	return c.Enqueue(payload)
}

func (c *Client) SendError(message string) bool {
	return c.SendJSON(ErrorEnvelope(message))
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.done)
		_ = c.conn.Close()
	})
}
