package gltf_test

import (
	"bytes"
	"image/color"
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestWriteBasicTri(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute, []vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		)
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: tri,
				Material: &gltf.PolyformMaterial{
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.White,
					},
				},
			},
		},
	}, &buf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `{
    "accessors": [
        {
            "bufferView": 0,
            "componentType": 5126,
            "type": "VEC3",
            "count": 3,
            "max": [
                1,
                1,
                0
            ],
            "min": [
                0,
                0,
                0
            ]
        },
        {
            "bufferView": 1,
            "componentType": 5123,
            "type": "SCALAR",
            "count": 3
        }
    ],
    "asset": {
        "version": "2.0",
        "generator": "https://github.com/EliCDavis/polyform"
    },
    "buffers": [
        {
            "byteLength": 42,
            "uri": "data:application/octet-stream;base64,AAAAAAAAAAAAAAAAAAAAAAAAgD8AAAAAAACAPwAAAAAAAAAAAAABAAIA"
        }
    ],
    "bufferViews": [
        {
            "buffer": 0,
            "byteLength": 36,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 36,
            "byteLength": 6,
            "target": 34963
        }
    ],
    "materials": [
        {
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    1,
                    1,
                    1
                ]
            }
        }
    ],
    "meshes": [
        {
            "name": "mesh",
            "primitives": [
                {
                    "attributes": {
                        "POSITION": 0
                    },
                    "indices": 1,
                    "material": 0
                }
            ]
        }
    ],
    "nodes": [
        {
            "mesh": 0
        }
    ],
    "scenes": [
        {
            "nodes": [
                0
            ]
        }
    ]
}`, buf.String())
}

func TestWriteColorTri(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute, []vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		).
		SetFloat3Attribute(
			modeling.ColorAttribute, []vector3.Float64{
				vector3.New(1., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(0., 0., 1.),
			},
		)

	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := gltf.WriteText(gltf.PolyformScene{
		[]gltf.PolyformModel{
			gltf.PolyformModel{
				Name: "mesh",
				Mesh: tri,
			},
		},
	}, &buf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `{
    "accessors": [
        {
            "bufferView": 0,
            "componentType": 5126,
            "type": "VEC3",
            "count": 3,
            "max": [
                1,
                1,
                1
            ],
            "min": [
                0,
                0,
                0
            ]
        },
        {
            "bufferView": 1,
            "componentType": 5126,
            "type": "VEC3",
            "count": 3,
            "max": [
                1,
                1,
                0
            ],
            "min": [
                0,
                0,
                0
            ]
        },
        {
            "bufferView": 2,
            "componentType": 5123,
            "type": "SCALAR",
            "count": 3
        }
    ],
    "asset": {
        "version": "2.0",
        "generator": "https://github.com/EliCDavis/polyform"
    },
    "buffers": [
        {
            "byteLength": 78,
            "uri": "data:application/octet-stream;base64,AACAPwAAAAAAAAAAAAAAAAAAgD8AAAAAAAAAAAAAAAAAAIA/AAAAAAAAAAAAAAAAAAAAAAAAgD8AAAAAAACAPwAAAAAAAAAAAAABAAIA"
        }
    ],
    "bufferViews": [
        {
            "buffer": 0,
            "byteLength": 36,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 36,
            "byteLength": 36,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 72,
            "byteLength": 6,
            "target": 34963
        }
    ],
    "meshes": [
        {
            "name": "mesh",
            "primitives": [
                {
                    "attributes": {
                        "COLOR_0": 0,
                        "POSITION": 1
                    },
                    "indices": 2
                }
            ]
        }
    ],
    "nodes": [
        {
            "mesh": 0
        }
    ],
    "scenes": [
        {
            "nodes": [
                0
            ]
        }
    ]
}`, buf.String())
}

func TestWriteTexturedTriWithMaterialWithColor(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute,
			[]vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		).
		SetFloat3Attribute(
			modeling.NormalAttribute,
			[]vector3.Float64{
				vector3.New(1., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(0., 0., 1.),
			},
		).
		SetFloat2Attribute(
			modeling.TexCoordAttribute,
			[]vector2.Float64{
				vector2.New(0., 0.),
				vector2.New(0., 1.),
				vector2.New(1., 0.),
			},
		)

	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := gltf.WriteText(gltf.PolyformScene{
		[]gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: tri,
				Material: &gltf.PolyformMaterial{
					Name: "My Material",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.RGBA{255, 100, 80, 255},
					},
				},
			},
		},
	}, &buf)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `{
    "accessors": [
        {
            "bufferView": 0,
            "componentType": 5126,
            "type": "VEC3",
            "count": 3,
            "max": [
                1,
                1,
                1
            ],
            "min": [
                0,
                0,
                0
            ]
        },
        {
            "bufferView": 1,
            "componentType": 5126,
            "type": "VEC3",
            "count": 3,
            "max": [
                1,
                1,
                0
            ],
            "min": [
                0,
                0,
                0
            ]
        },
        {
            "bufferView": 2,
            "componentType": 5126,
            "type": "VEC2",
            "count": 3,
            "max": [
                1,
                1
            ],
            "min": [
                0,
                0
            ]
        },
        {
            "bufferView": 3,
            "componentType": 5123,
            "type": "SCALAR",
            "count": 3
        }
    ],
    "asset": {
        "version": "2.0",
        "generator": "https://github.com/EliCDavis/polyform"
    },
    "buffers": [
        {
            "byteLength": 102,
            "uri": "data:application/octet-stream;base64,AACAPwAAAAAAAAAAAAAAAAAAgD8AAAAAAAAAAAAAAAAAAIA/AAAAAAAAAAAAAAAAAAAAAAAAgD8AAAAAAACAPwAAAAAAAAAAAAAAAAAAAAAAAAAAAACAPwAAgD8AAAAAAAABAAIA"
        }
    ],
    "bufferViews": [
        {
            "buffer": 0,
            "byteLength": 36,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 36,
            "byteLength": 36,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 72,
            "byteLength": 24,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 96,
            "byteLength": 6,
            "target": 34963
        }
    ],
    "materials": [
        {
            "name": "My Material",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.39215686274509803,
                    0.3137254901960784,
                    1
                ]
            }
        }
    ],
    "meshes": [
        {
            "name": "mesh",
            "primitives": [
                {
                    "attributes": {
                        "NORMAL": 0,
                        "POSITION": 1,
                        "TEXCOORD_0": 2
                    },
                    "indices": 3,
                    "material": 0
                }
            ]
        }
    ],
    "nodes": [
        {
            "mesh": 0
        }
    ],
    "scenes": [
        {
            "nodes": [
                0
            ]
        }
    ]
}`, buf.String())
}
