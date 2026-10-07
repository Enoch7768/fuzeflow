package auth

import "testing"

func TestPasswordHash(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if hash == password {
		t.Fatal("password must not be stored as plaintext")
	}
	if !CheckPassword(hash, password) {
		t.Fatal("expected password to verify")
	}
	if CheckPassword(hash, "wrong password") {
		t.Fatal("wrong password verified")
	}
}
