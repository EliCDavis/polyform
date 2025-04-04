// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package room

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/vector"
	"github.com/EliCDavis/vector/vector3"
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
	"Raven",
	"Chelsea",
	"Eddie",
}

type Vec4[T vector.Number] struct {
	X T
	Y T
	Z T
	W T
}

type PlayerRepresentation struct {
	Type     byte                          `json:"type"`
	Position vector3.Serializable[float32] `json:"position"`
	Rotation Vec4[float32]                 `json:"rotation"`
}

type Player struct {
	Name           string                 `json:"name"`
	Representation []PlayerRepresentation `json:"representation"`
}

type RoomState struct {
	ModelVersion uint32
	WebScene     *schema.WebScene
	Players      map[string]*Player
}

type clientUpdate struct {
	client *Client
	update Message
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	ClientConfig *ClientConfig

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

	graphInstance *graph.Instance
}

func NewHub(webScene *schema.WebScene, graphInstance *graph.Instance) *Hub {
	return &Hub{
		broadcast:     make(chan []byte),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		clients:       make(map[*Client]string),
		clientUpdates: make(chan clientUpdate),
		sceneUpdate:   make(chan time.Time),
		state: RoomState{
			Players:  map[string]*Player{},
			WebScene: webScene,
		},
		graphInstance: graphInstance,
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
			client.send <- SeverSetClientIDMessage(clientID)

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
				h.state.Players[clientID].Name = update.ClientSetDisplayName()

			case ClientSetOrientationMessageType:
				orientation, err := update.ClientSetOrientation()
				if err != nil {
					panic(fmt.Errorf("unable to set orientation data: %w", err))
				}

				h.state.Players[clientID].Representation = orientation.Objects
			case ClientSetSceneMessageType:
				scene := update.ClientSetSceneData()
				h.state.WebScene = &scene
			}

		case <-h.sceneUpdate:
			h.state.ModelVersion = h.graphInstance.ModelVersion()
			for client := range h.clients {
				select {
				case client.send <- ServerRoomStateUpdate(h.state):
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func (hub *Hub) ServeWs(w http.ResponseWriter, r *http.Request, clientConfig *ClientConfig) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan Message, 256),
		Config: clientConfig,
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
