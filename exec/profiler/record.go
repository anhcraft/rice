package profiler

import (
	"github.com/anhcraft/rice/exec/ast"
	"time"
)

type ProfileRecord struct {
	Label     string
	key       string
	Pos       ast.Pos
	Depth     int
	StartTime time.Time
	Duration  time.Duration
	Children  []*ProfileRecord
}

func (pr *ProfileRecord) Reset() {
	pr.Label = ""
	pr.key = ""
	pr.Pos = ast.Pos{}
	pr.Depth = 0
	pr.StartTime = time.Time{}
	pr.Duration = 0
	pr.Children = pr.Children[:0] // preserve capacity
}

func (pr *ProfileRecord) Key() string {
	if pr.key == "" {
		pr.key = pr.Label + "@" + pr.Pos.String()
	}
	return pr.key
}

func (pr *ProfileRecord) StopClock() {
	pr.Duration = time.Since(pr.StartTime)
}

func (pr *ProfileRecord) AddChild(child *ProfileRecord) {
	pr.Children = append(pr.Children, child)
}
