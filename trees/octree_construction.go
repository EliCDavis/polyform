package trees

type OctreeConsturctionSettings struct {
	elements  []Element
	maxDepth  int
	tolerance float64
}

type OctreeConsturctionSetting interface {
	Configure(*OctreeConsturctionSettings)
}

// Max Depth ==================================================================

// MaxDepth caps how many times the tree subdivides. Defaults to log8 of the
// element count.
func MaxDepth(depth int) OctreeConsturctionSetting {
	return maxDepthSetting{depth: depth}
}

type maxDepthSetting struct {
	depth int
}

func (mds maxDepthSetting) Configure(setting *OctreeConsturctionSettings) {
	setting.maxDepth = mds.depth
}

// Tolerance ==================================================================

// Tolerance keeps an element at a node instead of pushing it into a child when
// its bounds span at least (1 - tolerance) of the node's along some axis.
// 0 keeps only elements as wide as the node, 1 keeps everything at the root.
func Tolerance(tolerance float64) OctreeConsturctionSetting {
	return toleranceSetting{tolerance: tolerance}
}

type toleranceSetting struct {
	tolerance float64
}

func (mds toleranceSetting) Configure(setting *OctreeConsturctionSettings) {
	setting.tolerance = mds.tolerance
}

// NoRetention pushes every element down to a leaf.
func NoRetention() OctreeConsturctionSetting {
	return toleranceSetting{tolerance: noRetention}
}
