package shared

import (
	"encoding/hex"
	"github.com/trying2016/common-tools/utils"
)

// PostMetadata is the data associated with the PoST init procedure, persisted in the datadir next to the init files.
type PostMetadata struct {
	NodeId          []byte
	CommitmentAtxId []byte

	LabelsPerUnit uint64
	NumUnits      uint32
	MaxFileSize   uint64
	Nonce         *uint64 `json:",omitempty"`
	NonceValue    []byte
	LastPosition  *uint64 `json:",omitempty"`
}

func (p *PostMetadata) NodeIdStr() string {
	return hex.EncodeToString(p.NodeId)
}

func (p *PostMetadata) Marshal() ([]byte, error) {
	data := utils.Map{
		"NodeId":          p.NodeId,
		"CommitmentAtxId": p.CommitmentAtxId,
		"LabelsPerUnit":   p.LabelsPerUnit,
		"NumUnits":        p.NumUnits,
		"MaxFileSize":     p.MaxFileSize,
		"Nonce":           p.Nonce,
		"NonceValue":      hex.EncodeToString(p.NonceValue),
		"LastPosition":    p.LastPosition,
	}.ToJson()
	return []byte(data), nil
}
