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
}

func newMerge(a, x, b string) *Merge {
	sample, err := read(a)
	if err != nil {
		log.Fatal(err)
	}
	return &Merge{
		a, x, b, newCdiff(), newHighlighter(sample, "python"),
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
		m.highlighter.print(merged)
		result = append(result, merged...)
		merged = []string{}

		fmt.Println("<<<<<< >>>>>>")
		m.cdiff.print(conflictA, conflictB)

		reader := bufio.NewReader(os.Stdin)
	resolution:
		for {
			fmt.Print("% ")
			text, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%q", text[0])
			switch text[0] {
			case 'r':
				fallthrough
			case 'a':
				result = append(result, conflictA...)
				break resolution
			case 'g':
				fallthrough
			case 'b':
				result = append(result, conflictB...)
				break resolution
			case 'm':
				result = append(result, "<<<<<<\n")
				result = append(result, conflictA...)
				result = append(result, "======\n")
				result = append(result, conflictB...)
				result = append(result, ">>>>>>\n")
				break resolution
			}
		}
	}

	m.highlighter.print(merged)
	result = append(result, merged...)

	fmt.Println("Final:")
	m.highlighter.print(result)

	return
}
