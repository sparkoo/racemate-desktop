package upload_test

import (
	"testing"

	"github.com/sparkoo/racemate-desktop/pkg/upload"
)

func TestUploadFile(t *testing.T) {
	uploadErr := upload.UploadFile("../../1739393602.lap.gzip")
	if uploadErr != nil {
		t.Error("failed to upload the file", uploadErr)
	}

	t.FailNow()
}
