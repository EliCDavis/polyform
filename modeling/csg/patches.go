package csg

// Section 8, "Marking Vertices".
//
// "The polygon classification routine could be called to classify each polygon
// in both objects, but this would be time-consuming. Instead, all the vertices
// of the object are classified by classifying just a few of the polygons."
//
// The paper propagates through vertices. The same argument works a level up
// and is easier to reason about: a face's answer can only differ from its
// neighbour's where the two surfaces actually cross, so faces reachable from
// each other without stepping over the intersection curve share one answer.
// Whole regions then cost a handful of rays between them rather than one each,
// and because a region is large the rays can be fired from faces well clear of
// any edge.

// Rays are cheap once there are only a few of them, so a region is sampled
// rather than trusted to a single face.
const votesPerPatch = 5

// Faces sharing an edge that does not sit on the intersection curve. Each
// patch is a list of face indices.
//
// Marking an edge by whether both its ends lie on the curve is a slight over
// count: an edge can join two curve points without lying on the curve itself.
// That only ever splits a region in two, costing a few more rays, where the
// opposite mistake would hand one region's answer to another.
//
// A face lying on the other solid's surface registers no edge at all, so
// nothing joins to it and it answers for itself. No cut separates it from its
// neighbours, but it classifies as same or opposite where they do not.
func patchesOf(cornerIDs [][3]int, onCurve map[int]bool, touching []bool) [][]int {
	parent := make([]int, len(cornerIDs))
	for i := range parent {
		parent[i] = i
	}

	root := func(faceIndex int) int {
		for parent[faceIndex] != faceIndex {
			parent[faceIndex] = parent[parent[faceIndex]]
			faceIndex = parent[faceIndex]
		}
		return faceIndex
	}

	// Only which faces end up together matters, so edges are joined as they
	// are met and never stored. Holding a face list per edge costs more than
	// the answer is worth.
	firstFaceOnEdge := make(map[[2]int]int, len(cornerIDs)*3)
	for faceIndex, corners := range cornerIDs {
		if touching[faceIndex] {
			continue
		}
		for k := 0; k < 3; k++ {
			from, to := corners[k], corners[(k+1)%3]
			if from == to || (onCurve[from] && onCurve[to]) {
				continue
			}
			if from > to {
				from, to = to, from
			}

			edge := [2]int{from, to}
			earlierFace, seen := firstFaceOnEdge[edge]
			if !seen {
				firstFaceOnEdge[edge] = faceIndex
				continue
			}
			if earlierRoot, thisRoot := root(earlierFace), root(faceIndex); earlierRoot != thisRoot {
				parent[earlierRoot] = thisRoot
			}
		}
	}

	byRoot := make(map[int][]int)
	for faceIndex := range cornerIDs {
		r := root(faceIndex)
		byRoot[r] = append(byRoot[r], faceIndex)
	}

	patches := make([][]int, 0, len(byRoot))
	for _, patch := range byRoot {
		patches = append(patches, patch)
	}
	return patches
}

func sampleStride(patch []int) int {
	if stride := len(patch) / votesPerPatch; stride > 1 {
		return stride
	}
	return 1
}

// How many rays classifying these patches will cost, which is what decides
// whether the solid being cast against is worth indexing.
func rayBudget(patches [][]int) int {
	total := 0
	for _, patch := range patches {
		stride := sampleStride(patch)
		for i := 0; i < len(patch); i += stride {
			total++
		}
	}
	return total
}

func classifyPatches(faces []face, patches [][]int, against *solid) []classification {
	answers := make([]classification, len(faces))

	for _, patch := range patches {
		stride := sampleStride(patch)

		votes := make(map[classification]int, 4)
		for i := 0; i < len(patch); i += stride {
			votes[against.classify(faces[patch[i]])]++
		}

		winner, mostVotes := outside, -1
		for answer, count := range votes {
			if count > mostVotes || (count == mostVotes && answer < winner) {
				winner, mostVotes = answer, count
			}
		}

		for _, faceIndex := range patch {
			answers[faceIndex] = winner
		}
	}

	return answers
}
