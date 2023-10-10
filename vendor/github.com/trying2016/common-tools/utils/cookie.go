package utils

import (
	"strings"

	"github.com/xxtea/xxtea-go/xxtea"
)

type Cookie struct {
	cookieMap Map
}

func ParseCookie(cookies string) *Cookie {
	cookie := &Cookie{
		cookieMap: Map{},
	}
	arr := strings.Split(cookies, ";")
	for _, info := range arr {
		ck := strings.Split(info, "=")
		if len(ck) == 2 {
			cookie.cookieMap[ck[0]] = ck[1]
		}
	}
	return cookie
}

func (ck *Cookie) GetString(key string) string {
	return ck.cookieMap.GetString(key)
}

func (ck *Cookie) Decode(key, cookieKey string) string {
	value := ck.cookieMap.GetString(key)
	value, _ = xxtea.DecryptString(value, cookieKey)
	return value
}
