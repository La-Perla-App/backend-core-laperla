package internal

import (
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func ParseGoMod(basepath string) (string, error) {
	file := path.Join(basepath, "go.mod")
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	modFile, err := modfile.Parse(file, data, nil)
	if err != nil {
		return "", err
	}

	return modFile.Module.Mod.Path, nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func CreateFilesFromFS(f fs.FS, targetDir string, force bool, replacer *strings.Replacer) error {
	// Walk through the embedded file system
	return fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, path)
		if d.IsDir() {
			if _, err := os.Stat(targetPath); os.IsNotExist(err) {
				if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
					return err
				}
			}
		} else {
			if _, err := os.Stat(targetPath); os.IsNotExist(err) || force {
				srcFile, err := f.Open(path)
				if err != nil {
					return err
				}
				defer srcFile.Close()

				dstFile, err := os.Create(targetPath)
				if err != nil {
					return err
				}
				defer dstFile.Close()

				srcContent, err := io.ReadAll(srcFile)
				if err != nil {
					return err
				}

				content := string(srcContent)
				if replacer != nil {
					content = replacer.Replace(content)
				}

				if filepath.Ext(path) == ".go" {
					formattedContent, err := format.Source([]byte(content))
					if err != nil {
						return fmt.Errorf("failed to format Go source: %w", err)
					}
					content = string(formattedContent)
				}

				if _, err := dstFile.Write([]byte(content)); err != nil {
					return err
				}
			} else {
				fmt.Printf("[WARN] File \"%s\" already exists, skipping. Use --force to overwrite.\n", targetPath)
			}
		}
		return nil
	})
}
