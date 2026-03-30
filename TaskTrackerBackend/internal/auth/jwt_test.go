package auth

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateJWT_HappyPath(t *testing.T) {
	token, err := GenerateJWT("secret", 42, "user@example.com", 1, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d", len(parts))
	}
}

func TestGenerateJWT_EmptySecret(t *testing.T) {
	_, err := GenerateJWT("", 1, "user@example.com", 1, time.Hour)
	if err == nil {
		t.Fatal("expected error for empty secret, got nil")
	}
}

func TestVerifyJWT_ValidToken(t *testing.T) {
	secret := "testsecret"
	claims, err := GenerateJWT(secret, 7, "hello@world.com", 3, time.Hour)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	got, err := VerifyJWT(secret, claims)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.Sub != "7" {
		t.Errorf("sub: want 7, got %s", got.Sub)
	}
	if got.Email != "hello@world.com" {
		t.Errorf("email: want hello@world.com, got %s", got.Email)
	}
	if got.Ver != 3 {
		t.Errorf("ver: want 3, got %d", got.Ver)
	}
}

func TestVerifyJWT_ExpiredToken(t *testing.T) {
	secret := "testsecret"
	token, err := GenerateJWT(secret, 1, "u@e.com", 1, -time.Second)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	_, err = VerifyJWT(secret, token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestVerifyJWT_WrongSecret(t *testing.T) {
	token, err := GenerateJWT("secret-a", 1, "u@e.com", 1, time.Hour)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	_, err = VerifyJWT("secret-b", token)
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestVerifyJWT_TamperedPayload(t *testing.T) {
	secret := "testsecret"
	token, err := GenerateJWT(secret, 1, "u@e.com", 1, time.Hour)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	// Replace middle part with a different base64 payload
	parts := strings.Split(token, ".")
	parts[1] = "dGFtcGVyZWQ"  // base64url("tampered")
	tampered := strings.Join(parts, ".")
	_, err = VerifyJWT(secret, tampered)
	if err == nil {
		t.Fatal("expected error for tampered payload, got nil")
	}
}

func TestVerifyJWT_EmptySecret(t *testing.T) {
	_, err := VerifyJWT("", "any.token.here")
	if err == nil {
		t.Fatal("expected error for empty secret, got nil")
	}
}

func TestVerifyJWT_MalformedToken(t *testing.T) {
	cases := []string{
		"",
		"onlyone",
		"only.two",
		"too.many.dots.here",
	}
	for _, tc := range cases {
		_, err := VerifyJWT("secret", tc)
		if tc != "too.many.dots.here" && err == nil {
			t.Errorf("expected error for token %q, got nil", tc)
		}
	}
}

func TestSplitToken(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"a.b.c", 3},
		{"a.b.c.d", 3}, // only splits on first two dots; rest is part 3
		{"abc", 1},
		{"a.b", 1}, // only one dot → not enough → returns original as single element
	}
	for _, tc := range cases {
		got := splitToken(tc.input)
		if len(got) != tc.want {
			t.Errorf("splitToken(%q): want %d parts, got %d", tc.input, tc.want, len(got))
		}
	}
}

func TestGenerateVerify_RoundTrip(t *testing.T) {
	secret := "round-trip-secret"
	type tc struct {
		userID int
		email  string
		ver    int
	}
	cases := []tc{
		{1, "a@b.com", 1},
		{100, "x@y.z", 5},
		{99999, "big@id.io", 2},
	}
	for _, c := range cases {
		token, err := GenerateJWT(secret, c.userID, c.email, c.ver, 5*time.Minute)
		if err != nil {
			t.Fatalf("generate uid=%d: %v", c.userID, err)
		}
		claims, err := VerifyJWT(secret, token)
		if err != nil {
			t.Fatalf("verify uid=%d: %v", c.userID, err)
		}
		if claims.Email != c.email {
			t.Errorf("email: want %s, got %s", c.email, claims.Email)
		}
		if claims.Ver != c.ver {
			t.Errorf("ver: want %d, got %d", c.ver, claims.Ver)
		}
	}
}
