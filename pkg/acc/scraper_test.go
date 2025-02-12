package acc

import "testing"

func TestConvertToString(t *testing.T) {
	arr := []uint16{72, 117, 110, 103, 97, 114, 111, 114, 105, 110, 103, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	result := uint16SliceToString(arr)
	if result != "hello" {
		t.Errorf("failed to convert the string: '%s'", result)
	}
}
