package post

import "C"
import (
	"errors"
	"runtime"
	"sync"
)

// 常量定义
var (
	// spacemesh randomx seed
	randomxSeed = []byte("spacemesh-randomx-cache-key")
)
var (
	errCacheFailed = errors.New("cache allocation failed")
	errDataset     = errors.New("dataset allocation failed")
)

const (
	RANDOMX_FLAG_DEFAULT      = 0
	RANDOMX_FLAG_LARGE_PAGES  = 1
	RANDOMX_FLAG_HARD_AES     = 2
	RANDOMX_FLAG_FULL_MEM     = 4
	RANDOMX_FLAG_JIT          = 8
	RANDOMX_FLAG_SECURE       = 16
	RANDOMX_FLAG_ARGON2_SSSE3 = 32
	RANDOMX_FLAG_ARGON2_AVX2  = 64
	RANDOMX_FLAG_ARGON2       = 96
)

// 单例
var singleRandomX *RandomX

func init() {
	singleRandomX = &RandomX{}
}

// GetRandomX 获取单例
func GetRandomX() *RandomX {
	return singleRandomX
}

type RandomX struct {
	cache        RandomXCache
	dataset      RandomXDataset
	flags        int32
	thread       int32
	affinity     int32
	affinityStep int32
}

func NewRandomX(flags, thread, affinity, affinityStep int32) (*RandomX, error) {
	r := &RandomX{
		flags:        flags,
		thread:       thread,
		affinity:     affinity,
		affinityStep: affinityStep,
	}
	err := r.initRandomX()
	if err != nil {
		return nil, err
	}
	return r, nil
}
func (r *RandomX) Release() {
	if r.cache != nil {
		FreeRandomXCache(r.cache)
		r.cache = nil
	}
	if r.dataset != nil {
		FreeRandomXDataset(r.dataset)
		r.dataset = nil
	}
}

func (r *RandomX) initRandomX() error {
	if r.cache != nil && r.dataset != nil {
		return nil
	}

	cache := NewRandomXCache(uint(r.flags))
	if cache == nil {
		return errCacheFailed
	}

	dataset := MallocDataset(uint(r.flags), cache)
	if dataset == nil {
		return errDataset
	}

	datasetItemCount := DatasetItemCount()
	initThreadCount := runtime.NumCPU()
	perThread := datasetItemCount / uint64(initThreadCount)
	remainder := datasetItemCount % uint64(initThreadCount)
	startItem := uint64(0)
	var job sync.WaitGroup
	for i := 0; i < initThreadCount; i++ {
		job.Add(1)
		count := perThread
		if i == initThreadCount-1 {
			count += remainder
		}
		go func(start, itemCount uint64) {
			InitDataset(dataset, start, itemCount)
			job.Done()
		}(startItem, count)
		startItem += count
	}
	job.Wait()

	r.cache = cache
	r.dataset = dataset
	return nil
}

// Pow pow
func (r *RandomX) Pow(input []byte, difficulty []byte) uint64 {
	pow := CallRandomXProve(uint(r.flags), r.cache, r.dataset, input, difficulty, r.thread, r.affinity, r.affinityStep)
	return uint64(pow)
}

// GetFlags 获取flags
func (r *RandomX) GetFlags() int32 {
	return r.flags
}
