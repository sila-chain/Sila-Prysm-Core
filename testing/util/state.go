package util

import (
	"context"
	"fmt"
	"testing"

	"github.com/sila-chain/go-bitfield"
	b "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db/iface"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila/common/hexutil"
)

// FillRootsNaturalOpt is meant to be used as an option when calling NewBeaconState.
// It fills state and block roots with hex representations of natural numbers starting with 0.
// Example: 16 becomes 0x00...0f.
func FillRootsNaturalOpt(state *silapb.BeaconState) error {
	roots, err := PrepareRoots(int(params.BeaconConfig().SlotsPerHistoricalRoot))
	if err != nil {
		return err
	}
	state.StateRoots = roots
	state.BlockRoots = roots
	return nil
}

// FillRootsNaturalOptAltair is meant to be used as an option when calling NewBeaconStateAltair.
// It fills state and block roots with hex representations of natural numbers starting with 0.
// Example: 16 becomes 0x00...0f.
func FillRootsNaturalOptAltair(state *silapb.BeaconStateAltair) error {
	roots, err := PrepareRoots(int(params.BeaconConfig().SlotsPerHistoricalRoot))
	if err != nil {
		return err
	}
	state.StateRoots = roots
	state.BlockRoots = roots
	return nil
}

// FillRootsNaturalOptBellatrix is meant to be used as an option when calling NewBeaconStateBellatrix.
// It fills state and block roots with hex representations of natural numbers starting with 0.
// Example: 16 becomes 0x00...0f.
func FillRootsNaturalOptBellatrix(state *silapb.BeaconStateBellatrix) error {
	roots, err := PrepareRoots(int(params.BeaconConfig().SlotsPerHistoricalRoot))
	if err != nil {
		return err
	}
	state.StateRoots = roots
	state.BlockRoots = roots
	return nil
}

// FillRootsNaturalOptCapella is meant to be used as an option when calling NewBeaconStateCapella.
// It fills state and block roots with hex representations of natural numbers starting with 0.
// Example: 16 becomes 0x00...0f.
func FillRootsNaturalOptCapella(state *silapb.BeaconStateCapella) error {
	roots, err := PrepareRoots(int(params.BeaconConfig().SlotsPerHistoricalRoot))
	if err != nil {
		return err
	}
	state.StateRoots = roots
	state.BlockRoots = roots
	return nil
}

type NewBeaconStateOption func(state *silapb.BeaconState) error

// NewBeaconState creates a beacon state with minimum marshalable fields.
func NewBeaconState(options ...NewBeaconStateOption) (state.BeaconState, error) {
	seed := &silapb.BeaconState{
		BlockRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		StateRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		Slashings:                  make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector),
		RandaoMixes:                filledByteSlice2D(uint64(params.BeaconConfig().EpochsPerHistoricalVector), 32),
		Validators:                 make([]*silapb.Validator, 0),
		CurrentJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		SilaData: &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		},
		Fork: &silapb.Fork{
			PreviousVersion: make([]byte, 4),
			CurrentVersion:  make([]byte, 4),
		},
		SilaDataVotes:               make([]*silapb.SilaData, 0),
		HistoricalRoots:             make([][]byte, 0),
		JustificationBits:           bitfield.Bitvector4{0x0},
		FinalizedCheckpoint:         &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		LatestBlockHeader:           HydrateBeaconHeader(&silapb.BeaconBlockHeader{}),
		PreviousEpochAttestations:   make([]*silapb.PendingAttestation, 0),
		CurrentEpochAttestations:    make([]*silapb.PendingAttestation, 0),
		PreviousJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
	}

	for _, opt := range options {
		err := opt(seed)
		if err != nil {
			return nil, err
		}
	}

	var st, err = state_native.InitializeFromProtoUnsafePhase0(seed)
	if err != nil {
		return nil, err
	}

	return st, nil
}

// NewBeaconStateAltair creates a beacon state with minimum marshalable fields.
func NewBeaconStateAltair(options ...func(state *silapb.BeaconStateAltair) error) (state.BeaconState, error) {
	pubkeys := make([][]byte, 512)
	for i := range pubkeys {
		pubkeys[i] = make([]byte, 48)
	}

	seed := &silapb.BeaconStateAltair{
		BlockRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		StateRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		Slashings:                  make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector),
		RandaoMixes:                filledByteSlice2D(uint64(params.BeaconConfig().EpochsPerHistoricalVector), 32),
		Validators:                 make([]*silapb.Validator, 0),
		CurrentJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		SilaData: &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		},
		Fork: &silapb.Fork{
			PreviousVersion: make([]byte, 4),
			CurrentVersion:  make([]byte, 4),
		},
		SilaDataVotes:               make([]*silapb.SilaData, 0),
		HistoricalRoots:             make([][]byte, 0),
		JustificationBits:           bitfield.Bitvector4{0x0},
		FinalizedCheckpoint:         &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		LatestBlockHeader:           HydrateBeaconHeader(&silapb.BeaconBlockHeader{}),
		PreviousJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		PreviousEpochParticipation:  make([]byte, 0),
		CurrentEpochParticipation:   make([]byte, 0),
		CurrentSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		NextSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
	}

	for _, opt := range options {
		err := opt(seed)
		if err != nil {
			return nil, err
		}
	}

	var st, err = state_native.InitializeFromProtoUnsafeAltair(seed)
	if err != nil {
		return nil, err
	}

	return st, nil
}

// NewBeaconStateBellatrix creates a beacon state with minimum marshalable fields.
func NewBeaconStateBellatrix(options ...func(state *silapb.BeaconStateBellatrix) error) (state.BeaconState, error) {
	pubkeys := make([][]byte, 512)
	for i := range pubkeys {
		pubkeys[i] = make([]byte, 48)
	}

	seed := &silapb.BeaconStateBellatrix{
		BlockRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		StateRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		Slashings:                  make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector),
		RandaoMixes:                filledByteSlice2D(uint64(params.BeaconConfig().EpochsPerHistoricalVector), 32),
		Validators:                 make([]*silapb.Validator, 0),
		CurrentJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		SilaData: &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		},
		Fork: &silapb.Fork{
			PreviousVersion: make([]byte, 4),
			CurrentVersion:  make([]byte, 4),
		},
		SilaDataVotes:               make([]*silapb.SilaData, 0),
		HistoricalRoots:             make([][]byte, 0),
		JustificationBits:           bitfield.Bitvector4{0x0},
		FinalizedCheckpoint:         &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		LatestBlockHeader:           HydrateBeaconHeader(&silapb.BeaconBlockHeader{}),
		PreviousJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		PreviousEpochParticipation:  make([]byte, 0),
		CurrentEpochParticipation:   make([]byte, 0),
		CurrentSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		NextSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		LatestSilaPayloadHeader: &silaenginev1.SilaPayloadHeader{
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
		},
	}

	for _, opt := range options {
		err := opt(seed)
		if err != nil {
			return nil, err
		}
	}

	var st, err = state_native.InitializeFromProtoUnsafeBellatrix(seed)
	if err != nil {
		return nil, err
	}

	return st, nil
}

// NewBeaconStateCapella creates a beacon state with minimum marshalable fields.
func NewBeaconStateCapella(options ...func(state *silapb.BeaconStateCapella) error) (state.BeaconState, error) {
	pubkeys := make([][]byte, 512)
	for i := range pubkeys {
		pubkeys[i] = make([]byte, 48)
	}

	seed := &silapb.BeaconStateCapella{
		BlockRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		StateRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		Slashings:                  make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector),
		RandaoMixes:                filledByteSlice2D(uint64(params.BeaconConfig().EpochsPerHistoricalVector), 32),
		Validators:                 make([]*silapb.Validator, 0),
		CurrentJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		SilaData: &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		},
		Fork: &silapb.Fork{
			PreviousVersion: make([]byte, 4),
			CurrentVersion:  make([]byte, 4),
		},
		SilaDataVotes:               make([]*silapb.SilaData, 0),
		HistoricalSummaries:         make([]*silapb.HistoricalSummary, 0),
		JustificationBits:           bitfield.Bitvector4{0x0},
		FinalizedCheckpoint:         &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		LatestBlockHeader:           HydrateBeaconHeader(&silapb.BeaconBlockHeader{}),
		PreviousJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		PreviousEpochParticipation:  make([]byte, 0),
		CurrentEpochParticipation:   make([]byte, 0),
		CurrentSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		NextSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		LatestSilaPayloadHeader: &silaenginev1.SilaPayloadHeaderCapella{
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
		},
	}

	for _, opt := range options {
		err := opt(seed)
		if err != nil {
			return nil, err
		}
	}

	var st, err = state_native.InitializeFromProtoUnsafeCapella(seed)
	if err != nil {
		return nil, err
	}

	return st, nil
}

// NewBeaconStateDeneb creates a beacon state with minimum marshalable fields.
func NewBeaconStateDeneb(options ...func(state *silapb.BeaconStateDeneb) error) (state.BeaconState, error) {
	pubkeys := make([][]byte, 512)
	for i := range pubkeys {
		pubkeys[i] = make([]byte, 48)
	}

	seed := &silapb.BeaconStateDeneb{
		BlockRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		StateRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		Slashings:                  make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector),
		RandaoMixes:                filledByteSlice2D(uint64(params.BeaconConfig().EpochsPerHistoricalVector), 32),
		Validators:                 make([]*silapb.Validator, 0),
		CurrentJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		SilaData: &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		},
		Fork: &silapb.Fork{
			PreviousVersion: make([]byte, 4),
			CurrentVersion:  make([]byte, 4),
		},
		SilaDataVotes:               make([]*silapb.SilaData, 0),
		HistoricalRoots:             make([][]byte, 0),
		JustificationBits:           bitfield.Bitvector4{0x0},
		FinalizedCheckpoint:         &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		LatestBlockHeader:           HydrateBeaconHeader(&silapb.BeaconBlockHeader{}),
		PreviousJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		PreviousEpochParticipation:  make([]byte, 0),
		CurrentEpochParticipation:   make([]byte, 0),
		CurrentSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		NextSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		LatestSilaPayloadHeader: &silaenginev1.SilaPayloadHeaderDeneb{
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
		},
	}

	for _, opt := range options {
		err := opt(seed)
		if err != nil {
			return nil, err
		}
	}

	var st, err = state_native.InitializeFromProtoUnsafeDeneb(seed)
	if err != nil {
		return nil, err
	}

	return st, nil
}

// NewBeaconStateElectra creates a beacon state with minimum marshalable fields.
func NewBeaconStateElectra(options ...func(state *silapb.BeaconStateElectra) error) (state.BeaconState, error) {
	pubkeys := make([][]byte, 512)
	for i := range pubkeys {
		pubkeys[i] = make([]byte, 48)
	}

	seed := &silapb.BeaconStateElectra{
		BlockRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		StateRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		Slashings:                  make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector),
		RandaoMixes:                filledByteSlice2D(uint64(params.BeaconConfig().EpochsPerHistoricalVector), 32),
		Validators:                 make([]*silapb.Validator, 0),
		CurrentJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		SilaData: &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		},
		Fork: &silapb.Fork{
			PreviousVersion: make([]byte, 4),
			CurrentVersion:  make([]byte, 4),
		},
		SilaDataVotes:               make([]*silapb.SilaData, 0),
		HistoricalRoots:             make([][]byte, 0),
		JustificationBits:           bitfield.Bitvector4{0x0},
		FinalizedCheckpoint:         &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		LatestBlockHeader:           HydrateBeaconHeader(&silapb.BeaconBlockHeader{}),
		PreviousJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		PreviousEpochParticipation:  make([]byte, 0),
		CurrentEpochParticipation:   make([]byte, 0),
		CurrentSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		NextSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		LatestSilaPayloadHeader: &silaenginev1.SilaPayloadHeaderDeneb{
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
		},
	}

	for _, opt := range options {
		err := opt(seed)
		if err != nil {
			return nil, err
		}
	}

	var st, err = state_native.InitializeFromProtoUnsafeElectra(seed)
	if err != nil {
		return nil, err
	}

	return st, nil
}

// NewBeaconStateFulu creates a beacon state with minimum marshalable fields.
func NewBeaconStateFulu(options ...func(state *silapb.BeaconStateFulu) error) (state.BeaconState, error) {
	pubkeys := make([][]byte, 512)
	for i := range pubkeys {
		pubkeys[i] = make([]byte, 48)
	}

	seed := &silapb.BeaconStateFulu{
		BlockRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		StateRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		Slashings:                  make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector),
		RandaoMixes:                filledByteSlice2D(uint64(params.BeaconConfig().EpochsPerHistoricalVector), 32),
		Validators:                 make([]*silapb.Validator, 0),
		CurrentJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		SilaData: &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		},
		Fork: &silapb.Fork{
			PreviousVersion: make([]byte, 4),
			CurrentVersion:  make([]byte, 4),
		},
		SilaDataVotes:               make([]*silapb.SilaData, 0),
		HistoricalRoots:             make([][]byte, 0),
		JustificationBits:           bitfield.Bitvector4{0x0},
		FinalizedCheckpoint:         &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		LatestBlockHeader:           HydrateBeaconHeader(&silapb.BeaconBlockHeader{}),
		PreviousJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		PreviousEpochParticipation:  make([]byte, 0),
		CurrentEpochParticipation:   make([]byte, 0),
		CurrentSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		NextSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		LatestSilaPayloadHeader: &silaenginev1.SilaPayloadHeaderDeneb{
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
		},
		ProposerLookahead: make([]primitives.ValidatorIndex, 64),
	}

	for _, opt := range options {
		err := opt(seed)
		if err != nil {
			return nil, err
		}
	}

	var st, err = state_native.InitializeFromProtoUnsafeFulu(seed)
	if err != nil {
		return nil, err
	}

	return st, nil
}

// NewBeaconStateGloas creates a beacon state with minimum marshalable fields.
func NewBeaconStateGloas(options ...func(state *silapb.BeaconStateGloas) error) (state.BeaconState, error) {
	pubkeys := make([][]byte, 512)
	for i := range pubkeys {
		pubkeys[i] = make([]byte, 48)
	}

	builderPendingPayments := make([]*silapb.BuilderPendingPayment, 64)
	for i := range builderPendingPayments {
		builderPendingPayments[i] = &silapb.BuilderPendingPayment{
			Withdrawal: &silapb.BuilderPendingWithdrawal{
				FeeRecipient: make([]byte, 20),
			},
		}
	}

	ptcWindow := make([]*silapb.PTCs, 3*params.BeaconConfig().SlotsPerEpoch)
	for i := range ptcWindow {
		ptcWindow[i] = &silapb.PTCs{
			ValidatorIndices: make([]primitives.ValidatorIndex, fieldparams.PTCSize),
		}
	}

	seed := &silapb.BeaconStateGloas{
		BlockRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		StateRoots:                 filledByteSlice2D(uint64(params.BeaconConfig().SlotsPerHistoricalRoot), 32),
		Slashings:                  make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector),
		RandaoMixes:                filledByteSlice2D(uint64(params.BeaconConfig().EpochsPerHistoricalVector), 32),
		Validators:                 make([]*silapb.Validator, 0),
		CurrentJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		SilaData: &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		},
		Fork: &silapb.Fork{
			PreviousVersion: make([]byte, 4),
			CurrentVersion:  make([]byte, 4),
		},
		SilaDataVotes:               make([]*silapb.SilaData, 0),
		HistoricalRoots:             make([][]byte, 0),
		JustificationBits:           bitfield.Bitvector4{0x0},
		FinalizedCheckpoint:         &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		LatestBlockHeader:           HydrateBeaconHeader(&silapb.BeaconBlockHeader{}),
		PreviousJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		PreviousEpochParticipation:  make([]byte, 0),
		CurrentEpochParticipation:   make([]byte, 0),
		CurrentSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		NextSyncCommittee: &silapb.SyncCommittee{
			Pubkeys:         pubkeys,
			AggregatePubkey: make([]byte, 48),
		},
		ProposerLookahead: make([]primitives.ValidatorIndex, 64),
		LatestSilaPayloadBid: &silapb.SilaPayloadBid{
			ParentBlockHash:       make([]byte, 32),
			ParentBlockRoot:       make([]byte, 32),
			BlockHash:             make([]byte, 32),
			PrevRandao:            make([]byte, 32),
			FeeRecipient:          make([]byte, 20),
			BlobKzgCommitments:    [][]byte{make([]byte, 48)},
			SilaRequestsRoot: make([]byte, 32),
		},
		Builders:                     make([]*silapb.Builder, 0),
		SilaPayloadAvailability: make([]byte, 1024),
		BuilderPendingPayments:       builderPendingPayments,
		BuilderPendingWithdrawals:    make([]*silapb.BuilderPendingWithdrawal, 0),
		LatestBlockHash:              make([]byte, 32),
		PayloadExpectedWithdrawals:   make([]*silaenginev1.Withdrawal, 0),
		PtcWindow:                    ptcWindow,
	}

	for _, opt := range options {
		err := opt(seed)
		if err != nil {
			return nil, err
		}
	}

	var st, err = state_native.InitializeFromProtoUnsafeGloas(seed)
	if err != nil {
		return nil, err
	}

	return st, nil
}

// SSZ will fill 2D byte slices with their respective values, so we must fill these in too for round
// trip testing.
func filledByteSlice2D(length, innerLen uint64) [][]byte {
	b := make([][]byte, length)
	for i := range length {
		b[i] = make([]byte, innerLen)
	}
	return b
}

// PrepareRoots returns a list of roots with hex representations of natural numbers starting with 0.
// Example: 16 becomes 0x00...0f.
func PrepareRoots(size int) ([][]byte, error) {
	roots := make([][]byte, size)
	for i := range size {
		roots[i] = make([]byte, fieldparams.RootLength)
	}
	for j := range roots {
		// Remove '0x' prefix and left-pad '0' to have 64 chars in total.
		s := fmt.Sprintf("%064s", hexutil.EncodeUint64(uint64(j))[2:])
		h, err := hexutil.Decode("0x" + s)
		if err != nil {
			return nil, err
		}
		roots[j] = h
	}
	return roots, nil
}

// DeterministicGenesisStateWithGenesisBlock creates a genesis state, saves the genesis block,
// genesis state and head block root. It returns the genesis state, genesis block's root and
// validator private keys.
func DeterministicGenesisStateWithGenesisBlock(
	t *testing.T,
	ctx context.Context,
	db iface.HeadAccessDatabase,
	numValidators uint64,
) (state.BeaconState, [32]byte, []bls.SecretKey) {
	genesisState, privateKeys := DeterministicGenesisState(t, numValidators)
	stateRoot, err := genesisState.HashTreeRoot(ctx)
	require.NoError(t, err, "Could not hash genesis state")

	genesis := b.NewGenesisBlock(stateRoot[:])
	SaveBlock(t, ctx, db, genesis)

	parentRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, genesisState, parentRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, parentRoot), "Could not save genesis state")

	return genesisState, parentRoot, privateKeys
}
