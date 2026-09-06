package entdefs

type EntityDefinition struct {
	Title        string                 `json:"title" yaml:"title,omitempty"`
	InitialState string                 `json:"initialState" yaml:"initialState"`
	States       map[string]EntityState `json:"states" yaml:"states"`
	// Solid entities block player movement (16×16 AABB at the entity origin).
	Solid bool `json:"solid,omitempty" yaml:"solid,omitempty"`
}

// EntityState is a named gameplay state. Visual frames come from the tileset
// tile whose type/class matches title or title_state — not from YAML.
type EntityState struct {
	Name string `json:"name" yaml:"name,omitempty"`
}
