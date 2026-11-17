package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUserJSONOmitsPasswordHash(t *testing.T) {
	u := User{ID: 1, Username: "alice", PasswordHash: "bcrypt-secret-hash"}

	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if strings.Contains(string(data), "bcrypt-secret-hash") {
		t.Errorf("JSON = %s, must not contain password hash", data)
	}
	if strings.Contains(string(data), "PasswordHash") {
		t.Errorf("JSON = %s, must not contain PasswordHash field", data)
	}
}
