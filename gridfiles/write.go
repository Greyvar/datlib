package gridfiles;

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	log "github.com/sirupsen/logrus"
)

func Write(g *Grid) {	
	log.Infof("%v", g.Entities)
	sg := &GridSerializable{
		RowCount: g.RowCount,
		ColCount: g.ColCount,
		Entities: g.Entities,
	}

	for _, pos := range g.CellIterator() {
		sg.Tiles = append(sg.Tiles, g.Tiles[pos.Row][pos.Col])
	}


	yml, err := yaml.Marshal(sg)
	
	if err != nil {
		log.Errorf("%v", err)
	}

	err = ioutil.WriteFile("../greyvar-server/dat/worlds/gen/grids/0.grid", yml, 0644)

	if err != nil {
		log.Errorf("%v", err)
	}

}

