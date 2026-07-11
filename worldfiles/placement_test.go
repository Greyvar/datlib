package worldfiles

import (
	"testing"

	"github.com/greyvar/datlib/gridfiles"
)

func TestPlacementForGridFromWorldMaps(t *testing.T) {
	def := &Definition{
		Maps: []MapPlacement{
			{FileName: "1.1.tmj", Width: 256, Height: 256, X: 256, Y: 256},
			{FileName: "1.2.tmj", Width: 256, Height: 256, X: 512, Y: 256},
		},
	}

	placement, ok := PlacementForGrid(def, nil, "1.2.tmj")
	if !ok {
		t.Fatalf("expected placement for 1.2.tmj")
	}
	if placement.X != 512 || placement.Y != 256 {
		t.Fatalf("unexpected placement: %+v", placement)
	}
}

func TestScrollDeltaBetween(t *testing.T) {
	def := &Definition{
		Maps: []MapPlacement{
			{FileName: "1.1.tmj", Width: 256, Height: 256, X: 256, Y: 256},
			{FileName: "0.1.tmj", Width: 256, Height: 256, X: 256, Y: 0},
			{FileName: "1.2.tmj", Width: 256, Height: 256, X: 512, Y: 256},
		},
	}

	dx, dy, ok := ScrollDeltaBetween(def, nil, "1.1.tmj", "1.2.tmj")
	if !ok || dx != 256 || dy != 0 {
		t.Fatalf("right scroll delta = (%d,%d), want (256,0)", dx, dy)
	}

	dx, dy, ok = ScrollDeltaBetween(def, nil, "1.1.tmj", "0.1.tmj")
	if !ok || dx != 0 || dy != -256 {
		t.Fatalf("up scroll delta = (%d,%d), want (0,-256)", dx, dy)
	}
}

func TestPlacementForGridDerivedFromFilename(t *testing.T) {
	grids := map[string]*gridfiles.Grid{
		"2.0.tmj": {
			Filename: "2.0.tmj",
			ColCount: 16,
			RowCount: 16,
		},
	}

	placement, ok := PlacementForGrid(nil, grids, "2.0.tmj")
	if !ok {
		t.Fatalf("expected derived placement")
	}
	if placement.X != 0 || placement.Y != 512 {
		t.Fatalf("derived placement = %+v, want x=0 y=512", placement)
	}
}
