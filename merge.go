package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"

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
	if ops[0].I1 > line {
		panic("line too big")
	} else if ops[0].I2 == line {
		ops = ops[1:]
	}
	return ops[0].Tag, ops
}

// Merge merges three files.
type Merge struct {
	yours, base, other string
	cdiff              *Cdiff
	highlighter        *Highlighter
	reader             *bufio.Reader
}

func newMerge(a, x, b string) *Merge {
	sample, err := read(a)
	if err != nil {
		log.Fatal(err)
	}
	return &Merge{
		a, x, b,
		newCdiff(),
		newHighlighter(sample, "python"),
		bufio.NewReader(os.Stdin),
	}
}

func (m *Merge) merge() (conflict bool, result []string, err error) {
	X, err := read(m.base)
	if err != nil {
		return
	}
	A, err := read(m.yours)
	if err != nil {
		return
	}
	B, err := read(m.other)
	if err != nil {
		return
	}
	xA := difflib.NewMatcher(X, A).GetOpCodes()
	xB := difflib.NewMatcher(X, B).GetOpCodes()
	fmt.Printf("XA:%#v\nXB:%#v\n", xA, xB)
	iA := 0
	iB := 0
	var a, b byte
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
			a, xA = getOp(iA, xA)
			if a == 'e' {
				break
			}
			conflictA = append(conflictA, A[iA])
			iA++
		}
		conflictB := []string{}
		for iB < len(B) {
			b, xB = getOp(iB, xB)
			if b == 'e' {
				break
			}
			conflictB = append(conflictB, B[iB])
			iB++
		}

		conflict = true
		m.highlighter.printSlice(merged)
		result = append(result, merged...)
		merged = []string{}

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
			case 'h':
				fmt.Println(`
	r/a: Take red/A-side
	g/b: Take green/B-side
	m: Mark conflict and continue
	c[e/m/s/l]: Change diff cleanup mode
	o[a/b]: Change diff output mode
	h: Show this message
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
			// Mark the conflict and continue
			case 'm':
				result = append(result, "<<<<<< "+m.yours+"\n")
				result = append(result, conflictA...)
				result = append(result, "======\n")
				result = append(result, conflictB...)
				result = append(result, ">>>>>> "+m.other+"\n")
				break resolution
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
	}

	m.highlighter.printSlice(merged)
	result = append(result, merged...)

	fmt.Println("Final:")
	m.highlighter.printSlice(result)

	return
}
