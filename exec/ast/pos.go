package ast

import (
	"strconv"
)

type Pos struct {
	Index int // the index in terms of runes (0-indexed)
	Line  int // the ordinal in terms of line (1-indexed)
	str   string
}

func (p Pos) String() string {
	if p.str == "" {
		p.str = strconv.Itoa(p.Line) + ":" + strconv.Itoa(p.Index)
	}
	return p.str
}
