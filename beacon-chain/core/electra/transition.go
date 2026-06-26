package electra

import (
	"context"
	"fmt"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/altair"
	e "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/epoch"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/epoch/precompute"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	"github.com/pkg/errors"
)

// Re-exports for methods that haven't changed in Electra.
var (
	InitializePrecomputeValidators       = altair.InitializePrecomputeValidators
	ProcessEpochParticipation            = altair.ProcessEpochParticipation
	ProcessInactivityScores              = altair.ProcessInactivityScores
	ProcessRewardsAndPenaltiesPrecompute = altair.ProcessRewardsAndPenaltiesPrecompute
	ProcessSlashings                     = e.ProcessSlashings
	ProcessSilaDataReset                 = e.ProcessSilaDataReset
	ProcessSlashingsReset                = e.ProcessSlashingsReset
	ProcessRandaoMixesReset              = e.ProcessRandaoMixesReset
	ProcessHistoricalDataUpdate          = e.ProcessHistoricalDataUpdate
	ProcessParticipationFlagUpdates      = altair.ProcessParticipationFlagUpdates
	ProcessSyncCommitteeUpdates          = altair.ProcessSyncCommitteeUpdates
	AttestationsDelta                    = altair.AttestationsDelta
)

// ProcessEpoch describes the per epoch operations that are performed on the beacon state.
// It's optimized by pre computing validator attested info and epoch total/attested balances upfront.
//
// Spec definition:
//
//	def process_epoch(state: BeaconState) -> None:
//	    process_justification_and_finalization(state)
//	    process_inactivity_updates(state)
//	    process_rewards_and_penalties(state)
//	    process_registry_updates(state)  # [Modified in Electra:SIP7251]
//	    process_slashings(state)  # [Modified in Electra:SIP7251]
//	    process_sila_data_reset(state)
//	    process_pending_deposits(state)  # [New in Electra:SIP7251]
//	    process_pending_consolidations(state)  # [New in Electra:SIP7251]
//	    process_effective_balance_updates(state)  # [Modified in Electra:SIP7251]
//	    process_slashings_reset(state)
//	    process_randao_mixes_reset(state)
//	    process_historical_summaries_update(state)
//	    process_participation_flag_updates(state)
//	    process_sync_committee_updates(state)
func ProcessEpoch(ctx context.Context, state state.BeaconState) error {
	ctx, span := trace.StartSpan(ctx, "electra.ProcessEpoch")
	defer span.End()

	if state == nil || state.IsNil() {
		return errors.New("nil state")
	}
	vp, bp, err := InitializePrecomputeValidators(ctx, state)
	if err != nil {
		return err
	}
	vp, bp, err = ProcessEpochParticipation(ctx, state, bp, vp)
	if err != nil {
		return err
	}
	state, err = precompute.ProcessJustificationAndFinalizationPreCompute(state, bp)
	if err != nil {
		return errors.Wrap(err, "could not process justification")
	}
	state, vp, err = ProcessInactivityScores(ctx, state, vp)
	if err != nil {
		return errors.Wrap(err, "could not process inactivity updates")
	}
	state, err = ProcessRewardsAndPenaltiesPrecompute(state, bp, vp)
	if err != nil {
		return errors.Wrap(err, "could not process rewards and penalties")
	}
	if err := ProcessRegistryUpdates(ctx, state); err != nil {
		return errors.Wrap(err, "could not process registry updates")
	}
	if err := ProcessSlashings(ctx, state); err != nil {
		return err
	}
	state, err = ProcessSilaDataReset(state)
	if err != nil {
		return err
	}
	if err = ProcessPendingDeposits(ctx, state, primitives.Gwei(bp.ActiveCurrentEpoch)); err != nil {
		return err
	}
	if err = ProcessPendingConsolidations(ctx, state); err != nil {
		return err
	}
	if err = ProcessEffectiveBalanceUpdates(state); err != nil {
		return err
	}
	state, err = ProcessSlashingsReset(state)
	if err != nil {
		return err
	}
	state, err = ProcessRandaoMixesReset(state)
	if err != nil {
		return err
	}
	state, err = ProcessHistoricalDataUpdate(state)
	if err != nil {
		return err
	}
	state, err = ProcessParticipationFlagUpdates(state)
	if err != nil {
		return err
	}
	_, err = ProcessSyncCommitteeUpdates(ctx, state)
	if err != nil {
		return err
	}
	return nil
}

// VerifyBlockDepositLength
//
// Spec definition:
//
//	# [Modified in Electra:SIP6110]
//	  # Disable former deposit mechanism once all prior deposits are processed
//	  silaexec_deposit_index_limit = min(state.sila_data.deposit_count, state.deposit_requests_start_index)
//	  if state.silaexec_deposit_index < silaexec_deposit_index_limit:
//	      assert len(body.deposits) == min(MAX_DEPOSITS, silaexec_deposit_index_limit - state.silaexec_deposit_index)
//	  else:
//	      assert len(body.deposits) == 0
func VerifyBlockDepositLength(body interfaces.ReadOnlyBeaconBlockBody, state state.BeaconState) error {
	silaexecData := state.SilaData()
	requestsStartIndex, err := state.DepositRequestsStartIndex()
	if err != nil {
		return errors.Wrap(err, "failed to get requests start index")
	}
	silaExecutionDepositIndexLimit := min(silaexecData.DepositCount, requestsStartIndex)
	if state.SilaExecutionDepositIndex() < silaExecutionDepositIndexLimit {
		if uint64(len(body.Deposits())) != min(params.BeaconConfig().MaxDeposits, silaExecutionDepositIndexLimit-state.SilaExecutionDepositIndex()) {
			return fmt.Errorf("incorrect outstanding deposits in block body, wanted: %d, got: %d", min(params.BeaconConfig().MaxDeposits, silaExecutionDepositIndexLimit-state.SilaExecutionDepositIndex()), len(body.Deposits()))
		}
	} else {
		if len(body.Deposits()) != 0 {
			return fmt.Errorf("incorrect outstanding deposits in block body, wanted: %d, got: %d", 0, len(body.Deposits()))
		}
	}
	return nil
}
