package state_native

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native/types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stateutil"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/ssz"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/pkg/errors"
)

// ComputeFieldRootsWithHasher hashes the provided state and returns its respective field roots.
func ComputeFieldRootsWithHasher(ctx context.Context, state *BeaconState) ([][]byte, error) {
	ctx, span := trace.StartSpan(ctx, "ComputeFieldRootsWithHasher")
	defer span.End()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if state == nil {
		return nil, errors.New("nil state")
	}
	var fieldRoots [][]byte
	switch state.version {
	case version.Phase0:
		fieldRoots = make([][]byte, params.BeaconConfig().BeaconStateFieldCount)
	case version.Altair:
		fieldRoots = make([][]byte, params.BeaconConfig().BeaconStateAltairFieldCount)
	case version.Bellatrix:
		fieldRoots = make([][]byte, params.BeaconConfig().BeaconStateBellatrixFieldCount)
	case version.Capella:
		fieldRoots = make([][]byte, params.BeaconConfig().BeaconStateCapellaFieldCount)
	case version.Deneb:
		fieldRoots = make([][]byte, params.BeaconConfig().BeaconStateDenebFieldCount)
	case version.Electra:
		fieldRoots = make([][]byte, params.BeaconConfig().BeaconStateElectraFieldCount)
	case version.Fulu:
		fieldRoots = make([][]byte, params.BeaconConfig().BeaconStateFuluFieldCount)
	case version.Gloas:
		fieldRoots = make([][]byte, params.BeaconConfig().BeaconStateGloasFieldCount)
	default:
		return nil, fmt.Errorf("unknown state version %s", version.String(state.version))
	}

	// Genesis time root.
	genesisRoot := ssz.Uint64Root(state.genesisTime)
	fieldRoots[types.GenesisTime.RealPosition()] = genesisRoot[:]

	// Genesis validators root.
	var r [32]byte
	copy(r[:], state.genesisValidatorsRoot[:])
	fieldRoots[types.GenesisValidatorsRoot.RealPosition()] = r[:]

	// Slot root.
	slotRoot := ssz.Uint64Root(uint64(state.slot))
	fieldRoots[types.Slot.RealPosition()] = slotRoot[:]

	// Fork data structure root.
	forkHashTreeRoot, err := ssz.ForkRoot(state.fork)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute fork merkleization")
	}
	fieldRoots[types.Fork.RealPosition()] = forkHashTreeRoot[:]

	// BeaconBlockHeader data structure root.
	headerHashTreeRoot, err := stateutil.BlockHeaderRoot(state.latestBlockHeader)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute block header merkleization")
	}
	fieldRoots[types.LatestBlockHeader.RealPosition()] = headerHashTreeRoot[:]

	// BlockRoots array root.
	blockRootsRoot, err := stateutil.ArraysRoot(state.blockRootsVal().Slice(), fieldparams.BlockRootsLength)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute block roots merkleization")
	}
	fieldRoots[types.BlockRoots.RealPosition()] = blockRootsRoot[:]

	// StateRoots array root.
	stateRootsRoot, err := stateutil.ArraysRoot(state.stateRootsVal().Slice(), fieldparams.StateRootsLength)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute state roots merkleization")
	}
	fieldRoots[types.StateRoots.RealPosition()] = stateRootsRoot[:]

	// HistoricalRoots slice root.
	hRoots := make([][]byte, len(state.historicalRoots))
	for i := range hRoots {
		hRoots[i] = state.historicalRoots[i][:]
	}
	historicalRootsRt, err := ssz.ByteArrayRootWithLimit(hRoots, fieldparams.HistoricalRootsLength)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute historical roots merkleization")
	}
	fieldRoots[types.HistoricalRoots.RealPosition()] = historicalRootsRt[:]

	// SilaData data structure root.
	silaexecHashTreeRoot, err := stateutil.SilaExecutionRoot(state.silaexecData)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute silaData merkleization")
	}
	fieldRoots[types.SilaData.RealPosition()] = silaexecHashTreeRoot[:]

	// SilaDataVotes slice root.
	silaexecVotesRoot, err := stateutil.SilaDataVotesRoot(state.silaDataVotes)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute silaData votes merkleization")
	}
	fieldRoots[types.SilaDataVotes.RealPosition()] = silaexecVotesRoot[:]

	// SilaExecutionDepositIndex root.
	silaExecutionDepositIndexBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(silaExecutionDepositIndexBuf, state.silaExecutionDepositIndex)
	silaexecDepositBuf := bytesutil.ToBytes32(silaExecutionDepositIndexBuf)
	fieldRoots[types.SilaExecutionDepositIndex.RealPosition()] = silaexecDepositBuf[:]

	// Validators slice root.
	validatorsRoot, err := stateutil.ValidatorRegistryRoot(state.validatorsCompactVal())
	if err != nil {
		return nil, errors.Wrap(err, "could not compute validator registry merkleization")
	}
	fieldRoots[types.Validators.RealPosition()] = validatorsRoot[:]

	// Balances slice root.
	balancesRoot, err := stateutil.Uint64ListRootWithRegistryLimit(state.balancesVal())
	if err != nil {
		return nil, errors.Wrap(err, "could not compute validator balances merkleization")
	}
	fieldRoots[types.Balances.RealPosition()] = balancesRoot[:]

	// RandaoMixes array root.
	randaoRootsRoot, err := stateutil.ArraysRoot(state.randaoMixesVal().Slice(), fieldparams.RandaoMixesLength)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute randao roots merkleization")
	}
	fieldRoots[types.RandaoMixes.RealPosition()] = randaoRootsRoot[:]

	// Slashings array root.
	slashingsRootsRoot, err := ssz.SlashingsRoot(state.slashings)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute slashings merkleization")
	}
	fieldRoots[types.Slashings.RealPosition()] = slashingsRootsRoot[:]

	if state.version == version.Phase0 {
		// PreviousEpochAttestations slice root.
		prevAttsRoot, err := stateutil.EpochAttestationsRoot(state.previousEpochAttestations)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute previous epoch attestations merkleization")
		}
		fieldRoots[types.PreviousEpochAttestations.RealPosition()] = prevAttsRoot[:]

		// CurrentEpochAttestations slice root.
		currAttsRoot, err := stateutil.EpochAttestationsRoot(state.currentEpochAttestations)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute current epoch attestations merkleization")
		}
		fieldRoots[types.CurrentEpochAttestations.RealPosition()] = currAttsRoot[:]
	}

	if state.version >= version.Altair {
		// PreviousEpochParticipation slice root.
		prevParticipationRoot, err := stateutil.ParticipationBitsRoot(state.previousEpochParticipation)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute previous epoch participation merkleization")
		}
		fieldRoots[types.PreviousEpochParticipationBits.RealPosition()] = prevParticipationRoot[:]

		// CurrentEpochParticipation slice root.
		currParticipationRoot, err := stateutil.ParticipationBitsRoot(state.currentEpochParticipation)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute current epoch participation merkleization")
		}
		fieldRoots[types.CurrentEpochParticipationBits.RealPosition()] = currParticipationRoot[:]
	}

	// JustificationBits root.
	justifiedBitsRoot := bytesutil.ToBytes32(state.justificationBits)
	fieldRoots[types.JustificationBits.RealPosition()] = justifiedBitsRoot[:]

	// PreviousJustifiedCheckpoint data structure root.
	prevCheckRoot, err := ssz.CheckpointRoot(state.previousJustifiedCheckpoint)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute previous justified checkpoint merkleization")
	}
	fieldRoots[types.PreviousJustifiedCheckpoint.RealPosition()] = prevCheckRoot[:]

	// CurrentJustifiedCheckpoint data structure root.
	currJustRoot, err := ssz.CheckpointRoot(state.currentJustifiedCheckpoint)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute current justified checkpoint merkleization")
	}
	fieldRoots[types.CurrentJustifiedCheckpoint.RealPosition()] = currJustRoot[:]

	// FinalizedCheckpoint data structure root.
	finalRoot, err := ssz.CheckpointRoot(state.finalizedCheckpoint)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute finalized checkpoint merkleization")
	}
	fieldRoots[types.FinalizedCheckpoint.RealPosition()] = finalRoot[:]

	if state.version >= version.Altair {
		// Inactivity scores root.
		inactivityScoresRoot, err := stateutil.Uint64ListRootWithRegistryLimit(state.inactivityScoresVal())
		if err != nil {
			return nil, errors.Wrap(err, "could not compute inactivityScoreRoot")
		}
		fieldRoots[types.InactivityScores.RealPosition()] = inactivityScoresRoot[:]

		// Current sync committee root.
		currentSyncCommitteeRoot, err := stateutil.SyncCommitteeRoot(state.currentSyncCommittee)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute sync committee merkleization")
		}
		fieldRoots[types.CurrentSyncCommittee.RealPosition()] = currentSyncCommitteeRoot[:]

		// Next sync committee root.
		nextSyncCommitteeRoot, err := stateutil.SyncCommitteeRoot(state.nextSyncCommittee)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute sync committee merkleization")
		}
		fieldRoots[types.NextSyncCommittee.RealPosition()] = nextSyncCommitteeRoot[:]
	}

	if state.version == version.Bellatrix {
		// Sila payload root.
		silaPayloadRoot, err := state.latestSilaPayloadHeader.HashTreeRoot()
		if err != nil {
			return nil, err
		}
		fieldRoots[types.LatestSilaPayloadHeader.RealPosition()] = silaPayloadRoot[:]
	}

	if state.version == version.Capella {
		// Sila payload root.
		silaPayloadRoot, err := state.latestSilaPayloadHeaderCapella.HashTreeRoot()
		if err != nil {
			return nil, err
		}
		fieldRoots[types.LatestSilaPayloadHeaderCapella.RealPosition()] = silaPayloadRoot[:]
	}

	if state.version >= version.Deneb && state.version < version.Gloas {
		// Sila payload root.
		silaPayloadRoot, err := state.latestSilaPayloadHeaderDeneb.HashTreeRoot()
		if err != nil {
			return nil, err
		}
		fieldRoots[types.LatestSilaPayloadHeaderDeneb.RealPosition()] = silaPayloadRoot[:]
	}

	if state.version >= version.Gloas {
		// Sila payload bid root for Gloas.
		bidRoot, err := state.latestSilaPayloadBid.HashTreeRoot()
		if err != nil {
			return nil, err
		}

		fieldRoots[types.LatestSilaPayloadBid.RealPosition()] = bidRoot[:]
	}

	if state.version >= version.Capella {
		// Next withdrawal index root.
		nextWithdrawalIndexRoot := make([]byte, 32)
		binary.LittleEndian.PutUint64(nextWithdrawalIndexRoot, state.nextWithdrawalIndex)
		fieldRoots[types.NextWithdrawalIndex.RealPosition()] = nextWithdrawalIndexRoot

		// Next partial withdrawal validator index root.
		nextWithdrawalValidatorIndexRoot := make([]byte, 32)
		binary.LittleEndian.PutUint64(nextWithdrawalValidatorIndexRoot, uint64(state.nextWithdrawalValidatorIndex))
		fieldRoots[types.NextWithdrawalValidatorIndex.RealPosition()] = nextWithdrawalValidatorIndexRoot

		// Historical summary root.
		historicalSummaryRoot, err := stateutil.HistoricalSummariesRoot(state.historicalSummaries)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute historical summary merkleization")
		}
		fieldRoots[types.HistoricalSummaries.RealPosition()] = historicalSummaryRoot[:]
	}

	if state.version >= version.Electra {
		// DepositRequestsStartIndex root.
		drsiRoot := ssz.Uint64Root(state.depositRequestsStartIndex)
		fieldRoots[types.DepositRequestsStartIndex.RealPosition()] = drsiRoot[:]

		// DepositBalanceToConsume root.
		dbtcRoot := ssz.Uint64Root(uint64(state.depositBalanceToConsume))
		fieldRoots[types.DepositBalanceToConsume.RealPosition()] = dbtcRoot[:]

		// ExitBalanceToConsume root.
		ebtcRoot := ssz.Uint64Root(uint64(state.exitBalanceToConsume))
		fieldRoots[types.ExitBalanceToConsume.RealPosition()] = ebtcRoot[:]

		// EarliestExitEpoch root.
		eeeRoot := ssz.Uint64Root(uint64(state.earliestExitEpoch))
		fieldRoots[types.EarliestExitEpoch.RealPosition()] = eeeRoot[:]

		// ConsolidationBalanceToConsume root.
		cbtcRoot := ssz.Uint64Root(uint64(state.consolidationBalanceToConsume))
		fieldRoots[types.ConsolidationBalanceToConsume.RealPosition()] = cbtcRoot[:]

		// EarliestConsolidationEpoch root.
		eceRoot := ssz.Uint64Root(uint64(state.earliestConsolidationEpoch))
		fieldRoots[types.EarliestConsolidationEpoch.RealPosition()] = eceRoot[:]

		// PendingDeposits root.
		pbdRoot, err := stateutil.PendingDepositsRoot(state.pendingDeposits)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute pending balance deposits merkleization")
		}
		fieldRoots[types.PendingDeposits.RealPosition()] = pbdRoot[:]

		// PendingPartialWithdrawals root.
		ppwRoot, err := stateutil.PendingPartialWithdrawalsRoot(state.pendingPartialWithdrawals)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute pending partial withdrawals merkleization")
		}
		fieldRoots[types.PendingPartialWithdrawals.RealPosition()] = ppwRoot[:]

		// PendingConsolidations root.
		pcRoot, err := stateutil.PendingConsolidationsRoot(state.pendingConsolidations)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute pending consolidations merkleization")
		}
		fieldRoots[types.PendingConsolidations.RealPosition()] = pcRoot[:]
	}

	if state.version >= version.Fulu {
		// Proposer lookahead root.
		proposerLookaheadRoot, err := stateutil.ProposerLookaheadRoot(state.proposerLookahead)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute proposer lookahead merkleization")
		}
		fieldRoots[types.ProposerLookahead.RealPosition()] = proposerLookaheadRoot[:]
	}

	if state.version >= version.Gloas {
		buildersRoot, err := stateutil.BuildersRoot(state.builders)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute builders merkleization")
		}
		fieldRoots[types.Builders.RealPosition()] = buildersRoot[:]

		nextWithdrawalBuilderIndexRoot := ssz.Uint64Root(uint64(state.nextWithdrawalBuilderIndex))
		fieldRoots[types.NextWithdrawalBuilderIndex.RealPosition()] = nextWithdrawalBuilderIndexRoot[:]

		epaRoot, err := stateutil.SilaPayloadAvailabilityRoot(state.silaPayloadAvailability)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute sila payload availability merkleization")
		}

		fieldRoots[types.SilaPayloadAvailability.RealPosition()] = epaRoot[:]

		bppRoot, err := stateutil.BuilderPendingPaymentsRoot(state.builderPendingPayments)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute builder pending payments merkleization")
		}

		fieldRoots[types.BuilderPendingPayments.RealPosition()] = bppRoot[:]

		bpwRoot, err := stateutil.BuilderPendingWithdrawalsRoot(state.builderPendingWithdrawals)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute builder pending withdrawals merkleization")
		}

		fieldRoots[types.BuilderPendingWithdrawals.RealPosition()] = bpwRoot[:]

		lbhRoot := bytesutil.ToBytes32(state.latestBlockHash)
		fieldRoots[types.LatestBlockHash.RealPosition()] = lbhRoot[:]

		expectedWithdrawalsRoot, err := ssz.WithdrawalSliceRoot(state.payloadExpectedWithdrawals, fieldparams.MaxWithdrawalsPerPayload)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute payload expected withdrawals root")
		}

		fieldRoots[types.PayloadExpectedWithdrawals.RealPosition()] = expectedWithdrawalsRoot[:]

		ptcWindowRoot, err := stateutil.PTCWindowRoot(state.ptcWindow)
		if err != nil {
			return nil, errors.Wrap(err, "could not compute ptc window merkleization")
		}

		fieldRoots[types.PTCWindow.RealPosition()] = ptcWindowRoot[:]
	}
	return fieldRoots, nil
}
