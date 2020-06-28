package d2audio

import (
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2interface"
)

func Bind(e d2interface.Od2EngineInterface) error {

	am := &audioManager{}

	if err := e.BindAudioManager(am); err != nil {
		return err
	}

	return nil

}
