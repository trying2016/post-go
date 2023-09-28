package randomx

import "C"

//// #cgo linux,amd64 LDFLAGS: -L./ -lrandomx -Wl,-rpath,./
// #cgo darwin,amd64 LDFLAGS:-L./ -lrandomx_macos -lm -stdlib=libc++ -lstdc++
// #cgo linux,amd64 LDFLAGS:-L./ -lrandomx_linux -lm -lstdc++
// #cgo windows,amd64 LDFLAGS:-L./ -lrandomx -Wl,-rpath,./
// #include "randomx.h"
// #include <stdlib.h>
import "C"

/*
//#cgo linux,amd64 LDFLAGS:-L./linux -lm -lpthread -static -static-libgcc -static-libstdc++
//#cgo darwin,amd64 LDFLAGS:-L./macos -lm -stdlib=libc++
//#cgo windows,amd64 LDFLAGS:-L./windows -static -static-libgcc -static-libstdc++
*/

type (
	RandomxFlags   C.randomx_flags
	RandomxCache   *C.randomx_cache
	RandomxDataset *C.randomx_dataset
)

const (
	RANDOMX_FLAG_DEFAULT      = C.RANDOMX_FLAG_DEFAULT
	RANDOMX_FLAG_LARGE_PAGES  = C.RANDOMX_FLAG_LARGE_PAGES
	RANDOMX_FLAG_HARD_AES     = C.RANDOMX_FLAG_HARD_AES
	RANDOMX_FLAG_FULL_MEM     = C.RANDOMX_FLAG_FULL_MEM
	RANDOMX_FLAG_JIT          = C.RANDOMX_FLAG_JIT
	RANDOMX_FLAG_SECURE       = C.RANDOMX_FLAG_SECURE
	RANDOMX_FLAG_ARGON2_SSSE3 = C.RANDOMX_FLAG_ARGON2_SSSE3
	RANDOMX_FLAG_ARGON2_AVX2  = C.RANDOMX_FLAG_ARGON2_AVX2
	RANDOMX_FLAG_ARGON2       = C.RANDOMX_FLAG_ARGON2
)

func RandomxGetFlags() RandomxFlags {
	return RandomxFlags(C.randomx_get_flags())
}

func RandomxAllocCache(flags RandomxFlags) RandomxCache {
	cache := C.randomx_alloc_cache(C.randomx_flags(flags))
	return RandomxCache(cache)
}

// RandomxInitCache void randomx_init_cache(randomx_cache *cache, const void *key, size_t keySize);
func RandomxInitCache(cache RandomxCache, key []byte) {
	cKey := C.CBytes(key)
	defer C.free(cKey)
	C.randomx_init_cache((*C.randomx_cache)(cache), cKey, C.size_t(len(key)))
}

func RandomxReleaseCache(cache RandomxCache) {
	//	_, _, _ = _randomxReleaseCache.Call(uintptr(unsafe.Pointer(cache)))
	C.randomx_release_cache((*C.randomx_cache)(cache))
}

func RandomxAllocDataset(flags RandomxFlags) RandomxDataset {
	//dataset, _, _ := _randomxAllocDataset.Call(uintptr(flags))
	//return RandomxDataset(unsafe.Pointer(dataset))
	return RandomxDataset(C.randomx_alloc_dataset(C.randomx_flags(flags)))
}

func RandomxDatasetItemCount() uint32 {
	//count, _, _ := _randomxDatasetItemCount.Call()
	//return uint32(count)
	return uint32(C.randomx_dataset_item_count())
}

func RandomxInitDataset(dataset RandomxDataset, cache RandomxCache, startItem, itemCount uint32) {
	//_, _, _ = _randomxInitDataset.Call(uintptr(unsafe.Pointer(dataset)), uintptr(unsafe.Pointer(cache)), uintptr(C.int(startItem)), uintptr(C.int(itemCount)))
	C.randomx_init_dataset((*C.randomx_dataset)(dataset), (*C.randomx_cache)(cache), C.int(startItem), C.int(itemCount))
}

func RandomxReleaseDataset(dataset RandomxDataset) {
	//_randomxReleaseDataset.Call(uintptr(unsafe.Pointer(dataset)))
	C.randomx_release_dataset((*C.randomx_dataset)(dataset))
}

func Prove(flags RandomxFlags, cache RandomxCache, dataset RandomxDataset, input, difficulty []byte, thread, affinity, affinityStep int32) uint64 {
	cInput := C.CBytes(input)
	defer C.free(cInput)
	cDifficulty := C.CBytes(difficulty)
	defer C.free(cDifficulty)

	//pow, _, _ := _prove.Call(uintptr(flags), uintptr(unsafe.Pointer(cache)), uintptr(unsafe.Pointer(dataset)),
	//	uintptr(unsafe.Pointer(cInput)), uintptr(len(input)), uintptr(unsafe.Pointer(cDifficulty)), uintptr(C.int(thread)),
	//	uintptr(C.int(affinity)), uintptr(C.int(affinityStep)))
	pow := C.prove(C.randomx_flags(flags), (*C.randomx_cache)(cache), (*C.randomx_dataset)(dataset),
		cInput, C.size_t(len(input)), cDifficulty, C.int(thread), C.int(affinity), C.int(affinityStep))
	return uint64(pow)
}
