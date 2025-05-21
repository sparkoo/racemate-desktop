package upload_test

import (
	"testing"

	"github.com/sparkoo/racemate-desktop/pkg/state"
	"github.com/sparkoo/racemate-desktop/pkg/upload"
)

func TestUploadFile(t *testing.T) {
	appState := &state.AppState{
		UploadURL: "https://unauthorizedupload-hwppiybqxq-ey.a.run.app",
	}
	uploadErr := upload.UploadFile("c:\\Users\\michal\\AppData\\Roaming\\RaceMate\\1739478580.lap.gzip", appState)
	if uploadErr != nil {
		t.Error("failed to upload the file", uploadErr)
	}

	t.FailNow()
}
