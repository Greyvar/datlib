package tiled

import "strings"

// CollisionType controls how the player interacts with an entity tile.
type CollisionType string

const (
	// CollisionTouch (default): player is blocked by collision shapes; triggers
	// when the player AABB comes within a small margin of those shapes.
	CollisionTouch CollisionType = "touch"
	// CollisionPressure: walkable; triggers when a majority of the player AABB
	// overlaps the entity collision / footprint.
	CollisionPressure CollisionType = "pressure"
	// CollisionPickup: like pressure, but picking up removes the entity and
	// adds it to inventory instead of a normal activation.
	CollisionPickup CollisionType = "pickup"
)

// ParseCollisionType maps a tileset property value to a CollisionType.
// Empty or unknown values default to CollisionTouch.
func ParseCollisionType(raw string) CollisionType {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(CollisionPressure), "press":
		return CollisionPressure
	case string(CollisionPickup), "collect":
		return CollisionPickup
	case string(CollisionTouch), "solid", "":
		return CollisionTouch
	default:
		return CollisionTouch
	}
}

// OrDefault returns CollisionTouch when c is empty; otherwise c.
func (c CollisionType) OrDefault() CollisionType {
	if c == "" {
		return CollisionTouch
	}
	return c
}

// BlocksMovement reports whether this collision mode should block walking.
// Empty collision_type is treated as touch.
func (c CollisionType) BlocksMovement() bool {
	return c.OrDefault() == CollisionTouch
}

func propertyStringValue(props []tsxProperty, name string) string {
	for _, p := range props {
		if strings.EqualFold(strings.TrimSpace(p.Name), name) {
			return strings.TrimSpace(p.Value)
		}
	}
	return ""
}
