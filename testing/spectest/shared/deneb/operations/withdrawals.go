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

func blockWithWithdrawals(ssz []byte) (interfaces.SignedBeaconBlock, error) {
	e := &silaenginev1.SilaPayloadDeneb{}
	if err := e.UnmarshalSSZ(ssz); err != nil {
		return nil, err
	}
	b := util.NewBeaconBlockDeneb()
	b.Block.Body = &silapb.BeaconBlockBodyDeneb{SilaPayload: e}
	return blocks.NewSignedBeaconBlock(b)
}

func RunWithdrawalsTest(t *testing.T, config string) {
	common.RunWithdrawalsTest(t, config, version.String(version.Deneb), blockWithWithdrawals, sszToState)
}
