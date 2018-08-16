package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/pmezard/go-difflib/difflib"
)

func read(fn string) ([]string, error) {
	x, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return difflib.SplitLines(string(x)), nil
}

func getOp(line int, ops []difflib.OpCode) (byte, []difflib.OpCode) {
	for ops[0].J2 <= line {
		ops = ops[1:]
	}
	return ops[0].Tag, ops
}

func getConflict(ops []difflib.OpCode, ll []string, i int, appetite int) (conflict []string) {
	nonE := i
	h := appetite
	tag, ops := getOp(i, ops)
	for j := i; j < len(ll); j++ {
		tag, ops = getOp(j, ops)
		if tag == 'e' {
			h--
			if h <= 0 {
				break
			}
		} else {
			nonE = j - i
			h = appetite
		}
		conflict = append(conflict, ll[j])
	}
	conflict = conflict[0 : nonE+1]
	return
}

func getFileType(fileName string) string {
	parts := strings.Split(fileName, ".")
	return parts[len(parts)-1]
}

// Merge merges three files.
type Merge struct {
	yours, base, other string
	cdiff              *Cdiff
	highlighter        *Highlighter
	reader             *bufio.Reader
	appetite           int
}

func newInteractiveMerge(style, a, x, b string) *Merge {
	return newMerge(style, a, x, b, bufio.NewReader(os.Stdin))
}

func newMerge(style, a, x, b string, input io.Reader) *Merge {
	sample, err := read(a)
	if err != nil {
		log.Fatal(err)
	}
	return &Merge{
		a, x, b,
		newCdiff(),
		newHighlighter(sample, getFileType(a), style),
		bufio.NewReader(input),
		5,
	}
}

func (m *Merge) merge() (bool, string, error) {
	X, err := read(m.base)
	if err != nil {
		return false, "", err
	}
	A, err := read(m.yours)
	if err != nil {
		return false, "", err
	}
	B, err := read(m.other)
	if err != nil {
		return false, "", err
	}
	xA := difflib.NewMatcher(X, A).GetOpCodes()
	xB := difflib.NewMatcher(X, B).GetOpCodes()
	iA := 0
	iB := 0
	var a, b byte
	marked := false
	result := []string{}
	merged := []string{}
	for iA < len(A) && iB < len(B) {
		a, xA = getOp(iA, xA)
		b, xB = getOp(iB, xB)

		switch {
		// two-sided equals
		case a == b && a == 'e':
			merged = append(merged, A[iA])
			iA++
			iB++
			continue
		// two-sided deletes
		case a == b && a == 'd':
			iA++
			iB++
			continue
		// one-sided deletes
		case a == 'd' && b == 'e':
			iA++
			iB++
			continue
		case a == 'e' && b == 'd':
			iA++
			iB++
			continue
		// one-sided replacements
		case a == 'r' && b == 'e':
			merged = append(merged, A[iA])
			iA++
			iB++
			continue
		case a == 'e' && b == 'r':
			merged = append(merged, B[iB])
			iA++
			iB++
			continue
		// one-sided inserts
		case a == 'i' && b == 'e':
			merged = append(merged, A[iA])
			iA++
			continue
		case a == 'e' && b == 'i':
			merged = append(merged, B[iB])
			iB++
			continue
		}
		m.highlighter.printSlice(merged)
		result = append(result, merged...)

		// conflict
		appetite := m.appetite
		outputMode := '-'
		conflictA := []string{}
		conflictB := []string{}
	resolution:
		for {
			conflictA = getConflict(xA, A, iA, appetite)
			conflictB = getConflict(xB, B, iB, appetite)
			diff := m.cdiff.diff(conflictA, conflictB, outputMode)
			switch outputMode {
			case 'a':
				fallthrough
			case 'b':
				m.highlighter.printString(diff)
			default:
				fmt.Print(color.RedString("<<<"), "●", color.GreenString(">>>"))
				fmt.Print("\n", diff)
			}
			fmt.Print("% ")
			text, err := m.reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}

			switch text[0] {
			default: // '?'
				fmt.Println(`
	r/a: Take red/A-side
	g/b: Take green/B-side
	m: Mark conflict and continue
	c[e/m/s/l]: Change diff cleanup mode
	o[r/a/g/b]: Change diff output mode
	p: Print the Previous merged section
	h[#]: Re-conflict with different line appetite
	u[a/b]: Take the union with A/B-side first
				`)
			// Take A-side (red edits)
			case 'r':
				fallthrough
			case 'a':
				result = append(result, conflictA...)
				break resolution
			// Take B-side (green edits)
			case 'g':
				fallthrough
			case 'b':
				result = append(result, conflictB...)
				break resolution
			case 'u':
				switch rune(text[1]) {
				default: // 'a'
					result = append(result, conflictA...)
					result = append(result, conflictB...)
				case 'b':
					result = append(result, conflictB...)
					result = append(result, conflictA...)
				}
				break resolution
			// Mark the conflict and continue
			case 'm':
				result = append(result, "<<<<<< LOCAL\n")
				result = append(result, conflictA...)
				result = append(result, "======\n")
				result = append(result, conflictB...)
				result = append(result, ">>>>>> OTHER\n")
				marked = true
				break resolution
			// Print the previous merged section again
			case 'p':
				m.highlighter.printSlice(merged)
			// Change diff cleanup mode
			case 'c':
				m.cdiff.CleanupMode = rune(text[1])
			case 'h':
				appetite = int(text[1] - '0')
			// Change diff output mode
			case 'o':
				outputMode = rune(text[1])
				switch outputMode {
				case 'r':
					outputMode = 'a'
				case 'g':
					outputMode = 'b'
				}
			}
		}
		iA += len(conflictA)
		iB += len(conflictB)
		merged = []string{}
	}

	m.highlighter.printSlice(merged)
	result = append(result, merged...)

	return marked, strings.Join(result, ""), nil
}
