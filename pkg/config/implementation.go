package config

import (
	"os"
	"strings"
	"testing"

	"github.com/La-Perla-App/backend-core-laperla/pkg/jsonparser"
)

var tjson *jsonparser.TJson

// JSONConfigurator implement interface provider.
type JSONConfigurator struct{}

// GetString gets a string json node from a json key.
// JSON key example: 'foo.bar'
func (conf JSONConfigurator) GetString(key string) string {
	return tjson.GetString(key)
}

// GetInt64 gets an integer 64 json node from a json key.
// JSON key example: 'foo.bar'
func (conf JSONConfigurator) GetInt64(key string) int64 {
	return tjson.GetInt64(key)
}

// GetInt gets an integer json node from a json key.
// JSON key example: 'foo.bar'
func (conf JSONConfigurator) GetInt(key string) int {
	return tjson.GetInt(key)
}

// GetFloat64 gets a float 64 json node from a json key.
// JSON key example: 'foo.bar'
func (conf JSONConfigurator) GetFloat64(key string) float64 {
	return tjson.GetFloat64(key)
}

// GetBool gets a bool json node from a json key.
// JSON key example: 'foo.bar'
func (conf JSONConfigurator) GetBool(key string) bool {
	return tjson.GetBool(key)
}

// Get gets a json node from a json key.
// JSON key example: 'foo.bar'
func (conf JSONConfigurator) Get(key string) any {
	return tjson.Get(key)
}

// GetArrayMaps gets a array json node from a json key. Returns
// a json array as map.
// JSON key example: 'foo.bar'
func (conf JSONConfigurator) GetArrayMaps(key string) []map[string]any {
	return tjson.GetArrayMaps(key)
}

// GetArrayStrings gets a array json node from a json key. Returns
// a json string array as []string.
// JSON key example: 'foo.bar'
func (conf JSONConfigurator) GetArrayStrings(key string) []string {
	return tjson.GetArrayStrings(key)
}

// Open file that contins a json object.
func (conf JSONConfigurator) Open(name string) error {
	var path string
	if !testing.Testing() {
		path = name
	}
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--config-file=") {
			elements := strings.Split(arg, "=")
			path = elements[1]
		}
	}
	if path != "" {
		var err error
		tjson, err = jsonparser.Open(path)
		if err != nil {
			return err
		}
	} else {
		var err error
		tjson, err = jsonparser.OpenFromBytes([]byte("{}"))
		if err != nil {
			return err
		}
	}
	return nil
}
