package main

import "testing"

func TestDist(t *testing.T) {
	if Dist(0, 0, 47.4049323, 8.6071845) != 5343.537867470866 {
		t.Errorf("failed")
	}
	if Dist(47.4049323, 8.6071845, 0, 0) != 5343.537867470866 {
		t.Errorf("failed")
	}
}
