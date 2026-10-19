package mcp

import "testing"

func TestResolvePortName(t *testing.T) {
	ports := map[string]struct{}{
		"Color Texture": {},
		"Radius 2":      {},
		"Height":        {},
	}

	cases := []struct {
		name      string
		requested string
		want      string
	}{
		{"exact match passes through", "Color Texture", "Color Texture"},
		{"missing internal space", "ColorTexture", "Color Texture"},
		{"missing space before digit", "Radius2", "Radius 2"},
		{"case-insensitive too", "colortexture", "Color Texture"},
		{"single-word port needs no resolution", "Height", "Height"},
		{"genuinely unknown port passes through unchanged", "NotAPort", "NotAPort"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := resolvePortName(c.requested, ports)
			if got != c.want {
				t.Errorf("resolvePortName(%q) = %q, want %q", c.requested, got, c.want)
			}
		})
	}
}

func TestResolvePortNameAmbiguous(t *testing.T) {
	// "A B" and "AB" both normalize to "ab" - two real ports collapse to
	// the same normalized key, so a caller typing "ab" is genuinely
	// ambiguous and must be passed through unchanged rather than guessed.
	ports := map[string]struct{}{
		"A B": {},
		"AB":  {},
	}
	got := resolvePortName("ab", ports)
	if got != "ab" {
		t.Errorf("resolvePortName on an ambiguous normalized match = %q, want the original %q unchanged", got, "ab")
	}
}

func TestNormalizePortKey(t *testing.T) {
	cases := map[string]string{
		"Color Texture": "colortexture",
		"ColorTexture":  "colortexture",
		"Radius 2":      "radius2",
		"Radius2":       "radius2",
	}
	for in, want := range cases {
		if got := normalizePortKey(in); got != want {
			t.Errorf("normalizePortKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveOutputPortFallsBackToPrefixThenOnlyPort(t *testing.T) {
	half := map[string]struct{}{"Int": {}, "Float 64": {}}
	negate := map[string]struct{}{"Out": {}}
	add := map[string]struct{}{"Int": {}, "Float": {}}

	cases := []struct {
		name      string
		requested string
		ports     map[string]struct{}
		want      string
	}{
		{"Float finds Float 64 by prefix", "Float", half, "Float 64"},
		{"Float on a node whose main output is Out", "Float", negate, "Out"},
		{"Float 64 on a node whose main output is Out", "Float 64", negate, "Out"},
		{"Out on a node whose main output is Float 64", "Out", half, "Float 64"},
		{"a specific name is never swapped for a generic one", "Mesh", negate, "Mesh"},
		{"Out never matches a specific name", "Out", map[string]struct{}{"Union": {}}, "Out"},
		{"exact still wins", "Float", add, "Float"},
		{"ambiguous prefix passes through", "I", map[string]struct{}{"Int": {}, "Index": {}}, "I"},
		{"unknown on a multi-output node passes through", "Mesh", half, "Mesh"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := resolveOutputPort(c.requested, c.ports); got != c.want {
				t.Errorf("resolveOutputPort(%q) = %q, want %q", c.requested, got, c.want)
			}
		})
	}
}
