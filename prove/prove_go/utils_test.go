package post_go

import (
	"fmt"
	"testing"
)

func TestNonceGroupRange(t *testing.T) {
	nonces := []uint32{0, 1, 2, 3, 4, 5, 6, 7}
	perAES := uint32(2)
	expected := []uint32{0, 1, 2, 3}
	actual := nonceGroupRange(nonces, perAES)
	if len(expected) != len(actual) {
		t.Errorf("expected %v, got %v", expected, actual)
	}
	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("expected %v, got %v", expected, actual)
		}
	}
}

func TestCompressIndices(t *testing.T) {
	indexes := []uint64{0, 0b1111_1111_1111_0101, 0, 0b1111_1111_0000_1111}
	compressed := CompressIndices(indexes, 35)
	fmt.Println(compressed)
	// [0, 0, 0, 0, 168, 255, 7, 0, 0, 0, 0, 0, 0, 30, 254, 1, 0, 0]
	compressed = CompressIndices(indexes, 16)
	fmt.Println(compressed)
}

func TestRequiredBits(t *testing.T) {
	t.Log("requiredBits:", requiredBits(4294967296*4), requiredBits1(4294967296*4))
}
