package entdefs

import (
	"os"
	"path/filepath"

	datlib "github.com/greyvar/datlib/common"
	"gopkg.in/yaml.v2"
)

func entdefsDir() string {
	return filepath.Join(datlib.DatDir(), "entdefs")
}

func entdefPath(name string) string {
	return filepath.Join(entdefsDir(), name+".yml")
}

func ListEntdefs() ([]string, error) {
	dir := entdefsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".yml" {
			continue
		}

		names = append(names, name[:len(name)-4])
	}

	return names, nil
}

func WriteEntdef(name string, entdef *EntityDefinition) error {
	if err := os.MkdirAll(entdefsDir(), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(entdef)
	if err != nil {
		return err
	}

	out := append([]byte("---\n"), data...)
	return os.WriteFile(entdefPath(name), out, 0o644)
}

func DeleteEntdef(name string) error {
	return os.Remove(entdefPath(name))
}
