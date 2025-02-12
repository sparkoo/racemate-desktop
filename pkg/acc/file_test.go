package acc

import (
	"testing"
)

func TestReadFile(t *testing.T) {
	lap, err := loadFromFileCompressed("../../1739392703.lap.gzip")
	if err != nil {
		t.Error("Failed to read the file", err)
	}

	t.Logf("lap data: %+v\n", lap)
	t.Log(lap.Track)
	t.Log(lap.CarModel)
	t.Log(lap.PlayerName)
	t.Log(lap.PlayerSurname)
}
