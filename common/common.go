package common

import (
	"os"
)

func DatDir() string {
	if _, err := os.Stat("dat"); err == nil {
		// Return as a absolute path
		absPath, _ := os.Getwd()
		return absPath + "/dat"
	}


	return os.Getenv("GREYVAR_DAT_DIR")
}
