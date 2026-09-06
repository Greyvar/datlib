package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/greyvar/datlib/tiled"
)

func TestValidateDatDirEmpty(t *testing.T) {
	if err := ValidateDatDir(""); err == nil {
		t.Fatal("expected error for empty dat dir")
	}
}

func TestValidateDatDirIncomplete(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "worlds"), 0o755); err != nil {
		t.Fatal(err)
	}
	err := ValidateDatDir(root)
	if err == nil {
		t.Fatal("expected error when entdefs/ is missing")
	}
	if !strings.Contains(err.Error(), "entdefs/") {
		t.Fatalf("error = %q, want missing entdefs/", err)
	}
}

func TestValidateDatDirOK(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"worlds", "entdefs"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := ValidateDatDir(root); err != nil {
		t.Fatalf("ValidateDatDir: %v", err)
	}
}

func TestDiscoverDatDirDiscoversServerDat(t *testing.T) {
	root := t.TempDir()
	serverDat := filepath.Join(root, "server", "dat")
	for _, name := range []string{"worlds", "entdefs"} {
		if err := os.MkdirAll(filepath.Join(serverDat, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	toolDir := filepath.Join(root, "datlint")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(toolDir); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GREYVAR_DAT_DIR", "")

	got, err := DiscoverDatDir()
	if err != nil {
		t.Fatalf("DiscoverDatDir: %v", err)
	}
	want, err := filepath.EvalSymlinks(serverDat)
	if err != nil {
		want = filepath.Clean(serverDat)
	}
	gotEval, err := filepath.EvalSymlinks(got)
	if err != nil {
		gotEval = filepath.Clean(got)
	}
	if gotEval != want {
		t.Fatalf("DiscoverDatDir = %q, want %q", gotEval, want)
	}
}

func TestCheckEntdefOK(t *testing.T) {
	dir := t.TempDir()
	tsx := `<?xml version="1.0"?><tileset name="t" tilewidth="16" tileheight="16" tilecount="1" columns="1">
 <image source="bob.png" width="16" height="16"/>
 <tile id="0" type="player"/>
</tileset>`
	if err := os.WriteFile(filepath.Join(dir, "bob.tsx"), []byte(tsx), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "player.yml")
	body := "title: player\ninitialState: idle\nstates:\n  idle: {}\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	visuals, err := tiled.LoadEntityVisualCatalog(dir)
	if err != nil {
		t.Fatal(err)
	}
	check := CheckEntdef(path, visuals)
	if len(check.Issues) != 0 {
		t.Fatalf("issues = %+v, want none", check.Issues)
	}
}

func TestCheckEntdefMissingTilesetType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ghost.yml")
	body := "title: ghost\ninitialState: idle\nstates:\n  idle: {}\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	check := CheckEntdef(path, &tiled.EntityVisualCatalog{})
	if !check.HasErrors() {
		t.Fatal("expected missing tileset type error")
	}
}

func TestCheckTiledWorldMissingMap(t *testing.T) {
	dir := t.TempDir()
	world := filepath.Join(dir, "demo.world")
	body := `{"type":"world","maps":[{"fileName":"missing.tmj","x":0,"y":0,"width":16,"height":16}]}`
	if err := os.WriteFile(world, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	check := CheckTiledWorld(world, dir)
	if !check.HasErrors() {
		t.Fatal("expected missing map error")
	}
}

func TestCheckTMJMissingTileset(t *testing.T) {
	dir := t.TempDir()
	tmj := filepath.Join(dir, "map.tmj")
	body := `{"tilesets":[{"firstgid":1,"source":"grass.tsx"}],"layers":[]}`
	if err := os.WriteFile(tmj, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	check := CheckTMJFile(tmj, nil)
	if !check.HasErrors() {
		t.Fatal("expected error for missing tileset on tmj")
	}
}

func TestCheckTMJUnknownEntdef(t *testing.T) {
	dir := t.TempDir()
	tmj := filepath.Join(dir, "map.tmj")
	body := `{"tilesets":[],"layers":[{"type":"objectgroup","objects":[{"type":"noSuchEntity","x":0,"y":0}]}]}`
	if err := os.WriteFile(tmj, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	check := CheckTMJFile(tmj, map[string]struct{}{"chest": {}})
	if !check.HasErrors() {
		t.Fatal("expected unknown entdef error")
	}
}

func TestCheckTexdefMissingTexture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rock.yaml")
	if err := os.WriteFile(path, []byte("---\ntraversable: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	check := CheckTexdef(path, t.TempDir())
	if !check.HasErrors() {
		t.Fatal("expected missing texture error")
	}
}

func TestFindFilesByExt(t *testing.T) {
	root := t.TempDir()
	world := filepath.Join(root, "worlds", "demo")
	if err := os.MkdirAll(world, 0o755); err != nil {
		t.Fatal(err)
	}
	tmj := filepath.Join(world, "0.0.tmj")
	if err := os.WriteFile(tmj, []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(world, "old.grid"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	tmjs, err := FindFilesByExt(root, ".tmj")
	if err != nil {
		t.Fatalf("FindFilesByExt: %v", err)
	}
	if len(tmjs) != 1 || tmjs[0] != tmj {
		t.Fatalf("tmj paths = %v, want [%q]", tmjs, tmj)
	}
}

func TestCount(t *testing.T) {
	check := Check{Type: "entdef", Filename: "x.yml"}
	check.AddWarning("w")
	check.AddError("e")
	errors, warnings := Count([]Check{check})
	if errors != 1 || warnings != 1 {
		t.Fatalf("Count = %d,%d", errors, warnings)
	}
}
