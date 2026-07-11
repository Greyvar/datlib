package texdefs

import (
	"os"
	"path/filepath"
	"strings"

	datlib "github.com/greyvar/datlib/common"
	"gopkg.in/yaml.v2"
)

type TileDefinition struct {
	Texture     string `yaml:"texture,omitempty"`
	Traversable   *bool  `yaml:"traversable,omitempty"`
	Filename      string `yaml:"-"`
	TextureName   string `yaml:"-"`
}

func tilesDir() string {
	return filepath.Join(datlib.DatDir(), "texdefs", "tiles")
}

func tiledefPath(name string) string {
	stem := strings.TrimSuffix(name, ".yaml")
	stem = strings.TrimSuffix(stem, ".yml")
	return filepath.Join(tilesDir(), stem+".yaml")
}

func textureNameForFile(filename string, def *TileDefinition) string {
	if def != nil && def.Texture != "" {
		return def.Texture
	}

	return strings.TrimSuffix(filename, ".yaml") + ".png"
}

func ListTiledefs() ([]*TileDefinition, error) {
	dir := tilesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*TileDefinition{}, nil
		}
		return nil, err
	}

	out := make([]*TileDefinition, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		def, err := ReadTiledefFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		out = append(out, def)
	}

	return out, nil
}

func ReadTiledef(name string) (*TileDefinition, error) {
	return ReadTiledefFile(tiledefPath(name))
}

func ReadTiledefFile(path string) (*TileDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	def := &TileDefinition{}
	if err := yaml.Unmarshal(data, def); err != nil {
		return nil, err
	}

	def.Filename = filepath.Base(path)
	def.TextureName = textureNameForFile(def.Filename, def)
	return def, nil
}

func WriteTiledef(name string, def *TileDefinition) error {
	if err := os.MkdirAll(tilesDir(), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(def)
	if err != nil {
		return err
	}

	out := append([]byte("---\n"), data...)
	return os.WriteFile(tiledefPath(name), out, 0o644)
}

func DeleteTiledef(name string) error {
	return os.Remove(tiledefPath(name))
}
