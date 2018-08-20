package main

import (
	"strconv"
	"strings"
	"testing"
)

const appetiteForTest = 5

func TestA(test *testing.T) {
	tt := []struct {
		input              []string
		id                 int
		A, X, B            string
		wantResult         string
		wantMarks, wantErr bool
	}{
		{
			nil,
			1,
			"a.py",
			"x.py",
			"b.py",
			"o.py",
			false,
			false,
		},
		{
			nil,
			2,
			"a.go",
			"x.go",
			"b.go",
			"o.go",
			false,
			false,
		},
		{
			[]string{"h1", "m", "u"},
			3,
			"SameLine.A.java",
			"SameLine.X.java",
			"SameLine.B.java",
			"SameLine.O.java",
			true,
			false,
		},
		{
			[]string{"h4", "m", "u"},
			3,
			"SameLine.A.java",
			"SameLine.X.java",
			"SameLine.B.java",
			"SameLine.O.java",
			true,
			false,
		},
		{
			[]string{"m", "u"},
			3,
			"SameLine.A.java",
			"SameLine.X.java",
			"SameLine.B.java",
			"SameLine.O.h5.java",
			true,
			false,
		},
		{
			[]string{"h1", "m!"},
			3,
			"SameLine.A.java",
			"SameLine.X.java",
			"SameLine.B.java",
			"SameLine.O.mbang.java",
			true,
			false,
		},
	}
	for _, t := range tt {
		m := newMerge(
			"",
			"tests/"+strconv.Itoa(t.id)+"/"+t.A,
			"tests/"+strconv.Itoa(t.id)+"/"+t.X,
			"tests/"+strconv.Itoa(t.id)+"/"+t.B,
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
		wantSplice, err := read("tests/" + strconv.Itoa(t.id) + "/" + t.wantResult)
		if err != nil {
			test.Fatal(err)
		}
		// TODO: Why must I trim the spaces?
		if wantResult := strings.Join(wantSplice, ""); strings.TrimSpace(result) != strings.TrimSpace(wantResult) {
			test.Error("Results didn't match")
		}
	}
}
