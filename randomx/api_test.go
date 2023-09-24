package randomx

import (
	"runtime"
	"sync"
	"testing"
)

func TestRandomxGetFlags(t *testing.T) {
	flag := RandomxGetFlags()
	t.Log(flag)
}

func TestProve(t *testing.T) {
	flag := RandomxGetFlags() | RANDOMX_FLAG_FULL_MEM
	cache := RandomxAllocCache(flag)
	if cache == nil {
		t.Fatal("Cache allocation failed")
	}
	seed := []byte("spacemesh-randomx-cache-key")
	RandomxInitCache(cache, seed)

	dataset := RandomxAllocDataset(flag)
	if dataset == nil {
		t.Fatal("Dataset allocation failed")
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
	input := []uint8{0, 0, 0, 0, 0, 0, 0, 1, 'h', 'e', 'l', 'l', 'o', '!', '!', '!',
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	difficulty := []uint8{0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff}
	pow := Prove(flag, cache, dataset, input, difficulty, 8, 0, 1)
	t.Logf("Pow: %v", pow)
}
