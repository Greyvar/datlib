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
	ID         int           `json:"id"`
	GID        uint32        `json:"gid"`
	X          float64       `json:"x"`
	Y          float64       `json:"y"`
	Width      float64       `json:"width"`
	Height     float64       `json:"height"`
	Rotation   float64       `json:"rotation"`
	// Visible is a pointer so omitted Tiled fields (common on template
	// instances) default to visible=true, matching Tiled's behaviour.
	Visible    *bool         `json:"visible"`
	Type       string        `json:"type"`
	Name       string        `json:"name"`
	Template   string        `json:"template"`
	Properties []tmjProperty `json:"properties"`
}

type tmjProperty struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
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
	g.BuildEmpty()

	for _, atlas := range catalog.AtlasTilesets() {
		g.Tilesets = append(g.Tilesets, GridTileset{
			Key:        atlas.Key,
			ImagePath:  atlas.ImagePath,
			TileWidth:  atlas.TileWidth,
			TileHeight: atlas.TileHeight,
			Columns:    atlas.Columns,
		})
	}

	mapDir := filepath.Dir(filename)
	for _, layer := range doc.Layers {
		if !layerVisible(layer) {
			continue
		}

		switch layer.Type {
		case "tilelayer":
			tileLayer := NewTileLayer(layer.Name, LayerKindTerrain)
			if err := applyTileLayer(&tileLayer, catalog, layer, g); err != nil {
				return nil, err
			}
			g.Layers = append(g.Layers, tileLayer)
		case "objectgroup":
			objectLayer := NewTileLayer(layer.Name, LayerKindObject)
			applyObjectLayer(g, catalog, doc, layer, mapDir, &objectLayer)
			g.Layers = append(g.Layers, objectLayer)
		}
	}

	g.SyncTopTiles()

	if doc.NextObjectID > 1 {
		g.LastEntityId = fmt.Sprintf("%d", doc.NextObjectID-1)
	}

	return g, nil
}

func layerVisible(layer tmjLayer) bool {
	return layer.Visible || layer.Type == "tilelayer" || layer.Type == "objectgroup"
}

func applyTileLayer(layer *TileLayer, catalog *tiled.TilesetCatalog, tmjLayer tmjLayer, g *Grid) error {
	width := tmjLayer.Width
	if width == 0 {
		width = g.ColCount
	}

	for index, rawGID := range tmjLayer.Data {
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

		tile := tileFromCatalog(decoded, catalog, row, col)
		layer.SetTile(row, col, tile)
	}

	return nil
}

func tileFromCatalog(decoded tiled.GID, catalog *tiled.TilesetCatalog, row, col uint32) *Tile {
	info, ok := catalog.Lookup(decoded.ID)
	texture := tiled.DefaultTexture
	tile := &Tile{
		Row:         row,
		Col:         col,
		Rot:         decoded.Rot,
		FlipH:       decoded.FlipH,
		FlipV:       decoded.FlipV,
		Traversable: true,
	}
	if ok {
		if info.Texture != "" {
			texture = info.Texture
		}
		tile.AtlasKey = info.AtlasKey
		tile.FrameIndex = info.FrameIndex
		tile.CollisionFromTSX = info.CollisionFromTSX
		tile.Hidden = info.Hidden
		if len(info.Collision) > 0 {
			tile.Collision = append([]CollisionRect(nil), info.Collision...)
		}
	}
	tile.Texture = texture
	return tile
}

func applyObjectLayer(g *Grid, catalog *tiled.TilesetCatalog, doc tmjMap, layer tmjLayer, mapDir string, objectLayer *TileLayer) {
	templates := map[string]*tiled.Template{}

	for _, obj := range layer.Objects {
		if !objectVisible(obj) {
			continue
		}

		entityName := strings.TrimSpace(obj.Type)
		if entityName == "" {
			entityName = strings.TrimSpace(obj.Name)
		}

		var tpl *tiled.Template
		if obj.Template != "" {
			var err error
			tpl, err = resolveTemplate(templates, mapDir, obj.Template)
			if err != nil {
				log.Warnf("Cannot load object template %q: %v", obj.Template, err)
				continue
			}
			if entityName == "" {
				entityName = tpl.EntityName
			}
		}

		if obj.GID != 0 && entityName == "" {
			decoded := tiled.DecodeGID(obj.GID)
			entityName = catalog.EntityName(decoded.ID)
		}

		col := uint32(math.Floor(obj.X / float64(doc.TileWidth)))
		row := uint32(math.Floor(obj.Y / float64(doc.TileHeight)))

		if isSpawnMarker(entityName, obj) {
			if sp, ok := spawnPointFromObject(obj, doc, g); ok {
				g.SpawnPoints = append(g.SpawnPoints, sp)
			}
			continue
		}

		if entityName != "" {
			ent := GridFileEntityInstance{
				Row:        row,
				Col:        col,
				Definition: entityName,
				Properties: tmjPropertiesMap(obj.Properties),
			}
			g.Entities = append(g.Entities, ent)

			if obj.ID > 0 {
				g.LastEntityId = fmt.Sprintf("%d", obj.ID)
			}
			continue
		}

		var decoded tiled.GID
		switch {
		case obj.GID != 0:
			decoded = tiled.DecodeGID(obj.GID)
		case tpl != nil && tpl.GID.ID != 0:
			decoded = tpl.GID
		default:
			continue
		}

		if decoded.ID == 0 {
			continue
		}

		objectLayer.SetTile(row, col, tileFromCatalog(decoded, catalog, row, col))

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

func objectVisible(obj tmjMapObject) bool {
	return obj.Visible == nil || *obj.Visible
}

func isSpawnMarker(entityName string, obj tmjMapObject) bool {
	if strings.EqualFold(strings.TrimSpace(entityName), "spawn") {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(obj.Type), "spawn")
}

func spawnPointFromObject(obj tmjMapObject, doc tmjMap, g *Grid) (SpawnPoint, bool) {
	x := int32(math.Floor(obj.X))
	y := int32(math.Floor(obj.Y))
	maxX := int32(g.ColCount)*int32(doc.TileWidth) - int32(doc.TileWidth)
	maxY := int32(g.RowCount)*int32(doc.TileHeight) - int32(doc.TileHeight)

	if x < 0 || y < 0 || x > maxX || y > maxY {
		log.Warnf("Ignoring out-of-bounds spawn point at (%d,%d) on %s", x, y, g.Filename)
		return SpawnPoint{}, false
	}

	row := uint32(y / int32(doc.TileHeight))
	col := uint32(x / int32(doc.TileWidth))

	return SpawnPoint{
		Row: row,
		Col: col,
		X:   x,
		Y:   y,
	}, true
}

func tmjPropertiesMap(props []tmjProperty) map[string]string {
	if len(props) == 0 {
		return nil
	}
	out := make(map[string]string, len(props))
	for _, p := range props {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			continue
		}
		out[name] = tmjPropertyString(p.Value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func tmjPropertyString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%v", t)
	default:
		return fmt.Sprintf("%v", t)
	}
}
