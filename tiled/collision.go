package tiled

// CollisionRect is a blocking rectangle in tile-local pixel coordinates.
type CollisionRect struct {
	X float64
	Y float64
	W float64
	H float64
}

func collisionRectsFromObjects(objects []tsxObject) []CollisionRect {
	if len(objects) == 0 {
		return nil
	}

	out := make([]CollisionRect, 0, len(objects))
	for _, obj := range objects {
		if obj.Width <= 0 || obj.Height <= 0 {
			continue
		}
		out = append(out, CollisionRect{
			X: obj.X,
			Y: obj.Y,
			W: obj.Width,
			H: obj.Height,
		})
	}

	return out
}

func collisionRectsFromTSXTile(tile tsxTile) []CollisionRect {
	return collisionRectsFromObjects(tile.ObjectGroup.Objects)
}

// TransformCollisionRect applies tile flip flags to a tile-local collision rect.
func TransformCollisionRect(r CollisionRect, tileWidth, tileHeight float64, flipH, flipV bool) CollisionRect {
	out := r
	if flipH {
		out.X = tileWidth - r.X - r.W
	}
	if flipV {
		out.Y = tileHeight - r.Y - r.H
	}
	return out
}
