package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

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
	if ops[0].J2 < line {
		panic("line too big")
	} else if ops[0].J2 <= line {
		ops = ops[1:]
	}
	return ops[0].Tag, ops
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
}

func newInteractiveMerge(a, x, b string) *Merge {
	return newMerge(a, x, b, bufio.NewReader(os.Stdin))
}

func newMerge(a, x, b string, input io.Reader) *Merge {
	sample, err := read(a)
	if err != nil {
		log.Fatal(err)
	}
	return &Merge{
		a, x, b,
		newCdiff(),
		newHighlighter(sample, getFileType(a)),
		bufio.NewReader(input),
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
	fmt.Printf("XA:%v\nXB:%v\n", xA, xB)
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
		case a == 'd' || b == 'd':
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

		// conflict
		conflictA := []string{}
		for iA < len(A) {
			conflictA = append(conflictA, A[iA])
			iA++
			a, xA = getOp(iA, xA)
			if a == 'e' {
				break
			}
		}
		conflictB := []string{}
		for iB < len(B) {
			conflictB = append(conflictB, B[iB])
			iB++
			b, xB = getOp(iB, xB)
			if b == 'e' {
				break
			}
		}

		m.highlighter.printSlice(merged)
		result = append(result, merged...)

		outputMode := '-'
	resolution:
		for {
			fmt.Println("<<<<<< >>>>>>")
			diff := m.cdiff.diff(conflictA, conflictB, outputMode)
			switch outputMode {
			case 'a':
				fallthrough
			case 'b':
				m.highlighter.printString(diff)
			default:
				fmt.Print(diff)
			}
			fmt.Print("% ")
			text, err := m.reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}

			switch text[0] {
			default: // 'h'
				fmt.Println(`
	r/a: Take red/A-side
	g/b: Take green/B-side
	m: Mark conflict and continue
	c[e/m/s/l]: Change diff cleanup mode
	o[a/b]: Change diff output mode
	p: Print the Previous merged section
	h: Show this message
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
				order := 'a'
				if len(text) > 1 {
					order = rune(text[1])
				}
				switch order {
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
				continue
			// Change diff output mode
			case 'o':
				outputMode = rune(text[1])
				continue
			}
		}
		merged = []string{}
	}

	m.highlighter.printSlice(merged)
	result = append(result, merged...)

	return marked, strings.Join(result, ""), nil
}
