package worldfiles

import (
	"testing"

	"github.com/greyvar/datlib/gridfiles"
)

func TestRandomPlayerSpawn(t *testing.T) {
	world := &World{
		SpawnGrid: "a.tmj",
		Grids: map[string]*gridfiles.Grid{
			"a.tmj": {
				Filename: "a.tmj",
				SpawnPoints: []gridfiles.SpawnPoint{
					{X: 16, Y: 32},
					{X: 48, Y: 64},
				},
			},
			"b.tmj": {
				Filename: "b.tmj",
				SpawnPoints: []gridfiles.SpawnPoint{
					{X: 0, Y: 0},
				},
			},
		},
	}

	if SpawnPointCount(world) != 3 {
		t.Fatalf("expected 3 spawn points, got %d", SpawnPointCount(world))
	}

	seen := map[string]bool{}
	for i := 0; i < 30; i++ {
		gridID, x, y, ok := RandomPlayerSpawn(world)
		if !ok {
			t.Fatalf("expected spawn location")
		}
		switch {
		case gridID == "a.tmj" && x == 16 && y == 32:
			seen["a1"] = true
		case gridID == "a.tmj" && x == 48 && y == 64:
			seen["a2"] = true
		case gridID == "b.tmj" && x == 0 && y == 0:
			seen["b1"] = true
		default:
			t.Fatalf("unexpected spawn %q (%d,%d)", gridID, x, y)
		}
	}

	if len(seen) < 2 {
		t.Fatalf("expected random selection to hit multiple spawn points, got %v", seen)
	}
}

func TestFallbackPlayerSpawn(t *testing.T) {
	world := &World{
		SpawnGrid: "start.tmj",
		Grids: map[string]*gridfiles.Grid{
			"start.tmj": {
				RowCount: 16,
				ColCount: 16,
			},
		},
	}

	gridID, x, y, ok := PlayerSpawnLocation(world)
	if !ok {
		t.Fatalf("expected fallback spawn")
	}
	if gridID != "start.tmj" || x != 128 || y != 128 {
		t.Fatalf("fallback spawn = (%q,%d,%d), want (start.tmj,128,128)", gridID, x, y)
	}
}

func TestPlayerSpawnLocationPrefersMarkers(t *testing.T) {
	world := &World{
		SpawnGrid: "start.tmj",
		Grids: map[string]*gridfiles.Grid{
			"start.tmj": {
				RowCount: 16,
				ColCount: 16,
				SpawnPoints: []gridfiles.SpawnPoint{
					{X: 32, Y: 48},
				},
			},
		},
	}

	for i := 0; i < 10; i++ {
		gridID, x, y, ok := PlayerSpawnLocation(world)
		if !ok || gridID != "start.tmj" || x != 32 || y != 48 {
			t.Fatalf("expected spawn marker at (32,48), got (%q,%d,%d,%v)", gridID, x, y, ok)
		}
	}
}
