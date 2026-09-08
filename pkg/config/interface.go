package config

// Configurator interface provider
type Configurator interface {
	// GetString gets a string json node from a json key.
	// JSON key example: 'foo.bar'
	GetString(key string) string

	// GetInt64 gets an integer 64 json node from a json key.
	// JSON key example: 'foo.bar'
	GetInt64(key string) int64

	// GetInt gets an integer json node from a json key.
	// JSON key example: 'foo.bar'
	GetInt(key string) int

	// GetFloat64 gets a float 64 json node from a json key.
	// JSON key example: 'foo.bar'
	GetFloat64(key string) float64

	// GetBool gets a bool json node from a json key.
	// JSON key example: 'foo.bar'
	GetBool(key string) bool

	// Get gets a json node from a json key.
	// JSON key example: 'foo.bar'
	Get(key string) any

	// GetArrayMaps gets a array json node from a json key. Returns
	// a json array as map.
	// JSON key example: 'foo.bar'
	GetArrayMaps(key string) []map[string]any

	// GetArrayStrings gets a array json node from a json key. Returns
	// a json string array as []string.
	// JSON key example: 'foo.bar'
	GetArrayStrings(key string) []string

	// Open file that contins a json object.
	Open(name string) error
}
