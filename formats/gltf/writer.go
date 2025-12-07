package gltf

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"

	"github.com/EliCDavis/bitlib"
	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/animation"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

const (
	magicNumber         uint32 = 0x46546C67
	jsonChunkIdentifier uint32 = 0x4E4F534A
	binChunkIdentifier  uint32 = 0x004E4942
	version             uint32 = 2
)

var ErrInvalidInput = errors.New("invalid input")

type Writer struct {
	buf          *bytes.Buffer
	bitW         *bitlib.Writer
	bytesWritten int

	accessors   []Accessor
	bufferViews []BufferView
	meshes      []Mesh
	nodes       []Node
	materials   []Material

	matIndices          materialIndices     // Tracks and deduplicates unique materials
	meshIndices         meshIndices         // Tracks and deduplicates unique meshes&materials
	writtenMeshData     attributeIndices    // Tracks and deduplicate written mesh data
	textureIndices      textureIndices      // Tracks and deduplicates unique textures
	embededImageIndices map[image.Image]int // Tracks and deduplicates unique written images to our buffer

	skins      []Skin
	animations []Animation

	textures []Texture
	images   []Image
	samplers []Sampler
	scenes   []Scene

	// Extension Stuff
	lights []KHR_LightsPunctual

	extensionsUsed     map[string]bool
	extensionsRequired map[string]bool

	Options WriterOptions
}

func NewWriter() *Writer {
	buf := &bytes.Buffer{}
	return &Writer{
		buf:         buf,
		bitW:        bitlib.NewWriter(buf, binary.LittleEndian),
		bufferViews: make([]BufferView, 0),
		accessors:   make([]Accessor, 0),
		nodes:       make([]Node, 0),
		meshes:      make([]Mesh, 0),
		materials:   make([]Material, 0),
		skins:       make([]Skin, 0),
		animations:  make([]Animation, 0),

		meshIndices:         make(meshIndices),
		writtenMeshData:     make(attributeIndices),
		textureIndices:      make(textureIndices),
		embededImageIndices: make(map[image.Image]int),

		// Extensions
		lights: make([]KHR_LightsPunctual, 0),

		extensionsUsed:     make(map[string]bool),
		extensionsRequired: make(map[string]bool),
	}
}

func NewWriterFromScene(scene PolyformScene) (*Writer, error) {
	writer := NewWriter()
	if err := writer.AddScene(scene); err != nil {
		return nil, fmt.Errorf("failed to add scene to writer: %w", err)
	}
	return writer, nil
}

// Align ensures the buffer is aligned to the specified byte boundary
func (w *Writer) Align(alignment int) {
	padding := (alignment - (w.bytesWritten % alignment)) % alignment
	if padding == 0 {
		return
	}
	for i := 0; i < int(padding); i++ {
		w.bitW.Byte(0)
	}
	w.bytesWritten += int(padding)
}

func (w Writer) WriteVector4AsFloat32(v vector4.Float64) {
	w.bitW.Float32(float32(v.X()))
	w.bitW.Float32(float32(v.Y()))
	w.bitW.Float32(float32(v.Z()))
	w.bitW.Float32(float32(v.W()))
}

func (w Writer) WriteVector4AsByte(v vector4.Float64) {
	w.bitW.Byte(uint8(v.X()))
	w.bitW.Byte(uint8(v.Y()))
	w.bitW.Byte(uint8(v.Z()))
	w.bitW.Byte(uint8(v.W()))
}

func (w Writer) WriteVector3AsFloat32(v vector3.Float64) {
	w.bitW.Float32(float32(v.X()))
	w.bitW.Float32(float32(v.Y()))
	w.bitW.Float32(float32(v.Z()))
}

func (w Writer) WriteVector3AsByte(v vector3.Float64) {
	w.bitW.Byte(uint8(v.X()))
	w.bitW.Byte(uint8(v.Y()))
	w.bitW.Byte(uint8(v.Z()))
}

func (w Writer) WriteVector2AsFloat32(v vector2.Float64) {
	w.bitW.Float32(float32(v.X()))
	w.bitW.Float32(float32(v.Y()))
}

func (w Writer) WriteVector2AsByte(v vector2.Float64) {
	w.bitW.Byte(uint8(v.X()))
	w.bitW.Byte(uint8(v.Y()))
}

func (w *Writer) WriteVector4(accessorComponentType AccessorComponentType, data *iter.ArrayIterator[vector4.Float64]) {
	w.Align(accessorComponentType.Size())

	accessorType := AccessorType_VEC4

	min := vector4.Fill(math.MaxFloat64)
	max := vector4.Fill(-math.MaxFloat64)

	if accessorComponentType == AccessorComponentType_FLOAT {
		for i := range data.Len() {
			v := data.At(i)
			min = vector4.Min(min, v)
			max = vector4.Max(max, v)
			w.WriteVector4AsFloat32(v)
		}
	}

	if accessorComponentType == AccessorComponentType_UNSIGNED_BYTE {
		for i := range data.Len() {
			v := data.At(i)
			min = vector4.Min(min, v)
			max = vector4.Max(max, v)
			w.WriteVector4AsByte(v)
		}
	}

	minArr := []float64{min.X(), min.Y(), min.Z(), min.W()}
	maxArr := []float64{max.X(), max.Y(), max.Z(), max.W()}
	datasize := data.Len() * 4 * accessorComponentType.Size()

	w.accessors = append(w.accessors, Accessor{
		BufferView:    ptrI(len(w.bufferViews)),
		ComponentType: accessorComponentType,
		Type:          accessorType,
		Count:         data.Len(),
		Min:           minArr,
		Max:           maxArr,
	})

	w.bufferViews = append(w.bufferViews, BufferView{
		Buffer:     0,
		ByteOffset: w.bytesWritten,
		ByteLength: datasize,
		Target:     ARRAY_BUFFER,
	})

	w.bytesWritten += datasize
}

func (w *Writer) WriteVector3(accessorComponentType AccessorComponentType, data *iter.ArrayIterator[vector3.Float64]) {
	w.Align(accessorComponentType.Size())

	accessorType := AccessorType_VEC3

	min := vector3.Fill(math.MaxFloat64)
	max := vector3.Fill(-math.MaxFloat64)

	if accessorComponentType == AccessorComponentType_FLOAT {
		for i := 0; i < data.Len(); i++ {
			v := data.At(i)
			w.WriteVector3AsFloat32(v)

			// Don't contaminate min/max with NaNs
			if v.ContainsNaN() {
				continue
			}
			min = vector3.Min(min, v)
			max = vector3.Max(max, v)
		}
	}

	if accessorComponentType == AccessorComponentType_UNSIGNED_BYTE {
		for i := 0; i < data.Len(); i++ {
			v := data.At(i)
			w.WriteVector3AsByte(v)

			// Don't contaminate min/max with NaNs
			if v.ContainsNaN() {
				continue
			}
			min = vector3.Min(min, v)
			max = vector3.Max(max, v)
		}
	}

	minArr := []float64{min.X(), min.Y(), min.Z()}
	maxArr := []float64{max.X(), max.Y(), max.Z()}
	datasize := data.Len() * 3 * accessorComponentType.Size()

	w.accessors = append(w.accessors, Accessor{
		BufferView:    ptrI(len(w.bufferViews)),
		ComponentType: accessorComponentType,
		Type:          accessorType,
		Count:         data.Len(),
		Min:           minArr,
		Max:           maxArr,
	})

	w.bufferViews = append(w.bufferViews, BufferView{
		Buffer:     0,
		ByteOffset: w.bytesWritten,
		ByteLength: datasize,
		Target:     ARRAY_BUFFER,
	})

	w.bytesWritten += datasize
}

func (w *Writer) WriteVector2(accessorComponentType AccessorComponentType, data *iter.ArrayIterator[vector2.Float64]) {
	w.Align(accessorComponentType.Size())

	accessorType := AccessorType_VEC2

	min := vector2.Fill(math.MaxFloat64)
	max := vector2.Fill(-math.MaxFloat64)

	if accessorComponentType == AccessorComponentType_FLOAT {
		for i := 0; i < data.Len(); i++ {
			v := data.At(i)
			w.WriteVector2AsFloat32(v)

			// Don't contaminate min/max with NaNs
			if v.ContainsNaN() {
				continue
			}
			min = vector2.Min(min, v)
			max = vector2.Max(max, v)
		}
	}

	if accessorComponentType == AccessorComponentType_UNSIGNED_BYTE {
		for i := 0; i < data.Len(); i++ {
			v := data.At(i)
			w.WriteVector2AsByte(v)

			// Don't contaminate min/max with NaNs
			if v.ContainsNaN() {
				continue
			}
			min = vector2.Min(min, v)
			max = vector2.Max(max, v)
		}
	}

	minArr := []float64{min.X(), min.Y()}
	maxArr := []float64{max.X(), max.Y()}
	datasize := data.Len() * 2 * accessorComponentType.Size()

	w.accessors = append(w.accessors, Accessor{
		BufferView:    ptrI(len(w.bufferViews)),
		ComponentType: accessorComponentType,
		Type:          accessorType,
		Count:         data.Len(),
		Min:           minArr,
		Max:           maxArr,
	})

	w.bufferViews = append(w.bufferViews, BufferView{
		Buffer:     0,
		ByteOffset: w.bytesWritten,
		ByteLength: datasize,
		Target:     ARRAY_BUFFER,
	})

	w.bytesWritten += datasize
}

func (w *Writer) WriteIndices(indices *iter.ArrayIterator[int], attributeSize int) {
	indiceSize := indices.Len()

	componentType := AccessorComponentType_UNSIGNED_INT
	if attributeSize > math.MaxUint16 {
		for i := range indices.Len() {
			w.bitW.UInt32(uint32(indices.At(i)))
		}
		indiceSize *= 4
	} else {
		for i := range indices.Len() {
			w.bitW.UInt16(uint16(indices.At(i)))
		}
		indiceSize *= 2
		componentType = AccessorComponentType_UNSIGNED_SHORT
	}

	w.Align(componentType.Size())

	w.accessors = append(w.accessors, Accessor{
		BufferView:    ptrI(len(w.bufferViews)),
		ComponentType: componentType,
		Type:          AccessorType_SCALAR,
		Count:         indices.Len(),
	})

	w.bufferViews = append(w.bufferViews, BufferView{
		Buffer:     0,
		ByteOffset: w.bytesWritten,
		ByteLength: indiceSize,
		Target:     ELEMENT_ARRAY_BUFFER,
	})

	w.bytesWritten += indiceSize
}

func (w *Writer) writeImageAsPng(image image.Image) (int, error) {
	buf := &bytes.Buffer{}
	err := png.Encode(buf, image)
	if err != nil {
		return -1, err
	}

	imageSize := buf.Len()
	_, err = w.bitW.Write(buf.Bytes())
	if err != nil {
		return -1, err
	}

	bufferViewIndex := len(w.bufferViews)

	w.bufferViews = append(w.bufferViews, BufferView{
		Buffer:     0,
		ByteOffset: w.bytesWritten,
		ByteLength: imageSize,
		Target:     ELEMENT_ARRAY_BUFFER,
	})

	w.bytesWritten += imageSize

	return bufferViewIndex, nil
}

func (w *Writer) imageToDataURI(img image.Image) (string, error) {
	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		return "", fmt.Errorf("failed to encode PNG image: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return "data:image/png;base64," + encoded, nil
}

func rgbaToFloatArr(c color.Color) [4]float64 {
	r, g, b, a := c.RGBA()
	return [4]float64{
		roundFloat(float64(r)/math.MaxUint16, 3),
		roundFloat(float64(g)/math.MaxUint16, 3),
		roundFloat(float64(b)/math.MaxUint16, 3),
		roundFloat(float64(a)/math.MaxUint16, 3),
	}
}

func rgbToFloatArr(c color.Color) [3]float64 {
	r, g, b, _ := c.RGBA()
	return [3]float64{
		roundFloat(float64(r)/math.MaxUint16, 3),
		roundFloat(float64(g)/math.MaxUint16, 3),
		roundFloat(float64(b)/math.MaxUint16, 3),
	}
}

func (w *Writer) AddScene(scene PolyformScene) error {

	sceneNodes := make([]GltfId, 0)

	childInstanceGroups := make(instancesCachce)

	for i, child := range scene.Models {
		switch w.Options.GpuInstancingStrategy {
		case WriterInstancingStrategy_Collapse:
			if canCollapseIntoInstance(child) {
				meshIndex, err := w.getOrAddMeshIndex(child)
				if err != nil {
					return err
				}
				childInstanceGroups.Add(meshIndex, child)
				continue
			}

		case WriterInstancingStrategy_Expand:
			if len(child.GpuInstances) > 0 {
				for _, instance := range child.GpuInstances {
					id, err := w.addModel(child, &instance)
					if err != nil {
						return fmt.Errorf("unable to add model[%d] %q to scene: %w", i, child.Name, err)
					}
					sceneNodes = append(sceneNodes, *id)
				}
			} else {
				id, err := w.addModel(child, nil)
				if err != nil {
					return fmt.Errorf("unable to add model[%d] %q to scene: %w", i, child.Name, err)
				}
				sceneNodes = append(sceneNodes, *id)
			}

		case WriterInstancingStrategy_Default:
			id, err := w.addModel(child, nil)
			if err != nil {
				return fmt.Errorf("unable to add model[%d] %q to scene: %w", i, child.Name, err)
			}
			sceneNodes = append(sceneNodes, *id)

		default:
			return fmt.Errorf("unimplemented writer instancing strategy: %d", WriterInstancingStrategy_Default)
		}
	}

	if w.Options.GpuInstancingStrategy == WriterInstancingStrategy_Collapse {

		// Write all children that weren't instanced
		for _, child := range scene.Models {

			if child.Mesh != nil {
				meshIndex, err := w.getOrAddMeshIndex(child)
				if err != nil {
					return err
				}
				if childInstanceGroups.IsInstanced(meshIndex) {
					continue
				}
			}

			childeNodeIndex, err := w.addModel(child, nil)
			if err != nil {
				return fmt.Errorf("error adding scene node %q: %w", child.Name, err)
			}
			sceneNodes = append(sceneNodes, *childeNodeIndex)
		}

		// Write all instance groups
		for mesh, instances := range childInstanceGroups {
			if len(instances) < 2 {
				continue
			}
			childIndex := len(w.nodes)
			childNode := Node{Name: "Instances", Mesh: &mesh.mesh}

			childRotations := make([]vector4.Float64, len(instances))
			for i, instance := range instances {
				childRotations[i] = instance.Rotation().Vector4()
			}

			childNode.Extensions = make(Extensions)
			childNode.Extensions[extGpuInstancingID] = w.addExtGpuInstancing(
				trs.Positions(instances),
				trs.Scales(instances),
				childRotations,
			)
			w.nodes = append(w.nodes, childNode)
			sceneNodes = append(sceneNodes, childIndex)
		}
	}

	// Add lights
	for _, light := range scene.Lights {
		sceneNodes = append(sceneNodes, w.AddLight(light))
	}

	if len(sceneNodes) == 0 {
		return nil
	}

	w.scenes = append(w.scenes, Scene{
		Nodes: sceneNodes,
	})

	return nil
}

func (w *Writer) getOrAddMeshIndex(model *PolyformModel) (meshIndex int, err error) {
	var matIndex *int
	if model.Material != nil {
		matIndex, err = w.AddMaterial(model.Material)
		if err != nil {
			return -1, fmt.Errorf("failed to add material %q from model %q: %w",
				model.Material.Name, model.Name, err)
		}
	}

	uniqueMesh := meshEntry{model.Mesh, -1}
	if matIndex != nil {
		uniqueMesh.materialIndex = *matIndex
	}

	// Check if mesh already exists
	if existingIndex, exists := w.meshIndices[uniqueMesh]; exists {
		return existingIndex, nil
	}

	// Create new mesh
	meshIndex = len(w.meshes)
	w.meshIndices[uniqueMesh] = meshIndex

	// Create the mesh - process geometry, materials etc

	var primitiveAttributes map[string]int
	var indicesIndex int

	writtenData, alreadyWrittenMesh := w.writtenMeshData[model.Mesh]

	if alreadyWrittenMesh {
		primitiveAttributes = writtenData.attribute
		indicesIndex = *writtenData.indices
	} else {
		primitiveAttributes = make(map[string]int)
		for _, val := range model.Mesh.Float4Attributes() {
			primitiveAttributes[polyformToGLTFAttribute(val)] = len(w.accessors)
			w.WriteVector4(attributeType(val), model.Mesh.Float4Attribute(val))
		}

		for _, val := range model.Mesh.Float3Attributes() {
			primitiveAttributes[polyformToGLTFAttribute(val)] = len(w.accessors)
			w.WriteVector3(attributeType(val), model.Mesh.Float3Attribute(val))
		}

		for _, val := range model.Mesh.Float2Attributes() {
			primitiveAttributes[polyformToGLTFAttribute(val)] = len(w.accessors)
			w.WriteVector2(attributeType(val), model.Mesh.Float2Attribute(val))
		}

		indicesIndex = len(w.accessors)
		w.WriteIndices(model.Mesh.Indices(), model.Mesh.AttributeLength())

		w.writtenMeshData[model.Mesh] = writtenMeshData{
			attribute: primitiveAttributes,
			indices:   &indicesIndex,
		}
	}

	var mode *PrimitiveMode = nil
	if model.Mesh.Topology() == modeling.PointTopology {
		p := PrimitiveMode_POINTS
		mode = &p
	}

	w.meshes = append(w.meshes, Mesh{
		ChildOfRootProperty: ChildOfRootProperty{Name: model.Name},
		Primitives: []Primitive{
			{
				Indices:    &indicesIndex,
				Attributes: primitiveAttributes,
				Material:   matIndex,
				Mode:       mode,
			},
		},
	})

	return meshIndex, nil
}

func canCollapseChildIntoParentAsInstance(parent, child *PolyformModel) bool {
	return len(child.Children) == 0 && child.Mesh == parent.Mesh && child.Material == parent.Material
}

func canCollapseIntoInstance(model *PolyformModel) bool {
	return len(model.Children) == 0 && model.Mesh != nil && model.Mesh.PrimitiveCount() > 0
}

func (w *Writer) addExtGpuInstancing(positions, scales []vector3.Float64, rotations []vector4.Float64) ExtGpuInstancing {
	w.extensionsRequired[extGpuInstancingID] = true
	w.extensionsUsed[extGpuInstancingID] = true
	instances := ExtGpuInstancing{
		Attributes: make(map[string]int),
	}

	instances.Attributes["TRANSLATION"] = len(w.accessors)
	w.WriteVector3(AccessorComponentType_FLOAT, iter.Array(positions))

	instances.Attributes["SCALE"] = len(w.accessors)
	w.WriteVector3(AccessorComponentType_FLOAT, iter.Array(scales))

	instances.Attributes["ROTATION"] = len(w.accessors)
	w.WriteVector4(AccessorComponentType_FLOAT, iter.Array(rotations))

	return instances
}

func (w *Writer) addModel(model *PolyformModel, parentTransformOverride *trs.TRS) (nodeId *GltfId, err error) {

	node := Node{
		Name: model.Name,
	}

	// if model.Mesh == nil && len(model.Children) == 0 {
	// 	return nil, fmt.Errorf("%w: nil mesh in model %q", ErrInvalidInput, model.Name)
	// }

	if model.Material != nil && model.Mesh == nil {
		return nil, errors.New("model can not reference a material without also referencing a mesh")
	}

	if model.TRS != nil || parentTransformOverride != nil {
		trs := trs.Identity()
		if model.TRS != nil {
			trs = *model.TRS
		}

		if parentTransformOverride != nil {
			trs = trs.Multiply(*parentTransformOverride)
		}

		if trs.Position() != vector3.Zero[float64]() {
			translation := trs.Position().ToFixedArr()
			node.Translation = &translation
		}

		if trs.Scale() != vector3.One[float64]() {
			scale := trs.Scale().ToFixedArr()
			node.Scale = &scale
		}

		if trs.Rotation() != quaternion.Identity() {
			rotation := trs.Rotation().ToArr()
			node.Rotation = &rotation
		}
	}

	if model.Mesh != nil && model.Mesh.PrimitiveCount() > 0 {
		meshIndex, err := w.getOrAddMeshIndex(model)
		if err != nil {
			return nil, err
		}
		node.Mesh = &meshIndex
	}

	positions := make([]vector3.Float64, 0, len(model.GpuInstances))
	rotations := make([]vector4.Float64, 0, len(model.GpuInstances))
	scales := make([]vector3.Float64, 0, len(model.GpuInstances))

	for _, instance := range model.GpuInstances {
		positions = append(positions, instance.Position())
		rotations = append(rotations, instance.Rotation().Vector4())
		scales = append(scales, instance.Scale())
	}

	node.Children = make([]GltfId, 0, len(model.Children))
	childInstanceGroups := make(instancesCachce)

	for i, child := range model.Children {
		if w.Options.GpuInstancingStrategy == WriterInstancingStrategy_Collapse {
			if canCollapseChildIntoParentAsInstance(model, child) {

				childTRS := trs.Identity()
				if child.TRS != nil {
					childTRS = *child.TRS
				}

				if len(child.GpuInstances) > 0 {
					for _, childInstance := range child.GpuInstances {
						adjustedTRS := childTRS.Multiply(childInstance)
						positions = append(positions, adjustedTRS.Position())
						rotations = append(rotations, adjustedTRS.Rotation().Vector4())
						scales = append(scales, adjustedTRS.Scale())
					}
				} else {
					positions = append(positions, childTRS.Position())
					rotations = append(rotations, childTRS.Rotation().Vector4())
					scales = append(scales, childTRS.Scale())
				}

				continue
			}

			if canCollapseIntoInstance(child) {
				meshIndex, err := w.getOrAddMeshIndex(child)
				if err != nil {
					return nil, err
				}
				childInstanceGroups.Add(meshIndex, child)
				continue
			}
		}

		if w.Options.GpuInstancingStrategy == WriterInstancingStrategy_Expand && len(child.GpuInstances) > 0 {
			for _, instance := range child.GpuInstances {
				childeNodeIndex, err := w.addModel(child, &instance)
				if err != nil {
					return nil, fmt.Errorf("error adding %q %d child %q: %w", model.Name, i, child.Name, err)
				}
				node.Children = append(node.Children, *childeNodeIndex)
			}
			continue
		}

		childeNodeIndex, err := w.addModel(child, nil)
		if err != nil {
			return nil, fmt.Errorf("error adding %q %d child %q: %w", model.Name, i, child.Name, err)
		}
		node.Children = append(node.Children, *childeNodeIndex)
	}

	if len(positions) > 0 {
		switch w.Options.GpuInstancingStrategy {
		case WriterInstancingStrategy_Default, WriterInstancingStrategy_Collapse:
			if node.Extensions == nil {
				node.Extensions = make(Extensions)
			}
			node.Extensions[extGpuInstancingID] = w.addExtGpuInstancing(positions, scales, rotations)
		}
	}

	if w.Options.GpuInstancingStrategy == WriterInstancingStrategy_Collapse {

		// Write all children that weren't instanced
		for i, child := range model.Children {
			if child.Mesh != nil {
				meshIndex, err := w.getOrAddMeshIndex(child)
				if err != nil {
					panic(err)
				}
				if childInstanceGroups.IsInstanced(meshIndex) {
					continue
				}
			}

			childeNodeIndex, err := w.addModel(child, nil)
			if err != nil {
				return nil, fmt.Errorf("error adding %q %d child %q: %w", model.Name, i, child.Name, err)
			}
			node.Children = append(node.Children, *childeNodeIndex)
		}

		// Write all instance groups
		for mesh, instances := range childInstanceGroups {
			if len(instances) < 2 {
				continue
			}
			childIndex := len(w.nodes)
			childNode := Node{Name: "Instances", Mesh: &mesh.mesh}

			childRotations := make([]vector4.Float64, len(instances))
			for i, instance := range instances {
				childRotations[i] = instance.Rotation().Vector4()
			}
			childNode.Extensions = make(Extensions)
			childNode.Extensions[extGpuInstancingID] = w.addExtGpuInstancing(
				trs.Positions(instances),
				trs.Scales(instances),
				childRotations,
			)
			w.nodes = append(w.nodes, childNode)
			node.Children = append(node.Children, childIndex)
		}
	}

	index := len(w.nodes)
	w.nodes = append(w.nodes, node)
	return &index, nil
}

func (w *Writer) AddTexture(polyTex *PolyformTexture) *TextureInfo {
	texIndex := -1
	var texFound bool
	for texPtr, index := range w.textureIndices {
		if texPtr == polyTex {
			texIndex = index
			texFound = true
			break
		}
	}

	texExt, texInfoExt := polyTex.prepareExtensions(w)
	newTexInfo := &TextureInfo{Extensions: texInfoExt, TexCoord: polyTex.TexCoord}
	if texFound { // This is exactly same texture object as was already added, reference it and return early
		newTexInfo.Index = texIndex
		return newTexInfo
	}

	// New texture may need to be created, but it still may be the same as existing one.
	newTex := Texture{Extensions: texExt}

	imageIndex := len(w.images)

	// If EmbedTextures is enabled and we have image data, prioritize embedding over URI
	if w.Options.EmbedTextures && polyTex.Image != nil {
		foundIndex, ok := w.embededImageIndices[polyTex.Image]
		if ok {
			imageIndex = foundIndex
		} else {
			// Encode image as data URI instead of writing to buffer
			dataURI, err := w.imageToDataURI(polyTex.Image)
			if err != nil {
				panic(err)
			}

			w.images = append(w.images, Image{URI: dataURI})
			w.embededImageIndices[polyTex.Image] = imageIndex
		}
	} else if polyTex.URI != "" { // check if an image with this URI was already added before
		var imageFound bool
		for i, im := range w.images {
			if im.URI == polyTex.URI {
				imageIndex = i
				imageFound = true
				break
			}
		}
		if !imageFound {
			w.images = append(w.images, Image{URI: polyTex.URI})
		}
	} else if polyTex.Image != nil {

		foundIndex, ok := w.embededImageIndices[polyTex.Image]
		if ok {
			imageIndex = foundIndex
		} else {
			bufferView, err := w.writeImageAsPng(polyTex.Image)
			if err != nil {
				panic(err)
			}

			w.images = append(w.images, Image{
				MimeType:   ImageMimeType_PNG,
				BufferView: &bufferView,
			})
			w.embededImageIndices[polyTex.Image] = imageIndex
		}
	} else {
		panic(fmt.Errorf("no uri or image"))
	}
	newTex.Source = ptrI(imageIndex)

	// Check if a sampler like existing was already aded
	if polyTex.Sampler != nil {
		samplerIndex := len(w.samplers)
		var samplerFound bool
		for i, sampler := range w.samplers {
			if polyTex.Sampler.equal(&sampler) {
				samplerIndex = i
				samplerFound = true
				break
			}
		}
		if !samplerFound {
			w.samplers = append(w.samplers, *polyTex.Sampler)
		}
		newTex.Sampler = ptrI(samplerIndex)
	}

	// Check if the newly built texture is exactly the same as existing, if so - reuse existing.
	texIndex = len(w.textures)
	for i, tex := range w.textures {
		if newTex.equal(tex) {
			texIndex = i
			texFound = true
			break
		}
	}

	newTexInfo.Index = texIndex
	if !texFound {
		w.textureIndices[polyTex] = texIndex
		w.textures = append(w.textures, newTex)
	}
	return newTexInfo
}

func (w *Writer) AddMaterial(mat *PolyformMaterial) (*int, error) {
	// Check if material already exists
	if existingId, ok := w.matIndices.findExistingMaterialID(mat); ok {
		return existingId, nil
	}
	var pbr = &PbrMetallicRoughness{
		BaseColorFactor: &[4]float64{1, 1, 1, 1},
	}

	extensions := make(map[string]any)

	if mat.PbrMetallicRoughness != nil {
		polyPBR := mat.PbrMetallicRoughness

		pbr.MetallicFactor = polyPBR.MetallicFactor
		pbr.RoughnessFactor = polyPBR.RoughnessFactor

		if polyPBR.BaseColorFactor != nil {
			factor := rgbaToFloatArr(polyPBR.BaseColorFactor)
			pbr.BaseColorFactor = &factor
		}

		if polyPBR.BaseColorTexture.canAddToGLTF() {
			pbr.BaseColorTexture = w.AddTexture(polyPBR.BaseColorTexture)
		}

		if polyPBR.MetallicRoughnessTexture.canAddToGLTF() {
			pbr.MetallicRoughnessTexture = w.AddTexture(polyPBR.MetallicRoughnessTexture)
		}
	}

	for _, ext := range mat.Extensions {
		id := ext.ExtensionID()
		extensions[id] = ext.ToMaterialExtensionData(w)
		w.extensionsUsed[id] = true
	}

	var emissiveFactor *[3]float64
	if mat.EmissiveFactor != nil {
		factor := rgbToFloatArr(mat.EmissiveFactor)
		emissiveFactor = &factor
	}

	if mat.AlphaCutoff != nil && (mat.AlphaMode == nil || *mat.AlphaMode != MaterialAlphaMode_MASK) {
		alphaModeStr := "nil"
		if mat.AlphaMode != nil {
			alphaModeStr = string(*mat.AlphaMode)
		}

		return nil, fmt.Errorf("%w: invalid material %q: "+
			"alphaCutOff can only be set when the alphaMode == MASK: got %q", ErrInvalidInput, mat.Name, alphaModeStr)
	}

	m := Material{
		ChildOfRootProperty: ChildOfRootProperty{
			Name: mat.Name,
			Property: Property{
				Extras:     mat.Extras,
				Extensions: extensions,
			},
		},
		AlphaMode:            mat.AlphaMode,
		AlphaCutoff:          mat.AlphaCutoff,
		PbrMetallicRoughness: pbr,
		EmissiveFactor:       emissiveFactor,
	}

	if mat.NormalTexture != nil {
		m.NormalTexture = &NormalTexture{
			TextureInfo: *w.AddTexture(mat.NormalTexture.PolyformTexture),
			Scale:       mat.NormalTexture.Scale,
		}
	}
	if mat.OcclusionTexture != nil {
		m.OcclusionTexture = &OcclusionTexture{
			TextureInfo: *w.AddTexture(mat.OcclusionTexture.PolyformTexture),
			Strength:    mat.OcclusionTexture.Strength,
		}
	}
	if mat.EmissiveTexture != nil {
		m.EmissiveTexture = w.AddTexture(mat.EmissiveTexture)
	}

	w.materials = append(w.materials, m)
	index := len(w.materials) - 1

	// Add to material tracker
	w.matIndices = append(w.matIndices, materialEntry{
		polyMaterial: mat,
		index:        index,
	})

	return &index, nil
}

func (w *Writer) AddSkin(skeleton animation.Skeleton) Node {

	skeletonNodes := flattenSkeletonToNodes(1, skeleton, w.buf)
	jointIndices := make([]int, len(skeletonNodes))
	for i := range skeletonNodes {
		jointIndices[i] = len(w.nodes) + i
	}

	w.nodes = append(w.nodes, skeletonNodes...)

	w.accessors = append(w.accessors, Accessor{
		BufferView:    ptrI(len(w.bufferViews)),
		ComponentType: AccessorComponentType_FLOAT,
		Type:          AccessorType_MAT4,
		Count:         len(skeletonNodes),
		// Min:           []float64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		// Max:           []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	})

	inverseBindMAtrixLen := len(skeletonNodes) * 4 * 16

	w.bufferViews = append(w.bufferViews, BufferView{
		Buffer:     0,
		ByteOffset: w.bytesWritten,
		ByteLength: inverseBindMAtrixLen,
		// Target:     ARRAY_BUFFER,
	})
	w.bytesWritten += inverseBindMAtrixLen

	w.skins = append(w.skins, Skin{
		Joints:              jointIndices,
		InverseBindMatrices: len(w.accessors) - 1,
	})

	return Node{
		Skin: ptrI(len(w.skins) - 1),
	}
}

func (w *Writer) AddAnimations(animations []animation.Sequence, skeleton animation.Skeleton, skeletonNode int) {
	for i, animation := range animations {

		min := vector3.New(math.MaxFloat64, math.MaxFloat64, math.MaxFloat64)
		max := vector3.New(-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64)

		for _, frame := range animation.Frames() {
			v := frame.Val()
			min = vector3.Min(min, v)
			max = vector3.Max(max, v)
			w.WriteVector3AsFloat32(v)
		}

		datasize := len(animation.Frames()) * 3 * 4

		animationDataBufferView := BufferView{
			Buffer:     0,
			ByteOffset: w.bytesWritten,
			ByteLength: datasize,
		}
		animationDataBufferViewIndex := len(w.bufferViews)

		animationDataAccessor := Accessor{
			BufferView:    ptrI(animationDataBufferViewIndex),
			ComponentType: AccessorComponentType_FLOAT,
			Type:          AccessorType_VEC3,
			Count:         len(animation.Frames()),
			Min:           []float64{min.X(), min.Y(), min.Z()},
			Max:           []float64{max.X(), max.Y(), max.Z()},
		}
		animationDataAccessorIndex := len(w.accessors)

		w.accessors = append(w.accessors, animationDataAccessor)
		w.bufferViews = append(w.bufferViews, animationDataBufferView)

		w.bytesWritten += datasize

		// Time Data ============================================================

		minTime := math.MaxFloat64
		maxTime := -math.MaxFloat64

		for _, frame := range animation.Frames() {
			minTime = math.Min(minTime, frame.Time())
			maxTime = math.Max(maxTime, frame.Time())
			w.bitW.Float32(float32(frame.Time()))
		}

		datasize = len(animation.Frames()) * 4

		timeBufferView := BufferView{
			Buffer:     0,
			ByteOffset: w.bytesWritten,
			ByteLength: datasize,
		}
		timeBufferViewIndex := len(w.bufferViews)

		timeAccessor := Accessor{
			BufferView:    ptrI(timeBufferViewIndex),
			ComponentType: AccessorComponentType_FLOAT,
			Type:          AccessorType_SCALAR,
			Count:         len(animation.Frames()),
			Min:           []float64{minTime},
			Max:           []float64{maxTime},
		}

		timeAccessorIndex := len(w.accessors)
		w.accessors = append(w.accessors, timeAccessor)
		w.bufferViews = append(w.bufferViews, timeBufferView)

		w.bytesWritten += datasize

		w.animations = append(w.animations, Animation{
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
						Node: skeleton.Lookup(animation.Joint()) + skeletonNode,
					},
					Sampler: i,
				},
			},
		})
	}
}

func (w *Writer) AddLight(light KHR_LightsPunctual) GltfId {
	nodeIndex := len(w.nodes)

	lightIndex := len(w.lights)
	w.lights = append(w.lights, light)

	var translation = [3]float64{
		light.Position.X(),
		light.Position.Y(),
		light.Position.Z(),
	}

	w.nodes = append(w.nodes, Node{
		ChildOfRootProperty: ChildOfRootProperty{
			Property: Property{
				Extensions: map[string]any{
					"KHR_lights_punctual": map[string]any{
						"light": lightIndex,
					},
				},
			},
		},
		Translation: &translation,
	})
	w.extensionsUsed["KHR_lights_punctual"] = true
	return nodeIndex
}

type BufferEmbeddingStrategy int

const (
	BufferEmbeddingStrategy_Base64Encode BufferEmbeddingStrategy = iota
	BufferEmbeddingStrategy_GLB
)

func (w Writer) ToGLTF(embeddingStrategy BufferEmbeddingStrategy) Gltf {
	buffers := []Buffer{}
	if w.bytesWritten > 0 {
		buffer := Buffer{
			ByteLength: w.bytesWritten,
		}

		if embeddingStrategy == BufferEmbeddingStrategy_Base64Encode {
			buffer.URI = "data:application/octet-stream;base64," + base64.StdEncoding.EncodeToString(w.buf.Bytes())
		}

		buffers = append(buffers, buffer)
	}

	extensionsUsedArr := make([]string, 0, len(w.extensionsUsed))
	for ext := range w.extensionsUsed {
		extensionsUsedArr = append(extensionsUsedArr, ext)
	}

	extensionsRequiredArr := make([]string, 0, len(w.extensionsRequired))
	for ext := range w.extensionsRequired {
		extensionsRequiredArr = append(extensionsRequiredArr, ext)
	}

	extensions := make(map[string]any)
	if len(w.lights) > 0 {
		arr := make([]map[string]any, 0)

		for _, l := range w.lights {
			arr = append(arr, l.ToExtension())
		}

		extensions["KHR_lights_punctual"] = map[string]any{
			"lights": arr,
		}
	}

	var scene *int
	if len(w.scenes) > 0 {
		zero := 0
		scene = &zero
	}

	return Gltf{
		Asset:       defaultAsset(),
		Buffers:     buffers,
		BufferViews: w.bufferViews,
		Accessors:   w.accessors,

		// Skins: skins,
		Scene:  scene,
		Scenes: w.scenes,

		Skins:      w.skins,
		Animations: w.animations,

		Nodes:     w.nodes,
		Meshes:    w.meshes,
		Materials: w.materials,
		Textures:  w.textures,
		Images:    w.images,
		Samplers:  w.samplers,

		ExtensionsUsed:     extensionsUsedArr,
		ExtensionsRequired: extensionsRequiredArr,
		Property: Property{
			Extensions: extensions,
		},
	}
}

func (w Writer) WriteGLB(out io.Writer, opts WriterOptions) error {
	var jsonBytes []byte
	var err error
	switch opts.JsonFormat {
	case PrettyJsonFormat:
		jsonBytes, err = json.MarshalIndent(w.ToGLTF(BufferEmbeddingStrategy_GLB), "", "    ")
	case DefaultJsonFormat, MinifyJsonFormat:
		fallthrough
	default:
		jsonBytes, err = json.Marshal(w.ToGLTF(BufferEmbeddingStrategy_GLB))
	}
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	jsonByteLen := len(jsonBytes)
	jsonPadding := (4 - (jsonByteLen % 4)) % 4
	jsonByteLen += jsonPadding

	binBytes := w.buf.Bytes()
	binByteLen := len(binBytes)
	binPadding := (4 - (binByteLen % 4)) % 4
	binByteLen += binPadding

	bitWriter := bitlib.NewWriter(out, binary.LittleEndian)

	// https://registry.khronos.org/glTF/specs/2.0/glTF-2.0.pdf
	// magic MUST be equal to equal 0x46546C67. It is ASCII string glTF and can
	// be used to identify data as Binary glTF
	bitWriter.UInt32(magicNumber)

	// GLB version
	bitWriter.UInt32(version)

	// Length of entire document
	totalLen := jsonByteLen + binByteLen + 12 + 8
	if binByteLen > 0 {
		totalLen += 8
	}
	bitWriter.UInt32(uint32(totalLen))

	// JSON CHUNK =============================================================

	// Chunk Length
	bitWriter.UInt32(uint32(jsonByteLen))

	// JSON Chunk Identifier
	bitWriter.UInt32(jsonChunkIdentifier)

	// JSON data
	bitWriter.ByteArray(jsonBytes)

	// Padding to make it align to a 4 byte boundary
	for i := 0; i < jsonPadding; i++ {
		bitWriter.Byte(0x20)
	}

	// OPTIONAL BIN CHUNK =====================================================

	// Don't write anything if the bin data is empty
	if binByteLen == 0 {
		return bitWriter.Error()
	}

	// Chunk Length
	bitWriter.UInt32(uint32(binByteLen))

	// BIN Chunk Identifier
	bitWriter.UInt32(binChunkIdentifier)

	// Bin data
	bitWriter.ByteArray(binBytes)

	// Padding to make it align to a 4 byte boundary
	for i := 0; i < binPadding; i++ {
		bitWriter.Byte(0x00)
	}

	return bitWriter.Error()
}

// Microvalue, anything below this is assumed to be floating point precision noise and will be discarded.
const epsilon = 1e-8

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	result := math.Round(val*ratio) / ratio
	if math.Abs(result) < epsilon {
		result = 0.0 // remove "-0.0" results
	}

	return result
}
