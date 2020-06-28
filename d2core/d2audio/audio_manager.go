package d2audio

import (
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2audio/ebiten"
)

type audioManager struct {
	audioType d2interface.AudioManagerType
	provider  d2interface.AudioManager
}

// CreateManager creates a sound provider
func (am *AudioManager) Initialize(t d2interface.AudioManagerType) error {

	switch t {
	case d2interface.AudioManagerNone:
		am.audioType = t
		am.provider = nil
	case d2interface.AudioManagerEbiten:
		am.audioType = t
		am.provider = ebiten.CreateAudio()
	default:
		am.audioType = d2interface.AudioManagerEbiten
		am.provider = ebiten.CreateAudio()
	}

	singleton = audioProvider
	return nil
}

// AudioManager.PlayBGM plays an infinitely looping background track
func (am *AudioManager) PlayBGM(song string) error {

	if am.provider == nil && am.audioType == d2interface.AudioManagerNone {
		return nil
	}

	if am.provider == nil && am.audioType != d2interface.AudioManagerNone {
		return errors.New("No audio provider, or audio manager not initialized")
	}

	go func() {
		am.provider.PlayBGM(song)
	}()

	return nil
}

func (am *AudioManager) LoadSoundEffect(sfx string) (SoundEffect, error) {
	if am.provider == nil {
		return nil, errors.New("No audio provider")
	}

	return am.LoadSoundEffect(sfx)
}

func (am *AudioManager) SetVolumes(bgmVolume, sfxVolume float64) {
	if am.provider == nil {
		return nil, errors.New("No audio provider")
	}

	am.SetVolumes(bgmVolume, sfxVolume)
}
