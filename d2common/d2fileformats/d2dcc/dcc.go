package d2dcc

import (
	"errors"

	bitstream2 "github.com/gravestench/bitstream"
)

const dccFileSignature = 0x74
const directionOffsetMultiplier = 8

// DCC represents a DCC file.
type DCC struct {
	Signature          int
	Version            int
	NumberOfDirections int
	FramesPerDirection int
	Directions         []*DCCDirection
	directionOffsets   []int
	fileData           []byte
}

// Load loads a DCC file.
func Load(fileData []byte) (*DCC, error) {
	result := &DCC{
		fileData: fileData,
	}

	var bitstream = bitstream2.FromBytes(fileData...)

	result.Signature = bitstream.ReadBits(8).AsInt()

	if result.Signature != dccFileSignature {
		return nil, errors.New("signature expected to be 0x74 but it is not")
	}

	result.Version = int(bitstream.ReadBits(8).AsByte())
	result.NumberOfDirections = int(bitstream.ReadBits(8).AsByte())
	result.FramesPerDirection = int(bitstream.ReadBits(32).AsInt32())

	result.Directions = make([]*DCCDirection, result.NumberOfDirections)

	if bitstream.ReadBits(32).AsInt32() != 1 {
		return nil, errors.New("this value isn't 1. It has to be 1")
	}

	bitstream.ReadBits(32).AsInt32() // TotalSizeCoded

	result.directionOffsets = make([]int, result.NumberOfDirections)

	for i := 0; i < result.NumberOfDirections; i++ {
		result.directionOffsets[i] = int(bitstream.ReadBits(32).AsInt32())
		result.Directions[i] = result.decodeDirection(i)
	}

	return result, nil
}

// decodeDirection decodes and returns the given direction
func (d *DCC) decodeDirection(direction int) *DCCDirection {
	//return CreateDCCDirection(d2datautils.CreateBitMuncher(d.fileData,
	//	d.directionOffsets[direction]*directionOffsetMultiplier), d)
	bs := bitstream2.FromBytes(d.fileData...)
	bs.SetPosition(d.directionOffsets[direction] * directionOffsetMultiplier)

	return CreateDCCDirection(bs, d)
}

// Clone creates a copy of the DCC
func (d *DCC) Clone() *DCC {
	clone := *d
	copy(clone.directionOffsets, d.directionOffsets)
	copy(clone.fileData, d.fileData)
	clone.Directions = make([]*DCCDirection, len(d.Directions))

	for i := range d.Directions {
		cloneDirection := *d.Directions[i]
		clone.Directions = append(clone.Directions, &cloneDirection)
	}

	return &clone
}
