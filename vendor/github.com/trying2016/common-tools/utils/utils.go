package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// 对账号等进行掩码
func MaskContent(str string) string {
	content := []rune(str)
	if len(content) < 2 {
		return str
	}
	reserveNum := 1
	if len(content)/2 > 2 {
		reserveNum = 2
	}
	const MaskLen = 3
	var contentLen = reserveNum*2 + MaskLen
	data := make([]rune, contentLen, contentLen)
	for i := 0; i < reserveNum; i++ {
		data[i] = content[i]
		data[contentLen-i-1] = content[len(content)-i-1]
	}
	for i := 0; i < MaskLen; i++ {
		data[i+reserveNum] = '*'
	}
	return string(data)
}

func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func Min64(a, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}

func ToShowBalance(balance int64) string {
	str := fmt.Sprintf("%v", balance)
	for i := len(str); i <= 8; i++ {
		str = fmt.Sprintf("0%s", str)
	}
	nSplit := len(str) - 8
	str = str[:nSplit] + "." + str[nSplit:]
	return str
}

func Md5Byte(genSin []byte, body []byte) string {
	arrByte := make([]byte, len(genSin)+len(body))
	copy(arrByte, genSin)
	copy(arrByte[len(genSin):], body)

	h := md5.New()
	h.Write(arrByte)
	return hex.EncodeToString(h.Sum(nil))
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	} else {
		return b
	}
}
func MinInt64(a, b int64) int64 {
	if a > b {
		return b
	} else {
		return a
	}
}

// 截取字符串
func SubleString(src, str1, str2 string) string {
	src = string([]rune(src))
	str1 = string([]rune(str1))
	str2 = string([]rune(str2))

	nBegine := strings.Index(src, str1)
	if nBegine == -1 || nBegine == len(src)-1 {
		return ""
	}
	tmp := src[nBegine+len(str1):]
	nEnd := strings.Index(tmp, str2)
	if nEnd == -1 {
		return ""
	}
	return tmp[:nEnd]
}

/*
 * 删除Slice中的元素。
 * params:
 *   s: slice对象指针，类型为*[]interface{}
 *   index: 要删除元素的索引
 * return:
 *   无
 * 说明：直接操作传入的Slice对象，传入的序列地址不变，但内容已经被修改
 */

func SliceRemove(s interface{}, index int) {
	arr := reflect.ValueOf(s)
	if arr.Kind() != reflect.Array {
		return
	}
	arr.Len()
	//var arr *[]interface{} = (*[]interface{})(unsafe.Pointer(&s))
	//(*arr) = append((*arr)[:index], (*arr)[index+1:]...)
	//*s = append((*s)[:index], (*s)[index+1:]...)
}

// PanicTrace trace panic stack info.
func PanicTrace(kb int) []byte {
	s := []byte("/src/runtime/panic.go")
	e := []byte("\ngoroutine ")
	line := []byte("\n")
	stack := make([]byte, kb<<10) //4KB
	length := runtime.Stack(stack, true)
	start := bytes.Index(stack, s)
	stack = stack[start:length]
	start = bytes.Index(stack, line) + 1
	stack = stack[start:]
	end := bytes.LastIndex(stack, line)
	if end != -1 {
		stack = stack[:end]
	}
	end = bytes.Index(stack, e)
	if end != -1 {
		stack = stack[:end]
	}
	stack = bytes.TrimRight(stack, "\n")
	return stack
}

// 安全的go run
func SafeGo(callBack func(), panicFn func(err interface{})) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				data := PanicTrace(4)
				if data != nil {
					fmt.Printf("panic : %v", string(data))
					if panicFn != nil {
						panicFn(string(data))
					}
				} else {
					fmt.Printf("panic : %v", err)
					if panicFn != nil {
						panicFn(err)
					}
				}
			}
		}()
		callBack()
	}()
}

// 最长执行时间，返回是否超时
func RunTimeout(fn func(), millisecond int64) bool {
	var job sync.WaitGroup
	chTimeout := make(chan struct{})
	job.Add(1)
	go func() {
		fn()
		job.Done()
	}()
	go func() {
		job.Wait()
		chTimeout <- struct{}{}
		close(chTimeout)
	}()

	select {
	case <-time.After(time.Millisecond * time.Duration(millisecond)):
		return true
	case <-chTimeout:
		return false
	}
}
