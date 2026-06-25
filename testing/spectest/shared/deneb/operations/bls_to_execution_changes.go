package operations

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	common "github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/common/operations"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func blockWithBlsChange(ssz []byte) (interfaces.SignedBeaconBlock, error) {
	c := &silapb.SignedBLSToExecutionChange{}
	if err := c.UnmarshalSSZ(ssz); err != nil {
		return nil, err
	}
	b := util.NewBeaconBlockDeneb()
	b.Block.Body = &silapb.BeaconBlockBodyDeneb{BlsToExecutionChanges: []*silapb.SignedBLSToExecutionChange{c}}
	return blocks.NewSignedBeaconBlock(b)
}

func RunBLSToExecutionChangeTest(t *testing.T, config string) {
	common.RunBLSToExecutionChangeTest(t, config, version.String(version.Deneb), blockWithBlsChange, sszToState)
}
