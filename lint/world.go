package lint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type tiledWorldDoc struct {
	Type       string               `json:"type"`
	Maps       []tiledWorldMapRef   `json:"maps"`
	Patterns   []json.RawMessage    `json:"patterns"`
	Properties []tiledWorldProperty `json:"properties"`
}

type tiledWorldMapRef struct {
	FileName string `json:"fileName"`
}

type tiledWorldProperty struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type yamlWorldDoc struct {
	SpawnGrid string `yaml:"spawnGrid"`
}

// LintWorlds validates every world directory under dat/worlds.
func LintWorlds(datDir string) []Check {
	worldsRoot := filepath.Join(datDir, "worlds")
	entries, err := os.ReadDir(worldsRoot)
	if err != nil {
		check := Check{Type: "world", Filename: worldsRoot}
		check.AddError(fmt.Sprintf("scan worlds: %v", err))
		return []Check{check}
	}

	var out []Check
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		out = append(out, lintWorldDir(filepath.Join(worldsRoot, entry.Name()))...)
	}
	return out
}

func lintWorldDir(dir string) []Check {
	worldFiles, err := filepath.Glob(filepath.Join(dir, "*.world"))
	if err != nil {
		check := Check{Type: "world", Filename: dir}
		check.AddError(err.Error())
		return []Check{check}
	}

	var out []Check
	yamlPath := filepath.Join(dir, "world.yml")
	if _, err := os.Stat(yamlPath); err == nil {
		out = append(out, CheckYAMLWorld(yamlPath, dir))
	}

	for _, path := range worldFiles {
		out = append(out, CheckTiledWorld(path, dir))
	}
	return out
}

// CheckTiledWorld validates a Tiled .world file.
func CheckTiledWorld(path, dir string) Check {
	check := Check{Type: "world", Filename: path}

	data, err := os.ReadFile(path)
	if err != nil {
		check.AddError(err.Error())
		return check
	}

	var doc tiledWorldDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		check.AddError(fmt.Sprintf("parse world: %v", err))
		return check
	}

	listed := map[string]struct{}{}
	for _, m := range doc.Maps {
		name := strings.TrimSpace(m.FileName)
		if name == "" {
			continue
		}
		listed[name] = struct{}{}
		mapPath := filepath.Join(dir, filepath.FromSlash(name))
		if _, err := os.Stat(mapPath); err != nil {
			check.AddError(fmt.Sprintf("missing map %q", name))
		}
	}

	if len(doc.Maps) == 0 && len(doc.Patterns) > 0 {
		check.AddWarning("pattern-based world; map membership not validated against maps[]")
	}

	ondisk := mapFilesInDir(dir)
	for name := range ondisk {
		if _, ok := listed[name]; !ok && len(listed) > 0 {
			check.AddWarning(fmt.Sprintf("map %q on disk but not listed in world file", name))
		}
	}

	spawn := worldSpawnGrid(doc.Properties)
	if spawn != "" {
		spawnPath := filepath.Join(dir, filepath.FromSlash(spawn))
		if _, err := os.Stat(spawnPath); err != nil {
			check.AddError(fmt.Sprintf("spawnGrid %q does not exist", spawn))
		}
	}

	return check
}

// CheckYAMLWorld validates a legacy world.yml file.
func CheckYAMLWorld(path, dir string) Check {
	check := Check{Type: "world", Filename: path}

	data, err := os.ReadFile(path)
	if err != nil {
		check.AddError(err.Error())
		return check
	}

	var doc yamlWorldDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		check.AddError(fmt.Sprintf("parse world.yml: %v", err))
		return check
	}

	if doc.SpawnGrid != "" {
		candidates := []string{
			filepath.Join(dir, "grids", doc.SpawnGrid),
			filepath.Join(dir, "grids", doc.SpawnGrid+".grid"),
			filepath.Join(dir, doc.SpawnGrid),
		}
		found := false
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				found = true
				break
			}
		}
		if !found {
			check.AddError(fmt.Sprintf("spawnGrid %q does not exist", doc.SpawnGrid))
		}
	}

	return check
}

func worldSpawnGrid(props []tiledWorldProperty) string {
	for _, prop := range props {
		if prop.Name == "spawnGrid" {
			if s, ok := prop.Value.(string); ok {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

func mapFilesInDir(dir string) map[string]struct{} {
	out := map[string]struct{}{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".tmj" || ext == ".tmx" {
			out[entry.Name()] = struct{}{}
		}
	}
	return out
}
