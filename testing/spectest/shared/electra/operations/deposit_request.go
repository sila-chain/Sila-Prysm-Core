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

func blockWithDepositRequest(ssz []byte) (interfaces.SignedBeaconBlock, error) {
	dr := &enginev1.DepositRequest{}
	if err := dr.UnmarshalSSZ(ssz); err != nil {
		return nil, err
	}
	er := &enginev1.ExecutionRequests{
		Deposits: []*enginev1.DepositRequest{dr},
	}
	b := util.NewBeaconBlockElectra()
	b.Block.Body = &silapb.BeaconBlockBodyElectra{ExecutionRequests: er}
	return blocks.NewSignedBeaconBlock(b)
}

func RunDepositRequestsTest(t *testing.T, config string) {
	common.RunDepositRequestsTest(t, config, version.String(version.Electra), blockWithDepositRequest, sszToState)
}
