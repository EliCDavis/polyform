package extrude

import (
	"math"
	"testing"

	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathFramesTurnOnlyAsMuchAsThePath(t *testing.T) {
	straight := []vector3.Float64{}
	for i := 0; i <= 6; i++ {
		straight = append(straight, vector3.New(0., float64(i)*0.5, 0.))
	}

	helix := []vector3.Float64{}
	for i := 0; i <= 24; i++ {
		a := float64(i) * 0.4
		helix = append(helix, vector3.New(math.Cos(a)*2, float64(i)*0.2, math.Sin(a)*2))
	}

	loop, saddle, trefoil := []vector3.Float64{}, []vector3.Float64{}, []vector3.Float64{}
	for i := 0; i < 24; i++ {
		a := math.Pi * 2 * float64(i) / 24
		loop = append(loop, vector3.New(math.Cos(a)*3, 0, math.Sin(a)*3))
		saddle = append(saddle, vector3.New(math.Cos(a)*3, math.Sin(2*a)*1.2, math.Sin(a)*3))
		trefoil = append(trefoil, vector3.New(
			math.Sin(a)+2*math.Sin(2*a), -math.Sin(3*a), math.Cos(a)-2*math.Cos(2*a)))
	}

	for name, tc := range map[string]struct {
		path   []vector3.Float64
		closed bool
	}{
		"straight run": {straight, false},
		"straight run into a bend": {[]vector3.Float64{
			vector3.New(0., 0., 0.), vector3.New(0., 1., 0.), vector3.New(0., 2., 0.),
			vector3.New(1., 3., 0.), vector3.New(2.5, 3.5, 0.), vector3.New(4.5, 3.5, 0.),
		}, false},
		"collinear but not axis aligned": {[]vector3.Float64{
			vector3.New(0., 0., 0.),
			vector3.New(0., 1.3001528736574441, -0.37084847944351573),
			vector3.New(0., 2.6003057473148883, -0.7416969588870315),
		}, false},
		"arch": {[]vector3.Float64{
			vector3.New(0., 0., 0.), vector3.New(0., 5., 0.),
			vector3.New(5., 5., 0.), vector3.New(5., 0., 0.),
		}, false},
		"helix":                  {helix, false},
		"closed loop":            {loop, true},
		"closed non planar loop": {saddle, true},
		"closed trefoil":         {trefoil, true},
	} {
		t.Run(name, func(t *testing.T) {
			frames := pathFrames(tc.path, tc.closed)
			require.Len(t, frames, len(tc.path))

			for i, f := range frames {
				require.InDeltaf(t, 1, f.x.Length(), 1e-9, "ring %d x is not a unit vector", i)
				require.InDeltaf(t, 1, f.tangent.Length(), 1e-9, "ring %d tangent is not a unit vector", i)
				assert.InDeltaf(t, 0, f.x.Dot(f.tangent), 1e-9, "ring %d x is not square to the tangent", i)
				assert.InDeltaf(t, 0, f.y.Sub(f.x.Cross(f.tangent)).Length(), 1e-9, "ring %d is mirrored", i)

				previous := i - 1
				if i == 0 {
					if !tc.closed {
						continue
					}
					previous = len(frames) - 1
				}

				// A closed loop cannot be swept without any twist at all: the
				// leftover is topological. It only has to be shared out, not
				// dumped on the segment that closes the loop.
				allowed := 1e-9
				if tc.closed {
					allowed = 1 * math.Pi / 180
				}

				chord := f.origin.Sub(frames[previous].origin).Normalized()
				turned := frames[previous].tangent.Angle(chord) + chord.Angle(f.tangent)
				twisted := frames[previous].x.Angle(f.x)
				assert.LessOrEqualf(t, twisted, turned+allowed,
					"ring %d twisted %.2f degrees while the path only turned %.2f",
					i, twisted*180/math.Pi, turned*180/math.Pi)
			}
		})
	}
}

func TestPathFramesLayAProfileFlatHoweverALevelPathStarts(t *testing.T) {
	for name, path := range map[string][]vector3.Float64{
		"along +X":       {vector3.New(0., 0., 0.), vector3.New(5., 0., 0.)},
		"along -X":       {vector3.New(0., 0., 0.), vector3.New(-5., 0., 0.)},
		"along +Z":       {vector3.New(0., 0., 0.), vector3.New(0., 0., 5.)},
		"along -Z":       {vector3.New(0., 0., 0.), vector3.New(0., 0., -5.)},
		"turning":        {vector3.New(0., 0., 0.), vector3.New(5., 0., 0.), vector3.New(5., 0., 5.)},
		"doubling back":  {vector3.New(0., 0., 0.), vector3.New(5., 0., 0.), vector3.New(5., 0., 1.), vector3.New(0., 0., 1.)},
		"climbing first": {vector3.New(0., 0., 0.), vector3.New(0., 3., 0.), vector3.New(4., 3., 0.)},
		"with a hairpin": {vector3.New(0., 0., 0.), vector3.New(5., 0., 0.), vector3.New(0., 0., .1)},
	} {
		t.Run(name, func(t *testing.T) {
			for i, f := range pathFrames(path, false) {
				if math.Abs(f.tangent.Y()) > 1e-9 {
					continue
				}
				assert.InDeltaf(t, 1, f.y.Y(), 1e-9, "ring %d does not point the profile's y up", i)
			}
		})
	}
}

func TestPathFramesBevelATurnTooSharpToMiter(t *testing.T) {
	frames := pathFrames([]vector3.Float64{
		vector3.New(0., 0., 0.), vector3.New(5., 0., 0.), vector3.New(0., 0., .1),
	}, false)

	require.Len(t, frames, 4, "the hairpin is two rings")
	for i, f := range frames {
		assert.LessOrEqualf(t, f.stretch, maxMiterStretch, "ring %d", i)
	}
	assert.Equal(t, frames[1].origin, frames[2].origin)
}
