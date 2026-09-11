package search

import (
	"context"
	"encoding/json"
)

// Searcher is the OpenSearch surface shared by product services.
// Services must not import opensearch-go directly — extend this package instead.
type Searcher interface {
	Ping(ctx context.Context) error
	EnsureIndex(ctx context.Context, index string, mapping json.RawMessage) error
	Index(ctx context.Context, index, id string, doc any) error
	Delete(ctx context.Context, index, id string) error
	Search(ctx context.Context, index string, req Query) (*Result, error)
	Close() error
}

// Query describes a free-text + filter search.
type Query struct {
	// Text is multi_match over TextFields (or default name/slug/tags).
	Text string
	// TextFields overrides default text fields when Text is set.
	TextFields []string
	// Term exact matches (keyword / boolean / numeric).
	Terms map[string]any
	// TermsAny matches documents where field is in values (terms query).
	TermsAny map[string][]any
	From     int
	Size     int
	// Sort: e.g. "name.keyword:asc", "sort_order:asc", "_score:desc".
	Sort []string
}

// Hit is one search hit; Source is the indexed JSON document.
type Hit struct {
	ID     string
	Score  float64
	Source json.RawMessage
}

// Result is a page of hits.
type Result struct {
	Total int
	Hits  []Hit
}

// Index names used by La Perla (directory).
const (
	IndexDestinations = "laperla-destinations"
	IndexCategories   = "laperla-categories"
	IndexBusinesses   = "laperla-businesses"
	IndexFeatured     = "laperla-featured"
)
