package security

import (
	"encoding/base64"
	"testing"
)

func TestNewAesCodec(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng==" // "1234567890123456" in base64
	codec := NewAesCodec(testKey)

	if codec == nil {
		t.Fatal("NewAesCodec should not return nil")
	}

	if len(codec.key) != 16 {
		t.Errorf("Expected key length 16, got %d", len(codec.key))
	}
}

func TestNewAesCodec_InvalidBase64(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid base64 key")
		}
	}()

	NewAesCodec("invalid-base64!")
}

func TestAesCodec_EncryptDecrypt_RoundTrip(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng==" // "1234567890123456" in base64
	codec := NewAesCodec(testKey)

	testCases := []string{
		"Hello, World!",
		"This is a test message",
		"x", // Single character
		"A very long message that spans multiple blocks to test the padding functionality properly",
		"Special chars: !@#$%^&*() +-={}[]|\\:;\"'<>?,./",
		"Unicode: 你好世界 🚀 ñáéíóú",
	}

	for _, plaintext := range testCases {
		t.Run("plaintext "+plaintext, func(t *testing.T) {
			encrypted, err := codec.Encrypt(plaintext)
			if err != nil {
				t.Errorf("Encrypt failed: %v", err)
			}
			if encrypted == "" {
				t.Error("Encrypt should not return empty string")
			}

			decrypted, err := codec.Decrypt(encrypted)
			if err != nil {
				t.Errorf("Decrypt failed: %v", err)
			}

			if decrypted != plaintext {
				t.Errorf("Round trip failed: expected '%s', got '%s'", plaintext, decrypted)
			}
		})
	}
}

func TestAesCodec_EncryptWithoutPadding_DecryptWithoutPadding_RoundTrip(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng==" // "1234567890123456" in base64
	codec := NewAesCodec(testKey)

	testCases := []string{
		"Hello, World!",
		"This is a test message",
		"", // Empty string
		"A longer message for testing",
		"Special chars: !@#$%^&*()",
	}

	for _, plaintext := range testCases {
		t.Run("no padding "+plaintext, func(t *testing.T) {
			encrypted, err := codec.EncryptWithoutPadding(plaintext)
			if err != nil {
				t.Errorf("EncryptWithoutPadding failed: %v", err)
			}
			if encrypted == "" && plaintext != "" {
				t.Error("EncryptWithoutPadding should not return empty string for non-empty input")
			}

			decrypted, err := codec.DecryptWithoutPadding(encrypted)
			if err != nil {
				t.Errorf("DecryptWithoutPadding failed: %v", err)
			}

			if decrypted != plaintext {
				t.Errorf("Round trip failed: expected '%s', got '%s'", plaintext, decrypted)
			}
		})
	}
}

func TestAesCodec_Decrypt_InvalidBase64(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	codec := NewAesCodec(testKey)

	_, err := codec.Decrypt("invalid-base64!")
	if err == nil {
		t.Error("Expected error for invalid base64 input")
	}
}

func TestAesCodec_Decrypt_TooShortCiphertext(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	codec := NewAesCodec(testKey)

	// Create a ciphertext that's too short (less than block size)
	shortData := []byte{1, 2, 3}
	shortCiphertext := base64.StdEncoding.EncodeToString(shortData)

	_, err := codec.Decrypt(shortCiphertext)
	if err == nil {
		t.Error("Expected error for too short ciphertext")
	}
}

func TestAesCodec_Decrypt_InvalidBlockSize(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	codec := NewAesCodec(testKey)

	// Create a ciphertext with invalid block size (not multiple of 16)
	invalidData := make([]byte, 17) // 17 bytes is not multiple of 16
	invalidCiphertext := base64.StdEncoding.EncodeToString(invalidData)

	_, err := codec.Decrypt(invalidCiphertext)
	if err == nil {
		t.Error("Expected error for invalid block size")
	}
}

func TestAesCodec_DecryptWithoutPadding_InvalidBase64(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	codec := NewAesCodec(testKey)

	_, err := codec.DecryptWithoutPadding("invalid-base64!")
	if err == nil {
		t.Error("Expected error for invalid base64 input")
	}
}

func TestAesCodec_DecryptWithoutPadding_TooShort(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	codec := NewAesCodec(testKey)

	// Create a ciphertext that's definitely too short - empty
	shortCiphertext := base64.StdEncoding.EncodeToString([]byte{})

	result, err := codec.DecryptWithoutPadding(shortCiphertext)
	// For GCM mode, empty ciphertext should either cause an error or return empty result
	if err != nil {
		// This is expected behavior - error on empty/invalid input
		t.Logf("Got expected error for empty GCM ciphertext: %v", err)
	} else if result == "" {
		// This is also acceptable - empty input returns empty output
		t.Log("Empty ciphertext returned empty result")
	} else {
		t.Errorf("Unexpected result for empty ciphertext: '%s'", result)
	}
}

func TestDigest(t *testing.T) {
	iv := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	plaintext := "test message"

	digest := Digest(iv, plaintext)

	if len(digest) != 20 { // SHA1 produces 20 bytes
		t.Errorf("Expected digest length 20, got %d", len(digest))
	}

	// Test deterministic behavior
	digest2 := Digest(iv, plaintext)
	if string(digest) != string(digest2) {
		t.Error("Digest should be deterministic")
	}

	// Test different inputs produce different digests
	digest3 := Digest(iv, "different message")
	if string(digest) == string(digest3) {
		t.Error("Different inputs should produce different digests")
	}
}

func TestPadding_Basic(t *testing.T) {
	testCases := []struct {
		input       []byte
		expected    int // expected padding length
		description string
	}{
		{[]byte("hello"), 11, "5 bytes -> needs 11 more to reach 16"},
		{[]byte("1234567890123456"), 16, "exactly 16 bytes -> still adds full block padding"},
		{[]byte("x"), 15, "1 byte -> needs 15 more"},
		{[]byte(""), 16, "empty -> needs full block"},
		{[]byte("12345678901234567"), 15, "17 bytes -> needs 15 more to reach 32"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := Padding(tc.input, 16) // AES block size is 16
			actualPaddingLen := len(result) - len(tc.input)

			if actualPaddingLen != tc.expected {
				t.Errorf("Input %q (%d bytes): expected padding length %d, got %d",
					tc.input, len(tc.input), tc.expected, actualPaddingLen)
			}

			// Check that result length is multiple of 16
			if len(result)%16 != 0 {
				t.Errorf("Padded result length %d is not multiple of 16", len(result))
			}

			// Check padding content for non-zero padding cases
			if tc.expected > 0 {
				// Should start with 0x80 and be followed by 0x00
				paddingStart := len(tc.input)
				if result[paddingStart] != 0x80 {
					t.Errorf("Padding should start with 0x80, got 0x%02x", result[paddingStart])
				}
				// Rest should be 0x00
				for i := paddingStart + 1; i < len(result); i++ {
					if result[i] != 0x00 {
						t.Errorf("Padding byte at position %d should be 0x00, got 0x%02x", i, result[i])
					}
				}
			}
		})
	}
}

func TestAesCodec_Decrypt_DecryptError(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	codec := NewAesCodec(testKey)

	// Create some random data that won't decrypt properly
	randomData := make([]byte, 32) // 2 blocks
	for i := range randomData {
		randomData[i] = byte(i)
	}
	invalidCiphertext := base64.StdEncoding.EncodeToString(randomData)

	_, err := codec.Decrypt(invalidCiphertext)
	if err == nil {
		t.Error("Expected error for invalid ciphertext")
	}
}
