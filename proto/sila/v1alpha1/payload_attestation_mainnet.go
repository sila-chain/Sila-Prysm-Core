//go:build !minimal

package v1alpha1

import "github.com/sila-chain/go-bitfield"

func NewPayloadAttestationAggregationBits() bitfield.Bitvector512 {
	return bitfield.NewBitvector512()
}
