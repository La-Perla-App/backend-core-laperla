package internal

import (
	"encoding/json"
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

func LookupAndUnmarshalFile[T any](fs fs.FS, basePath string, fileWithoutExt string, extensionsLookup []string, dst *T) (ext string, err error) {
	for _, targetExt := range extensionsLookup {
		if file, err := fs.Open(filepath.Join(basePath, fmt.Sprintf("%s.%s", fileWithoutExt, strings.TrimPrefix(targetExt, ".")))); err == nil {
			data, err := io.ReadAll(file)
			if err != nil {
				return ext, err
			}

			targetExt, _ = CutPrefix(targetExt, ".")
			targetExt = fmt.Sprintf(".%s", targetExt)

			switch targetExt {
			case FileExtensions.Json, FileExtensions.JsonKey:
				if err := json.Unmarshal(data, dst); err != nil {
					return ext, err
				}
			case FileExtensions.Yaml, FileExtensions.Yml, FileExtensions.YamlKey, FileExtensions.YmlKey:
				if err := yaml.Unmarshal(data, dst); err != nil {
					return ext, err
				}
			case FileExtensions.Toml, FileExtensions.TomlKey:
				if err := toml.Unmarshal(data, dst); err != nil {
					return ext, err
				}
			}

			ext = targetExt
		}
	}
	return
}

func UnmarshalFile[T any](fileName string, extensionsLookup []string, dst *T) (ext string, err error) {
	if file, err := os.Open(fileName); err == nil {
		data, err := io.ReadAll(file)
		if err != nil {
			return ext, err
		}

		targetExt, _ := CutPrefix(path.Ext(fileName), ".")
		targetExt = fmt.Sprintf(".%s", targetExt)

		switch targetExt {
		case FileExtensions.Json, FileExtensions.JsonKey:
			if err := json.Unmarshal(data, dst); err != nil {
				return ext, err
			}
		case FileExtensions.Yaml, FileExtensions.Yml, FileExtensions.YamlKey, FileExtensions.YmlKey:
			if err := yaml.Unmarshal(data, dst); err != nil {
				return ext, err
			}
		case FileExtensions.Toml, FileExtensions.TomlKey:
			if err := toml.Unmarshal(data, dst); err != nil {
				return ext, err
			}
		}

		ext = targetExt
	}
	return
}

func CutPrefix(s, prefix string) (after string, found bool) {
	if !strings.HasPrefix(s, prefix) {
		return s, false
	}
	return s[len(prefix):], true
}
