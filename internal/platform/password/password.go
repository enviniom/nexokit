package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	errInvalidHash         = errors.New("invalid hash format")
	errIncompatibleVersion = errors.New("incompatible argon2 version")
	errMismatchedHash      = errors.New("password does not match")
)

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var defaultParams = params{
	memory:      65536,
	iterations:  3,
	parallelism: 4,
	saltLength:  16,
	keyLength:   32,
}

// Manager handles password hashing and verification.
type Manager struct{}

// HashPassword creates an argon2id hash of the given password.
func (Manager) HashPassword(password string) (string, error) {
	return Hash(password)
}

// VerifyPassword compares a password against an encoded argon2id hash.
func (Manager) VerifyPassword(password, encodedHash string) error {
	return Verify(password, encodedHash)
}

// Hash creates an argon2id hash of the given password.
func Hash(password string) (string, error) {
	salt := make([]byte, defaultParams.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		defaultParams.iterations,
		defaultParams.memory,
		defaultParams.parallelism,
		defaultParams.keyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		defaultParams.memory,
		defaultParams.iterations,
		defaultParams.parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// Verify compares a password against an encoded argon2id hash.
func Verify(password, encodedHash string) error {
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return err
	}

	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		p.iterations,
		p.memory,
		p.parallelism,
		p.keyLength,
	)

	if subtle.ConstantTimeCompare(hash, otherHash) != 1 {
		return errMismatchedHash
	}

	return nil
}

func decodeHash(encodedHash string) (*params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, errInvalidHash
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, errInvalidHash
	}
	if version != argon2.Version {
		return nil, nil, nil, errIncompatibleVersion
	}

	p := &params{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, errInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, errInvalidHash
	}
	p.saltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, errInvalidHash
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}
