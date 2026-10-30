// Package devseed menyediakan pembuatan hash password argon2id untuk akun
// development/UAT. Rumusnya dipakai bersama oleh cmd/seed-dev dan tes login
// lokal sehingga keduanya tidak dapat menyimpang.
package devseed

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// HashPassword membuat hash argon2id dengan parameter yang sepadan dengan
// OWASP Password Storage Cheat Sheet (m=64 MiB, t=3, p=2).
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	const mem uint32 = 64 * 1024
	const iter uint32 = 3
	const par uint8 = 2
	kunci := argon2.IDKey([]byte(password), salt, iter, mem, par, 32)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		mem, iter, par,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(kunci)), nil
}
