package profiler

import (
	"fmt"
	"github.com/anhcraft/rice/exec/ast"
	"sort"
	"strings"
	"time"
)

type AggregatedRecord struct {
	Label         string
	Pos           ast.Pos
	Invocations   int
	TotalDuration time.Duration
	Children      []*AggregatedRecord
}

func (p *Impl) Report() string {
	if p.root == nil {
		return "(No record)"
	}

	aggr := p.aggregate(p.root)

	var sb strings.Builder
	p.buildReportString(&sb, aggr, 0, aggr.TotalDuration)

	return sb.String()
}

func (p *Impl) aggregate(raw *ProfileRecord) *AggregatedRecord {
	agg := &AggregatedRecord{
		Label:         raw.Label,
		Pos:           raw.Pos,
		Invocations:   1,
		TotalDuration: raw.Duration,
	}

	groups := make(map[string][]*ProfileRecord)
	for _, child := range raw.Children {
		key := child.Key()
		groups[key] = append(groups[key], child)
	}

	agg.Children = make([]*AggregatedRecord, 0, len(groups))

	for _, group := range groups {
		first := group[0]

		merged := &AggregatedRecord{
			Label:       first.Label,
			Pos:         first.Pos,
			Invocations: len(group),
		}

		grandChildren := make([]*ProfileRecord, 0)
		for _, rec := range group {
			merged.TotalDuration += rec.Duration
			grandChildren = append(grandChildren, rec.Children...)
		}

		if len(grandChildren) > 0 {
			tempParent := &ProfileRecord{Children: grandChildren}
			aggregatedGrandChildren := p.aggregate(tempParent)
			merged.Children = aggregatedGrandChildren.Children
		}

		agg.Children = append(agg.Children, merged)
	}
	return agg
}

func (p *Impl) buildReportString(sb *strings.Builder, node *AggregatedRecord, indent int, parentDuration time.Duration) {
	prefix := strings.Repeat("  ", indent)
	branch := "└─"
	if indent == 0 {
		branch = ""
	}

	var percent float64
	if parentDuration > 0 {
		percent = (float64(node.TotalDuration) / float64(parentDuration)) * 100.0
	}

	invcStr := ""
	if node.Invocations > 1 {
		invcStr = fmt.Sprintf(" (x%d)", node.Invocations)
	}

	line := fmt.Sprintf("%s%s[%s]%s taken %v (%.2f%%) at %s\n",
		prefix,
		branch,
		node.Label,
		invcStr,
		node.TotalDuration.Round(time.Microsecond),
		percent,
		node.Pos,
	)
	sb.WriteString(line)

	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].TotalDuration > node.Children[j].TotalDuration
	})

	for _, child := range node.Children {
		p.buildReportString(sb, child, indent+1, node.TotalDuration)
	}
}
