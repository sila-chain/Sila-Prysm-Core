package validator

import (
	"context"
	"math/big"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/features"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/hash"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/rand"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/pkg/errors"
	fastssz "github.com/sila-chain/fastssz"
)

// silaexecDataMajorityVote determines the appropriate silaData for a block proposal using
// an algorithm called Voting with the Majority. The algorithm works as follows:
//   - Determine the timestamp for the start slot for the silaexec voting period.
//   - Determine the earliest and latest timestamps that a valid block can have.
//   - Determine the first block not before the earliest timestamp. This block is the lower bound.
//   - Determine the last block not after the latest timestamp. This block is the upper bound.
//   - If the last block is too early, use current silaData from the beacon state.
//   - Filter out votes on unknown blocks and blocks which are outside of the range determined by the lower and upper bounds.
//   - If no blocks are left after filtering votes, use silaData from the latest valid block.
//   - Otherwise:
//   - Determine the vote with the highest count. Prefer the vote with the highest silaexec block height in the event of a tie.
//   - This vote's block is the silaexec block to use for the block proposal.
//
// After Electra and silaexec deposit transition period voting will no longer be needed
func (vs *Server) silaexecDataMajorityVote(ctx context.Context, beaconState state.BeaconState) (*silapb.SilaData, error) {
	ctx, cancel := context.WithTimeout(ctx, silaDataTimeout)
	defer cancel()

	// post silaexec deposits, the Eth 1 data will then be frozen
	if helpers.DepositRequestsStarted(beaconState) {
		return beaconState.SilaData(), nil
	}

	slot := beaconState.Slot()
	votingPeriodStartTime := vs.slotStartTime(slot)

	if vs.MockSilaExecutionVotes {
		return vs.mockSilaDataVote(ctx, slot)
	}
	if !vs.SilaExecutionInfoFetcher.ExecutionClientConnected() {
		return vs.randomSilaDataVote(ctx)
	}
	silaexecDataNotification = false

	genesisTime, _ := vs.SilaExecutionInfoFetcher.GenesisExecutionChainInfo()
	followDistanceSeconds := params.BeaconConfig().SilaExecutionFollowDistance * params.BeaconConfig().SecondsPerSilaBlock
	latestValidTime := votingPeriodStartTime - followDistanceSeconds
	earliestValidTime := votingPeriodStartTime - 2*followDistanceSeconds

	// Special case for starting from a pre-mined genesis: the silaexec vote should be genesis until the chain has advanced
	// by SilaExecution_FOLLOW_DISTANCE. The head state should maintain the same SILASilaData until this condition has passed, so
	// trust the existing head for the right silaexec vote until we can get a meaningful value from the sila deposit.
	if latestValidTime < genesisTime+followDistanceSeconds {
		log.WithField("genesisTime", genesisTime).WithField("latestValidTime", latestValidTime).Warn("Voting period before genesis + follow distance, using silaData from head")
		return vs.HeadFetcher.HeadSilaData(), nil
	}

	lastBlockByLatestValidTime, err := vs.SilaBlockFetcher.BlockByTimestamp(ctx, latestValidTime)
	if err != nil {
		log.WithError(err).Error("Could not get last block by latest valid time")
		return vs.randomSilaDataVote(ctx)
	}
	if lastBlockByLatestValidTime.Time < earliestValidTime {
		return vs.HeadFetcher.HeadSilaData(), nil
	}

	lastBlockDepositCount, lastBlockDepositRoot := vs.DepositFetcher.DepositsNumberAndRootAtHeight(ctx, lastBlockByLatestValidTime.Number)
	if lastBlockDepositCount == 0 {
		return vs.ChainStartFetcher.ChainStartSilaData(), nil
	}

	if lastBlockDepositCount >= vs.HeadFetcher.HeadSilaData().DepositCount {
		h, err := vs.SilaBlockFetcher.BlockHashByHeight(ctx, lastBlockByLatestValidTime.Number)
		if err != nil {
			log.WithError(err).Error("Could not get hash of last block by latest valid time")
			return vs.randomSilaDataVote(ctx)
		}
		return &silapb.SilaData{
			BlockHash:    h.Bytes(),
			DepositCount: lastBlockDepositCount,
			DepositRoot:  lastBlockDepositRoot[:],
		}, nil
	}
	return vs.HeadFetcher.HeadSilaData(), nil
}

func (vs *Server) slotStartTime(slot primitives.Slot) uint64 {
	startTime, _ := vs.SilaExecutionInfoFetcher.GenesisExecutionChainInfo()
	return slots.VotingPeriodStartTime(startTime, slot)
}

// canonicalSilaData determines the canonical silaData and silaexec block height to use for determining deposits.
func (vs *Server) canonicalSilaData(
	ctx context.Context,
	beaconState state.BeaconState,
	currentVote *silapb.SilaData) (*silapb.SilaData, *big.Int, error) {
	var silaexecBlockHash [32]byte

	// Add in current vote, to get accurate vote tally
	if err := beaconState.AppendSilaDataVotes(currentVote); err != nil {
		return nil, nil, errors.Wrap(err, "could not append silaexec data votes to state")
	}
	hasSupport, err := blocks.SilaDataHasEnoughSupport(beaconState, currentVote)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not determine if current silaData vote has enough support")
	}
	var canonicalSilaData *silapb.SilaData
	if hasSupport {
		canonicalSilaData = currentVote
		silaexecBlockHash = bytesutil.ToBytes32(currentVote.BlockHash)
	} else {
		canonicalSilaData = beaconState.SilaData()
		silaexecBlockHash = bytesutil.ToBytes32(beaconState.SilaData().BlockHash)
	}
	if features.Get().DisableStakinContractCheck && silaexecBlockHash == [32]byte{} {
		return canonicalSilaData, new(big.Int).SetInt64(0), nil
	}
	_, canonicalSilaDataHeight, err := vs.SilaBlockFetcher.BlockExists(ctx, silaexecBlockHash)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not fetch silaData height")
	}
	return canonicalSilaData, canonicalSilaDataHeight, nil
}

func (vs *Server) mockSilaDataVote(ctx context.Context, slot primitives.Slot) (*silapb.SilaData, error) {
	if !silaexecDataNotification {
		log.Warn("Beacon Node is no longer connected to an SILAEXEC chain, so SILAEXEC data votes are now mocked.")
		silaexecDataNotification = true
	}
	// If a mock silaexec data votes is specified, we use the following for the
	// silaData we provide to every proposer based on https://github.com/sila/sila-pm/issues/62:
	//
	// slot_in_voting_period = current_slot % SLOTS_PER_SilaExecution_VOTING_PERIOD
	// SilaData(
	//   DepositRoot = hash(current_epoch + slot_in_voting_period),
	//   DepositCount = state.silaexec_deposit_index,
	//   BlockHash = hash(hash(current_epoch + slot_in_voting_period)),
	// )
	slotInVotingPeriod := slot.ModSlot(params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().EpochsPerSilaExecutionVotingPeriod)))
	headState, err := vs.HeadFetcher.HeadStateReadOnly(ctx)
	if err != nil {
		return nil, err
	}
	var enc []byte
	enc = fastssz.MarshalUint64(enc, uint64(slots.ToEpoch(slot))+uint64(slotInVotingPeriod))
	depRoot := hash.Hash(enc)
	blockHash := hash.Hash(depRoot[:])
	return &silapb.SilaData{
		DepositRoot:  depRoot[:],
		DepositCount: headState.SilaExecutionDepositIndex(),
		BlockHash:    blockHash[:],
	}, nil
}

func (vs *Server) randomSilaDataVote(ctx context.Context) (*silapb.SilaData, error) {
	if !silaexecDataNotification {
		log.Warn("Beacon Node is no longer connected to an SILAEXEC chain, so SILAEXEC data votes are now random.")
		silaexecDataNotification = true
	}
	headState, err := vs.HeadFetcher.HeadStateReadOnly(ctx)
	if err != nil {
		return nil, err
	}

	// set random roots and block hashes to prevent a majority from being
	// built if the silaexec node is offline
	randGen := rand.NewGenerator()
	depRoot := hash.Hash(bytesutil.Bytes32(randGen.Uint64()))
	blockHash := hash.Hash(bytesutil.Bytes32(randGen.Uint64()))
	return &silapb.SilaData{
		DepositRoot:  depRoot[:],
		DepositCount: headState.SilaExecutionDepositIndex(),
		BlockHash:    blockHash[:],
	}, nil
}
