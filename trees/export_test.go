package trees

import "github.com/EliCDavis/polyform/math/geometry"

type TreeStats = treeStats

func Stats(tree *OctTree) TreeStats {
	return statsOf(tree, 1)
}

func RayTests(tree *OctTree, ray geometry.Ray, min, max float64) int {
	return rayTests(tree, ray, min, max)
}
