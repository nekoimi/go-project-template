package websocket

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer.
	defaultWriteWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	defaultPongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	defaultPingPeriod = 54 * time.Second

	// Maximum message size allowed from peer.
	defaultMaxMessageSize = 4096
)

type Client struct {
	manager  *Manager
	conn     *websocket.Conn
	send     chan []byte
	userID   string
	logger   *zap.Logger
}

func newClient(manager *Manager, conn *websocket.Conn, userID string, logger *zap.Logger) *Client {
	return &Client{
		manager: manager,
		conn:    conn,
		send:    make(chan []byte, 256),
		userID:  userID,
		logger:  logger,
	}
}

// ReadPump reads messages from the WebSocket connection and dispatches them.
func (c *Client) ReadPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(defaultMaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(defaultPongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(defaultPongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				c.logger.Warn("websocket read error", zap.String("userID", c.userID), zap.Error(err))
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			c.logger.Warn("invalid message format", zap.String("userID", c.userID), zap.Error(err))
			continue
		}

		msg.From = c.userID
		msg.Time = time.Now().UnixMilli()

		c.handleMessage(&msg)
	}
}

// WritePump writes messages to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(defaultPingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(defaultWriteWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.logger.Warn("websocket write error", zap.String("userID", c.userID), zap.Error(err))
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(defaultWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case MessageTypeChat:
		if msg.To != "" {
			c.manager.SendToUser(msg.To, msg)
		} else {
			c.manager.Broadcast(msg)
		}
	case MessageTypeNotify:
		if msg.To != "" {
			c.manager.SendToUser(msg.To, msg)
		}
	default:
		c.logger.Debug("unknown message type", zap.String("type", msg.Type))
	}
}

func (c *Client) sendJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		c.logger.Error("failed to marshal message", zap.Error(err))
		return
	}
	select {
	case c.send <- data:
	default:
		c.logger.Warn("send channel full, dropping message", zap.String("userID", c.userID))
	}
}
