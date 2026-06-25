package operations

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	enginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/engine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	common "github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/common/operations"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func blockWithWithdrawalRequest(ssz []byte) (interfaces.SignedBeaconBlock, error) {
	wr := &enginev1.WithdrawalRequest{}
	if err := wr.UnmarshalSSZ(ssz); err != nil {
		return nil, err
	}
	er := &enginev1.ExecutionRequests{
		Withdrawals: []*enginev1.WithdrawalRequest{wr},
	}
	b := util.NewBeaconBlockElectra()
	b.Block.Body = &silapb.BeaconBlockBodyElectra{ExecutionRequests: er}
	return blocks.NewSignedBeaconBlock(b)
}

func RunWithdrawalRequestTest(t *testing.T, config string) {
	common.RunWithdrawalRequestTest(t, config, version.String(version.Fulu), blockWithWithdrawalRequest, sszToState)
}
