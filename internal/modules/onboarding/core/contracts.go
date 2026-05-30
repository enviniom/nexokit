package core

type PasswordHasher interface {
	HashPassword(password string) (string, error)
}
