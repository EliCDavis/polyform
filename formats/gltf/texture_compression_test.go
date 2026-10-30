package gltf_test

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func noisyImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	seed := uint32(7)
	for y := range 32 {
		for x := range 32 {
			seed = seed*1664525 + 1013904223
			img.Set(x, y, color.RGBA{uint8(seed >> 24), uint8(x * 8), uint8(y * 8), 255})
		}
	}
	return img
}

func encodedAt(t *testing.T, img image.Image, level png.CompressionLevel) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	require.NoError(t, (&png.Encoder{CompressionLevel: level}).Encode(buf, img))
	return buf.Bytes()
}

func texturedScene(material *gltf.PolyformMaterial) gltf.PolyformScene {
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(modeling.PositionAttribute, []vector3.Float64{
			vector3.New(0., 0., 0.),
			vector3.New(0., 1., 0.),
			vector3.New(1., 0., 0.),
		}).
		SetFloat2Attribute(modeling.TexCoordAttribute, nil)
	return gltf.PolyformScene{Models: []*gltf.PolyformModel{{Name: "tri", Mesh: &tri, Material: material}}}
}

func writtenImages(t *testing.T, scene gltf.PolyformScene, embed bool) [][]byte {
	t.Helper()
	out := &bytes.Buffer{}
	require.NoError(t, gltf.WriteBinary(scene, out, &gltf.WriterOptions{EmbedTextures: embed}))
	glb := out.Bytes()

	jsonLength := binary.LittleEndian.Uint32(glb[12:16])
	var doc struct {
		Images []struct {
			URI        string `json:"uri"`
			BufferView *int   `json:"bufferView"`
		} `json:"images"`
		BufferViews []struct {
			ByteOffset int `json:"byteOffset"`
			ByteLength int `json:"byteLength"`
		} `json:"bufferViews"`
	}
	require.NoError(t, json.Unmarshal(glb[20:20+jsonLength], &doc))
	bin := glb[20+jsonLength:]
	if len(bin) >= 8 {
		bin = bin[8:]
	}

	images := make([][]byte, len(doc.Images))
	for i, img := range doc.Images {
		if img.BufferView != nil {
			view := doc.BufferViews[*img.BufferView]
			images[i] = bin[view.ByteOffset : view.ByteOffset+view.ByteLength]
			continue
		}
		data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(img.URI, "data:image/png;base64,"))
		require.NoError(t, err)
		images[i] = data
	}
	return images
}

func TestTexturesAreEncodedAtTheirCompression(t *testing.T) {
	img := noisyImage()
	levels := map[string]struct {
		compression gltf.ImageCompression
		level       png.CompressionLevel
	}{
		"unset is fast": {level: png.BestSpeed},
		"fast":          {compression: gltf.ImageCompressionFast, level: png.BestSpeed},
		"balanced":      {compression: gltf.ImageCompressionBalanced, level: png.DefaultCompression},
		"smallest":      {compression: gltf.ImageCompressionSmallest, level: png.BestCompression},
		"none":          {compression: gltf.ImageCompressionNone, level: png.NoCompression},
	}

	for name, tc := range levels {
		for _, embed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s embedded=%v", name, embed), func(t *testing.T) {
				scene := texturedScene(&gltf.PolyformMaterial{
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorTexture: &gltf.PolyformTexture{Image: img, Compression: tc.compression},
					},
				})

				images := writtenImages(t, scene, embed)
				require.Len(t, images, 1)
				assert.Equal(t, encodedAt(t, img, tc.level), images[0])
			})
		}
	}
}

func TestOneImageAtTwoCompressionsIsWrittenTwice(t *testing.T) {
	img := noisyImage()
	scene := texturedScene(&gltf.PolyformMaterial{
		PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
			BaseColorTexture:         &gltf.PolyformTexture{Image: img},
			MetallicRoughnessTexture: &gltf.PolyformTexture{Image: img, Compression: gltf.ImageCompressionNone},
		},
	})

	images := writtenImages(t, scene, false)
	require.Len(t, images, 2)
	assert.ElementsMatch(t, [][]byte{encodedAt(t, img, png.BestSpeed), encodedAt(t, img, png.NoCompression)}, images)
}

func TestOneImageAtOneCompressionIsWrittenOnce(t *testing.T) {
	img := noisyImage()
	scene := texturedScene(&gltf.PolyformMaterial{
		PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
			BaseColorTexture:         &gltf.PolyformTexture{Image: img},
			MetallicRoughnessTexture: &gltf.PolyformTexture{Image: img},
		},
	})

	assert.Len(t, writtenImages(t, scene, false), 1)
}
