package jsonmatcher

import "go.uber.org/atomic"

type MatchStats struct {
	NodeName     string        `json:"node_name"`
	TimesChecked int64         `json:"times_checked"`
	TimesMatched int64         `json:"times_matched"`
	Children     []*MatchStats `json:"children,omitempty"`
}

// TODO: When MarshalScalar/UnmarshalScalar land, these types can be replaced
// with the native sync/atomic scalar types
// ref: https://github.com/golang/go/issues/56235
// ref: https://github.com/golang/go/issues/54582
type nodeStats struct {
	nodeName     string
	timesChecked atomic.Int64 `json:"times_checked"`
	timesMatched atomic.Int64 `json:"times_matched"`
}

func (ns *nodeStats) mark(matched bool) {
	ns.timesChecked.Inc()
	if matched {
		ns.timesMatched.Inc()
	}
}

func (ns *nodeStats) toStatsNode(children ...boolNode) *MatchStats {
	sn := &MatchStats{
		NodeName:     ns.nodeName,
		TimesChecked: ns.timesChecked.Load(),
		TimesMatched: ns.timesMatched.Load(),
	}
	if len(children) > 0 {
		sn.Children = make([]*MatchStats, 0, len(children))
		for _, c := range children {
			sn.Children = append(sn.Children, c.stats())
		}
	}
	return sn
}
