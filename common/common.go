package common

import (
	"os"
)

func DatDir() string {
	return os.Getenv("GREYVAR_DAT_DIR")
}
