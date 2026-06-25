package stateutil

import (
	"fmt"

	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/ssz"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// PTCWindowRoot computes the merkle root of the cached PTC window.
func PTCWindowRoot(ptcs []*silapb.PTCs) ([32]byte, error) {
	roots := make([][32]byte, len(ptcs))

	for i, ptc := range ptcs {
		if ptc == nil {
			return [32]byte{}, fmt.Errorf("invalid PTC at position %d", i)
		}

		r, err := ptc.HashTreeRoot()
		if err != nil {
			return [32]byte{}, err
		}

		roots[i] = r
	}

	return ssz.MerkleizeVector(roots, uint64(len(roots))), nil
}
