package upload_test

import (
	"testing"

	"github.com/sparkoo/racemate-desktop/pkg/upload"
)

func TestUploadFile(t *testing.T) {
	uploadErr := upload.UploadFile("c:\\Users\\michal\\AppData\\Roaming\\RaceMate\\1739478580.lap.gzip")
	if uploadErr != nil {
		t.Error("failed to upload the file", uploadErr)
	}

	t.FailNow()
}
