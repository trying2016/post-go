//go:build post_gpu_cuda_static
// +build post_gpu_cuda_static

package post_go

/*
#cgo LDFLAGS: -L./ -lpost_prove_cuda -Wl,-rpath,./
#include <stdlib.h>
#include "capi.h"
void logGpuCallback(int level, char* message);
*/
import "C"

import (
	"time"
	"unsafe"
)

const (
	LogLevelInfo  = C.LOG_INFO
	LogLevelError = C.LOG_ERROR
)

type LogCallback func(level int, message string)

var callback LogCallback

//export logGpuCallback
func logGpuCallback(level C.int, message *C.char) {
	callback(int(level), C.GoString(message))
	//fmt.Printf("Log (Level %d): %s\n", int(level), C.GoString(message))
}

type PostGPU *C.post_gpu

type Result struct {
	Index uint64
	Nonce uint32
}

// 通过dylib加载动态库，实现capi.h中的方法
func InitLibrary(filename string) error {
	//lib := dylib.NewLazyDLL(filename)
	//if err := lib.Load(); err != nil {
	//	return err
	//}
	//_setLogCallback = lib.NewProc("set_log_callback")
	//_postCreate = lib.NewProc("post_create")
	//_postDestroy = lib.NewProc("post_destroy")
	//_postProve = lib.NewProc("post_prove")
	//_postGetResult = lib.NewProc("post_get_results")
	//_postDeviceCount = lib.NewProc("post_device_count")
	//_postDeviceName = lib.NewProc("post_device_name")
	return nil
}

// SetLogCallback 设置日志回调函数
func SetLogCallback(fn LogCallback) {
	callback = fn
	C.set_log_callback((C.log_callback)(C.logGpuCallback))
	//_setLogCallback.Call(uintptr(unsafe.Pointer(C.logGpuCallback)))
	//_setLogCallback.Call(uintptr(callback))
}

// PostCreate 创建Post对象
/*
post_gpu* post_create(int device,
                    int start,
                    int nonces,
                    uint8_t *ciphers_keys,
                    uint8_t *lazy_ciphers_keys,
                    uint64_t difficulty_lsb,
                    uint8_t difficulty_msb,
                    int input_size,
                    const char *sources,
                    int source_size)
*/
func PostCreate(device, start, nonces int, ciphersKeys, lazyCiphersKeys []byte, difficultyLsb uint64, difficultyMsb uint8, inputSize int, source []byte) PostGPU {
	cCiphersKeys := C.CBytes(ciphersKeys)
	defer C.free(cCiphersKeys)
	cLazyCiphersKeys := C.CBytes(lazyCiphersKeys)
	defer C.free(cLazyCiphersKeys)
	cSource := C.CBytes(source)
	defer C.free(cSource)

	result := C.post_create(C.int(device), C.int(start), C.int(nonces), (*C.uint8_t)(cCiphersKeys), (*C.uint8_t)(cLazyCiphersKeys), C.uint64_t(difficultyLsb), C.uint8_t(difficultyMsb), C.int(inputSize), (*C.char)(cSource), C.int(len(source)))
	return PostGPU(result)

	//result, _, _ := _postCreate.Call(uintptr(device),
	//	uintptr(start),
	//	uintptr(nonces),
	//	uintptr(cCiphersKeys),
	//	uintptr(cLazyCiphersKeys),
	//	uintptr(difficultyLsb),
	//	uintptr(difficultyMsb),
	//	uintptr(inputSize),
	//	uintptr(cSource),
	//	uintptr(len(source)))

	//return PostGPU(unsafe.Pointer(result))
}

// PostDestroy 释放post_cuda对象
func PostDestroy(ctx PostGPU) {
	C.post_destroy((*C.post_gpu)(ctx))
	//_postDestroy.Call(uintptr(unsafe.Pointer(ctx)))
}

type CResult C.Result

// PostProve 生成证明
/*
int post_prove(post_cuda* ctx, uint64_t base_index, uint8_t* data, int data_size, Result* out);
*/
func PostProve(ctx PostGPU, baseIndex uint64, data []byte) ([]Result, error) {
	cData := C.CBytes(data)
	defer C.free(cData)

	tick := time.Now().UnixMilli()
	//count, _, _ := _postProve.Call(uintptr(unsafe.Pointer(ctx)), uintptr(C.uint64_t(baseIndex)), uintptr(cData), uintptr(C.int(len(data))))

	count := int(C.post_prove((*C.post_gpu)(ctx), C.uint64_t(baseIndex), (*C.uint8_t)(cData), C.int(len(data))))

	if count != 0 {
		println("count:", count, "consume:", time.Now().UnixMilli()-tick)
	}

	var list []Result
	// 遍历out,将结果放入list中
	for i := 0; i < int(count); i++ {
		var out *C.Result = &C.Result{}
		C.post_get_results((*C.post_gpu)(ctx), C.int(i), out)
		//_postGetResult.Call(uintptr(unsafe.Pointer(ctx)), uintptr(i), uintptr(unsafe.Pointer(out)))
		list = append(list, Result{
			Index: uint64(out.index),
			Nonce: uint32(out.nonce),
		})
	}

	return list, nil
}

// PostDeviceCount 获取设备数量
func PostDeviceCount() int {
	count := C.post_device_count()
	//count, _, _ := _postDeviceCount.Call()
	return int(count)
}

// PostDeviceName 获取设备名称
// post_device_name(int device, char* name, int size);
func PostDeviceName(device int) string {
	name := make([]byte, 1024)
	size := C.post_device_name(C.int(device), (*C.char)(unsafe.Pointer(&name[0])), C.int(len(name)))
	//size, _, _ := _postDeviceName.Call(uintptr(device), uintptr(unsafe.Pointer(&name[0])), uintptr(len(name)))
	// 获取失败
	if size < 0 || size >= 1024 {
		return ""
	}
	return string(name[:size])
}
