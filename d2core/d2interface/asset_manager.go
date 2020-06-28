package d2interface

type AssetManager interface {
	Initialize() error
	Shutdown() error
	LoadArchive(archivePath string) (MPQ, error)
	LoadFile(filePath string) ([]byte, error)
	FileExists(filePath string) (bool, error)
	LoadAnimation(animationPath, palettePath string) (Animation, error)
	LoadPaletteTransform(pl2Path string) (PL2File, error)
	LoadAnimationWithTransparency(string, string, int) (Animation, error)
	LoadComposite(ObjectLookupRecord, string) (Composite, error)
	LoadFont(tablePath, spritePath, palettePath string) (Font, error)
	LoadPalette(palettePath string) (DATPalette, error)
	BindTerminalCommands(TermManager) error
	UnbindTerminalCommands(TermManager) error
}
