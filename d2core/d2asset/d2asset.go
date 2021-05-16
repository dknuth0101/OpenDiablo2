package d2asset

import (
	"github.com/gravestench/weightedcache"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2tbl"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2loader"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2util"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2records"
)

// NewAssetManager creates and assigns all necessary dependencies for the AssetManager top-level functions to work correctly
func NewAssetManager(logLevel d2util.LogLevel) (*AssetManager, error) {
	loader, err := d2loader.NewLoader(logLevel)
	if err != nil {
		return nil, err
	}

	records, err := d2records.NewRecordManager(logLevel)
	if err != nil {
		return nil, err
	}

	logger := d2util.NewLogger()
	logger.SetPrefix(logPrefix)
	logger.SetLevel(logLevel)

	manager := &AssetManager{
		Logger:     logger,
		Loader:     loader,
		tables:     make([]d2tbl.TextDictionary, 0),
		animations: weightedcache.New(animationBudget),
		fonts:      weightedcache.New(fontBudget),
		palettes:   weightedcache.New(paletteBudget),
		transforms: weightedcache.New(paletteTransformBudget),
		dt1s:       weightedcache.New(dt1Budget),
		ds1s:       weightedcache.New(ds1Budget),
		cofs:       weightedcache.New(cofBudget),
		dccs:       weightedcache.New(dccBudget),
		Records:    records,
	}

	return manager, err
}
