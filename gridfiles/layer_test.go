package gridfiles

import "testing"

func TestTopTileAtLayerPrecedence(t *testing.T) {
	grid := &Grid{
		RowCount: 2,
		ColCount: 2,
		Layers: []TileLayer{
			func() TileLayer {
				l := NewTileLayer("biome", LayerKindTerrain)
				l.SetTile(0, 0, &Tile{Row: 0, Col: 0, Texture: "grass.png"})
				l.SetTile(0, 1, &Tile{Row: 0, Col: 1, Texture: "sand.png"})
				return l
			}(),
			func() TileLayer {
				l := NewTileLayer("terrain", LayerKindTerrain)
				l.SetTile(0, 0, &Tile{Row: 0, Col: 0, Texture: "fence.png"})
				return l
			}(),
			func() TileLayer {
				l := NewTileLayer("objects", LayerKindObject)
				l.SetTile(1, 0, &Tile{Row: 1, Col: 0, Texture: "stone.png"})
				return l
			}(),
		},
	}
	grid.BuildEmpty()

	top := grid.TopTileAt(0, 0)
	if top == nil || top.Texture != "fence.png" {
		t.Fatalf("expected terrain layer to override biome at 0,0, got %+v", top)
	}

	fallback := grid.TopTileAt(0, 1)
	if fallback == nil || fallback.Texture != "sand.png" {
		t.Fatalf("expected biome to show through empty terrain at 0,1, got %+v", fallback)
	}

	objectTop := grid.TopTileAt(1, 0)
	if objectTop == nil || objectTop.Texture != "stone.png" {
		t.Fatalf("expected object layer tile at 1,0, got %+v", objectTop)
	}

	if grid.TopTileAt(2, 0) != nil {
		t.Fatalf("expected nil for out-of-bounds row")
	}

	empty := grid.TopTileAt(1, 1)
	if empty == nil || !empty.Traversable || empty.Texture != "" {
		t.Fatalf("expected traversable empty in-bounds cell at 1,1, got %+v", empty)
	}
}
