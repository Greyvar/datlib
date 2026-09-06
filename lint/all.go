package lint

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	datlib "github.com/greyvar/datlib/common"
)

// Options configures a full dat-tree lint pass.
type Options struct {
	DatDir string
	ResDir string // optional; resolved from DatDir / GREYVAR_RES_DIR when empty
}

// LintAll loads and validates everything under DatDir in the standard order
// used by datlint (entdefs → texdefs → grids → worlds → tmj/tilesets → tmx → unknown).
func LintAll(opts Options) []Check {
	datDir := filepath.Clean(opts.DatDir)
	resDir := opts.ResDir
	if resDir == "" {
		resDir = ResolveResDir(datDir)
	}

	entdefNames := ListEntdefNames(datDir)

	var out []Check
	out = append(out, LintEntdefs(datDir, resDir)...)
	out = append(out, LintTexdefs(datDir, resDir)...)
	out = append(out, LintLegacyGrids(datDir)...)
	out = append(out, LintWorlds(datDir)...)
	out = append(out, LintTMJMaps(datDir, entdefNames)...)
	out = append(out, LintTMXMaps(datDir)...)
	out = append(out, LintUnknownFiles(datDir)...)
	return out
}

// CommonDatDirCandidates are checked relative to the process working directory
// when GREYVAR_DAT_DIR / ./dat are unset or invalid.
var CommonDatDirCandidates = []string{
	"dat",
	"../server/dat",
	"../greyvar-server/dat",
}

// DiscoverDatDir finds a usable Greyvar dat/ directory.
func DiscoverDatDir() (string, error) {
	primary := datlib.DatDir()
	if err := ValidateDatDir(primary); err == nil {
		return filepath.Clean(primary), nil
	}

	var tried []string
	if primary != "" {
		tried = append(tried, primary)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot resolve working directory: %w", err)
	}

	for _, rel := range CommonDatDirCandidates {
		candidate := filepath.Clean(filepath.Join(cwd, rel))
		if primary != "" && candidate == filepath.Clean(primary) {
			continue
		}
		tried = append(tried, candidate)
		if err := ValidateDatDir(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("could not find a Greyvar dat/ directory (tried %v); set GREYVAR_DAT_DIR", tried)
}

// ValidateDatDir ensures path looks like a Greyvar dat/ tree.
func ValidateDatDir(datDir string) error {
	if datDir == "" {
		return errors.New("dat dir is empty")
	}

	info, err := os.Stat(datDir)
	if err != nil {
		return fmt.Errorf("dat dir %q is not usable: %w", datDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("dat dir %q is not a directory", datDir)
	}

	required := []string{"worlds", "entdefs"}
	var missing []string
	for _, name := range required {
		path := filepath.Join(datDir, name)
		st, err := os.Stat(path)
		if err != nil || !st.IsDir() {
			missing = append(missing, name+"/")
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("dat dir %q does not look like a Greyvar dat/ tree (missing %v)", datDir, missing)
	}

	return nil
}

// ResolveResDir locates the Greyvar res/ tree beside dat/.
func ResolveResDir(datDir string) string {
	if env := strings.TrimSpace(os.Getenv("GREYVAR_RES_DIR")); env != "" {
		if st, err := os.Stat(env); err == nil && st.IsDir() {
			return filepath.Clean(env)
		}
	}

	candidates := []string{
		filepath.Join(datDir, "..", "..", "res"), // server/dat -> greyvar/res
		filepath.Join(datDir, "..", "res"),       // dat next to res
	}
	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return candidate
		}
	}
	return ""
}

func entityTexturePath(resDir, texture string) string {
	return filepath.Join(resDir, "img", "textures", "entities", texture)
}

func tileTexturePath(resDir, texture string) string {
	return filepath.Join(resDir, "img", "textures", "tiles", texture)
}

// ListEntdefNames returns entdef stems (filename without .yml) under dat/entdefs.
func ListEntdefNames(datDir string) map[string]struct{} {
	out := map[string]struct{}{}
	matches, err := filepath.Glob(filepath.Join(datDir, "entdefs", "*.yml"))
	if err != nil {
		return out
	}
	for _, match := range matches {
		stem := strings.TrimSuffix(filepath.Base(match), filepath.Ext(match))
		out[stem] = struct{}{}
	}
	return out
}
