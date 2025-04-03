package meshops

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
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
	RequireTopology(m, modeling.TriangleTopology)

	originalIndices := m.Indices()
	numFaces := originalIndices.Len() / 3

	belowPlaneIndices := make([]int, 0)
	abovePlaneIndices := make([]int, 0)

	// Mark which tris belong in retained or clipped
	for t := 0; t < numFaces; t++ {
		tri := m.Tri(t)

		aClip := plane.Normal().Dot(tri.P1Vec3Attr(attribute).Sub(plane.Origin())) < 0
		bClip := plane.Normal().Dot(tri.P2Vec3Attr(attribute).Sub(plane.Origin())) < 0
		cClip := plane.Normal().Dot(tri.P3Vec3Attr(attribute).Sub(plane.Origin())) < 0

		if !aClip && !bClip && !cClip {
			belowPlaneIndices = append(belowPlaneIndices, tri.P1(), tri.P2(), tri.P3())
		} else if aClip && bClip && cClip {
			abovePlaneIndices = append(abovePlaneIndices, tri.P1(), tri.P2(), tri.P3())
		} else {

			lineIntersections := make([]geometry.Line3D, 0, 2)
			if (aClip && !bClip) || (!aClip && bClip) {
				lineIntersections = append(lineIntersections, tri.L1(attribute))
			}

			if (bClip && !cClip) || (!bClip && cClip) {
				lineIntersections = append(lineIntersections, tri.L2(attribute))
			}

			if (aClip && !cClip) || (!aClip && cClip) {
				lineIntersections = append(lineIntersections, tri.L3(attribute))
			}

			// intersectionA, _ := lineIntersections[0].IntersectionTimeOnPlane(plane)
			// intersectionB, _ := lineIntersections[1].IntersectionTimeOnPlane(plane)

		}
	}

	v4Data := readAllFloat4Data(m)
	v3Data := readAllFloat3Data(m)
	v2Data := readAllFloat2Data(m)
	v1Data := readAllFloat1Data(m)

	above := modeling.NewMesh(m.Topology(), abovePlaneIndices).
		SetFloat4Data(v4Data).
		SetFloat3Data(v3Data).
		SetFloat2Data(v2Data).
		SetFloat1Data(v1Data)

	below := modeling.NewMesh(m.Topology(), belowPlaneIndices).
		SetFloat4Data(v4Data).
		SetFloat3Data(v3Data).
		SetFloat2Data(v2Data).
		SetFloat1Data(v1Data)

	return RemovedUnreferencedVertices(above), RemovedUnreferencedVertices(below)
}
