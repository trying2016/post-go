package post_go

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/trying2016/post-go/randomx"
	"github.com/trying2016/post-go/shared"
	"testing"
)

var (
	TestNetPowDifficulty, _ = hex.DecodeString("0ff37ec8ec25e6d2c00000000000000000000000000000000000000000000000")
	MainNetPowDifficulty, _ = hex.DecodeString("00037ec8ec25e6d2c00000000000000000000000000000000000000000000000")
)

// 日志回调函数
func logCallback(level int, msg string) {
	var msgLevel = []string{"", "INFO", "ERROR"}
	fmt.Println(msgLevel[level], msg)
}

func TestGenerateProof(t *testing.T) {
	var err error

	// 设置randomx
	randomxFlag := randomx.RandomxGetFlags()
	randomxFlag = randomxFlag | randomx.RANDOMX_FLAG_FULL_MEM

	if err := randomx.GetSpacemesh().Init(int32(randomxFlag), 16, 0, 0); err != nil {
		t.Fatal(err)
	}
	defer randomx.GetSpacemesh().Release()

	challenge := sha256.Sum256([]byte("1"))

	// 生成证明
	dir := "/Users/trying/Documents/plot/post_1"
	proof, err := GenerateProof(dir, challenge[:], 128, shared.K1, shared.K2, TestNetPowDifficulty)
	if err != nil {
		t.Fatal(err)
	}

	proofData, _ := json.Marshal(proof)
	t.Log(string(proofData))
}
