package post

import (
	"io"
	"unsafe"
)

// #cgo LDFLAGS: -L./ -lpost -Wl,-rpath,./
// #include "post.h"
/*
#define BATCH_SIZE 16
	void batch_encrypt_aes(struct Aes *aes_ptr,
                 const uint8_t *input_data,
                 uint8_t *output_data,
                 int size, int batchSize)
{
	for(int i=0;i<size;i += batchSize){
		encrypt_aes(aes_ptr, input_data+i, output_data+i, batchSize);
	}
}
*/
import "C"

import (
	"errors"
	"fmt"
)

// ErrScryptClosed is returned when calling a method on an already closed Scrypt instance.
var ErrScryptClosed = errors.New("scrypt has been closed")

func OpenCLProviders() ([]Provider, error) {
	return cGetProviders()
}

func CPUProviderID() uint {
	return cCPUProviderID()
}

// ScryptPositionsResult is the result of a ScryptPositions call.
type ScryptPositionsResult struct {
	Output      []byte  // The output of the scrypt computation.
	IdxSolution *uint64 // The index of a solution to the proof of work (if checked for).
}

type Scrypter interface {
	io.Closer
	Positions(start, end uint64) (ScryptPositionsResult, error)
}

type option struct {
	providerID *uint

	commitment    []byte
	n             uint
	vrfDifficulty []byte
}

func (o *option) validate() error {
	if o.providerID == nil {
		return errors.New("`providerID` is required")
	}

	if o.commitment == nil {
		return errors.New("`commitment` is required")
	}

	if o.n > 0 && o.n&(o.n-1) != 0 {
		return fmt.Errorf("invalid `n`; expected: power of 2, given: %v", o.n)
	}

	return nil
}

// OptionFunc is a function that sets an option for a Scrypt instance.
type OptionFunc func(*option) error

// WithProviderID sets the ID of the openCL provider to use.
func WithProviderID(id uint) OptionFunc {
	return func(opts *option) error {
		opts.providerID = new(uint)
		*opts.providerID = id
		return nil
	}
}

// WithCommitment sets the commitment to use for the scrypt computation.
func WithCommitment(commitment []byte) OptionFunc {
	return func(opts *option) error {
		if len(commitment) != 32 {
			return fmt.Errorf("invalid `commitment` length; expected: 32, given: %v", len(commitment))
		}

		opts.commitment = commitment
		return nil
	}
}

// WithScryptN sets the N parameter for the scrypt computation.
func WithScryptN(n uint) OptionFunc {
	return func(opts *option) error {
		opts.n = n
		return nil
	}
}

// WithVRFDifficulty sets the difficulty for the VRF nonce computation.
func WithVRFDifficulty(difficulty []byte) OptionFunc {
	return func(opts *option) error {
		if len(difficulty) != 32 {
			return fmt.Errorf("invalid `difficulty` length; expected: 32, given: %v", len(difficulty))
		}

		opts.vrfDifficulty = difficulty
		return nil
	}
}

// Scrypt is a scrypt computation instance. It communicates with post-rs to perform
// the scrypt computation on the GPU or CPU.
type Scrypt struct {
	options *option
	init    *C.Initializer
}

// NewScrypt creates a new Scrypt instance.
func NewScrypt(opts ...OptionFunc) (*Scrypt, error) {
	options := &option{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	if err := options.validate(); err != nil {
		return nil, err
	}

	init, err := NewInitializer(uint32(*options.providerID), uint32(options.n), options.commitment, options.vrfDifficulty)
	if err != nil {
		return nil, err
	}
	if *options.providerID != cCPUProviderID() {
		gpuMtx.Device(*options.providerID).Lock()
	}

	return &Scrypt{
		options: options,
		init:    init,
	}, nil
}

// Close closes the Scrypt instance.
func (s *Scrypt) Close() error {
	if s.init == nil {
		return ErrScryptClosed
	}

	FreeInitializer(s.init)
	if *s.options.providerID != cCPUProviderID() {
		gpuMtx.Device(*s.options.providerID).Unlock()
	}
	s.init = nil
	return nil
}

// Positions computes the scrypt output for the given options.
func (s *Scrypt) Positions(start, end uint64) (ScryptPositionsResult, error) {
	if s.init == nil {
		return ScryptPositionsResult{}, ErrScryptClosed
	}

	if start > end {
		return ScryptPositionsResult{}, fmt.Errorf("invalid `start` and `end`; expected: start <= end, given: %v > %v", start, end)
	}

	if err := s.options.validate(); err != nil {
		return ScryptPositionsResult{}, err
	}

	output, idxSolution, err := ScryptPositions(s.init, start, end)
	return ScryptPositionsResult{
		Output:      output,
		IdxSolution: idxSolution,
	}, err
}

// CreateAes struct Aes *create_aes(const uint8_t *key);
func CreateAes(key []byte) *C.Aes {
	return C.create_aes((*C.uint8_t)(&key[0]))
}

/*
EncryptAes
void encrypt_aes(struct Aes *aes_ptr,

	const uint8_t *input_data,
	uint8_t *output_data,
	uintptr_t size);
*/
func EncryptAes(aes *C.Aes, input []byte, output []byte, batchSize int) {
	C.batch_encrypt_aes(aes, (*C.uint8_t)(&input[0]), (*C.uint8_t)(&output[0]), C.int(len(input)), C.int(batchSize))
}
func EncryptAesUint(aes *C.Aes, input []byte, output []uint64, batchSize int) {
	C.batch_encrypt_aes(aes, (*C.uint8_t)(&input[0]), (*C.uint8_t)(unsafe.Pointer(&output[0])), C.int(len(input)), C.int(batchSize))
}

// FreeAes void free_aes(struct Aes *aes_ptr);
func FreeAes(aes *C.Aes) {
	C.free_aes(aes)
}
