package jsonmatcher

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/flowchartsman/aql/parser/ast"
	"github.com/valyala/fastjson"
)

type netClause struct {
	value netip.Prefix
	op    ast.Op
}

func (n *netClause) matches(values []*fastjson.Value) bool {
	for _, v := range values {
		sv, ok := getStringVal(v)
		if !ok {
			continue
		}
		switch n.op {
		case ast.EQ:
			switch {
			case strings.Contains(sv, `/`):
				// net block
				pfx, err := netip.ParsePrefix(sv)
				if err != nil {
					// report incorrect field
					return false
				}
				return n.value.Overlaps(pfx)
			case strings.Contains(sv, `:`):
				// addr w/ port
				addrport, err := netip.ParseAddrPort(sv)
				if err != nil {
					return false
				}
				return n.value.Contains(addrport.Addr())
			default:
				// plain addr
				addr, err := netip.ParseAddr(sv)
				if err != nil {
					return false
				}
				return n.value.Contains(addr)
			}
		default:
			panic(fmt.Sprintf("invalid op for net value comparison: %s", n.op))
		}
	}
	return false
}
