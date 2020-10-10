package jsonmatch

var operations = map[string]Operation{
	`>`:  opGT,
	`<`:  opLT,
	`>=`: opGTE,
	`<=`: opLTE,
	`><`: opBetween,
	`==`: opEQ,
}
