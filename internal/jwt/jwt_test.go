package jwt

import (
	"errors"
	"testing"
	"time"
)

func TestGenerateParseRoundTrip(t *testing.T) {
	m := NewManager("secret", time.Hour, "issuer")

	token, err := m.Generate(42)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	claims, err := m.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("UserID = %d, want 42", claims.UserID)
	}
	if claims.Subject != "42" {
		t.Errorf("Subject = %q, want 42", claims.Subject)
	}
	if claims.Issuer != "issuer" {
		t.Errorf("Issuer = %q, want issuer", claims.Issuer)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	signer := NewManager("secret-a", time.Hour, "issuer")
	verifier := NewManager("secret-b", time.Hour, "issuer")

	token, err := signer.Generate(1)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := verifier.Parse(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Parse error = %v, want ErrInvalidToken", err)
	}
}

func TestParseRejectsWrongIssuer(t *testing.T) {
	signer := NewManager("secret", time.Hour, "issuer-a")
	verifier := NewManager("secret", time.Hour, "issuer-b")

	token, err := signer.Generate(1)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := verifier.Parse(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Parse error = %v, want ErrInvalidToken", err)
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	m := NewManager("secret", time.Minute, "issuer")

	token, err := m.Generate(1)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	m.now = func() time.Time { return time.Now().Add(2 * time.Minute) }

	if _, err := m.Parse(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Parse error = %v, want ErrInvalidToken", err)
	}
}

func TestParseRejectsTamperedToken(t *testing.T) {
	m := NewManager("secret", time.Hour, "issuer")

	token, err := m.Generate(1)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	tampered := token[:len(token)-1] + "x"
	if tampered == token {
		tampered = token[:len(token)-1] + "y"
	}
	if _, err := m.Parse(tampered); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("Parse error = %v, want ErrInvalidToken", err)
	}
}
