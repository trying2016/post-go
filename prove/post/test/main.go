package main

import (
	"fmt"
	post_cpu "github.com/trying2016/post-go/prove/post"
	"github.com/trying2016/post-go/randomx"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("RandomxGetFlags", randomx.RandomxGetFlags())
	dataSize := 1024 * 1024 * 16
	loopCount := 1

	key := []byte{116, 14, 225, 154, 198, 165, 75, 70, 183, 102, 247, 44, 171, 175, 107, 76}
	data := make([]byte, dataSize)
	out := make([]byte, dataSize)
	aes := post_cpu.CreateAes(key)
	rand.Read(data)

	for k := 0; k < 1024; k += 10 {
		start := time.Now().UnixMilli()
		batchSize := 1024 * (k + 1)
		
		for i := 0; i < loopCount; i++ {
			post_cpu.EncryptAes(aes, data, out, batchSize)
		}
		costTime := time.Now().UnixMilli() - start

		totalCount := int64(loopCount) * (int64(dataSize) / 16)
		println("batchSize:", batchSize, "Total:", totalCount, "Cost:", costTime, "Speed:", fmt.Sprintf("%.02f", float64(totalCount)/float64(costTime)*1000/1000/1000), "MB/s")
	}
}
