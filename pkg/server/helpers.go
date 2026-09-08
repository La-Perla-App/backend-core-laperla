package server

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"
	"github.com/La-Perla-App/backend-core-laperla/pkg/server/interceptors"
	"go.akshayshah.org/connectproto"
	"google.golang.org/protobuf/encoding/protojson"
)

var defaultHandlerOptions = []connect.HandlerOption{
	interceptors.DefaultInterceptors,
	connectproto.WithJSON(
		protojson.MarshalOptions{EmitDefaultValues: true},
		protojson.UnmarshalOptions{DiscardUnknown: true},
	),
}

func ServiceHandlerOptions(opts ...connect.HandlerOption) []connect.HandlerOption {
	return append(opts, defaultHandlerOptions...)
}

func createProtoZipFile(fsys fs.FS, filesBasepath string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	if err := fs.WalkDir(fsys, filesBasepath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".proto" && !d.IsDir() {
			f, err := zw.Create(strings.TrimPrefix(strings.TrimPrefix(path, filesBasepath), "/"))
			if err != nil {
				return err
			}
			data, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}
			_, err = f.Write([]byte(data))
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
