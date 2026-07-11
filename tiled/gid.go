package tiled

const (
	flippedHorizontallyFlag = 0x80000000
	flippedVerticallyFlag   = 0x40000000
	flippedDiagonallyFlag   = 0x20000000
	gidMask                 = 0x1FFFFFFF
)

type GID struct {
	ID    uint32
	FlipH bool
	FlipV bool
	Rot   int32
}

func DecodeGID(raw uint32) GID {
	flipH := raw&flippedHorizontallyFlag != 0
	flipV := raw&flippedVerticallyFlag != 0
	flipD := raw&flippedDiagonallyFlag != 0
	id := raw & gidMask

	rot := int32(0)
	if flipD {
		switch {
		case flipH && flipV:
			rot = 90
			flipH = false
			flipV = true
		case flipH:
			rot = 90
			flipH = false
			flipV = false
		case flipV:
			rot = 270
			flipH = false
			flipV = false
		default:
			rot = 90
			flipH = true
			flipV = false
		}
	}

	return GID{
		ID:    id,
		FlipH: flipH,
		FlipV: flipV,
		Rot:   rot,
	}
}

func EncodeGID(g GID) uint32 {
	raw := g.ID & gidMask
	flipH := g.FlipH
	flipV := g.FlipV
	flipD := false

	switch g.Rot {
	case 90:
		flipD = true
		if flipV {
			flipH = true
			flipV = true
		} else {
			flipH = true
			flipV = false
		}
	case 270:
		flipD = true
		flipH = false
		flipV = true
	}

	if flipH {
		raw |= flippedHorizontallyFlag
	}
	if flipV {
		raw |= flippedVerticallyFlag
	}
	if flipD {
		raw |= flippedDiagonallyFlag
	}

	return raw
}
