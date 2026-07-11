package tiled

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type templateDoc struct {
	XMLName xml.Name        `xml:"template"`
	Tileset templateTileset `xml:"tileset"`
	Object  templateObject  `xml:"object"`
}

type templateTileset struct {
	FirstGID uint32 `xml:"firstgid,attr"`
	Source   string `xml:"source,attr"`
}

type templateObject struct {
	GID    uint32 `xml:"gid,attr"`
	Type   string `xml:"type,attr"`
	Class  string `xml:"class,attr"`
	Name   string `xml:"name,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
}

// Template is a Tiled object template (.tx / extensionless XML).
type Template struct {
	GID        GID
	EntityName string
}

// LoadTemplate reads a Tiled XML object template and resolves the entity
// name via the template's own tileset reference.
func LoadTemplate(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template %q: %w", path, err)
	}

	doc := templateDoc{}
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse template %q: %w", path, err)
	}

	tpl := &Template{
		GID: DecodeGID(doc.Object.GID),
	}

	entityName := strings.TrimSpace(doc.Object.Type)
	if entityName == "" {
		entityName = strings.TrimSpace(doc.Object.Class)
	}
	if entityName == "" {
		entityName = strings.TrimSpace(doc.Object.Name)
	}

	if entityName == "" && doc.Object.GID != 0 && doc.Tileset.Source != "" {
		ref := tilesetRef{
			FirstGID: doc.Tileset.FirstGID,
			Source:   doc.Tileset.Source,
		}
		catalog := &TilesetCatalog{base: make(map[uint32]tileInfo)}
		if err := catalog.loadExternalTileset(filepath.Dir(path), ref); err != nil {
			return nil, err
		}
		entityName = catalog.EntityName(tpl.GID.ID)
	}

	tpl.EntityName = entityName
	return tpl, nil
}
