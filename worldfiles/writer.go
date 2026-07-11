package worldfiles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/greyvar/datlib/gridfiles"
)

type TiledExportOptions struct {
	DestDir       string
	WorldFileName string
	TileSize      int
	TexturePrefix string
	SkipMaps      bool
}

type tiledWorldExport struct {
	Type                   string          `json:"type"`
	Maps                   []MapPlacement  `json:"maps"`
	OnlyShowAdjacentMaps   bool            `json:"onlyShowAdjacentMaps"`
	Properties             []tiledProperty `json:"properties"`
}

func ExportYAMLWorldAsTiled(name string, opts TiledExportOptions) (*World, error) {
	world, err := LoadWorld(name)
	if err != nil {
		return nil, err
	}

	if world.Definition == nil || world.Definition.Format != "yaml" {
		return nil, fmt.Errorf("world %q is not a YAML world", name)
	}

	destDir := opts.DestDir
	if destDir == "" {
		destDir = filepath.Join(WorldDir(name) + "_tiled")
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, err
	}

	tileSize := opts.TileSize
	if tileSize <= 0 {
		tileSize = defaultTileSize
	}

	tmjOpts := gridfiles.TMJWriteOptions{
		TileSize:      tileSize,
		TexturePrefix: opts.TexturePrefix,
	}

	placements := MapPlacementsFromGrids(world.Grids, tileSize)
	if !opts.SkipMaps {
		for gridName, grid := range world.Grids {
			tmjName := GridToTMJFilename(gridName)
			tmjPath := filepath.Join(destDir, tmjName)
			if err := gridfiles.WriteGridTMJ(grid, tmjPath, tmjOpts); err != nil {
				return nil, fmt.Errorf("write map %q: %w", tmjName, err)
			}
		}
	}

	worldFileName := opts.WorldFileName
	if worldFileName == "" {
		worldFileName = name + ".world"
	}

	spawnTMJ := SpawnGridToTMJ(world.SpawnGrid)
	properties := []tiledProperty{
		{Name: "author", Type: "string", Value: world.Author},
		{Name: "title", Type: "string", Value: world.Title},
		{Name: "spawnGrid", Type: "string", Value: spawnTMJ},
	}

	if len(world.Definition.Triggers) > 0 {
		triggerJSON, err := marshalYAMLValue(world.Definition.Triggers)
		if err != nil {
			return nil, fmt.Errorf("encode triggers: %w", err)
		}
		properties = append(properties, tiledProperty{
			Name:  "triggers",
			Type:  "string",
			Value: string(triggerJSON),
		})
	}

	export := tiledWorldExport{
		Type:                 "world",
		Maps:                 placements,
		OnlyShowAdjacentMaps: false,
		Properties:           properties,
	}

	worldPath := filepath.Join(destDir, worldFileName)
	out, err := json.MarshalIndent(export, "", "    ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(worldPath, out, 0o644); err != nil {
		return nil, err
	}

	return LoadWorldFromDir(name, destDir)
}

func LoadWorldFromDir(id, dir string) (*World, error) {
	def, worldPath, err := loadTiledWorld(dir)
	if err != nil {
		return nil, err
	}
	_ = worldPath
	return loadTiledWorldGrids(id, dir, def)
}

func WriteTiledWorld(world *World, worldPath string) error {
	if world == nil || world.Definition == nil {
		return fmt.Errorf("world is nil")
	}

	placements := world.Definition.Maps
	if len(placements) == 0 {
		placements = MapPlacementsFromGrids(world.Grids, defaultTileSize)
	}

	spawnTMJ := SpawnGridToTMJ(world.SpawnGrid)
	if strings.HasSuffix(world.SpawnGrid, ".tmj") {
		spawnTMJ = world.SpawnGrid
	}

	properties := []tiledProperty{
		{Name: "author", Type: "string", Value: world.Author},
		{Name: "title", Type: "string", Value: world.Title},
		{Name: "spawnGrid", Type: "string", Value: spawnTMJ},
	}
	for key, value := range world.Definition.Properties {
		if key == "author" || key == "title" || key == "spawnGrid" {
			continue
		}
		properties = append(properties, tiledProperty{
			Name:  key,
			Type:  "string",
			Value: value,
		})
	}

	export := tiledWorldExport{
		Type:                 "world",
		Maps:                 placements,
		OnlyShowAdjacentMaps: false,
		Properties:           properties,
	}

	out, err := json.MarshalIndent(export, "", "    ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(worldPath), 0o755); err != nil {
		return err
	}

	return os.WriteFile(worldPath, out, 0o644)
}

func ListYAMLWorlds() ([]Summary, error) {
	all, err := ListWorlds()
	if err != nil {
		return nil, err
	}

	out := make([]Summary, 0)
	for _, summary := range all {
		if summary.Format == "yaml" {
			out = append(out, summary)
		}
	}
	return out, nil
}
