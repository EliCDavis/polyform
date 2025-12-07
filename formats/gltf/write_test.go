package gltf_test

import (
	"bytes"
	"image/color"
	"math"
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

type mockTextureExtension struct{}

func (m mockTextureExtension) ExtensionID() string {
	return "mocXtension"
}
func (m mockTextureExtension) ToTextureExtensionData(w *gltf.Writer) map[string]any {
	return nil
}
func (m mockTextureExtension) IsRequired() bool {
	return false
}
func (m mockTextureExtension) IsInfo() bool {
	return false
}

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
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: &tri,
				Material: &gltf.PolyformMaterial{
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.White,
					},
				},
			},
		},
	}, &buf, nil)

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
            "mesh": 0,
            "name": "mesh"
        }
    ],
    "scene": 0,
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
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: &tri,
			},
		},
	}, &buf, nil)

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
            "mesh": 0,
            "name": "mesh"
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0
            ]
        }
    ]
}`, buf.String())
}

func TestWrite_TexturedTriWithMaterialWithColor(t *testing.T) {
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
	roughness := 0.
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: &tri,
				Material: &gltf.PolyformMaterial{
					Name: "My Material",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor:  color.RGBA{255, 100, 80, 255},
						RoughnessFactor:  &roughness,
						BaseColorTexture: &gltf.PolyformTexture{URI: "this_is_a_test.png"},
					},
				},
			},
		},
	}, &buf, nil)

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
    "images": [
        {
            "uri": "this_is_a_test.png"
        }
    ],
    "materials": [
        {
            "name": "My Material",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "baseColorTexture": {
                    "index": 0
                },
                "roughnessFactor": 0
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
            "mesh": 0,
            "name": "mesh"
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0
            ]
        }
    ],
    "textures": [
        {
            "source": 0
        }
    ]
}`, buf.String())
}

func TestWrite_TexturedTriWithMaterialWithColor_ImageSampleDedupe(t *testing.T) {
	// ARRANGE ================================================================
	tri1 := modeling.NewTriangleMesh([]int{0, 1, 2}).
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
	tri2 := modeling.NewTriangleMesh([]int{0, 1, 2}).
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
	roughness := 0.
	sampler := &gltf.Sampler{
		WrapS:     gltf.SamplerWrap_REPEAT,
		WrapT:     gltf.SamplerWrap_REPEAT,
		MinFilter: gltf.SamplerMinFilter_LINEAR_MIPMAP_LINEAR,
		MagFilter: gltf.SamplerMagFilter_LINEAR,
	}
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: &tri1,
				Material: &gltf.PolyformMaterial{
					Name: "My Material1",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.RGBA{255, 100, 80, 255},
						RoughnessFactor: &roughness,
						BaseColorTexture: &gltf.PolyformTexture{
							URI:     "this_is_a_test.png",
							Sampler: sampler,
						},
					},
				},
			},
			{
				Name: "mesh",
				Mesh: &tri2,
				Material: &gltf.PolyformMaterial{
					Name: "My Material2",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.RGBA{255, 100, 80, 255},
						RoughnessFactor: &roughness,
						BaseColorTexture: &gltf.PolyformTexture{
							URI:        "this_is_a_test.png",
							Sampler:    sampler,
							Extensions: []gltf.TextureExtension{&mockTextureExtension{}},
						},
					},
				},
			},
		},
	}, &buf, nil)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `{
    "extensionsUsed": [
        "mocXtension"
    ],
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
        },
        {
            "bufferView": 4,
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
            "bufferView": 5,
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
            "bufferView": 6,
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
            "bufferView": 7,
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
            "byteLength": 206,
            "uri": "data:application/octet-stream;base64,AACAPwAAAAAAAAAAAAAAAAAAgD8AAAAAAAAAAAAAAAAAAIA/AAAAAAAAAAAAAAAAAAAAAAAAgD8AAAAAAACAPwAAAAAAAAAAAAAAAAAAAAAAAAAAAACAPwAAgD8AAAAAAAABAAIAAAAAAIA/AAAAAAAAAAAAAAAAAACAPwAAAAAAAAAAAAAAAAAAgD8AAAAAAAAAAAAAAAAAAAAAAACAPwAAAAAAAIA/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAIA/AACAPwAAAAAAAAEAAgA="
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
        },
        {
            "buffer": 0,
            "byteOffset": 104,
            "byteLength": 36,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 140,
            "byteLength": 36,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 176,
            "byteLength": 24,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 200,
            "byteLength": 6,
            "target": 34963
        }
    ],
    "images": [
        {
            "uri": "this_is_a_test.png"
        }
    ],
    "materials": [
        {
            "name": "My Material1",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "baseColorTexture": {
                    "index": 0
                },
                "roughnessFactor": 0
            }
        },
        {
            "name": "My Material2",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "baseColorTexture": {
                    "index": 1
                },
                "roughnessFactor": 0
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
        },
        {
            "name": "mesh",
            "primitives": [
                {
                    "attributes": {
                        "NORMAL": 4,
                        "POSITION": 5,
                        "TEXCOORD_0": 6
                    },
                    "indices": 7,
                    "material": 1
                }
            ]
        }
    ],
    "nodes": [
        {
            "mesh": 0,
            "name": "mesh"
        },
        {
            "mesh": 1,
            "name": "mesh"
        }
    ],
    "samplers": [
        {
            "magFilter": 9729,
            "minFilter": 9987,
            "wrapS": 10497,
            "wrapT": 10497
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0,
                1
            ]
        }
    ],
    "textures": [
        {
            "sampler": 0,
            "source": 0
        },
        {
            "sampler": 0,
            "source": 0,
            "extensions": {
                "mocXtension": null
            }
        }
    ]
}`, buf.String())
}

func TestWrite_TexturedTriWithMaterialWithColor_TextureValueDedupe(t *testing.T) {
	// ARRANGE ================================================================
	tri1 := modeling.NewTriangleMesh([]int{0, 1, 2}).
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
	tri2 := modeling.NewTriangleMesh([]int{0, 1, 2}).
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
	roughness := 0.
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: &tri1,
				Material: &gltf.PolyformMaterial{
					Name: "My Material1",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.RGBA{255, 100, 80, 255},
						RoughnessFactor: &roughness,
						BaseColorTexture: &gltf.PolyformTexture{
							URI: "this_is_a_test.png",
							Sampler: &gltf.Sampler{
								WrapS:     gltf.SamplerWrap_REPEAT,
								WrapT:     gltf.SamplerWrap_REPEAT,
								MinFilter: gltf.SamplerMinFilter_LINEAR_MIPMAP_LINEAR,
								MagFilter: gltf.SamplerMagFilter_LINEAR,
							},
						},
					},
				},
			},
			{
				Name: "mesh",
				Mesh: &tri2,
				Material: &gltf.PolyformMaterial{
					Name: "My Material2",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.RGBA{255, 100, 80, 255},
						RoughnessFactor: &roughness,
						BaseColorTexture: &gltf.PolyformTexture{
							URI: "this_is_a_test.png",
							Sampler: &gltf.Sampler{
								WrapS:     gltf.SamplerWrap_REPEAT,
								WrapT:     gltf.SamplerWrap_REPEAT,
								MinFilter: gltf.SamplerMinFilter_LINEAR_MIPMAP_LINEAR,
								MagFilter: gltf.SamplerMagFilter_LINEAR,
							},
						},
					},
				},
			},
		},
	}, &buf, nil)

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
        },
        {
            "bufferView": 4,
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
            "bufferView": 5,
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
            "bufferView": 6,
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
            "bufferView": 7,
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
            "byteLength": 206,
            "uri": "data:application/octet-stream;base64,AACAPwAAAAAAAAAAAAAAAAAAgD8AAAAAAAAAAAAAAAAAAIA/AAAAAAAAAAAAAAAAAAAAAAAAgD8AAAAAAACAPwAAAAAAAAAAAAAAAAAAAAAAAAAAAACAPwAAgD8AAAAAAAABAAIAAAAAAIA/AAAAAAAAAAAAAAAAAACAPwAAAAAAAAAAAAAAAAAAgD8AAAAAAAAAAAAAAAAAAAAAAACAPwAAAAAAAIA/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAIA/AACAPwAAAAAAAAEAAgA="
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
        },
        {
            "buffer": 0,
            "byteOffset": 104,
            "byteLength": 36,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 140,
            "byteLength": 36,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 176,
            "byteLength": 24,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 200,
            "byteLength": 6,
            "target": 34963
        }
    ],
    "images": [
        {
            "uri": "this_is_a_test.png"
        }
    ],
    "materials": [
        {
            "name": "My Material1",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "baseColorTexture": {
                    "index": 0
                },
                "roughnessFactor": 0
            }
        },
        {
            "name": "My Material2",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "baseColorTexture": {
                    "index": 0
                },
                "roughnessFactor": 0
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
        },
        {
            "name": "mesh",
            "primitives": [
                {
                    "attributes": {
                        "NORMAL": 4,
                        "POSITION": 5,
                        "TEXCOORD_0": 6
                    },
                    "indices": 7,
                    "material": 1
                }
            ]
        }
    ],
    "nodes": [
        {
            "mesh": 0,
            "name": "mesh"
        },
        {
            "mesh": 1,
            "name": "mesh"
        }
    ],
    "samplers": [
        {
            "magFilter": 9729,
            "minFilter": 9987,
            "wrapS": 10497,
            "wrapT": 10497
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0,
                1
            ]
        }
    ],
    "textures": [
        {
            "sampler": 0,
            "source": 0
        }
    ]
}`, buf.String())
}

func TestWrite_TexturedTriWithMaterialWithColor_NormalOcclusion(t *testing.T) {
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
	roughness := 0.
	scale := 1.1
	strength := 0.1
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: &tri,
				Material: &gltf.PolyformMaterial{
					Name: "My Material",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor:  color.RGBA{255, 100, 80, 255},
						RoughnessFactor:  &roughness,
						BaseColorTexture: &gltf.PolyformTexture{URI: "this_is_a_test.png"},
					},
					NormalTexture: &gltf.PolyformNormal{
						PolyformTexture: &gltf.PolyformTexture{URI: "this_is_a_normal_test.png"},
						Scale:           &scale,
					},
					OcclusionTexture: &gltf.PolyformOcclusion{
						PolyformTexture: &gltf.PolyformTexture{URI: "this_is_an_occlusion_test.png"},
						Strength:        &strength,
					},
				},
			},
		},
	}, &buf, nil)

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
    "images": [
        {
            "uri": "this_is_a_test.png"
        },
        {
            "uri": "this_is_a_normal_test.png"
        },
        {
            "uri": "this_is_an_occlusion_test.png"
        }
    ],
    "materials": [
        {
            "name": "My Material",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "baseColorTexture": {
                    "index": 0
                },
                "roughnessFactor": 0
            },
            "normalTexture": {
                "index": 1,
                "scale": 1.1
            },
            "occlusionTexture": {
                "index": 2,
                "strength": 0.1
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
            "mesh": 0,
            "name": "mesh"
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0
            ]
        }
    ],
    "textures": [
        {
            "source": 0
        },
        {
            "source": 1
        },
        {
            "source": 2
        }
    ]
}`, buf.String())

}

func TestWrite_TexturedTriWithTexExtension(t *testing.T) {
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
	roughness := 0.
	textureOffset := vector2.New[float64](1.1, 0.1)
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: &tri,
				Material: &gltf.PolyformMaterial{
					Name: "My Material",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.RGBA{255, 100, 80, 255},
						RoughnessFactor: &roughness,
						BaseColorTexture: &gltf.PolyformTexture{
							URI: "this_is_a_test.png",
							Sampler: &gltf.Sampler{
								WrapS:     gltf.SamplerWrap_REPEAT,
								WrapT:     gltf.SamplerWrap_REPEAT,
								MinFilter: gltf.SamplerMinFilter_LINEAR_MIPMAP_LINEAR,
								MagFilter: gltf.SamplerMagFilter_LINEAR,
							},
							Extensions: []gltf.TextureExtension{gltf.PolyformTextureTransform{
								Offset: &textureOffset,
							}},
						},
					},
				},
			},
		},
	}, &buf, nil)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `{
    "extensionsUsed": [
        "KHR_texture_transform"
    ],
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
    "images": [
        {
            "uri": "this_is_a_test.png"
        }
    ],
    "materials": [
        {
            "name": "My Material",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "baseColorTexture": {
                    "index": 0,
                    "extensions": {
                        "KHR_texture_transform": {
                            "offset": [
                                1.1,
                                0.1
                            ]
                        }
                    }
                },
                "roughnessFactor": 0
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
            "mesh": 0,
            "name": "mesh"
        }
    ],
    "samplers": [
        {
            "magFilter": 9729,
            "minFilter": 9987,
            "wrapS": 10497,
            "wrapT": 10497
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0
            ]
        }
    ],
    "textures": [
        {
            "sampler": 0,
            "source": 0
        }
    ]
}`, buf.String())
}

func TestWrite_TexturedTriWithTexExtension_Required(t *testing.T) {
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
	roughness := 0.
	textureOffset := vector2.New[float64](1.1, 0.1)
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: &tri,
				Material: &gltf.PolyformMaterial{
					Name: "My Material",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.RGBA{255, 100, 80, 255},
						RoughnessFactor: &roughness,
						BaseColorTexture: &gltf.PolyformTexture{
							URI: "this_is_a_test.png",
							Sampler: &gltf.Sampler{
								WrapS:     gltf.SamplerWrap_REPEAT,
								WrapT:     gltf.SamplerWrap_REPEAT,
								MinFilter: gltf.SamplerMinFilter_LINEAR_MIPMAP_LINEAR,
								MagFilter: gltf.SamplerMagFilter_LINEAR,
							},
							Extensions: []gltf.TextureExtension{gltf.PolyformTextureTransform{
								Required: true,
								Offset:   &textureOffset,
							}},
						},
					},
				},
			},
		},
	}, &buf, nil)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `{
    "extensionsUsed": [
        "KHR_texture_transform"
    ],
    "extensionsRequired": [
        "KHR_texture_transform"
    ],
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
    "images": [
        {
            "uri": "this_is_a_test.png"
        }
    ],
    "materials": [
        {
            "name": "My Material",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "baseColorTexture": {
                    "index": 0,
                    "extensions": {
                        "KHR_texture_transform": {
                            "offset": [
                                1.1,
                                0.1
                            ]
                        }
                    }
                },
                "roughnessFactor": 0
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
            "mesh": 0,
            "name": "mesh"
        }
    ],
    "samplers": [
        {
            "magFilter": 9729,
            "minFilter": 9987,
            "wrapS": 10497,
            "wrapT": 10497
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0
            ]
        }
    ],
    "textures": [
        {
            "sampler": 0,
            "source": 0
        }
    ]
}`, buf.String())
}

func TestWrite_MaterialAlphaMode(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute,
			[]vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		)
	buf := bytes.Buffer{}

	// ACT ====================================================================
	roughness := 0.
	alphaBlend := gltf.MaterialAlphaMode_BLEND
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: &tri,
				Material: &gltf.PolyformMaterial{
					Name: "My Material",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.RGBA{255, 100, 80, 255},
						RoughnessFactor: &roughness,
					},
					AlphaMode: &alphaBlend,
				},
			},
		},
	}, &buf, nil)

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
            "name": "My Material",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "roughnessFactor": 0
            },
            "alphaMode": "BLEND"
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
            "mesh": 0,
            "name": "mesh"
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0
            ]
        }
    ]
}`, buf.String())
}

func TestWrite_MaterialAlphaModeWithCutOff(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute,
			[]vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		)

	buf := bytes.Buffer{}

	// ACT ====================================================================
	roughness := 0.
	alphaMode := gltf.MaterialAlphaMode_MASK
	alphaCutOff := 0.8
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: &tri,
				Material: &gltf.PolyformMaterial{
					Name: "My Material",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.RGBA{255, 100, 80, 255},
						RoughnessFactor: &roughness,
					},
					AlphaMode:   &alphaMode,
					AlphaCutoff: &alphaCutOff,
				},
			},
		},
	}, &buf, nil)

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
            "name": "My Material",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "roughnessFactor": 0
            },
            "alphaMode": "MASK",
            "alphaCutoff": 0.8
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
            "mesh": 0,
            "name": "mesh"
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0
            ]
        }
    ]
}`, buf.String())

}

func TestWrite_MaterialAlphaCutOffError(t *testing.T) {
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
	roughness := 0.
	alphaCutOff := 0.8
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: &tri,
				Material: &gltf.PolyformMaterial{
					Name: "My Material",
					PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
						BaseColorFactor: color.RGBA{255, 100, 80, 255},
						RoughnessFactor: &roughness,
					},
					AlphaCutoff: &alphaCutOff,
				},
			},
		},
	}, &buf, nil)

	// ASSERT =================================================================
	assert.Error(t, err)
}

func TestWrite_MaterialsDeduplicated(t *testing.T) {
	// ARRANGE ================================================================
	tri0 := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute,
			[]vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		)

	tri1 := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute,
			[]vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		)

	buf := bytes.Buffer{}

	// ACT ====================================================================
	roughness := 0.
	material := &gltf.PolyformMaterial{
		Name: "My Material",
		PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
			BaseColorFactor: color.RGBA{255, 100, 80, 255},
			RoughnessFactor: &roughness,
		},
	}
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{Name: "mesh0", Mesh: &tri0, Material: material},
			{Name: "mesh1", Mesh: &tri1, Material: material},
		},
	}, &buf, nil)

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
        },
        {
            "bufferView": 2,
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
            "byteLength": 86,
            "uri": "data:application/octet-stream;base64,AAAAAAAAAAAAAAAAAAAAAAAAgD8AAAAAAACAPwAAAAAAAAAAAAABAAIAAAAAAAAAAAAAAAAAAAAAAAAAAACAPwAAAAAAAIA/AAAAAAAAAAAAAAEAAgA="
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
        },
        {
            "buffer": 0,
            "byteOffset": 44,
            "byteLength": 36,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 80,
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
                    0.392,
                    0.314,
                    1
                ],
                "roughnessFactor": 0
            }
        }
    ],
    "meshes": [
        {
            "name": "mesh0",
            "primitives": [
                {
                    "attributes": {
                        "POSITION": 0
                    },
                    "indices": 1,
                    "material": 0
                }
            ]
        },
        {
            "name": "mesh1",
            "primitives": [
                {
                    "attributes": {
                        "POSITION": 2
                    },
                    "indices": 3,
                    "material": 0
                }
            ]
        }
    ],
    "nodes": [
        {
            "mesh": 0,
            "name": "mesh0"
        },
        {
            "mesh": 1,
            "name": "mesh1"
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0,
                1
            ]
        }
    ]
}`, buf.String())
}

func TestWrite_MeshesDeduplicated(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute,
			[]vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		)

	buf := bytes.Buffer{}

	// ACT ====================================================================
	roughness := 0.
	material := &gltf.PolyformMaterial{
		Name: "My Material",
		PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
			BaseColorFactor: color.RGBA{255, 100, 80, 255},
			RoughnessFactor: &roughness,
		},
	}
	rightV := vector3.New[float64](20, 0, 0)
	leftV := vector3.New[float64](-20, 0, -0)
	scaleUniform15 := vector3.New[float64](1.5, 1.5, 1.5)
	scaleDistort := vector3.New[float64](0.5, 2.5, 0.5)
	rotQuat := quaternion.FromTheta(-math.Pi/2, vector3.New[float64](1, 0, 0))

	trsRight := trs.New(rightV, quaternion.Identity(), scaleUniform15)
	trsLeft := trs.New(leftV, rotQuat, scaleDistort)

	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{Name: "mesh_right", Mesh: &tri, Material: material, TRS: &trsRight},
			{Name: "mesh_left", Mesh: &tri, Material: material, TRS: &trsLeft},
		},
	}, &buf, nil)

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
            "name": "My Material",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "roughnessFactor": 0
            }
        }
    ],
    "meshes": [
        {
            "name": "mesh_right",
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
            "mesh": 0,
            "scale": [
                1.5,
                1.5,
                1.5
            ],
            "translation": [
                20,
                0,
                0
            ],
            "name": "mesh_right"
        },
        {
            "mesh": 0,
            "rotation": [
                -0.7071067811865475,
                -0,
                -0,
                0.7071067811865476
            ],
            "scale": [
                0.5,
                2.5,
                0.5
            ],
            "translation": [
                -20,
                0,
                0
            ],
            "name": "mesh_left"
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0,
                1
            ]
        }
    ]
}`, buf.String())
}

func TestWrite_MeshesGpuInstanced(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute,
			[]vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		)

	buf := bytes.Buffer{}

	// ACT ====================================================================
	roughness := 0.
	material := &gltf.PolyformMaterial{
		Name: "My Material",
		PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
			BaseColorFactor: color.RGBA{255, 100, 80, 255},
			RoughnessFactor: &roughness,
		},
	}
	rightV := vector3.New[float64](2, 0, 0)
	leftV := vector3.New[float64](-2, 0, -0)
	scaleUniform15 := vector3.New[float64](1.5, 1.5, 1.5)
	scaleDistort := vector3.New[float64](0.5, 2.5, 0.5)
	rotQuat := quaternion.FromTheta(-math.Pi/2, vector3.New[float64](1, 0, 0))

	trsRight := trs.New(rightV, quaternion.Identity(), scaleUniform15)
	trsLeft := trs.New(leftV, rotQuat, scaleDistort)

	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{Name: "mesh_right", Mesh: &tri, Material: material, TRS: &trsRight},
			{Name: "mesh_left", Mesh: &tri, Material: material, TRS: &trsLeft},
		},
	}, &buf, &gltf.WriterOptions{
		GpuInstancingStrategy: gltf.WriterInstancingStrategy_Collapse,
	})

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `{
    "extensionsUsed": [
        "EXT_mesh_gpu_instancing"
    ],
    "extensionsRequired": [
        "EXT_mesh_gpu_instancing"
    ],
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
        },
        {
            "bufferView": 2,
            "componentType": 5126,
            "type": "VEC3",
            "count": 2,
            "max": [
                2,
                0,
                0
            ],
            "min": [
                -2,
                0,
                0
            ]
        },
        {
            "bufferView": 3,
            "componentType": 5126,
            "type": "VEC3",
            "count": 2,
            "max": [
                1.5,
                2.5,
                1.5
            ],
            "min": [
                0.5,
                1.5,
                0.5
            ]
        },
        {
            "bufferView": 4,
            "componentType": 5126,
            "type": "VEC4",
            "count": 2,
            "max": [
                0,
                0,
                0,
                1
            ],
            "min": [
                -0.7071067811865475,
                -0,
                -0,
                0.7071067811865476
            ]
        }
    ],
    "asset": {
        "version": "2.0",
        "generator": "https://github.com/EliCDavis/polyform"
    },
    "buffers": [
        {
            "byteLength": 124,
            "uri": "data:application/octet-stream;base64,AAAAAAAAAAAAAAAAAAAAAAAAgD8AAAAAAACAPwAAAAAAAAAAAAABAAIAAAAAAABAAAAAAAAAAAAAAADAAAAAAAAAAAAAAMA/AADAPwAAwD8AAAA/AAAgQAAAAD8AAAAAAAAAAAAAAAAAAIA/8wQ1vwAAAIAAAACA8wQ1Pw=="
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
        },
        {
            "buffer": 0,
            "byteOffset": 44,
            "byteLength": 24,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 68,
            "byteLength": 24,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 92,
            "byteLength": 32,
            "target": 34962
        }
    ],
    "materials": [
        {
            "name": "My Material",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "roughnessFactor": 0
            }
        }
    ],
    "meshes": [
        {
            "name": "mesh_right",
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
            "extensions": {
                "EXT_mesh_gpu_instancing": {
                    "attributes": {
                        "ROTATION": 4,
                        "SCALE": 3,
                        "TRANSLATION": 2
                    }
                }
            },
            "mesh": 0,
            "name": "Instances"
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0
            ]
        }
    ]
}`, buf.String())
}

func TestWrite_GpuInstancedMeshes_Expanded(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute,
			[]vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		)

	buf := bytes.Buffer{}

	// ACT ====================================================================
	roughness := 0.
	material := &gltf.PolyformMaterial{
		Name: "My Material",
		PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
			BaseColorFactor: color.RGBA{255, 100, 80, 255},
			RoughnessFactor: &roughness,
		},
	}
	trsRight := trs.New(
		vector3.New(2., 0, 0),
		quaternion.Identity(),
		vector3.New(1.5, 1.5, 1.5),
	)
	trsLeft := trs.New(
		vector3.New(-2., 0, -0),
		quaternion.FromTheta(-math.Pi/2, vector3.New(1., 0, 0)),
		vector3.New(0.5, 2.5, 0.5),
	)

	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{
				Name: "parent",
				Mesh: &tri, Material: material,
				GpuInstances: []trs.TRS{
					trsRight,
					trsLeft,
				},
			},
		},
	}, &buf, &gltf.WriterOptions{
		GpuInstancingStrategy: gltf.WriterInstancingStrategy_Expand,
	})

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
            "name": "My Material",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "roughnessFactor": 0
            }
        }
    ],
    "meshes": [
        {
            "name": "parent",
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
            "mesh": 0,
            "scale": [
                1.5,
                1.5,
                1.5
            ],
            "translation": [
                2,
                0,
                0
            ],
            "name": "parent"
        },
        {
            "mesh": 0,
            "rotation": [
                -0.7071067811865475,
                0,
                0,
                0.7071067811865476
            ],
            "scale": [
                0.5,
                2.5,
                0.5
            ],
            "translation": [
                -2,
                0,
                0
            ],
            "name": "parent"
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0,
                1
            ]
        }
    ]
}`, buf.String())
}

func TestWrite_MeshesGpuInstanced_UnderChild(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute,
			[]vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		)

	buf := bytes.Buffer{}

	// ACT ====================================================================
	roughness := 0.
	material := &gltf.PolyformMaterial{
		Name: "My Material",
		PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
			BaseColorFactor: color.RGBA{255, 100, 80, 255},
			RoughnessFactor: &roughness,
		},
	}
	rightV := vector3.New[float64](2, 0, 0)
	leftV := vector3.New[float64](-2, 0, -0)
	scaleUniform15 := vector3.New[float64](1.5, 1.5, 1.5)
	scaleDistort := vector3.New[float64](0.5, 2.5, 0.5)
	rotQuat := quaternion.FromTheta(-math.Pi/2, vector3.New[float64](1, 0, 0))

	trsRight := trs.New(rightV, quaternion.Identity(), scaleUniform15)
	trsLeft := trs.New(leftV, rotQuat, scaleDistort)

	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{Name: "parent", Children: []*gltf.PolyformModel{
				{Name: "mesh_right", Mesh: &tri, Material: material, TRS: &trsRight},
				{Name: "mesh_left", Mesh: &tri, Material: material, TRS: &trsLeft},
			}},
		},
	}, &buf, &gltf.WriterOptions{
		GpuInstancingStrategy: gltf.WriterInstancingStrategy_Collapse,
	})

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `{
    "extensionsUsed": [
        "EXT_mesh_gpu_instancing"
    ],
    "extensionsRequired": [
        "EXT_mesh_gpu_instancing"
    ],
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
        },
        {
            "bufferView": 2,
            "componentType": 5126,
            "type": "VEC3",
            "count": 2,
            "max": [
                2,
                0,
                0
            ],
            "min": [
                -2,
                0,
                0
            ]
        },
        {
            "bufferView": 3,
            "componentType": 5126,
            "type": "VEC3",
            "count": 2,
            "max": [
                1.5,
                2.5,
                1.5
            ],
            "min": [
                0.5,
                1.5,
                0.5
            ]
        },
        {
            "bufferView": 4,
            "componentType": 5126,
            "type": "VEC4",
            "count": 2,
            "max": [
                0,
                0,
                0,
                1
            ],
            "min": [
                -0.7071067811865475,
                -0,
                -0,
                0.7071067811865476
            ]
        }
    ],
    "asset": {
        "version": "2.0",
        "generator": "https://github.com/EliCDavis/polyform"
    },
    "buffers": [
        {
            "byteLength": 124,
            "uri": "data:application/octet-stream;base64,AAAAAAAAAAAAAAAAAAAAAAAAgD8AAAAAAACAPwAAAAAAAAAAAAABAAIAAAAAAABAAAAAAAAAAAAAAADAAAAAAAAAAAAAAMA/AADAPwAAwD8AAAA/AAAgQAAAAD8AAAAAAAAAAAAAAAAAAIA/8wQ1vwAAAIAAAACA8wQ1Pw=="
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
        },
        {
            "buffer": 0,
            "byteOffset": 44,
            "byteLength": 24,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 68,
            "byteLength": 24,
            "target": 34962
        },
        {
            "buffer": 0,
            "byteOffset": 92,
            "byteLength": 32,
            "target": 34962
        }
    ],
    "materials": [
        {
            "name": "My Material",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "roughnessFactor": 0
            }
        }
    ],
    "meshes": [
        {
            "name": "mesh_right",
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
            "extensions": {
                "EXT_mesh_gpu_instancing": {
                    "attributes": {
                        "ROTATION": 4,
                        "SCALE": 3,
                        "TRANSLATION": 2
                    }
                }
            },
            "mesh": 0,
            "name": "Instances"
        },
        {
            "children": [
                0
            ],
            "name": "parent"
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                1
            ]
        }
    ]
}`, buf.String())
}

func TestWrite_MeshesDifferentMatsPreserved(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewTriangleMesh([]int{0, 1, 2}).
		SetFloat3Attribute(
			modeling.PositionAttribute,
			[]vector3.Float64{
				vector3.New(0., 0., 0.),
				vector3.New(0., 1., 0.),
				vector3.New(1., 0., 0.),
			},
		)

	buf := bytes.Buffer{}

	// ACT ====================================================================
	roughness := 0.
	materialLeft := &gltf.PolyformMaterial{
		Name: "Material Left",
		PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
			BaseColorFactor: color.RGBA{255, 100, 80, 255},
			RoughnessFactor: &roughness,
		},
	}
	materialRight := &gltf.PolyformMaterial{
		Name: "Material Right",
		PbrMetallicRoughness: &gltf.PolyformPbrMetallicRoughness{
			BaseColorFactor: color.RGBA{100, 255, 80, 255},
			RoughnessFactor: &roughness,
		},
	}

	rightV := vector3.New[float64](20, 0, 0)
	leftV := vector3.New[float64](-20, 0, -0)
	scaleUniform15 := vector3.New[float64](1.5, 1.5, 1.5)
	scaleDistort := vector3.New[float64](0.5, 2.5, 0.5)
	rotQuat := quaternion.FromTheta(-math.Pi/2, vector3.New[float64](1, 0, 0))

	trsRight := trs.New(rightV, quaternion.Identity(), scaleUniform15)
	trsLeft := trs.New(leftV, rotQuat, scaleDistort)

	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{Name: "mesh_right", Mesh: &tri, Material: materialLeft, TRS: &trsRight},
			{Name: "mesh_left", Mesh: &tri, Material: materialRight, TRS: &trsLeft},
		},
	}, &buf, nil)

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
            "name": "Material Left",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    1,
                    0.392,
                    0.314,
                    1
                ],
                "roughnessFactor": 0
            }
        },
        {
            "name": "Material Right",
            "pbrMetallicRoughness": {
                "baseColorFactor": [
                    0.392,
                    1,
                    0.314,
                    1
                ],
                "roughnessFactor": 0
            }
        }
    ],
    "meshes": [
        {
            "name": "mesh_right",
            "primitives": [
                {
                    "attributes": {
                        "POSITION": 0
                    },
                    "indices": 1,
                    "material": 0
                }
            ]
        },
        {
            "name": "mesh_left",
            "primitives": [
                {
                    "attributes": {
                        "POSITION": 0
                    },
                    "indices": 1,
                    "material": 1
                }
            ]
        }
    ],
    "nodes": [
        {
            "mesh": 0,
            "scale": [
                1.5,
                1.5,
                1.5
            ],
            "translation": [
                20,
                0,
                0
            ],
            "name": "mesh_right"
        },
        {
            "mesh": 1,
            "rotation": [
                -0.7071067811865475,
                -0,
                -0,
                0.7071067811865476
            ],
            "scale": [
                0.5,
                2.5,
                0.5
            ],
            "translation": [
                -20,
                0,
                0
            ],
            "name": "mesh_left"
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0,
                1
            ]
        }
    ]
}`, buf.String())
}

func TestWriteEmptyMesh(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.EmptyMesh(modeling.TriangleTopology)
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Mesh: &tri,
			},
		},
	}, &buf, nil)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `{
    "asset": {
        "version": "2.0",
        "generator": "https://github.com/EliCDavis/polyform"
    },
    "nodes": [
        {
            "name": "mesh"
        }
    ],
    "scene": 0,
    "scenes": [
        {
            "nodes": [
                0
            ]
        }
    ]
}`, buf.String())
}

func aTestWrite_NestedModels(t *testing.T) {
	// ARRANGE ================================================================
	tri := modeling.NewPointCloud(
		nil,
		map[string][]vector3.Float64{
			modeling.PositionAttribute: {
				vector3.New(0., 0, 0),
			},
		},
		nil,
		nil,
	)
	buf := bytes.Buffer{}

	// ACT ====================================================================
	err := gltf.WriteText(gltf.PolyformScene{
		Models: []*gltf.PolyformModel{
			{
				Name: "mesh",
				Children: []*gltf.PolyformModel{
					{
						Name: "child",
						Mesh: &tri,
					},
				},
			},
		},
	}, &buf, nil)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `{
    "asset": {
        "version": "2.0",
        "generator": "https://github.com/EliCDavis/polyform",
        "nodes": [
            [1],
            [{something}]
        ]
    }
}`, buf.String())
}
