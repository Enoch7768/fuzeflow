package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const SessionTokenBytes = 32

func NewSessionToken() (string, []byte, error) {
	raw := make([]byte, SessionTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256(raw)
	return token, sum[:], nil
}

func HashSessionToken(token string) []byte {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil
	}
	sum := sha256.Sum256(raw)
	return sum[:]
}
