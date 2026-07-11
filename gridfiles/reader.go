package gridfiles

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	log "github.com/sirupsen/logrus"
)

func ReadGrid(filename string) (*Grid, error) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".tmj":
		return ReadGridTMJ(filename)
	}

	return readGridYAML(filename)
}

func readGridYAML(filename string) (*Grid, error) {
	log.Infof("Loading grid: %v", filename)

	file, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	gf := GridSerializable{}

	err = yaml.UnmarshalStrict(file, &gf)
	if err != nil {
		return nil, err
	}

	g := &Grid{
		RowCount: gf.RowCount,
		ColCount: gf.ColCount,
		Entities: gf.Entities,
		LastEntityId: gf.LastEntityId,
	}

	g.Build()

	for _, t := range gf.Tiles {
		g.Tiles[t.Row][t.Col] = t
	}

	return g, err
}
