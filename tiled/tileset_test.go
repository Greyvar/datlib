package tiled

import (
	"path/filepath"
	"testing"
)

func TestLoadTilesetsFromMapMissingExternalFallsBack(t *testing.T) {
	mapPath := filepath.Join(t.TempDir(), "map.tmj")
	refs := []tilesetRef{
		{
			FirstGID: 1,
			Source:   "missing_tileset.tsx",
		},
	}

	catalog, err := LoadTilesetsFromMap(mapPath, refs)
	if err != nil {
		t.Fatalf("LoadTilesetsFromMap: %v", err)
	}

	info, ok := catalog.Lookup(1)
	if !ok {
		t.Fatal("expected fallback entry for firstgid")
	}
	if info.Texture != DefaultTexture {
		t.Fatalf("texture = %q, want %q", info.Texture, DefaultTexture)
	}
	if !info.CollisionFromTSX {
		t.Fatal("expected CollisionFromTSX so fallback tiles stay walkable")
	}

	info, ok = catalog.Lookup(500)
	if !ok {
		t.Fatal("expected fallback entry for mid-range gid")
	}
	if info.Texture != DefaultTexture {
		t.Fatalf("texture = %q, want %q", info.Texture, DefaultTexture)
	}
}
