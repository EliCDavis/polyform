package gltf

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/animation"
)

// JsonFormat specifies the JSON output format for GLTF export
type JsonFormat int

const (
	// DefaultJsonFormat uses the library's default JSON formatting (pretty-printed with indentation)
	DefaultJsonFormat JsonFormat = iota
	// MinifyJsonFormat produces compact JSON output without indentation
	MinifyJsonFormat
	// PrettyJsonFormat explicitly specifies pretty-printed JSON with indentation (same as default)
	PrettyJsonFormat
)

// WriterOptions configures GLTF export behavior
type WriterOptions struct {
	// GpuInstancingStrategy Determines how to utilize GLTF's
	// EXT_mesh_gpu_instancing extension, if at all. The model mesh and material
	// deduplication applies regardless.
	GpuInstancingStrategy WriterInstancingStrategy

	// EmbedTextures forces texture images to be embedded as data URIs instead of external file references
	EmbedTextures bool

	// JsonFormat specifies the JSON output format (default is pretty-printed with indentation)
	JsonFormat JsonFormat
}

func defaultAsset() Asset {
	return Asset{
		Version:   "2.0",
		Generator: "https://github.com/EliCDavis/polyform",
	}
}

func attributeType(key string) AccessorComponentType {
	switch key {
	case modeling.JointAttribute:
		return AccessorComponentType_UNSIGNED_BYTE

	default:
		return AccessorComponentType_FLOAT
	}
}

func polyformToGLTFAttribute(key string) string {
	switch key {
	case modeling.PositionAttribute:
		return POSITION

	case modeling.ColorAttribute:
		return COLOR_0

	case modeling.JointAttribute:
		return JOINTS_0

	case modeling.WeightAttribute:
		return WEIGHTS_0

	case modeling.TexCoordAttribute:
		return TEXCOORD_0

	case modeling.NormalAttribute:
		return NORMAL
	}

	if strings.HasPrefix(key, modeling.TexCoordAttribute) {
		i, err := strconv.Atoi(key[len(modeling.TexCoordAttribute):])
		if err != nil {
			return fmt.Sprintf("TEXCOORD_%d", i-1)
		}
	}

	return key
}

func isVec4Atr(key string) bool {
	return key == modeling.JointAttribute || key == modeling.WeightAttribute
}

func ptrI(i int) *int {
	return &i
}

func ptrIEqual(i, j *int) bool {
	return (i == nil && j == nil) || (i != nil && j != nil && *i == *j)
}

func flattenSkeletonToNodes(offset int, skeleton animation.Skeleton, out *bytes.Buffer) []Node {
	nodes := make([]Node, 0)

	for i := 0; i < skeleton.JointCount(); i++ {
		children := skeleton.Children(i)
		for i, c := range children {
			children[i] = c + offset
		}

		// relativeMatrix := skeleton.RelativeMatrix(i)

		pos := skeleton.RelativePosition(i)

		node := Node{
			Translation: &[3]float64{pos.X(), pos.Y(), pos.Z()},
			// Matrix: &[16]float64{
			// 	relativeMatrix.X00,
			// 	relativeMatrix.X10,
			// 	relativeMatrix.X20,
			// 	relativeMatrix.X30,

			// 	relativeMatrix.X01,
			// 	relativeMatrix.X11,
			// 	relativeMatrix.X21,
			// 	relativeMatrix.X31,

			// 	relativeMatrix.X02,
			// 	relativeMatrix.X12,
			// 	relativeMatrix.X22,
			// 	relativeMatrix.X32,

			// 	relativeMatrix.X03,
			// 	relativeMatrix.X13,
			// 	relativeMatrix.X23,
			// 	relativeMatrix.X33,
			// },
			Children: children,
		}

		// mat := skeleton.InverseBindMatrix(i)
		mat := mat.Identity()
		worldPos := skeleton.WorldPosition(i)
		mat.X03 = -worldPos.X()
		mat.X13 = -worldPos.Y()
		mat.X23 = -worldPos.Z()
		// binary.Write(out, binary.LittleEndian, float32(mat.X00))
		// binary.Write(out, binary.LittleEndian, float32(mat.X01))
		// binary.Write(out, binary.LittleEndian, float32(mat.X02))
		// binary.Write(out, binary.LittleEndian, float32(mat.X03))

		// binary.Write(out, binary.LittleEndian, float32(mat.X10))
		// binary.Write(out, binary.LittleEndian, float32(mat.X11))
		// binary.Write(out, binary.LittleEndian, float32(mat.X12))
		// binary.Write(out, binary.LittleEndian, float32(mat.X13))

		// binary.Write(out, binary.LittleEndian, float32(mat.X20))
		// binary.Write(out, binary.LittleEndian, float32(mat.X21))
		// binary.Write(out, binary.LittleEndian, float32(mat.X22))
		// binary.Write(out, binary.LittleEndian, float32(mat.X23))

		// binary.Write(out, binary.LittleEndian, float32(mat.X30))
		// binary.Write(out, binary.LittleEndian, float32(mat.X31))
		// binary.Write(out, binary.LittleEndian, float32(mat.X32))
		// binary.Write(out, binary.LittleEndian, float32(mat.X33))

		binary.Write(out, binary.LittleEndian, float32(mat.X00))
		binary.Write(out, binary.LittleEndian, float32(mat.X10))
		binary.Write(out, binary.LittleEndian, float32(mat.X20))
		binary.Write(out, binary.LittleEndian, float32(mat.X30))

		binary.Write(out, binary.LittleEndian, float32(mat.X01))
		binary.Write(out, binary.LittleEndian, float32(mat.X11))
		binary.Write(out, binary.LittleEndian, float32(mat.X21))
		binary.Write(out, binary.LittleEndian, float32(mat.X31))

		binary.Write(out, binary.LittleEndian, float32(mat.X02))
		binary.Write(out, binary.LittleEndian, float32(mat.X12))
		binary.Write(out, binary.LittleEndian, float32(mat.X22))
		binary.Write(out, binary.LittleEndian, float32(mat.X32))

		binary.Write(out, binary.LittleEndian, float32(mat.X03))
		binary.Write(out, binary.LittleEndian, float32(mat.X13))
		binary.Write(out, binary.LittleEndian, float32(mat.X23))
		binary.Write(out, binary.LittleEndian, float32(mat.X33))

		nodes = append(nodes, node)
	}

	return nodes
}

func WriteText(scene PolyformScene, out io.Writer, options *WriterOptions) error {
	var opts WriterOptions
	if options != nil {
		opts = *options
	}

	writer := NewWriter()
	writer.Options = opts

	if err := writer.AddScene(scene); err != nil {
		return fmt.Errorf("failed to add scene to writer: %w", err)
	}

	outline := writer.ToGLTF(BufferEmbeddingStrategy_Base64Encode)

	var bolB []byte
	var err error
	switch opts.JsonFormat {
	case MinifyJsonFormat:
		bolB, err = json.Marshal(outline)
	case DefaultJsonFormat, PrettyJsonFormat:
		fallthrough
	default:
		bolB, err = json.MarshalIndent(outline, "", "    ")
	}
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if _, err = out.Write(bolB); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}
	return nil
}

func WriteBinary(scene PolyformScene, out io.Writer, options *WriterOptions) error {
	var opts WriterOptions
	if options != nil {
		opts = *options
	}

	writer := NewWriter()
	writer.Options = opts

	if err := writer.AddScene(scene); err != nil {
		return fmt.Errorf("failed to add scene to writer: %w", err)
	}

	if err := writer.WriteGLB(out, opts); err != nil {
		return fmt.Errorf("failed to write GLB: %w", err)
	}

	return nil
}
