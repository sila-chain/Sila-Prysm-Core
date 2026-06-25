package util

import (
	"context"
	"testing"
	"time"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stateutil"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// DeterministicGenesisStateBellatrix returns a genesis state in Bellatrix format made using the deterministic deposits.
func DeterministicGenesisStateBellatrix(t testing.TB, numValidators uint64) (state.BeaconState, []bls.SecretKey) {
	deposits, privKeys, err := DeterministicDepositsAndKeys(numValidators)
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get %d deposits", numValidators))
	}
	silaexecData, err := DeterministicSilaExecutionData(len(deposits))
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get silaExecutionData for %d deposits", numValidators))
	}
	beaconState, err := genesisBeaconStateBellatrix(t.Context(), deposits, time.Unix(0, 0), silaexecData)
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get genesis beacon state of %d validators", numValidators))
	}
	resetCache()
	return beaconState, privKeys
}

// genesisBeaconStateBellatrix returns the genesis beacon state.
func genesisBeaconStateBellatrix(ctx context.Context, deposits []*silapb.Deposit, genesisTime time.Time, silaexecData *silapb.SilaExecutionData) (state.BeaconState, error) {
	st, err := emptyGenesisStateBellatrix()
	if err != nil {
		return nil, err
	}

	// Process initial deposits.
	st, err = helpers.UpdateGenesisSilaExecutionData(st, deposits, silaexecData)
	if err != nil {
		return nil, err
	}

	st, err = processPreGenesisDeposits(ctx, st, deposits)
	if err != nil {
		return nil, errors.Wrap(err, "could not process validator deposits")
	}

	return buildGenesisBeaconStateBellatrix(genesisTime, st, st.SilaExecutionData())
}

// emptyGenesisStateBellatrix returns an empty genesis state in Bellatrix format.
func emptyGenesisStateBellatrix() (state.BeaconState, error) {
	st := &silapb.BeaconStateBellatrix{
		// Misc fields.
		Slot: 0,
		Fork: &silapb.Fork{
			PreviousVersion: params.BeaconConfig().AltairForkVersion,
			CurrentVersion:  params.BeaconConfig().BellatrixForkVersion,
			Epoch:           0,
		},
		// Validator registry fields.
		Validators:       []*silapb.Validator{},
		Balances:         []uint64{},
		InactivityScores: []uint64{},

		JustificationBits:          []byte{0},
		HistoricalRoots:            [][]byte{},
		CurrentEpochParticipation:  []byte{},
		PreviousEpochParticipation: []byte{},

		// SilaExecution data.
		SilaExecutionData:         &silapb.SilaExecutionData{},
		SilaExecutionDataVotes:    []*silapb.SilaExecutionData{},
		SilaExecutionDepositIndex: 0,

		LatestSilaPayloadHeader: &silaenginev1.SilaPayloadHeader{},
	}
	return state_native.InitializeFromProtoUnsafeBellatrix(st)
}

func buildGenesisBeaconStateBellatrix(genesisTime time.Time, preState state.BeaconState, silaexecData *silapb.SilaExecutionData) (state.BeaconState, error) {
	if silaexecData == nil {
		return nil, errors.New("no silaExecutionData provided for genesis state")
	}

	randaoMixes := make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector)
	for i := range randaoMixes {
		h := make([]byte, 32)
		copy(h, silaexecData.BlockHash)
		randaoMixes[i] = h
	}

	zeroHash := params.BeaconConfig().ZeroHash[:]

	activeIndexRoots := make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector)
	for i := range activeIndexRoots {
		activeIndexRoots[i] = zeroHash
	}

	blockRoots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := range blockRoots {
		blockRoots[i] = zeroHash
	}

	stateRoots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := range stateRoots {
		stateRoots[i] = zeroHash
	}

	slashings := make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector)

	compactValidators := stateutil.CompactValidatorsFromProto(preState.Validators())
	genesisValidatorsRoot, err := stateutil.ValidatorRegistryRoot(compactValidators)
	if err != nil {
		return nil, errors.Wrapf(err, "could not hash tree root genesis validators %v", err)
	}

	prevEpochParticipation, err := preState.PreviousEpochParticipation()
	if err != nil {
		return nil, err
	}
	currEpochParticipation, err := preState.CurrentEpochParticipation()
	if err != nil {
		return nil, err
	}
	scores, err := preState.InactivityScores()
	if err != nil {
		return nil, err
	}
	scoresMissing := len(preState.Validators()) - len(scores)
	if scoresMissing > 0 {
		for range scoresMissing {
			scores = append(scores, 0)
		}
	}
	st := &silapb.BeaconStateBellatrix{
		// Misc fields.
		Slot:                  0,
		GenesisTime:           uint64(genesisTime.Unix()),
		GenesisValidatorsRoot: genesisValidatorsRoot[:],

		Fork: &silapb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			Epoch:           0,
		},

		// Validator registry fields.
		Validators:                 preState.Validators(),
		Balances:                   preState.Balances(),
		PreviousEpochParticipation: prevEpochParticipation,
		CurrentEpochParticipation:  currEpochParticipation,
		InactivityScores:           scores,

		// Randomness and committees.
		RandaoMixes: randaoMixes,

		// Finality.
		PreviousJustifiedCheckpoint: &silapb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},
		CurrentJustifiedCheckpoint: &silapb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},
		JustificationBits: []byte{0},
		FinalizedCheckpoint: &silapb.Checkpoint{
			Epoch: 0,
			Root:  params.BeaconConfig().ZeroHash[:],
		},

		HistoricalRoots: [][]byte{},
		BlockRoots:      blockRoots,
		StateRoots:      stateRoots,
		Slashings:       slashings,

		// SilaExecution data.
		SilaExecutionData:         silaexecData,
		SilaExecutionDataVotes:    []*silapb.SilaExecutionData{},
		SilaExecutionDepositIndex: preState.SilaExecutionDepositIndex(),
	}

	var scBits [fieldparams.SyncAggregateSyncCommitteeBytesLength]byte
	bodyRoot, err := (&silapb.BeaconBlockBodyBellatrix{
		RandaoReveal: make([]byte, 96),
		SilaExecutionData: &silapb.SilaExecutionData{
			DepositRoot: make([]byte, 32),
			BlockHash:   make([]byte, 32),
		},
		Graffiti: make([]byte, 32),
		SyncAggregate: &silapb.SyncAggregate{
			SyncCommitteeBits:      scBits[:],
			SyncCommitteeSignature: make([]byte, 96),
		},
		SilaPayload: &silaenginev1.SilaPayload{
			ParentHash:    make([]byte, 32),
			FeeRecipient:  make([]byte, 20),
			StateRoot:     make([]byte, 32),
			ReceiptsRoot:  make([]byte, 32),
			LogsBloom:     make([]byte, 256),
			PrevRandao:    make([]byte, 32),
			ExtraData:     make([]byte, 0),
			BaseFeePerGas: make([]byte, 32),
			BlockHash:     make([]byte, 32),
			Transactions:  make([][]byte, 0),
		},
	}).HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "could not hash tree root empty block body")
	}

	st.LatestBlockHeader = &silapb.BeaconBlockHeader{
		ParentRoot: zeroHash,
		StateRoot:  zeroHash,
		BodyRoot:   bodyRoot[:],
	}

	var pubKeys [][]byte
	for i := uint64(0); i < params.BeaconConfig().SyncCommitteeSize; i++ {
		pubKeys = append(pubKeys, bytesutil.PadTo([]byte{}, params.BeaconConfig().BLSPubkeyLength))
	}
	st.CurrentSyncCommittee = &silapb.SyncCommittee{
		Pubkeys:         pubKeys,
		AggregatePubkey: bytesutil.PadTo([]byte{}, params.BeaconConfig().BLSPubkeyLength),
	}
	st.NextSyncCommittee = &silapb.SyncCommittee{
		Pubkeys:         bytesutil.SafeCopy2dBytes(pubKeys),
		AggregatePubkey: bytesutil.PadTo([]byte{}, params.BeaconConfig().BLSPubkeyLength),
	}

	st.LatestSilaPayloadHeader = &silaenginev1.SilaPayloadHeader{
		ParentHash:       make([]byte, 32),
		FeeRecipient:     make([]byte, 20),
		StateRoot:        make([]byte, 32),
		ReceiptsRoot:     make([]byte, 32),
		LogsBloom:        make([]byte, 256),
		PrevRandao:       make([]byte, 32),
		ExtraData:        make([]byte, 0),
		BaseFeePerGas:    make([]byte, 32),
		BlockHash:        make([]byte, 32),
		TransactionsRoot: make([]byte, 32),
	}

	bs, err := state_native.InitializeFromProtoUnsafeBellatrix(st)
	if err != nil {
		return nil, err
	}
	is, err := bs.InactivityScores()
	if err != nil {
		return nil, err
	}
	if bs.NumValidators() != len(is) {
		return nil, errors.New("inactivity score mismatch with num vals")
	}
	return bs, nil
}
