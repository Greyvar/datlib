package tiled

import (
	"path/filepath"
	"sort"
	"strings"
)

// AtlasTileset describes a Tiled TSX tileset backed by a single spritesheet image.
type AtlasTileset struct {
	Key        string
	ImagePath  string
	TileWidth  int
	TileHeight int
	Columns    int
}

// ResWebPath converts a filesystem path that contains the project's res/ tree
// into a URL path relative to the web client's /res/ base
// (e.g. img/textures/tilesets/grass.png).
func ResWebPath(path string) string {
	normalized := filepath.ToSlash(filepath.Clean(path))
	idx := strings.Index(normalized, "res/")
	if idx < 0 {
		return ""
	}

	suffix := normalized[idx+len("res/"):]
	for _, prefix := range []string{"img/", "atlas/", "entdefs/"} {
		if strings.HasPrefix(suffix, prefix) {
			return suffix
		}
	}

	return ""
}

func (c *TilesetCatalog) registerAtlas(atlas AtlasTileset) {
	if atlas.Key == "" || atlas.ImagePath == "" {
		return
	}
	if c.atlases == nil {
		c.atlases = make(map[string]AtlasTileset)
	}
	if _, exists := c.atlases[atlas.Key]; !exists {
		c.atlases[atlas.Key] = atlas
	}
}

// AtlasTilesets returns atlas tilesets referenced by the loaded map, sorted by key.
func (c *TilesetCatalog) AtlasTilesets() []AtlasTileset {
	if len(c.atlases) == 0 {
		return nil
	}

	result := make([]AtlasTileset, 0, len(c.atlases))
	for _, atlas := range c.atlases {
		result = append(result, atlas)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})

	return result
}
