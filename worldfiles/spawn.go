package worldfiles

import (
	"math/rand"
)

type playerSpawn struct {
	GridID string
	X      int32
	Y      int32
}

// AllSpawnPoints returns every spawn marker across all grids in the world.
func AllSpawnPoints(world *World) []playerSpawn {
	if world == nil {
		return nil
	}

	out := make([]playerSpawn, 0)
	for gridID, grid := range world.Grids {
		if grid == nil {
			continue
		}
		for _, sp := range grid.SpawnPoints {
			out = append(out, playerSpawn{
				GridID: gridID,
				X:      sp.X,
				Y:      sp.Y,
			})
		}
	}

	return out
}

// RandomPlayerSpawn picks a random spawn marker from the world.
func RandomPlayerSpawn(world *World) (gridID string, x, y int32, ok bool) {
	points := AllSpawnPoints(world)
	if len(points) == 0 {
		return "", 0, 0, false
	}

	pick := points[rand.Intn(len(points))]
	return pick.GridID, pick.X, pick.Y, true
}

// FallbackPlayerSpawn returns the center of the world's spawn grid.
func FallbackPlayerSpawn(world *World) (gridID string, x, y int32, ok bool) {
	if world == nil {
		return "", 0, 0, false
	}

	gridID = world.SpawnGrid
	grid := world.Grids[gridID]
	if grid == nil {
		return "", 0, 0, false
	}

	row := grid.RowCount / 2
	col := grid.ColCount / 2
	return gridID, int32(col * 16), int32(row * 16), true
}

// PlayerSpawnLocation picks a random world spawn marker, or falls back to spawn grid center.
func PlayerSpawnLocation(world *World) (gridID string, x, y int32, ok bool) {
	if gridID, x, y, ok = RandomPlayerSpawn(world); ok {
		return gridID, x, y, true
	}
	return FallbackPlayerSpawn(world)
}

// SpawnPointCount returns the number of spawn markers loaded for a world.
func SpawnPointCount(world *World) int {
	return len(AllSpawnPoints(world))
}
