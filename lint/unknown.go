package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var knownDatExtensions = map[string]struct{}{
	".yml":   {},
	".yaml":  {},
	".tmj":   {},
	".tmx":   {},
	".tsx":   {},
	".tx":    {},
	".grid":  {},
	".world": {},
}

// LintUnknownFiles warns on unrecognized files under dat/.
func LintUnknownFiles(datDir string) []Check {
	var out []Check
	err := filepath.WalkDir(datDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			base := d.Name()
			if base == ".git" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		base := d.Name()
		if strings.HasPrefix(base, ".") {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(base))
		if ext == "" {
			check := Check{Type: "unknown", Filename: path}
			check.AddWarning("unrecognized dat file (no extension)")
			out = append(out, check)
			return nil
		}

		if _, ok := knownDatExtensions[ext]; ok {
			return nil
		}

		check := Check{Type: "unknown", Filename: path}
		check.AddWarning(fmt.Sprintf("unrecognized dat file type %q", ext))
		out = append(out, check)
		return nil
	})
	if err != nil {
		check := Check{Type: "unknown", Filename: datDir}
		check.AddError(fmt.Sprintf("scan dat tree: %v", err))
		out = append(out, check)
	}
	return out
}

// FindFilesByExt walks root and returns files with the given extension.
func FindFilesByExt(root, ext string) ([]string, error) {
	ext = strings.ToLower(ext)
	var out []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) == ext {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}
