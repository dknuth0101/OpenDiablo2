package d2interface

type AudioManagerType int

const (
	AudioManagerNone AudioManagerType = iota
	AudioManagerEbiten
)

type AudioManager interface {
	Initialize(AudioManagerType) error
	PlayBGM(song string)
	LoadSoundEffect(sfx string) (SoundEffect, error)
	SetVolumes(bgmVolume, sfxVolume float64)
}
