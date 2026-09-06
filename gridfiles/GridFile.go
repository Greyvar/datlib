package gridfiles

type LayerKind string

const (
	LayerKindTerrain LayerKind = "terrain"
	LayerKindObject  LayerKind = "object"
)

type Tile struct {
	Row uint32
	Col uint32
	Rot int32
	FlipH bool `yaml:"flipH"`
	FlipV bool `yaml:"flipV"`
	Traversable bool
	Texture string
	AtlasKey string `yaml:"-"`
	FrameIndex int32 `yaml:"-"`
	Collision []CollisionRect
	CollisionFromTSX bool `yaml:"-"`
	TeleportDst string `yaml:"teleportDst,omitempty"`
	TeleportDir string `yaml:"teleportDir,omitempty"`
	TeleportX int `yaml:"teleportX,omitempty"`
	TeleportY int `yaml:"teleportY,omitempty"`
	Hidden bool `yaml:"-"`
}

type TileLayer struct {
	Name  string
	Kind  LayerKind
	Tiles map[uint32]map[uint32]*Tile
}

func NewTileLayer(name string, kind LayerKind) TileLayer {
	return TileLayer{
		Name:  name,
		Kind:  kind,
		Tiles: make(map[uint32]map[uint32]*Tile),
	}
}

func (l *TileLayer) SetTile(row, col uint32, tile *Tile) {
	if l.Tiles[row] == nil {
		l.Tiles[row] = make(map[uint32]*Tile)
	}
	l.Tiles[row][col] = tile
}

func (l *TileLayer) TileAt(row, col uint32) *Tile {
	if l.Tiles[row] == nil {
		return nil
	}
	return l.Tiles[row][col]
}

type GridTileset struct {
	Key        string
	ImagePath  string
	TileWidth  int
	TileHeight int
	Columns    int
}

type GridFileEntityInstance struct {
	Row        uint32
	Col        uint32
	Definition string
	GridID     string `yaml:"id"` // Ignored from map file for now. Overwritten by server. Will need to change this.

	// Properties are Tiled object custom properties (e.g. msg on signs).
	Properties map[string]string `yaml:"-" json:"-"`

	Spawned bool   `yaml:"-"`
	State   string `yaml:"-"`
}

// SpawnPoint marks a player spawn location loaded from a Tiled spawn tile/object.
type SpawnPoint struct {
	Row uint32
	Col uint32
	X   int32
	Y   int32
}

type GridSerializable struct {
	ColCount uint32
	RowCount uint32
	Tiles []*Tile
	Entities []GridFileEntityInstance
	LastEntityId string `yaml:"lastEntityId"`
}

type Grid struct {
	Filename string
	ColCount uint32
	RowCount uint32
	Tiles map[uint32]map[uint32]*Tile
	Layers []TileLayer
	Entities []GridFileEntityInstance
	SpawnPoints []SpawnPoint
	LastEntityId string `yaml:"lastEntityId"`
	Tilesets []GridTileset
}

func (g *Grid) BuildEmpty() {
	g.Tiles = make(map[uint32]map[uint32]*Tile)

	for row := uint32(0); row < g.RowCount; row++ {
		g.Tiles[row] = make(map[uint32]*Tile)
	}
}

// TopTileAt returns the topmost non-empty tile at a cell across all layers.
// Upper layers take precedence. In-bounds empty cells return a traversable tile.
// Out-of-bounds coordinates return nil.
func (g *Grid) TopTileAt(row, col uint32) *Tile {
	if row >= g.RowCount || col >= g.ColCount {
		return nil
	}

	for i := len(g.Layers) - 1; i >= 0; i-- {
		if tile := g.Layers[i].TileAt(row, col); tile != nil {
			return tile
		}
	}

	if g.Tiles != nil {
		if rowTiles, ok := g.Tiles[row]; ok {
			if tile := rowTiles[col]; tile != nil {
				return tile
			}
		}
	}

	return &Tile{
		Row:         row,
		Col:         col,
		Traversable: true,
	}
}

// SyncTopTiles updates the flat Tiles map to the topmost tile per cell.
func (g *Grid) SyncTopTiles() {
	if g.Tiles == nil {
		g.BuildEmpty()
	}

	for _, pos := range g.CellIterator() {
		top := g.TopTileAt(pos.Row, pos.Col)
		if top != nil {
			g.Tiles[pos.Row][pos.Col] = top
			continue
		}

		g.Tiles[pos.Row][pos.Col] = &Tile{
			Row:         pos.Row,
			Col:         pos.Col,
			Traversable: true,
		}
	}
}

func (g *Grid) Build() {
	g.Tiles = make(map[uint32]map[uint32]*Tile)

	for row := uint32(0); row < g.RowCount; row++ {
		g.Tiles[row] = make(map[uint32]*Tile)

		for col := uint32(0); col < g.ColCount; col++ {
			t := &Tile{
				Row: row,
				Col: col,
				Texture: "water",
			}

			g.Tiles[row][col] = t
		}
	}
}

type position struct {
	Row uint32
	Col uint32
}

func (g *Grid) CellIterator() []*position {
	ret := make([]*position, 0)

	for row := uint32(0); row < g.RowCount; row++ {
		for col := uint32(0); col < g.ColCount; col++ {
			ret = append(ret, &position{
				Row: row,
				Col: col,
			})
		}
	}

	return ret
}
