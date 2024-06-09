package ply_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/stretchr/testify/assert"
)

func TestHeaderTextureFiles(t *testing.T) {
	tests := map[string]struct {
		input ply.Header
		want  []string
	}{
		"no textures": {
			input: ply.Header{
				Comments: []string{
					"test",
				},
			},
			want: []string{},
		},
		"single tex": {
			input: ply.Header{
				Comments: []string{
					"texturefile a.png",
				},
			},
			want: []string{
				"a.png",
			},
		},
		"multiple textures": {
			input: ply.Header{
				Comments: []string{
					"something",
					"texturefile a.png",
					"other something",
					"TEXTUREFILE b.png",
					"TextureFile  with a  space.jpg",
				},
			},
			want: []string{
				"a.png",
				"b.png",
				"with a  space.jpg",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.input.TextureFiles())
		})
	}
}

func TestHeaderWrite(t *testing.T) {
	tests := map[string]struct {
		input ply.Header
		want  string
	}{
		"empty ascii": {
			input: ply.Header{
				Format: ply.ASCII,
			},
			want: `ply
format ascii 1.0
end_header
`,
		},
		"empty little endian": {
			input: ply.Header{
				Format: ply.BinaryLittleEndian,
			},
			want: `ply
format binary_little_endian 1.0
end_header
`,
		},
		"empty big endian": {
			input: ply.Header{
				Format: ply.BinaryBigEndian,
			},
			want: `ply
format binary_big_endian 1.0
end_header
`,
		},
		"obj_info": {
			input: ply.Header{
				Format: ply.ASCII,
				ObjInfo: []string{
					"test one two",
					"test three four",
				},
			},
			want: `ply
format ascii 1.0
obj_info test one two
obj_info test three four
end_header
`,
		},
		"comments": {
			input: ply.Header{
				Format: ply.ASCII,
				Comments: []string{
					"test one two",
					"test three four",
				},
			},
			want: `ply
format ascii 1.0
comment test one two
comment test three four
end_header
`,
		},
		"single element": {
			input: ply.Header{
				Format: ply.ASCII,
				Comments: []string{
					"Test Comment",
				},
				ObjInfo: []string{
					"Test OBJ",
				},
				Elements: []ply.Element{
					{
						Name:  "test",
						Count: 12345678,
						Properties: []ply.Property{
							ply.ScalarProperty{
								PropertyName: "foo",
								Type:         ply.UChar,
							},
							ply.ScalarProperty{
								PropertyName: "Bar",
								Type:         ply.Double,
							},
						},
					},
				},
			},
			want: `ply
format ascii 1.0
comment Test Comment
obj_info Test OBJ
element test 12345678
property uchar foo
property double Bar
end_header
`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			assert.NoError(t, tc.input.Write(buf))
			assert.Equal(t, tc.want, buf.String())
			assert.Equal(t, tc.want, string(tc.input.Bytes()))
		})
	}
}
