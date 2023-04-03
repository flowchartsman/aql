//go:build tools
// +build tools

package grammar

// Necessary to include dependency on pigeon
import (
	_ "github.com/mna/pigeon"
)
