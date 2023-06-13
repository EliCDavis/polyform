package pipeline

type Command interface {
	Run()
	ReadPermissions() MeshReadPermission
	WritePermissions() MeshWritePermission
}
