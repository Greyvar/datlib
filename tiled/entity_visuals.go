package tiled

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultEntityVisualType is the tileset type/class used when no matching
// entdef visual exists (placeholder art in general-entities.tsx).
const DefaultEntityVisualType = "construct_entity"

// AnimFrame is one step of a Tiled tile animation.
type AnimFrame struct {
	TileID     int
	DurationMs int
}

// TypedTileVisual is the render data for a tileset tile identified by type/class.
type TypedTileVisual struct {
	Type          string
	Texture       string // basename under img/textures/entities/ (or path under /res/)
	AtlasKey      string
	FrameIndex    int32
	Frames        []int32 // spritesheet frame indices (tile ids in atlas)
	Holds         []int32 // ms; parallel to Frames when from Tiled animation
	Collision     []CollisionRect
	CollisionType CollisionType
	// TileProps are custom properties from the tileset tile (msg default,
	// collision_transition, etc.). Instance map properties override these.
	TileProps  map[string]string
	TileWidth  int
	TileHeight int
	Columns    int
}

type tsxAnimFrame struct {
	TileID   int `xml:"tileid,attr"`
	Duration int `xml:"duration,attr"`
}

type tsxAnimation struct {
	Frames []tsxAnimFrame `xml:"frame"`
}

// extended tile fields for animation + class (Tiled 1.9+)
type tsxTileVisual struct {
	ID          int            `xml:"id,attr"`
	Type        string         `xml:"type,attr"`
	Class       string         `xml:"class,attr"`
	Image       tsxImage       `xml:"image"`
	Animation   *tsxAnimation  `xml:"animation"`
	ObjectGroup tsxObjectGroup `xml:"objectgroup"`
	Properties  tsxProperties  `xml:"properties"`
}

type tsxTilesetVisual struct {
	Name       string          `xml:"name,attr"`
	TileWidth  int             `xml:"tilewidth,attr"`
	TileHeight int             `xml:"tileheight,attr"`
	TileCount  int             `xml:"tilecount,attr"`
	Columns    int             `xml:"columns,attr"`
	Image      tsxImage        `xml:"image"`
	Tiles      []tsxTileVisual `xml:"tile"`
}

// EntityVisualCatalog indexes tileset tiles by type/class for entity rendering.
type EntityVisualCatalog struct {
	byType map[string]TypedTileVisual
}

// LoadEntityVisualCatalog loads every *.tsx under dir (typically dat/entdefs).
func LoadEntityVisualCatalog(dir string) (*EntityVisualCatalog, error) {
	cat := &EntityVisualCatalog{byType: make(map[string]TypedTileVisual)}
	matches, err := filepath.Glob(filepath.Join(dir, "*.tsx"))
	if err != nil {
		return nil, err
	}
	for _, path := range matches {
		if err := cat.loadTSX(path); err != nil {
			return nil, err
		}
	}
	return cat, nil
}

func tsxPropsToStringMap(props []tsxProperty) map[string]string {
	if len(props) == 0 {
		return nil
	}
	out := make(map[string]string, len(props))
	for _, p := range props {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			continue
		}
		out[name] = strings.TrimSpace(p.Value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (c *EntityVisualCatalog) loadTSX(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read tileset %q: %w", path, err)
	}

	var ts tsxTilesetVisual
	if err := xml.Unmarshal(data, &ts); err != nil {
		return fmt.Errorf("parse tileset %q: %w", path, err)
	}

	tilesetDir := filepath.Dir(path)
	atlasImage := strings.TrimSpace(ts.Image.Source)
	textureBase := ""
	atlasKey := strings.TrimSpace(ts.Name)
	if atlasKey == "" {
		atlasKey = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	if atlasImage != "" {
		abs := atlasImage
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(tilesetDir, abs)
		}
		textureBase = filepath.Base(abs)
		// Client loads entity textures from img/textures/entities/<file>
		if web := ResWebPath(abs); web != "" {
			if strings.HasPrefix(web, "img/textures/entities/") {
				textureBase = strings.TrimPrefix(web, "img/textures/entities/")
			} else {
				textureBase = filepath.Base(web)
			}
		}
	}

	for _, tile := range ts.Tiles {
		typeName := strings.TrimSpace(tile.Type)
		if typeName == "" {
			typeName = strings.TrimSpace(tile.Class)
		}
		if typeName == "" {
			continue
		}

		vis := TypedTileVisual{
			Type:          typeName,
			Texture:       textureBase,
			AtlasKey:      atlasKey,
			FrameIndex:    int32(tile.ID),
			Collision:     collisionRectsFromObjects(tile.ObjectGroup.Objects),
			CollisionType: ParseCollisionType(propertyStringValue(tile.Properties.Properties, "collision_type")),
			TileProps:     tsxPropsToStringMap(tile.Properties.Properties),
			TileWidth:     ts.TileWidth,
			TileHeight:    ts.TileHeight,
			Columns:       ts.Columns,
		}

		if tile.Image.Source != "" {
			imgPath := tile.Image.Source
			if !filepath.IsAbs(imgPath) {
				imgPath = filepath.Join(tilesetDir, imgPath)
			}
			if _, err := os.Stat(imgPath); err != nil {
				// Skip broken standalone images so they cannot clobber a valid
				// atlas type of the same name (e.g. missing pot.png vs general-entities).
				continue
			}
			vis.Texture = textureFromImagePath(tile.Image.Source)
			vis.FrameIndex = 0
			vis.Frames = []int32{0}
		} else if tile.Animation != nil && len(tile.Animation.Frames) > 0 {
			vis.Frames = make([]int32, 0, len(tile.Animation.Frames))
			vis.Holds = make([]int32, 0, len(tile.Animation.Frames))
			for _, f := range tile.Animation.Frames {
				vis.Frames = append(vis.Frames, int32(f.TileID))
				vis.Holds = append(vis.Holds, int32(f.Duration))
			}
		} else {
			vis.Frames = []int32{int32(tile.ID)}
		}

		c.byType[typeName] = vis
	}

	return nil
}

// Lookup returns the visual for an exact tileset type/class name.
func (c *EntityVisualCatalog) Lookup(typeName string) (TypedTileVisual, bool) {
	if c == nil || c.byType == nil {
		return TypedTileVisual{}, false
	}
	vis, ok := c.byType[typeName]
	return vis, ok
}

// ResolveStateExact picks the tileset type for an entdef title + state with no
// default fallback. Order: title_state, title (when state is initial or empty),
// then state alone.
func (c *EntityVisualCatalog) ResolveStateExact(title, state, initialState string) (TypedTileVisual, bool) {
	title = strings.TrimSpace(title)
	state = strings.TrimSpace(state)
	initialState = strings.TrimSpace(initialState)

	candidates := make([]string, 0, 3)
	if title != "" && state != "" {
		candidates = append(candidates, title+"_"+state)
	}
	if state == "" || state == initialState || state == title {
		if title != "" {
			candidates = append(candidates, title)
		}
	}
	if state != "" && state != title {
		candidates = append(candidates, state)
	}

	seen := map[string]bool{}
	for _, name := range candidates {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		if vis, ok := c.Lookup(name); ok {
			return vis, true
		}
	}
	return TypedTileVisual{}, false
}

// ResolveState is ResolveStateExact, then DefaultEntityVisualType (construct_entity).
func (c *EntityVisualCatalog) ResolveState(title, state, initialState string) (TypedTileVisual, bool) {
	if vis, ok := c.ResolveStateExact(title, state, initialState); ok {
		return vis, true
	}
	return c.Lookup(DefaultEntityVisualType)
}

// HasType reports whether any loaded tileset defines this type/class.
func (c *EntityVisualCatalog) HasType(typeName string) bool {
	_, ok := c.Lookup(typeName)
	return ok
}

// TypeCount returns how many typed tiles are indexed.
func (c *EntityVisualCatalog) TypeCount() int {
	if c == nil || c.byType == nil {
		return 0
	}
	return len(c.byType)
}

// SetTypeForTest registers a typed tile visual (tests only).
func (c *EntityVisualCatalog) SetTypeForTest(typeName string, vis TypedTileVisual) {
	if c == nil {
		return
	}
	if c.byType == nil {
		c.byType = make(map[string]TypedTileVisual)
	}
	vis.Type = typeName
	c.byType[typeName] = vis
}
