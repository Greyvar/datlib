package worldfiles

import (
	"github.com/greyvar/datlib/gridfiles"
	"strings"
)

type MapPlacement struct {
	FileName string `json:"fileName"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

type Definition struct {
	Author     string            `yaml:"author" json:"author"`
	Title      string            `yaml:"title" json:"title"`
	SpawnGrid  string            `yaml:"spawnGrid" json:"spawnGrid"`
	Triggers   []interface{}     `yaml:"triggers,omitempty" json:"triggers,omitempty"`
	Properties map[string]string `yaml:"-" json:"-"`
	Maps       []MapPlacement    `yaml:"-" json:"maps,omitempty"`
	Format     string            `yaml:"-" json:"-"`
}

type World struct {
	ID         string
	Definition *Definition
	Title      string
	Author     string
	SpawnGrid  string
	Grids      map[string]*gridfiles.Grid
}

type Summary struct {
	ID        string
	Title     string
	SpawnGrid string
	GridCount int
	Format    string
}

func finalizeWorld(world *World) {
	if world == nil || world.Definition == nil {
		return
	}

	world.Title = world.Definition.Title
	if world.Title == "" {
		world.Title = world.Definition.Properties["title"]
	}
	if world.Title == "" {
		world.Title = world.ID
	}

	world.Author = world.Definition.Author
	if world.Author == "" {
		world.Author = world.Definition.Properties["author"]
	}

	world.SpawnGrid = world.Definition.SpawnGrid
	if world.SpawnGrid == "" {
		world.SpawnGrid = world.Definition.Properties["spawnGrid"]
	}
	if world.SpawnGrid == "" {
		for _, placement := range world.Definition.Maps {
			if strings.Contains(strings.ToLower(placement.FileName), "welcome") {
				world.SpawnGrid = placement.FileName
				break
			}
		}
	}
	if world.SpawnGrid == "" && len(world.Definition.Maps) > 0 {
		world.SpawnGrid = world.Definition.Maps[0].FileName
	}
	if world.SpawnGrid == "" {
		for name := range world.Grids {
			world.SpawnGrid = name
			break
		}
	}
}
