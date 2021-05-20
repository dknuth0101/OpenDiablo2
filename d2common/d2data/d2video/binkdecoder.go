package d2video

import (
	"errors"
	"log"

	bitstream2 "github.com/gravestench/bitstream"
)

// BinkVideoMode is the video mode type
type BinkVideoMode uint32

const (
	// BinkVideoModeNormal is a normal video
	BinkVideoModeNormal BinkVideoMode = iota

	// BinkVideoModeHeightDoubled is a height-doubled video
	BinkVideoModeHeightDoubled

	// BinkVideoModeHeightInterlaced is a height-interlaced video
	BinkVideoModeHeightInterlaced

	// BinkVideoModeWidthDoubled is a width-doubled video
	BinkVideoModeWidthDoubled

	// BinkVideoModeWidthAndHeightDoubled is a width and height-doubled video
	BinkVideoModeWidthAndHeightDoubled

	// BinkVideoModeWidthAndHeightInterlaced is a width and height interlaced video
	BinkVideoModeWidthAndHeightInterlaced
)

const (
	numHeaderBytes            = 3
	bikHeaderStr              = "BIK"
	numAudioTrackUnknownBytes = 2
)

// BinkAudioAlgorithm represents the type of bink audio algorithm
type BinkAudioAlgorithm uint32

const (
	// BinkAudioAlgorithmFFT is the FTT audio algorithm
	BinkAudioAlgorithmFFT BinkAudioAlgorithm = iota

	// BinkAudioAlgorithmDCT is the DCT audio algorithm
	BinkAudioAlgorithmDCT
)

// BinkAudioTrack represents an audio track
type BinkAudioTrack struct {
	AudioChannels     uint16
	AudioSampleRateHz uint16
	Stereo            bool
	Algorithm         BinkAudioAlgorithm
	AudioTrackID      uint32
}

// BinkDecoder represents the bink decoder
type BinkDecoder struct {
	AudioTracks           []BinkAudioTrack
	FrameIndexTable       []uint32
	bitstream             *bitstream2.BitStream
	fileSize              uint32
	numberOfFrames        uint32
	largestFrameSizeBytes uint32
	VideoWidth            uint32
	VideoHeight           uint32
	FPS                   uint32
	FrameTimeMS           uint32
	VideoMode             BinkVideoMode
	frameIndex            uint32
	videoCodecRevision    byte
	HasAlphaPlane         bool
	Grayscale             bool

	// Mask bit 0, as this is defined as a keyframe

}

// CreateBinkDecoder returns a new instance of the bink decoder
func CreateBinkDecoder(source []byte) (*BinkDecoder, error) {
	result := &BinkDecoder{
		bitstream: bitstream2.FromBytes(source...),
	}

	err := result.loadHeaderInformation()

	return result, err
}

// GetNextFrame gets the next frame
func (v *BinkDecoder) GetNextFrame() error {
	//nolint:gocritic // v.bitstream.SetPosition(uint64(v.FrameIndexTable[i] & 0xFFFFFFFE))
	lengthOfAudioPackets := v.bitstream.ReadBits(32).AsUInt32()
	samplesInPacket := v.bitstream.ReadBits(32).AsUInt32()

	v.bitstream.ReadBits(8*int(lengthOfAudioPackets) - 4) //nolint:gomnd // decode magic

	log.Printf("Frame %d:\tSamp: %d", v.frameIndex, samplesInPacket)

	v.frameIndex++

	return nil
}

//nolint:gomnd,funlen,gocyclo // Decoder magic, can't help the long function length for now
func (v *BinkDecoder) loadHeaderInformation() error {
	v.bitstream.SetPosition(0)

	headerBytes := v.bitstream.ReadBits(numHeaderBytes).AsByte()

	if string(headerBytes) != bikHeaderStr {
		return errors.New("invalid header for bink video")
	}

	v.videoCodecRevision = v.bitstream.ReadBits(8).AsByte()
	v.fileSize = v.bitstream.ReadBits(32).AsUInt32()
	v.numberOfFrames = v.bitstream.ReadBits(32).AsUInt32()
	v.largestFrameSizeBytes = v.bitstream.ReadBits(32).AsUInt32()

	const numBytesToSkip = 4 // Number of frames again?

	v.bitstream.ReadBits(8 * numBytesToSkip)

	v.VideoWidth = v.bitstream.ReadBits(32).AsUInt32()
	v.VideoHeight = v.bitstream.ReadBits(32).AsUInt32()
	fpsDividend := v.bitstream.ReadBits(32).AsUInt32()
	fpsDivider := v.bitstream.ReadBits(32).AsUInt32()

	v.FPS = uint32(float32(fpsDividend) / float32(fpsDivider))
	v.FrameTimeMS = 1000 / v.FPS

	videoFlags := v.bitstream.ReadBits(32).AsUInt32()

	v.VideoMode = BinkVideoMode((videoFlags >> 28) & 0x0F)
	v.HasAlphaPlane = ((videoFlags >> 20) & 0x1) == 1
	v.Grayscale = ((videoFlags >> 17) & 0x1) == 1

	numberOfAudioTracks := v.bitstream.ReadBits(32).AsUInt32()

	v.AudioTracks = make([]BinkAudioTrack, numberOfAudioTracks)

	for i := 0; i < int(numberOfAudioTracks); i++ {
		v.bitstream.ReadBits(8 * numAudioTrackUnknownBytes)

		v.AudioTracks[i].AudioChannels = v.bitstream.ReadBits(16).AsUInt16()
	}

	for i := 0; i < int(numberOfAudioTracks); i++ {
		v.AudioTracks[i].AudioSampleRateHz = v.bitstream.ReadBits(16).AsUInt16()

		var flags uint16

		flags = v.bitstream.ReadBits(16).AsUInt16()

		v.AudioTracks[i].Stereo = ((flags >> 13) & 0x1) == 1
		v.AudioTracks[i].Algorithm = BinkAudioAlgorithm((flags >> 12) & 0x1)
	}

	for i := 0; i < int(numberOfAudioTracks); i++ {
		v.AudioTracks[i].AudioTrackID = v.bitstream.ReadBits(32).AsUInt32()
	}

	v.FrameIndexTable = make([]uint32, v.numberOfFrames+1)

	for i := 0; i < int(v.numberOfFrames+1); i++ {
		v.FrameIndexTable[i] = v.bitstream.ReadBits(32).AsUInt32()
	}

	return nil
}
