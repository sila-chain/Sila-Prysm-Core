package transition_test

import (
	"fmt"
	"testing"

	"github.com/sila-chain/go-bitfield"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/time"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/transition"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	consensusblocks "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/attestation"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func init() {
	transition.SkipSlotCache.Disable()
}

func TestExecuteStateTransition_IncorrectSlot(t *testing.T) {
	base := &silapb.BeaconState{
		Slot: 5,
	}
	beaconState, err := state_native.InitializeFromProtoPhase0(base)
	require.NoError(t, err)
	block := &silapb.SignedBeaconBlock{
		Block: &silapb.BeaconBlock{
			Slot: 4,
			Body: &silapb.BeaconBlockBody{},
		},
	}
	want := "expected state.slot"
	wsb, err := consensusblocks.NewSignedBeaconBlock(block)
	require.NoError(t, err)
	_, err = transition.ExecuteStateTransition(t.Context(), beaconState, wsb)
	assert.ErrorContains(t, want, err)
}

func TestExecuteStateTransition_FullProcess(t *testing.T) {
	beaconState, privKeys := util.DeterministicGenesisState(t, 100)

	eth1Data := &silapb.Eth1Data{
		DepositCount: 100,
		DepositRoot:  bytesutil.PadTo([]byte{2}, 32),
		BlockHash:    make([]byte, 32),
	}
	require.NoError(t, beaconState.SetSlot(params.BeaconConfig().SlotsPerEpoch-1))
	e := beaconState.Eth1Data()
	e.DepositCount = 100
	require.NoError(t, beaconState.SetEth1Data(e))
	bh := beaconState.LatestBlockHeader()
	bh.Slot = beaconState.Slot()
	require.NoError(t, beaconState.SetLatestBlockHeader(bh))
	require.NoError(t, beaconState.SetEth1DataVotes([]*silapb.Eth1Data{eth1Data}))

	oldMix, err := beaconState.RandaoMixAtIndex(1)
	require.NoError(t, err)

	require.NoError(t, beaconState.SetSlot(beaconState.Slot()+1))
	epoch := time.CurrentEpoch(beaconState)
	randaoReveal, err := util.RandaoReveal(beaconState, epoch, privKeys)
	require.NoError(t, err)
	require.NoError(t, beaconState.SetSlot(beaconState.Slot()-1))

	nextSlotState, err := transition.ProcessSlots(t.Context(), beaconState.Copy(), beaconState.Slot()+1)
	require.NoError(t, err)
	parentRoot, err := nextSlotState.LatestBlockHeader().HashTreeRoot()
	require.NoError(t, err)
	proposerIdx, err := helpers.BeaconProposerIndex(t.Context(), nextSlotState)
	require.NoError(t, err)
	block := util.NewBeaconBlock()
	block.Block.ProposerIndex = proposerIdx
	block.Block.Slot = beaconState.Slot() + 1
	block.Block.ParentRoot = parentRoot[:]
	block.Block.Body.RandaoReveal = randaoReveal
	block.Block.Body.Eth1Data = eth1Data

	wsb, err := consensusblocks.NewSignedBeaconBlock(block)
	require.NoError(t, err)
	stateRoot, err := transition.CalculateStateRoot(t.Context(), beaconState, wsb)
	require.NoError(t, err)

	block.Block.StateRoot = stateRoot[:]

	sig, err := util.BlockSignature(beaconState, block.Block, privKeys)
	require.NoError(t, err)
	block.Signature = sig.Marshal()

	wsb, err = consensusblocks.NewSignedBeaconBlock(block)
	require.NoError(t, err)
	beaconState, err = transition.ExecuteStateTransition(t.Context(), beaconState, wsb)
	require.NoError(t, err)

	assert.Equal(t, params.BeaconConfig().SlotsPerEpoch, beaconState.Slot(), "Unexpected Slot number")

	mix, err := beaconState.RandaoMixAtIndex(1)
	require.NoError(t, err)
	assert.DeepNotEqual(t, oldMix, mix, "Did not expect new and old randao mix to equal")
}

func TestProcessBlock_IncorrectProcessExits(t *testing.T) {
	beaconState, _ := util.DeterministicGenesisState(t, 100)

	proposerSlashings := []*silapb.ProposerSlashing{
		{
			Header_1: util.HydrateSignedBeaconHeader(&silapb.SignedBeaconBlockHeader{
				Header: &silapb.BeaconBlockHeader{
					ProposerIndex: 3,
					Slot:          1,
				},
				Signature: bytesutil.PadTo([]byte("A"), 96),
			}),
			Header_2: util.HydrateSignedBeaconHeader(&silapb.SignedBeaconBlockHeader{
				Header: &silapb.BeaconBlockHeader{
					ProposerIndex: 3,
					Slot:          1,
				},
				Signature: bytesutil.PadTo([]byte("B"), 96),
			}),
		},
	}
	attesterSlashings := []*silapb.AttesterSlashing{
		{
			Attestation_1: &silapb.IndexedAttestation{
				Data:             util.HydrateAttestationData(&silapb.AttestationData{}),
				AttestingIndices: []uint64{0, 1},
				Signature:        make([]byte, 96),
			},
			Attestation_2: &silapb.IndexedAttestation{
				Data:             util.HydrateAttestationData(&silapb.AttestationData{}),
				AttestingIndices: []uint64{0, 1},
				Signature:        make([]byte, 96),
			},
		},
	}
	var blockRoots [][]byte
	for i := uint64(0); i < uint64(params.BeaconConfig().SlotsPerHistoricalRoot); i++ {
		blockRoots = append(blockRoots, []byte{byte(i)})
	}
	require.NoError(t, beaconState.SetBlockRoots(blockRoots))
	blockAtt := util.HydrateAttestation(&silapb.Attestation{
		Data: &silapb.AttestationData{
			Target: &silapb.Checkpoint{Root: bytesutil.PadTo([]byte("hello-world"), 32)},
		},
		AggregationBits: bitfield.Bitlist{0xC0, 0xC0, 0xC0, 0xC0, 0x01},
	})
	attestations := []*silapb.Attestation{blockAtt}
	var exits []*silapb.SignedVoluntaryExit
	for i := uint64(0); i < params.BeaconConfig().MaxVoluntaryExits+1; i++ {
		exits = append(exits, &silapb.SignedVoluntaryExit{})
	}
	genesisBlock := blocks.NewGenesisBlock([]byte{})
	bodyRoot, err := genesisBlock.Block.HashTreeRoot()
	require.NoError(t, err)
	err = beaconState.SetLatestBlockHeader(util.HydrateBeaconHeader(&silapb.BeaconBlockHeader{
		Slot:       genesisBlock.Block.Slot,
		ParentRoot: genesisBlock.Block.ParentRoot,
		BodyRoot:   bodyRoot[:],
	}))
	require.NoError(t, err)
	parentRoot, err := beaconState.LatestBlockHeader().HashTreeRoot()
	require.NoError(t, err)
	block := util.NewBeaconBlock()
	block.Block.Slot = 1
	block.Block.ParentRoot = parentRoot[:]
	block.Block.Body.ProposerSlashings = proposerSlashings
	block.Block.Body.Attestations = attestations
	block.Block.Body.AttesterSlashings = attesterSlashings
	block.Block.Body.VoluntaryExits = exits
	block.Block.Body.Eth1Data.DepositRoot = bytesutil.PadTo([]byte{2}, 32)
	block.Block.Body.Eth1Data.BlockHash = bytesutil.PadTo([]byte{3}, 32)
	err = beaconState.SetSlot(beaconState.Slot() + params.BeaconConfig().MinAttestationInclusionDelay)
	require.NoError(t, err)
	cp := beaconState.CurrentJustifiedCheckpoint()
	cp.Root = []byte("hello-world")
	require.NoError(t, beaconState.SetCurrentJustifiedCheckpoint(cp))
	require.NoError(t, beaconState.AppendCurrentEpochAttestations(&silapb.PendingAttestation{}))
	wsb, err := consensusblocks.NewSignedBeaconBlock(block)
	require.NoError(t, err)
	_, err = transition.VerifyOperationLengths(t.Context(), beaconState, wsb.Block())
	wanted := "number of voluntary exits (17) in block body exceeds allowed threshold of 16"
	assert.ErrorContains(t, wanted, err)
}

func createFullBlockWithOperations(t *testing.T) (state.BeaconState,
	*silapb.SignedBeaconBlock, []*silapb.Attestation, []*silapb.ProposerSlashing, []*silapb.SignedVoluntaryExit) {
	beaconState, privKeys := util.DeterministicGenesisState(t, 32)
	genesisBlock := blocks.NewGenesisBlock([]byte{})
	bodyRoot, err := genesisBlock.Block.HashTreeRoot()
	require.NoError(t, err)
	err = beaconState.SetLatestBlockHeader(&silapb.BeaconBlockHeader{
		Slot:       genesisBlock.Block.Slot,
		ParentRoot: genesisBlock.Block.ParentRoot,
		StateRoot:  params.BeaconConfig().ZeroHash[:],
		BodyRoot:   bodyRoot[:],
	})
	require.NoError(t, err)
	err = beaconState.SetSlashings(make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector))
	require.NoError(t, err)
	cp := beaconState.CurrentJustifiedCheckpoint()
	var mockRoot [32]byte
	copy(mockRoot[:], "hello-world")
	cp.Root = mockRoot[:]
	require.NoError(t, beaconState.SetCurrentJustifiedCheckpoint(cp))
	require.NoError(t, beaconState.AppendCurrentEpochAttestations(&silapb.PendingAttestation{}))

	proposerSlashIdx := primitives.ValidatorIndex(3)
	slotsPerEpoch := params.BeaconConfig().SlotsPerEpoch
	err = beaconState.SetSlot(slotsPerEpoch.Mul(uint64(params.BeaconConfig().ShardCommitteePeriod)) + params.BeaconConfig().MinAttestationInclusionDelay)
	require.NoError(t, err)

	currentEpoch := time.CurrentEpoch(beaconState)
	header1 := util.HydrateSignedBeaconHeader(&silapb.SignedBeaconBlockHeader{
		Header: &silapb.BeaconBlockHeader{
			ProposerIndex: proposerSlashIdx,
			Slot:          1,
			StateRoot:     bytesutil.PadTo([]byte("A"), 32),
		},
	})
	header1.Signature, err = signing.ComputeDomainAndSign(beaconState, currentEpoch, header1.Header, params.BeaconConfig().DomainBeaconProposer, privKeys[proposerSlashIdx])
	require.NoError(t, err)

	header2 := util.HydrateSignedBeaconHeader(&silapb.SignedBeaconBlockHeader{
		Header: &silapb.BeaconBlockHeader{
			ProposerIndex: proposerSlashIdx,
			Slot:          1,
			StateRoot:     bytesutil.PadTo([]byte("B"), 32),
		},
	})
	header2.Signature, err = signing.ComputeDomainAndSign(beaconState, time.CurrentEpoch(beaconState), header2.Header, params.BeaconConfig().DomainBeaconProposer, privKeys[proposerSlashIdx])
	require.NoError(t, err)

	proposerSlashings := []*silapb.ProposerSlashing{
		{
			Header_1: header1,
			Header_2: header2,
		},
	}
	validators := beaconState.Validators()
	validators[proposerSlashIdx].PublicKey = privKeys[proposerSlashIdx].PublicKey().Marshal()
	require.NoError(t, beaconState.SetValidators(validators))

	mockRoot2 := [32]byte{'A'}
	att1 := util.HydrateIndexedAttestation(&silapb.IndexedAttestation{
		Data: &silapb.AttestationData{
			Source: &silapb.Checkpoint{Epoch: 0, Root: mockRoot2[:]},
		},
		AttestingIndices: []uint64{0, 1},
	})
	domain, err := signing.Domain(beaconState.Fork(), currentEpoch, params.BeaconConfig().DomainBeaconAttester, beaconState.GenesisValidatorsRoot())
	require.NoError(t, err)
	hashTreeRoot, err := signing.ComputeSigningRoot(att1.Data, domain)
	require.NoError(t, err)
	sig0 := privKeys[0].Sign(hashTreeRoot[:])
	sig1 := privKeys[1].Sign(hashTreeRoot[:])
	aggregateSig := bls.AggregateSignatures([]bls.Signature{sig0, sig1})
	att1.Signature = aggregateSig.Marshal()

	mockRoot3 := [32]byte{'B'}
	att2 := util.HydrateIndexedAttestation(&silapb.IndexedAttestation{
		Data: &silapb.AttestationData{
			Source: &silapb.Checkpoint{Epoch: 0, Root: mockRoot3[:]},
			Target: &silapb.Checkpoint{Epoch: 0, Root: make([]byte, fieldparams.RootLength)},
		},
		AttestingIndices: []uint64{0, 1},
	})

	hashTreeRoot, err = signing.ComputeSigningRoot(att2.Data, domain)
	require.NoError(t, err)
	sig0 = privKeys[0].Sign(hashTreeRoot[:])
	sig1 = privKeys[1].Sign(hashTreeRoot[:])
	aggregateSig = bls.AggregateSignatures([]bls.Signature{sig0, sig1})
	att2.Signature = aggregateSig.Marshal()

	attesterSlashings := []*silapb.AttesterSlashing{
		{
			Attestation_1: att1,
			Attestation_2: att2,
		},
	}

	var blockRoots [][]byte
	for i := uint64(0); i < uint64(params.BeaconConfig().SlotsPerHistoricalRoot); i++ {
		blockRoots = append(blockRoots, []byte{byte(i)})
	}
	require.NoError(t, beaconState.SetBlockRoots(blockRoots))

	aggBits := bitfield.NewBitlist(1)
	aggBits.SetBitAt(0, true)
	blockAtt := util.HydrateAttestation(&silapb.Attestation{
		Data: &silapb.AttestationData{
			Slot:   beaconState.Slot(),
			Target: &silapb.Checkpoint{Epoch: time.CurrentEpoch(beaconState)},
			Source: &silapb.Checkpoint{Root: mockRoot[:]}},
		AggregationBits: aggBits,
	})

	committee, err := helpers.BeaconCommitteeFromState(t.Context(), beaconState, blockAtt.Data.Slot, blockAtt.Data.CommitteeIndex)
	assert.NoError(t, err)
	attestingIndices, err := attestation.AttestingIndices(blockAtt, committee)
	require.NoError(t, err)
	assert.NoError(t, err)
	hashTreeRoot, err = signing.ComputeSigningRoot(blockAtt.Data, domain)
	assert.NoError(t, err)
	sigs := make([]bls.Signature, len(attestingIndices))
	for i, indice := range attestingIndices {
		sig := privKeys[indice].Sign(hashTreeRoot[:])
		sigs[i] = sig
	}
	blockAtt.Signature = bls.AggregateSignatures(sigs).Marshal()

	exit := &silapb.SignedVoluntaryExit{
		Exit: &silapb.VoluntaryExit{
			ValidatorIndex: 10,
			Epoch:          0,
		},
	}
	exit.Signature, err = signing.ComputeDomainAndSign(beaconState, currentEpoch, exit.Exit, params.BeaconConfig().DomainVoluntaryExit, privKeys[exit.Exit.ValidatorIndex])
	require.NoError(t, err)

	header := beaconState.LatestBlockHeader()
	prevStateRoot, err := beaconState.HashTreeRoot(t.Context())
	require.NoError(t, err)
	header.StateRoot = prevStateRoot[:]
	require.NoError(t, beaconState.SetLatestBlockHeader(header))
	parentRoot, err := beaconState.LatestBlockHeader().HashTreeRoot()
	require.NoError(t, err)
	copied := beaconState.Copy()
	require.NoError(t, copied.SetSlot(beaconState.Slot()+1))
	randaoReveal, err := util.RandaoReveal(copied, currentEpoch, privKeys)
	require.NoError(t, err)
	proposerIndex, err := helpers.BeaconProposerIndex(t.Context(), copied)
	require.NoError(t, err)
	block := util.HydrateSignedBeaconBlock(&silapb.SignedBeaconBlock{
		Block: &silapb.BeaconBlock{
			ParentRoot:    parentRoot[:],
			Slot:          beaconState.Slot() + 1,
			ProposerIndex: proposerIndex,
			Body: &silapb.BeaconBlockBody{
				RandaoReveal:      randaoReveal,
				ProposerSlashings: proposerSlashings,
				AttesterSlashings: attesterSlashings,
				Attestations:      []*silapb.Attestation{blockAtt},
				VoluntaryExits:    []*silapb.SignedVoluntaryExit{exit},
			},
		},
	})

	sig, err := util.BlockSignature(beaconState, block.Block, privKeys)
	require.NoError(t, err)
	block.Signature = sig.Marshal()

	require.NoError(t, beaconState.SetSlot(block.Block.Slot))
	return beaconState, block, []*silapb.Attestation{blockAtt}, proposerSlashings, []*silapb.SignedVoluntaryExit{exit}
}

func TestProcessEpochPrecompute_CanProcess(t *testing.T) {
	epoch := primitives.Epoch(1)

	atts := []*silapb.PendingAttestation{{Data: &silapb.AttestationData{Target: &silapb.Checkpoint{Root: make([]byte, 32)}}, InclusionDelay: 1}}
	slashing := make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector)
	base := &silapb.BeaconState{
		Slot:                       params.BeaconConfig().SlotsPerEpoch.Mul(uint64(epoch)) + 1,
		BlockRoots:                 make([][]byte, 128),
		Slashings:                  slashing,
		RandaoMixes:                make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
		CurrentEpochAttestations:   atts,
		FinalizedCheckpoint:        &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		JustificationBits:          bitfield.Bitvector4{0x00},
		CurrentJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)},
		Validators: []*silapb.Validator{
			{
				ExitEpoch:        params.BeaconConfig().FarFutureEpoch,
				EffectiveBalance: params.BeaconConfig().MinDepositAmount,
			},
		},
		Balances: []uint64{
			params.BeaconConfig().MinDepositAmount,
		},
	}
	s, err := state_native.InitializeFromProtoPhase0(base)
	require.NoError(t, err)
	newState, err := transition.ProcessEpochPrecompute(t.Context(), s)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), newState.Slashings()[2], "Unexpected slashed balance")
}

func TestProcessBlock_OverMaxProposerSlashings(t *testing.T) {
	maxSlashings := params.BeaconConfig().MaxProposerSlashings
	b := &silapb.SignedBeaconBlock{
		Block: &silapb.BeaconBlock{
			Body: &silapb.BeaconBlockBody{
				ProposerSlashings: make([]*silapb.ProposerSlashing, maxSlashings+1),
			},
		},
	}
	want := fmt.Sprintf("number of proposer slashings (%d) in block body exceeds allowed threshold of %d",
		len(b.Block.Body.ProposerSlashings), params.BeaconConfig().MaxProposerSlashings)
	s, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	wsb, err := consensusblocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = transition.VerifyOperationLengths(t.Context(), s, wsb.Block())
	assert.ErrorContains(t, want, err)
}

func TestProcessBlock_OverMaxAttesterSlashings(t *testing.T) {
	maxSlashings := params.BeaconConfig().MaxAttesterSlashings
	b := &silapb.SignedBeaconBlock{
		Block: &silapb.BeaconBlock{
			Body: &silapb.BeaconBlockBody{
				AttesterSlashings: make([]*silapb.AttesterSlashing, maxSlashings+1),
			},
		},
	}
	want := fmt.Sprintf("number of attester slashings (%d) in block body exceeds allowed threshold of %d",
		len(b.Block.Body.AttesterSlashings), params.BeaconConfig().MaxAttesterSlashings)
	s, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	wsb, err := consensusblocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = transition.VerifyOperationLengths(t.Context(), s, wsb.Block())
	assert.ErrorContains(t, want, err)
}

func TestProcessBlock_OverMaxAttesterSlashingsElectra(t *testing.T) {
	maxSlashings := params.BeaconConfig().MaxAttesterSlashingsElectra
	b := &silapb.SignedBeaconBlockElectra{
		Block: &silapb.BeaconBlockElectra{
			Body: &silapb.BeaconBlockBodyElectra{
				AttesterSlashings: make([]*silapb.AttesterSlashingElectra, maxSlashings+1),
			},
		},
	}
	want := fmt.Sprintf("number of attester slashings (%d) in block body exceeds allowed threshold of %d",
		len(b.Block.Body.AttesterSlashings), params.BeaconConfig().MaxAttesterSlashingsElectra)
	s, err := state_native.InitializeFromProtoUnsafeElectra(&silapb.BeaconStateElectra{})
	require.NoError(t, err)
	wsb, err := consensusblocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = transition.VerifyOperationLengths(t.Context(), s, wsb.Block())
	assert.ErrorContains(t, want, err)
}

func TestProcessBlock_OverMaxAttestations(t *testing.T) {
	b := &silapb.SignedBeaconBlock{
		Block: &silapb.BeaconBlock{
			Body: &silapb.BeaconBlockBody{
				Attestations: make([]*silapb.Attestation, params.BeaconConfig().MaxAttestations+1),
			},
		},
	}
	want := fmt.Sprintf("number of attestations (%d) in block body exceeds allowed threshold of %d",
		len(b.Block.Body.Attestations), params.BeaconConfig().MaxAttestations)
	s, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	wsb, err := consensusblocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = transition.VerifyOperationLengths(t.Context(), s, wsb.Block())
	assert.ErrorContains(t, want, err)
}

func TestProcessBlock_OverMaxAttestationsElectra(t *testing.T) {
	b := &silapb.SignedBeaconBlockElectra{
		Block: &silapb.BeaconBlockElectra{
			Body: &silapb.BeaconBlockBodyElectra{
				Attestations: make([]*silapb.AttestationElectra, params.BeaconConfig().MaxAttestationsElectra+1),
			},
		},
	}
	want := fmt.Sprintf("number of attestations (%d) in block body exceeds allowed threshold of %d",
		len(b.Block.Body.Attestations), params.BeaconConfig().MaxAttestationsElectra)
	s, err := state_native.InitializeFromProtoUnsafeElectra(&silapb.BeaconStateElectra{})
	require.NoError(t, err)
	wsb, err := consensusblocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = transition.VerifyOperationLengths(t.Context(), s, wsb.Block())
	assert.ErrorContains(t, want, err)
}

func TestProcessBlock_OverMaxVoluntaryExits(t *testing.T) {
	maxExits := params.BeaconConfig().MaxVoluntaryExits
	b := &silapb.SignedBeaconBlock{
		Block: &silapb.BeaconBlock{
			Body: &silapb.BeaconBlockBody{
				VoluntaryExits: make([]*silapb.SignedVoluntaryExit, maxExits+1),
			},
		},
	}
	want := fmt.Sprintf("number of voluntary exits (%d) in block body exceeds allowed threshold of %d",
		len(b.Block.Body.VoluntaryExits), maxExits)
	s, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	wsb, err := consensusblocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = transition.VerifyOperationLengths(t.Context(), s, wsb.Block())
	assert.ErrorContains(t, want, err)
}

func TestProcessBlock_IncorrectDeposits(t *testing.T) {
	base := &silapb.BeaconState{
		Eth1Data:         &silapb.Eth1Data{DepositCount: 100},
		Eth1DepositIndex: 98,
	}
	s, err := state_native.InitializeFromProtoPhase0(base)
	require.NoError(t, err)
	b := &silapb.SignedBeaconBlock{
		Block: &silapb.BeaconBlock{
			Body: &silapb.BeaconBlockBody{
				Deposits: []*silapb.Deposit{{}},
			},
		},
	}
	want := fmt.Sprintf("incorrect outstanding deposits in block body, wanted: %d, got: %d",
		s.Eth1Data().DepositCount-s.Eth1DepositIndex(), len(b.Block.Body.Deposits))
	wsb, err := consensusblocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	_, err = transition.VerifyOperationLengths(t.Context(), s, wsb.Block())
	assert.ErrorContains(t, want, err)
}

func TestProcessSlots_SameSlotAsParentState(t *testing.T) {
	slot := primitives.Slot(2)
	parentState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{Slot: slot})
	require.NoError(t, err)

	_, err = transition.ProcessSlots(t.Context(), parentState, slot)
	assert.ErrorContains(t, "expected state.slot 2 < slot 2", err)
}

func TestProcessSlots_LowerSlotAsParentState(t *testing.T) {
	slot := primitives.Slot(2)
	parentState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{Slot: slot})
	require.NoError(t, err)

	_, err = transition.ProcessSlots(t.Context(), parentState, slot-1)
	assert.ErrorContains(t, "expected state.slot 2 < slot 1", err)
}

func TestProcessSlots_ThroughAltairEpoch(t *testing.T) {
	transition.SkipSlotCache.Disable()
	params.SetupTestConfigCleanup(t)
	conf := params.BeaconConfig()
	conf.AltairForkEpoch = 5
	params.OverrideBeaconConfig(conf)

	st, _ := util.DeterministicGenesisState(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	st, err := transition.ProcessSlots(t.Context(), st, params.BeaconConfig().SlotsPerEpoch*10)
	require.NoError(t, err)
	require.Equal(t, version.Altair, st.Version())

	require.Equal(t, params.BeaconConfig().SlotsPerEpoch*10, st.Slot())

	s, err := st.InactivityScores()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(s)))

	p, err := st.PreviousEpochParticipation()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(p)))

	p, err = st.CurrentEpochParticipation()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(p)))

	sc, err := st.CurrentSyncCommittee()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().SyncCommitteeSize, uint64(len(sc.Pubkeys)))

	sc, err = st.NextSyncCommittee()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().SyncCommitteeSize, uint64(len(sc.Pubkeys)))
}

func TestProcessSlots_OnlyAltairEpoch(t *testing.T) {
	transition.SkipSlotCache.Disable()
	params.SetupTestConfigCleanup(t)
	conf := params.BeaconConfig()
	conf.AltairForkEpoch = 5
	params.OverrideBeaconConfig(conf)

	st, _ := util.DeterministicGenesisStateAltair(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	require.NoError(t, st.SetSlot(params.BeaconConfig().SlotsPerEpoch*6))
	st, err := transition.ProcessSlots(t.Context(), st, params.BeaconConfig().SlotsPerEpoch*10)
	require.NoError(t, err)
	require.Equal(t, version.Altair, st.Version())

	require.Equal(t, params.BeaconConfig().SlotsPerEpoch*10, st.Slot())

	s, err := st.InactivityScores()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(s)))

	p, err := st.PreviousEpochParticipation()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(p)))

	p, err = st.CurrentEpochParticipation()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(p)))

	sc, err := st.CurrentSyncCommittee()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().SyncCommitteeSize, uint64(len(sc.Pubkeys)))

	sc, err = st.NextSyncCommittee()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().SyncCommitteeSize, uint64(len(sc.Pubkeys)))
}

func TestProcessSlots_OnlyBellatrixEpoch(t *testing.T) {
	transition.SkipSlotCache.Disable()
	params.SetupTestConfigCleanup(t)
	conf := params.BeaconConfig().Copy()
	conf.BellatrixForkEpoch = 5
	params.OverrideBeaconConfig(conf)

	st, _ := util.DeterministicGenesisStateBellatrix(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	require.NoError(t, st.SetSlot(params.BeaconConfig().SlotsPerEpoch*6))
	require.Equal(t, version.Bellatrix, st.Version())
	st, err := transition.ProcessSlots(t.Context(), st, params.BeaconConfig().SlotsPerEpoch*10)
	require.NoError(t, err)
	require.Equal(t, version.Bellatrix, st.Version())

	require.Equal(t, params.BeaconConfig().SlotsPerEpoch*10, st.Slot())

	s, err := st.InactivityScores()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(s)))

	p, err := st.PreviousEpochParticipation()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(p)))

	p, err = st.CurrentEpochParticipation()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().MaxValidatorsPerCommittee, uint64(len(p)))

	sc, err := st.CurrentSyncCommittee()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().SyncCommitteeSize, uint64(len(sc.Pubkeys)))

	sc, err = st.NextSyncCommittee()
	require.NoError(t, err)
	require.Equal(t, params.BeaconConfig().SyncCommitteeSize, uint64(len(sc.Pubkeys)))
}

func TestProcessSlots_ThroughBellatrixEpoch(t *testing.T) {
	transition.SkipSlotCache.Disable()
	params.SetupTestConfigCleanup(t)
	conf := params.BeaconConfig()
	conf.BellatrixForkEpoch = 5
	params.OverrideBeaconConfig(conf)

	st, _ := util.DeterministicGenesisStateAltair(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	st, err := transition.ProcessSlots(t.Context(), st, params.BeaconConfig().SlotsPerEpoch*10)
	require.NoError(t, err)
	require.Equal(t, version.Bellatrix, st.Version())

	require.Equal(t, params.BeaconConfig().SlotsPerEpoch*10, st.Slot())
}

func TestProcessSlots_ThroughDenebEpoch(t *testing.T) {
	transition.SkipSlotCache.Disable()
	params.SetupTestConfigCleanup(t)
	conf := params.BeaconConfig()
	conf.DenebForkEpoch = 5
	params.OverrideBeaconConfig(conf)

	st, _ := util.DeterministicGenesisStateCapella(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	st, err := transition.ProcessSlots(t.Context(), st, params.BeaconConfig().SlotsPerEpoch*10)
	require.NoError(t, err)
	require.Equal(t, version.Deneb, st.Version())
	require.Equal(t, params.BeaconConfig().SlotsPerEpoch*10, st.Slot())
}

func TestProcessSlots_ThroughElectraEpoch(t *testing.T) {
	transition.SkipSlotCache.Disable()
	params.SetupTestConfigCleanup(t)
	conf := params.BeaconConfig()
	conf.ElectraForkEpoch = 5
	params.OverrideBeaconConfig(conf)

	st, _ := util.DeterministicGenesisStateDeneb(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	st, err := transition.ProcessSlots(t.Context(), st, params.BeaconConfig().SlotsPerEpoch*10)
	require.NoError(t, err)
	require.Equal(t, version.Electra, st.Version())
	require.Equal(t, params.BeaconConfig().SlotsPerEpoch*10, st.Slot())
}

func TestProcessSlots_ThroughFuluEpoch(t *testing.T) {
	transition.SkipSlotCache.Disable()
	params.SetupTestConfigCleanup(t)
	conf := params.BeaconConfig()
	conf.FuluForkEpoch = 5
	params.OverrideBeaconConfig(conf)

	st, _ := util.DeterministicGenesisStateElectra(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	st, err := transition.ProcessSlots(t.Context(), st, params.BeaconConfig().SlotsPerEpoch*10)
	require.NoError(t, err)
	require.Equal(t, version.Fulu, st.Version())
	require.Equal(t, params.BeaconConfig().SlotsPerEpoch*10, st.Slot())
}

func TestProcessSlotsUsingNextSlotCache(t *testing.T) {
	s, _ := util.DeterministicGenesisState(t, 1)
	r := []byte{'a'}
	s, err := transition.ProcessSlotsUsingNextSlotCache(t.Context(), s, r, 5)
	require.NoError(t, err)
	require.Equal(t, primitives.Slot(5), s.Slot())
}

func TestProcessSlotsConditionally(t *testing.T) {
	ctx := t.Context()
	s, _ := util.DeterministicGenesisState(t, 1)

	t.Run("target slot below current slot", func(t *testing.T) {
		require.NoError(t, s.SetSlot(5))
		s, err := transition.ProcessSlotsIfPossible(ctx, s, 4)
		require.NoError(t, err)
		assert.Equal(t, primitives.Slot(5), s.Slot())
	})

	t.Run("target slot equal current slot", func(t *testing.T) {
		require.NoError(t, s.SetSlot(5))
		s, err := transition.ProcessSlotsIfPossible(ctx, s, 5)
		require.NoError(t, err)
		assert.Equal(t, primitives.Slot(5), s.Slot())
	})

	t.Run("target slot above current slot", func(t *testing.T) {
		require.NoError(t, s.SetSlot(5))
		s, err := transition.ProcessSlotsIfPossible(ctx, s, 6)
		require.NoError(t, err)
		assert.Equal(t, primitives.Slot(6), s.Slot())
	})
}

func BenchmarkProcessSlots_Capella(b *testing.B) {
	st, _ := util.DeterministicGenesisStateCapella(b, params.BeaconConfig().MaxValidatorsPerCommittee)

	var err error

	for b.Loop() {
		st, err = transition.ProcessSlots(b.Context(), st, st.Slot()+1)
		if err != nil {
			b.Fatalf("Failed to process slot %v", err)
		}
	}
}

func BenchmarkProcessSlots_Deneb(b *testing.B) {
	st, _ := util.DeterministicGenesisStateDeneb(b, params.BeaconConfig().MaxValidatorsPerCommittee)

	var err error

	for b.Loop() {
		st, err = transition.ProcessSlots(b.Context(), st, st.Slot()+1)
		if err != nil {
			b.Fatalf("Failed to process slot %v", err)
		}
	}
}

func BenchmarkProcessSlots_Electra(b *testing.B) {
	st, _ := util.DeterministicGenesisStateElectra(b, params.BeaconConfig().MaxValidatorsPerCommittee)

	var err error

	for b.Loop() {
		st, err = transition.ProcessSlots(b.Context(), st, st.Slot()+1)
		if err != nil {
			b.Fatalf("Failed to process slot %v", err)
		}
	}
}
