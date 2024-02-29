package room

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type ClientConfig struct {
	// Time allowed to write a message to the peer.
	WriteWait time.Duration

	// Time allowed to read the next pong message from the peer.
	PongWait time.Duration

	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod time.Duration

	// Maximum message size allowed from peer.
	MaxMessageSize int64
}

const defaultPointWait = 60 * time.Second

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		WriteWait:      10 * time.Second,
		PongWait:       defaultPointWait,
		PingPeriod:     (defaultPointWait * 9) / 10,
		MaxMessageSize: 1024 * 10,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan Message

	// config
	Config *ClientConfig
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(c.Config.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(c.Config.PongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(c.Config.PongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		parsedMessage := Message{}
		parseErr := json.Unmarshal(message, &parsedMessage)
		if parseErr != nil {
			log.Printf("unable to parse message '%s': %s", string(message), parseErr.Error())
			continue
		}
		c.hub.clientUpdates <- clientUpdate{
			client: c,
			update: parsedMessage,
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(c.Config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("unable to serialize message: %v; %s", message, err.Error())
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(c.Config.WriteWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(data)

			// Add queued chat messages to the current websocket message.
			// n := len(c.send)
			// for i := 0; i < n; i++ {
			// 	w.Write(newline)
			// 	w.Write(<-c.send)
			// }

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.Config.WriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
