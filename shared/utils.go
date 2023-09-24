package shared

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"github.com/trying2016/common-tools/utils"
	"github.com/zeebo/blake3"
	"os"
	"path"
	"path/filepath"
)

const (
	postExeName      = "post"
	postLibName      = "libpost"
	postLibToolsName = "libtools"
)

const (
	metadataName = "postdata_metadata.json"
	KeyName      = "key.bin"
	bitsPerLabel = 8 * 16
)

// FileSizeToNumLabels 文件大小转 NumLabels
func FileSizeToNumLabels(fileSize int64) int64 {
	return fileSize * 8 / bitsPerLabel
}

// NumLabelsToFileSize NumLabels转文件大小
func NumLabelsToFileSize(numLabels int64) int64 {
	return numLabels * 16
}

// ReadMetadata 读取metadata
func ReadMetadata(dir string) (*PostMetadata, error) {
	data, err := os.ReadFile(path.Join(dir, metadataName))
	if err != nil {
		return nil, err
	}
	var metadata PostMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

// ReadPrivateKey 读取私钥
func ReadPrivateKey(dir string) ([]byte, error) {
	data, err := os.ReadFile(path.Join(dir, KeyName))
	if err != nil {
		return nil, err
	}
	if len(data) == 128 {
		return hex.DecodeString(string(data))
	}
	return data, err
}

// CheckPrivate 检测私钥是否正确
func CheckPrivate(privateKey []byte, nodeId []byte) bool {
	msg := []byte("h9-spacemesh")
	sign := ed25519.Sign(privateKey, msg)
	return ed25519.Verify(nodeId, msg, sign)
}

// CheckPlotComplete 检测是否Plot完成
func CheckPlotComplete(dir string) bool {
	metadata, err := ReadMetadata(dir)
	if err != nil {
		return false
	}
	// 文件大小
	filesize := PlotFilesize(dir)
	return uint64(metadata.NumUnits)*metadata.LabelsPerUnit == uint64(FileSizeToNumLabels(filesize))
}

// PlotFilesize Plot的文件大小
func PlotFilesize(dir string) int64 {
	files, err := filepath.Glob(path.Join(dir, "postdata_*.bin"))
	if err != nil {
		return 0
	}
	size := int64(0)
	for _, filename := range files {
		size += utils.GetFileSize(filename)
	}
	return size
}

// HashMembershipTreeNode calculates internal node of
// the membership merkle tree.
func HashMembershipTreeNode(buf, lChild, rChild []byte) []byte {
	hasher := blake3.New()
	_, _ = hasher.Write([]byte{0x01})
	_, _ = hasher.Write(lChild)
	_, _ = hasher.Write(rChild)
	return hasher.Sum(buf)
}
