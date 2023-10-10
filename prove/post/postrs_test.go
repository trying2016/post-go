package post

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/trying2016/post-go/randomx"
	"github.com/trying2016/post-go/shared"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

var (
	MainNetPowDifficulty, _ = hex.DecodeString("00037ec8ec25e6d2c00000000000000000000000000000000000000000000000")
	TestNetPowDifficulty, _ = hex.DecodeString("0ff37ec8ec25e6d2c00000000000000000000000000000000000000000000000")
	RandomxSeed             = []byte("spacemesh-randomx-cache-key")
)

const (
	maxNonce = 16 // 最大nonce数量
	K1       = 26
	K2       = 37
	K3       = 37
)

func initRandomx(flag randomx.RandomxFlags) (randomx.RandomxCache, randomx.RandomxDataset, error) {
	cache := randomx.RandomxAllocCache(flag)
	if cache == nil {
		return nil, nil, errCacheFailed
	}
	randomx.RandomxInitCache(cache, RandomxSeed)
	dataset := randomx.RandomxAllocDataset(flag)
	if dataset == nil {
		return nil, nil, errDataset
	}

	datasetItemCount := randomx.RandomxDatasetItemCount()
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
			randomx.RandomxInitDataset(dataset, cache, start, itemCount)
			job.Done()
		}(startItem, count)
		startItem += count
	}
	job.Wait()

	return cache, dataset, nil
}

func TestOpenCLProviders(t *testing.T) {
	list, err := OpenCLProviders()
	if err != nil {
		t.Fatal(err)
	}
	for i, device := range list {
		t.Log(i, device.ID, device.Model, device.DeviceType)
	}
}
func TestProof(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	SetLogCallback(Debug)
	dir := "/Users/trying/Documents/plot/post_1"
	metadata, err := shared.ReadMetadata(dir)
	if err != nil {
		t.Fatal(err)
	}

	randomxFlag := randomx.RandomxGetFlags()
	randomxFlag = randomxFlag | randomx.RANDOMX_FLAG_FULL_MEM
	cache, dataset, err := initRandomx(randomxFlag)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		randomx.RandomxReleaseCache(cache)
		randomx.RandomxReleaseDataset(dataset)
	}()
	SetRandomxCallback(func(input, difficulty []byte) uint64 {
		t.Logf("Input: %v diff: %v", hex.EncodeToString(input), hex.EncodeToString(difficulty))
		return randomx.Prove(randomxFlag, cache, dataset, input, difficulty, 15, -1, 1)
	})
	verify, err := NewVerifier(GetRecommendedPowFlags())
	if err != nil {
		t.Fatal(err)
	}

	challenge := sha256.Sum256([]byte("1"))

	proof, err := GenerateProof(dir, challenge[:], 16, 16, K1, K2, TestNetPowDifficulty, GetRecommendedPowFlags(), metadata.NodeId, -1)
	if err != nil {
		t.Fatal(err)
	}

	err = verify.VerifyProof(proof, metadata, K1, K2, K3, challenge[:], TestNetPowDifficulty, metadata.NodeId, TranslateScryptParams(8192, 1, 1))
	if err != nil {
		t.Fatal(err)
	}

	proofData, _ := json.Marshal(proof)
	metaData, _ := json.Marshal(metadata)
	t.Log(string(proofData), string(metaData))
}

func TestAes(t *testing.T) {
	key := []byte{116, 14, 225, 154, 198, 165, 75, 70, 183, 102, 247, 44, 171, 175, 107, 76}
	data := []byte{68, 147, 113, 158, 232, 185, 224, 236, 161, 131, 101, 213, 9, 224, 99, 170}
	out := make([]byte, len(data))
	aes := CreateAes(key)
	EncryptAes(aes, data, out, 16)
	FreeAes(aes)
	t.Log(out)
}

func TestAesPerformance(t *testing.T) {
	dataSize := 1024 * 1024 * 16
	loopCount := 100

	key := []byte{116, 14, 225, 154, 198, 165, 75, 70, 183, 102, 247, 44, 171, 175, 107, 76}
	data := make([]byte, dataSize)
	out := make([]byte, dataSize)
	aes := CreateAes(key)
	rand.Read(data)
	start := time.Now().UnixMilli()
	for i := 0; i < loopCount; i++ {
		EncryptAes(aes, data, out, 128)
	}
	costTime := time.Now().UnixMilli() - start

	totalCount := int64(loopCount) * (int64(dataSize) / 16)
	t.Log("Total:", totalCount, "Cost:", costTime, "Speed:", float64(totalCount)/float64(costTime)*1000/1000/1000, "MB/s")
}

func TestK2Pow(t *testing.T) {
	flags := GetRecommendedPowFlags()
	cache := NewRandomXCache(uint(flags))
	defer FreeRandomXCache(cache)
	dataset := MallocDataset(uint(flags), cache)
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

	//dataset := NewRandomXDataset(uint(flags), cache, 0, 0)
	defer FreeRandomXDataset(dataset)

	input := []uint8{0, 0, 0, 0, 0, 0, 0, 1, 'h', 'e', 'l', 'l', 'o', '!', '!', '!',
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	difficulty := []uint8{0x00, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff}

	pow := CallRandomXProve(uint(flags), cache, dataset, input, difficulty, 8, 0, 1)

	t.Logf("Pow: %v", pow)
}
