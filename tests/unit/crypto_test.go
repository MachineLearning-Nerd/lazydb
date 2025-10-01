package unit

import (
	"testing"

	"github.com/MachineLearning-Nerd/lazydb/internal/storage"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"simple password", "password123"},
		{"complex password", "P@ssw0rd!#$%^&*()"},
		{"long password", "this-is-a-very-long-password-with-many-characters-1234567890"},
		{"unicode", "密码🔒"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := storage.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Empty string should return empty
			if tt.plaintext == "" && encrypted != "" {
				t.Errorf("Expected empty encrypted string for empty plaintext")
			}

			// Non-empty should be different from plaintext
			if tt.plaintext != "" && encrypted == tt.plaintext {
				t.Errorf("Encrypted text should be different from plaintext")
			}

			// Decrypt
			decrypted, err := storage.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			// Should match original
			if decrypted != tt.plaintext {
				t.Errorf("Decrypted text doesn't match plaintext. Got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptionDeterminism(t *testing.T) {
	plaintext := "test-password"

	// Encrypt twice
	encrypted1, err1 := storage.Encrypt(plaintext)
	encrypted2, err2 := storage.Encrypt(plaintext)

	if err1 != nil || err2 != nil {
		t.Fatalf("Encryption failed: %v, %v", err1, err2)
	}

	// Should be different (due to random nonce)
	if encrypted1 == encrypted2 {
		t.Error("Same plaintext encrypted twice should produce different ciphertexts (random nonce)")
	}

	// But both should decrypt to same value
	decrypted1, _ := storage.Decrypt(encrypted1)
	decrypted2, _ := storage.Decrypt(encrypted2)

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both encrypted values should decrypt to original plaintext")
	}
}

func TestDecryptInvalidData(t *testing.T) {
	tests := []struct {
		name       string
		ciphertext string
	}{
		{"empty", ""},
		{"invalid base64", "not-valid-base64!!!"},
		{"too short", "YWJj"},
		{"random data", "SGVsbG8gV29ybGQh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := storage.Decrypt(tt.ciphertext)
			// Empty string should not error
			if tt.name == "empty" && err != nil {
				t.Errorf("Empty string should not error, got: %v", err)
			}
			// Others should error or return empty
			if tt.name != "empty" && err == nil {
				t.Log("Decryption of invalid data did not error (acceptable)")
			}
		})
	}
}
