package security

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"time"

	"github.com/La-Perla-App/backend-core-laperla/pkg/config"
	"github.com/La-Perla-App/backend-core-laperla/pkg/jsonparser"
	"github.com/La-Perla-App/backend-core-laperla/pkg/security/keys"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type tokenResourceOptions struct {
	JTI            any
	ExpirationTime time.Time
}

type TokenResourceOption func(options *tokenResourceOptions)

func WithExpirationTime(expirationTime time.Time) func(options *tokenResourceOptions) {
	return func(options *tokenResourceOptions) {
		options.ExpirationTime = expirationTime
	}
}

func WithJTI(jti any) func(options *tokenResourceOptions) {
	return func(options *tokenResourceOptions) {
		options.JTI = jti
	}
}

func GenerateResourceToken(data any, options ...TokenResourceOption) (string, error) {
	k := config.GetString("security.rsa.privateKey")
	if k == "" {
		k = string(keys.PrivateKey)
	}
	privateKey, err := loadPrivateKey([]byte(k))
	if err != nil {
		return "", fmt.Errorf("error loading private key: %v", err)
	}

	var opts tokenResourceOptions
	for _, opt := range options {
		opt(&opts)
	}

	// Configurar claims (payload)
	claims := jwt.MapClaims{
		"jti":     uuid.New().String(),
		"payload": jsonparser.ConvertToJSON(data),
		"iss":     "laperla",
	}

	if opts.JTI != nil {
		claims["jti"] = opts.JTI
	}

	if !opts.ExpirationTime.IsZero() {
		claims["exp"] = opts.ExpirationTime.Unix()
	}

	// Crear token con algoritmo HS256
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Firmar el token con la clave secreta
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("error al firmar el token: %v", err)
	}

	return signedToken, nil
}

func DecryptToken[T any](tokenString string) (T, jwt.MapClaims, error) {
	var data T
	k := config.GetString("security.rsa.publicKey")
	if k == "" {
		k = string(keys.PublicKey)
	}
	publicKey, err := loadPublicKey([]byte(k))
	if err != nil {
		return data, nil, fmt.Errorf("error loading public key: %v", err)
	}

	// Parsear el token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// Verificar el algoritmo de firma
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("invalid token alg: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return data, nil, fmt.Errorf("invalid token: %v", err)
	}

	if !token.Valid {
		return data, nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return data, nil, fmt.Errorf("invalid token payload")
	}

	if jsonData, ok := claims["payload"]; ok {
		jsonStr, err := json.Marshal(jsonData)
		if err != nil {
			return data, nil, err
		}

		err = json.Unmarshal(jsonStr, &data)
		if err != nil {
			return data, nil, err
		}
	}

	return data, claims, nil
}

func loadPrivateKey(privateKeyData []byte) (*rsa.PrivateKey, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func loadPublicKey(publicKeyData []byte) (*rsa.PublicKey, error) {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}
