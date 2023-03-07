package jsonmatcher

import "go.uber.org/atomic"

type StatsNode struct {
	NodeName     string       `json:"node_name"`
	TimesChecked int64        `json:"times_checked`
	TimesMatched int64        `json:"times_matched`
	Children     []*StatsNode `json:"children,omitempty"`
}

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

func (ns *nodeStats) toStatsNode(children ...matcherNode) *StatsNode {
	sn := &StatsNode{
		NodeName:     ns.nodeName,
		TimesChecked: ns.timesChecked.Load(),
		TimesMatched: ns.timesMatched.Load(),
	}
	if len(children) > 0 {
		sn.Children = make([]*StatsNode, 0, len(children))
		for _, c := range children {
			sn.Children = append(sn.Children, c.stats())
		}
	}
	return sn
}
