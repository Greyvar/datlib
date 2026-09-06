package lint

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	grids "github.com/greyvar/datlib/gridfiles"
	"github.com/greyvar/datlib/tiled"
)

type tmjDoc struct {
	Tilesets []tmjTilesetRef `json:"tilesets"`
	Layers   []tmjLayer      `json:"layers"`
}

type tmjTilesetRef struct {
	Source string `json:"source"`
}

type tmjLayer struct {
	Type    string         `json:"type"`
	Objects []tmjMapObject `json:"objects"`
}

type tmjMapObject struct {
	Template string `json:"template"`
	Type     string `json:"type"`
	Class    string `json:"class"`
	Name     string `json:"name"`
}

// LintTMJMaps loads referenced tilesets then every worlds/**/*.tmj map.
func LintTMJMaps(datDir string, entdefNames map[string]struct{}) []Check {
	worldsRoot := filepath.Join(datDir, "worlds")
	tmjPaths, err := FindFilesByExt(worldsRoot, ".tmj")
	if err != nil {
		check := Check{Type: "map", Filename: worldsRoot}
		check.AddError(fmt.Sprintf("scan tmj maps: %v", err))
		return []Check{check}
	}

	if len(tmjPaths) == 0 {
		return nil
	}

	var out []Check
	tilesets, collectChecks := collectTilesetPaths(tmjPaths)
	out = append(out, collectChecks...)

	for _, path := range tilesets {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		out = append(out, CheckTileset(path))
	}

	for _, path := range tmjPaths {
		out = append(out, CheckTMJFile(path, entdefNames))
	}
	return out
}

// LintTMXMaps warns on legacy worlds/**/*.tmx maps.
func LintTMXMaps(datDir string) []Check {
	worldsRoot := filepath.Join(datDir, "worlds")
	tmxPaths, err := FindFilesByExt(worldsRoot, ".tmx")
	if err != nil {
		check := Check{Type: "map", Filename: worldsRoot}
		check.AddError(fmt.Sprintf("scan tmx maps: %v", err))
		return []Check{check}
	}

	out := make([]Check, 0, len(tmxPaths))
	for _, path := range tmxPaths {
		check := Check{Type: "map", Filename: path}
		check.AddWarning("legacy .tmx map; server loads .tmj/.grid only")
		out = append(out, check)
	}
	return out
}

func collectTilesetPaths(tmjPaths []string) ([]string, []Check) {
	seen := map[string]struct{}{}
	var out []string
	var checks []Check

	for _, tmjPath := range tmjPaths {
		data, err := os.ReadFile(tmjPath)
		if err != nil {
			check := Check{Type: "map", Filename: tmjPath}
			check.AddError(fmt.Sprintf("read map while collecting tilesets: %v", err))
			checks = append(checks, check)
			continue
		}

		var doc tmjDoc
		if err := json.Unmarshal(data, &doc); err != nil {
			check := Check{Type: "map", Filename: tmjPath}
			check.AddError(fmt.Sprintf("parse map while collecting tilesets: %v", err))
			checks = append(checks, check)
			continue
		}

		mapDir := filepath.Dir(tmjPath)
		for _, ref := range doc.Tilesets {
			source := strings.TrimSpace(ref.Source)
			if source == "" {
				continue
			}
			resolved := filepath.Clean(filepath.Join(mapDir, filepath.FromSlash(source)))
			if _, ok := seen[resolved]; ok {
				continue
			}
			seen[resolved] = struct{}{}
			out = append(out, resolved)
		}
	}

	sort.Strings(out)
	return out, checks
}

// CheckTileset loads one .tsx tileset file.
func CheckTileset(path string) Check {
	check := Check{Type: "tileset", Filename: path}

	var warnings []string
	var err error
	withQuietLogs(func() {
		_, warnings, err = tiled.LoadTilesetFile(path)
	})
	for _, warning := range warnings {
		check.AddWarning(warning)
	}
	if err != nil {
		check.AddError(err.Error())
	}

	return check
}

// CheckTMJFile loads one .tmj map and validates refs / entdefs.
func CheckTMJFile(path string, entdefNames map[string]struct{}) Check {
	check := Check{Type: "map", Filename: path}

	data, err := os.ReadFile(path)
	if err != nil {
		check.AddError(err.Error())
		return check
	}

	var doc tmjDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		check.AddError(fmt.Sprintf("parse tmj: %v", err))
		return check
	}

	mapDir := filepath.Dir(path)
	for _, ref := range doc.Tilesets {
		source := strings.TrimSpace(ref.Source)
		if source == "" {
			continue
		}
		resolved := filepath.Clean(filepath.Join(mapDir, filepath.FromSlash(source)))
		if _, err := os.Stat(resolved); err != nil {
			check.AddError(fmt.Sprintf("missing tileset %q", source))
		}
	}

	for _, layer := range doc.Layers {
		if layer.Type != "objectgroup" {
			continue
		}
		for _, obj := range layer.Objects {
			template := strings.TrimSpace(obj.Template)
			if template != "" {
				if issue := templateRefIssue(mapDir, template); issue != "" {
					check.AddError(issue)
				}
			}
			addObjectEntdefIssue(&check, mapDir, obj, entdefNames)
		}
	}

	withQuietLogs(func() {
		_, err = grids.ReadGrid(path)
	})
	if err != nil {
		check.AddError(err.Error())
	}

	return check
}

func addObjectEntdefIssue(check *Check, mapDir string, obj tmjMapObject, entdefNames map[string]struct{}) {
	name := firstNonEmpty(obj.Type, obj.Class, obj.Name)

	if template := strings.TrimSpace(obj.Template); template != "" {
		if tplPath := resolveTemplatePath(mapDir, template); tplPath != "" {
			var tpl *tiled.Template
			withQuietLogs(func() {
				tpl, _ = tiled.LoadTemplate(tplPath)
			})
			if tpl != nil && strings.TrimSpace(tpl.EntityName) != "" && name == "" {
				name = strings.TrimSpace(tpl.EntityName)
			}
		}
	}

	if name == "" || strings.EqualFold(name, "spawn") {
		return
	}

	if _, ok := entdefNames[name]; !ok {
		check.AddError(fmt.Sprintf("unknown entdef %q", name))
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func resolveTemplatePath(mapDir, template string) string {
	path := filepath.Clean(filepath.Join(mapDir, filepath.FromSlash(template)))
	candidates := []string{path}
	if filepath.Ext(path) == "" {
		candidates = append(candidates, path+".tx")
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func templateRefIssue(mapDir, template string) string {
	if resolveTemplatePath(mapDir, template) != "" {
		return ""
	}

	path := filepath.Clean(filepath.Join(mapDir, filepath.FromSlash(template)))
	candidates := []string{path}
	if filepath.Ext(path) == "" {
		candidates = append(candidates, path+".tx")
	}
	return fmt.Sprintf("missing object template %q (tried %s)", template, strings.Join(candidates, ", "))
}
