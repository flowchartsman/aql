package jsonmatcher

import (
	"net/netip"
	"regexp"
	"strings"

	"github.com/valyala/fastjson"
)

var getNets = regexp.MustCompile(`(?:\d{1,3}\.){3}\d{1,3}(?:/\d{1,2})?`)

type netMatcher struct {
	value netip.Prefix
}

func (n *netMatcher) matches(values []*fastjson.Value) bool {
	for _, v := range values {
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
				if n.value.Overlaps(netblock) {
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
				if n.value.Contains(netaddr) {
					return true
				}
			}
		}
	}
	return false
}
