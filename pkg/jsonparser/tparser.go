// Package json provides functions to get json nodes from a json path.
// Also contains functions that returns a specific json node type.
// JSON path example: 'foo.bar'
package jsonparser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gojson "github.com/goccy/go-json"
	"gopkg.in/yaml.v3"
)

// TJson provides json parser toolkit.
type TJson struct {
	Buf  []byte
	node jsonNode
}

// Open file that contains a json or yaml object.
// Automatically detects file extension (.json, .yaml, .yml) and converts to JSON.
func Open(name string) (*TJson, error) {
	ext := strings.ToLower(filepath.Ext(name))

	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	if ext == ".yaml" || ext == ".yml" {
		data, err = yamlToJson(data)
		if err != nil {
			return nil, err
		}
	}

	return OpenFromBytes(data)
}

func OpenFromBytes(data []byte) (*TJson, error) {
	var node jsonNode
	err := gojson.Unmarshal(data, &node)
	if err != nil {
		return nil, err
	}
	sjson := TJson{
		Buf:  data,
		node: node,
	}
	return &sjson, nil
}

// AsMap returns the root object (may be nil for empty config).
func (g *TJson) AsMap() map[string]any {
	if g == nil || g.node == nil {
		return map[string]any{}
	}
	return map[string]any(g.node)
}

// ReplaceMap replaces the root object and refreshes Buf as JSON.
func (g *TJson) ReplaceMap(node map[string]any) error {
	if g == nil {
		return fmt.Errorf("nil TJson")
	}
	if node == nil {
		node = map[string]any{}
	}
	buf, err := gojson.Marshal(node)
	if err != nil {
		return err
	}
	g.node = jsonNode(node)
	g.Buf = buf
	return nil
}

// OpenYAMLBytes parses YAML into a TJson (does not read a file).
func OpenYAMLBytes(data []byte) (*TJson, error) {
	jsonData, err := yamlToJson(data)
	if err != nil {
		return nil, err
	}
	return OpenFromBytes(jsonData)
}

// Get json node from a json path.
// JSON path example: 'foo.bar'
func (g TJson) Get(key string) any {
	return getJSONPathValue(key, g.node)
}

// GetString get json string node from a json path.
// JSON path example: 'foo.bar'
func (g TJson) GetString(key string) string {
	v := getJSONPathValue(key, g.node)
	str, ok := v.(string)
	if !ok {
		return ""
	}
	return str
}

// GetInt get json integer 32 bit node from a json path.
// JSON path example: 'foo.bar'
func (g TJson) GetInt(key string) int {
	v := getJSONPathValue(key, g.node)
	value, valid := v.(float64)
	if !valid {
		return 0
	}
	return int(value)
}

// GetInt64 get json integer 64 node from a json path.
// JSON path example: 'foo.bar'
func (g TJson) GetInt64(key string) int64 {
	v := getJSONPathValue(key, g.node)
	value, valid := v.(float64)
	if !valid {
		return 0
	}
	return int64(value)
}

// GetFloat64 get json float 64 node from a json path.
// JSON path example: 'foo.bar'
func (g TJson) GetFloat64(key string) float64 {
	v := getJSONPathValue(key, g.node)
	valueFloat64, ok := v.(float64)
	if !ok {
		return 0.0
	}
	return valueFloat64
}

// GetBool get json bool node from a json path.
// JSON path example: 'foo.bar'
func (g TJson) GetBool(key string) bool {
	v := getJSONPathValue(key, g.node)
	valueBool, ok := v.(bool)
	if !ok {
		return false
	}
	return valueBool
}

// GetArrayMaps get json array as map from a json path
// JSON path example: 'foo.bar'
// Returns json array as a map.
func (g TJson) GetArrayMaps(key string) []map[string]any {
	v := getJSONPathValue(key, g.node)
	value, ok := v.([]any)
	if !ok {
		return nil
	}
	var a []map[string]any
	for _, m := range value {
		nm, ok := m.(map[string]any)
		if !ok {
			return nil
		}
		a = append(a, nm)
	}
	return a
}

// GetArrayMaps gets a array json node from a json key. Returns
// a json string array as []string.
func (g TJson) GetArrayStrings(key string) []string {
	v := getJSONPathValue(key, g.node)
	value, ok := v.([]any)
	if !ok {
		return nil
	}
	var a []string
	for _, m := range value {
		nm, ok := m.(string)
		if !ok {
			return nil
		}
		a = append(a, nm)
	}
	return a
}

func yamlToJson(data []byte) ([]byte, error) {
	var yamlDoc any
	if err := yaml.Unmarshal(data, &yamlDoc); err != nil {
		return nil, err
	}
	return gojson.Marshal(yamlDoc)
}
