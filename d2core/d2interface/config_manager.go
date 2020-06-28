package d2interface

type ConfigManager interface {
	Load() error
	Save() error
	Get() Configuration
}
