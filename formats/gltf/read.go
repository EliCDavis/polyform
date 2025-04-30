package gltf

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func decodeTopology(mode *PrimitiveMode) modeling.Topology {
	if mode == nil {
		return modeling.TriangleTopology
	}

	switch *mode {
	case PrimitiveMode_POINTS:
		return modeling.PointTopology

	case PrimitiveMode_LINES:
		return modeling.LineTopology

	case PrimitiveMode_LINE_LOOP:
		return modeling.LineLoopTopology

	case PrimitiveMode_LINE_STRIP:
		return modeling.LineStripTopology

	default:
		return modeling.TriangleTopology
	}
}

func decodeIndices(doc *Gltf, id *GltfId, buffers [][]byte) ([]int, error) {
	accessor := doc.Accessors[*id]
	if accessor.Type != AccessorType_SCALAR {
		return nil, fmt.Errorf("unexpected accessor type for indices: %s", accessor.Type)
	}

	bufferView := doc.BufferViews[*accessor.BufferView]
	start := bufferView.ByteOffset + accessor.ByteOffset
	end := bufferView.ByteOffset + bufferView.ByteLength
	buffer := buffers[bufferView.Buffer][start:end]

	indices := make([]int, accessor.Count)
	switch accessor.ComponentType {
	case AccessorComponentType_UNSIGNED_INT:
		for i := range indices {
			indices[i] = int(binary.LittleEndian.Uint32(buffer[i*4:]))
		}

	case AccessorComponentType_UNSIGNED_SHORT:
		for i := range indices {
			indices[i] = int(binary.LittleEndian.Uint16(buffer[i*2:]))
		}

	case AccessorComponentType_UNSIGNED_BYTE:
		for i := range indices {
			indices[i] = int(buffer[i])
		}

	default:
		return nil, fmt.Errorf("unexpected accessor component type for indices: %d", accessor.ComponentType)

	}

	return indices, nil
}

func decodePrimitiveAttributeName(name string) string {
	switch name {
	case POSITION:
		return modeling.PositionAttribute

	case COLOR_0:
		return modeling.ColorAttribute

	case JOINTS_0:
		return modeling.JointAttribute

	case WEIGHTS_0:
		return modeling.WeightAttribute

	case TEXCOORD_0:
		return modeling.TexCoordAttribute

	case NORMAL:
		return modeling.NormalAttribute

	default:
		return name
	}
}

func decodeVector2Accessor(doc *Gltf, id GltfId, buffers [][]byte) ([]vector2.Float64, error) {
	accessor := doc.Accessors[id]
	if accessor.Type != AccessorType_VEC2 {
		return nil, fmt.Errorf("unexpected accessor type for vec2: %s", accessor.Type)
	}

	bufferView := doc.BufferViews[*accessor.BufferView]
	start := bufferView.ByteOffset + accessor.ByteOffset
	end := bufferView.ByteOffset + bufferView.ByteLength
	buffer := buffers[bufferView.Buffer][start:end]

	stride := accessor.ComponentType.Size() * 2
	if bufferView.ByteStride != nil {
		stride = *bufferView.ByteStride
	}

	vectors := make([]vector2.Float64, accessor.Count)

	switch accessor.ComponentType {

	case AccessorComponentType_FLOAT:
		for i := range accessor.Count {
			offset := (i * stride)
			x := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset:])))
			y := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset+4:])))
			vectors[i] = vector2.New(x, y)
		}

	case AccessorComponentType_UNSIGNED_BYTE:
		div := 1.
		if accessor.Normalized {
			div = 255.
		}

		for i := range accessor.Count {
			offset := (i * stride)
			x := float64(buffer[offset]) / div
			y := float64(buffer[offset+1]) / div
			vectors[i] = vector2.New(x, y)
		}

	default:
		return nil, fmt.Errorf("unimplemented accessor component type: %d", accessor.ComponentType)
	}

	return vectors, nil
}

func decodeVector3Accessor(doc *Gltf, id GltfId, buffers [][]byte) ([]vector3.Float64, error) {
	accessor := doc.Accessors[id]
	if accessor.Type != AccessorType_VEC3 {
		return nil, fmt.Errorf("unexpected accessor type for vec3: %s", accessor.Type)
	}

	bufferView := doc.BufferViews[*accessor.BufferView]
	start := bufferView.ByteOffset + accessor.ByteOffset
	end := bufferView.ByteOffset + bufferView.ByteLength
	buffer := buffers[bufferView.Buffer][start:end]

	stride := accessor.ComponentType.Size() * 3
	if bufferView.ByteStride != nil {
		stride = *bufferView.ByteStride
	}

	vectors := make([]vector3.Float64, accessor.Count)

	switch accessor.ComponentType {

	case AccessorComponentType_FLOAT:
		for i := range accessor.Count {
			offset := (i * stride)
			x := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset:])))
			y := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset+4:])))
			z := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset+8:])))
			vectors[i] = vector3.New(x, y, z)
		}

	case AccessorComponentType_UNSIGNED_BYTE:
		div := 1.
		if accessor.Normalized {
			div = 255.
		}

		for i := range accessor.Count {
			offset := (i * stride)
			x := float64(buffer[offset]) / div
			y := float64(buffer[offset+1]) / div
			z := float64(buffer[offset+2]) / div
			vectors[i] = vector3.New(x, y, z)
		}

	default:
		return nil, fmt.Errorf("unimplemented accessor component type: %d", accessor.ComponentType)
	}

	return vectors, nil
}

func decodeVector4Accessor(doc *Gltf, id GltfId, buffers [][]byte) ([]vector4.Float64, error) {
	accessor := doc.Accessors[id]
	if accessor.Type != AccessorType_VEC4 {
		return nil, fmt.Errorf("unexpected accessor type for vec4: %s", accessor.Type)
	}

	bufferView := doc.BufferViews[*accessor.BufferView]
	start := bufferView.ByteOffset + accessor.ByteOffset
	bufferEnd := bufferView.ByteOffset + bufferView.ByteLength
	buffer := buffers[bufferView.Buffer][start:bufferEnd]

	stride := accessor.ComponentType.Size() * 4
	if bufferView.ByteStride != nil {
		stride = *bufferView.ByteStride
	}

	vectors := make([]vector4.Float64, accessor.Count)
	switch accessor.ComponentType {
	case AccessorComponentType_FLOAT:
		for i := range vectors {
			offset := (i * stride)
			x := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset:])))
			y := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset+4:])))
			z := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset+8:])))
			w := float64(math.Float32frombits(binary.LittleEndian.Uint32(buffer[offset+12:])))
			vectors[i] = vector4.New(x, y, z, w)
		}

	case AccessorComponentType_UNSIGNED_BYTE:
		div := 1.
		if accessor.Normalized {
			div = 255.
		}

		for i := range vectors {
			offset := (i * stride)
			x := float64(buffer[offset]) / div
			y := float64(buffer[offset+1]) / div
			z := float64(buffer[offset+2]) / div
			w := float64(buffer[offset+3]) / div
			vectors[i] = vector4.New(x, y, z, w)
		}

	default:
		return nil, fmt.Errorf("unimplemented accessor component type: %d", accessor.ComponentType)
	}

	return vectors, nil
}

func decodePrimitive(doc *Gltf, buffers [][]byte, n Node, m Mesh, p Primitive) (*PolyformModel, error) {
	indices, err := decodeIndices(doc, p.Indices, buffers)
	if err != nil {
		return nil, err
	}
	mesh := modeling.NewMesh(decodeTopology(p.Mode), indices)

	for attr, gltfId := range p.Attributes {
		acessor := doc.Accessors[gltfId]
		attributeName := decodePrimitiveAttributeName(attr)
		switch acessor.Type {
		case AccessorType_VEC2:
			v2, err := decodeVector2Accessor(doc, gltfId, buffers)
			if err != nil {
				return nil, err
			}
			mesh = mesh.SetFloat2Attribute(attributeName, v2)

		case AccessorType_VEC3:
			v3, err := decodeVector3Accessor(doc, gltfId, buffers)
			if err != nil {
				return nil, err
			}
			mesh = mesh.SetFloat3Attribute(attributeName, v3)

		case AccessorType_VEC4:
			v4, err := decodeVector4Accessor(doc, gltfId, buffers)
			if err != nil {
				return nil, err
			}
			mesh = mesh.SetFloat4Attribute(attributeName, v4)
		}
	}

	transform := trs.Identity()
	if n.Translation != nil {
		data := *n.Translation
		p := vector3.New(data[0], data[1], data[2])
		transform = transform.SetTranslation(p)
	}

	if n.Scale != nil {
		data := *n.Scale
		p := vector3.New(data[0], data[1], data[2])
		transform = transform.SetScale(p)
	}

	if n.Rotation != nil {
		data := *n.Rotation
		p := quaternion.New(vector3.New(data[0], data[1], data[2]), data[3])
		transform = transform.SetRotation(p)
	}

	return &PolyformModel{
		Name: n.Name,
		Mesh: &mesh,
		TRS:  &transform,
	}, nil
}

func ExperimentalDecodeModels(doc *Gltf, buffers [][]byte) ([]PolyformModel, error) {
	models := make([]PolyformModel, 0)

	for nodeIndex, node := range doc.Nodes {
		if node.Mesh == nil {
			continue
		}

		mesh := doc.Meshes[*node.Mesh]

		for primitiveIndex, p := range mesh.Primitives {
			model, err := decodePrimitive(doc, buffers, node, mesh, p)
			if err != nil {
				return nil, fmt.Errorf("Node %d Meshes[%d].primitives[%d]: %w", nodeIndex, *node.Mesh, primitiveIndex, err)
			}
			models = append(models, *model)
		}

	}

	return models, nil
}

func ExperimentalLoad(gltfPath string) (*Gltf, [][]byte, error) {
	gltfContents, err := os.ReadFile(gltfPath)
	if err != nil {
		return nil, nil, err
	}

	g := &Gltf{}
	err = json.Unmarshal(gltfContents, g)
	if err != nil {
		return nil, nil, err
	}

	allBuffers := make([][]byte, 0, len(g.Buffers))

	for bufIndex, buf := range g.Buffers {
		if strings.Index(buf.URI, "data:") == 0 {
			stringBuf := buf.URI[5:]

			base64Str := "application/octet-stream;base64,"
			if strings.Index(stringBuf, base64Str) == 0 {
				buf64, err := base64.StdEncoding.DecodeString(stringBuf[len(base64Str):])
				if err != nil {
					return g, allBuffers, err
				}

				allBuffers = append(allBuffers, buf64)
			} else {
				return g, allBuffers, fmt.Errorf("unimplemented buffer encoding on buffer %d", bufIndex)
			}
		} else {
			buf, err := os.ReadFile(filepath.Join(filepath.Dir(gltfPath), buf.URI))
			if err != nil {
				return g, allBuffers, err
			}
			allBuffers = append(allBuffers, buf)
		}
	}

	return g, allBuffers, nil
}
