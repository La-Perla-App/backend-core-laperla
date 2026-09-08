package config

import (
	"fmt"

	gojson "github.com/goccy/go-json"
	"github.com/La-Perla-App/backend-core-laperla/pkg/jsonparser"
)

// View is an immutable request-scoped config snapshot (base ∘ overlay).
// It does not mutate the process-wide singleton loaded from config.yaml.
type View struct {
	doc *jsonparser.TJson
}

// Ensure View satisfies Configurator.
var _ Configurator = (*View)(nil)

func (v *View) GetString(key string) string {
	if v == nil || v.doc == nil {
		return ""
	}
	return v.doc.GetString(key)
}

func (v *View) GetInt64(key string) int64 {
	if v == nil || v.doc == nil {
		return 0
	}
	return v.doc.GetInt64(key)
}

func (v *View) GetInt(key string) int {
	if v == nil || v.doc == nil {
		return 0
	}
	return v.doc.GetInt(key)
}

func (v *View) GetFloat64(key string) float64 {
	if v == nil || v.doc == nil {
		return 0
	}
	return v.doc.GetFloat64(key)
}

func (v *View) GetBool(key string) bool {
	if v == nil || v.doc == nil {
		return false
	}
	return v.doc.GetBool(key)
}

func (v *View) Get(key string) any {
	if v == nil || v.doc == nil {
		return nil
	}
	return v.doc.Get(key)
}

func (v *View) GetArrayMaps(key string) []map[string]any {
	if v == nil || v.doc == nil {
		return nil
	}
	return v.doc.GetArrayMaps(key)
}

func (v *View) GetArrayStrings(key string) []string {
	if v == nil || v.doc == nil {
		return nil
	}
	return v.doc.GetArrayStrings(key)
}

// Open is unsupported on Views (immutable snapshots).
func (v *View) Open(name string) error {
	return fmt.Errorf("config.View is read-only")
}

// BaseMap returns a deep copy of the process config root (never mutates singleton).
func BaseMap() (map[string]any, error) {
	if err := ValidateConfiguratorRegister(); err != nil {
		return nil, err
	}
	if tjson == nil {
		return map[string]any{}, nil
	}
	return cloneMap(tjson.AsMap())
}

// ViewFromMap deep-merges overlay onto a copy of the process config (overlay wins).
func ViewFromMap(overlay map[string]any) (*View, error) {
	base, err := BaseMap()
	if err != nil {
		return nil, err
	}
	if overlay != nil {
		if err := mergeInto(base, overlay); err != nil {
			return nil, err
		}
	}
	doc, err := jsonparser.OpenFromBytes(mustJSON(base))
	if err != nil {
		return nil, err
	}
	return &View{doc: doc}, nil
}

// ViewFromYAML parses YAML and returns a View = process config ∘ overlay (no singleton mutation).
func ViewFromYAML(yamlBytes []byte) (*View, error) {
	if len(yamlBytes) == 0 {
		return ViewFromMap(nil)
	}
	overlayDoc, err := jsonparser.OpenYAMLBytes(yamlBytes)
	if err != nil {
		return nil, fmt.Errorf("parse config yaml: %w", err)
	}
	return ViewFromMap(overlayDoc.AsMap())
}

func cloneMap(m map[string]any) (map[string]any, error) {
	if m == nil {
		return map[string]any{}, nil
	}
	buf, err := gojson.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("clone config map: %w", err)
	}
	var out map[string]any
	if err := gojson.Unmarshal(buf, &out); err != nil {
		return nil, fmt.Errorf("clone config map: %w", err)
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func mustJSON(m map[string]any) []byte {
	buf, err := gojson.Marshal(m)
	if err != nil {
		return []byte("{}")
	}
	return buf
}
