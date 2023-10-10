package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"math/rand"
	"strings"
)

const (
	saltLength = 8
)

func generateSalt(len int) []byte {
	data := make([]byte, len)
	for i := 0; i < len; i++ {
		idx := rand.Intn(codeLen)
		data[i] = byte(codes[idx])
	}
	return data
}

func generateHash(salt, password string) string {
	key := []byte(salt)
	for i := 0; i < 1; i++ {
		mac := hmac.New(sha1.New, key)
		mac.Write([]byte(password))
		password = hex.EncodeToString(mac.Sum([]byte("")))
	}
	//Hmac(salt, password)
	return "sha1" + "$" + salt + "$" + "1" + "$" + password
}

func GeneratePassword(password string) string {
	salt := generateSalt(saltLength)
	return generateHash(string(salt), password)
}

func VerifyPassword(password, hashedPassword string) bool {
	arrInfos := strings.SplitN(hashedPassword, "$", -1)
	if len(arrInfos) != 4 {
		return false
	}
	return generateHash(arrInfos[1], password) == hashedPassword
}
