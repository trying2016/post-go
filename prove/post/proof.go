package post

import "C"

// #cgo LDFLAGS: -lpost
// #include <stdlib.h>
// #include "post.h"
// uint64_t randomXPow(const char*, uintptr_t,const char*, uintptr_t);
import "C"

import (
	"errors"
	"fmt"
	"github.com/trying2016/post-go/shared"
	"math"
	"reflect"
	"sync"
	"unsafe"
)

type ScryptParams = C.ScryptParams

// Translate scrypt parameters expressed as N,R,P to Nfactor, Rfactor and Pfactor
// that are understood by scrypt-jane.
// Relation:
// N = 1 << (nfactor + 1)
// r = 1 << rfactor
// p = 1 << pfactor
func TranslateScryptParams(n, r, p uint) ScryptParams {
	return ScryptParams{
		nfactor: C.uint8_t(math.Log2(float64(n))) - 1,
		rfactor: C.uint8_t(math.Log2(float64(r))),
		pfactor: C.uint8_t(math.Log2(float64(p))),
	}
}

type postOptions struct {
	powCreatorId []byte
}

type PostOptionFunc func(*postOptions) error

func WithPowCreator(id []byte) PostOptionFunc {
	return func(opts *postOptions) error {
		opts.powCreatorId = id
		return nil
	}
}

func GenerateProof(dataDir string, challenge []byte, nonces, threads uint, K1, K2 uint32, powDifficulty []byte, powFlags PowFlags, creatorId []byte) (*shared.Proof, error) {
	dataDirPtr := C.CString(dataDir)
	defer C.free(unsafe.Pointer(dataDirPtr))

	challengePtr := C.CBytes(challenge)
	defer C.free(challengePtr)

	var powCreatorId unsafe.Pointer
	if creatorId != nil {
		powCreatorId = C.CBytes(creatorId)
		defer C.free(powCreatorId)
	}

	config := C.Config{
		k1: C.uint32_t(K1),
		k2: C.uint32_t(K2),
	}
	for i, b := range powDifficulty {
		config.pow_difficulty[i] = C.uchar(b)
	}

	cProof := C.generate_proof(
		dataDirPtr,
		(*C.uchar)(challengePtr),
		config,
		C.size_t(nonces),
		C.size_t(threads),
		powFlags,
		(*C.uchar)(powCreatorId),
		C.randomXPow,
	)

	if cProof == nil {
		return nil, fmt.Errorf("got nil")
	}
	defer C.free_proof(cProof)

	indices := make([]uint8, cProof.indices.len)
	copy(indices, unsafe.Slice((*uint8)(unsafe.Pointer(cProof.indices.ptr)), cProof.indices.len))

	return &shared.Proof{
		Nonce:   uint32(cProof.nonce),
		Indices: indices,
		Pow:     uint64(cProof.pow),
	}, nil
}

type PowFlags = C.RandomXFlag

// Get the recommended PoW flags.
//
// Does not include:
// * FLAG_LARGE_PAGES
// * FLAG_FULL_MEM
// * FLAG_SECURE
//
// The above flags need to be set manually, if required.
func GetRecommendedPowFlags() PowFlags {
	return C.recommended_pow_flags()
}

const (
	// Use the full dataset. AKA "Fast mode".
	PowFastMode = C.RandomXFlag_FLAG_FULL_MEM
	// Allocate memory in large pages.
	PowLargePages = C.RandomXFlag_FLAG_LARGE_PAGES
	// Use JIT compilation support.
	PowJIT = C.RandomXFlag_FLAG_JIT
	// When combined with FLAG_JIT, the JIT pages are never writable and executable at the same time.
	PowSecure = C.RandomXFlag_FLAG_SECURE
	// Use hardware accelerated AES.
	PowHardAES = C.RandomXFlag_FLAG_HARD_AES
	// Optimize Argon2 for CPUs with the SSSE3 instruction set.
	PowArgon2SSSE3 = C.RandomXFlag_FLAG_ARGON2_SSSE3
	// Optimize Argon2 for CPUs with the SSSE3 instruction set.
	PowArgon2AVX2 = C.RandomXFlag_FLAG_ARGON2_AVX2
	// Optimize Argon2 for CPUs without the AVX2 or SSSE3 instruction sets.
	PowArgon2 = C.RandomXFlag_FLAG_ARGON2
)

type Verifier struct {
	inner     *C.Verifier
	closeOnce sync.Once
}

// Create a new verifier.
// The verifier must be closed after use with Close().
func NewVerifier(powFlags PowFlags) (*Verifier, error) {
	verifier := Verifier{}
	result := C.new_verifier(powFlags, &verifier.inner)
	if result != C.Ok {
		return nil, fmt.Errorf("failed to create verifier")
	}
	return &verifier, nil
}

func (v *Verifier) Close() error {
	v.closeOnce.Do(func() { C.free_verifier(v.inner) })
	return nil
}

func (v *Verifier) VerifyProof(
	proof *shared.Proof,
	metadata *shared.PostMetadata,
	k1, k2, k3 uint32,
	challenge, powDifficulty, creatorId []byte,
	scryptParams ScryptParams,
) error {

	if proof == nil {
		return errors.New("proof cannot be nil")
	}
	if metadata == nil {
		return errors.New("metadata cannot be nil")
	}
	if len(metadata.NodeId) != 32 {
		return errors.New("node id length must be 32")
	}
	if len(metadata.CommitmentAtxId) != 32 {
		return errors.New("commitment atx id length must be 32")
	}
	if len(challenge) != 32 {
		return errors.New("challenge length must be 32")
	}
	if len(proof.Indices) == 0 {
		return errors.New("proof indices are empty")
	}

	config := C.Config{
		k1:     C.uint32_t(k1),
		k2:     C.uint32_t(k2),
		k3:     C.uint32_t(k3),
		scrypt: scryptParams,
	}
	for i, b := range powDifficulty {
		config.pow_difficulty[i] = C.uchar(b)
	}

	indicesSliceHdr := (*reflect.SliceHeader)(unsafe.Pointer(&proof.Indices))
	cProof := C.Proof{
		nonce: C.uint32_t(proof.Nonce),
		pow:   C.uint64_t(proof.Pow),
		indices: C.ArrayU8{
			ptr: (*C.uchar)(unsafe.Pointer(indicesSliceHdr.Data)),
			len: C.size_t(indicesSliceHdr.Len),
			cap: C.size_t(indicesSliceHdr.Cap),
		},
	}

	if creatorId != nil {
		minerIdSliceHdr := (*reflect.SliceHeader)(unsafe.Pointer(&creatorId))
		cProof.pow_creator = C.ArrayU8{
			ptr: (*C.uchar)(unsafe.Pointer(minerIdSliceHdr.Data)),
			len: C.size_t(minerIdSliceHdr.Len),
			cap: C.size_t(minerIdSliceHdr.Cap),
		}
	}

	cMetadata := C.ProofMetadata{
		node_id:           *(*[32]C.uchar)(unsafe.Pointer(&metadata.NodeId[0])),
		commitment_atx_id: *(*[32]C.uchar)(unsafe.Pointer(&metadata.CommitmentAtxId[0])),
		challenge:         *(*[32]C.uchar)(unsafe.Pointer(&challenge[0])),
		num_units:         C.uint32_t(metadata.NumUnits),
		labels_per_unit:   C.uint64_t(metadata.LabelsPerUnit),
	}
	result := C.verify_proof(
		v.inner,
		cProof,
		&cMetadata,
		config,
	)

	switch result {
	case C.Ok:
		return nil
	case C.Invalid:
		return fmt.Errorf("invalid proof")
	case C.InvalidArgument:
		return fmt.Errorf("invalid argument")
	default:
		return fmt.Errorf("unknown error")
	}
}

// VerifyVRFNonce ensures the validity of a nonce for a given node.
// AtxId is the id of the ATX that was selected by the node for its commitment.
func VerifyVRFNonce(nonce *uint64, m *shared.PostMetadata, labelScrypt shared.ScryptParams) error {
	if nonce == nil {
		return errors.New("invalid `nonce` value; expected: non-nil, given: nil")
	}

	if len(m.NodeId) != 32 {
		return fmt.Errorf("invalid `nodeId` length; expected: 32, given: %v", len(m.NodeId))
	}

	if len(m.CommitmentAtxId) != 32 {
		return fmt.Errorf("invalid `commitmentAtxId` length; expected: 32, given: %v", len(m.CommitmentAtxId))
	}

	numLabels := uint64(m.NumUnits) * uint64(m.LabelsPerUnit)
	difficulty := shared.PowDifficulty(numLabels)

	initializer, err := NewInitializer(uint32(cCPUProviderID()), uint32(labelScrypt.N), shared.CommitmentBytes(m.NodeId, m.CommitmentAtxId), difficulty)
	if err != nil {
		return err
	}
	defer FreeInitializer(initializer)

	_, calcNonce, err := ScryptPositions(initializer, *nonce, *nonce)
	if err != nil {
		return err
	}
	if calcNonce == nil || *calcNonce != *nonce {
		return fmt.Errorf("nonce %v is not valid for node %v", *nonce, m.NodeId)
	}
	return nil
}

// VerifyProof Verify proof data
func VerifyProof(context *Verifier,
	proof *shared.Proof,
	metadata *shared.PostMetadata,
	challenge,
	powDifficulty []byte,
	powCreator []byte,
) error {
	params := shared.DefaultLabelParams()
	return context.VerifyProof(proof,
		metadata,
		shared.K1,
		shared.K2,
		shared.K3,
		challenge,
		powDifficulty,
		powCreator,
		TranslateScryptParams(params.N, params.R, params.P))

}
