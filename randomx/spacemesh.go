package randomx

import (
	"errors"
	"runtime"
	"sync"
)

// error 定义
var (
	// 创建cache失败
	errCacheFailed = errors.New("cache allocation failed")
	// 创建dataset失败
	errDataset = errors.New("dataset allocation failed")
)

// 常量定义
var (
	// spacemesh randomx seed
	randomxSeed = []byte("spacemesh-randomx-cache-key")
)

// ProveCallback 代理Prove回调
type ProveCallback func(input, difficulty []byte) uint64

// 单列模式
var singleSpacemesh *Spacemesh

func init() {
	singleSpacemesh = &Spacemesh{}
}

// GetSpacemesh 获取单列
func GetSpacemesh() *Spacemesh {
	return singleSpacemesh
}

type Spacemesh struct {
	cache        RandomxCache
	dataset      RandomxDataset
	flags        RandomxFlags
	thread       int32
	affinity     int32
	affinityStep int32
	// 代理Prove回调
	proveCallback ProveCallback
}

// Init 初始化
func (s *Spacemesh) Init(flag, thread, affinity, affinityStep int32) error {
	s.flags = RandomxFlags(flag)
	s.thread = thread
	s.affinity = affinity
	s.affinityStep = affinityStep
	cache, dataset, err := s.initRandomX(s.flags)
	if err != nil {
		return err
	}
	s.cache = cache
	s.dataset = dataset
	return nil
}

// initRandomX 根据flag创建RandomXCache,RandomXDataset
func (s *Spacemesh) initRandomX(flag RandomxFlags) (RandomxCache, RandomxDataset, error) {
	cache := RandomxAllocCache(flag)
	if cache == nil {
		return nil, nil, errCacheFailed
	}
	RandomxInitCache(cache, randomxSeed)
	dataset := RandomxAllocDataset(flag)
	if dataset == nil {
		return nil, nil, errDataset
	}

	datasetItemCount := RandomxDatasetItemCount()
	initThreadCount := runtime.NumCPU()
	perThread := datasetItemCount / uint32(initThreadCount)
	remainder := datasetItemCount % uint32(initThreadCount)
	startItem := uint32(0)
	var job sync.WaitGroup
	for i := 0; i < initThreadCount; i++ {
		job.Add(1)
		count := perThread
		if i == initThreadCount-1 {
			count += remainder
		}
		go func(start, itemCount uint32) {
			RandomxInitDataset(dataset, cache, start, itemCount)
			job.Done()
		}(startItem, count)
		startItem += count
	}
	job.Wait()
	return cache, dataset, nil
}

// Release 释放
func (s *Spacemesh) Release() {
	if s.cache != nil {
		RandomxReleaseCache(s.cache)
	}
	if s.dataset != nil {
		RandomxReleaseDataset(s.dataset)
	}
}

// GetFlags 获取flags
func (s *Spacemesh) GetFlags() RandomxFlags {
	return s.flags
}

// Pow pow
func (s *Spacemesh) Pow(powInput, difficulty []byte) uint64 {
	if s.proveCallback != nil {
		return s.proveCallback(powInput, difficulty)
	}
	return Prove(s.flags, s.cache, s.dataset, powInput, difficulty, s.thread, s.affinity, s.affinityStep)
}

// SetProveCallback 设置代理Prove回调
func (s *Spacemesh) SetProveCallback(proveCallback ProveCallback) {
	s.proveCallback = proveCallback
}
