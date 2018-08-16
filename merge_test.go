package main

import (
	"strings"
	"testing"
)

const appetiteForTest = 5

func TestA(test *testing.T) {
	tt := []struct {
		input              []string
		A, X, B            string
		wantResult         string
		wantMarks, wantErr bool
	}{
		{
			nil,
			"tests/a.py",
			"tests/x.py",
			"tests/b.py",
			"tests/o.py",
			false,
			false,
		},
		{
			[]string{"h1", "m", "u"},
			"tests/conins/SameLine.A.java",
			"tests/conins/SameLine.X.java",
			"tests/conins/SameLine.B.java",
			"tests/conins/SameLine.O.java",
			true,
			false,
		},
		{
			[]string{"h4", "m", "u"},
			"tests/conins/SameLine.A.java",
			"tests/conins/SameLine.X.java",
			"tests/conins/SameLine.B.java",
			"tests/conins/SameLine.O.java",
			true,
			false,
		},
		{
			[]string{"m", "u"},
			"tests/conins/SameLine.A.java",
			"tests/conins/SameLine.X.java",
			"tests/conins/SameLine.B.java",
			"tests/conins/SameLine.O.h5.java",
			true,
			false,
		},
		{
			[]string{"h1", "m!"},
			"tests/conins/SameLine.A.java",
			"tests/conins/SameLine.X.java",
			"tests/conins/SameLine.B.java",
			"tests/conins/SameLine.O.mbang.java",
			true,
			false,
		},
	}
	for _, t := range tt {
		m := newMerge(
			"",
			t.A, t.X, t.B,
			appetiteForTest,
			strings.NewReader(strings.Join(t.input, "\n")+"\n"),
		)
		marks, result, err := m.merge()
		if marks != t.wantMarks {
			test.Error("Expected marks:", t.wantMarks)
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
