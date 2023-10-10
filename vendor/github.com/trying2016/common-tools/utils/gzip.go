package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
)

func Compress(data []byte) []byte {
	var zBuf bytes.Buffer
	zipWrite := gzip.NewWriter(&zBuf)

	if _, err := zipWrite.Write(data); err != nil {
		fmt.Println("-----gzip is faild,err:", err)
	}
	zipWrite.Close()

	return zBuf.Bytes()
}

func UnCompress(data []byte) []byte {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return data
	}
	unBody, err := ioutil.ReadAll(gzipReader)
	gzipReader.Close()
	return unBody
}
