package d2audio

import (
	"errors"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2interface"
)

func Bind(e d2interface.od2EngineInterface) error {

	am := &audioManager{}

	if err := e.BindAudioManager(am); err != nil {
		return err
	}

	return nil

}
