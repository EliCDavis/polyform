package noise

import (
	"math"
	"math/rand"

	"github.com/EliCDavis/vector/vector2"
)

// type TilingNoiseConfiguration struct {
// 	Size     int
// 	Frequncy float64
// 	Octaves  int
// }

// Tiling noise implementation comes from this answer:
// https://gamedev.stackexchange.com/questions/23625/how-do-you-generate-tileable-perlin-noise
//
// And simply takes advantage of perlin noise repeatability
type TilingNoise struct {
	dirs []vector2.Float64
	perm []int

	// For aggregation of noise across octaves
	size     int
	frequncy float64
	octaves  int
}

func NewTilingNoise(size int, frequncy float64, octaves int) *TilingNoise {
	tn := &TilingNoise{
		size:     size,
		frequncy: frequncy,
		octaves:  octaves,
	}
	tn.init()
	return tn
}

func NewDefaultTilingNoise() *TilingNoise {
	return NewTilingNoise(256, 1/64., 5)
}

func (tn *TilingNoise) init() {
	size := tn.size

	tn.perm = make([]int, size)
	for i := range size {
		tn.perm[i] = i
	}
	rand.Shuffle(len(tn.perm), func(i, j int) { tn.perm[i], tn.perm[j] = tn.perm[j], tn.perm[i] })
	tn.perm = append(tn.perm, tn.perm...)

	tn.dirs = make([]vector2.Float64, size)
	r := (2. * math.Pi) / float64(size)
	for i := range size {
		a := float64(i) * r
		tn.dirs[i] = vector2.New(math.Cos(a), math.Sin(a))
	}
}

func (tn *TilingNoise) surflet(v vector2.Float64, g vector2.Int, per int) float64 {
	dist := v.Sub(g.ToFloat64()).Abs()
	polyX := 1 - (6 * math.Pow(dist.X(), 5)) + (15 * math.Pow(dist.X(), 4)) - (10 * math.Pow(dist.X(), 3))
	polyY := 1 - (6 * math.Pow(dist.Y(), 5)) + (15 * math.Pow(dist.Y(), 4)) - (10 * math.Pow(dist.Y(), 3))

	hashed := tn.perm[tn.perm[g.X()%per]+(g.Y()%per)]

	hashedDir := tn.dirs[hashed]
	grad := ((v.X() - float64(g.X())) * hashedDir.X()) + ((v.Y() - float64(g.Y())) * hashedDir.Y())
	return polyX * polyY * grad
}

func (tn *TilingNoise) NoiseAtPermutation(v vector2.Float64, per int) float64 {
	i := v.FloorToInt()
	return tn.surflet(v, i, per) +
		tn.surflet(v, i.Add(vector2.Right[int]()), per) +
		tn.surflet(v, i.Add(vector2.Up[int]()), per) +
		tn.surflet(v, i.Add(vector2.One[int]()), per)
}

func (tn *TilingNoise) Noise(x, y int) float64 {
	val := 0.
	sf := int(float64(tn.size) * tn.frequncy)
	octaveFreq := 1.
	octaveStrength := 1.
	for range tn.octaves {
		n := tn.NoiseAtPermutation(
			vector2.New(
				(float64(x)*tn.frequncy)*octaveFreq,
				(float64(y)*tn.frequncy)*octaveFreq,
			),
			sf*int(octaveFreq),
		)
		val += n * octaveStrength

		octaveStrength *= 0.5
		octaveFreq *= 2
	}
	return val
}
