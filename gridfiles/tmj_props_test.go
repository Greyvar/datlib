package gridfiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTMJObjectPropertiesMsg(t *testing.T) {
	dir := t.TempDir()
	mapPath := filepath.Join(dir, "signs.tmj")
	tsxPath := filepath.Join(dir, "ents.tsx")

	tsx := `<?xml version="1.0"?>
<tileset name="ents" tilewidth="16" tileheight="16" tilecount="1" columns="1">
 <image source="ents.png" width="16" height="16"/>
 <tile id="0" type="sign"/>
</tileset>`
	if err := os.WriteFile(tsxPath, []byte(tsx), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ents.png"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	body := `{
  "width": 4, "height": 4, "tilewidth": 16, "tileheight": 16,
  "tilesets": [{"firstgid": 1, "source": "ents.tsx"}],
  "layers": [{
    "type": "objectgroup", "name": "entities", "objects": [{
      "id": 1, "gid": 1, "x": 16, "y": 16, "width": 16, "height": 16,
      "properties": [{"name": "msg", "type": "string", "value": "Welcome!"}]
    }]
  }]
}`
	if err := os.WriteFile(mapPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	grid, err := ReadGridTMJ(mapPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(grid.Entities) != 1 {
		t.Fatalf("entities = %d", len(grid.Entities))
	}
	ent := grid.Entities[0]
	if ent.Definition != "sign" {
		t.Fatalf("definition = %q", ent.Definition)
	}
	if ent.Properties["msg"] != "Welcome!" {
		t.Fatalf("properties = %#v", ent.Properties)
	}
}
