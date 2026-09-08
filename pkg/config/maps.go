package config

import "encoding/json"

// AsStringAnyMap normalizes nested config objects (incl. jsonparser.jsonNode aliases)
// into map[string]any so type assertions succeed across the defined-type boundary.
func AsStringAnyMap(raw any) map[string]any {
	if raw == nil {
		return nil
	}
	if m, ok := raw.(map[string]any); ok {
		return m
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil
	}
	return out
}

// AsStringSlice normalizes YAML/JSON list nodes into []string.
func AsStringSlice(v any) []string {
	switch list := v.(type) {
	case []string:
		return list
	case []any:
		out := make([]string, 0, len(list))
		for _, item := range list {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

// GetMap returns a nested object for key as map[string]any.
func GetMap(key string) map[string]any {
	return AsStringAnyMap(Get(key))
}
