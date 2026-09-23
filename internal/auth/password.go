package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"log"
)

func Hash(password string, iterations, keyLen, saltLen int) (hash, salt []byte, err error) {
	salt = make([]byte, saltLen)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, nil, err
	}

	hash, err = pbkdf2.Key(sha256.New, password, salt, iterations, keyLen)

	return hash, salt, err
}

func Verify(password string, salt, want []byte, iterations, keylen int) bool {
	hash, err := pbkdf2.Key(sha256.New, password, salt, iterations, keylen)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		return false
	}

	if subtle.ConstantTimeCompare(hash, want) == 1 {
		return true
	}

	return false
}
