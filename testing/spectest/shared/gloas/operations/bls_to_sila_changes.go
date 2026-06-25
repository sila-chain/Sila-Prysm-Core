package operations

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	common "github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/common/operations"
)

func blockWithBlsChange(ssz []byte) (interfaces.SignedBeaconBlock, error) {
	c := &silapb.SignedBLSToSilaChange{}
	if err := c.UnmarshalSSZ(ssz); err != nil {
		return nil, err
	}
	b := &silapb.SignedBeaconBlockGloas{
		Block: &silapb.BeaconBlockGloas{
			Body: &silapb.BeaconBlockBodyGloas{BlsToSilaChanges: []*silapb.SignedBLSToSilaChange{c}},
		},
	}
	return blocks.NewSignedBeaconBlock(b)
}

func RunBLSToSilaChangeTest(t *testing.T, config string) {
	common.RunBLSToSilaChangeTest(t, config, version.String(version.Gloas), blockWithBlsChange, sszToState)
}
