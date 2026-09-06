package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/greyvar/datlib/texdefs"
)

// LintTexdefs loads and validates dat/texdefs/tiles/*.yaml.
func LintTexdefs(datDir, resDir string) []Check {
	dir := filepath.Join(datDir, "texdefs", "tiles")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		check := Check{Type: "texdef", Filename: dir}
		check.AddError(err.Error())
		return []Check{check}
	}

	var out []Check
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		out = append(out, CheckTexdef(filepath.Join(dir, name), resDir))
	}
	return out
}

// CheckTexdef loads one tile texdef and verifies its texture exists under res/.
func CheckTexdef(path, resDir string) Check {
	check := Check{Type: "texdef", Filename: path}

	def, err := texdefs.ReadTiledefFile(path)
	if err != nil {
		check.AddError(err.Error())
		return check
	}

	if resDir == "" {
		check.AddWarning("res dir not found; skipped texture existence check")
		return check
	}

	texture := def.TextureName
	if texture == "" {
		check.AddWarning("texdef has no texture name")
		return check
	}

	texPath := tileTexturePath(resDir, texture)
	if _, err := os.Stat(texPath); err != nil {
		check.AddError(fmt.Sprintf("missing texture %q (looked in %s)", texture, filepath.Dir(texPath)))
	}

	return check
}
