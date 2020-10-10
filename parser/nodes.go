package parser

// NodeType represents the type of a node in the parse tree
type NodeType int

// Node Types
const (
	_ NodeType = iota
	NodeAnd
	NodeOr
	NodeTerminal
)

// Node is a node in the query parse tree
type Node struct {
	NodeType   NodeType
	Comparison Comparison
	Left       *Node
	Right      *Node
}

// Comparison is an individual comparision operation on a terminal node
type Comparison struct {
	Op      string
	Negated bool
	Field   []string
	Values  []string
}
