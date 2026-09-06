package tiled

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"
)

func TestParseTSXCollisionRects(t *testing.T) {
	path := filepath.Clean("../../res/img/textures/tilesets/grass_terains.tsx")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("grass_terains.tsx not available: %v", err)
	}

	var ts tsxTileset
	if err := xml.Unmarshal(data, &ts); err != nil {
		t.Fatalf("parse tsx: %v", err)
	}

	var tile42 *tsxTile
	for i := range ts.Tiles {
		if ts.Tiles[i].ID == 42 {
			tile42 = &ts.Tiles[i]
			break
		}
	}
	if tile42 == nil {
		t.Fatalf("expected tile id 42 in grass_terains.tsx")
	}

	rects := collisionRectsFromTSXTile(*tile42)
	if len(rects) != 1 {
		t.Fatalf("expected 1 collision rect on tile 42, got %d", len(rects))
	}
	if rects[0].X != 0 || rects[0].Y != 7 || rects[0].W != 10 || rects[0].H != 9 {
		t.Fatalf("unexpected rect: %+v", rects[0])
	}
}

func TestTransformCollisionRectFlipH(t *testing.T) {
	r := CollisionRect{X: 0, Y: 7, W: 10, H: 9}
	got := TransformCollisionRect(r, 16, 16, true, false)
	if got.X != 6 || got.Y != 7 || got.W != 10 || got.H != 9 {
		t.Fatalf("flipH transform = %+v, want x=6", got)
	}
}

func TestLoadTilesetCollisionIntoCatalog(t *testing.T) {
	tsxPath := filepath.Clean("../../res/img/textures/tilesets/grass_terains.tsx")
	if _, err := os.Stat(tsxPath); err != nil {
		t.Skipf("grass_terains.tsx not available: %v", err)
	}

	ref := tilesetRef{FirstGID: 1, Source: "grass_terains.tsx"}
	catalog := &TilesetCatalog{base: make(map[uint32]tileInfo)}
	if err := catalog.loadExternalTileset(filepath.Dir(tsxPath), ref); err != nil {
		t.Fatalf("loadExternalTileset: %v", err)
	}

	info, ok := catalog.Lookup(1 + 42)
	if !ok {
		t.Fatalf("expected gid for tile 42")
	}
	if !info.CollisionFromTSX {
		t.Fatalf("expected CollisionFromTSX on parsed tile")
	}
	if len(info.Collision) != 1 {
		t.Fatalf("expected collision rects on tile 42, got %d", len(info.Collision))
	}
}
