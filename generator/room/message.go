package room

import (
	"encoding/json"
)

type ClientSetOrientationMessage struct {
	Representation []PlayerRepresentation `json:"representation"`
}

type MessageType string

const (
	ClientSetOrientationMessageType MessageType = "Client-SetOrientation"
	ClientSetDisplayNameMessageType MessageType = "Client-SetDisplayName"
	ClientSetPointerMessageType     MessageType = "Client-SetPointer"
	ClientRemovePointerMessageType  MessageType = "Client-RemovePointer"
	ClientSetSceneMessageType       MessageType = "Client-SetScene"

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

func (m Message) ClientSetSceneData() (WebScene, error) {
	msg := WebScene{}
	err := json.Unmarshal(m.Data, &msg)
	return msg, err
}
