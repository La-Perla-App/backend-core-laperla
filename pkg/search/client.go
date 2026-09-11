package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

// Config mirrors cache/nats: addresses + basic auth from config.yaml / ConfigMap.
type Config struct {
	Addresses          []string
	Username           string
	Password           string
	InsecureSkipVerify bool
}

type clientImpl struct {
	api *opensearchapi.Client
}

// New creates an OpenSearch API client.
func New(cfg Config) (Searcher, error) {
	addrs := cfg.Addresses
	if len(addrs) == 0 {
		return nil, fmt.Errorf("opensearch: addrs vacío")
	}
	osCfg := opensearch.Config{
		Addresses:          addrs,
		Username:           cfg.Username,
		Password:           cfg.Password,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		MaxRetries:         3,
		RequestTimeout:     10 * time.Second,
	}
	api, err := opensearchapi.NewClient(opensearchapi.Config{Client: osCfg})
	if err != nil {
		return nil, fmt.Errorf("opensearch client: %w", err)
	}
	return &clientImpl{api: api}, nil
}

func (c *clientImpl) Ping(ctx context.Context) error {
	resp, err := c.api.Ping(ctx, nil)
	if err != nil {
		return err
	}
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if resp != nil && resp.IsError() {
		return fmt.Errorf("opensearch ping: status %s", resp.Status())
	}
	return nil
}

func (c *clientImpl) EnsureIndex(ctx context.Context, index string, mapping json.RawMessage) error {
	exists, err := c.api.Indices.Exists(ctx, opensearchapi.IndicesExistsReq{Indices: []string{index}})
	if err != nil {
		return fmt.Errorf("opensearch exists %s: %w", index, err)
	}
	if exists != nil && exists.Body != nil {
		defer exists.Body.Close()
	}
	if exists != nil && exists.StatusCode == http.StatusOK {
		return nil
	}
	var body io.Reader
	if len(mapping) > 0 {
		body = bytes.NewReader(mapping)
	}
	_, err = c.api.Indices.Create(ctx, opensearchapi.IndicesCreateReq{
		Index: index,
		Body:  body,
	})
	if err != nil {
		// race: another pod created it
		if strings.Contains(err.Error(), "resource_already_exists_exception") {
			return nil
		}
		return fmt.Errorf("opensearch create %s: %w", index, err)
	}
	return nil
}

func (c *clientImpl) Index(ctx context.Context, index, id string, doc any) error {
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	_, err = c.api.Index(ctx, opensearchapi.IndexReq{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(raw),
		Params:     opensearchapi.IndexParams{Refresh: "true"},
	})
	if err != nil {
		return fmt.Errorf("opensearch index %s/%s: %w", index, id, err)
	}
	return nil
}

func (c *clientImpl) Delete(ctx context.Context, index, id string) error {
	_, err := c.api.Document.Delete(ctx, opensearchapi.DocumentDeleteReq{
		Index:      index,
		DocumentID: id,
	})
	if err != nil {
		return fmt.Errorf("opensearch delete %s/%s: %w", index, id, err)
	}
	return nil
}

func (c *clientImpl) Search(ctx context.Context, index string, req Query) (*Result, error) {
	body, err := buildSearchBody(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.api.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{index},
		Body:    bytes.NewReader(body),
	})
	if err != nil {
		return nil, fmt.Errorf("opensearch search %s: %w", index, err)
	}
	out := &Result{Total: resp.Hits.Total.Value}
	for _, h := range resp.Hits.Hits {
		out.Hits = append(out.Hits, Hit{
			ID:     h.ID,
			Score:  float64(h.Score),
			Source: h.Source,
		})
	}
	return out, nil
}

func (c *clientImpl) Close() error {
	if c.api == nil {
		return nil
	}
	return c.api.Close()
}

func buildSearchBody(req Query) ([]byte, error) {
	must := make([]map[string]any, 0, 4)
	filter := make([]map[string]any, 0, 8)

	text := strings.TrimSpace(req.Text)
	if text != "" {
		fields := req.TextFields
		if len(fields) == 0 {
			fields = []string{"name^3", "name.es^3", "name.en^3", "slug^2", "tags", "description.es", "description.en", "address", "title"}
		}
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  text,
				"fields": fields,
				"type":   "best_fields",
				"operator": "and",
			},
		})
	} else {
		must = append(must, map[string]any{"match_all": map[string]any{}})
	}

	for k, v := range req.Terms {
		if v == nil {
			continue
		}
		if s, ok := v.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		filter = append(filter, map[string]any{
			"term": map[string]any{k: v},
		})
	}
	for k, vals := range req.TermsAny {
		if len(vals) == 0 {
			continue
		}
		filter = append(filter, map[string]any{
			"terms": map[string]any{k: vals},
		})
	}

	size := req.Size
	if size <= 0 {
		size = 20
	}
	from := req.From
	if from < 0 {
		from = 0
	}

	body := map[string]any{
		"from": from,
		"size": size,
		"query": map[string]any{
			"bool": map[string]any{
				"must":   must,
				"filter": filter,
			},
		},
		"track_total_hits": true,
	}
	if len(req.Sort) > 0 {
		sort := make([]any, 0, len(req.Sort))
		for _, s := range req.Sort {
			parts := strings.SplitN(s, ":", 2)
			field := parts[0]
			order := "asc"
			if len(parts) == 2 && parts[1] != "" {
				order = parts[1]
			}
			sort = append(sort, map[string]any{field: map[string]any{"order": order}})
		}
		body["sort"] = sort
	}

	return json.Marshal(body)
}
