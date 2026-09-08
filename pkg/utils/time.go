package utils

import (
	"slices"
	"strings"
	"time"
)

func FormatTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000000-07:00")
}

var timeLayouts = slices.Compact([]string{
	"2006-01-02T15:04:05.000000-07:00",
	"02/01/2006 15:04:05",
	"02/01/2006 3:04:05 PM",
	"2006-01-02T15:04:05.000",
	"2006-01-02T15:04:05.000000",
	"2006-01-02T15:04:05.999999Z07:00",
	time.RFC3339Nano,
	time.Layout,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
	time.DateTime,
	time.DateOnly,
	time.TimeOnly,
})

func ParseTime(t string) (parsed time.Time, err error) {
	t = strings.ReplaceAll(t, "p. m.", "PM")
	t = strings.ReplaceAll(t, "a. m.", "AM")
	for _, layout := range timeLayouts {
		parsed, err = time.Parse(layout, t)
		if err == nil {
			return
		}
	}
	return parsed, err
}

func NormalizeTimeString(t string) string {
	formatted, err := ParseTime(t)
	if err != nil {
		return t
	}
	return FormatTime(formatted)
}
