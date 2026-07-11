package worldfiles

import (
	"sort"
	"strconv"
	"strings"

	"github.com/greyvar/datlib/gridfiles"
)

const defaultTileSize = 16

func ParseGridFilename(name string) (row int, col int, ok bool) {
	stem := strings.TrimSuffix(name, ".grid")
	stem = strings.TrimSuffix(stem, ".tmj")
	parts := strings.Split(stem, ".")
	if len(parts) < 2 {
		return 0, 0, false
	}

	row, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}

	col, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, false
	}

	return row, col, true
}

func GridToTMJFilename(gridFilename string) string {
	return strings.TrimSuffix(gridFilename, ".grid") + ".tmj"
}

func SpawnGridToTMJ(spawnGrid string) string {
	if spawnGrid == "" {
		return ""
	}
	return GridToTMJFilename(spawnGrid)
}

func MapPlacementsFromGrids(grids map[string]*gridfiles.Grid, tileSize int) []MapPlacement {
	if tileSize <= 0 {
		tileSize = defaultTileSize
	}

	names := make([]string, 0, len(grids))
	for name := range grids {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]MapPlacement, 0, len(names))
	for _, name := range names {
		grid := grids[name]
		if grid == nil {
			continue
		}

		widthPx := int(grid.ColCount) * tileSize
		heightPx := int(grid.RowCount) * tileSize
		if widthPx <= 0 {
			widthPx = tileSize
		}
		if heightPx <= 0 {
			heightPx = tileSize
		}

		placement := MapPlacement{
			FileName: GridToTMJFilename(name),
			Width:    widthPx,
			Height:   heightPx,
		}

		if row, col, ok := ParseGridFilename(name); ok {
			placement.X = col * widthPx
			placement.Y = row * heightPx
		}

		out = append(out, placement)
	}

	return out
}

// PlacementForGrid resolves a grid's pixel placement within a world definition.
func PlacementForGrid(def *Definition, grids map[string]*gridfiles.Grid, gridId string) (MapPlacement, bool) {
	if gridId == "" {
		return MapPlacement{}, false
	}

	if def != nil {
		for _, placement := range def.Maps {
			if placement.FileName == gridId {
				return placement, true
			}
		}

		if row, col, ok := ParseGridFilename(gridId); ok {
			for _, placement := range def.Maps {
				mapRow, mapCol, mapOk := ParseGridFilename(placement.FileName)
				if mapOk && mapRow == row && mapCol == col {
					return placement, true
				}
			}
		}
	}

	grid, ok := grids[gridId]
	if !ok || grid == nil {
		return MapPlacement{}, false
	}

	tileSize := defaultTileSize
	widthPx := int(grid.ColCount) * tileSize
	heightPx := int(grid.RowCount) * tileSize
	if widthPx <= 0 {
		widthPx = tileSize
	}
	if heightPx <= 0 {
		heightPx = tileSize
	}

	placement := MapPlacement{
		FileName: gridId,
		Width:    widthPx,
		Height:   heightPx,
	}

	if row, col, ok := ParseGridFilename(gridId); ok {
		placement.X = col * widthPx
		placement.Y = row * heightPx
	}

	return placement, true
}

// ScrollDeltaBetween returns the pixel offset from one grid to another.
func ScrollDeltaBetween(def *Definition, grids map[string]*gridfiles.Grid, fromGridId string, toGridId string) (int32, int32, bool) {
	from, okFrom := PlacementForGrid(def, grids, fromGridId)
	to, okTo := PlacementForGrid(def, grids, toGridId)
	if !okFrom || !okTo {
		return 0, 0, false
	}

	return int32(to.X - from.X), int32(to.Y - from.Y), true
}
