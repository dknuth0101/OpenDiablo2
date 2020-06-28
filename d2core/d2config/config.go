package d2config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"runtime"
)

// Configuration defines the configuration for the engine, loaded from config.json
type configManager struct {
	MpqLoadOrder    []string
	Language        string
	MpqPath         string
	TicksPerSecond  int
	FpsCap          int
	SfxVolume       float64
	BgmVolume       float64
	FullScreen      bool
	RunInBackground bool
	VsyncEnabled    bool
}

var singleton = getDefaultConfig()

func (cf *configManager) Load() error {
	configPaths := []string{
		getLocalConfigPath(),
		getDefaultConfigPath(),
	}

	var loaded bool
	for _, configPath := range configPaths {
		log.Printf("loading configuration file from %s...", configPath)
		if err := load(configPath); err == nil {
			loaded = true
			break
		}
	}

	if !loaded {
		log.Println("failed to load configuration file, saving default configuration...")
		if err := Save(); err != nil {
			return err
		}
	}

	return nil
}

func (cf *configManager) Save() error {
	configPath := getDefaultConfigPath()
	log.Printf("saving configuration file to %s...", configPath)

	var err error
	if err = save(configPath); err != nil {
		log.Printf("failed to write configuration file (%s)", err)
	}

	return err
}

func (cf *configManager) Get() Configuration {
	if singleton == nil {
		panic("configuration is not initialized")
	}

	return *singleton
}

func getDefaultConfig() *Configuration {
	config := &Configuration{
		Language:        "ENG",
		FullScreen:      false,
		TicksPerSecond:  -1,
		RunInBackground: true,
		VsyncEnabled:    true,
		SfxVolume:       1.0,
		BgmVolume:       0.3,
		MpqPath:         "C:/Program Files (x86)/Diablo II",
		MpqLoadOrder: []string{
			"Patch_D2.mpq",
			"d2exp.mpq",
			"d2xmusic.mpq",
			"d2xtalk.mpq",
			"d2xvideo.mpq",
			"d2data.mpq",
			"d2char.mpq",
			"d2music.mpq",
			"d2sfx.mpq",
			"d2video.mpq",
			"d2speech.mpq",
		},
	}

	switch runtime.GOOS {
	case "windows":
		switch runtime.GOARCH {
		case "386":
			config.MpqPath = "C:/Program Files/Diablo II"
		}
	case "darwin":
		config.MpqPath = "/Applications/Diablo II/"
		config.MpqLoadOrder = []string{
			"Diablo II Patch",
			"Diablo II Expansion Data",
			"Diablo II Expansion Movies",
			"Diablo II Expansion Music",
			"Diablo II Expansion Speech",
			"Diablo II Game Data",
			"Diablo II Graphics",
			"Diablo II Movies",
			"Diablo II Music",
			"Diablo II Sounds",
			"Diablo II Speech",
		}
	case "linux":
		if usr, err := user.Current(); err == nil {
			config.MpqPath = path.Join(usr.HomeDir, ".wine/drive_c/Program Files (x86)/Diablo II")
		}
	}

	return config
}

func (cf *configManager) getDefaultConfigPath() string {
	if configDir, err := os.UserConfigDir(); err == nil {
		return path.Join(configDir, "OpenDiablo2", "config.json")
	}

	return getLocalConfigPath()
}

func (cf *configManager) getLocalConfigPath() string {
	return path.Join(path.Dir(os.Args[0]), "config.json")
}

func (cf *configManager) load(configPath string) error {
	configFile, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer configFile.Close()
	data, err := ioutil.ReadAll(configFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &singleton); err != nil {
		return err
	}

	return nil
}

func (cf *configManager) save(configPath string) error {
	configDir := path.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	data, err := json.MarshalIndent(singleton, "", "    ")
	if err != nil {
		return err
	}

	if _, err := configFile.Write(data); err != nil {
		return err
	}

	return nil
}
