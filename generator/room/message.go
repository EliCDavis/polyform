package room

import (
	"encoding/json"

	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type ClientSetOrientationMessage struct {
	Position vector3.Float64 `json:"position"`
	Rotation vector4.Float64 `json:"rotation"`
}

type MessageType string

const (
	ClientSetOrientationMessageType MessageType = "Client-SetOrientation"
	ClientSetDisplayNameMessageType MessageType = "Client-SetDisplayName"
	ClientSetPointerMessageType     MessageType = "Client-SetPointer"
	ClientRemovePointerMessageType  MessageType = "Client-RemovePointer"

	ServerSetClientIDMessageType       MessageType = "Server-SetClientID"
	ServerRoomStateUpdateMessageType   MessageType = "Server-RoomStateUpdate"
	ServerRefrershGeneratorMessageType MessageType = "Server-RefreshGenerator"
	ServerBroadcastMessageType         MessageType = "Server-Broadcast"
)

type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (m Message) ClientSetDisplayNameData() string {
	return string(m.Data)
}

func (m Message) ClientSetOrientationData() (ClientSetOrientationMessage, error) {
	msg := ClientSetOrientationMessage{}
	err := json.Unmarshal(m.Data, &msg)
	return msg, err
}
