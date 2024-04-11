package websocket

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/rejdeboer/multiplayer-server/internal/sync"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512

	// Note: Message types
	UPDATE     uint8 = 0
	QUERY_DIFF uint8 = 1
)

type Client struct {
	Context Context
	Doc     sync.Doc
	Hub     *Hub
	Conn    *websocket.Conn
	Send    chan []byte
}

func (c *Client) ReadPump() {
	log := c.Context.Log
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("unexpected close error")
			}
			break
		}

		messageType, message := message[0], message[1:]
		switch messageType {
		case UPDATE:
			go c.Doc.StoreUpdate(c.Context.Pool, message)
			c.Hub.Broadcast <- message
		case QUERY_DIFF:
			// TODO: Compute diff
			c.Send <- c.Doc.StateVector
		default:
			log.Error().Uint8("message_type", uint8(messageType)).Msg("unknown message type for reader")
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
