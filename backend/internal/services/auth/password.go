package auth

import "golang.org/x/crypto/bcrypt"

// PasswordVerifier проверяет соответствие
// открытого пароля его bcrypt-хешу.
type PasswordVerifier interface {
	Compare(
		hash string,
		password string,
	) error
}

// BcryptPasswordVerifier реализует PasswordVerifier
// с использованием bcrypt.
type BcryptPasswordVerifier struct{}

func (BcryptPasswordVerifier) Compare(
	hash string,
	password string,
) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}
