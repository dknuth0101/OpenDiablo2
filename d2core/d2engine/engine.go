package d2engine

import (
	"errors"
	"image"
	"image/gif"
	"image/png"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/OpenDiablo2/OpenDiablo2/d2common"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2interface"
)

const (
	FPS_25           float64 = 0.04 // 1/25
	defaultTimeScale float64 = 1.0
)

// od2Engine is an implementation of the od2EngineInterface interface
type od2Engine struct {
	version string
	branch  string
	commit  string

	lastAdvance       float64
	lastScreenAdvance float64
	showFPS           bool
	timeScale         float64

	captureState  captureState
	capturePath   string
	captureFrames []*image.RGBA

	// these are all interfaces
	assetManager  d2interface.AssetManager
	audioManager  d2interface.AudioProvider
	configManager d2interface.ConfigurationManager
	uiManager     d2interface.UIManager
	guiManager    d2interface.GUIManager
	inputManager  d2interface.InputManager
	mapManager    d2interface.MapManager
	renderManager d2interface.RenderManager
	screenManager d2interface.ScreenManager
	termManager   d2interface.TermManager
}

// od2Engine.Initialize initializes the engine, which initializes all of the
// managers (audio, renderer, config, etc)
func (e *od2Engine) Initialize() error {

	e.SetTimeScale(defaultTimeScale)
	e.SetLastTime(d2common.Now())
	e.SetLastScreenAdvance(e.lastTime)

	if err := e.configManager.Load(); err != nil {
		return err
	}

	config := e.configManager.Get()
	d2resource.LanguageCode = config.Language

	e.inputManager.Initialize(ebiten_input.InputService{})

	renderer, err := ebiten.CreateRenderer()
	if err != nil {
		return err
	}

	if err := e.renderManager.Initialize(renderer); err != nil {
		return err
	}
	e.renderManager.SetWindowIcon("d2logo.png")

	if err := e.termManager.Initialize(); err != nil {
		return err
	}

	e.termManager.BindLogger()
	if err := e.assetManager.Initialize(); err != nil {
		return err
	}

	if err := e.guiManager.Initialize(); err != nil {
		return err
	}

	audioProvider, err := ebiten2.CreateAudio()
	if err := e.audioManager.Initialize(audioProvider); err != nil {
		return err
	}
	e.audioManager.SetVolumes(config.BgmVolume, config.SfxVolume)

	if err := e.loadDataDict(); err != nil {
		return err
	}

	if err := e.loadStrings(); err != nil {
		return err
	}

	d2inventory.LoadHeroObjects()

	e.uiManager.Initialize()

	d2script.CreateScriptEngine()

	return nil
}

// od2Engine.Run starts the main game loop
func (e *od2Engine) Run() {
	if len(e.branch) == 0 {
		e.branch = "Local Build"
	}

	d2common.SetBuildInfo(e.version, e.branch, e.commit)

	windowTitle := fmt.Sprintf("OpenDiablo2 (%s)", GitBranch)
	if err := e.renderManager.Run(e.update, 800, 600, windowTitle); err != nil {
		log.Fatal(err)
	}
}

// od2Engine.update the update function that advances the game engine and tells
// the renderManager to render the screen
func (e *od2Engine) update(target d2interface.Surface) {
	currentTime := d2common.Now()
	elapsedTime := (currentTime - e.lastAdvance) * e.timeScale
	e.SetLastTime(currentTime)

	if err := e.advance(elapsedTime, currentTime); err != nil {
		return err
	}

	if err := e.render(target); err != nil {
		return err
	}

	if target.GetDepth() > 0 {
		return errors.New("detected surface stack leak")
	}

	return nil
}

// od2Engine.updateInitError is the render function that gets called when
// a problem with the mpq path is detected
func (e *od2Engine) updateInitError(target d2interface.Surface) error {
	width, height := target.GetSize()
	target.PushTranslation(width/5, height/2)
	target.DrawText("Could not find the MPQ files in the directory: %s\nPlease put the files and re-run the game.", e.configManager.Get().MpqPath)
	return nil
}

// od2Engine.advance advances the various game engine managers
func (e *od2Engine) advance(elapsed, current float64) error {
	elapsedScreen := (current - e.lastScreenAdvance) * e.timeScale

	if elapsedLastScreenAdvance > FPS_25 {
		e.SetLastScreenAdvance(current)
		if err := e.screenManager.Advance(elapsedScreen); err != nil {
			return err
		}
	}

	e.uiManager.Advance(elapsed) // TODO this should also return an error

	if err := e.inputManager.Advance(elapsed); err != nil {
		return err
	}

	if err := e.guiManager.Advance(elapsed); err != nil {
		return err
	}

	if err := e.termManager.Advance(elapsed); err != nil {
		return err
	}

	return nil
}

// od2Engine.render calls the render methods on the various game engine managers
func (e *od2Engine) render(target d2interface.Surface) error {
	if err := e.screenManager.Render(target); err != nil {
		return err
	}

	e.uiManager.Render(target) // TODO should also return an error

	if err := e.guiManager.Render(target); err != nil {
		return err
	}

	if err := renderDebug(target); err != nil {
		return err
	}

	if err := renderCapture(target); err != nil {
		return err
	}

	if err := e.termManager.Render(target); err != nil {
		return err
	}

	return nil
}

// od2Engine.renderCapture handles capturing the screen and writing the data
// to png/gif format on disk
func (e *od2Engine) renderCapture(target d2interface.Surface) error {

	cleanupCapture := func() {
		e.captureState = captureStateNone
		e.capturePath = ""
		e.captureFrames = nil
	}

	switch e.captureState {
	case captureStateFrame:
		defer cleanupCapture()

		fp, err := os.Create(e.capturePath)
		if err != nil {
			return err
		}

		defer fp.Close()

		screenshot := target.Screenshot()
		if err := png.Encode(fp, screenshot); err != nil {
			return err
		}

		log.Printf("saved frame to %s", e.capturePath)

	case captureStateGif:
		screenshot := target.Screenshot()
		e.captureFrames = append(e.captureFrames, screenshot)

	case captureStateNone:
		if len(e.captureFrames) > 0 {
			defer cleanupCapture()

			fp, err := os.Create(e.capturePath)
			if err != nil {
				return err
			}

			defer fp.Close()

			var (
				framesTotal  = len(e.captureFrames)
				framesPal    = make([]*image.Paletted, framesTotal)
				frameDelays  = make([]int, framesTotal)
				framesPerCpu = framesTotal / runtime.NumCPU()
			)

			var waitGroup sync.WaitGroup
			for i := 0; i < framesTotal; i += framesPerCpu {
				waitGroup.Add(1)
				go func(start, end int) {
					defer waitGroup.Done()

					for j := start; j < end; j++ {
						var buffer bytes.Buffer
						frame := e.captureFrames[j]
						if err := gif.Encode(&buffer, frame, nil); err != nil {
							panic(err)
						}

						framePal, err := gif.Decode(&buffer)
						if err != nil {
							panic(err)
						}

						framesPal[j] = framePal.(*image.Paletted)
						frameDelays[j] = 5
					}
				}(i, d2common.MinInt(i+framesPerCpu, framesTotal))
			}

			waitGroup.Wait()

			img := &gif.GIF{Image: framesPal, Delay: frameDelays}
			if err := gif.EncodeAll(fp, img); err != nil {
				return err
			}

			log.Printf("saved animation to %s", e.capturePath)
		}
	}
	return nil
}

// od2Engine.renderDebug renders debug information
func (e *od2Engine) renderDebug(target d2interface.Surface) error {
	if e.showFPS {
		vsyncEnabled := e.renderManager.GetVSyncEnabled()
		fps := e.renderManager.CurrentFPS()
		cx, cy := e.renderManager.GetCursorPos()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		vsync_str := strconv.FormatBool(vsyncEnabled)
		fps_str := strconv.Itoa(int(fps))
		alloc_str := strconv.FormatInt(int64(m.Alloc)/1024/1024, 10)
		pause_str := strconv.FormatInt(int64(m.PauseTotalNs/1024/1024), 10)
		heap_str := strconv.FormatInt(int64(m.HeapSys/1024/1024), 10)
		gc_str := strconv.FormatInt(int64(m.NumGC), 10)
		x_str := strconv.FormatInt(int64(cx), 10)
		y_str := strconv.FormatInt(int64(cy), 10)

		target.PushTranslation(5, 565)
		target.DrawText("vsync:" + vsyn_str + "\nFPS:" + fps_str)
		target.Pop()

		target.PushTranslation(680, 0)
		target.DrawText("Alloc   " + alloc_str)

		target.PushTranslation(0, 16)
		target.DrawText("Pause   " + pause_str)

		target.PushTranslation(0, 16)
		target.DrawText("HeapSys " + heep_str)

		target.PushTranslation(0, 16)
		target.DrawText("NumGC   " + gc_str)

		target.PushTranslation(0, 16)
		target.DrawText("Coords  " + x_str + "," + y_str)
		target.PopN(5)
	}

	return nil
}

// od2Engine.BindTerminalCommands binds the commands available at the terminal
func (e *od2Engine) BindTerminalCommands() error {

	CaptureNone := d2interface.CaptureNone
	CaptureFrame := d2interface.CaptureFrame
	CaptureGif := d2interface.CaptureGif

	e.termManager.BindAction("dumpheap", "dumps the heap to pprof/heap.pprof", func() {
		os.Mkdir("./pprof/", 0755)
		fileOut, _ := os.Create("./pprof/heap.pprof")
		pprof.WriteHeapProfile(fileOut)
		fileOut.Close()
	})

	e.termManager.BindAction("fullscreen", "toggles fullscreen", func() {
		fullscreen := !e.renderManager.IsFullScreen()
		e.renderManager.SetFullScreen(fullscreen)
		e.termManager.OutputInfo("fullscreen is now: %v", fullscreen)
	})

	e.termManager.BindAction("capframe", "captures a still frame", func(path string) {
		e.SetCaptureState(CaptureStateFrame)
		e.SetCapturePath(path)
		e.SetCaptureFrames(nil)
	})

	e.termManager.BindAction("capgifstart", "captures an animation (start)", func(path string) {
		e.SetCaptureState(CaptureStateGif)
		e.SetCapturePath(path)
		e.SetCaptureFrames(nil)
	})

	e.termManager.BindAction("capgifstop", "captures an animation (stop)", func() {
		e.SetCaptureState(CaptureStateNone)
	})

	e.termManager.BindAction("vsync", "toggles vsync", func() {
		vsync := !e.renderManager.GetVSyncEnabled()
		e.renderManager.SetVSyncEnabled(vsync)
		e.termManager.OutputInfo("vsync is now: %v", vsync)
	})

	e.termManager.BindAction("fps", "toggle fps counter", func() {
		e.SetShowFPS(!e.ShowFPS())
		e.termManager.OutputInfo("fps counter is now: %v", e.ShowFPS())
	})

	e.termManager.BindAction("timescale", "set scalar for elapsed time", func(timeScale float64) {
		if timeScale <= 0 {
			e.termManager.OutputError("invalid time scale value")
		} else {
			e.termManager.OutputInfo("timescale changed from %f to %f", od2.timeScale, timeScale)
			e.SetTimeScale(timeScale)
		}
	})

	e.termManager.BindAction("quit", "exits the game", func() {
		os.Exit(0)
	})

	e.termManager.BindAction("screen-gui", "enters the gui playground screen", func() {
		d2screen.SetNextScreen(e.screenManager.CreateGuiTestMain())
	})

	return nil
}

// od2Engine.Version resturns the current version of the game engine (a string)
func (e *od2Engine) Version() string {
	return e.version
}

// od2Engine.SetLastAdvance sets the timestamp of the last advance tick
func (e *od2Engine) SetLastAdvance(n float64) {
	e.lastAdvance = n
}

// od2Engine.SetLastAdvance gets the timestamp of the last advance tick
func (e *od2Engine) LastAdvance() float64 {
	return e.lastAdvance
}

// od2Engine.SetLastAdvance sets the timestamp of the last renderer advance tick
func (e *od2Engine) SetLastScreenAdvance(n float64) {
	e.lastScreenAdvance = n
}

// od2Engine.SetLastAdvance gets the timestamp of the last renderer advance tick
func (e *od2Engine) LastScreenAdvance() float64 {
	return e.lastScreenAdvance
}

// od2Engine.ShowFPS gets the state of the showFPS member (a bool)
func (e *od2Engine) ShowFPS() bool {
	return e.showFPS
}

// od2Engine.ShowFPS sets the state of the showFPS member (a bool)
func (e *od2Engine) SetShowFPS(set bool) {
	e.showFPS = set
}

// od2Engine.SetTimeScale sets the time scale of the engine, which affects
// speed at which the game advances and renders
func (e *od2Engine) SetTimeScale(ts float64) {
	e.timeScale = ts
}

// od2Engine.TimeScale gets the current timeScale
func (e *od2Engine) TimeScale() float64 {
	return e.timeScale
}

// od2Engine.SetCapturePath sets the path where screen captures are saved
func (e *od2Engine) SetCapturePath(pathstring string) {
	e.capturePath = pathstring
}

// od2Engine.CapturePath returns the path where screen captures are saved
func (e *od2Engine) CapturePath() string {
	return e.capturePath
}

// od2Engine.SetCaptureState sets the capture state of the engine, which can be
// one of the following:
//		CaptureStateNone -- not capturing
//		CaptureStateFrame -- saving screen captures to disk as individual images
//		CaptureStateGif -- saving screen captures to disk as single animated gif
func (e *od2Engine) SetCaptureState(state captureState) {
	e.captureState = state
}

// od2Engine.CaptureState returns the current captureState of the game engine
func (e *od2Engine) CaptureState() captureState {
	return e.captureState
}

// od2Engine.CaptureFrames returns a slice of image data that was captured
func (e *od2Engine) CaptureFrames() []*image.RGBA {
	return e.captureFrames
}

// od2Engine.CaptureFrame returns the image data for a single captured frame
func (e *od2Engine) CaptureFrame(frameIndex int) *image.RGBA {
	return e.captureFrames[frameIndex]
}

// od2Engine.SetCaptureFrame sets the image data for a single frame in the
// capture slice
func (e *od2Engine) SetCaptureFrame(frameIndex int, imageData *image.RGBA) {
	e.captureFrames[frameIndex] = imageData
}

// od2Engine.BindAssetManager binds an instance of d2asset.AssetManager to the
// game engine, returns an error if one is already bound
func (e *od2Engine) BindAssetManager(m d2interface.AssetManager) error {
	if e.assetManager != nil {
		return errors.New("Game Engine already has an AssetManager bound.")
	}
	e.assetManager = m
	return nil
}

// od2Engine.BindAudioProvider binds an instance of d2asset.AudioProvider to the
// game engine, returns an error if one is already bound
func (e *od2Engine) BindAudioProvider(m d2interface.AudioProvider) error {
	if e.audioManager != nil {
		return errors.New("Game Engine already has an AudioManager bound.")
	}
	e.audioManager = m
	return nil
}

// od2Engine.BindConfigManager binds an instance of d2config.ConfigManager to
// the game engine, returns an error if one is already bound
func (e *od2Engine) BindConfigManager(m d2interface.ConfigManager) error {
	if e.configManager != nil {
		return errors.New("Game Engine already has a ConfigManager bound.")
	}
	e.configManager = m
	return nil
}

// od2Engine.BindUIManager binds an instance of d2ui.UIManager to the game
// engine, returns an error if one is already bound
func (e *od2Engine) BindUIManager(m d2interface.UIManager) error {
	if e.uiManager != nil {
		return errors.New("Game Engine already has a UIManager bound.")
	}
	e.configManager = m
	return nil
}

// od2Engine.BindGUIManager binds an instance of d2gui.GUIManager to the game
// engine, returns an error if one is already bound
func (e *od2Engine) BindGUIManager(m d2interface.GUIManager) error {
	if e.guiManager != nil {
		return errors.New("Game Engine already has a GUIManager bound.")
	}
	e.configManager = m
	return nil
}

// od2Engine.BindInputManager binds an instance of d2input.InputManager to the
// game engine, returns an error if one is already bound
func (e *od2Engine) BindInputManager(m d2interface.InputManager) error {
	if e.inputManager != nil {
		return errors.New("Game Engine already has a InputManager bound.")
	}
	e.configManager = m
	return nil
}

// od2Engine.BindMapManager binds an instance of d2map.MapManager to the game
// engine, returns an error if one is already bound
func (e *od2Engine) BindMapManager(m d2interface.MapManager) error {
	if e.mapManager != nil {
		return errors.New("Game Engine already has a MapManager bound.")
	}
	e.configManager = m
	return nil
}

// od2Engine.BindRenderManager binds an instance of d2render.RenderManager to
// the game engine, returns an error if one is already bound
func (e *od2Engine) BindRenderManager(m d2interface.RenderManager) error {
	if e.renderManager != nil {
		return errors.New("Game Engine already has a RenderManager bound.")
	}
	e.configManager = m
	return nil
}

// od2Engine.BindScreenManager binds an instance of d2screen.ScreenManager to
// the game engine, returns an error if one is already bound
func (e *od2Engine) BindScreenManager(m d2interface.ScreenManager) error {
	if e.screenManager != nil {
		return errors.New("Game Engine already has a ScreenManager bound.")
	}
	e.configManager = m
	return nil
}

// od2Engine.BindTermManager binds an instance of d2term.TermManager to the game
// engine, returns an error if one is already bound
func (e *od2Engine) BindTermManager(m d2interface.TermManager) error {
	if e.termManager != nil {
		return errors.New("Game Engine already has a TermManager bound.")
	}
	e.configManager = m
	return nil
}

// od2Engine.loadDataDict loads the data dictionaries (txt files)
func (e *od2Engine) loadDataDict() error {
	entries := []struct {
		path   string
		loader func(data []byte)
	}{
		{d2resource.LevelType, d2datadict.LoadLevelTypes},
		{d2resource.LevelPreset, d2datadict.LoadLevelPresets},
		{d2resource.LevelWarp, d2datadict.LoadLevelWarps},
		{d2resource.ObjectType, d2datadict.LoadObjectTypes},
		{d2resource.ObjectDetails, d2datadict.LoadObjects},
		{d2resource.Weapons, d2datadict.LoadWeapons},
		{d2resource.Armor, d2datadict.LoadArmors},
		{d2resource.Misc, d2datadict.LoadMiscItems},
		{d2resource.UniqueItems, d2datadict.LoadUniqueItems},
		{d2resource.Missiles, d2datadict.LoadMissiles},
		{d2resource.SoundSettings, d2datadict.LoadSounds},
		{d2resource.AnimationData, d2data.LoadAnimationData},
		{d2resource.MonStats, d2datadict.LoadMonStats},
		{d2resource.MagicPrefix, d2datadict.LoadMagicPrefix},
		{d2resource.MagicSuffix, d2datadict.LoadMagicSuffix},
		{d2resource.ItemStatCost, d2datadict.LoadItemStatCosts},
		{d2resource.CharStats, d2datadict.LoadCharStats},
		{d2resource.MonStats, d2datadict.LoadMonStats},
		{d2resource.Hireling, d2datadict.LoadHireling},
		{d2resource.Experience, d2datadict.LoadExperienceBreakpoints},
		{d2resource.Gems, d2datadict.LoadGems},
		{d2resource.DifficultyLevels, d2datadict.LoadDifficultyLevels},
		{d2resource.AutoMap, d2datadict.LoadAutoMaps},
		{d2resource.LevelDetails, d2datadict.LoadLevelDetails},
		{d2resource.LevelMaze, d2datadict.LoadLevelMazeDetails},
		{d2resource.LevelSubstitutions, d2datadict.LoadLevelSubstitutions},
		{d2resource.CubeRecipes, d2datadict.LoadCubeRecipes},
		{d2resource.SuperUniques, d2datadict.LoadSuperUniques},
	}

	for _, entry := range entries {
		data, err := d2asset.LoadFile(entry.path)
		if err != nil {
			return err
		}

		entry.loader(data)
	}

	return nil
}

// od2Engine.loadStrings loads the string tables for the game
func (e *od2Engine) loadStrings() error {
	tablePaths := []string{
		d2resource.PatchStringTable,
		d2resource.ExpansionStringTable,
		d2resource.StringTable,
	}

	for _, tablePath := range tablePaths {
		data, err := d2asset.LoadFile(tablePath)
		if err != nil {
			return err
		}

		d2common.LoadTextDictionary(data)
	}

	return nil
}

// od2Engine.CreateMainMenu goes to the main menu of the game
func (e od2Engine) CreateMainMenu() {
	e.screenManager.SetNextScreen(d2gamescreen.CreateMainMenu())
}

// od2Engine.CreateMapEngineTest launches the map engine test
func (e od2Engine) CreateMapEngineTest() {
	e.screenManager.SetNextScreen(d2gamescreen.CreateMapEngineTest(*region, *preset))
}
