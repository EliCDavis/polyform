let FILE = {
    "buffers": [
        {
            "byteLength": 0,
            "uri": "data:application/octet-stream;base64,"
        }
    ],
    "data": {
        "authors": [
            {
                "name": "Eli C Davis",
                "contactInfo": [
                    {
                        "medium": "bsky.app",
                        "value": "@elicdavis.bsky.social"
                    },
                    {
                        "medium": "github.com",
                        "value": "EliCDavis"
                    }
                ]
            }
        ],
        "description": "Immutable mesh processing program",
        "metadata": {
            "nodes": {
                "Node-0": {
                    "position": {
                        "x": 1567.4258954194406,
                        "y": 322.63335724847826
                    }
                },
                "Node-1": {
                    "position": {
                        "x": 837.7246414995578,
                        "y": 273.01783165449234
                    }
                },
                "Node-10": {
                    "position": {
                        "x": 1021.4335776975993,
                        "y": 437.7334269441334
                    }
                },
                "Node-11": {
                    "position": {
                        "x": -72.21064043753051,
                        "y": 814.1968953730265
                    }
                },
                "Node-12": {
                    "position": {
                        "x": -400.58549030660896,
                        "y": 793.3844048883665
                    }
                },
                "Node-13": {
                    "position": {
                        "x": 506.76915349605804,
                        "y": 398.84339970988816
                    }
                },
                "Node-2": {
                    "position": {
                        "x": 1267.197259560411,
                        "y": 289.65614653835723
                    }
                },
                "Node-3": {
                    "position": {
                        "x": 840,
                        "y": 104.47726440429688
                    }
                },
                "Node-4": {
                    "position": {
                        "x": 496.2982045268315,
                        "y": 135.89384852503375
                    }
                },
                "Node-5": {
                    "position": {
                        "x": 478.8344727350086,
                        "y": 264.2434048072408
                    }
                },
                "Node-7": {
                    "position": {
                        "x": 284.71445395426133,
                        "y": 617.9658341476188
                    }
                },
                "Node-8": {
                    "position": {
                        "x": -95.49429571427606,
                        "y": 563.0821443419094
                    }
                },
                "Node-9": {
                    "position": {
                        "x": 286.3052855428326,
                        "y": 480.35890173620214
                    }
                }
            },
            "notes": {
                "1": {
                    "position": {
                        "x": 21,
                        "y": 22.477264404296875
                    },
                    "text": "# Tutorial\n\nThis graph is meant to showcase and explain the different features of polyform",
                    "width": 677
                },
                "2": {
                    "position": {
                        "x": 1296,
                        "y": 48.477264404296875
                    },
                    "text": "# Artifacts\n\nThe output of the polyform graphs are called Artifacts, and their nodes are colored purple.\n\nTo render out artifacts, we must name them. You can do so by right-clicking on an artifact node and selecting: *Edit > Title*. \n\nBe sure to include the file extension at the end of the file name",
                    "width": 523
                },
                "3": {
                    "position": {
                        "x": -5.802989335091781,
                        "y": 182.39666491174742
                    },
                    "text": "# Parameters\n\nParameters are how users can provide input into the graph. All parameter nodes are colored green. \n\nIn order to create a parameter, right-click on the graph and select: *New Node > Polyform > Parameters > Your parameter type*.\n\nIf you want the parameter to be exposed to the command line or swagger, you can name it. You can name it by right-clicking the parameter node, and selecting *Edit > Title*.\n\nBe sure to give it a unique name!",
                    "width": 451.1423537224773
                }
            }
        },
        "name": "Polyform",
        "nodes": {
            "Node-0": {
                "type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/generator/artifact.Artifact,github.com/EliCDavis/polyform/formats/gltf.ArtifactNodeData]",
                "dependencies": [
                    {
                        "dependencyID": "Node-2",
                        "dependencyPort": "Out",
                        "name": "Models.0"
                    }
                ]
            },
            "Node-1": {
                "type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/modeling.Mesh,github.com/EliCDavis/polyform/modeling/primitives.CubeNodeData]",
                "dependencies": [
                    {
                        "dependencyID": "Node-13",
                        "dependencyPort": "Out",
                        "name": "Width"
                    }
                ]
            },
            "Node-10": {
                "type": "github.com/EliCDavis/polyform/nodes.Struct[[]github.com/EliCDavis/polyform/math/trs.TRS,github.com/EliCDavis/polyform/modeling/repeat.TRSNodeData]",
                "dependencies": [
                    {
                        "dependencyID": "Node-3",
                        "dependencyPort": "Out",
                        "name": "Input"
                    },
                    {
                        "dependencyID": "Node-6",
                        "dependencyPort": "Out",
                        "name": "Transforms"
                    }
                ]
            },
            "Node-11": {
                "type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/quaternion.Quaternion,github.com/EliCDavis/polyform/math/quaternion.FromEulerAnglesNodeData]",
                "dependencies": [
                    {
                        "dependencyID": "Node-12",
                        "dependencyPort": "Out",
                        "name": "Angles"
                    }
                ]
            },
            "Node-12": {
                "type": "github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/vector/vector3.Vector[float64]]",
                "data": {
                    "name": "",
                    "description": "",
                    "currentValue": {
                        "x": 0,
                        "y": 1,
                        "z": 0
                    },
                    "defaultValue": {
                        "x": 0,
                        "y": 0,
                        "z": 0
                    },
                    "cli": null
                }
            },
            "Node-13": {
                "type": "github.com/EliCDavis/polyform/generator/parameter.Value[float64]",
                "data": {
                    "name": "",
                    "description": "",
                    "currentValue": 2,
                    "defaultValue": 0,
                    "cli": null
                }
            },
            "Node-2": {
                "type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/formats/gltf.PolyformModel,github.com/EliCDavis/polyform/formats/gltf.ModelNodeData]",
                "dependencies": [
                    {
                        "dependencyID": "Node-10",
                        "dependencyPort": "Out",
                        "name": "GpuInstances"
                    },
                    {
                        "dependencyID": "Node-1",
                        "dependencyPort": "Out",
                        "name": "Mesh"
                    }
                ]
            },
            "Node-3": {
                "type": "github.com/EliCDavis/polyform/nodes.Struct[[]github.com/EliCDavis/polyform/math/trs.TRS,github.com/EliCDavis/polyform/modeling/repeat.CircleNodeData]",
                "dependencies": [
                    {
                        "dependencyID": "Node-4",
                        "dependencyPort": "Out",
                        "name": "Radius"
                    },
                    {
                        "dependencyID": "Node-5",
                        "dependencyPort": "Out",
                        "name": "Times"
                    }
                ]
            },
            "Node-4": {
                "type": "github.com/EliCDavis/polyform/generator/parameter.Value[float64]",
                "data": {
                    "name": "",
                    "description": "",
                    "currentValue": 5,
                    "defaultValue": 0,
                    "cli": null
                }
            },
            "Node-5": {
                "type": "github.com/EliCDavis/polyform/generator/parameter.Value[int]",
                "data": {
                    "name": "",
                    "description": "",
                    "currentValue": 15,
                    "defaultValue": 0,
                    "cli": null
                }
            },
            "Node-6": {
                "type": "github.com/EliCDavis/polyform/nodes.Struct[[]github.com/EliCDavis/polyform/math/trs.TRS,github.com/EliCDavis/polyform/modeling/repeat.TransformationNodeData]",
                "dependencies": [
                    {
                        "dependencyID": "Node-9",
                        "dependencyPort": "Out",
                        "name": "Samples"
                    },
                    {
                        "dependencyID": "Node-7",
                        "dependencyPort": "Out",
                        "name": "Transformation"
                    }
                ]
            },
            "Node-7": {
                "type": "github.com/EliCDavis/polyform/nodes.Struct[github.com/EliCDavis/polyform/math/trs.TRS,github.com/EliCDavis/polyform/math/trs.NewNodeData]",
                "dependencies": [
                    {
                        "dependencyID": "Node-8",
                        "dependencyPort": "Out",
                        "name": "Position"
                    },
                    {
                        "dependencyID": "Node-11",
                        "dependencyPort": "Out",
                        "name": "Rotation"
                    }
                ]
            },
            "Node-8": {
                "type": "github.com/EliCDavis/polyform/generator/parameter.Value[github.com/EliCDavis/vector/vector3.Vector[float64]]",
                "data": {
                    "name": "",
                    "description": "",
                    "currentValue": {
                        "x": 0,
                        "y": 1,
                        "z": 0
                    },
                    "defaultValue": {
                        "x": 0,
                        "y": 0,
                        "z": 0
                    },
                    "cli": null
                }
            },
            "Node-9": {
                "type": "github.com/EliCDavis/polyform/generator/parameter.Value[int]",
                "data": {
                    "name": "",
                    "description": "",
                    "currentValue": 5,
                    "defaultValue": 0,
                    "cli": null
                }
            }
        },
        "producers": {
            "test.glb": {
                "nodeID": "Node-0",
                "port": "Out"
            }
        },
        "version": "0.22.0"
    }
}
const params = {}

for (let entry in FILE.data.nodes) {
    const node = FILE.data.nodes[entry];
    params[entry] = node.type.indexOf("github.com/EliCDavis/polyform/generator/parameter.Value") === 0;
}

for (let entry in FILE.data.nodes) {
    const node = FILE.data.nodes[entry];

    if (!node.dependencies) {
        continue
    }

    input = {}
    for (let i = 0; i < node.dependencies.length; i++) {
        const dep = node.dependencies[i];
        input[dep.name] = {
            "dependencyID": dep.dependencyID,
            "dependencyPort": dep.dependencyPort,
        }

        if (params[dep.dependencyID] === true) {
            input[dep.name].dependencyPort = "Value"
        }
    }
    node.assignedInput = input;
    delete node.dependencies
}

console.log(JSON.stringify(FILE))