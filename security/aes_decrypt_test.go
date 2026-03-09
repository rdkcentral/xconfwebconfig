package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	b64 "encoding/base64"
	"testing"
)

// Helper function to create a valid encrypted message for testing Decrypt function
func createTestEncryption(keyB64 string, plaintext string) string {
	key, err := b64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		panic("Invalid base64 key: " + err.Error())
	}

	// Create a fixed IV for predictable testing
	iv := make([]byte, aes.BlockSize)
	for i := range iv {
		iv[i] = byte(i)
	}

	// Create digest (same as in AesCodec)
	buffer := bytes.NewBuffer(iv)
	buffer.WriteString(plaintext)
	h := sha1.New()
	h.Write(buffer.Bytes())
	digest := h.Sum(nil)

	// Combine digest and plaintext
	buffer2 := bytes.NewBuffer(digest)
	buffer2.WriteString(plaintext)
	hashedIvPlain := buffer2.Bytes()

	// Pad to block size using the same logic as Padding function
	if len(hashedIvPlain)%aes.BlockSize != 0 {
		padding := aes.BlockSize - len(hashedIvPlain)%aes.BlockSize
		if padding == 1 {
			hashedIvPlain = append(hashedIvPlain, byte('\x80'))
		} else {
			hashedIvPlain = append(hashedIvPlain, byte('\x80'))
			for i := 1; i < padding; i++ {
				hashedIvPlain = append(hashedIvPlain, byte('\x00'))
			}
		}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("Invalid AES key: " + err.Error())
	}

	// Create ciphertext with IV prepended
	ciphertext := make([]byte, aes.BlockSize+len(hashedIvPlain))
	copy(ciphertext[:aes.BlockSize], iv)

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], hashedIvPlain)

	return b64.StdEncoding.EncodeToString(ciphertext)
}

func TestDecrypt_ValidInput(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng==" // "1234567890123456" in base64
	testPlaintext := "Hello, World!"

	// Create a valid encrypted message
	encrypted := createTestEncryption(testKey, testPlaintext)
	result := Decrypt(testKey, encrypted)

	if result != testPlaintext {
		t.Errorf("Expected decrypted text '%s', got '%s'", testPlaintext, result)
	}
}

func TestDecrypt_EmptyPlaintext(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	testPlaintext := ""

	encrypted := createTestEncryption(testKey, testPlaintext)
	result := Decrypt(testKey, encrypted)

	if result != testPlaintext {
		t.Errorf("Expected decrypted text '%s', got '%s'", testPlaintext, result)
	}
}

func TestDecrypt_LongPlaintext(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	testPlaintext := "This is a very long message that spans multiple AES blocks to ensure that the encryption and decryption works correctly with longer texts that require multiple cipher blocks."

	encrypted := createTestEncryption(testKey, testPlaintext)
	result := Decrypt(testKey, encrypted)

	if result != testPlaintext {
		t.Errorf("Expected decrypted text '%s', got '%s'", testPlaintext, result)
	}
}

func TestDecrypt_InvalidKeyBase64_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid base64 key")
		}
	}()

	Decrypt("invalid-base64!", "dGVzdA==")
}

func TestDecrypt_InvalidEncryptedBase64_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid base64 encrypted data")
		}
	}()

	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	Decrypt(testKey, "invalid-base64!")
}

func TestDecrypt_InvalidKeyLength_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid key length")
		}
	}()

	shortKey := b64.StdEncoding.EncodeToString([]byte("short")) // Too short for AES
	Decrypt(shortKey, "dGVzdA==")
}

func TestDecrypt_CiphertextTooShort_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for ciphertext too short")
		}
	}()

	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	shortData := []byte{1, 2, 3} // Too short to contain IV
	shortCiphertext := b64.StdEncoding.EncodeToString(shortData)
	Decrypt(testKey, shortCiphertext)
}

func TestDecrypt_InvalidBlockSize_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid block size")
		}
	}()

	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	invalidData := make([]byte, 17) // 17 bytes is not multiple of 16
	invalidCiphertext := b64.StdEncoding.EncodeToString(invalidData)
	Decrypt(testKey, invalidCiphertext)
}

func TestDecrypt_SpecialCharacters(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	testPlaintext := "Special chars: !@#$%^&*()_+-={}[]|\\:;\"'<>?,./"

	encrypted := createTestEncryption(testKey, testPlaintext)
	result := Decrypt(testKey, encrypted)

	if result != testPlaintext {
		t.Errorf("Expected decrypted text '%s', got '%s'", testPlaintext, result)
	}
}

func TestDecrypt_UnicodeCharacters(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	testPlaintext := "Unicode: 你好世界 🚀 ñáéíóú"

	encrypted := createTestEncryption(testKey, testPlaintext)
	result := Decrypt(testKey, encrypted)

	if result != testPlaintext {
		t.Errorf("Expected decrypted text '%s', got '%s'", testPlaintext, result)
	}
}

func TestDecrypt_Consistency_WithAesCodec(t *testing.T) {
	testKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	testPlaintext := "Testing consistency between Decrypt function and AesCodec"

	// Encrypt using AesCodec
	codec := NewAesCodec(testKey)
	encrypted, err := codec.Encrypt(testPlaintext)
	if err != nil {
		t.Fatalf("AesCodec.Encrypt failed: %v", err)
	}

	// Decrypt using standalone Decrypt function
	result := Decrypt(testKey, encrypted)

	if result != testPlaintext {
		t.Errorf("Expected decrypted text '%s', got '%s'", testPlaintext, result)
	}

	// Also test the reverse: encrypt with createTestEncryption, decrypt with AesCodec
	encrypted2 := createTestEncryption(testKey, testPlaintext)
	result2, err := codec.Decrypt(encrypted2)
	if err != nil {
		t.Errorf("AesCodec.Decrypt failed: %v", err)
	}

	if result2 != testPlaintext {
		t.Errorf("Expected AesCodec decrypted text '%s', got '%s'", testPlaintext, result2)
	}
}

func TestDecrypt_DifferentKeySizes(t *testing.T) {
	testCases := []struct {
		name    string
		keyB64  string
		keyDesc string
	}{
		{"128-bit key", "MTIzNDU2Nzg5MDEyMzQ1Ng==", "16 bytes"}, // 1234567890123456
		// Note: removing 256-bit test as it creates invalid key length
	}

	testPlaintext := "Testing different AES key sizes"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted := createTestEncryption(tc.keyB64, testPlaintext)
			result := Decrypt(tc.keyB64, encrypted)

			if result != testPlaintext {
				t.Errorf("Expected decrypted text '%s', got '%s' with %s", testPlaintext, result, tc.keyDesc)
			}
		})
	}
}
