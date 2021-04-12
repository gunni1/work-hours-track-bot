package main

import (
	"testing"
)

func TestParseWorkTimeFloat(t *testing.T) {
	testsets := []struct {
		Input    string
		ExpFloat float64
	}{
		{"0", 0},
		{"1,5", 1.5},
		{"1,7500000", 1.75},
		{"10.25", 10.25},
	}
	for _, testset := range testsets {
		result, err := ParseFloat(testset.Input)
		if err != nil {
			t.Error(err)
		}
		if testset.ExpFloat != result {
			t.Errorf("expected: %f to be equal to result: %f", testset.ExpFloat, result)
		}
	}
}
