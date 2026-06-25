package operations

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	common "github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/common/operations"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func blockWithConsolidation(ssz []byte) (interfaces.SignedBeaconBlock, error) {
	cr := &silaenginev1.ConsolidationRequest{}
	if err := cr.UnmarshalSSZ(ssz); err != nil {
		return nil, err
	}
	er := &silaenginev1.ExecutionRequests{
		Consolidations: []*silaenginev1.ConsolidationRequest{cr},
	}
	b := util.NewBeaconBlockElectra()
	b.Block.Body = &silapb.BeaconBlockBodyElectra{ExecutionRequests: er}
	return blocks.NewSignedBeaconBlock(b)
}

func RunConsolidationTest(t *testing.T, config string) {
	common.RunConsolidationTest(t, config, version.String(version.Gloas), blockWithConsolidation, sszToState)
}
