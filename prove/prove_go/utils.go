package post_go

import (
	"encoding/binary"
	"math"
	"math/bits"
	"strconv"
)

func calcNonceGroup(nonce, perAES uint32) uint32 {
	return nonce / perAES
}
func splitDifficulty(difficulty uint64) (uint8, uint64) {
	return uint8(difficulty >> 56), difficulty & 0x00ff_ffff_ffff_ffff
}

/*
#[inline(always)]

	fn nonce_group_range(nonces: Range<u32>, per_aes: u32) -> Range<u32> {
	    let start_group = nonces.start / per_aes;
	    let end_group = std::cmp::max(start_group + 1, (nonces.end + per_aes - 1) / per_aes);
	    start_group..end_group
	}
*/
func nonceGroupRange(nonces []uint32, perAES uint32) []uint32 {
	startGroup := nonces[0] / perAES
	endGroup := startGroup + 1
	if endGroup < (nonces[len(nonces)-1]+perAES-1)/perAES {
		endGroup = (nonces[len(nonces)-1] + perAES - 1) / perAES
	}
	var list []uint32
	for i := startGroup; i < endGroup; i++ {
		list = append(list, i)
	}
	return list
}

// nonceRange returns a list of nonces starting at start and ending at start+nonces
func nonceRange(start, nonces uint32) []uint32 {
	var list []uint32
	for i := start; i < start+nonces; i++ {
		list = append(list, i)
	}
	return list
}

/*
/// Calculate the number of bits required to store the value.

	pub(crate) fn required_bits(value: u64) -> usize {
	    if value == 0 {
	        return 0;
	    }
	    (value.ilog2() + 1) as usize
	}
*/
func requiredBits(value uint64) uint64 {
	if value == 0 {
		return 0
	}
	return uint64(math.Log2(float64(value)) + 1)
}

// CompressIndices Compress indexes into a byte slice.
/// The number of bits used to store each index is `keep_bits`.
/*
pub(crate) fn compress_indices(indexes: &[u64], keep_bits: usize) -> Vec<u8> {
    let mut bv = bitvec![u8, Lsb0;];
    for index in indexes {
        bv.extend_from_bitslice(&index.to_le_bytes().view_bits::<Lsb0>()[..keep_bits]);
    }
    bv.as_raw_slice().to_owned()
}
*/
//func CompressIndices(indexes []uint64, keepBits uint64) []byte {
//	bv := make([]byte, 0)
//	for _, index := range indexes {
//		bits := make([]byte, 8)
//		binary.LittleEndian.PutUint64(bits, index)
//		bv = append(bv, bits[:keepBits]...)
//	}
//	return bv
//}

// Compress indexes into a byte slice.
// The number of bits used to store each index is `keepBits`.
func CompressIndices(indexes []uint64, keepBits int) []byte {
	var bv string
	for _, index := range indexes {
		var b8 [8]byte
		binary.LittleEndian.PutUint64(b8[:], index)
		var bits string
		// b8转成二进制
		for i := 0; i < len(b8); i++ {
			tmp := ""

			bvBytes := strconv.FormatInt(int64(b8[i]), 2)
			// bvBytes 补齐8位
			for j := 0; j < 8-len(bvBytes); j++ {
				tmp += "0"
			}
			tmp += bvBytes
			//
			// tmp 反序
			for j := len(tmp) - 1; j >= 0; j-- {
				bits += string(tmp[j])
			}
		}
		// bits 保留 keepBits
		bv += bits[:keepBits]
		//fmt.Println("bv:", bv)
		//bv = append(bv, bits[:keepBits]...)
	}
	var result []byte
	for i := 0; i < len(bv); i += 8 {
		// 剩余长度判断是否满足8
		size := 8
		if len(bv)-i < size {
			size = len(bv) - i
		}
		bit := bv[i : i+size]
		// 反转bit
		var tmp string
		for j := len(bit) - 1; j >= 0; j-- {
			tmp += string(bit[j])
		}
		v, _ := strconv.ParseInt(tmp, 2, 64)

		result = append(result, byte(v))
	}

	return result
}

// Decompress indexes from a byte slice, previously compressed with `compressIndices`.
// Might return more indexes than the original, if the last byte contains unused bits.
func decompressIndexes(indexes []byte, bits int) []uint64 {
	result := []uint64{}
	for i := 0; i < len(indexes)*8/bits; i++ {
		index := uint64(0)
		for j := 0; j < bits/8; j++ {
			index |= uint64(indexes[i*bits/8+j]) << (8 * j)
		}
		result = append(result, index)
	}
	return result
}

// Calculate the number of bits required to store the value.
func requiredBits1(value uint64) int {
	if value == 0 {
		return 0
	}
	return bits.Len64(value)
}

/*
// Calculate nonce value given nonce group and its offset within the group.

	uint32_t calc_nonce(uint32_t nonce_group, uint32_t per_aes, uint32_t offset) {
	    return nonce_group * per_aes + offset % per_aes;
	}
*/
func calcNonce(nonceGroup, offset, perAES uint32) uint32 {
	return nonceGroup*perAES + offset%perAES
}
