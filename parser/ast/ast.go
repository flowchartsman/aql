package ast

import (
	"fmt"
	"net"
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
	Op    Op
	Field []string
	RVals []Val
}

func (e *ExprNode) IsNode() {}
func (e *ExprNode) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("(%s %s", e.Op, FieldString(e.Field)))
	switch len(e.RVals) {
	case 0:
	case 1:
		sb.WriteString(` `)
		sb.WriteString(e.RVals[0].ValStr())
	default:
		sb.WriteString(` `)
		sb.WriteString(`[`)
		for i, rv := range e.RVals {
			sb.WriteString(rv.ValStr())
			if i < len(e.RVals)-1 {
				sb.WriteString(`, `)
			}
		}
		sb.WriteString(`]`)
	}
	sb.WriteString(`)`)
	return sb.String()
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
		sb.WriteString(e.RVals[0].ValStr())
	default:
		sb.WriteString(`(`)
		for i, rv := range e.RVals {
			sb.WriteString(rv.ValStr())
			if i < len(e.RVals)-1 {
				sb.WriteString(`,`)
			}
		}
		sb.WriteString(`)`)
	}
	return sb.String()
}

type Val interface {
	ValStr() string
	FriendlyType() string
}

type IntVal int

func (i IntVal) ValStr() string {
	return fmt.Sprint(i)
}
func (i IntVal) Value() int {
	return int(i)
}
func (i IntVal) FriendlyType() string {
	return "integer"
}

type FloatVal float64

func (f FloatVal) ValStr() string {
	return fmt.Sprint(f)
}

func (f FloatVal) Value() float64 {
	return float64(f)
}
func (f FloatVal) FriendlyType() string {
	return "float"
}

type StringVal string

func (s StringVal) ValStr() string {
	return strconv.Quote(string(s))
}

func (s StringVal) Value() string {
	return string(s)
}
func (s StringVal) FriendlyType() string {
	return "string"
}

type BoolVal bool

func (b BoolVal) ValStr() string {
	return fmt.Sprint(b)
}

func (b BoolVal) Value() bool {
	return bool(b)
}
func (b BoolVal) FriendlyType() string {
	return "boolean"
}

type RegexpVal struct {
	sv string
	rv *regexp.Regexp
}

func NewRegexpVal(str string) *RegexpVal {
	return &RegexpVal{sv: str}
}

func (r *RegexpVal) ValStr() string {
	return fmt.Sprintf(`/%s/`, r.sv)
}

func (r *RegexpVal) FriendlyType() string {
	return "regular expression"
}

func (r *RegexpVal) Value() (*regexp.Regexp, error) {
	if r.rv == nil {
		rv, err := regexp.Compile(r.sv)
		if err != nil {
			return nil, err
		}
		r.rv = rv
	}
	return r.rv, nil
}

type NetVal struct {
	sv string
	nv *net.IPNet
}

func NewNetVal(str string) *NetVal {
	return &NetVal{sv: str}
}

func (n *NetVal) ValStr() string {
	return n.sv
}

func (n *NetVal) FriendlyType() string {
	return "net block"
}

func (n *NetVal) Value() (*net.IPNet, error) {
	if n.nv == nil {
		_, nv, err := net.ParseCIDR(n.sv)
		if err != nil {
			return nil, err
		}
		n.nv = nv
	}
	return n.nv, nil
}

type TimeVal struct {
	sv string
	tv time.Time
}

func NewTimeVal(str string) *TimeVal {
	return &TimeVal{sv: str}
}

func (t *TimeVal) ValStr() string {
	return t.sv
}

func (t *TimeVal) FriendlyType() string {
	return "timestamp"
}

func (t *TimeVal) Value() (time.Time, error) {
	if t.tv.IsZero() {
		if tv, err := time.Parse(time.RFC3339, t.sv); err == nil {
			t.tv = tv
		} else {
			if tv, err := time.Parse("2006-01-02", t.sv); err == nil {
				t.tv = tv
			} else {
				return time.Time{}, err
			}
		}
	}
	return t.tv, nil
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
