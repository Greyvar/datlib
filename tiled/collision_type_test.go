package tiled

import "testing"

func TestParseCollisionTypeEmptyDefaultsToTouch(t *testing.T) {
	for _, raw := range []string{"", "  ", "TOUCH", "touch", "solid"} {
		if got := ParseCollisionType(raw); got != CollisionTouch {
			t.Fatalf("ParseCollisionType(%q) = %q, want touch", raw, got)
		}
	}
	if ParseCollisionType("pressure") != CollisionPressure {
		t.Fatal("expected pressure")
	}
	if ParseCollisionType("pickup") != CollisionPickup {
		t.Fatal("expected pickup")
	}
}

func TestCollisionTypeOrDefault(t *testing.T) {
	if CollisionType("").OrDefault() != CollisionTouch {
		t.Fatal("empty OrDefault should be touch")
	}
	if CollisionPressure.OrDefault() != CollisionPressure {
		t.Fatal("pressure should be unchanged")
	}
	if !CollisionType("").BlocksMovement() {
		t.Fatal("empty should block like touch")
	}
	if CollisionPressure.BlocksMovement() {
		t.Fatal("pressure should not block")
	}
}
