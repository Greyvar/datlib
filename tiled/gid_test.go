package tiled

import "testing"

func TestEncodeDecodeGIDRoundTrip(t *testing.T) {
	original := GID{ID: 42, FlipH: true, FlipV: false, Rot: 0}
	encoded := EncodeGID(original)
	decoded := DecodeGID(encoded)
	if decoded.ID != original.ID {
		t.Fatalf("id mismatch: %d vs %d", decoded.ID, original.ID)
	}
	if decoded.FlipH != original.FlipH || decoded.FlipV != original.FlipV || decoded.Rot != original.Rot {
		t.Fatalf("flags mismatch after roundtrip: %+v vs %+v", decoded, original)
	}
}

func TestDecodeGID(t *testing.T) {
	decoded := DecodeGID(0x80000005)
	if decoded.ID != 5 {
		t.Fatalf("expected id 5, got %d", decoded.ID)
	}
	if !decoded.FlipH {
		t.Fatalf("expected horizontal flip")
	}
}
