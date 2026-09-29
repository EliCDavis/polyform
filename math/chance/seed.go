package chance

import "math/rand/v2"

const seedSalt = 0x9E3779B97F4A7C15

func FromSeed(seed int) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(seed), seedSalt))
}
