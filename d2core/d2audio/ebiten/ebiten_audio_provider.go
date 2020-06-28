package ebiten

import (
	"log"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2audio"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2interface"
	"github.com/hajimehoshi/ebiten/audio/wav"

	"github.com/hajimehoshi/ebiten/audio"
)

func CreateAudio() (d2interface.AudioManager, error) {
	result := &AudioManager{}
	var err error
	result.audioContext, err = audio.NewContext(44100)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return result, nil
}

type AudioManager struct {
	audioContext *audio.Context // The Audio context
	bgmAudio     *audio.Player  // The audio player
	lastBgm      string
	sfxVolume    float64
	bgmVolume    float64
}

func (eap *AudioManager) PlayBGM(song string) {
	if eap.lastBgm == song {
		return
	}
	eap.lastBgm = song
	if song == "" && eap.bgmAudio != nil && eap.bgmAudio.IsPlaying() {
		_ = eap.bgmAudio.Pause()
		return
	}

	if eap.bgmAudio != nil {
		err := eap.bgmAudio.Close()
		if err != nil {
			log.Panic(err)
		}
	}
	audioData, err := d2asset.LoadFile(song)
	if err != nil {
		panic(err)
	}
	d, err := wav.Decode(eap.audioContext, audio.BytesReadSeekCloser(audioData))
	if err != nil {
		log.Fatal(err)
	}
	s := audio.NewInfiniteLoop(d, d.Length())
	eap.bgmAudio, err = audio.NewPlayer(eap.audioContext, s)
	if err != nil {
		log.Fatal(err)
	}
	eap.bgmAudio.SetVolume(eap.bgmVolume)
	// Play the infinite-length stream. This never ends.
	err = eap.bgmAudio.Rewind()
	if err != nil {
		panic(err)
	}
	err = eap.bgmAudio.Play()
	if err != nil {
		panic(err)
	}
}

func (eap *AudioManager) LoadSoundEffect(sfx string) (d2audio.SoundEffect, error) {
	result := CreateSoundEffect(sfx, eap.audioContext, eap.sfxVolume) // TODO: Split
	return result, nil
}

func (eap *AudioManager) SetVolumes(bgmVolume, sfxVolume float64) {
	eap.sfxVolume = sfxVolume
	eap.bgmVolume = bgmVolume
}
