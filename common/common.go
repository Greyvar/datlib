package common

import (
	"os"
)

func DatDir() string {
	if _, err := os.Stat("dat"); err == nil {
		return "dat"
	}

	return os.Getenv("GREYVAR_DAT_DIR")
}
