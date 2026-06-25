package operations

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/altair"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	common "github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/common/operations"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func blockWithDeposit(ssz []byte) (interfaces.SignedBeaconBlock, error) {
	d := &silapb.Deposit{}
	if err := d.UnmarshalSSZ(ssz); err != nil {
		return nil, err
	}
	b := util.NewBeaconBlock()
	b.Block.Body = &silapb.BeaconBlockBody{Deposits: []*silapb.Deposit{d}}
	return blocks.NewSignedBeaconBlock(b)
}

func RunDepositTest(t *testing.T, config string) {
	common.RunDepositTest(t, config, version.String(version.Phase0), blockWithDeposit, altair.ProcessDeposits, sszToState)
}
