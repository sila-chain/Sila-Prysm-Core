//go:build !minimal

package v1alpha1

import (
	"github.com/sila-chain/go-bitfield"
)

func NewSyncCommitteeAggregationBits() bitfield.Bitvector128 {
	return bitfield.NewBitvector128()
}

func ConvertToSyncContributionBitVector(b []byte) bitfield.Bitvector128 {
	return b
}
