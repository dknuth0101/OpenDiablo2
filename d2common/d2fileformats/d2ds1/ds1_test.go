package d2ds1

import (
	"testing"

	"log"
	"os"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2path"
)

func exampleData() *DS1 {
	exampleFloor1 := Tile{
		// common fields
		tileCommonFields: tileCommonFields{
			Prop1:       2,
			Sequence:    89,
			Unknown1:    123,
			Style:       20,
			Unknown2:    53,
			HiddenBytes: 1,
			RandomIndex: 2,
			YAdjust:     21,
		},
		tileFloorShadowFields: tileFloorShadowFields{
			Animated: false,
		},
	}

	exampleFloor2 := Tile{
		// common fields
		tileCommonFields: tileCommonFields{
			Prop1:       3,
			Sequence:    89,
			Unknown1:    213,
			Style:       28,
			Unknown2:    53,
			HiddenBytes: 7,
			RandomIndex: 3,
			YAdjust:     28,
		},
		tileFloorShadowFields: tileFloorShadowFields{
			Animated: true,
		},
	}

	exampleWall1 := Tile{
		// common fields
		tileCommonFields: tileCommonFields{
			Prop1:       3,
			Sequence:    89,
			Unknown1:    213,
			Style:       28,
			Unknown2:    53,
			HiddenBytes: 7,
			RandomIndex: 3,
			YAdjust:     28,
		},
		tileWallFields: tileWallFields{
			Type: d2enum.TileRightWall,
		},
	}

	exampleWall2 := Tile{
		// common fields
		tileCommonFields: tileCommonFields{
			Prop1:       3,
			Sequence:    93,
			Unknown1:    193,
			Style:       17,
			Unknown2:    13,
			HiddenBytes: 1,
			RandomIndex: 1,
			YAdjust:     22,
		},
		tileWallFields: tileWallFields{
			Type: d2enum.TileLeftWall,
		},
	}

	exampleShadow := Tile{
		// common fields
		tileCommonFields: tileCommonFields{
			Prop1:       3,
			Sequence:    93,
			Unknown1:    173,
			Style:       17,
			Unknown2:    12,
			HiddenBytes: 1,
			RandomIndex: 1,
			YAdjust:     22,
		},
		tileFloorShadowFields: tileFloorShadowFields{
			Animated: false,
		},
	}

	result := &DS1{
		ds1Layers: &ds1Layers{
			width:  20,
			height: 80,
			Floors: layerGroup{
				// number of floors (one floor)
				{
					// tile grid = []tileRow
					tiles: tileGrid{
						// tile rows = []Tile
						// 2x2 tiles
						{
							exampleFloor1,
							exampleFloor2,
						},
						{
							exampleFloor2,
							exampleFloor1,
						},
					},
				},
			},
			Walls: layerGroup{
				// number of walls (two floors)
				{
					// tile grid = []tileRow
					tiles: tileGrid{
						// tile rows = []Tile
						// 2x2 tiles
						{
							exampleWall1,
							exampleWall2,
						},
						{
							exampleWall2,
							exampleWall1,
						},
					},
				},
				{
					// tile grid = []tileRow
					tiles: tileGrid{
						// tile rows = []Tile
						// 2x2 tiles
						{
							exampleWall1,
							exampleWall2,
						},
						{
							exampleWall2,
							exampleWall1,
						},
					},
				},
			},
			Shadows: layerGroup{
				// number of shadows (always 1)
				{
					// tile grid = []tileRow
					tiles: tileGrid{
						// tile rows = []Tile
						// 2x2 tiles
						{
							exampleShadow,
							exampleShadow,
						},
						{
							exampleShadow,
							exampleShadow,
						},
					},
				},
			},
		},
		Files: []string{"a.dt1", "bfile.dt1"},
		Objects: []Object{
			{0, 0, 0, 0, 0, nil},
			{0, 1, 0, 0, 0, []d2path.Path{{}}},
			{0, 2, 0, 0, 0, nil},
			{0, 3, 0, 0, 0, nil},
		},
		substitutionGroups: nil,
		version:            17,
		Act:                1,
		substitutionType:   0,
		unknown1:           make([]byte, 8),
		unknown2:           20,
	}

	return result
}

func TestDS1_Load(t *testing.T) {
	testFile, fileErr := os.Open("testdata/testdata.ds1")
	if fileErr != nil {
		t.Error("cannot open test data file")
		return
	}

	data := make([]byte, 0)
	buf := make([]byte, 16)

	for {
		numRead, err := testFile.Read(buf)

		data = append(data, buf[:numRead]...)

		if err != nil {
			break
		}
	}

	_, loadErr := Unmarshal(data)
	if loadErr != nil {
		t.Error(loadErr)
	}

	err := testFile.Close()
	if err != nil {
		t.Fail()
		log.Print(err)
	}

}
