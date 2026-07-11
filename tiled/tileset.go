package tiled

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type tilesetRef struct {
	FirstGID uint32 `json:"firstgid"`
	Source   string `json:"source"`
	Name     string `json:"name"`
	TileWidth int   `json:"tilewidth"`
	TileHeight int  `json:"tileheight"`
	Columns  int    `json:"columns"`
	Image    string `json:"image"`
	Tiles    []embeddedTile `json:"tiles"`
}

type embeddedTile struct {
	ID         int               `json:"id"`
	Type       string            `json:"type"`
	Image      string            `json:"image"`
	Properties []propertyEntry   `json:"properties"`
}

type propertyEntry struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type tsxTileset struct {
	XMLName    xml.Name `xml:"tileset"`
	Name       string   `xml:"name,attr"`
	TileWidth  int      `xml:"tilewidth,attr"`
	TileHeight int      `xml:"tileheight,attr"`
	TileCount  int      `xml:"tilecount,attr"`
	Columns    int      `xml:"columns,attr"`
	Image      tsxImage `xml:"image"`
	Tiles      []tsxTile `xml:"tile"`
}

type tsxImage struct {
	Source string `xml:"source,attr"`
}

type tsxTile struct {
	ID    int      `xml:"id,attr"`
	Type  string   `xml:"type,attr"`
	Image tsxImage `xml:"image"`
}

type tileInfo struct {
	Texture    string
	Type       string
	AtlasKey   string
	FrameIndex int32
}

type TilesetCatalog struct {
	refs    []tilesetRef
	base    map[uint32]tileInfo
	atlases map[string]AtlasTileset
}

func LoadTilesetsFromMap(mapPath string, refs []tilesetRef) (*TilesetCatalog, error) {
	catalog := &TilesetCatalog{
		refs: refs,
		base: make(map[uint32]tileInfo),
	}

	mapDir := filepath.Dir(mapPath)
	for _, ref := range refs {
		if ref.Source != "" {
			if err := catalog.loadExternalTileset(mapDir, ref); err != nil {
				return nil, err
			}
			continue
		}

		catalog.loadEmbeddedTileset(ref)
	}

	return catalog, nil
}

func (c *TilesetCatalog) loadExternalTileset(mapDir string, ref tilesetRef) error {
	sourcePath := filepath.Clean(filepath.Join(mapDir, filepath.FromSlash(ref.Source)))
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read tileset %q: %w", sourcePath, err)
	}

	var ts tsxTileset
	if err := xml.Unmarshal(data, &ts); err != nil {
		return fmt.Errorf("parse tileset %q: %w", sourcePath, err)
	}

	tilesetDir := filepath.Dir(sourcePath)
	atlasImage := strings.TrimSpace(ts.Image.Source)
	atlasBase := ""
	atlasKey := ""
	if atlasImage != "" {
		if !filepath.IsAbs(atlasImage) {
			atlasImage = filepath.Join(tilesetDir, atlasImage)
		}
		atlasBase = filepath.Base(atlasImage)
		atlasKey = strings.TrimSpace(ts.Name)
		if atlasKey == "" {
			atlasKey = strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
		}
		webPath := ResWebPath(atlasImage)
		if webPath != "" && ts.Columns > 0 {
			c.registerAtlas(AtlasTileset{
				Key:        atlasKey,
				ImagePath:  webPath,
				TileWidth:  ts.TileWidth,
				TileHeight: ts.TileHeight,
				Columns:    ts.Columns,
			})
		}
	}

	for _, tile := range ts.Tiles {
		gid := ref.FirstGID + uint32(tile.ID)
		info := tileInfo{Type: strings.TrimSpace(tile.Type)}

		if tile.Image.Source != "" {
			info.Texture = textureFromImagePath(tile.Image.Source)
		} else if atlasBase != "" {
			info.Texture = fmt.Sprintf("%s#%d", atlasBase, tile.ID)
			info.AtlasKey = atlasKey
			info.FrameIndex = int32(tile.ID)
		}

		c.base[gid] = info
	}

	if atlasBase != "" && ts.Columns > 0 {
		// Fill atlas slots that do not have explicit tile entries.
		for localID := 0; localID < countAtlasTiles(ts); localID++ {
			gid := ref.FirstGID + uint32(localID)
			if _, ok := c.base[gid]; ok {
				continue
			}
			c.base[gid] = tileInfo{
				Texture:    fmt.Sprintf("%s#%d", atlasBase, localID),
				AtlasKey:   atlasKey,
				FrameIndex: int32(localID),
			}
		}
	}

	return nil
}

func (c *TilesetCatalog) loadEmbeddedTileset(ref tilesetRef) {
	atlasBase := filepath.Base(ref.Image)
	atlasKey := strings.TrimSpace(ref.Name)
	if atlasKey == "" {
		atlasKey = strings.TrimSuffix(atlasBase, filepath.Ext(atlasBase))
	}
	webPath := ResWebPath(ref.Image)
	if webPath != "" && ref.Columns > 0 {
		tileWidth := ref.TileWidth
		if tileWidth <= 0 {
			tileWidth = 16
		}
		tileHeight := ref.TileHeight
		if tileHeight <= 0 {
			tileHeight = 16
		}
		c.registerAtlas(AtlasTileset{
			Key:        atlasKey,
			ImagePath:  webPath,
			TileWidth:  tileWidth,
			TileHeight: tileHeight,
			Columns:    ref.Columns,
		})
	}

	for _, tile := range ref.Tiles {
		gid := ref.FirstGID + uint32(tile.ID)
		info := tileInfo{Type: strings.TrimSpace(tile.Type)}

		if tile.Image != "" {
			info.Texture = textureFromImagePath(tile.Image)
		} else if atlasBase != "" {
			info.Texture = fmt.Sprintf("%s#%d", atlasBase, tile.ID)
			info.AtlasKey = atlasKey
			info.FrameIndex = int32(tile.ID)
		}

		c.base[gid] = info
	}

	if atlasBase != "" && ref.Columns > 0 {
		tileCount := len(ref.Tiles)
		if tileCount == 0 {
			tileCount = ref.Columns
		}
		for localID := 0; localID < tileCount; localID++ {
			gid := ref.FirstGID + uint32(localID)
			if _, ok := c.base[gid]; ok {
				continue
			}
			c.base[gid] = tileInfo{
				Texture:    fmt.Sprintf("%s#%d", atlasBase, localID),
				AtlasKey:   atlasKey,
				FrameIndex: int32(localID),
			}
		}
	}
}

func countAtlasTiles(ts tsxTileset) int {
	if ts.Columns <= 0 || ts.Image.Source == "" {
		return 0
	}

	if ts.TileCount > 0 {
		return ts.TileCount
	}

	maxID := 0
	for _, tile := range ts.Tiles {
		if tile.ID > maxID {
			maxID = tile.ID
		}
	}

	if maxID > 0 {
		return maxID + 1
	}

	return ts.Columns
}

func (c *TilesetCatalog) Lookup(gid uint32) (tileInfo, bool) {
	info, ok := c.base[gid]
	return info, ok
}

func (c *TilesetCatalog) EntityName(gid uint32) string {
	info, ok := c.Lookup(gid)
	if !ok {
		return ""
	}

	if info.Type != "" {
		// Includes "spawn"; callers decide whether to skip marker types.
		return info.Type
	}

	// Atlas slots (texture like "gameplay.png#2") without an explicit type
	// cannot be mapped to an entdef name reliably.
	if info.Texture == "" || strings.Contains(info.Texture, "#") {
		return ""
	}

	stem := strings.TrimSuffix(filepath.Base(info.Texture), filepath.Ext(info.Texture))

	switch stem {
	case "playerBob", "playerBobLeft", "playerBobRight", "playerBobUp":
		return "player"
	case "pressureButtonPressed":
		return "pressureButton"
	case "stonePotBottom":
		return "pot"
	default:
		return stem
	}
}

func textureFromImagePath(source string) string {
	base := filepath.Base(source)
	if base == "" || base == "." {
		return ""
	}
	return base
}

func ParseTilesetRefs(raw json.RawMessage) ([]tilesetRef, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var refs []tilesetRef
	if err := json.Unmarshal(raw, &refs); err != nil {
		return nil, err
	}

	return refs, nil
}
