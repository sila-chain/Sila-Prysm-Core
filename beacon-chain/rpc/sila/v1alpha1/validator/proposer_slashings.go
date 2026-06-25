package validator

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	v "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/validators"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

func (vs *Server) getSlashings(ctx context.Context, head state.BeaconState) ([]*silapb.ProposerSlashing, []silapb.AttSlashing) {
	var err error
	proposerSlashings := vs.SlashingsPool.PendingProposerSlashings(ctx, head, false /*noLimit*/)
	attSlashings := vs.SlashingsPool.PendingAttesterSlashings(ctx, head, false /*noLimit*/)
	validProposerSlashings := make([]*silapb.ProposerSlashing, 0, len(proposerSlashings))
	validAttSlashings := make([]silapb.AttSlashing, 0, len(attSlashings))
	if len(proposerSlashings) == 0 && len(attSlashings) == 0 {
		return validProposerSlashings, validAttSlashings
	}
	// ExitInformation is expensive to compute, only do it if we need it.
	exitInfo := v.ExitInformation(head)
	if err := helpers.UpdateTotalActiveBalanceCache(head, exitInfo.TotalActiveBalance); err != nil {
		log.WithError(err).Warn("Could not update total active balance cache")
	}
	for _, slashing := range proposerSlashings {
		_, err = blocks.ProcessProposerSlashing(ctx, head, slashing, exitInfo)
		if err != nil {
			log.WithError(err).Warn("Could not validate proposer slashing for block inclusion")
			continue
		}
		validProposerSlashings = append(validProposerSlashings, slashing)
	}
	for _, slashing := range attSlashings {
		_, err = blocks.ProcessAttesterSlashing(ctx, head, slashing, exitInfo)
		if err != nil {
			log.WithError(err).Warn("Could not validate attester slashing for block inclusion")
			continue
		}
		validAttSlashings = append(validAttSlashings, slashing)
	}
	return validProposerSlashings, validAttSlashings
}
