package gloas

import (
	"bytes"
	"context"

	requests "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/requests"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/pkg/errors"
)

// ProcessParentSilaPayload must run before process_block_header and
// process_sila_payload_bid, which overwrite the state fields it reads.
//
//	<spec fn="process_parent_sila_payload" fork="gloas" hash="defer_payload">
func ProcessParentSilaPayload(ctx context.Context, st state.BeaconState, blk interfaces.ReadOnlyBeaconBlock) error {
	_, span := trace.StartSpan(ctx, "gloas.ProcessParentSilaPayload")
	defer span.End()

	body := blk.Body()
	signedBid, err := body.SignedSilaPayloadBid()
	if err != nil {
		return errors.Wrap(err, "could not get signed sila payload bid")
	}
	bid := signedBid.Message

	parentBid, err := st.LatestSilaPayloadBid()
	if err != nil {
		return errors.Wrap(err, "could not get parent sila payload bid")
	}

	parentExecutionRequests, err := body.ParentExecutionRequests()
	if err != nil {
		return errors.Wrap(err, "could not get parent execution requests")
	}

	parentBidBlockHash := parentBid.BlockHash()
	isParentFull := bytes.Equal(bid.ParentBlockHash, parentBidBlockHash[:])

	if !isParentFull {
		if !IsEmptyExecutionRequests(parentExecutionRequests) {
			return errors.New("parent was empty but parent_execution_requests is non-empty")
		}
		return nil
	}

	requestsRoot, err := parentExecutionRequests.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "could not compute parent execution requests root")
	}
	parentBidRequestRoot := parentBid.ExecutionRequestsRoot()
	if requestsRoot != parentBidRequestRoot {
		return errors.Errorf("parent execution requests root mismatch: block=%#x, bid=%#x", requestsRoot, parentBidRequestRoot)
	}

	return ApplyParentSilaPayload(ctx, st, parentExecutionRequests)
}

// ApplyParentSilaPayload reads parent_bid from state.latest_sila_payload_bid
// and mutates st. Called by ProcessParentSilaPayload and by the validator during
// block production before computing withdrawals.
//
//	<spec fn="apply_parent_sila_payload" fork="gloas" hash="defer_payload">
func ApplyParentSilaPayload(
	ctx context.Context,
	st state.BeaconState,
	reqs *silaenginev1.ExecutionRequests,
) error {
	parentBid, err := st.LatestSilaPayloadBid()
	if err != nil {
		return errors.Wrap(err, "could not get latest sila payload bid")
	}
	parentSlot := parentBid.Slot()

	if err := processExecutionRequests(ctx, st, reqs); err != nil {
		return errors.Wrap(err, "could not process parent execution requests")
	}

	if err := st.QueueBuilderPaymentForSlot(parentSlot); err != nil {
		return errors.Wrap(err, "could not queue builder payment")
	}

	if err := st.SetSilaPayloadAvailability(parentSlot, true); err != nil {
		return errors.Wrap(err, "could not set parent sila payload availability")
	}

	blockHash := parentBid.BlockHash()
	if err := st.SetLatestBlockHash(blockHash); err != nil {
		return errors.Wrap(err, "could not set latest block hash")
	}

	return nil
}

func processExecutionRequests(ctx context.Context, st state.BeaconState, rqs *silaenginev1.ExecutionRequests) error {
	if err := ProcessDepositRequests(ctx, st, rqs.Deposits, prefetchedDepositSigs(rqs)); err != nil {
		return errors.Wrap(err, "could not process deposit requests")
	}
	var err error
	st, err = requests.ProcessWithdrawalRequests(ctx, st, rqs.Withdrawals)
	if err != nil {
		return errors.Wrap(err, "could not process withdrawal requests")
	}
	return requests.ProcessConsolidationRequests(ctx, st, rqs.Consolidations)
}

// IsEmptyExecutionRequests returns true if the execution requests contain no entries.
func IsEmptyExecutionRequests(r *silaenginev1.ExecutionRequests) bool {
	if r == nil {
		return true
	}
	return len(r.Deposits) == 0 && len(r.Withdrawals) == 0 && len(r.Consolidations) == 0
}
