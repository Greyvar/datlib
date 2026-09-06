package tiled

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEntityVisualCatalogChest(t *testing.T) {
	dir := filepath.Clean("../../server/dat/entdefs")
	if _, err := os.Stat(filepath.Join(dir, "general-entities.tsx")); err != nil {
		t.Skip(err)
	}

	cat, err := LoadEntityVisualCatalog(dir)
	if err != nil {
		t.Fatal(err)
	}

	vis, ok := cat.Lookup("chest")
	if !ok {
		t.Fatal("expected chest type")
	}
	if vis.Texture != "general-entities.png" {
		t.Fatalf("texture = %q", vis.Texture)
	}
	if len(vis.Frames) != 1 || vis.Frames[0] != 10 {
		t.Fatalf("chest frames = %#v", vis.Frames)
	}

	open, ok := cat.ResolveState("chest", "open", "closed")
	if !ok {
		t.Fatal("expected chest_open")
	}
	if len(open.Frames) < 1 {
		t.Fatalf("open frames = %#v", open.Frames)
	}

	player, ok := cat.ResolveState("player", "idle", "idle")
	if !ok {
		t.Fatal("expected player idle")
	}
	if player.Texture != "ss_playerBob.png" {
		t.Fatalf("player texture = %q", player.Texture)
	}
	if len(player.Frames) != 8 || len(player.Holds) != 8 {
		t.Fatalf("idle frames/holds = %#v %#v", player.Frames, player.Holds)
	}

	fallback, ok := cat.Lookup(DefaultEntityVisualType)
	if !ok {
		t.Fatal("expected construct_entity in general-entities.tsx")
	}
	if fallback.FrameIndex != 9 {
		t.Fatalf("construct_entity frame = %d, want 9", fallback.FrameIndex)
	}

	signpost, ok := cat.Lookup("signpost")
	if !ok {
		t.Fatal("expected signpost type")
	}
	if len(signpost.Collision) < 1 {
		t.Fatal("expected signpost collision rects from tileset")
	}
	if signpost.CollisionType != CollisionTouch && signpost.CollisionType != "" {
		t.Fatalf("signpost collision type = %q", signpost.CollisionType)
	}

	pot, ok := cat.ResolveState("pot", "potlike", "potlike")
	if !ok {
		t.Fatal("expected pot visual")
	}
	if pot.Texture != "general-entities.png" {
		t.Fatalf("pot texture = %q, want general-entities.png (not missing pot.png)", pot.Texture)
	}
	if len(pot.Frames) != 1 || pot.Frames[0] != 23 {
		t.Fatalf("pot frames = %#v, want [23]", pot.Frames)
	}
	if len(pot.Collision) < 1 {
		t.Fatal("expected pot collision from general-entities tile")
	}

	key, ok := cat.Lookup("key")
	if !ok {
		t.Fatal("expected key")
	}
	if key.CollisionType != CollisionPickup {
		t.Fatalf("key collision type = %q, want pickup", key.CollisionType)
	}

	btn, ok := cat.Lookup("pressureButton")
	if !ok {
		t.Fatal("expected pressureButton")
	}
	if btn.CollisionType != CollisionPressure {
		t.Fatalf("button collision type = %q", btn.CollisionType)
	}

	missing, ok := cat.ResolveState("noSuchEnt", "default", "default")
	if !ok {
		t.Fatal("expected ResolveState fallback to construct_entity")
	}
	if missing.Type != DefaultEntityVisualType {
		t.Fatalf("fallback type = %q", missing.Type)
	}
	if _, ok := cat.ResolveStateExact("noSuchEnt", "default", "default"); ok {
		t.Fatal("ResolveStateExact should not fall back")
	}
}
