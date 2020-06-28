package d2engine

import (
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2audio"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2config"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2gui"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2input"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2interface"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2render"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2screen"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2term"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2ui"
)

func Create(version, branch, commit string) (*od2Engine, error) {
	engine := &od2Engine{
		version: version,
		branch:  branch,
		commit:  commit,
	}

	defer engine.BindTerminalCommands()

	// implicit singletons are bound to the engine :)
	if err := d2term.Bind(engine); err != nil {
		return err
	}

	// IMPORTANT!! config must be bound to the engine so that other managers
	// can use it
	if err := d2config.Bind(engine); err != nil {
		return err
	}

	if err := d2asset.Bind(engine); err != nil {
		return err
	}

	if err := d2audio.Bind(engine); err != nil {
		return err
	}

	if err := d2gui.Bind(engine); err != nil {
		return err
	}

	if err := d2input.Bind(engine); err != nil {
		return err
	}

	if err := d2map.Bind(engine); err != nil {
		return err
	}

	if err := d2render.Bind(engine); err != nil {
		return err
	}

	if err := d2screen.Bind(engine); err != nil {
		return err
	}

	if err := d2ui.Bind(engine); err != nil {
		return err
	}

	return engine, nil
}

// bindTerminalCommands calls the BindTerminalCommands method on each of the
// game engine's managers
func bindTerminalCommands(e d2interface.Od2EngineInterface) {
	e.assetManager.BindTerminalCommands()
	e.audioManager.BindTerminalCommands()
	e.configManager.BindTerminalCommands()
	e.uiManager.BindTerminalCommands()
	e.guiManager.BindTerminalCommands()
	e.inputManager.BindTerminalCommands()
	e.mapManager.BindTerminalCommands()
	e.renderManager.BindTerminalCommands()
	e.screen.BindTerminalCommands()

	e.BindTerminalCommands()
}
