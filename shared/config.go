package shared

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultDataDirName = "data"

	DefaultComputeBatchSize = 1 << 20

	MinBitsPerLabel = 1
	MaxBitsPerLabel = 256
	BitsPerLabel    = 8 * 16

	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB

	defaultMaxFileSize = 4 * GiB
	minFileSize        = 1024
)

var DefaultDataDir string

func init() {
	home, _ := os.UserHomeDir()
	DefaultDataDir = filepath.Join(home, "post", DefaultDataDirName)
}

type Config struct {
	MinNumUnits   uint32
	MaxNumUnits   uint32
	LabelsPerUnit uint64

	K1 uint32 // K1 specifies the difficulty for a label to be a candidate for a proof.
	K2 uint32 // K2 is the number of labels below the required difficulty required for a proof.
	K3 uint32 // K3 is the size of the subset of proof indices that is validated.

	PowDifficulty [32]byte
}

// MainnetConfig returns the default config for mainnet.
func MainnetConfig() Config {
	cfg := Config{
		MinNumUnits:   4,
		MaxNumUnits:   1048576,    // max post size 64 PiB
		LabelsPerUnit: 4294967296, // 64GiB units
		K1:            26,
		K2:            37,
		K3:            37,
	}
	_, err := hex.Decode(cfg.PowDifficulty[:], []byte("00037ec8ec25e6d2c00000000000000000000000000000000000000000000000"))
	if err != nil {
		panic(err)
	}
	return cfg
}

// DefaultConfig returns the default config. These are intended for testing.
func DefaultConfig() Config {
	cfg := Config{
		MinNumUnits:   1,
		MaxNumUnits:   100,
		LabelsPerUnit: 512, // 8kB units
		K1:            26,
		K2:            37,
		K3:            37,
	}
	for i := range cfg.PowDifficulty {
		cfg.PowDifficulty[i] = 0xFF
	}
	return cfg
}

type InitOpts struct {
	DataDir     string
	NumUnits    uint32
	MaxFileSize uint64
	ProviderID  int
	Throttle    bool
	Scrypt      ScryptParams
	// ComputeBatchSize must be greater than 0
	ComputeBatchSize uint64
}

type ScryptParams struct {
	N, R, P uint
}

func (p *ScryptParams) Validate() error {
	if p.N == 0 {
		return errors.New("scrypt parameter N cannot be 0")
	}
	if p.R == 0 {
		return errors.New("scrypt parameter r cannot be 0")
	}
	if p.P == 0 {
		return errors.New("scrypt parameter p cannot be 0")
	}
	return nil
}

func DefaultLabelParams() ScryptParams {
	return ScryptParams{
		N: 8192,
		R: 1,
		P: 1,
	}
}

// BestProviderID can be used for selecting the most performant provider
// based on a short benchmarking session.
const BestProviderID = -1

// MainnetInitOpts returns the default InitOpts for mainnet.
func MainnetInitOpts() InitOpts {
	return InitOpts{
		DataDir:          DefaultDataDir,
		NumUnits:         4,
		MaxFileSize:      defaultMaxFileSize,
		ProviderID:       BestProviderID,
		Throttle:         false,
		Scrypt:           DefaultLabelParams(),
		ComputeBatchSize: DefaultComputeBatchSize,
	}
}

// DefaultInitOpts returns the default InitOpts. These are intended for testing.
func DefaultInitOpts() InitOpts {
	return InitOpts{
		DataDir:          DefaultDataDir,
		NumUnits:         2,
		MaxFileSize:      defaultMaxFileSize,
		ProviderID:       BestProviderID,
		Throttle:         false,
		Scrypt:           DefaultLabelParams(),
		ComputeBatchSize: DefaultComputeBatchSize,
	}
}

func Validate(cfg Config, opts InitOpts) error {
	if opts.NumUnits < cfg.MinNumUnits {
		return fmt.Errorf("invalid `opts.NumUnits`; expected: >= %d, given: %d", cfg.MinNumUnits, opts.NumUnits)
	}

	if opts.NumUnits > cfg.MaxNumUnits {
		return fmt.Errorf("invalid `opts.NumUnits`; expected: <= %d, given: %d", cfg.MaxNumUnits, opts.NumUnits)
	}

	if opts.MaxFileSize < minFileSize {
		return fmt.Errorf("invalid `opts.MaxFileSize`; expected: >= %d, given: %d", minFileSize, opts.MaxFileSize)
	}

	if opts.ComputeBatchSize == 0 {
		return fmt.Errorf("invalid `opts.ComputeBatchSize` expected: > 0, given: %d", opts.ComputeBatchSize)
	}

	if res := Uint64MulOverflow(cfg.LabelsPerUnit, uint64(opts.NumUnits)); res {
		return fmt.Errorf("uint64 overflow: `cfg.LabelsPerUnit` (%v) * `opts.NumUnits` (%v) exceeds the range allowed by uint64",
			cfg.LabelsPerUnit, opts.NumUnits)
	}

	numLabels := cfg.LabelsPerUnit * uint64(opts.NumUnits)
	if res := Uint64MulOverflow(numLabels, uint64(cfg.K1)); res {
		return fmt.Errorf("uint64 overflow: `cfg.LabelsPerUnit` * `opts.NumUnits` (%v) * `cfg.K1` (%v) exceeds the range allowed by uint64",
			numLabels, cfg.K1)
	}

	return nil
}
