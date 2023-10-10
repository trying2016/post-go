// Package prove Prove
/**
 * @Author: trying
 * @Description: 生成proof
 * @File:  prove.go
 * @Version: 1.0.0
 * @Date: 2023/9/24 15:17
 */
package prove

import (
	"errors"
	"github.com/trying2016/post-go/prove/post"
	post_go "github.com/trying2016/post-go/prove/prove_go"
	"github.com/trying2016/post-go/shared"
	"sync"
)

const (
	Error = iota + 1
	Warn
	Info
	Debug
	Trace
)

// ProofType proof 类型
type ProofType int32

const (
	ProofType_Rust ProofType = iota
	PowType_Go
)

var postLogOnce sync.Once

// NewProve
func NewProve(proofType ProofType, thread, nonces int32) (*Prove, error) {
	switch proofType {
	case ProofType_Rust:
		postLogOnce.Do(func() {
			post.SetRandomxCallback(func(input, difficulty []byte) uint64 {
				return post.GetRandomX().Pow(input, difficulty)
			})
			//post.SetLogCallback(post.Info)
		})

		return &Prove{
			proofType: proofType,
			thread:    thread,
			nonces:    nonces,
		}, nil
	case PowType_Go:
		post_go.SetRandomxCallback(func(input, difficulty []byte) uint64 {
			return post.GetRandomX().Pow(input, difficulty)
		})
		return &Prove{
			proofType: proofType,
			thread:    thread,
			nonces:    nonces,
		}, nil
	default:
		return nil, errors.New("unknown proof type")
	}
}

type Prove struct {
	proofType ProofType
	thread    int32
	nonces    int32
}

// GenerateProof 生成proof
func (p *Prove) GenerateProof(dataDir string, challenge []byte, powDifficulty []byte, creatorId []byte, threadId int32) (*shared.Proof, error) {
	switch p.proofType {
	case ProofType_Rust:
		return post.GenerateProof(dataDir,
			challenge,
			uint(p.nonces),
			uint(p.thread),
			shared.K1,
			shared.K2,
			powDifficulty,
			post.PowFlags(post.GetRandomX().GetFlags()),
			creatorId,
			threadId)
	case PowType_Go:
		return post_go.GenerateProof(dataDir,
			challenge,
			uint32(p.nonces),
			shared.K1,
			shared.K2,
			powDifficulty)
	default:
		return nil, errors.New("unknown proof type")
	}
}

// SetPostLogLevel 设置日志级别
func SetPostLogLevel(level int32) {
	post.SetLogCallback(int(level))
}
