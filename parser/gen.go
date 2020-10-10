// +build gen

// dummy package for dependencies
package parser

import (
	_ "github.com/mna/pigeon"
)

//go:generate pigeon -no-recover -o parser.go aql.peg
