package jsonmatcher

import (
	"net/netip"
	"regexp"
	"strings"
)

var getNets = regexp.MustCompile(`(?:\d{1,3}\.){3}\d{1,3}(?:/\d{1,2})?`)

type exprNet struct {
	value netip.Prefix
}

func (e *exprNet) matches(field *field) bool {
	for _, v := range field.scalarValues() {
		sv, ok := getStringVal(v)
		if !ok {
			continue
		}
		// find all CIDRs/IPAddrs in string
		for _, ipIdx := range getNets.FindAllStringIndex(sv, -1) {
			netsv := sv[ipIdx[0]:ipIdx[1]]
			switch {
			case strings.Contains(netsv, `/`):
				// net block, check for overlap
				netblock, err := netip.ParsePrefix(netsv)
				if err != nil {
					// report incorrect field
					continue
				}
				if e.value.Overlaps(netblock) {
					return true
				}
			// would add port awareness here if needed
			// case strings.Contains(netsv, `:`):
			default:
				// plain netaddr, check if it's contained/equals
				netaddr, err := netip.ParseAddr(netsv)
				if err != nil {
					continue
				}
				if e.value.Contains(netaddr) {
					return true
				}
			}
		}
	}
	return false
}
