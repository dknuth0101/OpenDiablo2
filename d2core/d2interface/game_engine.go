package d2interface

import (
	"image"
)

type Od2EngineInterface interface {
	Run(func(Surface) error) error

	Version() string
	SetLastAdvance(float64)
	LastAdvance() float64
	SetLastScreenAdvance(float64)
	LastScreenAdvance() float64
	ShowFPS() bool
	SetShowFPS(bool)
	SetTimeScale(float64)
	TimeScale() float64

	SetCapturePath(string)
	CapturePath() string
	SetCaptureState(CaptureState)
	CaptureState() CaptureState
	CaptureFrames() []*image.RGBA
	CaptureFrame(int) *image.RGBA
	SetCaptureFrame(int, *image.RGBA)

	CreateMainMenu()
	CreateMapEngineTest()

	BindAssetManager(AssetManager) error
	BindAudioProvider(AudioManager) error
	BindConfigManager(ConfigManager) error
	BindUIManager(UIManager) error
	BindGUIManager(GUIManager) error
	BindInputManager(InputManager) error
	BindMapManager(MapManager) error
	BindRenderManager(RenderManager) error
	BindScreenManager(ScreenManager) error
	BindTermManager(TermManager) error

	BindTerminalCommands()
	UnbindTerminalCommands()
}
