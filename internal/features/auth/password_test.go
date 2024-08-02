package auth

import "testing"

const (
	correctPassword   = "12345"
	incorrectPassword = "54321"
)

func TestHashPassword(t *testing.T) {
	hashed, err := HashPassword(correctPassword)

	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}

	if hashed == "" {
		t.Fatalf("Expected hashed password, got empty string")
	}

	if hashed == correctPassword {
		t.Fatalf("Expected hashed password to be different from password")
	}

}

func TestComparePassword(t *testing.T) {
	hashed, err := HashPassword(correctPassword)

	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}

	err = ComparePassword(hashed, correctPassword)
	if err != nil {
		t.Fatalf("Error comparing password: %v", err)
	}

	err = ComparePassword(hashed, incorrectPassword)
	if err == nil {
		t.Fatalf("Expected password to not match password: %v", err)
	}
}
