package middlewares

var HeadersKeys = struct {
	Insecure             string
	ShowGRPCErrorDetails string
	EncryptedPrefix      string
	CustomHandler        string
	HasErrors            string
}{
	Insecure:             "X-Insecure",
	ShowGRPCErrorDetails: "X-Show-GRPC-Errors-Details",
	EncryptedPrefix:      "X-Encrypted-Prefix",
	CustomHandler:        "Custom-Handler",
	HasErrors:            "Has-Errors",
}
