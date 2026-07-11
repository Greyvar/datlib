package entdefs;

type EntityDefinition struct {
	Title        string                  `json:"title" yaml:"title,omitempty"`
	InitialState string                  `json:"initialState" yaml:"initialState"`
	States       map[string]EntityState  `json:"states" yaml:"states"`
	Texture      string                  `json:"texture" yaml:"texture,omitempty"`
}

type EntityState struct {
	Name   string  `json:"name" yaml:"name"`
	Frames []int32 `json:"frames" yaml:"frames"`
}
