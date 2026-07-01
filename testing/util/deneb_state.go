package util

import (
	"context"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stateutil"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// DeterministicGenesisStateDeneb returns a genesis state in Deneb format made using the deterministic deposits.
func DeterministicGenesisStateDeneb(t testing.TB, numValidators uint64) (state.BeaconState, []bls.SecretKey) {
	deposits, privKeys, err := DeterministicDepositsAndKeys(numValidators)
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get %d deposits", numValidators))
	}
	silaexecData, err := DeterministicSilaData(len(deposits))
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get silaData for %d deposits", numValidators))
	}
	beaconState, err := genesisBeaconStateDeneb(context.Background(), deposits, uint64(0), silaexecData)
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get genesis beacon state of %d validators", numValidators))
	}
	resetCache()
	return beaconState, privKeys
}

// genesisBeaconStateDeneb returns the genesis beacon state.
func genesisBeaconStateDeneb(ctx context.Context, deposits []*silapb.Deposit, genesisTime uint64, silaexecData *silapb.SilaData) (state.BeaconState, error) {
	st, err := emptyGenesisStateDeneb()
	if err != nil {
		return nil, err
	}

	// Process initial deposits.
	st, err = helpers.UpdateGenesisSilaData(st, deposits, silaexecData)
	if err != nil {
		return nil, err
	}

	st, err = processPreGenesisDeposits(ctx, st, deposits)
	if err != nil {
		return nil, errors.Wrap(err, "could not process validator deposits")
	}

	return buildGenesisBeaconStateDeneb(genesisTime, st, st.SilaData())
}

// emptyGenesisStateDeneb returns an empty genesis state in Deneb format.
func emptyGenesisStateDeneb() (state.BeaconState, error) {
	st := &silapb.BeaconStateDeneb{
		// Misc fields.
		Slot: 0,
		Fork: &silapb.Fork{
			PreviousVersion: params.BeaconConfig().BellatrixForkVersion,
			CurrentVersion:  params.BeaconConfig().DenebForkVersion,
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
		SilaData:         &silapb.SilaData{},
		SilaDataVotes:    []*silapb.SilaData{},
		SilaexecDepositIndex: 0,

		LatestSilaPayloadHeader: &silaenginev1.SilaPayloadHeaderDeneb{},
	}
	return state_native.InitializeFromProtoUnsafeDeneb(st)
}

func buildGenesisBeaconStateDeneb(genesisTime uint64, preState state.BeaconState, silaexecData *silapb.SilaData) (state.BeaconState, error) {
	if silaexecData == nil {
		return nil, errors.New("no silaData provided for genesis state")
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
	st := &silapb.BeaconStateDeneb{
		// Misc fields.
		Slot:                  0,
		GenesisTime:           genesisTime,
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
		SilaData:         silaexecData,
		SilaDataVotes:    []*silapb.SilaData{},
		SilaexecDepositIndex: preState.SilaExecutionDepositIndex(),
	}

	var scBits [fieldparams.SyncAggregateSyncCommitteeBytesLength]byte
	bodyRoot, err := (&silapb.BeaconBlockBodyDeneb{
		RandaoReveal: make([]byte, 96),
		SilaData: &silapb.SilaData{
			DepositRoot: make([]byte, 32),
			BlockHash:   make([]byte, 32),
		},
		Graffiti: make([]byte, 32),
		SyncAggregate: &silapb.SyncAggregate{
			SyncCommitteeBits:      scBits[:],
			SyncCommitteeSignature: make([]byte, 96),
		},
		SilaPayload: &silaenginev1.SilaPayloadDeneb{
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
			Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
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
	vals := preState.Validators()
	for i := uint64(0); i < params.BeaconConfig().SyncCommitteeSize; i++ {
		j := i % uint64(len(vals))
		pubKeys = append(pubKeys, vals[j].PublicKey)
	}
	aggregated, err := bls.AggregatePublicKeys(pubKeys)
	if err != nil {
		return nil, err
	}
	st.CurrentSyncCommittee = &silapb.SyncCommittee{
		Pubkeys:         pubKeys,
		AggregatePubkey: aggregated.Marshal(),
	}
	st.NextSyncCommittee = &silapb.SyncCommittee{
		Pubkeys:         pubKeys,
		AggregatePubkey: aggregated.Marshal(),
	}

	st.LatestSilaPayloadHeader = &silaenginev1.SilaPayloadHeaderDeneb{
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
		WithdrawalsRoot:  make([]byte, 32),
	}

	return state_native.InitializeFromProtoUnsafeDeneb(st)
}
