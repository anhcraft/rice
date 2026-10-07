package exec

import (
	"strings"
)

func buildErrorStacktrace(re RuntimeError) string {
	frames := re.Frames()
	var sb strings.Builder
	sb.WriteString("RuntimeError: ")
	if len(frames) == 0 {
		sb.WriteString(re.Error())
		sb.WriteRune('\n')
		return sb.String()
	}

	inner := frames[len(frames)-1]
	sb.WriteString(inner.formatError())
	sb.WriteRune('\n')

	for i := 0; i < len(frames); {
		site := frames[i]
		sb.WriteString("└─ at ")
		sb.WriteString(site.siteString())
		sb.WriteRune('\n')
		j := i
		for j < len(frames) && frames[j].sameSite(site) {
			sb.WriteString("    while ")
			sb.WriteString(frames[j].Message)
			sb.WriteRune('\n')
			j++
		}
		i = j
	}
	return sb.String()
}
