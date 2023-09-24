package shared

import "github.com/trying2016/post-go/codec"

type Proof struct {
	Nonce   uint32
	Indices []byte
	Pow     uint64
}

func EncodeProof(proof *Proof) ([]byte, error) {
	return codec.Encode(proof)
}

func DecodeProof(data []byte) (*Proof, error) {
	var proof Proof
	err := codec.Decode(data, &proof)
	return &proof, err
}

// EncodeScale implements scale codec interface.
func (p *Proof) EncodeScale(enc *codec.Encoder) (total int, err error) {
	{
		n, err := codec.EncodeCompact32(enc, uint32(p.Nonce))
		if err != nil {
			return total, err
		}
		total += n
	}
	{
		n, err := codec.EncodeByteSliceWithLimit(enc, p.Indices, 8000) // needs to hold K2*8 bytes at most
		if err != nil {
			return total, err
		}
		total += n
	}
	{
		n, err := codec.EncodeCompact64(enc, p.Pow)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// DecodeScale implements scale codec interface.
func (p *Proof) DecodeScale(dec *codec.Decoder) (total int, err error) {
	{
		field, n, err := codec.DecodeCompact32(dec)
		if err != nil {
			return total, err
		}
		total += n
		p.Nonce = field
	}
	{
		field, n, err := codec.DecodeByteSliceWithLimit(dec, 8000) // needs to hold K2*8 bytes at most
		if err != nil {
			return total, err
		}
		total += n
		p.Indices = field
	}
	{
		field, n, err := codec.DecodeCompact64(dec)
		if err != nil {
			return total, err
		}
		total += n
		p.Pow = field
	}
	return total, nil
}

type ProofMetadata struct {
	NodeId          []byte
	CommitmentAtxId []byte

	Challenge     Challenge
	NumUnits      uint32
	LabelsPerUnit uint64
}

type VRFNonce uint64

type VRFNonceMetadata struct {
	NodeId          []byte
	CommitmentAtxId []byte

	NumUnits      uint32
	LabelsPerUnit uint64
}
