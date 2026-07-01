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
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/pkg/errors"
)

type ElectraStateOption func(*silapb.BeaconStateElectra) error

func WithElectraStateSlot(slot primitives.Slot) ElectraStateOption {
	return func(s *silapb.BeaconStateElectra) error {
		s.Slot = slot
		return nil
	}
}

// DeterministicGenesisStateElectra returns a genesis state in Electra format made using the deterministic deposits.
func DeterministicGenesisStateElectra(t testing.TB, numValidators uint64, opts ...ElectraStateOption) (state.BeaconState, []bls.SecretKey) {
	deposits, privKeys, err := DeterministicDepositsAndKeys(numValidators)
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get %d deposits", numValidators))
	}
	silaexecData, err := DeterministicSilaData(len(deposits))
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get silaData for %d deposits", numValidators))
	}
	beaconState, err := genesisBeaconStateElectra(t.Context(), deposits, uint64(0), silaexecData, opts...)
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get genesis beacon state of %d validators", numValidators))
	}
	if err := setKeysToActive(beaconState); err != nil {
		t.Fatal(errors.Wrapf(err, "failed to set keys to active"))
	}
	resetCache()
	return beaconState, privKeys
}

// setKeysToActive is a function to set the validators to active post electra, electra no longer processes deposits based on silaData
func setKeysToActive(beaconState state.BeaconState) error {
	vals := make([]*silapb.Validator, len(beaconState.Validators()))
	for i, val := range beaconState.Validators() {
		val.ActivationEpoch = 0
		val.EffectiveBalance = params.BeaconConfig().MinActivationBalance
		vals[i] = val
	}
	return beaconState.SetValidators(vals)
}

// genesisBeaconStateElectra returns the genesis beacon state.
func genesisBeaconStateElectra(ctx context.Context, deposits []*silapb.Deposit, genesisTime uint64, silaexecData *silapb.SilaData, opts ...ElectraStateOption) (state.BeaconState, error) {
	st, err := emptyGenesisStateElectra()
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

	return buildGenesisBeaconStateElectra(ctx, genesisTime, st, st.SilaData(), opts...)
}

// emptyGenesisStateElectra returns an empty genesis state in Electra format.
func emptyGenesisStateElectra() (state.BeaconState, error) {
	st := &silapb.BeaconStateElectra{
		// Misc fields.
		Slot: 0,
		Fork: &silapb.Fork{
			PreviousVersion: params.BeaconConfig().DenebForkVersion,
			CurrentVersion:  params.BeaconConfig().ElectraForkVersion,
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

		DepositBalanceToConsume:       primitives.Gwei(0),
		ExitBalanceToConsume:          primitives.Gwei(0),
		ConsolidationBalanceToConsume: primitives.Gwei(0),
	}
	return state_native.InitializeFromProtoUnsafeElectra(st)
}

func buildGenesisBeaconStateElectra(ctx context.Context, genesisTime uint64, preState state.BeaconState, silaexecData *silapb.SilaData, opts ...ElectraStateOption) (state.BeaconState, error) {
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
	tab, err := helpers.TotalActiveBalance(ctx, preState)
	if err != nil {
		return nil, err
	}
	st := &silapb.BeaconStateElectra{
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

		// Electra Data
		DepositRequestsStartIndex:     params.BeaconConfig().UnsetDepositRequestsStartIndex,
		ExitBalanceToConsume:          helpers.ActivationExitChurnLimit(primitives.Gwei(tab)),
		EarliestConsolidationEpoch:    helpers.ActivationExitEpoch(slots.ToEpoch(preState.Slot())),
		ConsolidationBalanceToConsume: helpers.ConsolidationChurnLimit(primitives.Gwei(tab)),
		PendingDeposits:               make([]*silapb.PendingDeposit, 0),
		PendingPartialWithdrawals:     make([]*silapb.PendingPartialWithdrawal, 0),
		PendingConsolidations:         make([]*silapb.PendingConsolidation, 0),
	}
	for _, opt := range opts {
		if err := opt(st); err != nil {
			return nil, err
		}
	}

	var scBits [fieldparams.SyncAggregateSyncCommitteeBytesLength]byte
	bodyRoot, err := (&silapb.BeaconBlockBodyElectra{
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
		},
		SilaRequests: &silaenginev1.SilaRequests{
			Deposits:       make([]*silaenginev1.DepositRequest, 0),
			Withdrawals:    make([]*silaenginev1.WithdrawalRequest, 0),
			Consolidations: make([]*silaenginev1.ConsolidationRequest, 0),
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

	return state_native.InitializeFromProtoUnsafeElectra(st)
}
