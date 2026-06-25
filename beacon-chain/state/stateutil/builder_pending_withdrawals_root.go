package stateutil

import (
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/ssz"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// BuilderPendingWithdrawalsRoot computes the SSZ root of a slice of BuilderPendingWithdrawal.
func BuilderPendingWithdrawalsRoot(slice []*silapb.BuilderPendingWithdrawal) ([32]byte, error) {
	return ssz.SliceRoot(slice, fieldparams.BuilderPendingWithdrawalsLimit)
}
