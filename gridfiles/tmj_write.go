package gridfiles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/greyvar/datlib/tiled"
)

// TMJWriteOptions configures TMJ export.
type TMJWriteOptions struct {
	TileSize      int
	TexturePrefix string
}

type tmjEmbeddedTile struct {
	ID          int    `json:"id"`
	Image       string `json:"image"`
	ImageWidth  int    `json:"imagewidth"`
	ImageHeight int    `json:"imageheight"`
}

type tmjEmbeddedTileset struct {
	FirstGID   uint32            `json:"firstgid"`
	Name       string            `json:"name"`
	TileWidth  int               `json:"tilewidth"`
	TileHeight int               `json:"tileheight"`
	TileCount  int               `json:"tilecount"`
	Columns    int               `json:"columns"`
	Tiles      []tmjEmbeddedTile `json:"tiles"`
}

type tmjWriteDoc struct {
	CompressionLevel int                  `json:"compressionlevel"`
	Height           uint32               `json:"height"`
	Width            uint32               `json:"width"`
	Infinite         bool                 `json:"infinite"`
	Orientation      string               `json:"orientation"`
	RenderOrder      string               `json:"renderorder"`
	TiledVersion     string               `json:"tiledversion"`
	Type             string               `json:"type"`
	Version          string               `json:"version"`
	TileWidth        int                  `json:"tilewidth"`
	TileHeight       int                  `json:"tileheight"`
	NextLayerID      int                  `json:"nextlayerid"`
	NextObjectID     int                  `json:"nextobjectid"`
	Layers           []tmjLayer           `json:"layers"`
	Tilesets         []tmjEmbeddedTileset `json:"tilesets"`
}

func WriteGridTMJ(grid *Grid, filename string, opts TMJWriteOptions) error {
	if grid == nil {
		return fmt.Errorf("grid is nil")
	}

	tileSize := opts.TileSize
	if tileSize <= 0 {
		tileSize = 16
	}

	texturePrefix := opts.TexturePrefix
	if texturePrefix == "" {
		texturePrefix = "../../../../res/img/textures/tiles/"
	}
	if !strings.HasSuffix(texturePrefix, "/") {
		texturePrefix += "/"
	}

	textureIDs := map[string]uint32{}
	embeddedTiles := make([]tmjEmbeddedTile, 0)
	nextLocalID := 0

	registerTexture := func(texture string) uint32 {
		texture = strings.TrimSpace(texture)
		if texture == "" {
			texture = "construct.png"
		}
		if id, ok := textureIDs[texture]; ok {
			return id
		}

		localID := nextLocalID
		nextLocalID++
		gid := uint32(localID + 1)
		textureIDs[texture] = gid
		embeddedTiles = append(embeddedTiles, tmjEmbeddedTile{
			ID:          localID,
			Image:       texturePrefix + texture,
			ImageWidth:  tileSize,
			ImageHeight: tileSize,
		})
		return gid
	}

	data := make([]uint32, grid.ColCount*grid.RowCount)
	for row := uint32(0); row < grid.RowCount; row++ {
		for col := uint32(0); col < grid.ColCount; col++ {
			tile := grid.Tiles[row][col]
			if tile == nil {
				continue
			}

			gid := registerTexture(tile.Texture)
			raw := tiled.EncodeGID(tiled.GID{
				ID:    gid,
				FlipH: tile.FlipH,
				FlipV: tile.FlipV,
				Rot:   tile.Rot,
			})
			data[int(row*grid.ColCount+col)] = raw
		}
	}

	objects := make([]tmjMapObject, 0, len(grid.Entities))
	nextObjectID := 1
	for _, ent := range grid.Entities {
		objects = append(objects, tmjMapObject{
			ID:       nextObjectID,
			Type:     ent.Definition,
			X:        float64(ent.Col*uint32(tileSize)) + float64(tileSize)/2,
			Y:        float64(ent.Row*uint32(tileSize)) + float64(tileSize)/2,
			Width:    float64(tileSize),
			Height:   float64(tileSize),
			Rotation: 0,
			Visible:  true,
		})
		nextObjectID++
	}

	doc := tmjWriteDoc{
		CompressionLevel: -1,
		Height:           grid.RowCount,
		Width:            grid.ColCount,
		Infinite:         false,
		Orientation:      "orthogonal",
		RenderOrder:      "right-down",
		TiledVersion:     "1.12.2",
		Type:             "map",
		Version:          "1.10",
		TileWidth:        tileSize,
		TileHeight:       tileSize,
		NextLayerID:      3,
		NextObjectID:     nextObjectID,
		Layers: []tmjLayer{
			{
				Type:    "tilelayer",
				Name:    "Tile Layer 1",
				Width:   grid.ColCount,
				Height:  grid.RowCount,
				Data:    data,
				Visible: true,
				Opacity: 1,
			},
			{
				Type:      "objectgroup",
				Name:      "Object Layer 1",
				DrawOrder: "topdown",
				Objects:   objects,
				Visible:   true,
				Opacity:   1,
			},
		},
	}

	if len(embeddedTiles) > 0 {
		doc.Tilesets = []tmjEmbeddedTileset{{
			FirstGID:   1,
			Name:       "export",
			TileWidth:  tileSize,
			TileHeight: tileSize,
			TileCount:  len(embeddedTiles),
			Columns:    len(embeddedTiles),
			Tiles:      embeddedTiles,
		}}
	}

	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}

	out, err := json.MarshalIndent(doc, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, out, 0o644)
}
