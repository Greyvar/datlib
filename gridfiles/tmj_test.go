package gridfiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadGridTMJ(t *testing.T) {
	datDir := os.Getenv("GREYVAR_DAT_DIR")
	if datDir == "" {
		datDir = filepath.Clean("../../server/dat")
	}

	path := filepath.Join(datDir, "worlds/startWorldTmj/ow-p01-n01-o0000.tmj")
	grid, err := ReadGridTMJ(path)
	if err != nil {
		t.Fatalf("ReadGridTMJ: %v", err)
	}

	if grid.ColCount != 16 || grid.RowCount != 16 {
		t.Fatalf("expected 16x16 grid, got %dx%d", grid.ColCount, grid.RowCount)
	}

	if grid.Tiles[0][0] == nil || grid.Tiles[0][0].Texture == "" {
		t.Fatalf("expected populated tile at 0,0")
	}

	if len(grid.Entities) == 0 {
		t.Fatalf("expected at least one entity from object layer")
	}

	if len(grid.Tilesets) == 0 {
		t.Fatalf("expected atlas tilesets from grass.tsx")
	}

	foundGrass := false
	for _, tileset := range grid.Tilesets {
		if tileset.Key == "grass" {
			foundGrass = true
			if tileset.ImagePath != "img/textures/tilesets/grass.png" {
				t.Fatalf("expected grass tileset image path, got %q", tileset.ImagePath)
			}
			if tileset.Columns != 20 {
				t.Fatalf("expected grass tileset columns 20, got %d", tileset.Columns)
			}
		}
	}
	if !foundGrass {
		t.Fatalf("expected grass tileset in grid.Tilesets")
	}

	atlasTiles := 0
	highAtlasTiles := 0
	for _, pos := range grid.CellIterator() {
		tile := grid.Tiles[pos.Row][pos.Col]
		if tile.AtlasKey == "grass" {
			atlasTiles++
			if tile.FrameIndex >= 300 {
				highAtlasTiles++
			}
		}
	}
	if atlasTiles == 0 {
		t.Fatalf("expected atlas-backed tiles in grid")
	}
	if highAtlasTiles == 0 {
		t.Fatalf("expected high-index atlas tiles (tilecount fix)")
	}

	foundTV := false
	for _, ent := range grid.Entities {
		if ent.Definition == "tv" {
			foundTV = true
			if ent.Row != 6 || ent.Col != 6 {
				t.Fatalf("expected tv at row 6 col 6, got row %d col %d", ent.Row, ent.Col)
			}
		}
	}

	if !foundTV {
		t.Fatalf("expected tv entity from gid 403")
	}
}

func TestReadGridDispatch(t *testing.T) {
	datDir := os.Getenv("GREYVAR_DAT_DIR")
	if datDir == "" {
		datDir = filepath.Clean("../../server/dat")
	}

	tmjPath := filepath.Join(datDir, "worlds/startWorldTmj/ow-p01-n01-o0000.tmj")
	if _, err := ReadGrid(tmjPath); err != nil {
		t.Fatalf("ReadGrid tmj dispatch: %v", err)
	}

	yamlPath := "../worldfiles/testdata/dat/worlds/yamlWorld_dev/grids/0.0.grid"
	if _, err := ReadGrid(yamlPath); err != nil {
		t.Fatalf("ReadGrid yaml dispatch: %v", err)
	}
}
