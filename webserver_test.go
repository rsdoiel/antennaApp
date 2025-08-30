package antennaApp

import (
	"testing"
)

func TestIsDotPath(t *testing.T) {
	boolTests := map[string]bool{
		"":                        false,
		".":                       false,
		"./":                      false,
		"..":                      false,
		"recent/articles":         false,
		"./something else":        false,
		"./something/.git/config": true,
		"../../../":               false,
		".git":                    true,
		".ssh":                    true,
		"../../reoirwepoiewr/../poierwer/../.git/ewrpoiewrrwe/../../": false,
		"../../reoirwepoiewr/../poierwer/../.git/ewrpoiewrrwe/..":     true,
	}

	for p, expected := range boolTests {
		r := IsDotPath(p)
		if r != expected {
			t.Errorf("expected %t, got %t for %s", expected, r, p)
		}
	}
}
