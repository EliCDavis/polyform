package pipeline

type Operate func(*View)

type Command struct {
	operator         Operate
	readPermissions  Permission
	writePermissions Permission
}

func NewCommand(readPermissions, writePermissions Permission, operator Operate) Command {
	return Command{
		operator:         operator,
		readPermissions:  readPermissions,
		writePermissions: writePermissions,
	}
}
