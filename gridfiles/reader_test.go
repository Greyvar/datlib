package gridfiles

import (
	"testing"
)

func TestReadGridMissingFile(t *testing.T) {
	_, readError := ReadGrid("404.yml")

	if readError == nil {
		t.Errorf("Should not be able to find this grid")
	}
}
