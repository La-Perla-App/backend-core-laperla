package auth

import (
	"encoding/json"
	"time"

	"github.com/La-Perla-App/backend-core-laperla/pkg/jsonparser"
	"github.com/google/uuid"
)

const (
	RolePoliciesWildcard = "*"
)

type SessionData struct {
	UserID         string              `json:"userId"`
	SessionID      string              `json:"sessionId"`
	ExpirationTime time.Time           `json:"-"`
	SessionPayload map[string]any      `json:"sessionPayload"`
	Policies       map[string]Policies `json:"policies"`
}

type Policies struct {
	Allow []string
	Block []string
}

func (p Policies) CheckAction(targetAction string) bool {
	var allowed bool
	for _, action := range p.Allow {
		if action == RolePoliciesWildcard || targetAction == action {
			allowed = true
			break
		}
	}
	for _, action := range p.Block {
		if action == RolePoliciesWildcard || targetAction == action {
			allowed = false
			break
		}
	}
	return allowed
}

func (sessionData SessionData) MaxAge() int {
	currentTime := time.Now().Unix()
	maxAge := sessionData.ExpirationTime.Unix() - currentTime
	if maxAge < 0 {
		return 0
	}
	return int(maxAge)
}

func NewSessionData(userID string) SessionData {
	return NewSessionDataWithExpiration(userID, time.Time{})
}

func NewSessionDataWithExpiration(userID string, expirationTime time.Time) SessionData {
	if expirationTime.IsZero() {
		expirationTime = time.Now().Add(time.Hour * 24)
	}
	return SessionData{
		UserID:         userID,
		ExpirationTime: expirationTime,
		SessionID:      uuid.New().String(),
	}
}

func (sessionData SessionData) IsValid() bool {
	if sessionData.SessionID == "" {
		return false
	}
	if sessionData.UserID == "" {
		return sessionData.Policies != nil
	}
	return sessionData.UserID != ""
}

func (sessionData *SessionData) SetSessionPayloadData(data any, key string) {
	if data == nil {
		delete(sessionData.SessionPayload, key)
		return
	}
	if sessionData.SessionPayload == nil {
		sessionData.SessionPayload = make(map[string]any)
	}
	sessionData.SessionPayload[key] = jsonparser.ConvertToJSON(data)
}

func GetSessionPayloadData[T any](session *SessionData, key string) (data T, ok bool) {
	if session == nil {
		return
	}
	val, found := session.SessionPayload[key]
	if !found {
		return
	}
	jsonStr, err := json.Marshal(val)
	if err != nil {
		return
	}
	if err := json.Unmarshal(jsonStr, &data); err != nil {
		return
	}
	ok = found
	return
}
