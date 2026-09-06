package lint

import (
	"fmt"
	"path/filepath"

	"github.com/greyvar/datlib/entdefs"
	"github.com/greyvar/datlib/tiled"
)

// LintEntdefs loads and validates every entdefs/*.yml file against tileset types.
func LintEntdefs(datDir, resDir string) []Check {
	matches, err := filepath.Glob(filepath.Join(datDir, "entdefs", "*.yml"))
	if err != nil {
		check := Check{Type: "entdef", Filename: filepath.Join(datDir, "entdefs")}
		check.AddError(err.Error())
		return []Check{check}
	}

	visuals, visErr := tiled.LoadEntityVisualCatalog(filepath.Join(datDir, "entdefs"))
	if visErr != nil {
		check := Check{Type: "entdef", Filename: filepath.Join(datDir, "entdefs")}
		check.AddError(fmt.Sprintf("cannot load entity tilesets: %v", visErr))
		return []Check{check}
	}

	out := make([]Check, 0, len(matches))
	for _, match := range matches {
		out = append(out, CheckEntdef(match, visuals))
	}
	return out
}

// CheckEntdef loads one entdef file and returns lint issues.
func CheckEntdef(filename string, visuals *tiled.EntityVisualCatalog) Check {
	check := Check{Type: "entdef", Filename: filename}

	var entdef *entdefs.EntityDefinition
	var err error
	withQuietLogs(func() {
		entdef, err = entdefs.ReadEntdefFile(filename)
	})
	if err != nil {
		check.AddError(err.Error())
		return check
	}

	title := entdef.Title
	if title == "" {
		check.AddWarning("entdef has no title")
		title = filepath.Base(filename)
	}

	if len(entdef.States) == 0 {
		check.AddError("entdef has no states")
		return check
	}

	for name := range entdef.States {
		if _, ok := visuals.ResolveStateExact(title, name, entdef.InitialState); !ok {
			check.AddError(fmt.Sprintf(
				"state %q has no matching tileset type/class (tried %s_%s / %s / %s)",
				name, title, name, title, name,
			))
		}
	}

	return check
}
