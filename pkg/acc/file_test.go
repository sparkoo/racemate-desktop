package acc

import (
	"fmt"
	"testing"
)

func TestReadFile(t *testing.T) {
	frames, err := loadFromFile("../../1739041520.gob")
	if err != nil {
		t.Error(fmt.Errorf("failed to read file: %w", err))
	}

	for _, f := range frames {
		fmt.Printf("frame: %+v\n", f)
	}
}
