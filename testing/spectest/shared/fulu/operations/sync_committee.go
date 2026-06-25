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

func blockWithSyncAggregate(ssz []byte) (interfaces.SignedBeaconBlock, error) {
	sa := &silapb.SyncAggregate{}
	if err := sa.UnmarshalSSZ(ssz); err != nil {
		return nil, err
	}
	b := util.NewBeaconBlockElectra()
	b.Block.Body = &silapb.BeaconBlockBodyElectra{SyncAggregate: sa}
	return blocks.NewSignedBeaconBlock(b)
}

func RunSyncCommitteeTest(t *testing.T, config string) {
	common.RunSyncCommitteeTest(t, config, version.String(version.Fulu), blockWithSyncAggregate, sszToState)
}
