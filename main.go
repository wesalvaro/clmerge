package main

import (
	"flag"
	"io/ioutil"
	"log"
)

var base = flag.String("base", "", "Base file")
var local = flag.String("local", "", "Local file")
var other = flag.String("other", "", "Other file")
var output = flag.String("output", "", "Output file")

func main() {
	flag.Parse()

	m := newInteractiveMerge(*local, *base, *other)
	merged, result, err := m.merge()
	if err != nil {
		log.Fatal(err)
	}
	if *output != "" {
		err := ioutil.WriteFile(*output, []byte(result), 0644)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		m.highlighter.printString(result)
	}
	if !merged {
		log.Fatal("Files were not be merged completely.")
	}
}
