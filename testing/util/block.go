package util

import (
	"context"
	rd "crypto/rand"
	"fmt"
	"math/big"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/time"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db/iface"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/rand"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	v1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaapi/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assertions"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila/common"
	"github.com/pkg/errors"
)

// BlockGenConfig is used to define the requested conditions
// for block generation.
type BlockGenConfig struct {
	NumProposerSlashings     uint64
	NumAttesterSlashings     uint64
	NumAttestations          uint64
	NumDeposits              uint64
	NumVoluntaryExits        uint64
	NumTransactions          uint64 // Only for post Bellatrix blocks
	FullSyncAggregate        bool
	NumBLSChanges            uint64 // Only for post Capella blocks
	NumWithdrawals           uint64
	NumDepositRequests       uint64 // Only for post Electra blocks
	NumWithdrawalRequests    uint64 // Only for post Electra blocks
	NumConsolidationRequests uint64 // Only for post Electra blocks
	NumBlobKzgCommitments    uint64 // Only for post Deneb blocks
}

// DefaultBlockGenConfig returns the block config that utilizes the
// current params in the beacon config.
func DefaultBlockGenConfig() *BlockGenConfig {
	return &BlockGenConfig{
		NumProposerSlashings:     0,
		NumAttesterSlashings:     0,
		NumAttestations:          1,
		NumDeposits:              0,
		NumVoluntaryExits:        0,
		NumTransactions:          0,
		NumBLSChanges:            0,
		NumWithdrawals:           0,
		NumConsolidationRequests: 0,
		NumWithdrawalRequests:    0,
		NumDepositRequests:       0,
		NumBlobKzgCommitments:    0,
	}
}

// ----------------------------------------------------------------------------
// Phase 0
// ----------------------------------------------------------------------------

// NewBeaconBlock creates a beacon block with minimum marshalable fields.
func NewBeaconBlock() *silapb.SignedBeaconBlock {
	return &silapb.SignedBeaconBlock{
		Block: &silapb.BeaconBlock{
			ParentRoot: make([]byte, fieldparams.RootLength),
			StateRoot:  make([]byte, fieldparams.RootLength),
			Body: &silapb.BeaconBlockBody{
				RandaoReveal: make([]byte, fieldparams.BLSSignatureLength),
				SilaData: &silapb.SilaData{
					DepositRoot: make([]byte, fieldparams.RootLength),
					BlockHash:   make([]byte, fieldparams.RootLength),
				},
				Graffiti:          make([]byte, fieldparams.RootLength),
				Attestations:      []*silapb.Attestation{},
				AttesterSlashings: []*silapb.AttesterSlashing{},
				Deposits:          []*silapb.Deposit{},
				ProposerSlashings: []*silapb.ProposerSlashing{},
				VoluntaryExits:    []*silapb.SignedVoluntaryExit{},
			},
		},
		Signature: make([]byte, fieldparams.BLSSignatureLength),
	}
}

// GenerateFullBlock generates a fully valid block with the requested parameters.
// Use BlockGenConfig to declare the conditions you would like the block generated under.
func GenerateFullBlock(
	bState state.BeaconState,
	privs []bls.SecretKey,
	conf *BlockGenConfig,
	slot primitives.Slot,
) (*silapb.SignedBeaconBlock, error) {
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
				return nil, fmt.Errorf("attester slashing has the wrong type (expected %T, got %T)", &silapb.AttesterSlashing{}, s)
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
			return nil, errors.Wrapf(err, "failed generating %d voluntary exits:", numToGen)
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

	if slot == currentSlot {
		slot = currentSlot + 1
	}

	// Temporarily incrementing the beacon state slot here since BeaconProposerIndex is a
	// function deterministic on beacon state slot.
	if err := bState.SetSlot(slot); err != nil {
		return nil, err
	}
	reveal, err := RandaoReveal(bState, time.CurrentEpoch(bState), privs)
	if err != nil {
		return nil, err
	}

	idx, err := helpers.BeaconProposerIndex(ctx, bState)
	if err != nil {
		return nil, err
	}

	block := &silapb.BeaconBlock{
		Slot:          slot,
		ParentRoot:    parentRoot[:],
		ProposerIndex: idx,
		Body: &silapb.BeaconBlockBody{
			SilaData:          silaexecData,
			RandaoReveal:      reveal,
			ProposerSlashings: pSlashings,
			AttesterSlashings: aSlashings,
			Attestations:      atts,
			VoluntaryExits:    exits,
			Deposits:          newDeposits,
			Graffiti:          make([]byte, fieldparams.RootLength),
		},
	}
	if err := bState.SetSlot(currentSlot); err != nil {
		return nil, err
	}

	signature, err := BlockSignature(bState, block, privs)
	if err != nil {
		return nil, err
	}

	return &silapb.SignedBeaconBlock{Block: block, Signature: signature.Marshal()}, nil
}

// GenerateProposerSlashingForValidator for a specific validator index.
func GenerateProposerSlashingForValidator(
	bState state.BeaconState,
	priv bls.SecretKey,
	idx primitives.ValidatorIndex,
) (*silapb.ProposerSlashing, error) {
	header1 := HydrateSignedBeaconHeader(&silapb.SignedBeaconBlockHeader{
		Header: &silapb.BeaconBlockHeader{
			ProposerIndex: idx,
			Slot:          bState.Slot(),
			BodyRoot:      bytesutil.PadTo([]byte{0, 1, 0}, fieldparams.RootLength),
		},
	})
	currentEpoch := time.CurrentEpoch(bState)
	var err error
	header1.Signature, err = signing.ComputeDomainAndSign(bState, currentEpoch, header1.Header, params.BeaconConfig().DomainBeaconProposer, priv)
	if err != nil {
		return nil, err
	}

	header2 := &silapb.SignedBeaconBlockHeader{
		Header: &silapb.BeaconBlockHeader{
			ProposerIndex: idx,
			Slot:          bState.Slot(),
			BodyRoot:      bytesutil.PadTo([]byte{0, 2, 0}, fieldparams.RootLength),
			StateRoot:     make([]byte, fieldparams.RootLength),
			ParentRoot:    make([]byte, fieldparams.RootLength),
		},
	}
	header2.Signature, err = signing.ComputeDomainAndSign(bState, currentEpoch, header2.Header, params.BeaconConfig().DomainBeaconProposer, priv)
	if err != nil {
		return nil, err
	}

	return &silapb.ProposerSlashing{
		Header_1: header1,
		Header_2: header2,
	}, nil
}

func generateProposerSlashings(
	bState state.BeaconState,
	privs []bls.SecretKey,
	numSlashings uint64,
) ([]*silapb.ProposerSlashing, error) {
	proposerSlashings := make([]*silapb.ProposerSlashing, numSlashings)
	for i := range numSlashings {
		proposerIndex, err := randValIndex(bState)
		if err != nil {
			return nil, err
		}
		slashing, err := GenerateProposerSlashingForValidator(bState, privs[proposerIndex], proposerIndex)
		if err != nil {
			return nil, err
		}
		proposerSlashings[i] = slashing
	}
	return proposerSlashings, nil
}

// GenerateAttesterSlashingForValidator for a specific validator index.
func GenerateAttesterSlashingForValidator(
	bState state.BeaconState,
	priv bls.SecretKey,
	idx primitives.ValidatorIndex,
) (silapb.AttSlashing, error) {
	currentEpoch := time.CurrentEpoch(bState)

	if bState.Version() >= version.Electra {
		att1 := &silapb.IndexedAttestationElectra{
			Data: &silapb.AttestationData{
				Slot:            bState.Slot(),
				CommitteeIndex:  0,
				BeaconBlockRoot: make([]byte, fieldparams.RootLength),
				Target: &silapb.Checkpoint{
					Epoch: currentEpoch,
					Root:  params.BeaconConfig().ZeroHash[:],
				},
				Source: &silapb.Checkpoint{
					Epoch: currentEpoch + 1,
					Root:  params.BeaconConfig().ZeroHash[:],
				},
			},
			AttestingIndices: []uint64{uint64(idx)},
		}
		var err error
		att1.Signature, err = signing.ComputeDomainAndSign(bState, currentEpoch, att1.Data, params.BeaconConfig().DomainBeaconAttester, priv)
		if err != nil {
			return nil, err
		}

		att2 := &silapb.IndexedAttestationElectra{
			Data: &silapb.AttestationData{
				Slot:            bState.Slot(),
				CommitteeIndex:  0,
				BeaconBlockRoot: make([]byte, fieldparams.RootLength),
				Target: &silapb.Checkpoint{
					Epoch: currentEpoch,
					Root:  params.BeaconConfig().ZeroHash[:],
				},
				Source: &silapb.Checkpoint{
					Epoch: currentEpoch,
					Root:  params.BeaconConfig().ZeroHash[:],
				},
			},
			AttestingIndices: []uint64{uint64(idx)},
		}
		att2.Signature, err = signing.ComputeDomainAndSign(bState, currentEpoch, att2.Data, params.BeaconConfig().DomainBeaconAttester, priv)
		if err != nil {
			return nil, err
		}

		return &silapb.AttesterSlashingElectra{
			Attestation_1: att1,
			Attestation_2: att2,
		}, nil
	}

	att1 := &silapb.IndexedAttestation{
		Data: &silapb.AttestationData{
			Slot:            bState.Slot(),
			CommitteeIndex:  0,
			BeaconBlockRoot: make([]byte, fieldparams.RootLength),
			Target: &silapb.Checkpoint{
				Epoch: currentEpoch,
				Root:  params.BeaconConfig().ZeroHash[:],
			},
			Source: &silapb.Checkpoint{
				Epoch: currentEpoch + 1,
				Root:  params.BeaconConfig().ZeroHash[:],
			},
		},
		AttestingIndices: []uint64{uint64(idx)},
	}
	var err error
	att1.Signature, err = signing.ComputeDomainAndSign(bState, currentEpoch, att1.Data, params.BeaconConfig().DomainBeaconAttester, priv)
	if err != nil {
		return nil, err
	}

	att2 := &silapb.IndexedAttestation{
		Data: &silapb.AttestationData{
			Slot:            bState.Slot(),
			CommitteeIndex:  0,
			BeaconBlockRoot: make([]byte, fieldparams.RootLength),
			Target: &silapb.Checkpoint{
				Epoch: currentEpoch,
				Root:  params.BeaconConfig().ZeroHash[:],
			},
			Source: &silapb.Checkpoint{
				Epoch: currentEpoch,
				Root:  params.BeaconConfig().ZeroHash[:],
			},
		},
		AttestingIndices: []uint64{uint64(idx)},
	}
	att2.Signature, err = signing.ComputeDomainAndSign(bState, currentEpoch, att2.Data, params.BeaconConfig().DomainBeaconAttester, priv)
	if err != nil {
		return nil, err
	}

	return &silapb.AttesterSlashing{
		Attestation_1: att1,
		Attestation_2: att2,
	}, nil
}

func generateAttesterSlashings(
	bState state.BeaconState,
	privs []bls.SecretKey,
	numSlashings uint64,
) ([]silapb.AttSlashing, error) {
	attesterSlashings := make([]silapb.AttSlashing, numSlashings)
	randGen := rand.NewDeterministicGenerator()
	for i := range numSlashings {
		committeeIndex := randGen.Uint64() % helpers.SlotCommitteeCount(uint64(bState.NumValidators()))
		committee, err := helpers.BeaconCommitteeFromState(context.Background(), bState, bState.Slot(), primitives.CommitteeIndex(committeeIndex))
		if err != nil {
			return nil, err
		}
		randIndex := randGen.Uint64() % uint64(len(committee))
		valIndex := committee[randIndex]
		slashing, err := GenerateAttesterSlashingForValidator(bState, privs[valIndex], valIndex)
		if err != nil {
			return nil, err
		}
		attesterSlashings[i] = slashing
	}
	return attesterSlashings, nil
}

func generateDepositsAndSilaData(
	bState state.BeaconState,
	numDeposits uint64,
) (
	[]*silapb.Deposit,
	*silapb.SilaData,
	error,
) {
	previousDepsLen := bState.SilaExecutionDepositIndex()
	currentDeposits, _, err := DeterministicDepositsAndKeys(previousDepsLen + numDeposits)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not get deposits")
	}
	silaexecData, err := DeterministicSilaData(len(currentDeposits))
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not get silaData")
	}
	return currentDeposits[previousDepsLen:], silaexecData, nil
}

func GenerateVoluntaryExits(bState state.BeaconState, k bls.SecretKey, idx primitives.ValidatorIndex) (*silapb.SignedVoluntaryExit, error) {
	currentEpoch := time.CurrentEpoch(bState)
	exit := &silapb.SignedVoluntaryExit{
		Exit: &silapb.VoluntaryExit{
			Epoch:          time.PrevEpoch(bState),
			ValidatorIndex: idx,
		},
	}
	var err error
	exit.Signature, err = signing.ComputeDomainAndSign(bState, currentEpoch, exit.Exit, params.BeaconConfig().DomainVoluntaryExit, k)
	if err != nil {
		return nil, err
	}
	return exit, nil
}

func generateVoluntaryExits(
	bState state.BeaconState,
	privs []bls.SecretKey,
	numExits uint64,
) ([]*silapb.SignedVoluntaryExit, error) {
	currentEpoch := time.CurrentEpoch(bState)

	voluntaryExits := make([]*silapb.SignedVoluntaryExit, numExits)
	valMap := map[primitives.ValidatorIndex]bool{}
	for i := 0; i < len(voluntaryExits); i++ {
		valIndex, err := randValIndex(bState)
		if err != nil {
			return nil, err
		}
		// Retry if validator exit already exists.
		if valMap[valIndex] {
			i--
			continue
		}
		exit := &silapb.SignedVoluntaryExit{
			Exit: &silapb.VoluntaryExit{
				Epoch:          time.PrevEpoch(bState),
				ValidatorIndex: valIndex,
			},
		}
		exit.Signature, err = signing.ComputeDomainAndSign(bState, currentEpoch, exit.Exit, params.BeaconConfig().DomainVoluntaryExit, privs[valIndex])
		if err != nil {
			return nil, err
		}
		voluntaryExits[i] = exit
		valMap[valIndex] = true
	}
	return voluntaryExits, nil
}

func randValIndex(bState state.BeaconState) (primitives.ValidatorIndex, error) {
	activeCount, err := helpers.ActiveValidatorCount(context.Background(), bState, time.CurrentEpoch(bState))
	if err != nil {
		return 0, err
	}
	return primitives.ValidatorIndex(rand.NewGenerator().Uint64() % activeCount), nil
}

// HydrateSignedBeaconHeader hydrates a signed beacon block header with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBeaconHeader(h *silapb.SignedBeaconBlockHeader) *silapb.SignedBeaconBlockHeader {
	if h.Signature == nil {
		h.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	h.Header = HydrateBeaconHeader(h.Header)
	return h
}

// HydrateBeaconHeader hydrates a beacon block header with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconHeader(h *silapb.BeaconBlockHeader) *silapb.BeaconBlockHeader {
	if h == nil {
		h = &silapb.BeaconBlockHeader{}
	}
	if h.BodyRoot == nil {
		h.BodyRoot = make([]byte, fieldparams.RootLength)
	}
	if h.StateRoot == nil {
		h.StateRoot = make([]byte, fieldparams.RootLength)
	}
	if h.ParentRoot == nil {
		h.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	return h
}

// HydrateSignedBeaconBlock hydrates a signed beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBeaconBlock(b *silapb.SignedBeaconBlock) *silapb.SignedBeaconBlock {
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Block = HydrateBeaconBlock(b.Block)
	return b
}

// HydrateBeaconBlock hydrates a beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlock(b *silapb.BeaconBlock) *silapb.BeaconBlock {
	if b == nil {
		b = &silapb.BeaconBlock{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBeaconBlockBody(b.Body)
	return b
}

// HydrateBeaconBlockBody hydrates a beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockBody(b *silapb.BeaconBlockBody) *silapb.BeaconBlockBody {
	if b == nil {
		b = &silapb.BeaconBlockBody{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, fieldparams.RootLength)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, fieldparams.RootLength),
		}
	}
	return b
}

// HydrateV1SignedBeaconBlock hydrates a signed beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateV1SignedBeaconBlock(b *v1.SignedBeaconBlock) *v1.SignedBeaconBlock {
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Block = HydrateV1BeaconBlock(b.Block)
	return b
}

// HydrateV1BeaconBlock hydrates a beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateV1BeaconBlock(b *v1.BeaconBlock) *v1.BeaconBlock {
	if b == nil {
		b = &v1.BeaconBlock{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateV1BeaconBlockBody(b.Body)
	return b
}

// HydrateV1BeaconBlockBody hydrates a beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateV1BeaconBlockBody(b *v1.BeaconBlockBody) *v1.BeaconBlockBody {
	if b == nil {
		b = &v1.BeaconBlockBody{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, fieldparams.RootLength)
	}
	if b.SilaData == nil {
		b.SilaData = &v1.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, fieldparams.RootLength),
		}
	}
	return b
}

func SaveBlock(tb assertions.AssertionTestingTB, ctx context.Context, db iface.NoHeadAccessDatabase, b any) interfaces.SignedBeaconBlock {
	wsb, err := blocks.NewSignedBeaconBlock(b)
	require.NoError(tb, err)
	require.NoError(tb, db.SaveBlock(ctx, wsb))
	return wsb
}

// ----------------------------------------------------------------------------
// Altair
// ----------------------------------------------------------------------------

// HydrateSignedBeaconBlockAltair hydrates a signed beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBeaconBlockAltair(b *silapb.SignedBeaconBlockAltair) *silapb.SignedBeaconBlockAltair {
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Block = HydrateBeaconBlockAltair(b.Block)
	return b
}

// HydrateBeaconBlockAltair hydrates a beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockAltair(b *silapb.BeaconBlockAltair) *silapb.BeaconBlockAltair {
	if b == nil {
		b = &silapb.BeaconBlockAltair{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBeaconBlockBodyAltair(b.Body)
	return b
}

// HydrateBeaconBlockBodyAltair hydrates a beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockBodyAltair(b *silapb.BeaconBlockBodyAltair) *silapb.BeaconBlockBodyAltair {
	if b == nil {
		b = &silapb.BeaconBlockBodyAltair{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, fieldparams.RootLength)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, fieldparams.RootLength),
		}
	}
	if b.SyncAggregate == nil {
		b.SyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, 64),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		}
	}
	return b
}

// ----------------------------------------------------------------------------
// Bellatrix
// ----------------------------------------------------------------------------

// HydrateSignedBeaconBlockBellatrix hydrates a signed beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBeaconBlockBellatrix(b *silapb.SignedBeaconBlockBellatrix) *silapb.SignedBeaconBlockBellatrix {
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Block = HydrateBeaconBlockBellatrix(b.Block)
	return b
}

// HydrateBeaconBlockBellatrix hydrates a beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockBellatrix(b *silapb.BeaconBlockBellatrix) *silapb.BeaconBlockBellatrix {
	if b == nil {
		b = &silapb.BeaconBlockBellatrix{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBeaconBlockBodyBellatrix(b.Body)
	return b
}

// HydrateBeaconBlockBodyBellatrix hydrates a beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockBodyBellatrix(b *silapb.BeaconBlockBodyBellatrix) *silapb.BeaconBlockBodyBellatrix {
	if b == nil {
		b = &silapb.BeaconBlockBodyBellatrix{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, fieldparams.RootLength)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, fieldparams.RootLength),
		}
	}
	if b.SyncAggregate == nil {
		b.SyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, fieldparams.SyncAggregateSyncCommitteeBytesLength),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		}
	}
	if b.SilaPayload == nil {
		b.SilaPayload = &silaenginev1.SilaPayload{
			ParentHash:    make([]byte, fieldparams.RootLength),
			FeeRecipient:  make([]byte, 20),
			StateRoot:     make([]byte, fieldparams.RootLength),
			ReceiptsRoot:  make([]byte, fieldparams.RootLength),
			LogsBloom:     make([]byte, 256),
			PrevRandao:    make([]byte, fieldparams.RootLength),
			ExtraData:     make([]byte, 0),
			BaseFeePerGas: make([]byte, fieldparams.RootLength),
			BlockHash:     make([]byte, fieldparams.RootLength),
			Transactions:  make([][]byte, 0),
		}
	}
	return b
}

// HydrateSignedBlindedBeaconBlockBellatrix hydrates a signed blinded beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBlindedBeaconBlockBellatrix(b *silapb.SignedBlindedBeaconBlockBellatrix) *silapb.SignedBlindedBeaconBlockBellatrix {
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Block = HydrateBlindedBeaconBlockBellatrix(b.Block)
	return b
}

// HydrateBlindedBeaconBlockBellatrix hydrates a blinded beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBlindedBeaconBlockBellatrix(b *silapb.BlindedBeaconBlockBellatrix) *silapb.BlindedBeaconBlockBellatrix {
	if b == nil {
		b = &silapb.BlindedBeaconBlockBellatrix{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBlindedBeaconBlockBodyBellatrix(b.Body)
	return b
}

// HydrateBlindedBeaconBlockBodyBellatrix hydrates a blinded beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBlindedBeaconBlockBodyBellatrix(b *silapb.BlindedBeaconBlockBodyBellatrix) *silapb.BlindedBeaconBlockBodyBellatrix {
	if b == nil {
		b = &silapb.BlindedBeaconBlockBodyBellatrix{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, 32)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		}
	}
	if b.SyncAggregate == nil {
		b.SyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, fieldparams.SyncAggregateSyncCommitteeBytesLength),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		}
	}
	if b.SilaPayloadHeader == nil {
		b.SilaPayloadHeader = &silaenginev1.SilaPayloadHeader{
			ParentHash:       make([]byte, 32),
			FeeRecipient:     make([]byte, 20),
			StateRoot:        make([]byte, fieldparams.RootLength),
			ReceiptsRoot:     make([]byte, fieldparams.RootLength),
			LogsBloom:        make([]byte, 256),
			PrevRandao:       make([]byte, 32),
			ExtraData:        make([]byte, 0),
			BaseFeePerGas:    make([]byte, 32),
			BlockHash:        make([]byte, 32),
			TransactionsRoot: make([]byte, fieldparams.RootLength),
		}
	}
	return b
}

// ----------------------------------------------------------------------------
// Capella
// ----------------------------------------------------------------------------

// HydrateSignedBeaconBlockCapella hydrates a signed beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBeaconBlockCapella(b *silapb.SignedBeaconBlockCapella) *silapb.SignedBeaconBlockCapella {
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Block = HydrateBeaconBlockCapella(b.Block)
	return b
}

// HydrateBeaconBlockCapella hydrates a beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockCapella(b *silapb.BeaconBlockCapella) *silapb.BeaconBlockCapella {
	if b == nil {
		b = &silapb.BeaconBlockCapella{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBeaconBlockBodyCapella(b.Body)
	return b
}

// HydrateBeaconBlockBodyCapella hydrates a beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockBodyCapella(b *silapb.BeaconBlockBodyCapella) *silapb.BeaconBlockBodyCapella {
	if b == nil {
		b = &silapb.BeaconBlockBodyCapella{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, fieldparams.RootLength)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, fieldparams.RootLength),
		}
	}
	if b.SyncAggregate == nil {
		b.SyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, fieldparams.SyncAggregateSyncCommitteeBytesLength),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		}
	}
	if b.SilaPayload == nil {
		b.SilaPayload = &silaenginev1.SilaPayloadCapella{
			ParentHash:    make([]byte, fieldparams.RootLength),
			FeeRecipient:  make([]byte, 20),
			StateRoot:     make([]byte, fieldparams.RootLength),
			ReceiptsRoot:  make([]byte, fieldparams.RootLength),
			LogsBloom:     make([]byte, 256),
			PrevRandao:    make([]byte, fieldparams.RootLength),
			ExtraData:     make([]byte, 0),
			BaseFeePerGas: make([]byte, fieldparams.RootLength),
			BlockHash:     make([]byte, fieldparams.RootLength),
			Transactions:  make([][]byte, 0),
			Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
		}
	}
	return b
}

// HydrateSignedBlindedBeaconBlockCapella hydrates a signed blinded beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBlindedBeaconBlockCapella(b *silapb.SignedBlindedBeaconBlockCapella) *silapb.SignedBlindedBeaconBlockCapella {
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Block = HydrateBlindedBeaconBlockCapella(b.Block)
	return b
}

// HydrateBlindedBeaconBlockCapella hydrates a blinded beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBlindedBeaconBlockCapella(b *silapb.BlindedBeaconBlockCapella) *silapb.BlindedBeaconBlockCapella {
	if b == nil {
		b = &silapb.BlindedBeaconBlockCapella{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBlindedBeaconBlockBodyCapella(b.Body)
	return b
}

// HydrateBlindedBeaconBlockBodyCapella hydrates a blinded beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBlindedBeaconBlockBodyCapella(b *silapb.BlindedBeaconBlockBodyCapella) *silapb.BlindedBeaconBlockBodyCapella {
	if b == nil {
		b = &silapb.BlindedBeaconBlockBodyCapella{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, 32)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		}
	}
	if b.SyncAggregate == nil {
		b.SyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, fieldparams.SyncAggregateSyncCommitteeBytesLength),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		}
	}
	if b.SilaPayloadHeader == nil {
		b.SilaPayloadHeader = &silaenginev1.SilaPayloadHeaderCapella{
			ParentHash:       make([]byte, 32),
			FeeRecipient:     make([]byte, 20),
			StateRoot:        make([]byte, fieldparams.RootLength),
			ReceiptsRoot:     make([]byte, fieldparams.RootLength),
			LogsBloom:        make([]byte, 256),
			PrevRandao:       make([]byte, 32),
			ExtraData:        make([]byte, 0),
			BaseFeePerGas:    make([]byte, 32),
			BlockHash:        make([]byte, 32),
			TransactionsRoot: make([]byte, fieldparams.RootLength),
			WithdrawalsRoot:  make([]byte, fieldparams.RootLength),
		}
	}
	return b
}

// ----------------------------------------------------------------------------
// Deneb
// ----------------------------------------------------------------------------

// HydrateSignedBeaconBlockDeneb hydrates a signed beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBeaconBlockDeneb(b *silapb.SignedBeaconBlockDeneb) *silapb.SignedBeaconBlockDeneb {
	if b == nil {
		b = &silapb.SignedBeaconBlockDeneb{}
	}
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Block = HydrateBeaconBlockDeneb(b.Block)
	return b
}

// HydrateSignedBeaconBlockContentsDeneb hydrates a signed beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBeaconBlockContentsDeneb(b *silapb.SignedBeaconBlockContentsDeneb) *silapb.SignedBeaconBlockContentsDeneb {
	b.Block = HydrateSignedBeaconBlockDeneb(b.Block)
	return b
}

// HydrateBeaconBlockDeneb hydrates a beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockDeneb(b *silapb.BeaconBlockDeneb) *silapb.BeaconBlockDeneb {
	if b == nil {
		b = &silapb.BeaconBlockDeneb{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBeaconBlockBodyDeneb(b.Body)
	return b
}

// HydrateBeaconBlockBodyDeneb hydrates a beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockBodyDeneb(b *silapb.BeaconBlockBodyDeneb) *silapb.BeaconBlockBodyDeneb {
	if b == nil {
		b = &silapb.BeaconBlockBodyDeneb{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, fieldparams.RootLength)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, fieldparams.RootLength),
		}
	}
	if b.SyncAggregate == nil {
		b.SyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, fieldparams.SyncAggregateSyncCommitteeBytesLength),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		}
	}
	if b.SilaPayload == nil {
		b.SilaPayload = &silaenginev1.SilaPayloadDeneb{
			ParentHash:    make([]byte, fieldparams.RootLength),
			FeeRecipient:  make([]byte, 20),
			StateRoot:     make([]byte, fieldparams.RootLength),
			ReceiptsRoot:  make([]byte, fieldparams.RootLength),
			LogsBloom:     make([]byte, 256),
			PrevRandao:    make([]byte, fieldparams.RootLength),
			ExtraData:     make([]byte, 0),
			BaseFeePerGas: make([]byte, fieldparams.RootLength),
			BlockHash:     make([]byte, fieldparams.RootLength),
			Transactions:  make([][]byte, 0),
			Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
		}
	}
	return b
}

// HydrateSignedBlindedBeaconBlockDeneb hydrates a signed blinded beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBlindedBeaconBlockDeneb(b *silapb.SignedBlindedBeaconBlockDeneb) *silapb.SignedBlindedBeaconBlockDeneb {
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Message = HydrateBlindedBeaconBlockDeneb(b.Message)
	return b
}

// HydrateBlindedBeaconBlockBodyDeneb hydrates a blinded beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBlindedBeaconBlockBodyDeneb(b *silapb.BlindedBeaconBlockBodyDeneb) *silapb.BlindedBeaconBlockBodyDeneb {
	if b == nil {
		b = &silapb.BlindedBeaconBlockBodyDeneb{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, 32)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		}
	}
	if b.SyncAggregate == nil {
		b.SyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, fieldparams.SyncAggregateSyncCommitteeBytesLength),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		}
	}
	if b.SilaPayloadHeader == nil {
		b.SilaPayloadHeader = &silaenginev1.SilaPayloadHeaderDeneb{
			ParentHash:       make([]byte, 32),
			FeeRecipient:     make([]byte, 20),
			StateRoot:        make([]byte, fieldparams.RootLength),
			ReceiptsRoot:     make([]byte, fieldparams.RootLength),
			LogsBloom:        make([]byte, 256),
			PrevRandao:       make([]byte, 32),
			ExtraData:        make([]byte, 0),
			BaseFeePerGas:    make([]byte, 32),
			BlockHash:        make([]byte, 32),
			TransactionsRoot: make([]byte, fieldparams.RootLength),
			WithdrawalsRoot:  make([]byte, fieldparams.RootLength),
		}
	}
	return b
}

// HydrateBlindedBeaconBlockDeneb hydrates a blinded beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBlindedBeaconBlockDeneb(b *silapb.BlindedBeaconBlockDeneb) *silapb.BlindedBeaconBlockDeneb {
	if b == nil {
		b = &silapb.BlindedBeaconBlockDeneb{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBlindedBeaconBlockBodyDeneb(b.Body)
	return b
}

// ----------------------------------------------------------------------------
// Electra
// ----------------------------------------------------------------------------

// HydrateSignedBeaconBlockElectra hydrates a signed beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBeaconBlockElectra(b *silapb.SignedBeaconBlockElectra) *silapb.SignedBeaconBlockElectra {
	if b == nil {
		b = &silapb.SignedBeaconBlockElectra{}
	}
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Block = HydrateBeaconBlockElectra(b.Block)
	return b
}

// HydrateSignedBeaconBlockContentsElectra hydrates a signed beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBeaconBlockContentsElectra(b *silapb.SignedBeaconBlockContentsElectra) *silapb.SignedBeaconBlockContentsElectra {
	b.Block = HydrateSignedBeaconBlockElectra(b.Block)
	return b
}

// HydrateBeaconBlockElectra hydrates a beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockElectra(b *silapb.BeaconBlockElectra) *silapb.BeaconBlockElectra {
	if b == nil {
		b = &silapb.BeaconBlockElectra{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBeaconBlockBodyElectra(b.Body)
	return b
}

// HydrateBeaconBlockBodyElectra hydrates a beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockBodyElectra(b *silapb.BeaconBlockBodyElectra) *silapb.BeaconBlockBodyElectra {
	if b == nil {
		b = &silapb.BeaconBlockBodyElectra{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, fieldparams.RootLength)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, fieldparams.RootLength),
		}
	}
	if b.SyncAggregate == nil {
		b.SyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, fieldparams.SyncAggregateSyncCommitteeBytesLength),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		}
	}
	if b.SilaPayload == nil {
		b.SilaPayload = &silaenginev1.SilaPayloadDeneb{
			ParentHash:    make([]byte, fieldparams.RootLength),
			FeeRecipient:  make([]byte, 20),
			StateRoot:     make([]byte, fieldparams.RootLength),
			ReceiptsRoot:  make([]byte, fieldparams.RootLength),
			LogsBloom:     make([]byte, 256),
			PrevRandao:    make([]byte, fieldparams.RootLength),
			ExtraData:     make([]byte, 0),
			BaseFeePerGas: make([]byte, fieldparams.RootLength),
			BlockHash:     make([]byte, fieldparams.RootLength),
			Transactions:  make([][]byte, 0),
			Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
		}
	}
	b.SilaRequests = HydrateSilaRequests(b.SilaRequests)
	return b
}

// HydrateSilaRequests fills the sila requests with the correct field
// lengths
func HydrateSilaRequests(e *silaenginev1.SilaRequests) *silaenginev1.SilaRequests {
	if e == nil {
		e = &silaenginev1.SilaRequests{}
	}
	if e.Deposits == nil {
		e.Deposits = make([]*silaenginev1.DepositRequest, 0)
	}
	if e.Withdrawals == nil {
		e.Withdrawals = make([]*silaenginev1.WithdrawalRequest, 0)
	}
	if e.Consolidations == nil {
		e.Consolidations = make([]*silaenginev1.ConsolidationRequest, 0)
	}
	return e
}

// HydrateSignedBlindedBeaconBlockElectra hydrates a signed blinded beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBlindedBeaconBlockElectra(b *silapb.SignedBlindedBeaconBlockElectra) *silapb.SignedBlindedBeaconBlockElectra {
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Message = HydrateBlindedBeaconBlockElectra(b.Message)
	return b
}

// HydrateBlindedBeaconBlockElectra hydrates a blinded beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBlindedBeaconBlockElectra(b *silapb.BlindedBeaconBlockElectra) *silapb.BlindedBeaconBlockElectra {
	if b == nil {
		b = &silapb.BlindedBeaconBlockElectra{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBlindedBeaconBlockBodyElectra(b.Body)
	return b
}

// HydrateBlindedBeaconBlockBodyElectra hydrates a blinded beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBlindedBeaconBlockBodyElectra(b *silapb.BlindedBeaconBlockBodyElectra) *silapb.BlindedBeaconBlockBodyElectra {
	if b == nil {
		b = &silapb.BlindedBeaconBlockBodyElectra{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, 32)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		}
	}
	if b.SyncAggregate == nil {
		b.SyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, fieldparams.SyncAggregateSyncCommitteeBytesLength),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		}
	}
	if b.SilaPayloadHeader == nil {
		b.SilaPayloadHeader = &silaenginev1.SilaPayloadHeaderDeneb{
			ParentHash:       make([]byte, 32),
			FeeRecipient:     make([]byte, 20),
			StateRoot:        make([]byte, fieldparams.RootLength),
			ReceiptsRoot:     make([]byte, fieldparams.RootLength),
			LogsBloom:        make([]byte, 256),
			PrevRandao:       make([]byte, 32),
			ExtraData:        make([]byte, 0),
			BaseFeePerGas:    make([]byte, 32),
			BlockHash:        make([]byte, 32),
			TransactionsRoot: make([]byte, fieldparams.RootLength),
			WithdrawalsRoot:  make([]byte, fieldparams.RootLength),
		}
	}
	b.SilaRequests = HydrateSilaRequests(b.SilaRequests)
	return b
}

func generateWithdrawals(
	bState state.BeaconState,
	privs []bls.SecretKey,
	numWithdrawals uint64,
) ([]*silaenginev1.Withdrawal, error) {
	withdrawalRequests := make([]*silaenginev1.Withdrawal, numWithdrawals)
	for i := range numWithdrawals {
		valIndex, err := randValIndex(bState)
		if err != nil {
			return nil, err
		}
		amount := uint64(10000)
		bal, err := bState.BalanceAtIndex(valIndex)
		if err != nil {
			return nil, err
		}
		amounts := []uint64{
			amount, // some smaller amount
			bal,    // the entire balance
		}
		// Get a random index
		nBig, err := rd.Int(rd.Reader, big.NewInt(int64(len(amounts))))
		if err != nil {
			return nil, err
		}
		randomIndex := nBig.Uint64()
		withdrawalRequests[i] = &silaenginev1.Withdrawal{
			ValidatorIndex: valIndex,
			Address:        make([]byte, common.AddressLength),
			Amount:         amounts[randomIndex],
		}
	}
	return withdrawalRequests, nil
}

// ----------------------------------------------------------------------------
// Fulu
// ----------------------------------------------------------------------------

// HydrateSignedBeaconBlockFulu hydrates a signed beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBeaconBlockFulu(b *silapb.SignedBeaconBlockFulu) *silapb.SignedBeaconBlockFulu {
	if b == nil {
		b = &silapb.SignedBeaconBlockFulu{}
	}
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Block = HydrateBeaconBlockFulu(b.Block)
	return b
}

// HydrateSignedBeaconBlockContentsFulu hydrates a signed beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBeaconBlockContentsFulu(b *silapb.SignedBeaconBlockContentsFulu) *silapb.SignedBeaconBlockContentsFulu {
	b.Block = HydrateSignedBeaconBlockFulu(b.Block)
	return b
}

// HydrateBeaconBlockFulu hydrates a beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockFulu(b *silapb.BeaconBlockElectra) *silapb.BeaconBlockElectra {
	if b == nil {
		b = &silapb.BeaconBlockElectra{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBeaconBlockBodyFulu(b.Body)
	return b
}

// HydrateBeaconBlockBodyFulu hydrates a beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockBodyFulu(b *silapb.BeaconBlockBodyElectra) *silapb.BeaconBlockBodyElectra {
	if b == nil {
		b = &silapb.BeaconBlockBodyElectra{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, fieldparams.RootLength)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, fieldparams.RootLength),
		}
	}
	if b.SyncAggregate == nil {
		b.SyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, fieldparams.SyncAggregateSyncCommitteeBytesLength),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		}
	}
	if b.SilaPayload == nil {
		b.SilaPayload = &silaenginev1.SilaPayloadDeneb{
			ParentHash:    make([]byte, fieldparams.RootLength),
			FeeRecipient:  make([]byte, 20),
			StateRoot:     make([]byte, fieldparams.RootLength),
			ReceiptsRoot:  make([]byte, fieldparams.RootLength),
			LogsBloom:     make([]byte, 256),
			PrevRandao:    make([]byte, fieldparams.RootLength),
			ExtraData:     make([]byte, 0),
			BaseFeePerGas: make([]byte, fieldparams.RootLength),
			BlockHash:     make([]byte, fieldparams.RootLength),
			Transactions:  make([][]byte, 0),
			Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
		}
	}
	b.SilaRequests = HydrateSilaRequests(b.SilaRequests)
	return b
}

// HydrateSignedBlindedBeaconBlockFulu hydrates a signed blinded beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBlindedBeaconBlockFulu(b *silapb.SignedBlindedBeaconBlockFulu) *silapb.SignedBlindedBeaconBlockFulu {
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Message = HydrateBlindedBeaconBlockFulu(b.Message)
	return b
}

// HydrateBlindedBeaconBlockFulu hydrates a blinded beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBlindedBeaconBlockFulu(b *silapb.BlindedBeaconBlockFulu) *silapb.BlindedBeaconBlockFulu {
	if b == nil {
		b = &silapb.BlindedBeaconBlockFulu{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBlindedBeaconBlockBodyFulu(b.Body)
	return b
}

// HydrateBlindedBeaconBlockBodyFulu hydrates a blinded beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBlindedBeaconBlockBodyFulu(b *silapb.BlindedBeaconBlockBodyElectra) *silapb.BlindedBeaconBlockBodyElectra {
	if b == nil {
		b = &silapb.BlindedBeaconBlockBodyElectra{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, 32)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, 32),
		}
	}
	if b.SyncAggregate == nil {
		b.SyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, fieldparams.SyncAggregateSyncCommitteeBytesLength),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		}
	}
	if b.SilaPayloadHeader == nil {
		b.SilaPayloadHeader = &silaenginev1.SilaPayloadHeaderDeneb{
			ParentHash:       make([]byte, 32),
			FeeRecipient:     make([]byte, 20),
			StateRoot:        make([]byte, fieldparams.RootLength),
			ReceiptsRoot:     make([]byte, fieldparams.RootLength),
			LogsBloom:        make([]byte, 256),
			PrevRandao:       make([]byte, 32),
			ExtraData:        make([]byte, 0),
			BaseFeePerGas:    make([]byte, 32),
			BlockHash:        make([]byte, 32),
			TransactionsRoot: make([]byte, fieldparams.RootLength),
			WithdrawalsRoot:  make([]byte, fieldparams.RootLength),
		}
	}
	b.SilaRequests = HydrateSilaRequests(b.SilaRequests)
	return b
}

// HydrateSignedBeaconBlockGloas hydrates a signed beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedBeaconBlockGloas(b *silapb.SignedBeaconBlockGloas) *silapb.SignedBeaconBlockGloas {
	if b == nil {
		b = &silapb.SignedBeaconBlockGloas{}
	}
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Block = HydrateBeaconBlockGloas(b.Block)
	return b
}

// HydrateBeaconBlockGloas hydrates a beacon block with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockGloas(b *silapb.BeaconBlockGloas) *silapb.BeaconBlockGloas {
	if b == nil {
		b = &silapb.BeaconBlockGloas{}
	}
	if b.ParentRoot == nil {
		b.ParentRoot = make([]byte, fieldparams.RootLength)
	}
	if b.StateRoot == nil {
		b.StateRoot = make([]byte, fieldparams.RootLength)
	}
	b.Body = HydrateBeaconBlockBodyGloas(b.Body)
	return b
}

// HydrateBeaconBlockBodyGloas hydrates a beacon block body with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateBeaconBlockBodyGloas(b *silapb.BeaconBlockBodyGloas) *silapb.BeaconBlockBodyGloas {
	if b == nil {
		b = &silapb.BeaconBlockBodyGloas{}
	}
	if b.RandaoReveal == nil {
		b.RandaoReveal = make([]byte, fieldparams.BLSSignatureLength)
	}
	if b.Graffiti == nil {
		b.Graffiti = make([]byte, fieldparams.RootLength)
	}
	if b.SilaData == nil {
		b.SilaData = &silapb.SilaData{
			DepositRoot: make([]byte, fieldparams.RootLength),
			BlockHash:   make([]byte, fieldparams.RootLength),
		}
	}
	if b.SyncAggregate == nil {
		b.SyncAggregate = &silapb.SyncAggregate{
			SyncCommitteeBits:      make([]byte, fieldparams.SyncAggregateSyncCommitteeBytesLength),
			SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
		}
	}
	b.SignedSilaPayloadBid = HydrateSignedSilaPayloadBid(b.SignedSilaPayloadBid)
	if b.PayloadAttestations == nil {
		b.PayloadAttestations = make([]*silapb.PayloadAttestation, 0)
	}
	if b.ParentSilaRequests == nil {
		b.ParentSilaRequests = &silaenginev1.SilaRequests{}
	}
	return b
}

// HydrateSignedSilaPayloadBid hydrates a signed sila payload bid with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSignedSilaPayloadBid(b *silapb.SignedSilaPayloadBid) *silapb.SignedSilaPayloadBid {
	if b == nil {
		b = &silapb.SignedSilaPayloadBid{}
	}
	if b.Signature == nil {
		b.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	b.Message = HydrateSilaPayloadBid(b.Message)
	return b
}

// HydrateSilaPayloadBid hydrates an sila payload bid with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydrateSilaPayloadBid(b *silapb.SilaPayloadBid) *silapb.SilaPayloadBid {
	if b == nil {
		b = &silapb.SilaPayloadBid{}
	}
	if b.ParentBlockHash == nil {
		b.ParentBlockHash = make([]byte, fieldparams.RootLength)
	}
	if b.ParentBlockRoot == nil {
		b.ParentBlockRoot = make([]byte, fieldparams.RootLength)
	}
	if b.BlockHash == nil {
		b.BlockHash = make([]byte, fieldparams.RootLength)
	}
	if b.PrevRandao == nil {
		b.PrevRandao = make([]byte, fieldparams.RootLength)
	}
	if b.FeeRecipient == nil {
		b.FeeRecipient = make([]byte, fieldparams.FeeRecipientLength)
	}
	if b.BlobKzgCommitments == nil {
		b.BlobKzgCommitments = make([][]byte, 0)
	}
	if b.SilaRequestsRoot == nil {
		b.SilaRequestsRoot = make([]byte, fieldparams.RootLength)
	}
	return b
}

// HydratePayloadAttestation hydrates a payload attestation with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydratePayloadAttestation(p *silapb.PayloadAttestation) *silapb.PayloadAttestation {
	if p == nil {
		p = &silapb.PayloadAttestation{}
	}
	if p.AggregationBits == nil {
		p.AggregationBits = make([]byte, 64)
	}
	if p.Signature == nil {
		p.Signature = make([]byte, fieldparams.BLSSignatureLength)
	}
	p.Data = HydratePayloadAttestationData(p.Data)
	return p
}

// HydratePayloadAttestationData hydrates a payload attestation data with correct field length sizes
// to comply with fssz marshalling and unmarshalling rules.
func HydratePayloadAttestationData(d *silapb.PayloadAttestationData) *silapb.PayloadAttestationData {
	if d == nil {
		d = &silapb.PayloadAttestationData{}
	}
	if d.BeaconBlockRoot == nil {
		d.BeaconBlockRoot = make([]byte, fieldparams.RootLength)
	}
	return d
}

// GenerateTestPayloadAttestations generates a slice of payload attestations with non-zero test values.
// This is useful for testing Gloas-specific fields.
func GenerateTestPayloadAttestations(count int, slot primitives.Slot) []*silapb.PayloadAttestation {
	attestations := make([]*silapb.PayloadAttestation, count)
	for i := range count {
		aggregationBits := make([]byte, 64)
		aggregationBits[0] = 0x01 // Set at least one bit
		signature := make([]byte, fieldparams.BLSSignatureLength)
		signature[0] = byte(i + 1) // Make each signature unique
		beaconBlockRoot := make([]byte, fieldparams.RootLength)
		beaconBlockRoot[0] = byte(i + 1) // Make each root unique

		attestations[i] = &silapb.PayloadAttestation{
			AggregationBits: aggregationBits,
			Signature:       signature,
			Data: &silapb.PayloadAttestationData{
				BeaconBlockRoot:   beaconBlockRoot,
				Slot:              slot,
				PayloadPresent:    true,
				BlobDataAvailable: true,
			},
		}
	}
	return attestations
}

// GenerateTestSignedSilaPayloadBid generates a signed sila payload bid with non-zero test values.
// This is useful for testing Gloas-specific fields.
func GenerateTestSignedSilaPayloadBid(slot primitives.Slot) *silapb.SignedSilaPayloadBid {
	parentBlockHash := bytesutil.PadTo([]byte{0x01}, fieldparams.RootLength)
	parentBlockRoot := bytesutil.PadTo([]byte{0x02}, fieldparams.RootLength)
	blockHash := bytesutil.PadTo([]byte{0x03}, fieldparams.RootLength)
	prevRandao := bytesutil.PadTo([]byte{0x04}, fieldparams.RootLength)
	feeRecipient := bytesutil.PadTo([]byte{0x05}, fieldparams.FeeRecipientLength)
	blobKzgCommitment := bytesutil.PadTo([]byte{0x06}, fieldparams.BLSPubkeyLength)
	signature := bytesutil.PadTo([]byte{0x07}, fieldparams.BLSSignatureLength)

	return &silapb.SignedSilaPayloadBid{
		Message: &silapb.SilaPayloadBid{
			Slot:                  slot,
			BuilderIndex:          1,
			ParentBlockHash:       parentBlockHash,
			ParentBlockRoot:       parentBlockRoot,
			BlockHash:             blockHash,
			GasLimit:              30000000,
			PrevRandao:            prevRandao,
			FeeRecipient:          feeRecipient,
			Value:                 1000000,
			ExecutionPayment:      2000000,
			BlobKzgCommitments:    [][]byte{blobKzgCommitment},
			SilaRequestsRoot: make([]byte, fieldparams.RootLength),
		},
		Signature: signature,
	}
}
