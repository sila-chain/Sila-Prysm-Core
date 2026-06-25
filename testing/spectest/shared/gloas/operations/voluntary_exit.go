package operations

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	common "github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/common/operations"
)

func blockWithVoluntaryExit(ssz []byte) (interfaces.SignedBeaconBlock, error) {
	e := &silapb.SignedVoluntaryExit{}
	if err := e.UnmarshalSSZ(ssz); err != nil {
		return nil, err
	}
	b := &silapb.SignedBeaconBlockGloas{
		Block: &silapb.BeaconBlockGloas{
			Body: &silapb.BeaconBlockBodyGloas{VoluntaryExits: []*silapb.SignedVoluntaryExit{e}},
		},
	}
	return blocks.NewSignedBeaconBlock(b)
}

func RunVoluntaryExitTest(t *testing.T, config string) {
	common.RunVoluntaryExitTest(t, config, version.String(version.Gloas), blockWithVoluntaryExit, sszToState)
}
