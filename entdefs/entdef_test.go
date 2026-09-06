package entdefs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadEntdefSolid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "chest.yml")
	body := `title: chest
solid: true
initialState: closed
states:
  closed: {}
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	def, err := ReadEntdefFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !def.Solid {
		t.Fatal("expected solid: true")
	}
	if _, ok := def.States["closed"]; !ok {
		t.Fatal("expected closed state")
	}
}

func TestReadEntdefRejectsTexture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bob.yml")
	body := `title: bob
initialState: idle
texture: bob.png
states:
  idle: {}
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadEntdefFile(path)
	if err == nil {
		t.Fatal("expected unmarshal error for unknown texture field")
	}
}
