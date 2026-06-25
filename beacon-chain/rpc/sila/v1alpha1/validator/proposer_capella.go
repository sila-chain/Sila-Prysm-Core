package validator

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
)

// Sets the bls to exec data for a block.
func (vs *Server) setBlsToExecData(blk interfaces.SignedBeaconBlock, headState state.BeaconState) {
	if blk.Version() < version.Capella {
		return
	}
	if err := blk.SetBLSToExecutionChanges([]*silapb.SignedBLSToExecutionChange{}); err != nil {
		log.WithError(err).Error("Could not set bls to execution data in block")
		return
	}
	changes, err := vs.BLSChangesPool.BLSToExecChangesForInclusion(headState)
	if err != nil {
		log.WithError(err).Error("Could not get bls to execution changes")
		return
	} else {
		if err := blk.SetBLSToExecutionChanges(changes); err != nil {
			log.WithError(err).Error("Could not set bls to execution changes")
			return
		}
	}
}
