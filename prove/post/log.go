package post

import "C"
import (
	"encoding/hex"
	"github.com/trying2016/common-tools/logging"
)

/*
#include "post.h"
// forward declarations for callback C functions
void logCallback(ExternCRecord* record);
typedef void (*callback)(const struct ExternCRecord*);
*/
import "C"

const (
	Error = iota + 1
	Warn
	Info
	Debug
	Trace
)

var (
	levelMap = map[C.Level]uint32{
		C.Error: logging.ERROR,
		C.Warn:  logging.WARN,
		C.Info:  logging.INFO,
		C.Debug: logging.DEBUG,
		C.Trace: logging.TRACE,
	}
)

// SetLogCallback 设置日志回调
func SetLogCallback(level int) {
	C.set_logging_callback(C.Level(level), C.callback(C.logCallback))
}

//export logCallback
func logCallback(record *C.ExternCRecord) {
	msg := C.GoStringN(record.message.ptr, (C.int)(record.message.len))
	fields := logging.LogFormat{
		"module": C.GoStringN(record.module_path.ptr, (C.int)(record.module_path.len)),
		"file":   C.GoStringN(record.file.ptr, (C.int)(record.file.len)),
		"line":   int64(record.line),
	}
	logging.CPrint(levelMap[record.level], msg, fields)
}

//export randomXPow
func randomXPow(input *C.char, inputLen C.uintptr_t, difficulty *C.char, diffLen C.uintptr_t) C.uint64_t {
	sInput := C.GoStringN(input, C.int(inputLen))
	sDiff := C.GoStringN(difficulty, C.int(diffLen))
	//fmt.Println(sInput, sDiff)
	rawInput, _ := hex.DecodeString(sInput)
	rawDifficulty, _ := hex.DecodeString(sDiff)
	return C.uint64_t(randomxCallback(rawInput, rawDifficulty))
}

type RandomxCallback func(input, difficulty []byte) uint64

var randomxCallback RandomxCallback

func SetRandomxCallback(callback RandomxCallback) {
	randomxCallback = callback
}
