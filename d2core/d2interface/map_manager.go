package d2interface

type MapManager interface {
	Initialize() error
	Shutdown() error
	MakeMapRealm() error
	MakeMapAct() error
	BindTerminalCommands() error
	UnbindTerminalCommands() error
}
