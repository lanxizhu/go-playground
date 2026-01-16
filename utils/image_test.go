package utils

import "testing"

func TestFitImage(t *testing.T) {
	if err := FitImage("./original.png", 8); err != nil {
		t.Errorf("Error opening image: %v", err)
		return
	}
}
