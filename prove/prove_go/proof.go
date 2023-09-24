package post_go

import (
	"context"
	"errors"
	"fmt"
	"github.com/trying2016/post-go/shared"
	"sort"
)

func GenerateProof(dataDir string, challenge []byte, nonces, K1, K2 uint32, powDifficulty []byte) (*shared.Proof, error) {
	metadata, err := shared.ReadMetadata(dataDir)
	if err != nil {
		return nil, fmt.Errorf("loading metadata: %w", err)
	}
	params, err := NewProvingParams(metadata, &Config{
		PowDifficulty: powDifficulty,
		K1:            K1,
		K2:            K2,
	})
	if err != nil {
		return nil, fmt.Errorf("creating proving params: %w", err)
	}
	// let num_labels = metadata.num_units as u64 * metadata.labels_per_unit
	numLabels := uint64(metadata.NumUnits) * metadata.LabelsPerUnit
	fmt.Printf("Generating proof with params: %+v\n", params)
	startNonce := uint32(0)

	// 进行扫盘
	generate := func(startNonce uint32) (*shared.Proof, error) {
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
		}()

		indexes := make(map[uint32][]uint64)
		prove, err := NewProver8_56(challenge, nonceRange(startNonce, uint32(nonces)), params, metadata.NodeId)
		if err != nil {
			return nil, err
		}
		defer func() {
			prove.Destroy()
		}()
		var foundNonce int64 = -1

		//proveQueue := queue.NewNormal(ctx, 4, func(v interface{}) {
		//	if foundNonce != -1 {
		//		return
		//	}
		//	select {
		//	case <-ctx.Done():
		//		return
		//	default:
		//	}
		//	batch := v.(*Batch)
		//
		//	prove.prove(batch.Data, batch.Pos, func(nonce uint32, index uint64) bool {
		//		indexes[nonce] = append(indexes[nonce], index)
		//		if len(indexes[nonce]) >= int(K2) {
		//			foundNonce = int64(nonce)
		//			cancel()
		//			return true
		//		}
		//		return true
		//	})
		//
		//})

		err = ReadData(dataDir, BUNCH_SIZE, metadata.MaxFileSize, func(batch *Batch) bool {
			select {
			case <-ctx.Done():
				return false
			default:
			}
			prove.prove(batch.Data, batch.Pos/LABEL_SIZE, func(nonce uint32, index uint64) bool {
				indexes[nonce] = append(indexes[nonce], index)
				if len(indexes[nonce]) >= int(K2) {
					foundNonce = int64(nonce)
					cancel()
					return true
				}
				return false
			})
			//proveQueue.Push(batch)
			return true
		})

		if err != nil {
			return nil, err
		}

		//for _, v := range indexes {
		//	sort.Slice(v, func(i, j int) bool {
		//		return v[i] < v[j]
		//	})
		//}

		if foundNonce != -1 {
			list := indexes[uint32(foundNonce)]
			sort.Slice(list, func(i, j int) bool {
				return list[i] < list[j]
			})
			fmt.Print("with [")
			for _, v := range list {
				fmt.Print(", ", v)
			}
			fmt.Println("]")
			return &shared.Proof{
				Pow:     prove.Pow(uint32(foundNonce) - startNonce),
				Indices: CompressIndices(list, int(requiredBits(numLabels))),
				Nonce:   uint32(foundNonce),
			}, nil
		} else {
			return nil, errors.New("not found")
		}
	}

	for {
		if proof, err := generate(startNonce); err == nil {
			return proof, nil
		}
		startNonce += uint32(nonces)
	}
}
