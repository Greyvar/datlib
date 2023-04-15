package entdefs;

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	datlib "github.com/greyvar/datlib/common"
	log "github.com/sirupsen/logrus"
)

func ReadEntdef(name string) (*EntityDefinition, error) {
	return ReadEntdefFile(datlib.DatDir() + "/entdefs/" + name + ".yml")
}

func ReadEntdefFile(filename string) (*EntityDefinition, error) {
	log.Infof("Reading entdef %v", filename)

	file, err := ioutil.ReadFile(filename);

	if err != nil {
		log.Warnf("Cannot read entdef file: %v", err)
		return nil, err
	}

	entdef := &EntityDefinition{}

	err = yaml.UnmarshalStrict(file, &entdef);

	if err != nil {
		log.Warnf("Cannot unmarshal entdef file: %v %v", filename, err)
		return nil, err
	}

	if entdef.Texture == "" {
		log.Warnf("entdef has no texture %v", filename )
	}

	return entdef, err
}
