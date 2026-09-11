package search

import (
	"strings"
	"testing"
)

func TestBuildSearchBody_TermsAndText(t *testing.T) {
	raw, err := buildSearchBody(Query{
		Text: "café",
		Terms: map[string]any{
			"destination_id": "d1",
			"status":         "published",
			"is_featured":    true,
		},
		From: 0,
		Size: 10,
		Sort: []string{"name.keyword:asc"},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, want := range []string{
		`"destination_id"`,
		`"status"`,
		`"is_featured"`,
		`"multi_match"`,
		"café",
		`"name.keyword"`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("body missing %s: %s", want, s)
		}
	}
}
