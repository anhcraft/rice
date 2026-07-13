package datetime

import (
	"fmt"

	"github.com/anhcraft/rice/exec/fun"
	"github.com/anhcraft/rice/exec/stdlib"
	"github.com/anhcraft/rice/exec/types"
	"github.com/anhcraft/rice/exec/types/values"
	"time"
)

var Functions = fun.FunctionPackage{
	"now":    {stdlib.Define(Now)},
	"parse":  {stdlib.Define(Parse)},
	"format": {stdlib.Define(Format)},
}

// Now returns the current Unix timestamp in milliseconds.
func Now() (types.Value, error) {
	return values.Int(time.Now().UnixMilli()), nil
}

// Parse parses an ISO 8601 / RFC 3339 date-time string and returns the Unix
// timestamp in milliseconds. Supported formats:
//   - RFC 3339:     "2024-01-15T10:30:00Z", "2024-01-15T10:30:00+07:00"
//   - RFC 3339 Nano: "2024-01-15T10:30:00.123456789Z"
//   - Date only:     "2024-01-15" (treated as midnight UTC)
func Parse(s values.String) (types.Value, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02",
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, string(s))
		if err == nil {
			return values.Int(t.UnixMilli()), nil
		}
	}
	return nil, fmt.Errorf("datetime.parse: cannot parse %q as ISO 8601 date", string(s))
}

// Format formats a Unix millisecond timestamp into a human-readable string. All
// output is in UTC. Supported format names:
//
//	"rfc3339"  → "2024-01-15T10:30:00Z"
//	"date"     → "2024-01-15"
//	"time"     → "10:30:00"
//	"datetime" → "2024-01-15 10:30:00"
func Format(ts values.Int, format values.String) (types.Value, error) {
	t := time.UnixMilli(int64(ts)).UTC()
	switch string(format) {
	case "rfc3339":
		return values.String(t.Format(time.RFC3339)), nil
	case "date":
		return values.String(t.Format("2006-01-02")), nil
	case "time":
		return values.String(t.Format("15:04:05")), nil
	case "datetime":
		return values.String(t.Format("2006-01-02 15:04:05")), nil
	default:
		return nil, fmt.Errorf("datetime.format: unknown format %q, use rfc3339, date, time, or datetime", string(format))
	}
}
