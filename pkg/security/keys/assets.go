package keys

import _ "embed"

//go:embed rsa.key
var PrivateKey []byte

//go:embed rsa.key.pub
var PublicKey []byte
