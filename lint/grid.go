package lint

import (
	"fmt"
	"path/filepath"

	grids "github.com/greyvar/datlib/gridfiles"
)

// LintLegacyGrids loads every worlds/**/*.grid file.
func LintLegacyGrids(datDir string) []Check {
	worldsRoot := filepath.Join(datDir, "worlds")
	matches, err := FindFilesByExt(worldsRoot, ".grid")
	if err != nil {
		check := Check{Type: "grid", Filename: worldsRoot}
		check.AddError(fmt.Sprintf("scan grid files: %v", err))
		return []Check{check}
	}

	out := make([]Check, 0, len(matches))
	for _, match := range matches {
		out = append(out, CheckGridFile(match))
	}
	return out
}

// CheckGridFile loads one legacy .grid file.
func CheckGridFile(filename string) Check {
	check := Check{Type: "grid", Filename: filename}
	var err error
	withQuietLogs(func() {
		_, err = grids.ReadGrid(filename)
	})
	if err != nil {
		check.AddError(err.Error())
	}
	return check
}
