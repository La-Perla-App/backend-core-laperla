package hooks

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/gocql/gocql"
)

var isProd = os.Getenv("MODE") == "PROD"

type LogHook struct{}

func NewLogHook() *LogHook {
	return &LogHook{}
}

func (h *LogHook) ObserveQuery(ctx context.Context, query gocql.ObservedQuery) {
	if isProd || os.Getenv("CQL_NO_TRACE") != "" {
		return
	}

	//if query.End.Sub(query.Start) >= 2000*time.Millisecond {
	log.Printf(
		"CQL Trace -> Query: `%s`, Rows: %d, Args: `%s`, Attempt: %d, Latency: %s\n",
		trimQueryContent(query.Statement),
		query.Rows,
		argsToString(query.Values),
		query.Attempt,
		query.End.Sub(query.Start),
	)
	//}
}

func (h *LogHook) ObserveBatch(ctx context.Context, batch gocql.ObservedBatch) {
	if isProd || os.Getenv("CQL_NO_TRACE") != "" {
		return
	}

	//if batch.End.Sub(batch.Start) >= 2000*time.Millisecond {
	log.Printf(
		"CQL Trace -> Batch Latency: %s\n",
		batch.End.Sub(batch.Start),
	)
	//}

}

func trimQueryContent(query string) string {
	separator := " ••• skipped ••• "
	query = strings.TrimSpace(query)
	if os.Getenv("CQL_FULL_TRACE") != "" {
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
