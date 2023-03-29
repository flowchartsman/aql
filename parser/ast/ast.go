package ast

import (
	"errors"
	"fmt"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Node interface {
	IsNode()
	String() string
}

type AndNode struct {
	Left  Node
	Right Node
}

func (a *AndNode) IsNode() {}

func (a *AndNode) String() string {
	return fmt.Sprintf("(&& %s %s)", a.Left.String(), a.Right.String())
}

type OrNode struct {
	Left  Node
	Right Node
}

func (o *OrNode) IsNode() {}
func (o *OrNode) String() string {
	return fmt.Sprintf("(|| %s %s)", o.Left.String(), o.Right.String())
}

type NotNode struct {
	Expr Node
}

func (n *NotNode) IsNode() {}
func (n *NotNode) String() string {
	return fmt.Sprintf("(! %s)", n.Expr.String())
}

type SubdocNode struct {
	Field []string
	Expr  Node
}

func (s *SubdocNode) IsNode() {}
func (s *SubdocNode) String() string {
	return fmt.Sprintf(`(%s{%s})`, FieldString(s.Field), s.Expr.String())
}

type ExprNode struct {
	Op       Op
	Field    []string
	RVals    []Val
	Position Pos
}

func (e *ExprNode) IsNode() {}
func (e *ExprNode) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("(%s %s", e.Op, FieldString(e.Field)))
	switch len(e.RVals) {
	case 0:
	case 1:
		sb.WriteString(` `)
		sb.WriteString(e.RVals[0].String())
	default:
		sb.WriteString(` `)
		sb.WriteString(`[`)
		for i, rv := range e.RVals {
			sb.WriteString(rv.String())
			if i < len(e.RVals)-1 {
				sb.WriteString(`, `)
			}
		}
		sb.WriteString(`]`)
	}
	sb.WriteString(`)`)
	return sb.String()
}

func (e *ExprNode) Pos() Pos {
	return e.Position
}

func (e *ExprNode) FriendlyString() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`%s:`, FieldString(e.Field)))
	if e.Op != `==` {
		sb.WriteString(string(e.Op))
	}
	sb.WriteString(` `)

	switch len(e.RVals) {
	case 0:
	case 1:
		sb.WriteString(e.RVals[0].String())
	default:
		sb.WriteString(`(`)
		for i, rv := range e.RVals {
			sb.WriteString(rv.String())
			if i < len(e.RVals)-1 {
				sb.WriteString(`,`)
			}
		}
		sb.WriteString(`)`)
	}
	return sb.String()
}

// Pos represents the position of a node or token in the text.
type Pos struct {
	// Line is a 1-based integer representing the line on which the token was.
	// found.
	Line int `json:"line"`
	// Col is a 1-based integer representing the rune offset of the token on the line.
	Col int `json:"column"`
	// Offset is a 0-based offset of the token in the entire input text.
	Offset int `json:"offset"`
	// Len is the length of the token from Offset, in runes
	Len int `json:"length"`
}

func (p Pos) Start() Pos {
	if p.Line == 0 || p.Col == 0 || p.Offset == -1 {
		return p
	}
	return Pos{
		Line:   p.Line,
		Col:    p.Col,
		Offset: p.Offset,
		Len:    1,
	}
}

func (p Pos) IsZero() bool {
	return p == noPosition
}

// NoPosition returns a position with a negative offset for messages and errors
// that are not attached to a query feature.
func NoPosition() Pos {
	return noPosition
}

var noPosition = Pos{
	Line:   0,
	Col:    0,
	Offset: -1,
	Len:    0,
}

type ValType string

const (
	TypeInt    ValType = "integer"
	TypeFloat  ValType = "float"
	TypeString ValType = "string"
	TypeBool   ValType = "boolean"
	TypeRegex  ValType = "regex"
	TypeNet    ValType = "netaddr"
	TypeTime   ValType = "timestamp"
)

type Val interface {
	String() string
	Type() ValType
	Pos() Pos
}

type IntVal struct {
	iv  int64
	sv  string
	pos Pos
}

func NewIntVal(b []byte, pos Pos) (*IntVal, error) {
	sv := string(b)
	iv, err := strconv.ParseInt(sv, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer value [%s]", sv)
	}
	return &IntVal{
		iv:  iv,
		sv:  sv,
		pos: pos,
	}, nil
}

func (i *IntVal) String() string {
	return i.sv
}

func (i *IntVal) Value() int64 {
	return i.iv
}

func (i *IntVal) Type() ValType {
	return TypeInt
}

func (i *IntVal) Pos() Pos {
	return i.pos
}

type FloatVal struct {
	fv  float64
	sv  string
	pos Pos
}

func NewFloatVal(b []byte, pos Pos) (*FloatVal, error) {
	sv := string(b)
	fv, err := strconv.ParseFloat(sv, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid float value [%s]", sv)
	}
	return &FloatVal{
		fv:  fv,
		sv:  strconv.FormatFloat(fv, 'f', -1, 64),
		pos: pos,
	}, nil
}

func (f *FloatVal) String() string {
	return f.sv
}

func (f *FloatVal) Value() float64 {
	return f.fv
}

func (f *FloatVal) Type() ValType {
	return TypeFloat
}

func (f *FloatVal) Pos() Pos {
	return f.pos
}

type StringVal struct {
	sv  string
	qsv string
	pos Pos
}

func NewStringVal(b []byte, pos Pos) (*StringVal, error) {
	sv := string(b)
	return &StringVal{
		sv:  sv,
		qsv: strconv.Quote(sv),
		pos: pos,
	}, nil
}

func (s *StringVal) String() string {
	return s.qsv
}

func (s *StringVal) Value() string {
	return s.sv
}

func (s *StringVal) Type() ValType {
	return TypeString
}

func (s *StringVal) Pos() Pos {
	return s.pos
}

type BoolVal struct {
	bv  bool
	sv  string
	pos Pos
}

func NewBoolVal(b []byte, pos Pos) (*BoolVal, error) {
	sv := string(b)
	bv, err := strconv.ParseBool(strings.ToLower(sv))
	if err != nil {
		return nil, fmt.Errorf("invalid boolean value [%s]", sv)
	}
	sv2 := "false"
	if bv {
		sv2 = "true"
	}
	return &BoolVal{
		bv:  bv,
		sv:  sv2,
		pos: pos,
	}, nil
}

func (b *BoolVal) String() string {
	return b.sv
}

func (b *BoolVal) Value() bool {
	return bool(b.bv)
}

func (b *BoolVal) Type() ValType {
	return TypeBool
}

func (b *BoolVal) Pos() Pos {
	return b.pos
}

type RegexpVal struct {
	sv  string
	rv  *regexp.Regexp
	pos Pos
}

func NewRegexpVal(b []byte, pos Pos) (*RegexpVal, error) {
	sv := string(b)
	rv, err := regexp.Compile(strings.Trim(sv, `/`))
	if err != nil {
		return nil, fmt.Errorf("invalid regular expression [%s]: %v", sv, err)
	}
	return &RegexpVal{
		sv:  sv,
		rv:  rv,
		pos: pos,
	}, nil
}

func (r *RegexpVal) String() string {
	return r.sv
}

func (r *RegexpVal) Type() ValType {
	return TypeRegex
}

func (r *RegexpVal) Value() *regexp.Regexp {
	return r.rv
}

func (r *RegexpVal) Pos() Pos {
	return r.pos
}

type NetVal struct {
	sv  string
	nv  netip.Prefix
	pos Pos
}

var (
	removeParsePrefix = regexp.MustCompile(`^netip\.ParsePrefix\([^)]*\): `)
	removeParseAddr   = regexp.MustCompile(`^ParseAddr\([^)]*\): `)
)

func NewNetVal(b []byte, pos Pos) (*NetVal, error) {
	sv := string(b)
	if !strings.Contains(sv, `/`) {
		sv += "/32"
	}
	nv, err := netip.ParsePrefix(sv)
	if err != nil {
		errstr := err.Error()
		errstr = removeParsePrefix.ReplaceAllLiteralString(errstr, "")
		errstr = removeParseAddr.ReplaceAllLiteralString(errstr, "")
		return nil, fmt.Errorf("invalid network value [%s]: %s", sv, errstr)
	}
	return &NetVal{
		sv:  sv,
		nv:  nv,
		pos: pos,
	}, nil
}

func (n *NetVal) String() string {
	return n.sv
}

func (n *NetVal) Type() ValType {
	return TypeNet
}

func (n *NetVal) Value() netip.Prefix {
	return n.nv
}

func (n *NetVal) Pos() Pos {
	return n.pos
}

type TimeVal struct {
	sv  string
	tv  time.Time
	pos Pos
}

func NewTimeVal(b []byte, pos Pos) (*TimeVal, error) {
	sv := string(b)
	var (
		tv  time.Time
		err error
	)
	if len(sv) == 10 {
		tv, err = time.Parse(`2006-01-02`, sv)
	} else {
		tv, err = time.Parse(time.RFC3339, sv)
	}
	var message string
	if err != nil {
		var tpe *time.ParseError
		if errors.As(err, &tpe) {
			message = strings.TrimPrefix(tpe.Message, ": ")
		} else {
			message = err.Error()
		}
		return nil, fmt.Errorf("invalid datetime value [%s]: %s", sv, message)
	}

	return &TimeVal{
		sv:  sv,
		tv:  tv,
		pos: pos,
	}, nil
}

func (t *TimeVal) String() string {
	return t.sv
}

func (t *TimeVal) Value() time.Time {
	return t.tv
}

func (t *TimeVal) Type() ValType {
	return TypeTime
}

func (t *TimeVal) Pos() Pos {
	return t.pos
}

func FieldString(pathparts []string) string {
	var sb strings.Builder
	for i, p := range pathparts {
		if strings.ContainsAny(p, ` ."`) {
			sb.WriteString(strconv.Quote(p))
		} else {
			sb.WriteString(p)
		}
		if i < len(pathparts)-1 {
			sb.WriteString(`.`)
		}
	}
	return sb.String()
}

type Op string

const (
	EQ  Op = `==`
	LT  Op = `<`
	LTE Op = `<=`
	GT  Op = `>`
	GTE Op = `>=`
	BET Op = `><`
	SIM Op = `~`
	EXS Op = `exists`
	NUL Op = `null`
)
