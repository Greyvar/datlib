package tiled

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDrawFromGameplaySpawnTile(t *testing.T) {
	tsxPath := filepath.Clean("../../server/dat/entdefs/gameplay.tsx")
	if _, err := os.Stat(tsxPath); err != nil {
		t.Skipf("gameplay.tsx not available: %v", err)
	}

	ref := tilesetRef{FirstGID: 477, Source: "gameplay.tsx"}
	catalog := &TilesetCatalog{base: make(map[uint32]tileInfo)}
	if err := catalog.loadExternalTileset(filepath.Dir(tsxPath), ref); err != nil {
		t.Fatalf("loadExternalTileset: %v", err)
	}

	spawn, ok := catalog.Lookup(477)
	if !ok {
		t.Fatalf("expected spawn tile at gid 477")
	}
	if !spawn.Hidden {
		t.Fatalf("expected spawn tile hidden (draw=false)")
	}
	if spawn.Type != "spawn" {
		t.Fatalf("expected spawn type, got %q", spawn.Type)
	}

	visible, ok := catalog.Lookup(478)
	if !ok {
		t.Fatalf("expected atlas slot at gid 478")
	}
	if visible.Hidden {
		t.Fatalf("expected default visible tile without draw property")
	}
}

func TestDrawFromProperties(t *testing.T) {
	if drawFromProperties([]propertyEntry{{Name: "draw", Type: "bool", Value: "false"}}) {
		t.Fatalf("expected draw=false")
	}
	if !drawFromProperties([]propertyEntry{{Name: "draw", Type: "bool", Value: "true"}}) {
		t.Fatalf("expected draw=true")
	}
	if !drawFromProperties(nil) {
		t.Fatalf("expected default draw=true")
	}
}
