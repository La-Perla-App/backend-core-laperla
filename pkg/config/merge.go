package config

import (
	"fmt"

	"dario.cat/mergo"
	"github.com/La-Perla-App/backend-core-laperla/pkg/jsonparser"
)

func mergeInto(base, overlay map[string]any) error {
	if overlay == nil {
		return nil
	}
	if err := mergo.Merge(&base, overlay, mergo.WithOverride); err != nil {
		return fmt.Errorf("merge config map: %w", err)
	}
	return nil
}

// MergeMap deep-merges overlay into the process-wide config singleton (overlay wins).
// Prefer ViewFromMap / ViewFromYAML + WithView when you need a request-scoped view.
func MergeMap(overlay map[string]any) error {
	if err := ValidateConfiguratorRegister(); err != nil {
		return err
	}
	if overlay == nil {
		return nil
	}
	if tjson == nil {
		var err error
		tjson, err = jsonparser.OpenFromBytes([]byte("{}"))
		if err != nil {
			return err
		}
	}
	base := tjson.AsMap()
	if base == nil {
		base = map[string]any{}
	}
	if err := mergeInto(base, overlay); err != nil {
		return err
	}
	return tjson.ReplaceMap(base)
}

// MergeYAML parses YAML and deep-merges it into the process-wide config singleton.
// Prefer ViewFromYAML + WithView when you need a request-scoped view.
func MergeYAML(yamlBytes []byte) error {
	if len(yamlBytes) == 0 {
		return nil
	}
	overlayDoc, err := jsonparser.OpenYAMLBytes(yamlBytes)
	if err != nil {
		return fmt.Errorf("parse config yaml: %w", err)
	}
	return MergeMap(overlayDoc.AsMap())
}
