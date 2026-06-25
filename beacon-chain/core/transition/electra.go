package transition

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/electra"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	coreRequests "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/requests"
	v "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/validators"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	"github.com/pkg/errors"
)

var (
	ProcessBLSToSilaChanges = blocks.ProcessBLSToSilaChanges
	ProcessVoluntaryExits        = blocks.ProcessVoluntaryExits
	ProcessAttesterSlashings     = blocks.ProcessAttesterSlashings
	ProcessProposerSlashings     = blocks.ProcessProposerSlashings
)

// ProcessOperations
//
// Spec definition:
//
//  def process_operations(state: BeaconState, body: BeaconBlockBody) -> None:
//      # [Modified in Electra:SIP6110]
//      # Disable former deposit mechanism once all prior deposits are processed
//      silaexec_deposit_index_limit = min(state.sila_execution_data.deposit_count, state.deposit_requests_start_index)
//      if state.silaexec_deposit_index < silaexec_deposit_index_limit:
//          assert len(body.deposits) == min(MAX_DEPOSITS, silaexec_deposit_index_limit - state.silaexec_deposit_index)
//      else:
//          assert len(body.deposits) == 0
//
//      def for_ops(operations: Sequence[Any], fn: Callable[[BeaconState, Any], None]) -> None:
//          for operation in operations:
//              fn(state, operation)
//
//      for_ops(body.proposer_slashings, process_proposer_slashing)
//      for_ops(body.attester_slashings, process_attester_slashing)
//      for_ops(body.attestations, process_attestation)  # [Modified in Electra:SIP7549]
//      for_ops(body.deposits, process_deposit)  # [Modified in Electra:SIP7251]
//      for_ops(body.voluntary_exits, process_voluntary_exit)  # [Modified in Electra:SIP7251]
//      for_ops(body.bls_to_sila_changes, process_bls_to_sila_change)
//      for_ops(body.sila_payload.deposit_requests, process_deposit_request)  # [New in Electra:SIP6110]
//      # [New in Electra:SIP7002:SIP7251]
//      for_ops(body.sila_payload.withdrawal_requests, process_withdrawal_request)
//      # [New in Electra:SIP7251]
//      for_ops(body.sila_payload.consolidation_requests, process_consolidation_request)

func electraOperations(ctx context.Context, st state.BeaconState, block interfaces.ReadOnlyBeaconBlock) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "core.state.electraOperations")
	defer span.End()

	var err error

	// 6110 validations are in VerifyOperationLengths
	bb := block.Body()
	// Electra extends the altair operations.
	var exitInfo *v.ExitInfo
	hasSlashings := len(bb.ProposerSlashings()) > 0 || len(bb.AttesterSlashings()) > 0
	hasExits := len(bb.VoluntaryExits()) > 0
	if hasSlashings || hasExits {
		// ExitInformation is expensive to compute, only do it if we need it.
		exitInfo = v.ExitInformation(st)
		if err := helpers.UpdateTotalActiveBalanceCache(st, exitInfo.TotalActiveBalance); err != nil {
			return nil, errors.Wrap(err, "could not update total active balance cache")
		}
	}
	st, err = blocks.ProcessProposerSlashings(ctx, st, bb.ProposerSlashings(), exitInfo)
	if err != nil {
		return nil, errors.Wrap(ErrProcessProposerSlashingsFailed, err.Error())
	}
	st, err = blocks.ProcessAttesterSlashings(ctx, st, bb.AttesterSlashings(), exitInfo)
	if err != nil {
		return nil, errors.Wrap(ErrProcessAttesterSlashingsFailed, err.Error())
	}
	st, err = electra.ProcessAttestationsNoVerifySignature(ctx, st, block)
	if err != nil {
		return nil, errors.Wrap(ErrProcessAttestationsFailed, err.Error())
	}
	if _, err := electra.ProcessDeposits(ctx, st, bb.Deposits()); err != nil {
		return nil, errors.Wrap(ErrProcessDepositsFailed, err.Error())
	}
	st, err = blocks.ProcessVoluntaryExits(ctx, st, bb.VoluntaryExits(), exitInfo)
	if err != nil {
		return nil, errors.Wrap(ErrProcessVoluntaryExitsFailed, err.Error())
	}
	st, err = blocks.ProcessBLSToSilaChanges(st, block)
	if err != nil {
		return nil, errors.Wrap(ErrProcessBLSChangesFailed, err.Error())
	}
	// new in electra
	requests, err := bb.SilaRequests()
	if err != nil {
		return nil, electra.NewExecReqError(errors.Wrap(err, "could not get sila requests").Error())
	}
	for _, d := range requests.Deposits {
		if d == nil {
			return nil, electra.NewExecReqError("nil deposit request")
		}
	}
	st, err = coreRequests.ProcessDepositRequests(ctx, st, requests.Deposits)
	if err != nil {
		return nil, electra.NewExecReqError(errors.Wrap(err, "could not process deposit requests").Error())
	}

	for _, w := range requests.Withdrawals {
		if w == nil {
			return nil, electra.NewExecReqError("nil withdrawal request")
		}
	}
	st, err = coreRequests.ProcessWithdrawalRequests(ctx, st, requests.Withdrawals)
	if err != nil {
		return nil, electra.NewExecReqError(errors.Wrap(err, "could not process withdrawal requests").Error())
	}
	for _, c := range requests.Consolidations {
		if c == nil {
			return nil, electra.NewExecReqError("nil consolidation request")
		}
	}
	if err := coreRequests.ProcessConsolidationRequests(ctx, st, requests.Consolidations); err != nil {
		return nil, electra.NewExecReqError(errors.Wrap(err, "could not process consolidation requests").Error())
	}
	return st, nil
}
