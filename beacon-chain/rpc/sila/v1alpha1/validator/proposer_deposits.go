package validator

import (
	"bytes"
	"context"
	"math/big"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/container/trie"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (vs *Server) packDepositsAndAttestations(
	ctx context.Context,
	head state.BeaconState,
	blkSlot primitives.Slot,
	silaexecData *silapb.SilaExecutionData,
) ([]*silapb.Deposit, []silapb.Att, error) {
	eg, egctx := errgroup.WithContext(ctx)
	var deposits []*silapb.Deposit
	var atts []silapb.Att

	eg.Go(func() error {
		// Pack SILAEXEC deposits which have not been included in the beacon chain.
		localDeposits, err := vs.deposits(egctx, head, silaexecData)
		if err != nil {
			return status.Errorf(codes.Internal, "Could not get SILAEXEC deposits: %v", err)
		}
		// if the original context is cancelled, then cancel this routine too
		select {
		case <-egctx.Done():
			return egctx.Err()
		default:
		}
		deposits = localDeposits
		return nil
	})

	eg.Go(func() error {
		// Pack aggregated attestations which have not been included in the beacon chain.
		localAtts, err := vs.packAttestations(egctx, head, blkSlot)
		if err != nil {
			return status.Errorf(codes.Internal, "Could not get attestations to pack into block: %v", err)
		}
		// if the original context is cancelled, then cancel this routine too
		select {
		case <-egctx.Done():
			return egctx.Err()
		default:
		}
		atts = localAtts
		return nil
	})

	return deposits, atts, eg.Wait()
}

// deposits returns a list of pending deposits that are ready for inclusion in the next beacon
// block. Determining deposits depends on the current silaExecutionData vote for the block and whether or not
// this silaExecutionData has enough support to be considered for deposits inclusion. If current vote has
// enough support, then use that vote for basis of determining deposits, otherwise use current state
// silaExecutionData.
// In the post-electra phase, this function will usually return an empty list,
// as the legacy deposit process is deprecated. (SIP-6110)
// NOTE: During the transition period, the legacy deposit process
// may still be active and managed. This function handles that scenario.
func (vs *Server) deposits(
	ctx context.Context,
	beaconState state.BeaconState,
	currentVote *silapb.SilaExecutionData,
) ([]*silapb.Deposit, error) {
	ctx, span := trace.StartSpan(ctx, "ProposerServer.deposits")
	defer span.End()

	if vs.MockSilaExecutionVotes {
		return []*silapb.Deposit{}, nil
	}

	if !vs.SilaExecutionInfoFetcher.ExecutionClientConnected() {
		log.Warn("Not connected to silaexec node, skip pending deposit insertion")
		return []*silapb.Deposit{}, nil
	}

	// skip legacy deposits if silaexec deposit index is already at the index of deposit requests start
	if helpers.DepositRequestsStarted(beaconState) {
		return []*silapb.Deposit{}, nil
	}

	// Need to fetch if the deposits up to the state's latest silaexec data matches
	// the number of all deposits in this RPC call. If not, then we return nil.
	canonicalSilaExecutionData, canonicalSilaExecutionDataHeight, err := vs.canonicalSilaExecutionData(ctx, beaconState, currentVote)
	if err != nil {
		return nil, err
	}

	_, genesisSilaBlock := vs.SilaExecutionInfoFetcher.GenesisExecutionChainInfo()
	if genesisSilaBlock.Cmp(canonicalSilaExecutionDataHeight) == 0 {
		return []*silapb.Deposit{}, nil
	}

	// If there are no pending deposits, exit early.
	allPendingContainers := vs.PendingDepositsFetcher.PendingContainers(ctx, canonicalSilaExecutionDataHeight)
	if len(allPendingContainers) == 0 {
		log.Debug("No pending deposits for inclusion in block")
		return []*silapb.Deposit{}, nil
	}

	depositTrie, err := vs.depositTrie(ctx, canonicalSilaExecutionData, canonicalSilaExecutionDataHeight)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve deposit trie")
	}

	// Deposits need to be received in order of merkle index root, so this has to make sure
	// deposits are sorted from lowest to highest.
	var pendingDeps []*silapb.DepositContainer
	for _, dep := range allPendingContainers {
		if beaconState.Version() < version.Electra {
			// Add deposits up to min(MAX_DEPOSITS, sila_execution_data.deposit_count - state.silaexec_deposit_index)
			if uint64(dep.Index) >= beaconState.SilaExecutionDepositIndex() && uint64(dep.Index) < canonicalSilaExecutionData.DepositCount {
				pendingDeps = append(pendingDeps, dep)
			}
		} else {
			// Electra change SIP6110
			// def get_silaexec_pending_deposit_count(state: BeaconState) -> uint64:
			//    silaexec_deposit_index_limit = min(state.sila_execution_data.deposit_count, state.deposit_requests_start_index)
			//    if state.silaexec_deposit_index < silaexec_deposit_index_limit:
			//        return min(MAX_DEPOSITS, silaexec_deposit_index_limit - state.silaexec_deposit_index)
			//    else:
			//        return uint64(0)
			requestsStartIndex, err := beaconState.DepositRequestsStartIndex()
			if err != nil {
				return nil, errors.Wrap(err, "could not retrieve requests start index")
			}
			silaExecutionDepositIndexLimit := min(canonicalSilaExecutionData.DepositCount, requestsStartIndex)
			if beaconState.SilaExecutionDepositIndex() < silaExecutionDepositIndexLimit {
				if uint64(dep.Index) >= beaconState.SilaExecutionDepositIndex() && uint64(dep.Index) < silaExecutionDepositIndexLimit {
					pendingDeps = append(pendingDeps, dep)
				}
			}
			// just don't add any pending deps if it's not state.silaexec_deposit_index < silaexec_deposit_index_limit
		}

		// Don't try to pack more than the max allowed in a block
		if uint64(len(pendingDeps)) == params.BeaconConfig().MaxDeposits {
			break
		}
	}

	for i := range pendingDeps {
		pendingDeps[i].Deposit, err = constructMerkleProof(depositTrie, int(pendingDeps[i].Index), pendingDeps[i].Deposit)
		if err != nil {
			return nil, err
		}
	}

	var pendingDeposits []*silapb.Deposit
	for i := uint64(0); i < uint64(len(pendingDeps)); i++ {
		pendingDeposits = append(pendingDeposits, pendingDeps[i].Deposit)
	}
	return pendingDeposits, nil
}

func (vs *Server) depositTrie(ctx context.Context, canonicalSilaExecutionData *silapb.SilaExecutionData, canonicalSilaExecutionDataHeight *big.Int) (cache.MerkleTree, error) {
	ctx, span := trace.StartSpan(ctx, "ProposerServer.depositTrie")
	defer span.End()

	var depositTrie cache.MerkleTree

	finalizedDeposits, err := vs.DepositFetcher.FinalizedDeposits(ctx)
	if err != nil {
		return nil, err
	}
	depositTrie = finalizedDeposits.Deposits()
	upToSilaExecutionDataDeposits := vs.DepositFetcher.NonFinalizedDeposits(ctx, finalizedDeposits.MerkleTrieIndex(), canonicalSilaExecutionDataHeight)
	insertIndex := finalizedDeposits.MerkleTrieIndex() + 1

	if shouldRebuildTrie(canonicalSilaExecutionData.DepositCount, uint64(len(upToSilaExecutionDataDeposits))) {
		log.WithFields(logrus.Fields{
			"unfinalizedDeposits": len(upToSilaExecutionDataDeposits),
			"totalDepositCount":   canonicalSilaExecutionData.DepositCount,
		}).Warn("Too many unfinalized deposits, building a deposit trie from scratch.")
		return vs.rebuildDepositTrie(ctx, canonicalSilaExecutionData, canonicalSilaExecutionDataHeight)
	}
	for _, dep := range upToSilaExecutionDataDeposits {
		depHash, err := dep.Data.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "could not hash deposit data")
		}
		if err = depositTrie.Insert(depHash[:], int(insertIndex)); err != nil {
			return nil, err
		}
		insertIndex++
	}
	valid, err := validateDepositTrie(depositTrie, canonicalSilaExecutionData)
	// Log a warning here, as the cached trie is invalid.
	if !valid {
		log.WithError(err).Warn("Cached deposit trie is invalid, rebuilding it now")
		return vs.rebuildDepositTrie(ctx, canonicalSilaExecutionData, canonicalSilaExecutionDataHeight)
	}

	return depositTrie, nil
}

// rebuilds our deposit trie by recreating it from all processed deposits till
// specified silaexec block height.
func (vs *Server) rebuildDepositTrie(ctx context.Context, canonicalSilaExecutionData *silapb.SilaExecutionData, canonicalSilaExecutionDataHeight *big.Int) (cache.MerkleTree, error) {
	ctx, span := trace.StartSpan(ctx, "ProposerServer.rebuildDepositTrie")
	defer span.End()

	deposits := vs.DepositFetcher.AllDeposits(ctx, canonicalSilaExecutionDataHeight)
	trieItems := make([][]byte, 0, len(deposits))
	for _, dep := range deposits {
		depHash, err := dep.Data.HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "could not hash deposit data")
		}
		trieItems = append(trieItems, depHash[:])
	}
	depositTrie, err := trie.GenerateTrieFromItems(trieItems, params.BeaconConfig().SilaDepositTreeDepth)
	if err != nil {
		return nil, err
	}

	valid, err := validateDepositTrie(depositTrie, canonicalSilaExecutionData)
	// Log an error here, as even with rebuilding the trie, it is still invalid.
	if !valid {
		log.WithError(err).Error("Rebuilt deposit trie is invalid")
	}
	return depositTrie, nil
}

// validate that the provided deposit trie matches up with the canonical silaexec data provided.
func validateDepositTrie(trie cache.MerkleTree, canonicalSilaExecutionData *silapb.SilaExecutionData) (bool, error) {
	if trie == nil || canonicalSilaExecutionData == nil {
		return false, errors.New("nil trie or silaExecutionData provided")
	}
	if trie.NumOfItems() != int(canonicalSilaExecutionData.DepositCount) {
		return false, errors.Errorf("wanted the canonical count of %d but received %d", canonicalSilaExecutionData.DepositCount, trie.NumOfItems())
	}
	rt, err := trie.HashTreeRoot()
	if err != nil {
		return false, err
	}
	if !bytes.Equal(rt[:], canonicalSilaExecutionData.DepositRoot) {
		return false, errors.Errorf("wanted the canonical deposit root of %#x but received %#x", canonicalSilaExecutionData.DepositRoot, rt)
	}
	return true, nil
}

func constructMerkleProof(trie cache.MerkleTree, index int, deposit *silapb.Deposit) (*silapb.Deposit, error) {
	proof, err := trie.MerkleProof(index)
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate merkle proof for deposit at index %d", index)
	}
	// For every deposit, we construct a Merkle proof using the powchain service's
	// in-memory deposits trie, which is updated only once the state's LatestSilaExecutionData
	// property changes during a state transition after a voting period.
	deposit.Proof = proof
	return deposit, nil
}

// This checks whether we should fallback to rebuild the whole deposit trie.
func shouldRebuildTrie(totalDepCount, unFinalizedDeps uint64) bool {
	if totalDepCount == 0 || unFinalizedDeps == 0 {
		return false
	}
	// The total number interior nodes hashed in a binary trie would be
	// x - 1, where x is the total number of leaves of the trie. For simplicity's
	// sake we assume it as x here as this function is meant as a heuristic rather than
	// and exact calculation.
	//
	// Since the effective_depth = log(x) , the total depth can be represented as
	// depth = log(x) + k. We can then find the total number of nodes to be hashed by
	// calculating  y (log(x) + k) , where y is the number of unfinalized deposits. For
	// the deposit trie, the value of log(x) + k is fixed at 32.
	unFinalizedCompute := unFinalizedDeps * params.BeaconConfig().SilaDepositTreeDepth
	return unFinalizedCompute > totalDepCount
}
