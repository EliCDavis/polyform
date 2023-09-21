// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package room

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var defaultDisplayNames = []string{
	"Peggy Hill",
	"Madara Uchiha",
	"Dr. Eggman",
	"Eustace Bagge",
	"Buttercup",
	"Professor Utonium",
	"Schnitzel",
	"Eduardo",
	"Mandy",
	"Captain K'nuckles",
	"Snufkin",
	"Mortimer",
	"Bella",
}

type Player struct {
	Name     string           `json:"name"`
	Position vector3.Float64  `json:"position"`
	Rotation vector4.Float64  `json:"rotation"`
	Pointer  *vector3.Float64 `json:"pointer,omitempty"`
}

type RoomState struct {
	Players map[string]*Player
}

type clientUpdate struct {
	client *Client
	update Message
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]string

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	clientUpdates chan clientUpdate

	sceneUpdate chan time.Time

	state RoomState
}

func NewHub() *Hub {
	return &Hub{
		broadcast:     make(chan []byte),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]string),
		clientUpdates: make(chan clientUpdate),
		sceneUpdate:   make(chan time.Time),
		state: RoomState{
			Players: map[string]*Player{},
		},
	}
}

func (h *Hub) Run() {

	go func() {
		for {
			time.Sleep(time.Millisecond * 200)
			h.sceneUpdate <- time.Now()
		}
	}()

	for {
		select {
		case client := <-h.register:
			clientID := randString(10)
			h.clients[client] = clientID
			h.state.Players[clientID] = &Player{
				Name: defaultDisplayNames[rand.Intn(len(defaultDisplayNames))],
			}
			client.send <- Message{
				Type: ServerSetClientIDMessageType,
				Data: []byte(fmt.Sprintf("\"%s\"", clientID)),
			}

		case client := <-h.unregister:
			if id, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.state.Players, id)
				close(client.send)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- Message{
					Type: ServerBroadcastMessageType,
					Data: message,
				}:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

		case clientUpdate := <-h.clientUpdates:
			clientID := h.clients[clientUpdate.client]
			update := clientUpdate.update
			switch update.Type {
			case ClientSetDisplayNameMessageType:
				h.state.Players[clientID].Name = update.ClientSetDisplayNameData()

			case ClientSetOrientationMessageType:
				orientation, err := update.ClientSetOrientationData()
				if err != nil {
					panic(fmt.Errorf("unable to deseriale set orientation data: %w", err))
				}

				h.state.Players[clientID].Position = orientation.Position
				h.state.Players[clientID].Rotation = orientation.Rotation
			}

		case <-h.sceneUpdate:
			message, err := json.Marshal(h.state)
			if err != nil {
				panic(err)
			}
			for client := range h.clients {
				select {
				case client.send <- Message{
					Type: ServerRoomStateUpdateMessageType,
					Data: message,
				}:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}

	}
}

// serveWs handles websocket requests from the peer.
func (hub *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan Message, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
