package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"math/rand"
)

var (
	codes      = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	allCodes   = "abcdefghijklmnopqrstuvwxyz1234567890"
	codeLen    = len(codes)
	allCodeLen = len(allCodes)
)

//生成32位随机序列
func CreateRandomString(len int) string {
	data := make([]byte, len)

	for i := 0; i < len; i++ {
		idx := rand.Intn(codeLen)
		data[i] = byte(codes[idx])
	}
	return string(data)
}

func CreateRandomAllString(len int) string {
	data := make([]byte, len)

	for i := 0; i < len; i++ {
		idx := rand.Int() % allCodeLen
		data[i] = allCodes[idx]
	}
	return string(data)
}

func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func SHA256(str string) string {
	return SHA256Byte([]byte(str))
}

func SHA256Byte(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

func Hmac(key, data string) string {
	hmac := hmac.New(md5.New, []byte(key))
	hmac.Write([]byte(data))
	return hex.EncodeToString(hmac.Sum([]byte("")))
}

func Sha1(data string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(data))
	return hex.EncodeToString(sha1.Sum([]byte("")))
}

func Base64Encoding(src []byte) string {
	maxLen := base64.StdEncoding.EncodedLen(len(src))
	dst := make([]byte, maxLen)
	base64.StdEncoding.Encode(dst, src)
	return string(dst)
}

func Base64Decoding(src string) []byte {
	dst, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return nil
	}
	return dst
}
