package worldfiles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	datlib "github.com/greyvar/datlib/common"
	"github.com/greyvar/datlib/gridfiles"
	"gopkg.in/yaml.v2"
)

type tiledWorldFile struct {
	Type       string         `json:"type"`
	Maps       []MapPlacement `json:"maps"`
	Properties []tiledProperty `json:"properties"`
}

type tiledProperty struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func WorldDir(name string) string {
	return filepath.Join(datlib.DatDir(), "worlds", name)
}

func LoadWorld(name string) (*World, error) {
	base := WorldDir(name)

	if def, err := loadYAMLWorld(base); err == nil {
		return loadYAMLGrids(name, base, def)
	}

	if def, worldPath, err := loadTiledWorld(base); err == nil {
		return loadTiledWorldGrids(name, filepath.Dir(worldPath), def)
	}

	return nil, fmt.Errorf("world %q: no world.yml or .world file found", name)
}

func loadYAMLWorld(base string) (*Definition, error) {
	defPath := filepath.Join(base, "world.yml")
	data, err := os.ReadFile(defPath)
	if err != nil {
		return nil, err
	}

	def := &Definition{Format: "yaml", Properties: map[string]string{}}
	if err := yaml.Unmarshal(data, def); err != nil {
		return nil, err
	}

	return def, nil
}

func loadYAMLGrids(name, base string, def *Definition) (*World, error) {
	world := &World{
		ID:         name,
		Definition: def,
		Grids:      make(map[string]*gridfiles.Grid),
	}

	gridsDir := filepath.Join(base, "grids")
	entries, err := os.ReadDir(gridsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".grid") {
			continue
		}

		path := filepath.Join(gridsDir, entry.Name())
		grid, err := gridfiles.ReadGrid(path)
		if err != nil {
			return nil, err
		}

		grid.Filename = entry.Name()
		world.Grids[entry.Name()] = grid
	}

	finalizeWorld(world)
	return world, nil
}

func loadTiledWorld(base string) (*Definition, string, error) {
	matches, err := filepath.Glob(filepath.Join(base, "*.world"))
	if err != nil {
		return nil, "", err
	}

	if len(matches) == 0 {
		return nil, "", fmt.Errorf("no .world file in %s", base)
	}

	worldPath := matches[0]
	data, err := os.ReadFile(worldPath)
	if err != nil {
		return nil, "", err
	}

	doc := &tiledWorldFile{}
	if err := json.Unmarshal(data, doc); err != nil {
		return nil, "", err
	}

	def := &Definition{
		Format:     "tiled-world",
		Properties: map[string]string{},
		Maps:       doc.Maps,
	}

	for _, prop := range doc.Properties {
		def.Properties[prop.Name] = propertyToString(prop.Value)
	}

	if def.Title == "" {
		def.Title = def.Properties["title"]
	}
	if def.Author == "" {
		def.Author = def.Properties["author"]
	}
	if def.SpawnGrid == "" {
		def.SpawnGrid = def.Properties["spawnGrid"]
	}

	return def, worldPath, nil
}

func loadTiledWorldGrids(name, base string, def *Definition) (*World, error) {
	world := &World{
		ID:         name,
		Definition: def,
		Grids:      make(map[string]*gridfiles.Grid),
	}

	for _, placement := range def.Maps {
		if placement.FileName == "" {
			continue
		}

		path := filepath.Join(base, placement.FileName)
		grid, err := gridfiles.ReadGrid(path)
		if err != nil {
			return nil, fmt.Errorf("load map %q: %w", placement.FileName, err)
		}

		grid.Filename = placement.FileName
		world.Grids[placement.FileName] = grid
	}

	finalizeWorld(world)
	return world, nil
}

func ListWorlds() ([]Summary, error) {
	root := filepath.Join(datlib.DatDir(), "worlds")
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	out := make([]Summary, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		id := entry.Name()
		summary, ok, err := summarizeWorld(filepath.Join(root, id))
		if err != nil || !ok {
			continue
		}

		summary.ID = id
		out = append(out, summary)
	}

	return out, nil
}

func summarizeWorld(base string) (Summary, bool, error) {
	if def, err := loadYAMLWorld(base); err == nil {
		gridCount := 0
		gridsDir := filepath.Join(base, "grids")
		if gridEntries, err := os.ReadDir(gridsDir); err == nil {
			for _, gridEntry := range gridEntries {
				if !gridEntry.IsDir() && strings.HasSuffix(gridEntry.Name(), ".grid") {
					gridCount++
				}
			}
		}

		return Summary{
			Title:     def.Title,
			SpawnGrid: def.SpawnGrid,
			GridCount: gridCount,
			Format:    "yaml",
		}, true, nil
	}

	def, _, err := loadTiledWorld(base)
	if err != nil {
		return Summary{}, false, nil
	}

	return Summary{
		Title:     def.Title,
		SpawnGrid: def.SpawnGrid,
		GridCount: len(def.Maps),
		Format:    "tiled-world",
	}, true, nil
}

func propertyToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
