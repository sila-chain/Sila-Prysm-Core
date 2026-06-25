package operations

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	common "github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/common/operations"
)

func blockWithSignedExecutionPayloadBid(blockSSZ []byte) (interfaces.SignedBeaconBlock, error) {
	var block silapb.BeaconBlockGloas
	if err := block.UnmarshalSSZ(blockSSZ); err != nil {
		return nil, err
	}
	return blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockGloas{Block: &block})
}

func RunExecutionPayloadBidTest(t *testing.T, config string) {
	common.RunExecutionPayloadBidTest(t, config, version.String(version.Gloas), blockWithSignedExecutionPayloadBid, sszToState)
}
