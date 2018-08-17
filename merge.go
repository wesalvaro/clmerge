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

const interactiveUsage = `
Resolution:
  r/a[!]: Take red/A-side
  g/b[!]: Take green/B-side
  u[r/a/g/b]: Take both with A/B-side first
	m[!]: Mark conflict
Display:
  c[e/m/s/l]: Change diff cleanup mode
  o[r/a/g/b]: Change diff output mode
  p: Print the Previous merged section
Conflicting:
  h[#]: Re-conflict with different line appetite
  e: Increase line appetite (by one)
  f: Decrease line appetite (by one)
`

type resolution int

const (
	resolutionNone resolution = iota
	resolutionMark
	resolutionTakeA
	resolutionTakeB
	resolutionAlwaysMark
	resolutionAlwaysTakeA
	resolutionAlwaysTakeB
	resolutionFirstTakeA
	resolutionFirstTakeB
)

type command rune

const (
	commandNoop        command = '-'
	commandTakeRed     command = 'r'
	commandTakeA       command = 'a'
	commandTakeGreen   command = 'g'
	commandTakeB       command = 'b'
	commandTakeBoth    command = 'u'
	commandMark        command = 'm'
	commandAlways      command = '!'
	commandAppetiteSet command = 'h'
	commandAppetiteInc command = 'i'
	commandAppetiteDec command = 'd'
	commandCleanupMode command = 'c'
	commandOutputMode  command = 'o'
	commandReprint     command = 'p'
)

func read(fn string) ([]string, error) {
	x, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("Could not read: %s. (%s)", fn, err)
	}
	return difflib.SplitLines(string(x)), nil
}

func getOp(line int, ops []difflib.OpCode) (byte, []difflib.OpCode) {
	for ops[0].J2 <= line {
		ops = ops[1:]
	}
	return ops[0].Tag, ops
}

func getCommands(reader *bufio.Reader) []command {
	fmt.Print("% ")
	text, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	return []command(text)
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

func resolve(cmds []command, result, conflictA, conflictB []string) (resolution, []string) {
	always := false
	if len(cmds) > 1 && cmds[1] == commandAlways {
		always = true
	}
	switch cmds[0] {
	// Take A-side (red edits)
	case commandTakeRed:
		fallthrough
	case commandTakeA:
		r := resolutionTakeA
		if always {
			r = resolutionAlwaysTakeA
		}
		return r, append(result, conflictA...)
	// Take B-side (green edits)
	case commandTakeGreen:
		fallthrough
	case commandTakeB:
		r := resolutionTakeB
		if always {
			r = resolutionAlwaysTakeB
		}
		return r, append(result, conflictB...)
	case commandTakeBoth:
		switch cmds[1] {
		default: // commandTakeA
			return resolutionFirstTakeA, append(append(result, conflictA...), conflictB...)
		case commandTakeGreen:
			fallthrough
		case commandTakeB:
			return resolutionFirstTakeB, append(append(result, conflictB...), conflictA...)
		}
	// Mark the conflict and continue
	case commandMark:
		r := resolutionMark
		if always {
			r = resolutionAlwaysMark
		}
		result = append(result, "<<<<<<< LOCAL\n")
		result = append(result, conflictA...)
		result = append(result, "=======\n")
		result = append(result, conflictB...)
		result = append(result, ">>>>>>> OTHER\n")
		return r, result
	}
	return resolutionNone, result
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

func newInteractiveMerge(style, a, x, b string, appetite int) *Merge {
	return newMerge(style, a, x, b, appetite, bufio.NewReader(os.Stdin))
}

func newMerge(style, a, x, b string, appetite int, input io.Reader) *Merge {
	sample, err := read(a)
	if err != nil {
		log.Fatal(err)
	}
	return &Merge{
		a, x, b,
		newCdiff(),
		newHighlighter(sample, getFileType(a), style),
		bufio.NewReader(input),
		appetite,
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
	cmdDefault := commandNoop
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
		outputMode := commandNoop
		conflictA := []string{}
		conflictB := []string{}
	resolution:
		for {
			conflictA = getConflict(xA, A, iA, appetite)
			conflictB = getConflict(xB, B, iB, appetite)
			diff := m.cdiff.diff(conflictA, conflictB, outputMode)
			fmt.Print(m.cdiff.diff([]string{"<<<●"}, []string{"●>>>"}, outputMode))
			switch outputMode {
			case 'a':
				fallthrough
			case 'b':
				fmt.Println()
				m.highlighter.printString(diff)
			default:
				fmt.Print("\n", diff)
			}
			cmds := []command{cmdDefault}
			if cmds[0] == commandNoop {
				cmds = getCommands(m.reader)
			}
			r := resolutionNone
			r, result = resolve(cmds, result, conflictA, conflictB)
			if r != resolutionNone {
				switch r {
				case resolutionAlwaysTakeA:
					cmdDefault = commandTakeA
				case resolutionAlwaysTakeB:
					cmdDefault = commandTakeB
				case resolutionAlwaysMark:
					cmdDefault = commandMark
					fallthrough
				case resolutionMark:
					marked = true
				}
				break resolution
			}
			switch cmds[0] {
			// Print the previous merged section again
			case commandReprint:
				m.highlighter.printSlice(merged)
			// Change diff cleanup mode
			case commandCleanupMode:
				m.cdiff.CleanupMode = rune(cmds[1])
			case commandAppetiteSet:
				appetite = int(rune(cmds[1]) - '0')
				fmt.Printf("Appetite: %d\n", appetite)
			case commandAppetiteInc:
				appetite++
				fmt.Printf("Appetite: %d\n", appetite)
			case commandAppetiteDec:
				appetite--
				fmt.Printf("Appetite: %d\n", appetite)
			// Change diff output mode
			case commandOutputMode:
				outputMode = cmds[1]
				switch outputMode {
				case commandTakeRed:
					outputMode = commandTakeA
				case commandTakeGreen:
					outputMode = commandTakeB
				}
			default: // '?'
				fmt.Println(interactiveUsage)
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
