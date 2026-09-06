package tiled

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTilesetFileOK(t *testing.T) {
	dir := t.TempDir()
	png := filepath.Join(dir, "sheet.png")
	if err := os.WriteFile(png, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	tsx := filepath.Join(dir, "sheet.tsx")
	body := `<?xml version="1.0"?>
<tileset name="sheet" tilewidth="16" tileheight="16" tilecount="4" columns="2">
 <image source="sheet.png" width="32" height="32"/>
</tileset>`
	if err := os.WriteFile(tsx, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	catalog, warnings, err := LoadTilesetFile(tsx)
	if err != nil {
		t.Fatalf("LoadTilesetFile: %v", err)
	}
	if catalog == nil {
		t.Fatal("expected catalog")
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v, want none", warnings)
	}
}

func TestLoadTilesetFileMissingImage(t *testing.T) {
	dir := t.TempDir()
	tsx := filepath.Join(dir, "sheet.tsx")
	body := `<?xml version="1.0"?>
<tileset name="sheet" tilewidth="16" tileheight="16" tilecount="4" columns="2">
 <image source="missing.png" width="32" height="32"/>
</tileset>`
	if err := os.WriteFile(tsx, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := LoadTilesetFile(tsx)
	if err == nil {
		t.Fatal("expected missing image error")
	}
	if !strings.Contains(err.Error(), "missing.png") {
		t.Fatalf("error = %q, want missing.png", err)
	}
}
