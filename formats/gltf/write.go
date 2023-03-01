package gltf

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/animation"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

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
	return key
}

func isVec4Atr(key string) bool {
	return key == modeling.JointAttribute || key == modeling.WeightAttribute
}

func ptrI(i int) *int {
	return &i
}

func flattenSkeletonToNodes(offset int, skeleton animation.Joint, out *bytes.Buffer) []Node {
	// pos := skeleton.RelativePosition()

	childrenIndexes := make([]int, len(skeleton.Children()))
	for i := range skeleton.Children() {
		childrenIndexes[i] = i + offset + 1
	}

	relativeMatrix := skeleton.RelativeMatrix()

	nodes := []Node{
		{
			// Translation: &[3]float64{pos.X(), pos.Y(), pos.Z()},
			Matrix: &[16]float64{
				relativeMatrix.X00,
				relativeMatrix.X10,
				relativeMatrix.X20,
				relativeMatrix.X30,

				relativeMatrix.X01,
				relativeMatrix.X11,
				relativeMatrix.X21,
				relativeMatrix.X31,

				relativeMatrix.X02,
				relativeMatrix.X12,
				relativeMatrix.X22,
				relativeMatrix.X32,

				relativeMatrix.X03,
				relativeMatrix.X13,
				relativeMatrix.X23,
				relativeMatrix.X33,
			},
			Children: childrenIndexes,
		},
	}

	mat := skeleton.InverseBindMatrix()
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

	currentOffset := offset + len(skeleton.Children())
	for _, c := range skeleton.Children() {
		subChildren := flattenSkeletonToNodes(currentOffset, c, out)
		currentOffset += len(subChildren)
		nodes = append(nodes, subChildren...)
	}
	return nodes
}

func writeAnimations(animations []animation.Sequence, gltf *Gltf, bin *bytes.Buffer) []Animation {
	gltfAnimations := make([]Animation, len(animations))
	for i, animation := range animations {

		min := vector3.New(math.MaxFloat64, math.MaxFloat64, math.MaxFloat64)
		max := vector3.New(-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64)

		for _, frame := range animation.Frames() {
			v := frame.Val()
			min = vector3.Min(min, v)
			max = vector3.Max(max, v)
			binary.Write(bin, binary.LittleEndian, float32(v.X()))
			binary.Write(bin, binary.LittleEndian, float32(v.Y()))
			binary.Write(bin, binary.LittleEndian, float32(v.Z()))
		}

		datasize := len(animation.Frames()) * 3 * 4

		animationDataBufferView := BufferView{
			Buffer:     0,
			ByteOffset: gltf.Buffers[0].ByteLength,
			ByteLength: datasize,
		}
		animationDataBufferViewIndex := len(gltf.BufferViews)

		animationDataAccessor := Accessor{
			BufferView:    ptrI(animationDataBufferViewIndex),
			ComponentType: AccessorComponentType_FLOAT,
			Type:          AccessorType_VEC3,
			Count:         len(animation.Frames()),
			Min:           []float64{min.X(), min.Y(), min.Z()},
			Max:           []float64{max.X(), max.Y(), max.Z()},
		}
		animationDataAccessorIndex := len(gltf.Accessors)

		gltf.Accessors = append(gltf.Accessors, animationDataAccessor)
		gltf.BufferViews = append(gltf.BufferViews, animationDataBufferView)

		gltf.Buffers[0].ByteLength += datasize

		// Time Data ============================================================

		minTime := math.MaxFloat64
		maxTime := -math.MaxFloat64

		for _, frame := range animation.Frames() {
			minTime = math.Min(minTime, frame.Time())
			maxTime = math.Max(maxTime, frame.Time())
			binary.Write(bin, binary.LittleEndian, float32(frame.Time()))
		}

		datasize = len(animation.Frames()) * 4

		timeBufferView := BufferView{
			Buffer:     0,
			ByteOffset: gltf.Buffers[0].ByteLength,
			ByteLength: datasize,
		}
		timeBufferViewIndex := len(gltf.BufferViews)

		timeAccessor := Accessor{
			BufferView:    ptrI(timeBufferViewIndex),
			ComponentType: AccessorComponentType_FLOAT,
			Type:          AccessorType_SCALAR,
			Count:         len(animation.Frames()),
			Min:           []float64{minTime},
			Max:           []float64{maxTime},
		}
		timeAccessorIndex := len(gltf.Accessors)

		gltf.Accessors = append(gltf.Accessors, timeAccessor)
		gltf.BufferViews = append(gltf.BufferViews, timeBufferView)

		gltf.Buffers[0].ByteLength += datasize

		gltfAnimations[i] = Animation{
			Samplers: []AnimationSampler{
				{
					Interpolation: AnimationSamplerInterpolation_LINEAR,
					Input:         timeAccessorIndex,
					Output:        animationDataAccessorIndex,
				},
			},
			Channels: []AnimationChannel{
				{
					Target: AnimationChannelTarget{
						Path: AnimationChannelTargetPath_TRANSLATION,
						Node: animation.Joint(),
					},
					Sampler: i,
				},
			},
		}
	}

	gltf.Animations = gltfAnimations

	return gltfAnimations
}

func structureFromMesh(mesh modeling.Mesh, skeleton *animation.Joint, animations []animation.Sequence) Gltf {
	primitiveAttributes := make(map[string]int)

	bufferViews := make([]BufferView, 0)
	accessors := make([]Accessor, 0)

	bin := &bytes.Buffer{}

	bytesWritten := 0
	attributesWritten := 0
	for _, val := range mesh.Float3Attributes() {

		accessorType := AccessorType_VEC3
		attributeType := attributeType(val)
		vec4 := isVec4Atr(val)

		min := vector3.New(math.MaxFloat64, math.MaxFloat64, math.MaxFloat64)
		max := vector3.New(-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64)

		mesh.ScanFloat3Attribute(val, func(i int, v vector3.Float64) {
			min = vector3.Min(min, v)
			max = vector3.Max(max, v)
			if attributeType == AccessorComponentType_FLOAT {
				binary.Write(bin, binary.LittleEndian, float32(v.X()))
				binary.Write(bin, binary.LittleEndian, float32(v.Y()))
				binary.Write(bin, binary.LittleEndian, float32(v.Z()))
				if vec4 {
					binary.Write(bin, binary.LittleEndian, float32(0))
				}
			} else if attributeType == AccessorComponentType_UNSIGNED_BYTE {
				binary.Write(bin, binary.LittleEndian, uint8(v.X()))
				binary.Write(bin, binary.LittleEndian, uint8(v.Y()))
				binary.Write(bin, binary.LittleEndian, uint8(v.Z()))
				if vec4 {
					binary.Write(bin, binary.LittleEndian, uint8(0))
				}
			}
		})

		primitiveAttributes[polyformToGLTFAttribute(val)] = attributesWritten
		minArr := []float64{min.X(), min.Y(), min.Z()}
		maxArr := []float64{max.X(), max.Y(), max.Z()}
		datasize := mesh.AttributeLength() * 3

		if vec4 {
			accessorType = AccessorType_VEC4
			minArr = append(minArr, 0)
			maxArr = append(maxArr, 0)
			datasize = mesh.AttributeLength() * 4
		}

		if attributeType == AccessorComponentType_FLOAT {
			datasize *= 4
		}

		accessors = append(accessors, Accessor{
			BufferView:    ptrI(len(bufferViews)),
			ComponentType: attributeType,
			Type:          accessorType,
			Count:         mesh.AttributeLength(),
			Min:           minArr,
			Max:           maxArr,
		})
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

	nodes := []Node{
		{
			Mesh: &meshIndex,
		},
	}

	scene := Scene{
		Nodes: []int{
			0,
		},
	}

	var skins []Skin = nil

	if skeleton != nil {
		nodes[0].Skin = ptrI(0)

		skeletonNodes := flattenSkeletonToNodes(1, *skeleton, bin)
		nodes = append(nodes, skeletonNodes...)

		jointIndices := make([]int, len(skeletonNodes))
		for i := 0; i < len(skeletonNodes); i++ {
			jointIndices[i] = i + 1 // +1 because we're offsetting from mesh node
		}

		accessors = append(accessors, Accessor{
			BufferView:    ptrI(len(bufferViews)),
			ComponentType: AccessorComponentType_FLOAT,
			Type:          AccessorType_MAT4,
			Count:         len(skeletonNodes),
			// Min:           []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			// Max:           []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		})

		inverseBindMAtrixLen := len(skeletonNodes) * 4 * 16

		bufferViews = append(bufferViews, BufferView{
			Buffer:     0,
			ByteOffset: bytesWritten,
			ByteLength: inverseBindMAtrixLen,
			// Target:     ARRAY_BUFFER,
		})
		bytesWritten += inverseBindMAtrixLen

		skins = []Skin{
			{
				Joints:              jointIndices,
				InverseBindMatrices: len(accessors) - 1,
			},
		}
		scene.Nodes = append(scene.Nodes, 1)
	}

	gltf := Gltf{
		Asset: defaultAsset(),
		Buffers: []Buffer{
			{
				ByteLength: bytesWritten,
			},
		},
		BufferViews: bufferViews,
		Accessors:   accessors,

		Skins: skins,
		Scene: 0,
		Scenes: []Scene{
			scene,
		},
		Nodes: nodes,
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

	if len(animations) > 0 {
		writeAnimations(animations, &gltf, bin)
	}

	gltf.Buffers[0].URI = "data:application/octet-stream;base64," + base64.StdEncoding.EncodeToString(bin.Bytes())

	return gltf
}

func WriteText(mesh modeling.Mesh, out io.Writer) error {
	return WriteTextWithAnimations(mesh, out, nil, nil)
}

func WriteTextWithAnimations(mesh modeling.Mesh, out io.Writer, skeleton *animation.Joint, animations []animation.Sequence) error {
	outline := structureFromMesh(mesh, skeleton, animations)
	bolB, err := json.MarshalIndent(outline, "", "    ")
	if err != nil {
		return err
	}

	_, err = out.Write(bolB)
	return err
}
