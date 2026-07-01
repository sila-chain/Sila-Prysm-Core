//go:build minimal

package v1alpha1

import (
	"github.com/sila-chain/go-bitfield"
)

func NewSyncCommitteeAggregationBits() bitfield.Bitvector8 {
	return bitfield.NewBitvector8()
}

func ConvertToSyncContributionBitVector(b []byte) bitfield.Bitvector8 {
	return b
}
