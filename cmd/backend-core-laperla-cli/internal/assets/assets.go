package assets

import (
	"embed"
	"io/fs"
)

//go:embed all:_template
var templateFS embed.FS

var TemplateFS fs.FS

func init() {
	f, err := fs.Sub(templateFS, "_template")
	if err != nil {
		panic(err)
	}
	TemplateFS = f
}

//go:embed all:_handler
var handlerFS embed.FS

var HandlerFS fs.FS

func init() {
	f, err := fs.Sub(handlerFS, "_handler")
	if err != nil {
		panic(err)
	}
	HandlerFS = f
}

//go:embed all:_proto
var protoFS embed.FS

var ProtoFS fs.FS

func init() {
	f, err := fs.Sub(protoFS, "_proto")
	if err != nil {
		panic(err)
	}
	ProtoFS = f
}
