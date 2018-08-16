package main

import "github.com/sergi/go-diff/diffmatchpatch"
import "github.com/alecthomas/chroma"
import "github.com/alecthomas/chroma/formatters"
import "github.com/alecthomas/chroma/lexers"
import "github.com/alecthomas/chroma/styles"
import "strings"

import "log"
import "os"

// Cdiff diffs two files.
type Cdiff struct {
	dmp         *diffmatchpatch.DiffMatchPatch
	CleanupMode rune
}

func newCdiff() *Cdiff {
	return &Cdiff{
		diffmatchpatch.New(),
		's',
	}
}

func (c *Cdiff) diff(a, b []string, outputMode command) string {
	diffs := c.dmp.DiffMain(strings.Join(a, ""), strings.Join(b, ""), false)

	switch c.CleanupMode {
	case 'e':
		diffs = c.dmp.DiffCleanupEfficiency(diffs)
	case 'm':
		diffs = c.dmp.DiffCleanupMerge(diffs)
	case 's':
		diffs = c.dmp.DiffCleanupSemantic(diffs)
	case 'l':
		diffs = c.dmp.DiffCleanupSemanticLossless(diffs)
	}

	switch outputMode {
	default: // 'p':
		return c.dmp.DiffPrettyText(diffs)
	case commandTakeA:
		return c.dmp.DiffText1(diffs)
	case commandTakeB:
		return c.dmp.DiffText2(diffs)
	}
}

// Highlighter highlights a source file.
type Highlighter struct {
	lexer     chroma.Lexer
	formatter chroma.Formatter
	style     *chroma.Style
}

func newHighlighter(sample []string, lang, style string) *Highlighter {
	// Determine lexer.
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Analyse(strings.Join(sample, ""))
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	// Determine formatter.
	f := formatters.Get("terminal256")
	if f == nil {
		f = formatters.Fallback
	}

	// Determine style.
	s := styles.Get(style)
	if s == nil {
		s = styles.Fallback
	}
	return &Highlighter{
		l, f, s,
	}
}

func (h *Highlighter) printSlice(content []string) error {
	return h.printString(strings.Join(content, ""))
}
func (h *Highlighter) printString(content string) error {
	it, err := h.lexer.Tokenise(nil, content)
	if err != nil {
		log.Fatal(err)
	}
	return h.formatter.Format(os.Stdout, h.style, it)
}
