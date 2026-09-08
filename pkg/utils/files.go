package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

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

// Error to indicate no unmarshaler was found for a given extension.
var errUnsupportedUnmarshalExtension = errors.New("unsupported extension for unmarshaling")

// unmarshalByExtension is a helper to centralize unmarshaling logic.
// It expects extWithDot to have a leading dot (e.g., ".json").
func unmarshalByExtension[T any](data []byte, extWithDot string, dst *T) error {
	switch extWithDot {
	case FileExtensions.Json, FileExtensions.JsonKey:
		return json.Unmarshal(data, dst)
	case FileExtensions.Yaml, FileExtensions.Yml, FileExtensions.YamlKey, FileExtensions.YmlKey:
		return yaml.Unmarshal(data, dst)
	case FileExtensions.Toml, FileExtensions.TomlKey:
		return toml.Unmarshal(data, dst)
	default:
		// If the extension is not supported by our unmarshalers
		return errUnsupportedUnmarshalExtension
	}
}

func LookupAndUnmarshalFile[T any](fsys fs.FS, basePath string, fileWithoutExt string, extensionsLookup []string, dst *T) (ext string, err error) {
	for _, targetExtLoopVar := range extensionsLookup {
		// Normalize extension from extensionsLookup to ensure it's just 'ext' without a leading dot for consistent filename construction.
		processedExtForFilename := strings.TrimPrefix(targetExtLoopVar, ".")
		filenameToTry := fmt.Sprintf("%s.%s", fileWithoutExt, processedExtForFilename)
		fullPath := filepath.Join(basePath, filenameToTry)

		file, openErr := fsys.Open(fullPath)
		if openErr != nil {
			// If file doesn't exist or cannot be opened, try the next extension.
			// This preserves the original behavior of continuing on fs.Open error.
			continue
		}
		// Ensure file is closed. Using explicit close here because it's in a loop.
		// Defer would only run when the function exits, potentially leaving many files open.

		data, readErr := io.ReadAll(file)
		// It's important to attempt closing the file even if ReadAll fails.
		closeErr := file.Close()

		if readErr != nil {
			return "", readErr // Return empty ext and the read error, as per original pattern
		}
		if closeErr != nil {
			// If read was successful but close failed, report close error.
			return "", closeErr
		}

		// For unmarshaling and return value, ensure the extension has a leading dot.
		// The targetExtLoopVar might be "json" or ".json".
		normalizedExtForResult := targetExtLoopVar
		if !strings.HasPrefix(normalizedExtForResult, ".") {
			normalizedExtForResult = "." + normalizedExtForResult
		}

		unmarshalErr := unmarshalByExtension(data, normalizedExtForResult, dst)
		if unmarshalErr == nil {
			// Successful unmarshal
			return normalizedExtForResult, nil
		}

		if !errors.Is(unmarshalErr, errUnsupportedUnmarshalExtension) {
			// It was a genuine unmarshaling error (e.g., malformed content)
			return "", unmarshalErr // Return empty ext and the unmarshal error
		}
		// If it was errUnsupportedUnmarshalExtension, it means this particular file extension
		// (e.g. a .txt file accidentally listed in extensionsLookup) isn't handled by our switch.
		// The loop should continue to try other extensions, preserving original behavior.
	}

	// No file found and successfully unmarshaled from the lookup list.
	return "", nil // Return empty ext and nil error
}

func UnmarshalFile[T any](fileName string, extensionsLookup []string /* unused */, dst *T) (ext string, err error) {
	file, openErr := os.Open(fileName)
	if openErr != nil {
		// Preserve original behavior: return "", nil if file can't be opened.
		return "", nil
	}
	defer file.Close() // Defer is fine here as it's not in a loop for this function's scope.

	data, readErr := io.ReadAll(file)
	if readErr != nil {
		return "", readErr // Return empty ext and the read error
	}

	// path.Ext includes the dot if an extension exists (e.g., ".json")
	// or an empty string if no dot/extension.
	actualExt := path.Ext(fileName)

	if actualExt == "" {
		// No extension found on the file, or it's a dotfile without further extension.
		// The original switch wouldn't match, and it would return "", nil.
		return "", nil
	}

	// The extensionsLookup parameter is intentionally not used, as in the original.

	unmarshalErr := unmarshalByExtension(data, actualExt, dst)
	if unmarshalErr == nil {
		// Successful unmarshal
		return actualExt, nil
	}

	if errors.Is(unmarshalErr, errUnsupportedUnmarshalExtension) {
		// The file's extension is not supported by our unmarshalers.
		// Original behavior: would not match any switch case and return "", nil.
		return "", nil
	}

	// It was a genuine unmarshaling error (e.g., malformed content)
	return "", unmarshalErr // Return empty ext and the unmarshal error
}

func CutPrefix(s, prefix string) (after string, found bool) {
	if !strings.HasPrefix(s, prefix) {
		return s, false
	}
	return s[len(prefix):], true
}
