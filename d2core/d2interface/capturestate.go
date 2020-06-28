package d2interface

type CaptureState int

const (
	CaptureStateNone captureState = iota
	CaptureStateFrame
	CaptureStateGif
)
