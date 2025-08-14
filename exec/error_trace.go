package exec

import (
	"errors"
	"strings"
)

func buildErrorStacktrace(sb *strings.Builder, root *RuntimeError, depth int) {
	if root == nil {
		return
	}

	indent := strings.Repeat(" ", depth)
	sb.WriteString(indent)
	sb.WriteString("└─Caused at ")
	sb.WriteString(root.source.String())
	sb.WriteRune('\n')

	node := root
	source := root.source

	for node != nil {
		if node.source != source {
			break
		}

		sb.WriteString(indent)
		sb.WriteString(" ↑ ")
		sb.WriteString(node.Error())
		sb.WriteRune('\n')

		cause := node.cause
		if cause != nil {
			var re RuntimeError

			if errors.As(cause, &re) {
				node = &re
				continue
			} else {
				sb.WriteString(indent)
				sb.WriteString(" ↑ ")
				sb.WriteString(cause.Error())
				sb.WriteRune('\n')
			}
		}

		return
	}

	buildErrorStacktrace(sb, node, depth+1)
}
