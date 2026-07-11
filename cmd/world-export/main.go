package main

import (
	"flag"
	"fmt"
	"os"

	datlib "github.com/greyvar/datlib/common"
	"github.com/greyvar/datlib/worldfiles"
	log "github.com/sirupsen/logrus"
)

func main() {
	datDir := flag.String("dat", "", "path to dat directory (default: ./dat or GREYVAR_DAT_DIR)")
	worldName := flag.String("world", "", "YAML world id to export (e.g. isleOfStarting_dev)")
	all := flag.Bool("all", false, "export all YAML worlds")
	outDir := flag.String("out", "", "output directory (default: <worldDir>_tiled)")
	flag.Parse()

	if *datDir != "" {
		os.Setenv("GREYVAR_DAT_DIR", *datDir)
	}

	if *all {
		exportAll(*outDir)
		return
	}

	if *worldName == "" {
		log.Fatal("usage: world-export -world <id> | -all  (set -dat if needed)")
	}

	exportOne(*worldName, *outDir)
}

func exportAll(outDir string) {
	worlds, err := worldfiles.ListYAMLWorlds()
	if err != nil {
		log.Fatalf("list worlds: %v", err)
	}

	for _, summary := range worlds {
		dest := outDir
		if dest != "" {
			dest = fmt.Sprintf("%s/%s", dest, summary.ID)
		}
		exportOne(summary.ID, dest)
	}
}

func exportOne(name, outDir string) {
	opts := worldfiles.TiledExportOptions{DestDir: outDir}
	exported, err := worldfiles.ExportYAMLWorldAsTiled(name, opts)
	if err != nil {
		log.Fatalf("export %s: %v", name, err)
	}

	dest := opts.DestDir
	if dest == "" {
		dest = worldfiles.WorldDir(name) + "_tiled"
	}

	log.Infof("exported %s -> %s (%d maps, spawn %s)", name, dest, len(exported.Grids), exported.SpawnGrid)
	log.Infof("dat dir: %s", datlib.DatDir())
}
