package pipeline

type dataAccess int

const (
	none dataAccess = iota
	read
	write
)

type pipelineBuilder struct {
	waves       [][]Command
	currentWave []Command

	dataV4 map[string]dataAccess
	dataV3 map[string]dataAccess
	dataV2 map[string]dataAccess
	dataV1 map[string]dataAccess

	materials dataAccess
	indices   dataAccess
}

func (pb *pipelineBuilder) completeCurrentWave() {
	if len(pb.currentWave) == 0 {
		return
	}
	pb.waves = append(pb.waves, pb.currentWave)
	pb.currentWave = make([]Command, 0)

	pb.dataV1 = make(map[string]dataAccess)
	pb.dataV2 = make(map[string]dataAccess)
	pb.dataV3 = make(map[string]dataAccess)
	pb.dataV4 = make(map[string]dataAccess)

	pb.materials = none
	pb.indices = none
}

func checkOccupied[T, G any](ledger map[string]dataAccess, readRequests map[string]T, writeRequests map[string]G) bool {
	for attr := range readRequests {
		if v, ok := ledger[attr]; ok {
			return v == write
		}
	}

	for attr := range writeRequests {
		if _, ok := ledger[attr]; ok {
			return true
		}
	}

	return false
}

func updateLedger[T, G any](ledger map[string]dataAccess, readRequests map[string]T, writeRequests map[string]G) {
	for attr := range readRequests {
		ledger[attr] = read
	}

	for attr := range writeRequests {
		ledger[attr] = write
	}
}

func (pb *pipelineBuilder) add(command Command) {
	writePermissions := command.WritePermissions()
	readPermissions := command.ReadPermissions()
	if readPermissions.Everything != nil || writePermissions.Everything != nil {
		pb.completeCurrentWave()
		pb.waves = append(pb.waves, []Command{command})
		return
	}

	if checkOccupied(pb.dataV1, readPermissions.V1Permissions, writePermissions.V1Permissions) {
		pb.completeCurrentWave()
	}
	if checkOccupied(pb.dataV2, readPermissions.V2Permissions, writePermissions.V2Permissions) {
		pb.completeCurrentWave()
	}
	if checkOccupied(pb.dataV3, readPermissions.V3Permissions, writePermissions.V3Permissions) {
		pb.completeCurrentWave()
	}
	if checkOccupied(pb.dataV4, readPermissions.V4Permissions, writePermissions.V4Permissions) {
		pb.completeCurrentWave()
	}

	if readPermissions.Indices != nil {
		if pb.indices == write {
			pb.completeCurrentWave()
		}
		pb.indices = read
	}

	if writePermissions.Indices != nil {
		if pb.indices != none {
			pb.completeCurrentWave()
		}
		pb.indices = write
	}

	updateLedger(pb.dataV1, readPermissions.V1Permissions, writePermissions.V1Permissions)
	updateLedger(pb.dataV2, readPermissions.V2Permissions, writePermissions.V2Permissions)
	updateLedger(pb.dataV3, readPermissions.V3Permissions, writePermissions.V3Permissions)
	updateLedger(pb.dataV4, readPermissions.V4Permissions, writePermissions.V4Permissions)
}

func (pb pipelineBuilder) build() Pipeline {
	allWaves := pb.waves
	allWaves = append(allWaves, pb.currentWave)
	return Pipeline{allWaves}
}

func Schedule(commandsToSchedule ...Command) Pipeline {
	builder := pipelineBuilder{
		waves:       make([][]Command, 0),
		currentWave: make([]Command, 0),
	}

	for _, command := range commandsToSchedule {
		if command != nil {
			continue
		}

		builder.add(command)
	}

	return builder.build()
}
