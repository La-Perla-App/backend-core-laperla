package {{.Package}}client

var (
    mu{{.ServiceName}} = {{ call .GetIdent "sync" "Mutex" }}{}
    clients{{.ServiceName}} = map[string]{{ call .GetConnectIdent (printf "%sClient" .ServiceName) }}{}
    h2cHttp{{.ServiceName}}Client = &{{ call .GetIdent "net/http" "Client" }}{
        CheckRedirect: func(_ *{{ call .GetIdent "net/http" "Request" }}, _ []*{{ call .GetIdent "net/http" "Request" }}) error {
            return {{ call .GetIdent "net/http" "ErrUseLastResponse" }}
        },
        Transport: &{{ call .GetIdent "golang.org/x/net/http2" "Transport"}} {
            AllowHTTP: true,
            DialTLS: func(network, addr string, _ *{{ call .GetIdent "crypto/tls" "Config"}}) ({{ call .GetIdent "net" "Conn" }}, error) {
                return {{ call .GetIdent "net" "Dial" }}(network, addr)
            },
        },
    }
)


func instanceNew{{.ServiceName}}Client(addr string) {{ call .GetConnectIdent (printf "%sClient" .ServiceName) }} {
	client := h2cHttp{{.ServiceName}}Client
	if {{ call $.GetIdent "strings" "HasPrefix" }}(addr, "https://") {
		client = &{{ call .GetIdent "net/http" "Client" }}{}
	}
	return {{ .ConnectNewClientIdent }}(
		client,
		addr,
		{{ call .GetIdent "connectrpc.com/connect" "WithHTTPGet"}}(),
	)
}

// Instance a new client for `{{.ProtoPackage}}.{{.ServiceName}}`
func Get{{.ServiceName}}Client(ctx {{ call .GetIdent "context" "Context"}}) {{ call .GetConnectIdent (printf "%sClient" .ServiceName) }} {
    addr := {{ .ServerEndpointIdent }}(ctx)
    var client {{ call .GetConnectIdent (printf "%sClient" .ServiceName) }}
	mu{{.ServiceName}}.Lock()
	defer mu{{.ServiceName}}.Unlock()
    if c, ok := clients{{.ServiceName}}[addr]; ok && c != nil {
        client = c
    } else {
        client = instanceNew{{.ServiceName}}Client(addr)
        clients{{.ServiceName}}[addr] = client
    }
    return client
}

{{ range .Methods }}
{{- if .Comments }}
{{ .Comments }}
{{- else }}
// Do a remote call for `{{$.ProtoPackage}}.{{$.ServiceName}}@{{.MethodName}}({{.RequestType}}) -> {{.ResponseType}}`
{{- end }}
{{- if .HasGateway }}
// Requires api.GeneralParams (token, locale, platform, destination).
{{- end }}
{{- if .Deprecated }}
//
// Deprecated: this service method is deprecated.
{{- end }}
func {{ if .PrefixWithServiceName }}{{$.ServiceName}}{{ end }}{{.MethodName}}(ctx {{ call $.GetIdent "context" "Context" }}, {{ if .HasGateway }}generalParams {{ $.GeneralParamsIdent }}, {{ end }}req *{{.RequestType}}) (*{{.ResponseType}}, error) {
	jsonReq, _ := {{ call $.GetIdent "google.golang.org/protobuf/encoding/protojson" "Marshal" }}(req)
	{{ call $.GetIdent "log" "Println" }}("PROCESSING UNARY GRPC METHOD: {{$.ProtoPackage}}.{{$.ServiceName}}@{{.MethodName}}({{.RequestType}}) -> {{.ResponseType}}")
	{{ call $.GetIdent "log" "Printf" }}("UNARY GRPC REQUEST: {{.RequestType}} -> %s\n", string(jsonReq))
    var response *{{.ResponseType}}
    {{- if .HasGateway }}
    rpcRequest, err := {{$.NewRequestIdent}}(generalParams, req)
    if err != nil {
        return response, err
    }
    {{- else }}
	rpcRequest := {{ call $.GetIdent "connectrpc.com/connect" "NewRequest" }}(req)
    {{- end }}
    rpcResponse, err := Get{{$.ServiceName}}Client(ctx).{{.MethodName}}(ctx, rpcRequest)
    if rpcResponse != nil {
        response = rpcResponse.Msg
		jsonRes, _ := {{ call $.GetIdent "google.golang.org/protobuf/encoding/protojson" "Marshal" }}(response)
		{{ call $.GetIdent "log" "Printf" }}("UNARY GRPC RESPONSE: {{.ResponseType}} -> %s\n", string(jsonRes))
    }
    return response, err
}
{{ end }}
