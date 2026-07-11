package worldfiles

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarshalYAMLTriggers(t *testing.T) {
	triggers := []interface{}{
		map[interface{}]interface{}{
			"title": "firstSwitch",
			"conditions": []interface{}{
				map[interface{}]interface{}{
					"type": "conditionPlayerStepOn",
					"arguments": map[interface{}]interface{}{
						"pos": map[interface{}]interface{}{
							"type": "tileCoords",
							"x":    6,
							"y":    6,
						},
					},
				},
			},
		},
	}

	out, err := marshalYAMLValue(triggers)
	if err != nil {
		t.Fatalf("marshalYAMLValue: %v", err)
	}

	var decoded []map[string]interface{}
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if decoded[0]["title"] != "firstSwitch" {
		t.Fatalf("unexpected trigger payload: %s", string(out))
	}
}

func TestExportYAMLWorldAsTiled(t *testing.T) {
	t.Setenv("GREYVAR_DAT_DIR", fixtureDatDir(t))

	tmp := t.TempDir()
	exported, err := ExportYAMLWorldAsTiled("yamlWorld_dev", TiledExportOptions{
		DestDir: tmp,
	})
	if err != nil {
		t.Fatalf("ExportYAMLWorldAsTiled: %v", err)
	}

	worldPath := filepath.Join(tmp, "yamlWorld_dev.world")
	data, err := os.ReadFile(worldPath)
	if err != nil {
		t.Fatalf("read world file: %v", err)
	}

	doc := tiledWorldExport{}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse world file: %v", err)
	}

	if doc.Type != "world" {
		t.Fatalf("expected type world, got %q", doc.Type)
	}
	if len(doc.Maps) != 2 {
		t.Fatalf("expected 2 map placements, got %d", len(doc.Maps))
	}

	spawn := ""
	triggersProp := ""
	for _, prop := range doc.Properties {
		switch prop.Name {
		case "spawnGrid":
			spawn, _ = prop.Value.(string)
		case "triggers":
			triggersProp, _ = prop.Value.(string)
		}
	}
	if spawn != "0.0.tmj" {
		t.Fatalf("expected spawnGrid 0.0.tmj, got %q", spawn)
	}
	if !strings.Contains(triggersProp, "fixtureTrigger") {
		t.Fatalf("expected triggers property to contain fixtureTrigger, got %q", triggersProp)
	}

	if _, err := os.Stat(filepath.Join(tmp, "0.0.tmj")); err != nil {
		t.Fatalf("expected exported tmj map: %v", err)
	}

	if len(exported.Grids) != 2 {
		t.Fatalf("expected 2 loaded tmj grids after export, got %d", len(exported.Grids))
	}
}

func TestMapPlacementsFromGrids(t *testing.T) {
	t.Setenv("GREYVAR_DAT_DIR", fixtureDatDir(t))

	world, err := LoadWorld("yamlWorld_dev")
	if err != nil {
		t.Fatalf("LoadWorld: %v", err)
	}

	placements := MapPlacementsFromGrids(world.Grids, 16)
	if len(placements) != len(world.Grids) {
		t.Fatalf("expected %d placements, got %d", len(world.Grids), len(placements))
	}

	found := false
	for _, p := range placements {
		if !strings.HasSuffix(p.FileName, ".tmj") {
			t.Fatalf("expected .tmj filename, got %q", p.FileName)
		}
		if p.FileName == "1.1.tmj" {
			found = true
			// 2x2 grid at 16px tiles = 32px per grid; row 1 col 1 -> (32, 32)
			if p.X != 32 || p.Y != 32 {
				t.Fatalf("expected 1.1 at 32,32 got %d,%d", p.X, p.Y)
			}
		}
	}
	if !found {
		t.Fatalf("expected 1.1.tmj placement")
	}
}
