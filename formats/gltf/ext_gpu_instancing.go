package gltf

const extGpuInstancingID = "EXT_mesh_gpu_instancing"

type ExtGpuInstancing struct {
	Attributes map[string]int `json:"attributes"`
}
