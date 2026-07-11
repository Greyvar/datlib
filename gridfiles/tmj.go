package gridfiles

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/greyvar/datlib/tiled"
	log "github.com/sirupsen/logrus"
)

type tmjMap struct {
	Width      uint32          `json:"width"`
	Height     uint32          `json:"height"`
	TileWidth  int             `json:"tilewidth"`
	TileHeight int             `json:"tileheight"`
	Layers     []tmjLayer      `json:"layers"`
	Tilesets   json.RawMessage `json:"tilesets"`
	NextObjectID int           `json:"nextobjectid"`
}

type tmjLayer struct {
	Type      string         `json:"type"`
	Name      string         `json:"name"`
	Width     uint32         `json:"width"`
	Height    uint32         `json:"height"`
	Data      []uint32       `json:"data"`
	Visible   bool           `json:"visible"`
	Opacity   float64        `json:"opacity,omitempty"`
	DrawOrder string         `json:"draworder,omitempty"`
	Objects   []tmjMapObject `json:"objects"`
}

type tmjMapObject struct {
	ID       int     `json:"id"`
	GID      uint32  `json:"gid"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Rotation float64 `json:"rotation"`
	Visible  bool    `json:"visible"`
	Type     string  `json:"type"`
	Name     string  `json:"name"`
	Template string  `json:"template"`
}

func ReadGridTMJ(filename string) (*Grid, error) {
	log.Infof("Loading TMJ grid: %v", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var doc tmjMap
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse tmj %q: %w", filename, err)
	}

	if doc.TileWidth <= 0 {
		doc.TileWidth = 16
	}
	if doc.TileHeight <= 0 {
		doc.TileHeight = 16
	}

	refs, err := tiled.ParseTilesetRefs(doc.Tilesets)
	if err != nil {
		return nil, err
	}

	catalog, err := tiled.LoadTilesetsFromMap(filename, refs)
	if err != nil {
		return nil, err
	}

	g := &Grid{
		Filename: filepath.Base(filename),
		ColCount: doc.Width,
		RowCount: doc.Height,
	}
	g.Build()

	for _, atlas := range catalog.AtlasTilesets() {
		g.Tilesets = append(g.Tilesets, GridTileset{
			Key:        atlas.Key,
			ImagePath:  atlas.ImagePath,
			TileWidth:  atlas.TileWidth,
			TileHeight: atlas.TileHeight,
			Columns:    atlas.Columns,
		})
	}

	for _, layer := range doc.Layers {
		if !layerVisible(layer) {
			continue
		}

		switch layer.Type {
		case "tilelayer":
			if err := applyTileLayer(g, catalog, layer); err != nil {
				return nil, err
			}
		case "objectgroup":
			applyObjectLayer(g, catalog, doc, layer, filepath.Dir(filename))
		}
	}

	if doc.NextObjectID > 1 {
		g.LastEntityId = fmt.Sprintf("%d", doc.NextObjectID-1)
	}

	return g, nil
}

func layerVisible(layer tmjLayer) bool {
	return layer.Visible || layer.Type == "tilelayer" || layer.Type == "objectgroup"
}

func applyTileLayer(g *Grid, catalog *tiled.TilesetCatalog, layer tmjLayer) error {
	width := layer.Width
	if width == 0 {
		width = g.ColCount
	}

	for index, rawGID := range layer.Data {
		if rawGID == 0 {
			continue
		}

		decoded := tiled.DecodeGID(rawGID)
		if decoded.ID == 0 {
			continue
		}

		row := uint32(index) / width
		col := uint32(index) % width
		if row >= g.RowCount || col >= g.ColCount {
			continue
		}

		info, ok := catalog.Lookup(decoded.ID)
		texture := "construct.png"
		tile := &Tile{
			Row:         row,
			Col:         col,
			Rot:         decoded.Rot,
			FlipH:       decoded.FlipH,
			FlipV:       decoded.FlipV,
			Traversable: true,
		}
		if ok && info.Texture != "" {
			texture = info.Texture
			tile.AtlasKey = info.AtlasKey
			tile.FrameIndex = info.FrameIndex
		}
		tile.Texture = texture

		g.Tiles[row][col] = tile
	}

	return nil
}

func applyObjectLayer(g *Grid, catalog *tiled.TilesetCatalog, doc tmjMap, layer tmjLayer, mapDir string) {
	templates := map[string]*tiled.Template{}

	for _, obj := range layer.Objects {
		entityName := strings.TrimSpace(obj.Type)
		if entityName == "" {
			entityName = strings.TrimSpace(obj.Name)
		}

		if obj.GID != 0 {
			decoded := tiled.DecodeGID(obj.GID)
			if entityName == "" {
				entityName = catalog.EntityName(decoded.ID)
			}
		} else if obj.Template != "" {
			tpl, err := resolveTemplate(templates, mapDir, obj.Template)
			if err != nil {
				log.Warnf("Cannot load object template %q: %v", obj.Template, err)
				continue
			}
			if entityName == "" {
				entityName = tpl.EntityName
			}
		} else {
			continue
		}

		if entityName == "" || entityName == "spawn" {
			continue
		}

		col := uint32(math.Floor(obj.X / float64(doc.TileWidth)))
		row := uint32(math.Floor(obj.Y / float64(doc.TileHeight)))

		ent := GridFileEntityInstance{
			Row:        row,
			Col:        col,
			Definition: entityName,
		}
		g.Entities = append(g.Entities, ent)

		if obj.ID > 0 {
			g.LastEntityId = fmt.Sprintf("%d", obj.ID)
		}
	}
}

func resolveTemplate(cache map[string]*tiled.Template, mapDir string, ref string) (*tiled.Template, error) {
	if tpl, ok := cache[ref]; ok {
		return tpl, nil
	}

	path := filepath.Clean(filepath.Join(mapDir, filepath.FromSlash(ref)))

	candidates := []string{path}
	if filepath.Ext(path) == "" {
		candidates = append(candidates, path+".tx")
	}

	var lastErr error
	for _, candidate := range candidates {
		tpl, err := tiled.LoadTemplate(candidate)
		if err == nil {
			cache[ref] = tpl
			return tpl, nil
		}
		lastErr = err
	}

	return nil, lastErr
}
