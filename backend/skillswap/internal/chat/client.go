package chat

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait is the time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// pongWait is the time allowed to read the next pong message from the peer.
	// Must be greater than pingPeriod.
	pongWait = 40 * time.Second

	// pingPeriod is how often the server sends pings. Must be < pongWait.
	pingPeriod = 30 * time.Second

	// maxMessageSize is the maximum incoming WebSocket message size (4 KB).
	maxMessageSize = 4096

	// sendBufferSize is the per-client outbound message buffer.
	sendBufferSize = 256

	// rateLimitMax is the maximum number of business messages (send_message)
	// a single client connection may send per rateLimitWindow.
	rateLimitMax = 30

	// rateLimitWindow is the sliding-window duration for the rate limiter.
	rateLimitWindow = time.Minute
)

// Client represents a single WebSocket connection for a user.
// A user may have multiple Client instances (one per browser tab).
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	UserID  string
	send    chan []byte
	handler func(client *Client, message []byte)

	// Rate limiting state
	msgCount    int
	windowStart time.Time
	rateMu      sync.Mutex
}

// newClient creates a Client bound to the given hub and connection.
func newClient(hub *Hub, conn *websocket.Conn, userID string, handler func(*Client, []byte)) *Client {
	return &Client{
		hub:         hub,
		conn:        conn,
		UserID:      userID,
		send:        make(chan []byte, sendBufferSize),
		handler:     handler,
		windowStart: time.Now(),
	}
}

// checkRateLimit returns true if the client is allowed to send another
// business message, false if they have exceeded the rate limit.
func (c *Client) checkRateLimit() bool {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()

	now := time.Now()
	if now.Sub(c.windowStart) >= rateLimitWindow {
		c.msgCount = 0
		c.windowStart = now
	}

	if c.msgCount >= rateLimitMax {
		return false
	}

	c.msgCount++
	return true
}

// readPump reads messages from the WebSocket connection and passes them
// to the handler callback. It runs in its own goroutine.
//
// On any read error (including a clean close), the client unregisters
// from the hub and the underlying connection is closed.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
			) {
				log.Printf("[WS] Unexpected close for user %s: %v", c.UserID, err)
			}
			return
		}

		if c.handler != nil {
			c.handler(c, message)
		}
	}
}

// writePump pumps messages from the send channel to the WebSocket
// connection. It also sends periodic pings. Runs in its own goroutine.
//
// When the hub closes the send channel (on unregister or slow-client
// eviction), writePump sends a close frame and exits.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if !ok {
				// Hub closed the channel — send a close frame.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
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
