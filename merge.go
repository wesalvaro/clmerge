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

const interactiveUsage = `Interactive Command List:

Resolution:
  r/a: Take red/A-side
  g/b: Take green/B-side
	u[r/a/g/b]: Take both with A/B-side first
	m: Mark conflict and continue
Display:
  c[e/m/s/l]: Change diff cleanup mode
  o[r/a/g/b]: Change diff output mode
	p: Print the Previous merged section
Conflicting:
	h[#]: Re-conflict with different line appetite
	e: Increase line appetite (by one)
	f: Decrease line appetite (by one)
`

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

func merge(a, b byte, merged, A, B []string, iA, iB int) (bool, []string, int, int) {
	switch {
	// two-sided equals
	case a == b && a == 'e':
		return true, append(merged, A[iA]), iA + 1, iB + 1
	// two-sided deletes
	case a == b && a == 'd':
		return true, merged, iA + 1, iB + 1
	// one-sided deletes
	case a == 'd' && b == 'e':
		return true, merged, iA + 1, iB + 1
	case a == 'e' && b == 'd':
		return true, merged, iA + 1, iB + 1
	// one-sided replacements
	case a == 'r' && b == 'e':
		return true, append(merged, A[iA]), iA + 1, iB + 1
	case a == 'e' && b == 'r':
		return true, append(merged, B[iB]), iA + 1, iB + 1
	// one-sided inserts
	case a == 'i' && b == 'e':
		return true, append(merged, A[iA]), iA + 1, iB
	case a == 'e' && b == 'i':
		return true, append(merged, B[iB]), iA, iB + 1
	}
	return false, merged, iA, iB
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
	marked := false
	result := []string{}
	merged := []string{}
	for iA, iB := 0, 0; iA < len(A) && iB < len(B); {
		var a, b byte
		a, xA = getOp(iA, xA)
		b, xB = getOp(iB, xB)
		var ok bool
		if ok, merged, iA, iB = merge(a, b, merged, A, B, iA, iB); ok {
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
				fmt.Println(interactiveUsage)
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
				case 'g':
					fallthrough
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
			case 'e':
				appetite++
			case 'f':
				appetite--
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
