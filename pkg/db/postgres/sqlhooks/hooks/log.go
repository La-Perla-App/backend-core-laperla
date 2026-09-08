package hooks

import (
	"context"
	"log"
	"os"
	"strings"
	"time"
)

var isProd = os.Getenv("MODE") == "PROD"

var started int

type LogHook struct{}

func NewLogHook() *LogHook {
	return &LogHook{}
}

func (h *LogHook) Before(ctx context.Context, query string, args ...any) (context.Context, error) {
	return context.WithValue(ctx, &started, time.Now()), nil
}

func (h *LogHook) After(ctx context.Context, query string, args ...any) (context.Context, error) {
	if isProd || ctx.Value("silent") != nil || os.Getenv("SQL_NO_TRACE") != "" {
		return ctx, nil
	}
	log.Printf(
		"SQL Trace -> Query: `%s`, Args: `%s`. took: %s\n",
		trimQueryContent(query),
		argsToString(args),
		time.Since(ctx.Value(&started).(time.Time)),
	)
	return ctx, nil
}

func (h *LogHook) OnError(ctx context.Context, err error, query string, args ...any) error {
	if isProd || ctx.Value("silent") != nil || os.Getenv("SQL_NO_TRACE") != "" {
		return err
	}
	log.Printf(
		"SQL Trace -> Error: %v, Query: `%s`, Args: `%s`, Took: %s\n",
		err,
		trimQueryContent(query),
		argsToString(args),
		time.Since(ctx.Value(&started).(time.Time)),
	)
	return err
}

func trimQueryContent(query string) string {
	separator := " ••• skipped ••• "
	query = strings.TrimSpace(query)
	if os.Getenv("SQL_FULL_TRACE") != "" {
		return query
	}
	var segments []string
	for _, segment := range strings.Split(query, "\n") {
		segment = strings.TrimSpace(segment)
		if segment != "" {
			segments = append(segments, segment)
		}
	}
	if len(segments) > 1 {
		query = segments[0] + " " + segments[1]
		if len(segments) > 2 {
			query += separator + segments[len(segments)-1]
		}
	}
	if len(query) > 500 {
		query = strings.ReplaceAll(query, separator, "")
		query = string(query[:250]) + separator + string(query[len(query)-250:])
	}
	return query
}
