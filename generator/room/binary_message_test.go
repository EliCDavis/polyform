package room_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestFromClientMessage(t *testing.T) {
	// ARANGE =================================================================
	data := []byte{0, 1, 2, 3, 4}

	// ACT ====================================================================
	message := room.MessageFromClient(data)

	// ASSERT =================================================================
	assert.Equal(t, room.ClientSetOrientationMessageType, message.Type)
	if assert.Len(t, message.Data, 4) {
		assert.Equal(t, data[1], message.Data[0])
		assert.Equal(t, data[2], message.Data[1])
		assert.Equal(t, data[3], message.Data[2])
		assert.Equal(t, data[4], message.Data[3])
	}
}

func TestFromClientMessage_PanicsOnEmptyBody(t *testing.T) {
	// ACT ====================================================================
	assert.PanicsWithError(t, "can not construct client message from empty data", func() {
		room.MessageFromClient(nil)
	})
}

func TestMessage_ClientSetOrientationMessage(t *testing.T) {
	// ARANGE =================================================================
	data := []byte{0, 1, 2, 3, 4}

	// ACT ====================================================================
	message := room.MessageFromClient(data)

	// ASSERT =================================================================
	assert.Equal(t, room.ClientSetOrientationMessageType, message.Type)
	if assert.Len(t, message.Data, 4) {
		assert.Equal(t, data[1], message.Data[0])
		assert.Equal(t, data[2], message.Data[1])
		assert.Equal(t, data[3], message.Data[2])
		assert.Equal(t, data[4], message.Data[3])
	}

	tests := map[string]struct {
		input []byte
		want  room.ClientSetOrientationMessage
		err   error
	}{
		// Errors
		"err: message has incomplete orientation data": {
			input: []byte{0, 1},
			want: room.ClientSetOrientationMessage{
				Objects: make([]room.PlayerRepresentation, 0),
			},
			err: errors.New("message has incomplete orientation data"),
		},

		// Valid messages
		"no objects": {
			input: []byte{0},
			want: room.ClientSetOrientationMessage{
				Objects: make([]room.PlayerRepresentation, 0),
			},
			err: nil,
		},

		"1 object": {
			input: []byte{
				0,

				// Object Type
				1,

				// Position Data
				0, 0, 0, 0, // X
				0, 0, 0, 0, // Y
				0, 0, 0, 0, // Z

				// Rotation Data
				0, 0, 0, 0, // X
				0, 0, 0, 0, // Y
				0, 0, 0, 0, // Z
				0, 0, 0, 0, // W

			},
			want: room.ClientSetOrientationMessage{
				Objects: []room.PlayerRepresentation{
					{
						Type: 1,
						Position: vector3.Serializable[float32]{
							X: 0,
							Y: 0,
							Z: 0,
						},
						Rotation: room.Vec4[float32]{
							X: 0,
							Y: 0,
							Z: 0,
							W: 0,
						},
					},
				},
			},
			err: nil,
		},

		"2 object": {
			input: []byte{
				0,

				// Object Type
				1,

				// Position Data
				0x00, 0x00, 0x80, 0x3f, // X
				0x00, 0x00, 0x00, 0x40, // Y
				0x00, 0x00, 0x40, 0x40, // Z

				// Rotation Data
				0x00, 0x00, 0x80, 0x3f, // X
				0x00, 0x00, 0x00, 0x40, // Y
				0x00, 0x00, 0x40, 0x40, // Z
				0x00, 0x00, 0x80, 0x40, // W

				// Object Type
				2,

				// Position Data
				0x00, 0x00, 0x40, 0x40, // X
				0x00, 0x00, 0x00, 0x40, // Y
				0x00, 0x00, 0x80, 0x3f, // Z

				// Rotation Data
				0, 0, 0, 0, // X
				0, 0, 0, 0, // Y
				0, 0, 0, 0, // Z
				0, 0, 0, 0, // W
			},
			want: room.ClientSetOrientationMessage{
				Objects: []room.PlayerRepresentation{
					{
						Type: 1,
						Position: vector3.Serializable[float32]{
							X: 1,
							Y: 2,
							Z: 3,
						},
						Rotation: room.Vec4[float32]{
							X: 1,
							Y: 2,
							Z: 3,
							W: 4,
						},
					},
					{
						Type: 2,
						Position: vector3.Serializable[float32]{
							X: 3,
							Y: 2,
							Z: 1,
						},
						Rotation: room.Vec4[float32]{
							X: 0,
							Y: 0,
							Z: 0,
							W: 0,
						},
					},
				},
			},
			err: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fromClient := room.MessageFromClient(tc.input)

			if !assert.Equal(t, room.ClientSetOrientationMessageType, fromClient.Type) {
				return
			}

			msg, err := fromClient.ClientSetOrientation()
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}

			if assert.Len(t, msg.Objects, len(tc.want.Objects)) {
				for i, obj := range msg.Objects {
					assert.Equal(t, tc.want.Objects[i], obj)
				}
			}
		})
	}
}

func TestMessage_MessageFromClientMessage(t *testing.T) {

	tests := map[string]struct {
		input []byte
		want  string
	}{
		"empty": {
			input: []byte{1},
			want:  "",
		},
		"a": {
			input: []byte{1, 0x61},
			want:  "a",
		},
		"ab": {
			input: []byte{1, 0x61, 0x62},
			want:  "ab",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fromClient := room.MessageFromClient(tc.input)

			if !assert.Equal(t, room.ClientSetDisplayNameMessageType, fromClient.Type) {
				return
			}

			msg := fromClient.ClientSetDisplayName()
			assert.Equal(t, tc.want, msg)
		})
	}
}

func TestMessage_ClientSetSceneMessage(t *testing.T) {
	tests := map[string]struct {
		input []byte
		want  json.RawMessage
	}{
		"empty": {
			input: []byte{2},
			want:  json.RawMessage{},
		},
		"1": {
			input: []byte{2, 0x31},
			want:  json.RawMessage("1"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fromClient := room.MessageFromClient(tc.input)

			if !assert.Equal(t, room.ClientSetSceneMessageType, fromClient.Type) {
				return
			}

			msg := fromClient.ClientSetScene()
			assert.Equal(t, tc.want, msg)
		})
	}
}

func TestMessage_SeverSetClientID(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []byte
	}{
		"123": {
			input: "123",
			want:  []byte{128, 0x31, 0x32, 0x33},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			msg := room.SeverSetClientIDMessage(tc.input)

			if !assert.Equal(t, room.ServerSetClientIDMessageType, msg.Type) {
				return
			}

			assert.Equal(t, tc.input, msg.SeverSetClientID())

			buf := &bytes.Buffer{}
			err := msg.Write(buf)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, buf.Bytes())
		})
	}
}

func TestMessage_ServerRoomStateUpdate(t *testing.T) {
	tests := map[string]struct {
		input room.RoomState
		want  []byte
	}{
		"123": {
			input: room.RoomState{
				ModelVersion: 1234,
				Players: map[string]*room.Player{
					"id1": {
						Name: "Something",
						Representation: []room.PlayerRepresentation{
							{
								Type: 2,
								Position: vector3.Serializable[float32]{
									X: 1,
									Y: 2,
									Z: 3,
								},
								Rotation: room.Vec4[float32]{
									X: 4,
									Y: 5,
									Z: 6,
									W: 7,
								},
							},
						},
					},
				},
				WebScene: &schema.WebScene{
					AntiAlias:       true,
					RenderWireframe: true,
					XrEnabled:       true,
					Fog: schema.WebSceneFog{
						Color: "#00FF00",
						Near:  12,
						Far:   25,
					},
					Background: "#000000",
					Lighting:   "#FFFFFF",
					Ground:     "#0000FF",
				},
			},
			want: []byte{
				0x81, 0xd2, 0x4, 0x0, 0x0, 0x1, 0x1, 0x1, 0x0, 0xff, 0x0, 0xff,
				0x0, 0x0, 0x40, 0x41, 0x0, 0x0, 0xc8, 0x41, 0x0, 0x0, 0x0,
				0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0xff, 0xff, 0x1, 0x3,
				0x69, 0x64, 0x31, 0x9, 0x53, 0x6f, 0x6d, 0x65, 0x74, 0x68, 0x69,
				0x6e, 0x67, 0x1, 0x2, 0x0, 0x0, 0x80, 0x3f, 0x0, 0x0, 0x0, 0x40,
				0x0, 0x0, 0x40, 0x40, 0x0, 0x0, 0x80, 0x40, 0x0, 0x0, 0xa0, 0x40,
				0x0, 0x0, 0xc0, 0x40, 0x0, 0x0, 0xe0, 0x40,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			msg := room.ServerRoomStateUpdate(tc.input)

			if !assert.Equal(t, room.ServerRoomStateUpdateMessageType, msg.Type) {
				return
			}

			assert.Equal(t, tc.input, msg.ServerRoomStateUpdate())

			buf := &bytes.Buffer{}
			err := msg.Write(buf)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, buf.Bytes())
		})
	}
}
