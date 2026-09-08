package internal

var FileExtensions = struct {
	Json    string
	JsonKey string
	Toml    string
	TomlKey string
	Yaml    string
	YamlKey string
	Yml     string
	YmlKey  string
}{
	Json:    ".json",
	JsonKey: ".json.key",
	Toml:    ".toml",
	TomlKey: ".toml.key",
	Yaml:    ".yaml",
	YamlKey: ".yaml.key",
	Yml:     ".yml",
	YmlKey:  ".yml.key",
}

const ProtocPluginBin = "github.com/La-Perla-App/backend-core-laperla/cmd/protoc-gen-backend-core-laperla-cli"
