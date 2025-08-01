package meshops

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type SliceByPlaneTransformerSide int

const (
	AbovePlane SliceByPlaneTransformerSide = iota
	BelowPlane
)

type SliceByPlaneTransformer struct {
	Attribute   string
	SliceToKeep SliceByPlaneTransformerSide
	Plane       geometry.Plane
}

func (sbpt SliceByPlaneTransformer) attribute() string {
	return sbpt.Attribute
}

func (sbpt SliceByPlaneTransformer) Transform(m modeling.Mesh) (results modeling.Mesh, err error) {
	attribute := getAttribute(sbpt, modeling.PositionAttribute)

	if err = RequireV3Attribute(m, attribute); err != nil {
		return
	}

	above, below := SliceByPlaneWithAttribute(m, sbpt.Plane, attribute)
	if sbpt.SliceToKeep == AbovePlane {
		return above, nil
	}
	return below, nil
}

func SliceByPlaneWithAttribute(m modeling.Mesh, plane geometry.Plane, attribute string) (modeling.Mesh, modeling.Mesh) {
	return sliceTrianglesByPlaneWithAttribute(m, plane, attribute)
}

type SliceAttributeByPlaneNode struct {
	Mesh      nodes.Output[modeling.Mesh]
	Plane     nodes.Output[geometry.Plane]
	Attribute nodes.Output[string]
}

func (n SliceAttributeByPlaneNode) slice(out *nodes.StructOutput[modeling.Mesh], above bool) {
	if n.Mesh == nil {
		out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
		return
	}

	mesh := nodes.GetOutputValue(out, n.Mesh)
	if n.Plane == nil {
		out.Set(mesh)
		return
	}
	plane := nodes.GetOutputValue(out, n.Plane)

	aboveM, belowM := SliceByPlaneWithAttribute(
		mesh,
		plane,
		nodes.TryGetOutputValue(out, n.Attribute, modeling.PositionAttribute),
	)
	if above {
		out.Set(aboveM)
	} else {
		out.Set(belowM)
	}
}

func (n SliceAttributeByPlaneNode) AbovePlane(out *nodes.StructOutput[modeling.Mesh]) {
	n.slice(out, true)
}

func (n SliceAttributeByPlaneNode) BelowPlane(out *nodes.StructOutput[modeling.Mesh]) {
	n.slice(out, false)
}

func trianglePlaneIntersectionPoints(s1p1, s2p1, s2p2 vector3.Float64, plane geometry.Plane) (vector3.Float64, vector3.Float64) {
	l1 := geometry.NewLine3D(s1p1, s2p1)
	l2 := geometry.NewLine3D(s1p1, s2p2)

	time1, intersects := l1.IntersectionTimeOnPlane(plane)
	if !intersects {
		panic("unsupported scenario where triangle doesn't intersect plane")
	}

	time2, intersects := l2.IntersectionTimeOnPlane(plane)
	if !intersects {
		panic("unsupported scenario where triangle doesn't intersect plane")
	}

	return l1.AtTime(time1), l2.AtTime(time2)
}

func alignWithNormal(normal vector3.Float64, vertices []vector3.Float64, indices []int, i1, i2, i3 int) []int {
	normalized := vertices[i2].Sub(vertices[i1]).Cross(vertices[i3].Sub(vertices[i1])).Normalized()
	if normal.Dot(normalized) > 0 {
		return append(indices, i1, i2, i3)
	}
	return append(indices, i1, i3, i2)
}

func sliceTrianglesByPlaneWithAttribute(m modeling.Mesh, plane geometry.Plane, attribute string) (modeling.Mesh, modeling.Mesh) {
	if err := RequireTopology(m, modeling.TriangleTopology); err != nil {
		panic(err)
	}
	if err := RequireV3Attribute(m, attribute); err != nil {
		panic(err)
	}

	originalIndices := m.Indices()
	numFaces := originalIndices.Len() / 3

	belowPlaneIndices := make([]int, 0)
	abovePlaneIndices := make([]int, 0)

	attributeData := m.Float3Attribute(attribute)
	belowPlaneVertices := make([]vector3.Float64, attributeData.Len())
	abovePlaneVertices := make([]vector3.Float64, attributeData.Len())
	for i := range attributeData.Len() {
		belowPlaneVertices[i] = attributeData.At(i)
		abovePlaneVertices[i] = attributeData.At(i)
	}

	// Mark which tris belong in retained or clipped
	for t := range numFaces {
		tri := m.Tri(t)
		triNormal := tri.Normal(attribute)

		a := tri.P1Vec3Attr(attribute)
		b := tri.P2Vec3Attr(attribute)
		c := tri.P3Vec3Attr(attribute)

		aClip := plane.Normal().Dot(a.Sub(plane.Origin())) < 0
		bClip := plane.Normal().Dot(b.Sub(plane.Origin())) < 0
		cClip := plane.Normal().Dot(c.Sub(plane.Origin())) < 0

		keepVertices := make([]vector3.Float64, 0, 3)
		removedVertices := make([]vector3.Float64, 0, 3)

		keepIndices := make([]int, 0, 3)
		removedIndices := make([]int, 0, 3)
		if aClip {
			keepVertices = append(keepVertices, a)
			keepIndices = append(keepIndices, tri.P1())
		} else {
			removedVertices = append(removedVertices, a)
			removedIndices = append(removedIndices, tri.P1())
		}

		if bClip {
			keepVertices = append(keepVertices, b)
			keepIndices = append(keepIndices, tri.P2())
		} else {
			removedVertices = append(removedVertices, b)
			removedIndices = append(removedIndices, tri.P2())
		}

		if cClip {
			keepVertices = append(keepVertices, c)
			keepIndices = append(keepIndices, tri.P3())
		} else {
			removedVertices = append(removedVertices, c)
			removedIndices = append(removedIndices, tri.P3())
		}

		switch len(keepVertices) {
		case 0:
			belowPlaneIndices = append(belowPlaneIndices, tri.P1(), tri.P2(), tri.P3())

		case 1:
			// One point is kept
			// The triangle is discarded, and a new one generated
			// The two new points of the new clip triangle are the intersection of the plane.
			newV1, newV2 := trianglePlaneIntersectionPoints(keepVertices[0], removedVertices[0], removedVertices[1], plane)

			newKeepIndice1 := len(abovePlaneVertices)
			newKeepIndice2 := newKeepIndice1 + 1

			abovePlaneVertices = append(abovePlaneVertices, newV1, newV2)
			abovePlaneIndices = alignWithNormal(triNormal, abovePlaneVertices, abovePlaneIndices, keepIndices[0], newKeepIndice1, newKeepIndice2)

			newRemoveIndice1 := len(belowPlaneVertices)
			newRemoveIndice2 := newRemoveIndice1 + 1

			belowPlaneVertices = append(belowPlaneVertices, newV1, newV2)
			belowPlaneIndices = alignWithNormal(triNormal, belowPlaneVertices, belowPlaneIndices, newRemoveIndice1, removedIndices[0], newRemoveIndice2)
			belowPlaneIndices = alignWithNormal(triNormal, belowPlaneVertices, belowPlaneIndices, newRemoveIndice2, removedIndices[0], removedIndices[1])

		case 2:
			// Two points are kept.
			// The triangle is discarded, and two new triangles are generated
			// This is just the inverse of case 1:
			newV1, newV2 := trianglePlaneIntersectionPoints(removedVertices[0], keepVertices[0], keepVertices[1], plane)

			newKeepIndice1 := len(belowPlaneVertices)
			newKeepIndice2 := newKeepIndice1 + 1

			belowPlaneVertices = append(belowPlaneVertices, newV1, newV2)
			belowPlaneIndices = alignWithNormal(triNormal, belowPlaneVertices, belowPlaneIndices, removedIndices[0], newKeepIndice1, newKeepIndice2)

			newRemoveIndice1 := len(abovePlaneVertices)
			newRemoveIndice2 := newRemoveIndice1 + 1

			abovePlaneVertices = append(abovePlaneVertices, newV1, newV2)
			abovePlaneIndices = alignWithNormal(triNormal, abovePlaneVertices, abovePlaneIndices, newRemoveIndice1, keepIndices[0], newRemoveIndice2)
			abovePlaneIndices = alignWithNormal(triNormal, abovePlaneVertices, abovePlaneIndices, newRemoveIndice2, keepIndices[0], keepIndices[1])

		case 3:
			abovePlaneIndices = append(abovePlaneIndices, tri.P1(), tri.P2(), tri.P3())

		}

	}

	// v4Data := readAllFloat4Data(m)
	// v3Data := readAllFloat3Data(m)
	// v2Data := readAllFloat2Data(m)
	// v1Data := readAllFloat1Data(m)

	// above := modeling.NewMesh(m.Topology(), abovePlaneIndices).
	// 	SetFloat4Data(v4Data).
	// 	SetFloat3Data(v3Data).
	// 	SetFloat2Data(v2Data).
	// 	SetFloat1Data(v1Data)

	// below := modeling.NewMesh(m.Topology(), belowPlaneIndices).
	// 	SetFloat4Data(v4Data).
	// 	SetFloat3Data(v3Data).
	// 	SetFloat2Data(v2Data).
	// 	SetFloat1Data(v1Data)

	// return RemovedUnreferencedVertices(above), RemovedUnreferencedVertices(below)

	aboveMesh := modeling.NewTriangleMesh(abovePlaneIndices).
		SetFloat3Attribute(attribute, abovePlaneVertices)

	belowMesh := modeling.NewTriangleMesh(belowPlaneIndices).
		SetFloat3Attribute(attribute, belowPlaneVertices)

	return RemovedUnreferencedVertices(aboveMesh), RemovedUnreferencedVertices(belowMesh)
}
