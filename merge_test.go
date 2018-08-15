package main

import (
	"strings"
	"testing"
)

func TestA(test *testing.T) {
	tt := []struct {
		input                 []string
		A, X, B               string
		wantResult            string
		wantConflict, wantErr bool
	}{
		{
			[]string{"m", "u"},
			"tests/conins/SameLine.A.java",
			"tests/conins/SameLine.X.java",
			"tests/conins/SameLine.B.java",
			"tests/conins/SameLine.O.java",
			true,
			false,
		},
	}
	for _, t := range tt {
		m := newMerge(
			t.A, t.X, t.B,
			strings.NewReader(strings.Join(t.input, "\n")+"\n"))
		conflict, result, err := m.merge()
		if conflict != t.wantConflict {
			test.Error("Expected conflict:", conflict)
		}
		if err != nil != t.wantErr {
			test.Error("Expected error:", err)
		}
		wantSplice, err := read(t.wantResult)
		if err != nil {
			test.Fatal(err)
		}
		// TODO: Why must I trim the spaces?
		if wantResult := strings.Join(wantSplice, ""); strings.TrimSpace(result) != strings.TrimSpace(wantResult) {
			test.Error("Results didn't match")
		}
	}
}
