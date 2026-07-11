package tiled

import "testing"

func TestResWebPath(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"/home/user/greyvar/res/img/textures/tilesets/grass.png", "img/textures/tilesets/grass.png"},
		{"../../res/atlas/water.png", "atlas/water.png"},
		{"/no/res/here.png", ""},
	}

	for _, tc := range cases {
		got := ResWebPath(tc.path)
		if got != tc.want {
			t.Fatalf("ResWebPath(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}
