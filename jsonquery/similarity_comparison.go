package jsonquery

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// regexp to quickly identify CIDRs for selection pending deeper type integration
var cidrRex = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,3}$`)

var opSim = &similarityComparison{}

type similarityComparison struct{}

func (s *similarityComparison) GetComparator(rValues []string) (Comparator, error) {
	if len(rValues) != 1 {
		return nil, fmt.Errorf("too many values for similarity comparison. want 1, got %d", len(rValues))
	}
	cString := rValues[0]
	// Similar to equality comparison, this attempts to figure out what type the user wants.  In this case it's easier because if it starts and ends with a front slash it's a regexp, otherwise it's a wildcard. Special case: boolean true will search for "truthy" things, which are: boolean true, string "true", numeric 1, string "1", while boolean false will search for "falsy" things, which are boolean false, string "false", numeric 0, and string "0"
	switch {
	case cString == "true":
		return &boolSimComparator{true}, nil
	case cString == "false":
		return &boolSimComparator{false}, nil
	case len(cString) > 2 && cString[0] == '/' && cString[len(cString)-1] == '/':
		reg, err := regexp.Compile(rValues[0][1 : len(rValues[0])-1])
		if err != nil {
			return nil, fmt.Errorf("regular expression parse err: %w", err)
		}
		return &regexpSimComparator{reg}, nil
	case cidrRex.MatchString(cString):
		println("cidr found")
		_, net, err := net.ParseCIDR(cString)
		if err != nil {
			return nil, fmt.Errorf("CIDR parse err: %w", err)
		}
		return &netSimComparator{net}, nil
	}
	// otherwise, check for wildcards
	if strings.ContainsAny(cString, `*?`) {
		return &regexpSimComparator{wildCardRegexp(cString)}, nil
	}
	// finding none, consider it a regular string comparator
	return &stringEQComparator{cString}, nil
}

type regexpSimComparator struct {
	rex *regexp.Regexp
}

func (r *regexpSimComparator) Evaluate(lValues []string) (bool, error) {
	for _, ls := range lValues {
		if r.rex.MatchString(ls) {
			return true, nil
		}
	}
	return false, nil
}

type boolSimComparator struct {
	wantTrue bool
}

func (b *boolSimComparator) Evaluate(lValues []string) (bool, error) {
	for _, ls := range lValues {
		if b.wantTrue {
			switch ls {
			case "true", "1":
				return true, nil
			}
		} else {
			switch ls {
			case "false", "0":
				return true, nil
			}
		}
	}
	return false, nil
}

func wildCardRegexp(wcString string) *regexp.Regexp {
	var reStr strings.Builder
	var accum strings.Builder
	reStr.WriteString(`(?m)^`)
	for _, c := range wcString {
		switch c {
		case '?', '*':
			// write accumulated string so far, making sure nothing is treated as regexp syntax
			reStr.WriteString(regexp.QuoteMeta(accum.String()))
			switch c {
			case '?':
				reStr.WriteString(`.`)
			case '*':
				reStr.WriteString(`.*`)
			}
			accum.Reset()
		default:
			accum.WriteRune(c)
		}
	}
	reStr.WriteString(regexp.QuoteMeta(accum.String()))
	reStr.WriteString(`$`)
	return regexp.MustCompile(reStr.String())
}

type netSimComparator struct {
	net *net.IPNet
}

func (n *netSimComparator) Evaluate(lValues []string) (bool, error) {
	for _, ls := range lValues {
		if addr := net.ParseIP(ls); addr != nil && n.net.Contains(addr) {
			return true, nil
		}
	}
	return false, nil
}
