package gridfiles

type Tile struct {
	Row uint32
	Col uint32
	Rot int32
	FlipH bool `yaml:"flipH"`
	FlipV bool `yaml:"flipV"`
	Traversable bool
	Texture string
}

type GridFileEntityInstance struct {
	Row uint32
	Col uint32
	Definition string
	GridID string `yaml:"id"` // Ignored from map file for now. Overwritten by server. Will need to change this.

	Spawned bool `yaml:"-"`
	State string `yaml:"-"`
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
	Entities []GridFileEntityInstance
	LastEntityId string `yaml:"lastEntityId"`
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
