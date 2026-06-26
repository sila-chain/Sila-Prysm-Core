package util

import (
	"context"
	"fmt"
	"testing"

	"github.com/sila-chain/go-bitfield"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/altair"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/time"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/transition"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stateutil"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	consensusblocks "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// DeterministicGenesisStateAltair returns a genesis state in hard fork 1 format made using the deterministic deposits.
func DeterministicGenesisStateAltair(t testing.TB, numValidators uint64) (state.BeaconState, []bls.SecretKey) {
	deposits, privKeys, err := DeterministicDepositsAndKeys(numValidators)
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get %d deposits", numValidators))
	}
	silaexecData, err := DeterministicSilaData(len(deposits))
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get silaData for %d deposits", numValidators))
	}
	beaconState, err := GenesisBeaconState(context.Background(), deposits, uint64(0), silaexecData)
	if err != nil {
		t.Fatal(errors.Wrapf(err, "failed to get genesis beacon state of %d validators", numValidators))
	}
	resetCache()
	return beaconState, privKeys
}

// GenesisBeaconState returns the genesis beacon state.
func GenesisBeaconState(ctx context.Context, deposits []*silapb.Deposit, genesisTime uint64, silaexecData *silapb.SilaData) (state.BeaconState, error) {
	st, err := emptyGenesisState()
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

	return buildGenesisBeaconState(genesisTime, st, st.SilaData())
}

// processPreGenesisDeposits processes a deposit for the beacon state Altair before chain start.
func processPreGenesisDeposits(
	ctx context.Context,
	beaconState state.BeaconState,
	deposits []*silapb.Deposit,
) (state.BeaconState, error) {
	var err error
	beaconState, err = altair.ProcessDeposits(ctx, beaconState, deposits)
	if err != nil {
		return nil, errors.Wrap(err, "could not process deposit")
	}
	beaconState, err = helpers.ActivateValidatorWithEffectiveBalance(beaconState, deposits)
	if err != nil {
		return nil, err
	}
	return beaconState, nil
}

func buildGenesisBeaconState(genesisTime uint64, preState state.BeaconState, silaexecData *silapb.SilaData) (state.BeaconState, error) {
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
	st := &silapb.BeaconStateAltair{
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
		SilaExecutionDepositIndex: preState.SilaExecutionDepositIndex(),
	}

	var scBits [fieldparams.SyncAggregateSyncCommitteeBytesLength]byte
	bodyRoot, err := (&silapb.BeaconBlockBodyAltair{
		RandaoReveal: make([]byte, 96),
		SilaData: &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		},
		Graffiti: make([]byte, 32),
		SyncAggregate: &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, len(scBits[:])),
			SyncCommitteeSignature: make([]byte, 96),
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

	return state_native.InitializeFromProtoUnsafeAltair(st)
}

func emptyGenesisState() (state.BeaconState, error) {
	st := &silapb.BeaconStateAltair{
		// Misc fields.
		Slot: 0,
		Fork: &silapb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().AltairForkVersion,
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
		SilaExecutionDepositIndex: 0,
	}
	return state_native.InitializeFromProtoUnsafeAltair(st)
}

// NewBeaconBlockAltair creates a beacon block with minimum marshalable fields.
func NewBeaconBlockAltair() *silapb.SignedBeaconBlockAltair {
	var scBits [fieldparams.SyncAggregateSyncCommitteeBytesLength]byte
	return &silapb.SignedBeaconBlockAltair{
		Block: &silapb.BeaconBlockAltair{
			ParentRoot: make([]byte, fieldparams.RootLength),
			StateRoot:  make([]byte, fieldparams.RootLength),
			Body: &silapb.BeaconBlockBodyAltair{
				RandaoReveal: make([]byte, 96),
				SilaData: &silapb.SilaData{
					DepositRoot: make([]byte, fieldparams.RootLength),
					BlockHash:   make([]byte, 32),
				},
				Graffiti:          make([]byte, 32),
				Attestations:      []*silapb.Attestation{},
				AttesterSlashings: []*silapb.AttesterSlashing{},
				Deposits:          []*silapb.Deposit{},
				ProposerSlashings: []*silapb.ProposerSlashing{},
				VoluntaryExits:    []*silapb.SignedVoluntaryExit{},
				SyncAggregate: &silapb.SyncAggregate{
					SyncCommitteeBits:      scBits[:],
					SyncCommitteeSignature: make([]byte, 96),
				},
			},
		},
		Signature: make([]byte, 96),
	}
}

// BlockSignatureAltair calculates the post-state root of the block and returns the signature.
func BlockSignatureAltair(
	bState state.BeaconState,
	block *silapb.BeaconBlockAltair,
	privKeys []bls.SecretKey,
) (bls.Signature, error) {
	var err error
	wsb, err := consensusblocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockAltair{Block: block})
	if err != nil {
		return nil, err
	}
	s, err := transition.CalculateStateRoot(context.Background(), bState, wsb)
	if err != nil {
		return nil, err
	}
	block.StateRoot = s[:]
	domain, err := signing.Domain(bState.Fork(), time.CurrentEpoch(bState), params.BeaconConfig().DomainBeaconProposer, bState.GenesisValidatorsRoot())
	if err != nil {
		return nil, err
	}
	blockRoot, err := signing.ComputeSigningRoot(block, domain)
	if err != nil {
		return nil, err
	}
	// Temporarily increasing the beacon state slot here since BeaconProposerIndex is a
	// function deterministic on beacon state slot.
	currentSlot := bState.Slot()
	if err := bState.SetSlot(block.Slot); err != nil {
		return nil, err
	}
	proposerIdx, err := helpers.BeaconProposerIndex(context.Background(), bState)
	if err != nil {
		return nil, err
	}
	if err := bState.SetSlot(currentSlot); err != nil {
		return nil, err
	}
	return privKeys[proposerIdx].Sign(blockRoot[:]), nil
}

// GenerateFullBlockAltair generates a fully valid Altair block with the requested parameters.
// Use BlockGenConfig to declare the conditions you would like the block generated under.
func GenerateFullBlockAltair(
	bState state.BeaconState,
	privs []bls.SecretKey,
	conf *BlockGenConfig,
	slot primitives.Slot,
) (*silapb.SignedBeaconBlockAltair, error) {
	ctx := context.Background()
	currentSlot := bState.Slot()
	if currentSlot > slot {
		return nil, fmt.Errorf("current slot in state is larger than given slot. %d > %d", currentSlot, slot)
	}
	bState = bState.Copy()

	if conf == nil {
		conf = &BlockGenConfig{}
	}

	var err error
	var pSlashings []*silapb.ProposerSlashing
	numToGen := conf.NumProposerSlashings
	if numToGen > 0 {
		pSlashings, err = generateProposerSlashings(bState, privs, numToGen)
		if err != nil {
			return nil, errors.Wrapf(err, "failed generating %d proposer slashings:", numToGen)
		}
	}

	numToGen = conf.NumAttesterSlashings
	var aSlashings []*silapb.AttesterSlashing
	if numToGen > 0 {
		generated, err := generateAttesterSlashings(bState, privs, numToGen)
		if err != nil {
			return nil, errors.Wrapf(err, "failed generating %d attester slashings:", numToGen)
		}
		aSlashings = make([]*silapb.AttesterSlashing, len(generated))
		var ok bool
		for i, s := range generated {
			aSlashings[i], ok = s.(*silapb.AttesterSlashing)
			if !ok {
				return nil, fmt.Errorf("attester slashing has wrong type (expected %T, got %T)", &silapb.AttesterSlashing{}, s)
			}
		}
	}

	numToGen = conf.NumAttestations
	var atts []*silapb.Attestation
	if numToGen > 0 {
		generatedAtts, err := GenerateAttestations(bState, privs, numToGen, slot, false)
		if err != nil {
			return nil, errors.Wrapf(err, "failed generating %d attestations:", numToGen)
		}
		atts = make([]*silapb.Attestation, len(generatedAtts))
		var ok bool
		for i, a := range generatedAtts {
			atts[i], ok = a.(*silapb.Attestation)
			if !ok {
				return nil, fmt.Errorf("attestation has the wrong type (expected %T, got %T)", &silapb.Attestation{}, a)
			}
		}
	}

	numToGen = conf.NumDeposits
	var newDeposits []*silapb.Deposit
	silaexecData := bState.SilaData()
	if numToGen > 0 {
		newDeposits, silaexecData, err = generateDepositsAndSilaData(bState, numToGen)
		if err != nil {
			return nil, errors.Wrapf(err, "failed generating %d deposits:", numToGen)
		}
	}

	numToGen = conf.NumVoluntaryExits
	var exits []*silapb.SignedVoluntaryExit
	if numToGen > 0 {
		exits, err = generateVoluntaryExits(bState, privs, numToGen)
		if err != nil {
			return nil, errors.Wrapf(err, "failed generating %d attester slashings:", numToGen)
		}
	}

	newHeader := bState.LatestBlockHeader()
	prevStateRoot, err := bState.HashTreeRoot(ctx)
	if err != nil {
		return nil, err
	}
	newHeader.StateRoot = prevStateRoot[:]
	parentRoot, err := newHeader.HashTreeRoot()
	if err != nil {
		return nil, err
	}

	var newSyncAggregate *silapb.SyncAggregate
	if conf.FullSyncAggregate {
		newSyncAggregate, err = generateSyncAggregate(bState, privs, parentRoot)
		if err != nil {
			return nil, errors.Wrap(err, "failed generating syncAggregate")
		}
	} else {
		var syncCommitteeBits []byte
		currSize := new(silapb.SyncAggregate).SyncCommitteeBits.Len()
		switch currSize {
		case 512:
			syncCommitteeBits = bitfield.NewBitvector512()
		case 32:
			syncCommitteeBits = bitfield.NewBitvector32()
		default:
			return nil, errors.New("invalid bit vector size")
		}
		newSyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      syncCommitteeBits,
			SyncCommitteeSignature: append([]byte{0xC0}, make([]byte, 95)...),
		}
	}

	if slot == currentSlot {
		slot = currentSlot + 1
	}

	stCopy := bState.Copy()
	stCopy, err = transition.ProcessSlots(context.Background(), stCopy, slot)
	if err != nil {
		return nil, err
	}
	reveal, err := RandaoReveal(stCopy, time.CurrentEpoch(stCopy), privs)
	if err != nil {
		return nil, err
	}

	idx, err := helpers.BeaconProposerIndex(ctx, stCopy)
	if err != nil {
		return nil, err
	}

	block := &silapb.BeaconBlockAltair{
		Slot:          slot,
		ParentRoot:    parentRoot[:],
		ProposerIndex: idx,
		Body: &silapb.BeaconBlockBodyAltair{
			SilaData:          silaexecData,
			RandaoReveal:      reveal,
			ProposerSlashings: pSlashings,
			AttesterSlashings: aSlashings,
			Attestations:      atts,
			VoluntaryExits:    exits,
			Deposits:          newDeposits,
			Graffiti:          make([]byte, fieldparams.RootLength),
			SyncAggregate:     newSyncAggregate,
		},
	}

	signature, err := BlockSignature(bState, block, privs)
	if err != nil {
		return nil, err
	}

	return &silapb.SignedBeaconBlockAltair{Block: block, Signature: signature.Marshal()}, nil
}
