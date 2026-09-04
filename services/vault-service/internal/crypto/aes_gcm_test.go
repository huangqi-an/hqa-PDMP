package crypto

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	// 这里放一个 32 字节密钥的 base64 表示
	keyBase64 := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	encryptor, err := New(keyBase64)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	plaintext := "sk-test-1234567890"
	encoded, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decoded, err := encryptor.Decrypt(encoded)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decoded != plaintext {
		t.Fatalf("roundtrip failed: got %q, want %q", decoded, plaintext)
	}
}
