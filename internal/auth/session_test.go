package auth

import "testing"

func TestSessionToken(t *testing.T) {
	token, hash, err := NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(token) < 40 || len(hash) != 32 {
		t.Fatal("unexpected session token material")
	}
	got := HashSessionToken(token)
	if string(got) != string(hash) {
		t.Fatal("session hash mismatch")
	}
}
