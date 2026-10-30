package devseed

import (
	"encoding/base64"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
)

func TestHashPasswordSesuaiParameter(t *testing.T) {
	hash, err := HashPassword("password-lokal-panjang")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	// Format: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<kunci>
	bagian := strings.Split(hash, "$")
	if len(bagian) != 6 || bagian[1] != "argon2id" || bagian[2] != "v=19" {
		t.Fatalf("format hash tidak sesuai: %q", hash)
	}
	if bagian[3] != "m=65536,t=3,p=2" {
		t.Fatalf("parameter argon2id tidak sesuai: %q", bagian[3])
	}
	salt, err := base64.RawStdEncoding.DecodeString(bagian[4])
	if err != nil || len(salt) != 16 {
		t.Fatalf("salt tidak valid: len=%d err=%v", len(salt), err)
	}
	kunci, err := base64.RawStdEncoding.DecodeString(bagian[5])
	if err != nil || len(kunci) != 32 {
		t.Fatalf("kunci tidak valid: len=%d err=%v", len(kunci), err)
	}
	ulang := argon2.IDKey([]byte("password-lokal-panjang"), salt, 3, 64*1024, 2, 32)
	if string(ulang) != string(kunci) {
		t.Fatal("hash tidak dapat diverifikasi ulang dengan parameter yang sama")
	}
}

func TestHashPasswordSaltBerbeda(t *testing.T) {
	a, err := HashPassword("password-sama")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	b, err := HashPassword("password-sama")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if a == b {
		t.Fatal("dua hash dengan password sama tidak boleh identik (salt harus acak)")
	}
}
