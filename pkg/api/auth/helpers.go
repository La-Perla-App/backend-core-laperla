package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/La-Perla-App/backend-core-laperla/pkg/security"
)

const SessionCookieName = "laperla-access-token"

func getTokenFromCookie(header http.Header) (string, error) {
	var token string
	if cookieHeader := header.Get("Cookie"); cookieHeader != "" {
		cookies, err := http.ParseCookie(cookieHeader)
		if err != nil {
			return token, err
		}
		for _, cookie := range cookies {
			if cookie.Name == SessionCookieName {
				if !cookie.Expires.IsZero() && time.Now().After(cookie.Expires) {
					return token, errors.New("expired cookie")
				}
				token = cookie.Value
			}
		}
	}
	return token, nil
}

func getTokenFromAuthorizationHeader(header http.Header) (string, error) {
	authHeader := header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is missing")
	}

	// Check if the header uses the Bearer scheme
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	// Extract the token part
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}

func GetTokenFromHeader(header http.Header) (string, error) {
	cookieToken, cookieErr := getTokenFromCookie(header)

	token, tokenErr := getTokenFromAuthorizationHeader(header)
	if token != "" {
		return token, nil
	}
	if cookieToken == "" {
		return token, tokenErr
	}

	return cookieToken, cookieErr
}

func GenerateSessionToken(sessionData SessionData) (string, error) {
	data := NewSessionData(sessionData.UserID)

	if sessionData.SessionID == "" {
		sessionData.SessionID = data.SessionID
	}

	if sessionData.ExpirationTime.IsZero() {
		sessionData.ExpirationTime = data.ExpirationTime
	}

	return security.GenerateResourceToken(
		sessionData,
		security.WithJTI(sessionData.SessionID),
		security.WithExpirationTime(sessionData.ExpirationTime),
	)
}

func ValidateSessionToken(tokenString string) (*SessionData, error) {
	sessionData, claims, err := security.DecryptToken[SessionData](tokenString)
	if err != nil {
		return nil, err
	}

	if expTime, _ := claims.GetExpirationTime(); expTime != nil {
		sessionData.ExpirationTime = time.Unix(expTime.Unix(), 0)
	}

	if !sessionData.IsValid() {
		return nil, fmt.Errorf("invalid token payload")
	}

	return &sessionData, nil
}

func CookieFromSessionToken(sessionToken string) *http.Cookie {
	secure, _ := strconv.ParseBool(os.Getenv("SECURE_COOKIES"))
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionToken,
		Domain:   os.Getenv("SERVER_DOMAIN"),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteNoneMode,
		Path:     "/",
	}
	if sessionData, err := ValidateSessionToken(sessionToken); err == nil {
		cookie.MaxAge = sessionData.MaxAge()
	}
	return cookie
}
