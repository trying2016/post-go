package post_go

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/trying2016/post-go/prove/post"
	"github.com/trying2016/post-go/shared"
	"math"
	"math/big"
	"sync"
	"time"
)

const (
	LABEL_SIZE     = 16
	BLOCK_SIZE     = 16
	AES_BATCH      = 8
	CHUNK_SIZE     = BLOCK_SIZE * AES_BATCH
	NONCES_PER_AES = 16
	KEY_SIZE       = 16
	BUNCH_SIZE     = 1024 * 1024
)

var (
	// 创建PoST上下文失败
	errCreatePostContext = errors.New("create post context failed")
)

type RandomxCallback func(input, difficulty []byte) uint64

var randomxCallback RandomxCallback

func SetRandomxCallback(callback RandomxCallback) {
	randomxCallback = callback
}

func provingDifficulty(k1 uint32, numLabels uint64) (uint64, error) {
	if numLabels <= 0 {
		return 0, fmt.Errorf("number of label blocks must be > 0")
	}
	if numLabels <= uint64(k1) {
		return 0, fmt.Errorf("number of labels (%d) must be bigger than k1 (%d)", numLabels, k1)
	}
	difficulty := uint64(math.Pow(2, 64) * float64(k1) / float64(numLabels))
	return difficulty, nil
}

type ProvingParams struct {
	Difficulty    uint64
	PoWDifficulty [32]byte
}

/*
	pub fn new(metadata: &PostMetadata, cfg: &Config) -> eyre::Result<Self> {
	       let num_labels = metadata.num_units as u64 * metadata.labels_per_unit;
	       let mut pow_difficulty = [0u8; 32];
	       let difficulty_scaled = U256::from_big_endian(&cfg.pow_difficulty) / metadata.num_units;
	       difficulty_scaled.to_big_endian(&mut pow_difficulty);
	       Ok(Self {
	           difficulty: proving_difficulty(cfg.k1, num_labels)?,
	           pow_difficulty,
	       })
	   }
*/
func toBigEndian(nWords int, value []big.Word, bytes []byte) {
	for i := 0; i < nWords; i++ {
		binary.BigEndian.PutUint64(bytes[8*i:], uint64(value[nWords-i-1]))
	}
}
func NewProvingParams(metadata *shared.PostMetadata, cfg *Config) (*ProvingParams, error) {
	numLabels := uint64(metadata.NumUnits) * metadata.LabelsPerUnit
	var powDifficulty [32]byte
	difficultyScaled := new(big.Int).SetBytes(cfg.PowDifficulty)
	difficultyScaled = difficultyScaled.Div(difficultyScaled, new(big.Int).SetUint64(uint64(metadata.NumUnits)))
	//copy(powDifficulty[:], difficultyScaled.Bytes())
	toBigEndian(4, difficultyScaled.Bits(), powDifficulty[:])

	diff, err := provingDifficulty(cfg.K1, numLabels)
	if err != nil {
		return nil, err
	}
	return &ProvingParams{
		Difficulty:    diff,
		PoWDifficulty: powDifficulty,
	}, nil
}

type Cipher struct {
	Aes   *post.Aes
	GoAes cipher.Block
	Pow   uint64
}

type Prover8_56 struct {
	DifficultyMSB uint8
	DifficultyLSB uint64
	groupCipher   []*Cipher
	nonceCipher   []*Cipher
	tmpOut        []byte
	startNonce    uint32
	nonces        uint32
}

// 加个全局锁，防止randomx并发
var randomxLock sync.Mutex

func NewProver8_56(challenge []byte, nonces []uint32, params *ProvingParams, minerID []byte) (*Prover8_56, error) {
	randomxLock.Lock()
	defer randomxLock.Unlock()

	if nonces[0]%NONCES_PER_AES != 0 {
		return nil, errors.New("nonces must start at a multiple of 16")
	}
	if len(nonces) == 0 || len(nonces)%int(NONCES_PER_AES) != 0 {
		return nil, errors.New("nonces must be a multiple of 16")
	}

	fmt.Printf("calc nonces %v...%v \n", nonces[0], nonces[len(nonces)-1])

	nonceGroup := nonceGroupRange(nonces, NONCES_PER_AES)
	gropuKeys := make([]byte, KEY_SIZE*len(nonceGroup))
	noncesKeys := make([]byte, KEY_SIZE*len(nonces))
	var groupCipherList []*Cipher
	var nonceCipherList []*Cipher
	for i, group := range nonceGroup {
		//0~6: nonce, 7: group, 8~15: challenge, 16~47: minerID
		powInput := make([]byte, 8+8+32)
		powInput[7] = uint8(group)
		copy(powInput[8:8+8], challenge[:8])
		copy(powInput[16:16+32], minerID[:])

		//hexInput := hex.EncodeToString(powInput)
		//hexDifficulty := hex.EncodeToString(params.PoWDifficulty[:])
		pow := randomxCallback(powInput, params.PoWDifficulty[:])
		key := NewAesCipherKey(challenge, uint32(group), pow)

		cipher := &Cipher{
			Aes: post.NewAes(key),
			Pow: pow,
		}

		// fmt.Println("group key", hex.EncodeToString(key))
		copy(gropuKeys[i*KEY_SIZE:], key)
		groupCipherList = append(groupCipherList, cipher)
	}

	start := nonces[0]
	startGroup := calcNonceGroup(start, NONCES_PER_AES)
	for i, nonce := range nonces {
		group := calcNonceGroup(nonce, NONCES_PER_AES)
		pow := groupCipherList[group-startGroup].Pow
		key := NewLazyAesCipherKey(challenge, nonce, group, pow)

		goAes, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		nonceCipherList = append(nonceCipherList, &Cipher{
			Aes:   post.NewAes(key),
			Pow:   pow,
			GoAes: goAes,
		})
		copy(noncesKeys[i*KEY_SIZE:], key)
	}

	difficultyMSB, difficultyLSB := splitDifficulty(params.Difficulty)

	return &Prover8_56{
		groupCipher:   groupCipherList,
		nonceCipher:   nonceCipherList,
		DifficultyMSB: difficultyMSB,
		DifficultyLSB: difficultyLSB,
		tmpOut:        make([]byte, BUNCH_SIZE),
		startNonce:    nonces[0],
		nonces:        uint32(len(nonces)),
	}, nil
}

// Pow 根据nonce获取对应的pow
func (p *Prover8_56) Pow(nonce uint32) uint64 {
	group := calcNonceGroup(nonce, NONCES_PER_AES)
	return p.groupCipher[group].Pow
}

// Destroy 销毁
func (p *Prover8_56) Destroy() {
	for _, cipher := range p.groupCipher {
		cipher.Aes.Free()
	}
	for _, cipher := range p.nonceCipher {
		cipher.Aes.Free()
	}
}

// prove generates a proof for the given data and baseIndex.
/*
fn prove<F>(&self, batch: &[u8], mut index: u64, mut consume: F) -> Option<(u32, Vec<u64>)>
    where
        F: FnMut(u32, u64) -> Option<Vec<u64>>,
    {
        let mut u8s = [0u8; CHUNK_SIZE];

        for chunk in batch.chunks_exact(CHUNK_SIZE) {
            for cipher in &self.ciphers {
                _ = cipher.aes.encrypt_padded_b2b::<NoPadding>(chunk, &mut u8s);

                for (offset, &msb) in u8s.iter().enumerate() {
                    if msb <= self.difficulty_msb {
                        if msb == self.difficulty_msb {
                            // Check LSB
                            let nonce =
                                calc_nonce(cipher.nonce_group, Self::NONCES_PER_AES, offset);
                            let label_offset = offset / Self::NONCES_PER_AES as usize * LABEL_SIZE;
                            if let Some(p) = self.check_lsb(
                                &chunk[label_offset..label_offset + LABEL_SIZE],
                                nonce,
                                offset,
                                index,
                                &mut consume,
                            ) {
                                return Some(p);
                            }
                        } else {
                            // valid label
                            let index = index + (offset as u32 / Self::NONCES_PER_AES) as u64;
                            let nonce =
                                calc_nonce(cipher.nonce_group, Self::NONCES_PER_AES, offset);
                            if let Some(indexes) = consume(nonce, index) {
                                return Some((nonce, indexes));
                            }
                        }
                    }
                }
            }
            index += AES_BATCH as u64;
        }

        None
    }
*/

var showTime int64

// prove 满足consume的调节后，返回true，读盘停止运行
func (p *Prover8_56) prove(batch []byte, baseIndex uint64, consume func(uint32, uint64) bool) bool {
	count := int64(len(batch)) / 16 * int64(len(p.groupCipher))
	groupCost := int64(0)
	calcGroup := func(i int, cipher *Cipher) {
		group := uint32(i) + p.startNonce/16
		tmpOut := make([]byte, len(batch))
		t := time.Now().UnixNano()
		cipher.Aes.Encrypt(batch, tmpOut, len(batch))
		t = time.Now().UnixNano() - t
		groupCost += t
		for offset, msb := range tmpOut {
			if msb <= p.DifficultyMSB {
				if msb == p.DifficultyMSB {
					// Check LSB
					nonce := calcNonce(group, uint32(offset), NONCES_PER_AES)
					labelOffset := offset / int(NONCES_PER_AES) * LABEL_SIZE
					count++
					if p.checkLSB(batch[labelOffset:labelOffset+LABEL_SIZE], nonce, uint32(offset), baseIndex, consume) {
						return
					}
				} else {
					// valid label
					index := baseIndex + uint64(offset/int(NONCES_PER_AES))
					nonce := calcNonce(group, uint32(offset), NONCES_PER_AES)
					if consume(nonce, index) {
						return
					}
				}
			}
		}
	}
	tick := time.Now().UnixMilli()
	for i, cipher := range p.groupCipher {
		calcGroup(i, cipher)
	}

	if time.Now().Unix()-showTime > 10 {
		showTime = time.Now().Unix()
		costTime := time.Now().UnixMilli() - tick
		fmt.Println("cost time:", costTime, " ms count:", count, " group cost:", groupCost/1e6)
	}
	return false
}

/*
// LSB part of the difficulty is checked with second sequence of AES ciphers.

	void check_lsb(__global struct AES_ctx* nonce_cipher, uint64_t difficulty_lsb, const __global uint8_t *label, uint32_t nonce, uint64_t nonce_offset, uint64_t base_index, __global struct Result* out_index, __global int *out_nonce_offset) {
	    uint64_t temp[2];
	    __global struct AES_ctx *lazy = nonce_cipher+nonce;
	    AES_CBC_encrypt_buffer(lazy->RoundKey, label, (uint8_t *)temp, AES128_BLOCKLEN);
	    uint8_t lsb = temp[0] & 0x00ffffffffffffff;
	    if (lsb < difficulty_lsb) {
	        uint64_t index = base_index + (nonce_offset / NONCES_PER_AES);
	        int nonce_index = atomic_add(out_nonce_offset, 1);

	        __global struct Result* out = out_index + nonce_index;
	        out->index = index;
	        out->nonce = nonce;
	    }
	}
*/
var temp [16]byte

func (p *Prover8_56) checkLSB(label []byte, nonce, offset uint32, baseIndex uint64, consume func(uint32, uint64) bool) bool {
	//var temp [16]byte

	lazy := p.nonceCipher[nonce%p.nonces]
	lazy.GoAes.Encrypt(temp[:], label)
	//lazy.Aes.EncryptUint(label, temp[:], 16)
	lsb := binary.LittleEndian.Uint64(temp[:]) & 0x00ffffffffffffff
	if lsb < p.DifficultyLSB {
		index := baseIndex + uint64(offset/uint32(NONCES_PER_AES))
		return consume(nonce, index)
	}
	return false
}

// prove 满足consume的调节后，返回true，读盘停止运行
//func (p *Prover8_56) prove(batch []byte, baseIndex uint64, consume func(uint32, uint64) bool) bool {
//	list, err := PostProve(p.Context, baseIndex/BLOCK_SIZE, batch)
//	if err != nil {
//		return false
//	}
//	for _, item := range list {
//		if !consume(item.Nonce, item.Index) {
//			return false
//		}
//	}
//	return true
//}
