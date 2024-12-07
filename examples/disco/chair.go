package main

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type ChairNode = nodes.StructNode[modeling.Mesh, ChairNodeData]

type ChairNodeData struct {
	Height    nodes.NodeOutput[float64]
	Width     nodes.NodeOutput[float64]
	Length    nodes.NodeOutput[float64]
	Thickness nodes.NodeOutput[float64]

	BackHeight             nodes.NodeOutput[float64]
	BackingPieceHeight     nodes.NodeOutput[float64]
	BackingPieceHeightPegs nodes.NodeOutput[int]

	LegRadius nodes.NodeOutput[float64]
	LegInset  nodes.NodeOutput[float64]
}

func (cn ChairNodeData) Process() (modeling.Mesh, error) {
	chairHeight := cn.Height.Value()
	chairWidth := cn.Width.Value()
	chairLength := cn.Length.Value()

	halfHeight := chairHeight / 2
	halfWidth := chairWidth / 2
	halfLength := chairLength / 2

	// LEGS ===================================================================

	legRadius := cn.LegRadius.Value()
	legInset := cn.LegInset.Value()

	leg := primitives.Cylinder{
		Sides:  8,
		Height: chairHeight,
		Radius: legRadius,
	}.ToMesh()

	legRadiusAndInset := legRadius + legInset

	legSupportFrontBackRotation := quaternion.FromTheta(math.Pi/2, vector3.Forward[float64]())
	legFrontBackSupport := primitives.Cylinder{
		Sides:  8,
		Height: chairWidth - (legRadiusAndInset * 2),
		Radius: legRadius / 2,
	}.ToMesh().Transform(
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.PositionAttribute,
			Amount:    legSupportFrontBackRotation,
		},
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.NormalAttribute,
			Amount:    legSupportFrontBackRotation,
		},
	)

	legSupportLeftRightRotation := quaternion.FromTheta(math.Pi/2, vector3.Right[float64]())
	legLeftRightSupport := primitives.Cylinder{
		Sides:  8,
		Height: chairLength - (legRadiusAndInset * 2),
		Radius: legRadius / 2,
	}.ToMesh().Transform(
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.PositionAttribute,
			Amount:    legSupportLeftRightRotation,
		},
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.NormalAttribute,
			Amount:    legSupportLeftRightRotation,
		},
	)

	// BACK ===================================================================

	backHeight := cn.BackHeight.Value()
	halfBackHeight := backHeight / 2

	backPeg := primitives.Cylinder{
		Sides:  8,
		Height: backHeight,
		Radius: legRadius,
	}.ToMesh()

	backSupportRotation := quaternion.FromTheta(math.Pi/2, vector3.Forward[float64]())
	backSupport := primitives.Cylinder{
		Sides:  8,
		Height: chairWidth - (legRadiusAndInset * 2),
		Radius: legRadius / 1.1,
	}.ToMesh().Transform(
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.PositionAttribute,
			Amount:    backSupportRotation,
		},
		meshops.RotateAttribute3DTransformer{
			Attribute: modeling.NormalAttribute,
			Amount:    backSupportRotation,
		},
	)

	backSupportPegHeight := backHeight * cn.BackingPieceHeight.Value()
	backSupportPeg := primitives.Cylinder{
		Sides:  8,
		Height: backSupportPegHeight,
		Radius: legRadius / 1.4,
	}.ToMesh()

	backSupportPegs := repeat.Mesh(
		backSupportPeg,
		repeat.LineExlusive(
			vector3.New(halfWidth-legRadiusAndInset, 0., halfLength-legRadiusAndInset),
			vector3.New(-halfWidth+legRadiusAndInset, 0., halfLength-legRadiusAndInset),
			cn.BackingPieceHeightPegs.Value(),
		),
	)

	return primitives.Cube{
		Height: cn.Thickness.Value(),
		Width:  chairWidth,
		Depth:  chairLength,
		UVs:    primitives.DefaultCubeUVs(),
	}.
		UnweldedQuads().
		Translate(vector3.New(0, chairHeight, 0)).
		// LEGS ===============================================================
		Append(leg.Translate(
			vector3.New(-halfWidth+legRadiusAndInset, halfHeight, -halfLength+legRadiusAndInset),
		)).
		Append(leg.Translate(
			vector3.New(-halfWidth+legRadiusAndInset, halfHeight, halfLength-legRadiusAndInset),
		)).
		Append(leg.Translate(
			vector3.New(halfWidth-legRadiusAndInset, halfHeight, -halfLength+legRadiusAndInset),
		)).
		Append(leg.Translate(
			vector3.New(halfWidth-legRadiusAndInset, halfHeight, halfLength-legRadiusAndInset),
		)).

		// LEG SUPPORT ========================================================
		Append(legFrontBackSupport.Translate(
			vector3.New(0, chairHeight*0.85, halfLength-legRadiusAndInset),
		)).
		Append(legFrontBackSupport.Translate(
			vector3.New(0, chairHeight*0.6, -halfLength+legRadiusAndInset),
		)).
		Append(legFrontBackSupport.Translate(
			vector3.New(0, chairHeight*0.3, -halfLength+legRadiusAndInset),
		)).
		Append(legLeftRightSupport.Translate(
			vector3.New(-halfWidth+legRadiusAndInset, chairHeight*0.45, 0),
		)).
		Append(legLeftRightSupport.Translate(
			vector3.New(-halfWidth+legRadiusAndInset, chairHeight*0.7, 0),
		)).
		Append(legLeftRightSupport.Translate(
			vector3.New(halfWidth-legRadiusAndInset, chairHeight*0.45, 0),
		)).
		Append(legLeftRightSupport.Translate(
			vector3.New(halfWidth-legRadiusAndInset, chairHeight*0.7, 0),
		)).

		// BACK ===============================================================
		Append(backPeg.Translate(
			vector3.New(halfWidth-legRadiusAndInset, chairHeight+halfBackHeight, halfLength-legRadiusAndInset),
		)).
		Append(backPeg.Translate(
			vector3.New(-halfWidth+legRadiusAndInset, chairHeight+halfBackHeight, halfLength-legRadiusAndInset),
		)).
		Append(backSupport.Translate(
			vector3.New(0., chairHeight+halfBackHeight, halfLength-legRadiusAndInset),
		)).
		Append(backSupport.Translate(
			vector3.New(0., chairHeight+halfBackHeight+backSupportPegHeight, halfLength-legRadiusAndInset),
		)).
		Append(backSupportPegs.Translate(
			vector3.New(0., chairHeight+halfBackHeight+(backSupportPegHeight/2), 0),
		)).
		Append(backSupport.Translate(
			vector3.New(0., chairHeight+(halfBackHeight*0.8), halfLength-legRadiusAndInset),
		)).
		Append(backSupport.Translate(
			vector3.New(0., chairHeight+(halfBackHeight*0.55), halfLength-legRadiusAndInset),
		)), nil
}

type CushionNode = nodes.StructNode[modeling.Mesh, CushionNodeData]

type CushionNodeData struct {
	Thickness nodes.NodeOutput[float64]
	Width     nodes.NodeOutput[float64]
	Length    nodes.NodeOutput[float64]
}

func (cn CushionNodeData) Process() (modeling.Mesh, error) {
	return primitives.Cube{
		Height: cn.Thickness.Value(),
		Width:  cn.Width.Value(),
		Depth:  cn.Length.Value(),
		UVs:    primitives.DefaultCubeUVs(),
	}.UnweldedQuads(), nil
}
