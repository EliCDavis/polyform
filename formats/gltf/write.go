package gltf

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

func defaultAsset() Asset {
	return Asset{
		Version:   "2.0",
		Generator: "https://github.com/EliCDavis/polyform",
	}
}

func polyformToGLTFAttribute(key string) string {
	switch key {
	case modeling.PositionAttribute:
		return POSITION

	case modeling.ColorAttribute:
		return COLOR_0

	case modeling.TexCoordAttribute:
		return TEXCOORD_0

	case modeling.NormalAttribute:
		return NORMAL
	}
	return key
}

func ptrI(i int) *int {
	return &i
}

func structureFromMesh(mesh modeling.Mesh) Gltf {
	primitiveAttributes := make(map[string]int)

	bufferViews := make([]BufferView, 0)
	accessors := make([]Accessor, 0)

	bin := &bytes.Buffer{}

	bytesWritten := 0
	attributesWritten := 0
	for _, val := range mesh.Float3Attributes() {

		min := vector3.New(math.MaxFloat64, math.MaxFloat64, math.MaxFloat64)
		max := vector3.New(-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64)

		mesh.ScanFloat3Attribute(val, func(i int, v vector3.Float64) {
			min = vector3.Min(min, v)
			max = vector3.Max(max, v)
			binary.Write(bin, binary.LittleEndian, float32(v.X()))
			binary.Write(bin, binary.LittleEndian, float32(v.Y()))
			binary.Write(bin, binary.LittleEndian, float32(v.Z()))
		})

		primitiveAttributes[polyformToGLTFAttribute(val)] = attributesWritten
		accessors = append(accessors, Accessor{
			BufferView:    ptrI(len(bufferViews)),
			ComponentType: AccessorComponentType_FLOAT,
			Type:          AccessorType_VEC3,
			Count:         mesh.AttributeLength(),
			Min:           []float64{min.X(), min.Y(), min.Z()},
			Max:           []float64{max.X(), max.Y(), max.Z()},
		})

		datasize := mesh.AttributeLength() * 3 * 4

		bufferViews = append(bufferViews, BufferView{
			Buffer:     0,
			ByteOffset: bytesWritten,
			ByteLength: datasize,
			Target:     ARRAY_BUFFER,
		})

		bytesWritten += datasize
		attributesWritten++
	}

	for _, val := range mesh.Float2Attributes() {

		min := vector2.New(math.MaxFloat64, math.MaxFloat64)
		max := vector2.New(-math.MaxFloat64, -math.MaxFloat64)

		mesh.ScanFloat2Attribute(val, func(i int, v vector2.Float64) {
			min = vector2.Min(min, v)
			max = vector2.Max(max, v)
			binary.Write(bin, binary.LittleEndian, float32(v.X()))
			binary.Write(bin, binary.LittleEndian, float32(v.Y()))
		})

		primitiveAttributes[polyformToGLTFAttribute(val)] = attributesWritten
		accessors = append(accessors, Accessor{
			BufferView:    ptrI(len(bufferViews)),
			ComponentType: AccessorComponentType_FLOAT,
			Type:          AccessorType_VEC2,
			Count:         mesh.AttributeLength(),
			Min:           []float64{min.X(), min.Y()},
			Max:           []float64{max.X(), max.Y()},
		})

		datasize := mesh.AttributeLength() * 2 * 4

		bufferViews = append(bufferViews, BufferView{
			Buffer:     0,
			ByteOffset: bytesWritten,
			ByteLength: datasize,
			Target:     ARRAY_BUFFER,
		})

		bytesWritten += datasize
		attributesWritten++
	}

	indiceSize := mesh.PrimitiveCount() * 3 * 4

	for i := 0; i < mesh.PrimitiveCount(); i++ {
		tri := mesh.Tri(i)
		binary.Write(bin, binary.LittleEndian, uint32(tri.P1()))
		binary.Write(bin, binary.LittleEndian, uint32(tri.P2()))
		binary.Write(bin, binary.LittleEndian, uint32(tri.P3()))
	}

	indiceIndex := len(accessors)

	accessors = append(accessors, Accessor{
		BufferView:    ptrI(len(bufferViews)),
		ComponentType: AccessorComponentType_UNSIGNED_INT,
		Type:          AccessorType_SCALAR,
		Count:         mesh.PrimitiveCount() * 3,
	})

	bufferViews = append(bufferViews, BufferView{
		Buffer:     0,
		ByteOffset: bytesWritten,
		ByteLength: indiceSize,
		Target:     ELEMENT_ARRAY_BUFFER,
	})

	bytesWritten += indiceSize

	meshIndex := 0

	return Gltf{
		Asset: defaultAsset(),
		Buffers: []Buffer{
			{
				URI:        "data:application/octet-stream;base64," + base64.StdEncoding.EncodeToString(bin.Bytes()),
				ByteLength: bytesWritten,
			},
		},
		BufferViews: bufferViews,
		Accessors:   accessors,

		Scene: 0,
		Scenes: []Scene{
			{
				Nodes: []int{
					0,
				},
			},
		},
		Nodes: []Node{
			{
				Mesh: &meshIndex,
			},
		},
		Meshes: []Mesh{
			{
				ChildOfRootProperty: ChildOfRootProperty{
					Name: "mesh",
				},
				Primitives: []Primitive{
					{
						Indices:    &indiceIndex,
						Attributes: primitiveAttributes,
						Material:   ptrI(0),
					},
				},
			},
		},
		Materials: []Material{
			{
				PbrMetallicRoughness: &PbrMetallicRoughness{
					BaseColorFactor: &[4]float64{1, 1, 1, 1},
				},
			},
		},
	}
}

func WriteText(mesh modeling.Mesh, out io.Writer) error {
	outline := structureFromMesh(mesh)
	bolB, err := json.MarshalIndent(outline, "", "    ")
	if err != nil {
		return err
	}

	_, err = out.Write(bolB)
	return err
}
