package main

import (
	"flag"
	"io/ioutil"
	"log"
)

var style = flag.String("style", "vim", "Highlight style")

var base = flag.String("base", "", "Base file")
var local = flag.String("local", "", "Local file")
var other = flag.String("other", "", "Other file")
var output = flag.String("output", "", "Output file")

var appetite = flag.Int("appetite", 5, "Line appetite when looking for conflict chunks")

func main() {
	flag.Parse()

	if *local == "" || *base == "" || *other == "" {
		log.Fatal("Set `local`, `base`, and `other` file flags.")
	}

	m := newInteractiveMerge(*style, *local, *base, *other, *appetite)
	marks, result, err := m.merge()
	if err != nil {
		log.Fatal(err)
	}
	if *output != "" {
		err := ioutil.WriteFile(*output, []byte(result), 0644)
		if err != nil {
			log.Fatalf("Could not write output %s", err)
		}
	} else {
		m.highlighter.printString(result)
	}
	if marks {
		log.Fatal("Files were not be merged completely.")
	}
}
