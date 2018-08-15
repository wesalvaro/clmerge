package main

import (
	"fmt"

	"github.com/pmezard/go-difflib/difflib"
)

func explain(codes []difflib.OpCode) {
	for _, c := range codes {
		switch c.Tag {
		case 'e':
			fmt.Printf("E A[%d:%d] == B[%d:%d]\n", c.I1, c.I2, c.J1, c.J2)
		case 'i':
			fmt.Printf("I A[%d] <-- B[%d:%d]\n", c.I1, c.J1, c.J2)
		case 'r':
			fmt.Printf("R A[%d:%d] <== B[%d:%d]\n", c.I1, c.I2, c.J1, c.J2)
		case 'd':
			fmt.Printf("D A[%d:%d] XXX\n", c.I1, c.I2)
		}
	}
}
