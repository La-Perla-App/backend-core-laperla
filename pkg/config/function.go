package config

// GetString gets a string json node from a json key.
// JSON key example: 'foo.bar'
func GetString(key string) string {
	err := ValidateConfiguratorRegister()
	if err != nil {
		return ""
	}
	return configurator.GetString(key)
}

// GetInt64 gets an integer 64 json node from a json key.
// JSON key example: 'foo.bar'
func GetInt64(key string) int64 {
	err := ValidateConfiguratorRegister()
	if err != nil {
		return 0
	}
	return configurator.GetInt64(key)
}

// GetInt gets an integer json node from a json key.
// JSON key example: 'foo.bar'
func GetInt(key string) int {
	err := ValidateConfiguratorRegister()
	if err != nil {
		return 0
	}
	return configurator.GetInt(key)
}

// GetFloat64 gets a float 64 json node from a json key.
// JSON key example: 'foo.bar'
func GetFloat64(key string) float64 {
	err := ValidateConfiguratorRegister()
	if err != nil {
		return 0
	}
	return configurator.GetFloat64(key)
}

// GetBool gets a bool json node from a json key.
// JSON key example: 'foo.bar'
func GetBool(key string) bool {
	err := ValidateConfiguratorRegister()
	if err != nil {
		return false
	}
	return configurator.GetBool(key)
}

// Get gets a json node from a json key.
// JSON key example: 'foo.bar'
func Get(key string) any {
	err := ValidateConfiguratorRegister()
	if err != nil {
		return nil
	}
	return configurator.Get(key)
}

// GetArrayMaps gets a array json node from a json key. Returns
// a json array as map.
// JSON key example: 'foo.bar'
func GetArrayMaps(key string) []map[string]any {
	err := ValidateConfiguratorRegister()
	if err != nil {
		return nil
	}
	return configurator.GetArrayMaps(key)
}

// GetArrayStrings gets a array json node from a json key. Returns
// a json string array as []string.
// JSON key example: 'foo.bar'
func GetArrayStrings(key string) []string {
	err := ValidateConfiguratorRegister()
	if err != nil {
		return nil
	}
	return configurator.GetArrayStrings(key)
}

// LoadConfigFile Open file that contins a json object.
func LoadConfigFile(name string) error {
	err := ValidateConfiguratorRegister()
	if err != nil {
		return err
	}
	return configurator.Open(name)
}
