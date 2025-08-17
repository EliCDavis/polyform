package room

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"

	"github.com/EliCDavis/bitlib"
	"github.com/EliCDavis/polyform/generator/schema"
)

type MessageType byte

const (
	ClientSetOrientationMessageType MessageType = 0
	ClientSetDisplayNameMessageType MessageType = 1
	ClientSetSceneMessageType       MessageType = 2
	ClientSetPointerMessageType     MessageType = 3
	ClientRemovePointerMessageType  MessageType = 4

	ServerSetClientIDMessageType       MessageType = 0 + 128
	ServerRoomStateUpdateMessageType   MessageType = 1 + 128
	ServerRefrershGeneratorMessageType MessageType = 2 + 128
	ServerBroadcastMessageType         MessageType = 3 + 128
)

// func (csomo ClientSetOrientationMessageObject) Position() vector3.Float32 {
// 	return vector3.New(csomo.PosX, csomo.PosY, csomo.PosZ)
// }

// func (csomo ClientSetOrientationMessageObject) Rotation() vector4.Float32 {
// 	return vector4.New(csomo.RotX, csomo.RotY, csomo.RotZ, csomo.RotW)
// }

type ClientSetOrientationMessage struct {
	Objects []PlayerRepresentation
}

type Message struct {
	Type MessageType
	Data []byte
}

func (m Message) Write(out io.Writer) error {
	_, err := out.Write([]byte{byte(m.Type)})
	if err != nil {
		return err
	}
	_, err = out.Write(m.Data)
	return err
}

func (m Message) ClientSetOrientation() (ClientSetOrientationMessage, error) {
	if len(m.Data) == 0 {
		return ClientSetOrientationMessage{}, nil
	}

	// 1 byte for type, 3 floats for position, 4 floats for rotation
	const oriantationMessageSize = 1 + (4 * 3) + (4 * 4)

	if len(m.Data)%oriantationMessageSize != 0 {
		return ClientSetOrientationMessage{}, errors.New("message has incomplete orientation data")
	}

	numOfObjects := len(m.Data) / oriantationMessageSize
	orientationData := make([]PlayerRepresentation, numOfObjects)

	err := binary.Read(bytes.NewReader(m.Data), binary.LittleEndian, &orientationData)

	return ClientSetOrientationMessage{
		Objects: orientationData,
	}, err
}

func (m Message) ClientSetSceneData() schema.WebScene {
	data := bytes.NewBuffer(m.Data)
	reader := bitlib.NewReader(data, binary.LittleEndian)

	webScene, _ := bitlib.Read[schema.WebScene](reader)
	return webScene
}

func (m Message) ClientSetDisplayName() string {
	return string(m.Data)
}

func (m Message) SeverSetClientID() string {
	return string(m.Data)
}

func (m Message) ServerRoomStateUpdate() RoomState {
	data := bytes.NewBuffer(m.Data)
	reader := bitlib.NewReader(data, binary.LittleEndian)

	room := RoomState{
		Players: make(map[string]*Player),
	}

	room.ModelVersion = reader.UInt32()
	webScene, err := bitlib.Read[schema.WebScene](reader)
	if err != nil {
		panic(err)
	}
	room.WebScene = &webScene

	playerLen := reader.Byte()

	for i := 0; i < int(playerLen); i++ {
		id := readString(reader)

		p := Player{}
		p.Name = readString(reader)

		representationLen := reader.Byte()
		p.Representation, _ = bitlib.ReadArray[PlayerRepresentation](reader, int(representationLen))

		room.Players[id] = &p
	}

	err = reader.Error()
	if err != nil {
		panic(err)
	}

	return room
}

func (m Message) ClientSetScene() json.RawMessage {
	return json.RawMessage(m.Data)
}

func MessageFromClient(data []byte) Message {
	if len(data) == 0 {
		panic(errors.New("can not construct client message from empty data"))
	}

	return Message{
		Type: MessageType(data[0]),
		Data: data[1:],
	}
}

func SeverSetClientIDMessage(id string) Message {
	return Message{
		Type: ServerSetClientIDMessageType,
		Data: []byte(id),
	}
}

func writeString(w *bitlib.Writer, s string) {
	w.Byte(byte(len(s)))
	w.WriteString(s)
}

func readString(r *bitlib.Reader) string {
	return r.String(int(r.Byte()))
}

func ServerRoomStateUpdate(room RoomState) Message {
	data := bytes.Buffer{}
	writer := bitlib.NewWriter(&data, binary.LittleEndian)

	writer.UInt32(uint32(room.ModelVersion))
	bitlib.Write(writer, room.WebScene)

	writer.Byte(byte(len(room.Players)))
	for id, player := range room.Players {
		writeString(writer, id)
		writeString(writer, player.Name)
		writer.Byte(byte(len(player.Representation)))
		bitlib.WriteArray(writer, player.Representation)
	}

	err := writer.Error()
	if err != nil {
		panic(err)
	}

	return Message{
		Type: ServerRoomStateUpdateMessageType,
		Data: data.Bytes(),
	}
}
