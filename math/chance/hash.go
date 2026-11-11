package chance

import (
	"hash/fnv"
)

const (
	golden = 0x9E3779B97F4A7C15
	keyMul = 0xD1B54A32D192ED03
)

func mix(x uint64) uint64 {
	x ^= x >> 30
	x *= 0xBF58476D1CE4E5B9
	x ^= x >> 27
	x *= 0x94D049BB133111EB
	x ^= x >> 31
	return x
}

func Hash(seed, key int) uint64 {
	return mix(mix(uint64(seed)+golden) ^ (uint64(key)*keyMul + golden))
}

func unit(h uint64) float64 {
	return float64(h>>11) / (1 << 53)
}

// Unit is the value in [0, 1) that (seed, key) always maps to.
func Unit(seed, key int) float64 {
	return unit(Hash(seed, key))
}

// Stream hands out independent values in [0, 1).
type Stream func() float64

func Roll(seed, key int) Stream {
	h := Hash(seed, key)
	next, n := h, uint64(0)
	return func() float64 {
		v := unit(next)
		n++
		next = mix(h + n)
		return v
	}
}

// Fixed is a stream that only ever gives u.
func Fixed(u float64) Stream {
	return func() float64 { return u }
}

// Derive is a seed of its own for every label, so parts sharing one master
// seed draw unrelated values.
func Derive(seed int, label string) int {
	f := fnv.New64a()
	f.Write([]byte(label))
	return int(mix(mix(uint64(seed)+golden) ^ f.Sum64()))
}
