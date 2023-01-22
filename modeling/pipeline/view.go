package pipeline

import (
	"errors"
	"fmt"
	"sync"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

type View struct {
	mesh             *modeling.Mesh
	readPermissions  Permission
	writePermissions Permission
	meshMutex        sync.Mutex
}

// VertexNeighborTable
func (v View) VertexNeighborTable() modeling.VertexLUT {
	if !v.readPermissions.HasPrimitivePermission() {
		panic(errors.New("mesh view does not have access to primitive"))
	}
	return v.mesh.VertexNeighborTable()
}

func (v View) AttributeLength() int {
	if !v.readPermissions.HasAttributePermission() {
		panic(errors.New("mesh view does not have access to attribute information"))
	}
	return v.AttributeLength()
}

func (v View) requireV3ReadPermission(attribute string) {
	if !v.readPermissions.HasFloat3Permission(attribute) {
		panic(fmt.Errorf("mesh view does not have read permission to float3 mesh attribute: '%s'", attribute))
	}
}

func (v View) requireV2ReadPermission(attribute string) {
	if !v.readPermissions.HasFloat2Permission(attribute) {
		panic(fmt.Errorf("mesh view does not have read permission to float2 mesh attribute: '%s'", attribute))
	}
}

func (v View) requireV1ReadPermission(attribute string) {
	if !v.readPermissions.HasFloat1Permission(attribute) {
		panic(fmt.Errorf("mesh view does not have read permission to float1 mesh attribute: '%s'", attribute))
	}
}

func (v View) requireV3WritePermission(attribute string) {
	if !v.writePermissions.HasFloat3Permission(attribute) {
		panic(fmt.Errorf("mesh view does not have write permission to float3 mesh attribute: '%s'", attribute))
	}
}

func (v View) requireV2WritePermission(attribute string) {
	if !v.writePermissions.HasFloat2Permission(attribute) {
		panic(fmt.Errorf("mesh view does not have write permission to float2 mesh attribute: '%s'", attribute))
	}
}

func (v View) requireV1WritePermission(attribute string) {
	if !v.writePermissions.HasFloat1Permission(attribute) {
		panic(fmt.Errorf("mesh view does not have write permission to float1 mesh attribute: '%s'", attribute))
	}
}

func (v View) ScanFloat3Attribute(attribute string, f func(i int, v vector.Vector3)) {
	v.requireV3ReadPermission(attribute)
	v.mesh.ScanFloat3Attribute(attribute, f)
}

func (v View) ScanFloat3AttributeParallel(attribute string, f func(i int, v vector.Vector3)) {
	v.requireV3ReadPermission(attribute)
	v.mesh.ScanFloat3AttributeParallel(attribute, f)
}

func (v View) ScanFloat3AttributeParallelWithPoolSize(attribute string, size int, f func(i int, v vector.Vector3)) {
	v.requireV3ReadPermission(attribute)
	v.mesh.ScanFloat3AttributeParallelWithPoolSize(attribute, size, f)
}

func (v View) ScanFloat2Attribute(attribute string, f func(i int, v vector.Vector2)) {
	v.requireV2ReadPermission(attribute)
	v.mesh.ScanFloat2Attribute(attribute, f)
}

func (v View) ScanFloat2AttributeParallel(attribute string, f func(i int, v vector.Vector2)) {
	v.requireV2ReadPermission(attribute)
	v.mesh.ScanFloat2AttributeParallel(attribute, f)
}

func (v View) ScanFloat2AttributeParallelWithPoolSize(attribute string, size int, f func(i int, v vector.Vector2)) {
	v.requireV2ReadPermission(attribute)
	v.mesh.ScanFloat2AttributeParallelWithPoolSize(attribute, size, f)
}

func (v View) ScanFloat1Attribute(attribute string, f func(i int, v float64)) {
	v.requireV1ReadPermission(attribute)
	v.mesh.ScanFloat1Attribute(attribute, f)
}

func (v View) ScanFloat1AttributeParallel(attribute string, f func(i int, v float64)) {
	v.requireV1ReadPermission(attribute)
	v.mesh.ScanFloat1AttributeParallel(attribute, f)
}

func (v View) ScanFloat1AttributeParallelWithPoolSize(attribute string, size int, f func(i int, v float64)) {
	v.requireV1ReadPermission(attribute)
	v.mesh.ScanFloat1AttributeParallelWithPoolSize(attribute, size, f)
}

func (v *View) SetFloat3Attribute(attribute string, data []vector.Vector3) {
	v.requireV3WritePermission(attribute)

	mesh := v.mesh.SetFloat3Attribute(attribute, data)
	v.meshMutex.Lock()
	v.mesh = &mesh
	v.meshMutex.Unlock()
}

func (v *View) SetFloat2Attribute(attribute string, data []vector.Vector2) {
	v.requireV2WritePermission(attribute)

	mesh := v.mesh.SetFloat2Attribute(attribute, data)
	v.meshMutex.Lock()
	v.mesh = &mesh
	v.meshMutex.Unlock()
}

func (v *View) SetFloat1Attribute(attribute string, data []float64) {
	v.requireV1WritePermission(attribute)

	mesh := v.mesh.SetFloat1Attribute(attribute, data)
	v.meshMutex.Lock()
	v.mesh = &mesh
	v.meshMutex.Unlock()
}

func newView(mesh modeling.Mesh, read, write Permission) *View {
	return &View{
		mesh:             &mesh,
		readPermissions:  read,
		writePermissions: write,
	}
}
