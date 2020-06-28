package d2asset

import (
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2config"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2interface"
)

func Bind(e d2interface.od2EngineInterface) error {

	var (
		config                  = d2config.Get()
		archiveManager          = createArchiveManager(config)
		fileManager             = createFileManager(config, archiveManager)
		paletteManager          = createPaletteManager()
		paletteTransformManager = createPaletteTransformManager()
		animationManager        = createAnimationManager()
		fontManager             = createFontManager()
	)

	am := &assetManager{
		archiveManager,
		fileManager,
		paletteManager,
		paletteTransformManager,
		animationManager,
		fontManager,
	}

	if err := e.BindAssetManager(am); err != nil {
		return err
	}

	return nil

}
