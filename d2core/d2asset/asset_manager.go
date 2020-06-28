package d2asset

import (
	"errors"
	"log"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2data/d2datadict"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2cof"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2dat"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2dc6"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2dcc"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2mpq"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2pl2"
)

var (
	ErrWasInit = errors.New("asset system is already initialized")
	ErrNotInit = errors.New("asset system is not initialized")
)

type assetManager struct {
	archiveManager          *archiveManager
	fileManager             *fileManager
	paletteManager          *paletteManager
	paletteTransformManager *paletteTransformManager
	animationManager        *animationManager
	fontManager             *fontManager
}

func (am *assetManager) Shutdown() {
	am = nil
}

func (am *assetManager) LoadArchive(archivePath string) (*d2mpq.MPQ, error) {
	return am.archiveManager.loadArchive(archivePath)
}

func (am *assetManager) LoadFile(filePath string) ([]byte, error) {

	data, err := am.fileManager.loadFile(filePath)
	if err != nil {
		log.Printf("error loading file %s (%v)", filePath, err.Error())
	}

	return data, err
}

func (am *assetManager) FileExists(filePath string) (bool, error) {
	return am.fileManager.fileExists(filePath)
}

func (am *assetManager) LoadAnimation(animationPath, palettePath string) (*Animation, error) {
	return am.LoadAnimationWithTransparency(animationPath, palettePath, 255)
}

func (am *assetManager) LoadPaletteTransform(pl2Path string) (*d2pl2.PL2File, error) {
	return am.paletteTransformManager.loadPaletteTransform(pl2Path)
}

func (am *assetManager) LoadAnimationWithTransparency(animationPath, palettePath string, transparency int) (*Animation, error) {
	return am.animationManager.loadAnimation(animationPath, palettePath, transparency)
}

func (am *assetManager) LoadComposite(object *d2datadict.ObjectLookupRecord, palettePath string) (*Composite, error) {
	return CreateComposite(object, palettePath), nil
}

func (am *assetManager) LoadFont(tablePath, spritePath, palettePath string) (*Font, error) {
	return am.fontManager.loadFont(tablePath, spritePath, palettePath)
}

func (am *assetManager) LoadPalette(palettePath string) (*d2dat.DATPalette, error) {
	return am.paletteManager.loadPalette(palettePath)
}

func (am *assetManager) loadDC6(dc6Path string) (*d2dc6.DC6File, error) {
	dc6Data, err := am.LoadFile(dc6Path)
	if err != nil {
		return nil, err
	}

	dc6, err := d2dc6.LoadDC6(dc6Data)
	if err != nil {
		return nil, err
	}

	return dc6, nil
}

func (am *assetManager) loadDCC(dccPath string) (*d2dcc.DCC, error) {
	dccData, err := am.LoadFile(dccPath)
	if err != nil {
		return nil, err
	}

	return d2dcc.LoadDCC(dccData)
}

func (am *assetManager) loadCOF(cofPath string) (*d2cof.COF, error) {
	cofData, err := am.LoadFile(cofPath)
	if err != nil {
		return nil, err
	}

	return d2cof.LoadCOF(cofData)
}

func (am *assetManager) BindTerminalCommands(t d2interface.TermManager) {
	t.BindAction("assetspam", "display verbose asset manager logs", func(verbose bool) {
		if verbose {
			t.OutputInfo("asset manager verbose logging enabled")
		} else {
			t.OutputInfo("asset manager verbose logging disabled")
		}

		archiveManager.cache.SetVerbose(verbose)
		fileManager.cache.SetVerbose(verbose)
		paletteManager.cache.SetVerbose(verbose)
		paletteTransformManager.cache.SetVerbose(verbose)
		animationManager.cache.SetVerbose(verbose)
	})

	t.BindAction("assetstat", "display asset manager cache statistics", func() {
		t.OutputInfo("archive cache: %f", float64(archiveManager.cache.GetWeight())/float64(archiveManager.cache.GetBudget())*100.0)
		t.OutputInfo("file cache: %f", float64(fileManager.cache.GetWeight())/float64(fileManager.cache.GetBudget())*100.0)
		t.OutputInfo("palette cache: %f", float64(paletteManager.cache.GetWeight())/float64(paletteManager.cache.GetBudget())*100.0)
		t.OutputInfo("palette transform cache: %f", float64(paletteTransformManager.cache.GetWeight())/float64(paletteTransformManager.cache.GetBudget())*100.0)
		t.OutputInfo("animation cache: %f", float64(animationManager.cache.GetWeight())/float64(animationManager.cache.GetBudget())*100.0)
		t.OutputInfo("font cache: %f", float64(fontManager.cache.GetWeight())/float64(fontManager.cache.GetBudget())*100.0)
	})

	t.BindAction("assetclear", "clear asset manager cache", func() {
		archiveManager.cache.Clear()
		fileManager.cache.Clear()
		paletteManager.cache.Clear()
		paletteTransformManager.cache.Clear()
		animationManager.cache.Clear()
		fontManager.cache.Clear()
	})
}
