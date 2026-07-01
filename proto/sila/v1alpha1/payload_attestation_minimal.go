//go:build minimal

package v1alpha1

import "github.com/sila-chain/go-bitfield"

func NewPayloadAttestationAggregationBits() bitfield.Bitvector16 {
	return bitfield.NewBitvector16()
}
