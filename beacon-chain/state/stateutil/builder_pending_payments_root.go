package stateutil

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/ssz"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// BuilderPendingPaymentsRoot computes the merkle root of a slice of BuilderPendingPayment.
func BuilderPendingPaymentsRoot(slice []*silapb.BuilderPendingPayment) ([32]byte, error) {
	roots := make([][32]byte, len(slice))

	for i, payment := range slice {
		r, err := payment.HashTreeRoot()
		if err != nil {
			return [32]byte{}, err
		}

		roots[i] = r
	}

	return ssz.MerkleizeVector(roots, uint64(len(roots))), nil
}
