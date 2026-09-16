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

import "github.com/EliCDavis/vector/vector3"

// Rays are cheap once there are only a few of them, so a region is sampled
// rather than trusted to a single face. An even number would risk a tie.
const votesPerPatch = 5

func curvePoints(weld *welder, corners []vector3.Float64) map[int]bool {
	onCurve := make(map[int]bool, len(corners))
	for _, c := range corners {
		onCurve[weld.Index(c)] = true
	}
	return onCurve
}

// Faces sharing an edge that does not sit on the intersection curve.
//
// Marking an edge by whether both its ends lie on the curve is a slight over
// count: an edge can join two curve points without lying on the curve itself.
// That only ever splits a region in two, costing a few more rays, where the
// opposite mistake would hand one region's answer to another.
//
// A face lying on the other solid's surface registers no edge at all, so
// nothing joins to it and it answers for itself. No cut separates it from its
// neighbours, but it classifies as same or opposite where they do not.
func patchesOf(ids [][3]int, onCurve map[int]bool, touching []bool) [][]int {
	parent := make([]int, len(ids))
	for i := range parent {
		parent[i] = i
	}

	root := func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}

	// Only which faces end up together matters, so edges are joined as they
	// are met and never stored. Holding a face list per edge costs more than
	// the answer is worth.
	first := make(map[[2]int]int, len(ids)*3)
	for i, corners := range ids {
		if touching[i] {
			continue
		}
		for k := 0; k < 3; k++ {
			a, b := corners[k], corners[(k+1)%3]
			if a == b || (onCurve[a] && onCurve[b]) {
				continue
			}
			if a > b {
				a, b = b, a
			}

			edge := [2]int{a, b}
			seen, known := first[edge]
			if !known {
				first[edge] = i
				continue
			}
			if left, right := root(seen), root(i); left != right {
				parent[left] = right
			}
		}
	}

	grouped := make(map[int][]int)
	for i := range ids {
		r := root(i)
		grouped[r] = append(grouped[r], i)
	}

	patches := make([][]int, 0, len(grouped))
	for _, group := range grouped {
		patches = append(patches, group)
	}
	return patches
}

func sampleStride(group []int) int {
	if stride := len(group) / votesPerPatch; stride > 1 {
		return stride
	}
	return 1
}

// How many rays classifying these patches will cost, which is what decides
// whether the solid being cast against is worth indexing.
func rayBudget(patches [][]int) int {
	total := 0
	for _, group := range patches {
		stride := sampleStride(group)
		for i := 0; i < len(group); i += stride {
			total++
		}
	}
	return total
}

func classifyPatches(faces []face, patches [][]int, against *solid) []classification {
	answers := make([]classification, len(faces))

	for _, group := range patches {
		stride := sampleStride(group)

		votes := make(map[classification]int, 4)
		for i := 0; i < len(group); i += stride {
			votes[against.classify(faces[group[i]])]++
		}

		winner, best := outside, -1
		for answer, count := range votes {
			if count > best || (count == best && answer < winner) {
				winner, best = answer, count
			}
		}

		for _, i := range group {
			answers[i] = winner
		}
	}

	return answers
}
