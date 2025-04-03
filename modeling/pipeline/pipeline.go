package pipeline

import (
	"sync"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type Pipeline struct {
	waves [][]Command
}

type MeshView struct {
	v4Data   map[string][]vector4.Float64
	v3Data   map[string][]vector3.Float64
	v2Data   map[string][]vector2.Float64
	v1Data   map[string][]float64
	indices  []int
	topology modeling.Topology
}

func (wip MeshView) Mesh() modeling.Mesh {
	return modeling.EmptyMesh(wip.topology)
}

func setupVxReadAttr[T any](perms map[string]ReadArrayPermission[T], workingData map[string][]T) {
	for attr, perm := range perms {
		if data, ok := workingData[attr]; ok {
			perm.data = data
		}
	}
}

func setupVxWriteAttr[T any](perms map[string]WriteArrayPermission[T], workingData map[string][]T) {
	for attr, perm := range perms {
		if data, ok := workingData[attr]; ok {
			perm.data = data
		}
	}
}

func teardownVxReadAttr[T any](perms map[string]ReadArrayPermission[T]) {
	for _, perm := range perms {
		perm.data = nil
	}
}

func teardownVxWriteAttr[T any](perms map[string]WriteArrayPermission[T]) {
	for _, perm := range perms {
		perm.data = nil
	}
}

func setupCommand(command Command, mesh MeshView) {
	read := command.ReadPermissions()
	write := command.WritePermissions()

	m := mesh.Mesh()

	setupVxReadAttr(read.V4Permissions, mesh.v4Data)
	setupVxReadAttr(read.V3Permissions, mesh.v3Data)
	setupVxReadAttr(read.V2Permissions, mesh.v2Data)
	setupVxReadAttr(read.V1Permissions, mesh.v1Data)

	if read.Everything != nil {
		read.Everything.data = m
	}

	if read.Indices != nil {
		read.Indices.data = mesh.indices
		read.Indices.m = &m
	}

	setupVxWriteAttr(write.V4Permissions, mesh.v4Data)
	setupVxWriteAttr(write.V3Permissions, mesh.v3Data)
	setupVxWriteAttr(write.V2Permissions, mesh.v2Data)
	setupVxWriteAttr(write.V1Permissions, mesh.v1Data)

	if write.Everything != nil {
		write.Everything.data = m
	}

	if write.Indices != nil {
		write.Indices.data = mesh.indices
	}
}

func teardownCommand(command Command, wip *MeshView) {
	read := command.ReadPermissions()
	write := command.WritePermissions()

	if read.Everything != nil {
		read.Everything.data = modeling.Mesh{}
	}

	if read.Indices != nil {
		read.Indices.data = nil
		read.Indices.m = nil
	}

	teardownVxReadAttr(read.V4Permissions)
	teardownVxReadAttr(read.V3Permissions)
	teardownVxReadAttr(read.V2Permissions)
	teardownVxReadAttr(read.V1Permissions)

	teardownVxWriteAttr(write.V4Permissions)
	teardownVxWriteAttr(write.V3Permissions)
	teardownVxWriteAttr(write.V2Permissions)
	teardownVxWriteAttr(write.V1Permissions)

	if write.Everything != nil {
		write.Everything.data = modeling.Mesh{}
		write.Everything.written = false
	}

	if write.Indices != nil {
		write.Indices.data = nil
	}
}

func (s Pipeline) RunSynchronous(m modeling.Mesh) modeling.Mesh {
	finalMesh := m
	wip := MeshView{}
	for _, wave := range s.waves {
		for _, c := range wave {
			setupCommand(c, wip)
			c.Run()
			teardownCommand(c, &wip)
			finalMesh = wip.Mesh()
		}
	}
	return finalMesh
}

func (s Pipeline) Run(m modeling.Mesh) modeling.Mesh {
	finalMesh := m
	wip := MeshView{}
	for _, wave := range s.waves {
		var wg sync.WaitGroup
		for _, c := range wave {
			wg.Add(1)
			go func(c Command) {
				defer wg.Done()
				setupCommand(c, wip)
				c.Run()
				teardownCommand(c, &wip)
			}(c)
		}
		wg.Wait()
		finalMesh = wip.Mesh()
	}
	return finalMesh
}
