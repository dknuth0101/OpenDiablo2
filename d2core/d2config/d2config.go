package d2config

import (
	"log"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2interface"
)

func Bind(e d2interface.od2EngineInterface) error {

	cm := &configManager{}

	if err := e.BindConfigManager(cm); err != nil {
		return err
	}

	return nil

}
