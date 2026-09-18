package csg

// Section 8. Faces that can reach each other without crossing the
// intersection curve share one answer, so a region needs only a few rays.

// Rays are cheap once there are only a few of them, so a region is sampled
// rather than trusted to a single face.
const votesPerPatch = 5

// Groups faces sharing an edge off the intersection curve, as lists of face
// indices. A face on the other solid's surface joins nothing and stands alone.
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
	// are met and never stored.
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

func classifyPatches(faces []face, patches [][]int, against *target) []classification {
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
