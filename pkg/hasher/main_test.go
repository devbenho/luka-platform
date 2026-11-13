package hasher

import (
	"strings"
	"testing"
)

func TestHashAndCompare(t *testing.T) {
	h := NewHasher()

	hashed, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}
	if hashed == "correct horse battery staple" {
		t.Fatal("hash must not equal the plaintext password")
	}

	if err := h.Compare(hashed, "correct horse battery staple"); err != nil {
		t.Errorf("Compare with the right password failed: %v", err)
	}
	if err := h.Compare(hashed, "wrong password"); err == nil {
		t.Error("Compare with a wrong password must fail")
	}
}

func TestHash_ReturnsErrorForPasswordOverBcryptLimit(t *testing.T) {
	// bcrypt rejects inputs longer than 72 bytes; Hash must surface that
	// instead of returning an empty hash with a nil error.
	hashed, err := NewHasher().Hash(strings.Repeat("a", 73))
	if err == nil {
		t.Fatal("expected an error for a password longer than 72 bytes")
	}
	if hashed != "" {
		t.Errorf("hash = %q, want empty on error", hashed)
	}
}
