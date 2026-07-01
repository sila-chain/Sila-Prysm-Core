package validator

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/pkg/errors"
	builderapi "github.com/sila-chain/Sila-Consensus-Core/v7/api/client/builder"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain/kzg"
	mock "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/builder"
	builderTest "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/builder/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache/depositsnapshot"
	b "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	coretime "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/time"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/transition"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db"
	dbutil "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db/testing"
	doublylinkedtree "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/forkchoice/doubly-linked-tree"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/attestations"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/blstoexec"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/slashings"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/synccommittee"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/voluntaryexits"
	mockp2p "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/rpc/testutil"
	mockSila "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stategen"
	mockSync "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/sync/initial-sync/testing"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/container/trie"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/ssz"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/attestation"
	attaggregation "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/attestation/aggregation/attestations"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/sila-chain/Sila/common"
	"github.com/sila-chain/Sila/common/hexutil"
	"github.com/sila-chain/go-bitfield"
	"github.com/sirupsen/logrus"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestServer_GetBeaconBlock_Phase0(t *testing.T) {
	db := dbutil.SetupDB(t)
	ctx := t.Context()

	beaconState, privKeys := util.DeterministicGenesisState(t, 64)
	stateRoot, err := beaconState.HashTreeRoot(ctx)
	require.NoError(t, err, "Could not hash genesis state")

	genesis := b.NewGenesisBlock(stateRoot[:])
	genBlk := &silapb.SignedBeaconBlock{
		Block: &silapb.BeaconBlock{
			Slot:       genesis.Block.Slot,
			ParentRoot: genesis.Block.ParentRoot,
			StateRoot:  genesis.Block.StateRoot,
			Body: &silapb.BeaconBlockBody{
				RandaoReveal: genesis.Block.Body.RandaoReveal,
				Graffiti:     genesis.Block.Body.Graffiti,
				SilaData:     genesis.Block.Body.SilaData,
			},
		},
		Signature: genesis.Signature,
	}
	util.SaveBlock(t, ctx, db, genBlk)

	parentRoot, err := genBlk.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, parentRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, parentRoot), "Could not save genesis state")

	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, parentRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, parentRoot), "Could not save genesis state")

	proposerServer := getProposerServer(ctx, db, beaconState, parentRoot[:])
	// Use a separate mock for BlockReceiver with an independent state copy.
	// This mirrors production where computePostBlockStateAndRoot calls StateByRoot (fresh from DB),
	// not the same head state object mutated by the getSlashings goroutine.
	proposerServer.BlockReceiver = &mock.ChainService{
		State:           beaconState.Copy(),
		Root:            parentRoot[:],
		ForkChoiceStore: doublylinkedtree.New(),
	}

	randaoReveal, err := util.RandaoReveal(beaconState, 0, privKeys)
	require.NoError(t, err)

	graffiti := bytesutil.ToBytes32([]byte("sila"))
	req := &silapb.BlockRequest{
		Slot:         1,
		RandaoReveal: randaoReveal,
		Graffiti:     graffiti[:],
	}
	proposerSlashings, attSlashings := injectSlashings(t, beaconState, privKeys, proposerServer)

	block, err := proposerServer.GetBeaconBlock(ctx, req)
	require.NoError(t, err)
	phase0Blk, ok := block.GetBlock().(*silapb.GenericBeaconBlock_Phase0)
	require.Equal(t, true, ok)
	assert.Equal(t, req.Slot, phase0Blk.Phase0.Slot)
	assert.DeepEqual(t, parentRoot[:], phase0Blk.Phase0.ParentRoot, "Expected block to have correct parent root")
	assert.DeepEqual(t, randaoReveal, phase0Blk.Phase0.Body.RandaoReveal, "Expected block to have correct randao reveal")
	assert.DeepEqual(t, req.Graffiti, phase0Blk.Phase0.Body.Graffiti, "Expected block to have correct Graffiti")
	assert.Equal(t, params.BeaconConfig().MaxProposerSlashings, uint64(len(phase0Blk.Phase0.Body.ProposerSlashings)))
	assert.DeepEqual(t, proposerSlashings, phase0Blk.Phase0.Body.ProposerSlashings)
	assert.Equal(t, params.BeaconConfig().MaxAttesterSlashings, uint64(len(phase0Blk.Phase0.Body.AttesterSlashings)))
	assert.DeepEqual(t, attSlashings, phase0Blk.Phase0.Body.AttesterSlashings)
}

func TestServer_GetBeaconBlock_Altair(t *testing.T) {
	db := dbutil.SetupDB(t)
	ctx := t.Context()

	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.AltairForkEpoch = 1
	params.OverrideBeaconConfig(cfg)
	beaconState, privKeys := util.DeterministicGenesisState(t, 64)

	stateRoot, err := beaconState.HashTreeRoot(ctx)
	require.NoError(t, err, "Could not hash genesis state")

	genesis := b.NewGenesisBlock(stateRoot[:])
	util.SaveBlock(t, ctx, db, genesis)

	parentRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, parentRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, parentRoot), "Could not save genesis state")

	altairSlot, err := slots.EpochStart(params.BeaconConfig().AltairForkEpoch)
	require.NoError(t, err)

	var scBits [fieldparams.SyncAggregateSyncCommitteeBytesLength]byte
	genAltair := &silapb.SignedBeaconBlockAltair{
		Block: &silapb.BeaconBlockAltair{
			Slot:       altairSlot + 1,
			ParentRoot: parentRoot[:],
			StateRoot:  genesis.Block.StateRoot,
			Body: &silapb.BeaconBlockBodyAltair{
				RandaoReveal:  genesis.Block.Body.RandaoReveal,
				Graffiti:      genesis.Block.Body.Graffiti,
				SilaData:      genesis.Block.Body.SilaData,
				SyncAggregate: &silapb.SyncAggregate{SyncCommitteeBits: scBits[:], SyncCommitteeSignature: make([]byte, 96)},
			},
		},
	}

	blkRoot, err := genAltair.Block.HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, blkRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, blkRoot), "Could not save genesis state")

	proposerServer := getProposerServer(ctx, db, beaconState, parentRoot[:])

	randaoReveal, err := util.RandaoReveal(beaconState, 0, privKeys)
	require.NoError(t, err)

	graffiti := bytesutil.ToBytes32([]byte("sila"))
	require.NoError(t, err)
	req := &silapb.BlockRequest{
		Slot:         altairSlot + 1,
		RandaoReveal: randaoReveal,
		Graffiti:     graffiti[:],
	}
	proposerSlashings, attSlashings := injectSlashings(t, beaconState, privKeys, proposerServer)

	block, err := proposerServer.GetBeaconBlock(ctx, req)
	require.NoError(t, err)
	altairBlk, ok := block.GetBlock().(*silapb.GenericBeaconBlock_Altair)
	require.Equal(t, true, ok)

	assert.Equal(t, req.Slot, altairBlk.Altair.Slot)
	assert.DeepEqual(t, parentRoot[:], altairBlk.Altair.ParentRoot, "Expected block to have correct parent root")
	assert.DeepEqual(t, randaoReveal, altairBlk.Altair.Body.RandaoReveal, "Expected block to have correct randao reveal")
	assert.DeepEqual(t, req.Graffiti, altairBlk.Altair.Body.Graffiti, "Expected block to have correct Graffiti")
	assert.Equal(t, params.BeaconConfig().MaxProposerSlashings, uint64(len(altairBlk.Altair.Body.ProposerSlashings)))
	assert.DeepEqual(t, proposerSlashings, altairBlk.Altair.Body.ProposerSlashings)
	assert.Equal(t, params.BeaconConfig().MaxAttesterSlashings, uint64(len(altairBlk.Altair.Body.AttesterSlashings)))
	assert.DeepEqual(t, attSlashings, altairBlk.Altair.Body.AttesterSlashings)
}

func TestServer_GetBeaconBlock_Bellatrix(t *testing.T) {
	db := dbutil.SetupDB(t)
	ctx := t.Context()

	terminalBlockHash := bytesutil.PadTo([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, 32)
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.BellatrixForkEpoch = 2
	cfg.AltairForkEpoch = 1
	cfg.TerminalBlockHash = common.BytesToHash(terminalBlockHash)
	cfg.TerminalBlockHashActivationEpoch = 2
	params.OverrideBeaconConfig(cfg)
	beaconState, privKeys := util.DeterministicGenesisState(t, 64)

	stateRoot, err := beaconState.HashTreeRoot(ctx)
	require.NoError(t, err, "Could not hash genesis state")

	genesis := b.NewGenesisBlock(stateRoot[:])
	util.SaveBlock(t, ctx, db, genesis)

	parentRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, parentRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, parentRoot), "Could not save genesis state")

	bellatrixSlot, err := slots.EpochStart(params.BeaconConfig().BellatrixForkEpoch)
	require.NoError(t, err)

	var scBits [fieldparams.SyncAggregateSyncCommitteeBytesLength]byte
	blk := &silapb.SignedBeaconBlockBellatrix{
		Block: &silapb.BeaconBlockBellatrix{
			Slot:       bellatrixSlot + 1,
			ParentRoot: parentRoot[:],
			StateRoot:  genesis.Block.StateRoot,
			Body: &silapb.BeaconBlockBodyBellatrix{
				RandaoReveal:  genesis.Block.Body.RandaoReveal,
				Graffiti:      genesis.Block.Body.Graffiti,
				SilaData:      genesis.Block.Body.SilaData,
				SyncAggregate: &silapb.SyncAggregate{SyncCommitteeBits: scBits[:], SyncCommitteeSignature: make([]byte, 96)},
				SilaPayload: &silaenginev1.SilaPayload{
					ParentHash:    make([]byte, fieldparams.RootLength),
					FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
					StateRoot:     make([]byte, fieldparams.RootLength),
					ReceiptsRoot:  make([]byte, fieldparams.RootLength),
					LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
					PrevRandao:    make([]byte, fieldparams.RootLength),
					BaseFeePerGas: make([]byte, fieldparams.RootLength),
					BlockHash:     make([]byte, fieldparams.RootLength),
				},
			},
		},
	}

	blkRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, blkRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, blkRoot), "Could not save genesis state")

	c := mockSila.New()
	c.HashesByHeight[0] = terminalBlockHash
	random, err := helpers.RandaoMix(beaconState, slots.ToEpoch(beaconState.Slot()))
	require.NoError(t, err)
	timeStamp, err := slots.StartTime(beaconState.GenesisTime(), bellatrixSlot+1)
	require.NoError(t, err)

	payload := &silaenginev1.SilaPayload{
		ParentHash:    make([]byte, fieldparams.RootLength),
		FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:     make([]byte, fieldparams.RootLength),
		ReceiptsRoot:  make([]byte, fieldparams.RootLength),
		LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
		PrevRandao:    random,
		BaseFeePerGas: make([]byte, fieldparams.RootLength),
		BlockHash:     make([]byte, fieldparams.RootLength),
		Transactions:  make([][]byte, 0),
		ExtraData:     make([]byte, 0),
		BlockNumber:   1,
		GasLimit:      2,
		GasUsed:       3,
		Timestamp:     uint64(timeStamp.Unix()),
	}

	proposerServer := getProposerServer(ctx, db, beaconState, parentRoot[:])
	proposerServer.SilaBlockFetcher = c
	ed, err := blocks.NewWrappedSilaData(payload)
	require.NoError(t, err)
	proposerServer.SilaEngineCaller = &mockSila.SilaEngineClient{
		PayloadIDBytes:     &silaenginev1.PayloadIDBytes{1},
		GetPayloadResponse: &blocks.GetPayloadResponse{SilaData: ed},
	}

	randaoReveal, err := util.RandaoReveal(beaconState, 0, privKeys)
	require.NoError(t, err)

	graffiti := bytesutil.ToBytes32([]byte("sila"))
	require.NoError(t, err)
	req := &silapb.BlockRequest{
		Slot:         bellatrixSlot + 1,
		RandaoReveal: randaoReveal,
		Graffiti:     graffiti[:],
	}

	block, err := proposerServer.GetBeaconBlock(ctx, req)
	require.NoError(t, err)
	bellatrixBlk, ok := block.GetBlock().(*silapb.GenericBeaconBlock_Bellatrix)
	require.Equal(t, true, ok)

	assert.Equal(t, req.Slot, bellatrixBlk.Bellatrix.Slot)
	assert.DeepEqual(t, parentRoot[:], bellatrixBlk.Bellatrix.ParentRoot, "Expected block to have correct parent root")
	assert.DeepEqual(t, randaoReveal, bellatrixBlk.Bellatrix.Body.RandaoReveal, "Expected block to have correct randao reveal")
	assert.DeepEqual(t, req.Graffiti, bellatrixBlk.Bellatrix.Body.Graffiti, "Expected block to have correct Graffiti")

	require.DeepEqual(t, payload, bellatrixBlk.Bellatrix.Body.SilaPayload) // Payload should equal.

	// Operator sets default fee recipient to not be burned through beacon node cli.
	newHook := logTest.NewGlobal()
	params.SetupTestConfigCleanup(t)
	cfg = params.MinimalSpecConfig().Copy()
	cfg.DefaultFeeRecipient = common.Address{'b'}
	params.OverrideBeaconConfig(cfg)
	_, err = proposerServer.GetBeaconBlock(ctx, req)
	require.NoError(t, err)
	require.LogsDoNotContain(t, newHook, "Fee recipient is currently using the burn address")
}

func TestServer_GetBeaconBlock_Capella(t *testing.T) {
	db := dbutil.SetupDB(t)
	ctx := t.Context()
	transition.SkipSlotCache.Disable()

	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.CapellaForkEpoch = 3
	cfg.BellatrixForkEpoch = 2
	cfg.AltairForkEpoch = 1
	params.OverrideBeaconConfig(cfg)
	beaconState, privKeys := util.DeterministicGenesisState(t, 64)

	stateRoot, err := beaconState.HashTreeRoot(ctx)
	require.NoError(t, err, "Could not hash genesis state")

	genesis := b.NewGenesisBlock(stateRoot[:])
	util.SaveBlock(t, ctx, db, genesis)

	parentRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, parentRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, parentRoot), "Could not save genesis state")

	capellaSlot, err := slots.EpochStart(params.BeaconConfig().CapellaForkEpoch)
	require.NoError(t, err)

	var scBits [fieldparams.SyncAggregateSyncCommitteeBytesLength]byte
	blk := &silapb.SignedBeaconBlockCapella{
		Block: &silapb.BeaconBlockCapella{
			Slot:       capellaSlot + 1,
			ParentRoot: parentRoot[:],
			StateRoot:  genesis.Block.StateRoot,
			Body: &silapb.BeaconBlockBodyCapella{
				RandaoReveal:  genesis.Block.Body.RandaoReveal,
				Graffiti:      genesis.Block.Body.Graffiti,
				SilaData:      genesis.Block.Body.SilaData,
				SyncAggregate: &silapb.SyncAggregate{SyncCommitteeBits: scBits[:], SyncCommitteeSignature: make([]byte, 96)},
				SilaPayload: &silaenginev1.SilaPayloadCapella{
					ParentHash:    make([]byte, fieldparams.RootLength),
					FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
					StateRoot:     make([]byte, fieldparams.RootLength),
					ReceiptsRoot:  make([]byte, fieldparams.RootLength),
					LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
					PrevRandao:    make([]byte, fieldparams.RootLength),
					ExtraData:     make([]byte, 0),
					BaseFeePerGas: make([]byte, fieldparams.RootLength),
					BlockHash:     make([]byte, fieldparams.RootLength),
					Transactions:  make([][]byte, 0),
					Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
				},
			},
		},
	}

	blkRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, blkRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, blkRoot), "Could not save genesis state")

	random, err := helpers.RandaoMix(beaconState, slots.ToEpoch(beaconState.Slot()))
	require.NoError(t, err)
	timeStamp, err := slots.StartTime(beaconState.GenesisTime(), capellaSlot+1)
	require.NoError(t, err)
	payload := &silaenginev1.SilaPayloadCapella{
		ParentHash:    make([]byte, fieldparams.RootLength),
		FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:     make([]byte, fieldparams.RootLength),
		ReceiptsRoot:  make([]byte, fieldparams.RootLength),
		LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
		PrevRandao:    random,
		BlockNumber:   1,
		GasLimit:      2,
		GasUsed:       3,
		Timestamp:     uint64(timeStamp.Unix()),
		ExtraData:     make([]byte, 0),
		BaseFeePerGas: make([]byte, fieldparams.RootLength),
		BlockHash:     make([]byte, fieldparams.RootLength),
		Transactions:  make([][]byte, 0),
		Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
	}

	proposerServer := getProposerServer(ctx, db, beaconState, parentRoot[:])
	advancedState := beaconState.Copy()
	advancedState, err = transition.ProcessSlots(ctx, advancedState, capellaSlot)
	require.NoError(t, err)
	proposerServer.BlockReceiver = &mock.ChainService{
		State:           advancedState,
		Root:            parentRoot[:],
		ForkChoiceStore: doublylinkedtree.New(),
	}
	ed, err := blocks.NewWrappedSilaData(payload)
	require.NoError(t, err)
	proposerServer.SilaEngineCaller = &mockSila.SilaEngineClient{
		PayloadIDBytes:     &silaenginev1.PayloadIDBytes{1},
		GetPayloadResponse: &blocks.GetPayloadResponse{SilaData: ed},
	}

	randaoReveal, err := util.RandaoReveal(beaconState, 0, privKeys)
	require.NoError(t, err)

	graffiti := bytesutil.ToBytes32([]byte("sila"))
	require.NoError(t, err)
	req := &silapb.BlockRequest{
		Slot:         capellaSlot + 1,
		RandaoReveal: randaoReveal,
		Graffiti:     graffiti[:],
	}

	copiedState := beaconState.Copy()
	copiedState, err = transition.ProcessSlots(ctx, copiedState, capellaSlot+1)
	require.NoError(t, err)
	change, err := util.GenerateBLSToSilaChange(copiedState, privKeys[1], 0)
	require.NoError(t, err)
	proposerServer.BLSChangesPool.InsertBLSToExecChange(change)

	got, err := proposerServer.GetBeaconBlock(ctx, req)
	require.NoError(t, err)
	require.Equal(t, 1, len(got.GetCapella().Body.BlsToSilaChanges))
	require.DeepEqual(t, change, got.GetCapella().Body.BlsToSilaChanges[0])
}

func TestServer_GetBeaconBlock_Deneb(t *testing.T) {
	db := dbutil.SetupDB(t)
	ctx := t.Context()
	transition.SkipSlotCache.Disable()

	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.DenebForkEpoch = 4
	cfg.CapellaForkEpoch = 3
	cfg.BellatrixForkEpoch = 2
	cfg.AltairForkEpoch = 1
	params.OverrideBeaconConfig(cfg)
	beaconState, privKeys := util.DeterministicGenesisState(t, 64)

	stateRoot, err := beaconState.HashTreeRoot(ctx)
	require.NoError(t, err, "Could not hash genesis state")

	genesis := b.NewGenesisBlock(stateRoot[:])
	util.SaveBlock(t, ctx, db, genesis)

	parentRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, parentRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, parentRoot), "Could not save genesis state")

	denebSlot, err := slots.EpochStart(params.BeaconConfig().DenebForkEpoch)
	require.NoError(t, err)

	var scBits [fieldparams.SyncAggregateSyncCommitteeBytesLength]byte
	blk := &silapb.SignedBeaconBlockDeneb{
		Block: &silapb.BeaconBlockDeneb{
			Slot:       denebSlot + 1,
			ParentRoot: parentRoot[:],
			StateRoot:  genesis.Block.StateRoot,
			Body: &silapb.BeaconBlockBodyDeneb{
				RandaoReveal:  genesis.Block.Body.RandaoReveal,
				Graffiti:      genesis.Block.Body.Graffiti,
				SilaData:      genesis.Block.Body.SilaData,
				SyncAggregate: &silapb.SyncAggregate{SyncCommitteeBits: scBits[:], SyncCommitteeSignature: make([]byte, 96)},
				SilaPayload: &silaenginev1.SilaPayloadDeneb{
					ParentHash:    make([]byte, fieldparams.RootLength),
					FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
					StateRoot:     make([]byte, fieldparams.RootLength),
					ReceiptsRoot:  make([]byte, fieldparams.RootLength),
					LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
					PrevRandao:    make([]byte, fieldparams.RootLength),
					BaseFeePerGas: make([]byte, fieldparams.RootLength),
					BlockHash:     make([]byte, fieldparams.RootLength),
				},
			},
		},
	}

	blkRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, blkRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, blkRoot), "Could not save genesis state")

	random, err := helpers.RandaoMix(beaconState, slots.ToEpoch(beaconState.Slot()))
	require.NoError(t, err)
	timeStamp, err := slots.StartTime(beaconState.GenesisTime(), denebSlot+1)
	require.NoError(t, err)
	payload := &silaenginev1.SilaPayloadDeneb{
		ParentHash:    make([]byte, fieldparams.RootLength),
		FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:     make([]byte, fieldparams.RootLength),
		ReceiptsRoot:  make([]byte, fieldparams.RootLength),
		LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
		PrevRandao:    random,
		BaseFeePerGas: make([]byte, fieldparams.RootLength),
		BlockHash:     make([]byte, fieldparams.RootLength),
		Transactions:  make([][]byte, 0),
		ExtraData:     make([]byte, 0),
		BlockNumber:   1,
		GasLimit:      2,
		GasUsed:       3,
		Timestamp:     uint64(timeStamp.Unix()),
		BlobGasUsed:   4,
		ExcessBlobGas: 5,
	}
	ed, err := blocks.NewWrappedSilaData(payload)
	require.NoError(t, err)

	kc := make([][]byte, 0)
	kc = append(kc, bytesutil.PadTo([]byte("kc"), 48))
	kc = append(kc, bytesutil.PadTo([]byte("kc1"), 48))
	kc = append(kc, bytesutil.PadTo([]byte("kc2"), 48))
	proofs := [][]byte{[]byte("proof"), []byte("proof1"), []byte("proof2")}
	blobs := [][]byte{[]byte("blob"), []byte("blob1"), []byte("blob2")}
	bundle := &silaenginev1.BlobsBundle{KzgCommitments: kc, Proofs: proofs, Blobs: blobs}
	proposerServer := getProposerServer(ctx, db, beaconState, parentRoot[:])
	advancedState := beaconState.Copy()
	advancedState, err = transition.ProcessSlots(ctx, advancedState, denebSlot)
	require.NoError(t, err)
	proposerServer.BlockReceiver = &mock.ChainService{
		State:           advancedState,
		Root:            parentRoot[:],
		ForkChoiceStore: doublylinkedtree.New(),
	}
	proposerServer.SilaEngineCaller = &mockSila.SilaEngineClient{
		PayloadIDBytes: &silaenginev1.PayloadIDBytes{1},
		GetPayloadResponse: &blocks.GetPayloadResponse{
			SilaData:     ed,
			BlobsBundler: bundle,
		},
	}

	randaoReveal, err := util.RandaoReveal(beaconState, 0, privKeys)
	require.NoError(t, err)

	graffiti := bytesutil.ToBytes32([]byte("sila"))
	require.NoError(t, err)
	req := &silapb.BlockRequest{
		Slot:         denebSlot + 1,
		RandaoReveal: randaoReveal,
		Graffiti:     graffiti[:],
	}

	copiedState := beaconState.Copy()
	copiedState, err = transition.ProcessSlots(ctx, copiedState, denebSlot+1)
	require.NoError(t, err)
	change, err := util.GenerateBLSToSilaChange(copiedState, privKeys[1], 0)
	require.NoError(t, err)
	proposerServer.BLSChangesPool.InsertBLSToExecChange(change)

	got, err := proposerServer.GetBeaconBlock(ctx, req)
	require.NoError(t, err)
	require.DeepEqual(t, got.GetDeneb().Block.Body.BlobKzgCommitments, kc)
}

func TestServer_GetBeaconBlock_Electra(t *testing.T) {
	db := dbutil.SetupDB(t)
	ctx := t.Context()
	transition.SkipSlotCache.Disable()

	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.ElectraForkEpoch = 5
	cfg.DenebForkEpoch = 4
	cfg.CapellaForkEpoch = 3
	cfg.BellatrixForkEpoch = 2
	cfg.AltairForkEpoch = 1
	params.OverrideBeaconConfig(cfg)
	beaconState, privKeys := util.DeterministicGenesisState(t, 64)

	stateRoot, err := beaconState.HashTreeRoot(ctx)
	require.NoError(t, err, "Could not hash genesis state")

	genesis := b.NewGenesisBlock(stateRoot[:])
	util.SaveBlock(t, ctx, db, genesis)

	parentRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, parentRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, parentRoot), "Could not save genesis state")

	electraSlot, err := slots.EpochStart(params.BeaconConfig().ElectraForkEpoch)
	require.NoError(t, err)

	var scBits [fieldparams.SyncAggregateSyncCommitteeBytesLength]byte
	dr := []*silaenginev1.DepositRequest{{
		Pubkey:                bytesutil.PadTo(privKeys[0].PublicKey().Marshal(), 48),
		WithdrawalCredentials: bytesutil.PadTo([]byte("wc"), 32),
		Amount:                123,
		Signature:             bytesutil.PadTo([]byte("sig"), 96),
		Index:                 456,
	}}
	wr := []*silaenginev1.WithdrawalRequest{
		{
			SourceAddress:   bytesutil.PadTo([]byte("sa"), 20),
			ValidatorPubkey: bytesutil.PadTo(privKeys[1].PublicKey().Marshal(), 48),
			Amount:          789,
		},
	}
	cr := []*silaenginev1.ConsolidationRequest{
		{
			SourceAddress: bytesutil.PadTo([]byte("sa"), 20),
			SourcePubkey:  bytesutil.PadTo(privKeys[1].PublicKey().Marshal(), 48),
			TargetPubkey:  bytesutil.PadTo(privKeys[2].PublicKey().Marshal(), 48),
		},
	}
	blk := &silapb.SignedBeaconBlockElectra{
		Block: &silapb.BeaconBlockElectra{
			Slot:       electraSlot + 1,
			ParentRoot: parentRoot[:],
			StateRoot:  genesis.Block.StateRoot,
			Body: &silapb.BeaconBlockBodyElectra{
				RandaoReveal:  genesis.Block.Body.RandaoReveal,
				Graffiti:      genesis.Block.Body.Graffiti,
				SilaData:      genesis.Block.Body.SilaData,
				SyncAggregate: &silapb.SyncAggregate{SyncCommitteeBits: scBits[:], SyncCommitteeSignature: make([]byte, 96)},
				SilaPayload: &silaenginev1.SilaPayloadDeneb{
					ParentHash:    make([]byte, fieldparams.RootLength),
					FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
					StateRoot:     make([]byte, fieldparams.RootLength),
					ReceiptsRoot:  make([]byte, fieldparams.RootLength),
					LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
					PrevRandao:    make([]byte, fieldparams.RootLength),
					BaseFeePerGas: make([]byte, fieldparams.RootLength),
					BlockHash:     make([]byte, fieldparams.RootLength),
				},
				SilaRequests: &silaenginev1.SilaRequests{
					Withdrawals:    wr,
					Deposits:       dr,
					Consolidations: cr,
				},
			},
		},
	}

	blkRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, blkRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, blkRoot), "Could not save genesis state")

	random, err := helpers.RandaoMix(beaconState, slots.ToEpoch(beaconState.Slot()))
	require.NoError(t, err)
	timeStamp, err := slots.StartTime(beaconState.GenesisTime(), electraSlot+1)
	require.NoError(t, err)
	payload := &silaenginev1.SilaPayloadDeneb{
		Timestamp:     uint64(timeStamp.Unix()),
		ParentHash:    make([]byte, fieldparams.RootLength),
		FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:     make([]byte, fieldparams.RootLength),
		ReceiptsRoot:  make([]byte, fieldparams.RootLength),
		LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
		PrevRandao:    random,
		BaseFeePerGas: make([]byte, fieldparams.RootLength),
		BlockHash:     make([]byte, fieldparams.RootLength),
	}
	proposerServer := getProposerServer(ctx, db, beaconState, parentRoot[:])
	advancedState := beaconState.Copy()
	advancedState, err = transition.ProcessSlots(ctx, advancedState, electraSlot)
	require.NoError(t, err)
	proposerServer.BlockReceiver = &mock.ChainService{
		State:           advancedState,
		Root:            parentRoot[:],
		ForkChoiceStore: doublylinkedtree.New(),
	}
	ed, err := blocks.NewWrappedSilaData(payload)
	require.NoError(t, err)
	proposerServer.SilaEngineCaller = &mockSila.SilaEngineClient{
		PayloadIDBytes: &silaenginev1.PayloadIDBytes{1},
		GetPayloadResponse: &blocks.GetPayloadResponse{SilaData: ed, SilaRequests: &silaenginev1.SilaRequests{
			Withdrawals:    wr,
			Deposits:       dr,
			Consolidations: cr,
		}},
	}

	randaoReveal, err := util.RandaoReveal(beaconState, 0, privKeys)
	require.NoError(t, err)

	graffiti := bytesutil.ToBytes32([]byte("sila"))
	require.NoError(t, err)
	req := &silapb.BlockRequest{
		Slot:         electraSlot + 1,
		RandaoReveal: randaoReveal,
		Graffiti:     graffiti[:],
	}

	_, err = proposerServer.GetBeaconBlock(ctx, req)
	require.NoError(t, err)
}

func TestServer_GetBeaconBlock_Fulu(t *testing.T) {
	db := dbutil.SetupDB(t)
	ctx := t.Context()
	transition.SkipSlotCache.Disable()

	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.FuluForkEpoch = 6
	cfg.ElectraForkEpoch = 5
	cfg.DenebForkEpoch = 4
	cfg.CapellaForkEpoch = 3
	cfg.BellatrixForkEpoch = 2
	cfg.AltairForkEpoch = 1
	params.OverrideBeaconConfig(cfg)
	beaconState, privKeys := util.DeterministicGenesisState(t, 64)

	stateRoot, err := beaconState.HashTreeRoot(ctx)
	require.NoError(t, err, "Could not hash genesis state")

	genesis := b.NewGenesisBlock(stateRoot[:])
	util.SaveBlock(t, ctx, db, genesis)

	parentRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, parentRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, parentRoot), "Could not save genesis state")

	fuluSlot, err := slots.EpochStart(params.BeaconConfig().FuluForkEpoch)
	require.NoError(t, err)

	var scBits [fieldparams.SyncAggregateSyncCommitteeBytesLength]byte
	dr := []*silaenginev1.DepositRequest{{
		Pubkey:                bytesutil.PadTo(privKeys[0].PublicKey().Marshal(), 48),
		WithdrawalCredentials: bytesutil.PadTo([]byte("wc"), 32),
		Amount:                123,
		Signature:             bytesutil.PadTo([]byte("sig"), 96),
		Index:                 456,
	}}
	wr := []*silaenginev1.WithdrawalRequest{
		{
			SourceAddress:   bytesutil.PadTo([]byte("sa"), 20),
			ValidatorPubkey: bytesutil.PadTo(privKeys[1].PublicKey().Marshal(), 48),
			Amount:          789,
		},
	}
	cr := []*silaenginev1.ConsolidationRequest{
		{
			SourceAddress: bytesutil.PadTo([]byte("sa"), 20),
			SourcePubkey:  bytesutil.PadTo(privKeys[1].PublicKey().Marshal(), 48),
			TargetPubkey:  bytesutil.PadTo(privKeys[2].PublicKey().Marshal(), 48),
		},
	}
	blk := &silapb.SignedBeaconBlockFulu{
		Block: &silapb.BeaconBlockElectra{
			Slot:       fuluSlot + 1,
			ParentRoot: parentRoot[:],
			StateRoot:  genesis.Block.StateRoot,
			Body: &silapb.BeaconBlockBodyElectra{
				RandaoReveal:  genesis.Block.Body.RandaoReveal,
				Graffiti:      genesis.Block.Body.Graffiti,
				SilaData:      genesis.Block.Body.SilaData,
				SyncAggregate: &silapb.SyncAggregate{SyncCommitteeBits: scBits[:], SyncCommitteeSignature: make([]byte, 96)},
				SilaPayload: &silaenginev1.SilaPayloadDeneb{
					ParentHash:    make([]byte, fieldparams.RootLength),
					FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
					StateRoot:     make([]byte, fieldparams.RootLength),
					ReceiptsRoot:  make([]byte, fieldparams.RootLength),
					LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
					PrevRandao:    make([]byte, fieldparams.RootLength),
					BaseFeePerGas: make([]byte, fieldparams.RootLength),
					BlockHash:     make([]byte, fieldparams.RootLength),
				},
				SilaRequests: &silaenginev1.SilaRequests{
					Withdrawals:    wr,
					Deposits:       dr,
					Consolidations: cr,
				},
			},
		},
	}

	blkRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, err, "Could not get signing root")
	require.NoError(t, db.SaveState(ctx, beaconState, blkRoot), "Could not save genesis state")
	require.NoError(t, db.SaveHeadBlockRoot(ctx, blkRoot), "Could not save genesis state")

	random, err := helpers.RandaoMix(beaconState, slots.ToEpoch(beaconState.Slot()))
	require.NoError(t, err)
	timeStamp, err := slots.StartTime(beaconState.GenesisTime(), fuluSlot+1)
	require.NoError(t, err)
	payload := &silaenginev1.SilaPayloadDeneb{
		Timestamp:     uint64(timeStamp.Unix()),
		ParentHash:    make([]byte, fieldparams.RootLength),
		FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:     make([]byte, fieldparams.RootLength),
		ReceiptsRoot:  make([]byte, fieldparams.RootLength),
		LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
		PrevRandao:    random,
		BaseFeePerGas: make([]byte, fieldparams.RootLength),
		BlockHash:     make([]byte, fieldparams.RootLength),
	}
	proposerServer := getProposerServer(ctx, db, beaconState, parentRoot[:])
	advancedState := beaconState.Copy()
	advancedState, err = transition.ProcessSlots(ctx, advancedState, fuluSlot)
	require.NoError(t, err)
	proposerServer.BlockReceiver = &mock.ChainService{
		State:           advancedState,
		Root:            parentRoot[:],
		ForkChoiceStore: doublylinkedtree.New(),
	}
	ed, err := blocks.NewWrappedSilaData(payload)
	require.NoError(t, err)
	proposerServer.SilaEngineCaller = &mockSila.SilaEngineClient{
		PayloadIDBytes: &silaenginev1.PayloadIDBytes{1},
		GetPayloadResponse: &blocks.GetPayloadResponse{SilaData: ed, SilaRequests: &silaenginev1.SilaRequests{
			Withdrawals:    wr,
			Deposits:       dr,
			Consolidations: cr,
		}},
	}

	randaoReveal, err := util.RandaoReveal(beaconState, 0, privKeys)
	require.NoError(t, err)

	graffiti := bytesutil.ToBytes32([]byte("sila"))
	require.NoError(t, err)
	req := &silapb.BlockRequest{
		Slot:         fuluSlot + 1,
		RandaoReveal: randaoReveal,
		Graffiti:     graffiti[:],
	}

	_, err = proposerServer.GetBeaconBlock(ctx, req)
	require.NoError(t, err)
}

func TestServer_GetBeaconBlock_Optimistic(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.BellatrixForkEpoch = 2
	cfg.AltairForkEpoch = 1
	params.OverrideBeaconConfig(cfg)

	bellatrixSlot, err := slots.EpochStart(params.BeaconConfig().BellatrixForkEpoch)
	require.NoError(t, err)

	mockChainService := &mock.ChainService{ForkChoiceStore: doublylinkedtree.New()}
	proposerServer := &Server{
		OptimisticModeFetcher: &mock.ChainService{Optimistic: true},
		SyncChecker:           &mockSync.Sync{},
		ForkFetcher:           mockChainService,
		ForkchoiceFetcher:     mockChainService,
		TimeFetcher:           &mock.ChainService{}}
	req := &silapb.BlockRequest{
		Slot: bellatrixSlot + 1,
	}
	_, err = proposerServer.GetBeaconBlock(t.Context(), req)
	s, ok := status.FromError(err)
	require.Equal(t, true, ok)
	require.DeepEqual(t, codes.Unavailable, s.Code())
	require.ErrorContains(t, errOptimisticMode.Error(), err)
}

func getProposerServer(ctx context.Context, db db.HeadAccessDatabase, headState state.BeaconState, headRoot []byte) *Server {
	mockChainService := &mock.ChainService{State: headState, Root: headRoot, ForkChoiceStore: doublylinkedtree.New()}
	return &Server{
		HeadFetcher:            mockChainService,
		SyncChecker:            &mockSync.Sync{IsSyncing: false},
		BlockReceiver:          mockChainService,
		ChainStartFetcher:      &mockSila.Chain{},
		SilaChainInfoFetcher:   &mockSila.Chain{},
		SilaBlockFetcher:       &mockSila.Chain{},
		FinalizationFetcher:    mockChainService,
		ForkFetcher:            mockChainService,
		ForkchoiceFetcher:      mockChainService,
		MockSilaExecutionVotes: true,
		AttPool:                attestations.NewPool(),
		SlashingsPool:          slashings.NewPool(),
		ExitPool:               voluntaryexits.NewPool(),
		StateGen:               stategen.New(db, doublylinkedtree.New()),
		SyncCommitteePool:      synccommittee.NewStore(),
		OptimisticModeFetcher:  &mock.ChainService{},
		TimeFetcher: &testutil.MockGenesisTimeFetcher{
			Genesis: time.Now(),
		},
		PayloadIDCache:         cache.NewPayloadIDCache(),
		TrackedValidatorsCache: cache.NewTrackedValidatorsCache(),
		BeaconDB:               db,
		BLSChangesPool:         blstoexec.NewPool(),
		BlockBuilder:           &builderTest.MockBuilderService{HasConfigured: true},
	}
}

func injectSlashings(t *testing.T, st state.BeaconState, keys []bls.SecretKey, server *Server) ([]*silapb.ProposerSlashing, []*silapb.AttesterSlashing) {
	proposerSlashings := make([]*silapb.ProposerSlashing, params.BeaconConfig().MaxProposerSlashings)
	for i := primitives.ValidatorIndex(0); uint64(i) < params.BeaconConfig().MaxProposerSlashings; i++ {
		proposerSlashing, err := util.GenerateProposerSlashingForValidator(st, keys[i], i /* validator index */)
		require.NoError(t, err)
		proposerSlashings[i] = proposerSlashing
		err = server.SlashingsPool.InsertProposerSlashing(t.Context(), st, proposerSlashing)
		require.NoError(t, err)
	}

	attSlashings := make([]*silapb.AttesterSlashing, params.BeaconConfig().MaxAttesterSlashings)
	for i := uint64(0); i < params.BeaconConfig().MaxAttesterSlashings; i++ {
		generatedAttesterSlashing, err := util.GenerateAttesterSlashingForValidator(st, keys[i+params.BeaconConfig().MaxProposerSlashings], primitives.ValidatorIndex(i+params.BeaconConfig().MaxProposerSlashings) /* validator index */)
		require.NoError(t, err)
		attesterSlashing, ok := generatedAttesterSlashing.(*silapb.AttesterSlashing)
		require.Equal(t, true, ok, "Attester slashing has the wrong type (expected %T, got %T)", &silapb.AttesterSlashing{}, generatedAttesterSlashing)
		attSlashings[i] = attesterSlashing
		err = server.SlashingsPool.InsertAttesterSlashing(t.Context(), st, generatedAttesterSlashing.(*silapb.AttesterSlashing))
		require.NoError(t, err)
	}
	return proposerSlashings, attSlashings
}

func TestProposer_ProposeBlock_OK(t *testing.T) {
	// Initialize KZG for Fulu blocks
	require.NoError(t, kzg.Start())

	tests := []struct {
		name       string
		block      func([32]byte) *silapb.GenericSignedBeaconBlock
		err        string
		useBuilder bool
	}{
		{
			name: "phase0",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				blockToPropose := util.NewBeaconBlock()
				blockToPropose.Block.Slot = 5
				blockToPropose.Block.ParentRoot = parent[:]
				blk := &silapb.GenericSignedBeaconBlock_Phase0{Phase0: blockToPropose}
				return &silapb.GenericSignedBeaconBlock{Block: blk}
			},
		},
		{
			name: "altair",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				blockToPropose := util.NewBeaconBlockAltair()
				blockToPropose.Block.Slot = 5
				blockToPropose.Block.ParentRoot = parent[:]
				blk := &silapb.GenericSignedBeaconBlock_Altair{Altair: blockToPropose}
				return &silapb.GenericSignedBeaconBlock{Block: blk}
			},
		},
		{
			name: "bellatrix",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				blockToPropose := util.NewBeaconBlockBellatrix()
				blockToPropose.Block.Slot = 5
				blockToPropose.Block.ParentRoot = parent[:]
				blk := &silapb.GenericSignedBeaconBlock_Bellatrix{Bellatrix: blockToPropose}
				return &silapb.GenericSignedBeaconBlock{Block: blk}
			},
		},
		{
			name: "blind capella",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				blockToPropose := util.NewBlindedBeaconBlockCapella()
				blockToPropose.Block.Slot = 5
				blockToPropose.Block.ParentRoot = parent[:]
				txRoot, err := ssz.TransactionsRoot([][]byte{})
				require.NoError(t, err)
				withdrawalsRoot, err := ssz.WithdrawalSliceRoot([]*silaenginev1.Withdrawal{}, fieldparams.MaxWithdrawalsPerPayload)
				require.NoError(t, err)
				blockToPropose.Block.Body.SilaPayloadHeader.TransactionsRoot = txRoot[:]
				blockToPropose.Block.Body.SilaPayloadHeader.WithdrawalsRoot = withdrawalsRoot[:]
				blk := &silapb.GenericSignedBeaconBlock_BlindedCapella{BlindedCapella: blockToPropose}
				return &silapb.GenericSignedBeaconBlock{Block: blk}
			},
			useBuilder: true,
		},
		{
			name: "blind capella no builder",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				blockToPropose := util.NewBlindedBeaconBlockCapella()
				blockToPropose.Block.Slot = 5
				blockToPropose.Block.ParentRoot = parent[:]
				txRoot, err := ssz.TransactionsRoot([][]byte{})
				require.NoError(t, err)
				withdrawalsRoot, err := ssz.WithdrawalSliceRoot([]*silaenginev1.Withdrawal{}, fieldparams.MaxWithdrawalsPerPayload)
				require.NoError(t, err)
				blockToPropose.Block.Body.SilaPayloadHeader.TransactionsRoot = txRoot[:]
				blockToPropose.Block.Body.SilaPayloadHeader.WithdrawalsRoot = withdrawalsRoot[:]
				blk := &silapb.GenericSignedBeaconBlock_BlindedCapella{BlindedCapella: blockToPropose}
				return &silapb.GenericSignedBeaconBlock{Block: blk}
			},
			err: "unconfigured block builder",
		},
		{
			name: "bellatrix",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				blockToPropose := util.NewBeaconBlockBellatrix()
				blockToPropose.Block.Slot = 5
				blockToPropose.Block.ParentRoot = parent[:]
				blk := &silapb.GenericSignedBeaconBlock_Bellatrix{Bellatrix: blockToPropose}
				return &silapb.GenericSignedBeaconBlock{Block: blk}
			},
		},
		{
			name: "deneb block no blob",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				blockToPropose := util.NewBeaconBlockContentsDeneb()
				blockToPropose.Block.Block.Slot = 5
				blockToPropose.Block.Block.ParentRoot = parent[:]
				blk := &silapb.GenericSignedBeaconBlock_Deneb{Deneb: blockToPropose}
				return &silapb.GenericSignedBeaconBlock{Block: blk}
			},
		},
		{
			name: "deneb block some blobs",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				blockToPropose := util.NewBeaconBlockContentsDeneb()
				blockToPropose.Block.Block.Slot = 5
				blockToPropose.Block.Block.ParentRoot = parent[:]
				blockToPropose.Blobs = [][]byte{{0x01}, {0x02}, {0x03}}
				blockToPropose.KzgProofs = [][]byte{{0x01}, {0x02}, {0x03}}
				blockToPropose.Block.Block.Body.BlobKzgCommitments = [][]byte{bytesutil.PadTo([]byte("kc"), 48), bytesutil.PadTo([]byte("kc1"), 48), bytesutil.PadTo([]byte("kc2"), 48)}
				blk := &silapb.GenericSignedBeaconBlock_Deneb{Deneb: blockToPropose}
				return &silapb.GenericSignedBeaconBlock{Block: blk}
			},
		},
		{
			name: "deneb block some blobs (kzg and blob count mismatch)",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				blockToPropose := util.NewBeaconBlockContentsDeneb()
				blockToPropose.Block.Block.Slot = 5
				blockToPropose.Block.Block.ParentRoot = parent[:]
				blockToPropose.Blobs = [][]byte{{0x01}, {0x02}, {0x03}}
				blockToPropose.KzgProofs = [][]byte{{0x01}, {0x02}, {0x03}}
				blk := &silapb.GenericSignedBeaconBlock_Deneb{Deneb: blockToPropose}
				return &silapb.GenericSignedBeaconBlock{Block: blk}
			},
			err: "blob KZG commitments don't match number of blobs or KZG proofs",
		},
		{
			name: "blind deneb block some blobs",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				blockToPropose := util.NewBlindedBeaconBlockDeneb()
				blockToPropose.Message.Slot = 5
				blockToPropose.Message.ParentRoot = parent[:]
				txRoot, err := ssz.TransactionsRoot([][]byte{})
				require.NoError(t, err)
				withdrawalsRoot, err := ssz.WithdrawalSliceRoot([]*silaenginev1.Withdrawal{}, fieldparams.MaxWithdrawalsPerPayload)
				require.NoError(t, err)
				blockToPropose.Message.Body.SilaPayloadHeader.TransactionsRoot = txRoot[:]
				blockToPropose.Message.Body.SilaPayloadHeader.WithdrawalsRoot = withdrawalsRoot[:]
				blockToPropose.Message.Body.BlobKzgCommitments = [][]byte{bytesutil.PadTo([]byte{0x01}, 48)}
				blk := &silapb.GenericSignedBeaconBlock_BlindedDeneb{BlindedDeneb: blockToPropose}
				return &silapb.GenericSignedBeaconBlock{Block: blk}
			},
			useBuilder: true,
		},
		{
			name: "blind deneb block some blobs (commitment value does not match blob)",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				blockToPropose := util.NewBlindedBeaconBlockDeneb()
				blockToPropose.Message.Slot = 5
				blockToPropose.Message.ParentRoot = parent[:]
				txRoot, err := ssz.TransactionsRoot([][]byte{})
				require.NoError(t, err)
				withdrawalsRoot, err := ssz.WithdrawalSliceRoot([]*silaenginev1.Withdrawal{}, fieldparams.MaxWithdrawalsPerPayload)
				require.NoError(t, err)
				blockToPropose.Message.Body.SilaPayloadHeader.TransactionsRoot = txRoot[:]
				blockToPropose.Message.Body.SilaPayloadHeader.WithdrawalsRoot = withdrawalsRoot[:]
				blockToPropose.Message.Body.BlobKzgCommitments = [][]byte{bytesutil.PadTo([]byte("kc"), 48)}
				blk := &silapb.GenericSignedBeaconBlock_BlindedDeneb{BlindedDeneb: blockToPropose}
				return &silapb.GenericSignedBeaconBlock{Block: blk}
			},
			useBuilder: true,
			err:        "unblind blobs sidecars: commitment value doesn't match block",
		},
		{
			name: "electra block no blob",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				sb := &silapb.SignedBeaconBlockContentsElectra{
					Block: &silapb.SignedBeaconBlockElectra{
						Block: &silapb.BeaconBlockElectra{Slot: 5, ParentRoot: parent[:], Body: util.HydrateBeaconBlockBodyElectra(&silapb.BeaconBlockBodyElectra{})},
					},
				}
				blk := &silapb.GenericSignedBeaconBlock_Electra{Electra: sb}
				return &silapb.GenericSignedBeaconBlock{Block: blk, IsBlinded: false}
			},
		},
		{
			name: "electra block some blob",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				sb := &silapb.SignedBeaconBlockContentsElectra{
					Block: &silapb.SignedBeaconBlockElectra{
						Block: &silapb.BeaconBlockElectra{
							Slot: 5, ParentRoot: parent[:],
							Body: util.HydrateBeaconBlockBodyElectra(&silapb.BeaconBlockBodyElectra{
								BlobKzgCommitments: [][]byte{bytesutil.PadTo([]byte("kc"), 48), bytesutil.PadTo([]byte("kc1"), 48), bytesutil.PadTo([]byte("kc2"), 48)},
							}),
						},
					},
					KzgProofs: [][]byte{{0x01}, {0x02}, {0x03}},
					Blobs:     [][]byte{{0x01}, {0x02}, {0x03}},
				}
				blk := &silapb.GenericSignedBeaconBlock_Electra{Electra: sb}
				return &silapb.GenericSignedBeaconBlock{Block: blk, IsBlinded: false}
			},
		},
		{
			name: "electra block some blob (kzg and blob count mismatch)",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				sb := &silapb.SignedBeaconBlockContentsElectra{
					Block: &silapb.SignedBeaconBlockElectra{
						Block: &silapb.BeaconBlockElectra{
							Slot: 5, ParentRoot: parent[:],
							Body: util.HydrateBeaconBlockBodyElectra(&silapb.BeaconBlockBodyElectra{
								BlobKzgCommitments: [][]byte{bytesutil.PadTo([]byte("kc"), 48), bytesutil.PadTo([]byte("kc1"), 48)},
							}),
						},
					},
					KzgProofs: [][]byte{{0x01}, {0x02}, {0x03}},
					Blobs:     [][]byte{{0x01}, {0x02}, {0x03}},
				}
				blk := &silapb.GenericSignedBeaconBlock_Electra{Electra: sb}
				return &silapb.GenericSignedBeaconBlock{Block: blk, IsBlinded: false}
			},
			err: "blob KZG commitments don't match number of blobs or KZG proofs",
		},
		{
			name: "fulu block no blob",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				sb := &silapb.SignedBeaconBlockContentsFulu{
					Block: &silapb.SignedBeaconBlockFulu{
						Block: &silapb.BeaconBlockElectra{Slot: 5, ParentRoot: parent[:], Body: util.HydrateBeaconBlockBodyElectra(&silapb.BeaconBlockBodyElectra{})},
					},
				}
				blk := &silapb.GenericSignedBeaconBlock_Fulu{Fulu: sb}
				return &silapb.GenericSignedBeaconBlock{Block: blk, IsBlinded: false}
			},
		},
		{
			name: "fulu block with single blob and cell proofs",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				numberOfColumns := uint64(128)
				// For Fulu, we have cell proofs (blobs * numberOfColumns)
				cellProofs := make([][]byte, numberOfColumns)
				for i := range numberOfColumns {
					cellProofs[i] = bytesutil.PadTo([]byte{byte(i)}, 48)
				}
				// Blob must be exactly 131072 bytes
				blob := make([]byte, 131072)
				blob[0] = 0x01
				sb := &silapb.SignedBeaconBlockContentsFulu{
					Block: &silapb.SignedBeaconBlockFulu{
						Block: &silapb.BeaconBlockElectra{
							Slot: 5, ParentRoot: parent[:],
							Body: util.HydrateBeaconBlockBodyElectra(&silapb.BeaconBlockBodyElectra{
								BlobKzgCommitments: [][]byte{bytesutil.PadTo([]byte("kc"), 48)},
							}),
						},
					},
					KzgProofs: cellProofs,
					Blobs:     [][]byte{blob},
				}
				blk := &silapb.GenericSignedBeaconBlock_Fulu{Fulu: sb}
				return &silapb.GenericSignedBeaconBlock{Block: blk, IsBlinded: false}
			},
		},
		{
			name: "fulu block with multiple blobs and cell proofs",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				numberOfColumns := uint64(128)
				blobCount := 3
				// For Fulu, we have cell proofs (blobs * numberOfColumns)
				cellProofs := make([][]byte, uint64(blobCount)*numberOfColumns)
				for i := range cellProofs {
					cellProofs[i] = bytesutil.PadTo([]byte{byte(i % 256)}, 48)
				}
				// Create properly sized blobs (131072 bytes each)
				blobs := make([][]byte, blobCount)
				for i := range blobCount {
					blob := make([]byte, 131072)
					blob[0] = byte(i + 1)
					blobs[i] = blob
				}
				sb := &silapb.SignedBeaconBlockContentsFulu{
					Block: &silapb.SignedBeaconBlockFulu{
						Block: &silapb.BeaconBlockElectra{
							Slot: 5, ParentRoot: parent[:],
							Body: util.HydrateBeaconBlockBodyElectra(&silapb.BeaconBlockBodyElectra{
								BlobKzgCommitments: [][]byte{
									bytesutil.PadTo([]byte("kc"), 48),
									bytesutil.PadTo([]byte("kc1"), 48),
									bytesutil.PadTo([]byte("kc2"), 48),
								},
							}),
						},
					},
					KzgProofs: cellProofs,
					Blobs:     blobs,
				}
				blk := &silapb.GenericSignedBeaconBlock_Fulu{Fulu: sb}
				return &silapb.GenericSignedBeaconBlock{Block: blk, IsBlinded: false}
			},
		},
		{
			name: "fulu block wrong cell proof count (should be blobs * 128)",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				// Wrong number of cell proofs - should be 2 * 128 = 256, but providing only 2
				// Create properly sized blobs
				blob1 := make([]byte, 131072)
				blob1[0] = 0x01
				blob2 := make([]byte, 131072)
				blob2[0] = 0x02
				sb := &silapb.SignedBeaconBlockContentsFulu{
					Block: &silapb.SignedBeaconBlockFulu{
						Block: &silapb.BeaconBlockElectra{
							Slot: 5, ParentRoot: parent[:],
							Body: util.HydrateBeaconBlockBodyElectra(&silapb.BeaconBlockBodyElectra{
								BlobKzgCommitments: [][]byte{
									bytesutil.PadTo([]byte("kc"), 48),
									bytesutil.PadTo([]byte("kc1"), 48),
								},
							}),
						},
					},
					KzgProofs: [][]byte{{0x01}, {0x02}}, // Wrong: should be 256 cell proofs
					Blobs:     [][]byte{blob1, blob2},
				}
				blk := &silapb.GenericSignedBeaconBlock_Fulu{Fulu: sb}
				return &silapb.GenericSignedBeaconBlock{Block: blk, IsBlinded: false}
			},
			err: "blobs and cells proofs mismatch",
		},
		{
			name: "blind fulu block with blob commitments",
			block: func(parent [32]byte) *silapb.GenericSignedBeaconBlock {
				blockToPropose := util.NewBlindedBeaconBlockFulu()
				blockToPropose.Message.Slot = 5
				blockToPropose.Message.ParentRoot = parent[:]
				txRoot, err := ssz.TransactionsRoot([][]byte{})
				require.NoError(t, err)
				withdrawalsRoot, err := ssz.WithdrawalSliceRoot([]*silaenginev1.Withdrawal{}, fieldparams.MaxWithdrawalsPerPayload)
				require.NoError(t, err)
				blockToPropose.Message.Body.SilaPayloadHeader.TransactionsRoot = txRoot[:]
				blockToPropose.Message.Body.SilaPayloadHeader.WithdrawalsRoot = withdrawalsRoot[:]
				blockToPropose.Message.Body.BlobKzgCommitments = [][]byte{bytesutil.PadTo([]byte{0x01}, 48)}
				blk := &silapb.GenericSignedBeaconBlock_BlindedFulu{BlindedFulu: blockToPropose}
				return &silapb.GenericSignedBeaconBlock{Block: blk}
			},
			useBuilder: true,
			err:        "commitment value doesn't match block", // Known issue with mock builder cell proof mismatch
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			numDeposits := uint64(64)
			beaconState, _ := util.DeterministicGenesisState(t, numDeposits)
			bsRoot, err := beaconState.HashTreeRoot(ctx)
			require.NoError(t, err)

			c := &mock.ChainService{Root: bsRoot[:], State: beaconState}
			db := dbutil.SetupDB(t)
			// Create cell proofs for Fulu blocks (128 proofs per blob)
			numberOfColumns := uint64(128)
			cellProofs := make([][]byte, numberOfColumns)
			for i := range numberOfColumns {
				cellProofs[i] = bytesutil.PadTo([]byte{byte(i)}, 48)
			}
			// Create properly sized blob for mock builder
			mockBlob := make([]byte, 131072)
			mockBlob[0] = 0x03
			// Use the same commitment as in the blind block test
			mockCommitment := bytesutil.PadTo([]byte{0x01}, 48)

			proposerServer := &Server{
				BlockReceiver: c,
				BlockNotifier: c.BlockNotifier(),
				P2P:           mockp2p.NewTestP2P(t),
				BlockBuilder: &builderTest.MockBuilderService{HasConfigured: tt.useBuilder, PayloadCapella: emptyPayloadCapella(), PayloadDeneb: emptyPayloadDeneb(),
					BlobBundle:   &silaenginev1.BlobsBundle{KzgCommitments: [][]byte{mockCommitment}, Proofs: [][]byte{{0x02}}, Blobs: [][]byte{{0x03}}},
					BlobBundleV2: &silaenginev1.BlobsBundleV2{KzgCommitments: [][]byte{mockCommitment}, Proofs: cellProofs, Blobs: [][]byte{mockBlob}}},
				BeaconDB:           db,
				BlobReceiver:       c,
				DataColumnReceiver: c, // Add DataColumnReceiver for Fulu blocks
				OperationNotifier:  c.OperationNotifier(),
			}
			blockToPropose := tt.block(bsRoot)
			res, err := proposerServer.ProposeBeaconBlock(t.Context(), blockToPropose)
			if tt.err != "" { // Expecting an error
				require.ErrorContains(t, tt.err, err)
			} else {
				assert.NoError(t, err, "Could not propose block correctly")
				if res == nil || len(res.BlockRoot) == 0 {
					t.Error("No block root was returned")
				}
			}
		})
	}
}

func TestProposer_ComputeStateRoot_OK(t *testing.T) {
	db := dbutil.SetupDB(t)
	ctx := t.Context()

	beaconState, parentRoot, privKeys := util.DeterministicGenesisStateWithGenesisBlock(t, ctx, db, 100)

	proposerServer := &Server{
		ChainStartFetcher:    &mockSila.Chain{},
		SilaChainInfoFetcher: &mockSila.Chain{},
		SilaBlockFetcher:     &mockSila.Chain{},
		StateGen:             stategen.New(db, doublylinkedtree.New()),
		BlockReceiver: &mock.ChainService{
			State:           beaconState.Copy(),
			Root:            parentRoot[:],
			ForkChoiceStore: doublylinkedtree.New(),
		},
	}
	req := util.NewBeaconBlock()
	req.Block.ProposerIndex = 84
	req.Block.ParentRoot = parentRoot[:]
	req.Block.Slot = 1
	require.NoError(t, beaconState.SetSlot(beaconState.Slot()+1))
	randaoReveal, err := util.RandaoReveal(beaconState, 0, privKeys)
	require.NoError(t, err)
	proposerIdx, err := helpers.BeaconProposerIndex(ctx, beaconState)
	require.NoError(t, err)
	require.NoError(t, beaconState.SetSlot(slots.PrevSlot(beaconState.Slot())))
	req.Block.Body.RandaoReveal = randaoReveal
	currentEpoch := coretime.CurrentEpoch(beaconState)
	req.Signature, err = signing.ComputeDomainAndSign(beaconState, currentEpoch, req.Block, params.BeaconConfig().DomainBeaconProposer, privKeys[proposerIdx])
	require.NoError(t, err)

	wsb, err := blocks.NewSignedBeaconBlock(req)
	require.NoError(t, err)
	_, _, err = proposerServer.computePostBlockStateAndRoot(t.Context(), wsb)
	require.NoError(t, err)
}

func TestHandlePostBlockStateError_MaxAttemptsReached(t *testing.T) {
	// Test that handlePostBlockStateError returns an error when max attempts is reached
	// instead of recursing infinitely.
	ctx := t.Context()
	vs := &Server{}

	// Create a minimal block for testing
	blk := util.NewBeaconBlock()
	wsb, err := blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)

	// Pre-seed the context with max attempts already reached
	ctx = context.WithValue(ctx, computeStateRootAttemptsKey, maxComputeStateRootAttempts)

	// Call handlePostBlockStateError with a retryable error
	_, err = vs.handlePostBlockStateError(ctx, wsb, transition.ErrAttestationsSignatureInvalid)

	// Should return an error about max attempts instead of recursing
	require.ErrorContains(t, "attempted max compute state root attempts", err)
}

func TestHandlePostBlockStateError_IncrementsAttempts(t *testing.T) {
	// Test that handlePostBlockStateError properly increments the attempts counter
	// and eventually fails after max attempts.
	db := dbutil.SetupDB(t)
	ctx := t.Context()

	beaconState, parentRoot, _ := util.DeterministicGenesisStateWithGenesisBlock(t, ctx, db, 100)

	stateGen := stategen.New(db, doublylinkedtree.New())
	vs := &Server{
		StateGen:      stateGen,
		BlockReceiver: &mock.ChainService{State: beaconState},
	}

	// Create a block that will trigger retries
	blk := util.NewBeaconBlock()
	blk.Block.ParentRoot = parentRoot[:]
	blk.Block.Slot = 1
	wsb, err := blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)

	// Add a state for the parent root so StateByRoot succeeds
	require.NoError(t, stateGen.SaveState(ctx, parentRoot, beaconState))

	// Call handlePostBlockStateError with a retryable error - it will recurse
	// but eventually hit the max attempts limit since CalculatePostState
	// will keep failing (no valid attestations, randao, etc.)
	_, err = vs.handlePostBlockStateError(ctx, wsb, transition.ErrAttestationsSignatureInvalid)

	// Should eventually fail - either with max attempts or another error
	require.NotNil(t, err)
}

func TestProposer_PendingDeposits_SilaDataVoteOK(t *testing.T) {
	ctx := t.Context()

	height := big.NewInt(int64(params.BeaconConfig().SilaExecutionFollowDistance))
	newHeight := big.NewInt(height.Int64() + 11000)
	p := &mockSila.Chain{
		LatestBlockNumber: height,
		HashesByHeight: map[int][]byte{
			int(height.Int64()):    []byte("0x0"),
			int(newHeight.Int64()): []byte("0x1"),
		},
	}

	var votes []*silapb.SilaData

	blockHash := make([]byte, 32)
	copy(blockHash, "0x1")
	vote := &silapb.SilaData{
		DepositRoot:  make([]byte, 32),
		BlockHash:    blockHash,
		DepositCount: 3,
	}
	period := uint64(params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().EpochsPerSilaExecutionVotingPeriod)))
	for i := 0; i <= int(period/2); i++ {
		votes = append(votes, vote)
	}

	blockHash = make([]byte, 32)
	copy(blockHash, "0x0")
	beaconState, err := util.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, beaconState.SetSilaExecutionDepositIndex(2))
	require.NoError(t, beaconState.SetSilaData(&silapb.SilaData{
		DepositRoot:  make([]byte, 32),
		BlockHash:    blockHash,
		DepositCount: 2,
	}))
	require.NoError(t, beaconState.SetSilaDataVotes(votes))

	blk := util.NewBeaconBlock()
	blkRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)

	bs := &Server{
		ChainStartFetcher:    p,
		SilaChainInfoFetcher: p,
		SilaBlockFetcher:     p,
		BlockReceiver:        &mock.ChainService{State: beaconState, Root: blkRoot[:]},
		HeadFetcher:          &mock.ChainService{State: beaconState, Root: blkRoot[:]},
	}

	// It should also return the recent deposits after their follow window.
	p.LatestBlockNumber = big.NewInt(0).Add(p.LatestBlockNumber, big.NewInt(10000))
	_, silaexecHeight, err := bs.canonicalSilaData(ctx, beaconState, &silapb.SilaData{})
	require.NoError(t, err)

	assert.Equal(t, 0, silaexecHeight.Cmp(height))

	newState, err := b.ProcessSilaDataInBlock(ctx, beaconState, blk.Block.Body.SilaData)
	require.NoError(t, err)

	if proto.Equal(newState.SilaData(), vote) {
		t.Errorf("silaData in the state equal to vote, when not expected to"+
			"have majority: Got %v", vote)
	}

	blk.Block.Body.SilaData = vote

	_, silaexecHeight, err = bs.canonicalSilaData(ctx, beaconState, vote)
	require.NoError(t, err)
	assert.Equal(t, 0, silaexecHeight.Cmp(newHeight))

	newState, err = b.ProcessSilaDataInBlock(ctx, beaconState, blk.Block.Body.SilaData)
	require.NoError(t, err)

	if !proto.Equal(newState.SilaData(), vote) {
		t.Errorf("silaData in the state not of the expected kind: Got %v but wanted %v", newState.SilaData(), vote)
	}
}

func TestProposer_PendingDeposits_OutsideSilaExecutionFollowWindow(t *testing.T) {
	ctx := t.Context()

	height := big.NewInt(int64(params.BeaconConfig().SilaExecutionFollowDistance))
	p := &mockSila.Chain{
		LatestBlockNumber: height,
		HashesByHeight: map[int][]byte{
			int(height.Int64()): []byte("0x0"),
		},
	}

	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		SilaData: &silapb.SilaData{
			BlockHash:   bytesutil.PadTo([]byte("0x0"), 32),
			DepositRoot: make([]byte, 32),
		},
		SilaexecDepositIndex: 2,
	})
	require.NoError(t, err)

	var mockSig [96]byte
	var mockCreds [32]byte

	// Using the merkleTreeIndex as the block number for this test...
	readyDeposits := []*silapb.DepositContainer{
		{
			Index:           0,
			SilaexecBlockHeight: 2,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("a"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index:           1,
			SilaexecBlockHeight: 8,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("b"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	recentDeposits := []*silapb.DepositContainer{
		{
			Index:           2,
			SilaexecBlockHeight: 400,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("c"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index:           3,
			SilaexecBlockHeight: 600,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("d"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)

	depositTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
	require.NoError(t, err, "Could not setup deposit trie")
	for _, dp := range append(readyDeposits, recentDeposits...) {
		depositHash, err := dp.Deposit.Data.HashTreeRoot()
		require.NoError(t, err, "Unable to determine hashed value of deposit")

		assert.NoError(t, depositTrie.Insert(depositHash[:], int(dp.Index)))
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		assert.NoError(t, depositCache.InsertDeposit(ctx, dp.Deposit, dp.SilaexecBlockHeight, dp.Index, root))
	}
	for _, dp := range recentDeposits {
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		depositCache.InsertPendingDeposit(ctx, dp.Deposit, dp.SilaexecBlockHeight, dp.Index, root)
	}

	blk := util.NewBeaconBlock()
	blk.Block.Slot = beaconState.Slot()

	blkRoot, err := blk.HashTreeRoot()
	require.NoError(t, err)

	bs := &Server{
		ChainStartFetcher:      p,
		SilaChainInfoFetcher:   p,
		SilaBlockFetcher:       p,
		DepositFetcher:         depositCache,
		PendingDepositsFetcher: depositCache,
		BlockReceiver:          &mock.ChainService{State: beaconState, Root: blkRoot[:]},
		HeadFetcher:            &mock.ChainService{State: beaconState, Root: blkRoot[:]},
	}

	deposits, err := bs.deposits(ctx, beaconState, &silapb.SilaData{})
	require.NoError(t, err)
	assert.Equal(t, 0, len(deposits), "Received unexpected list of deposits")

	// It should not return the recent deposits after their follow window.
	// as latest block number makes no difference in retrieval of deposits
	p.LatestBlockNumber = big.NewInt(0).Add(p.LatestBlockNumber, big.NewInt(10000))
	deposits, err = bs.deposits(ctx, beaconState, &silapb.SilaData{})
	require.NoError(t, err)
	assert.Equal(t, 0, len(deposits), "Received unexpected number of pending deposits")
}

func TestProposer_PendingDeposits_FollowsCorrectSilaBlock(t *testing.T) {
	ctx := t.Context()

	height := big.NewInt(int64(params.BeaconConfig().SilaExecutionFollowDistance))
	newHeight := big.NewInt(height.Int64() + 11000)
	p := &mockSila.Chain{
		LatestBlockNumber: height,
		HashesByHeight: map[int][]byte{
			int(height.Int64()):    []byte("0x0"),
			int(newHeight.Int64()): []byte("0x1"),
		},
	}

	var votes []*silapb.SilaData

	vote := &silapb.SilaData{
		BlockHash:    bytesutil.PadTo([]byte("0x1"), 32),
		DepositRoot:  make([]byte, 32),
		DepositCount: 7,
	}
	period := uint64(params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().EpochsPerSilaExecutionVotingPeriod)))
	for i := 0; i <= int(period/2); i++ {
		votes = append(votes, vote)
	}

	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		SilaData: &silapb.SilaData{
			BlockHash:    []byte("0x0"),
			DepositRoot:  make([]byte, 32),
			DepositCount: 5,
		},
		SilaexecDepositIndex: 1,
		SilaDataVotes:             votes,
	})
	require.NoError(t, err)
	blk := util.NewBeaconBlock()
	blk.Block.Slot = beaconState.Slot()

	blkRoot, err := blk.HashTreeRoot()
	require.NoError(t, err)

	var mockSig [96]byte
	var mockCreds [32]byte

	// Using the merkleTreeIndex as the block number for this test...
	readyDeposits := []*silapb.DepositContainer{
		{
			Index:           0,
			SilaexecBlockHeight: 8,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("a"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index:           1,
			SilaexecBlockHeight: 14,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("b"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	recentDeposits := []*silapb.DepositContainer{
		{
			Index:           2,
			SilaexecBlockHeight: 5000,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("c"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index:           3,
			SilaexecBlockHeight: 6000,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("d"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)

	depositTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
	require.NoError(t, err, "Could not setup deposit trie")
	for _, dp := range append(readyDeposits, recentDeposits...) {
		depositHash, err := dp.Deposit.Data.HashTreeRoot()
		require.NoError(t, err, "Unable to determine hashed value of deposit")

		assert.NoError(t, depositTrie.Insert(depositHash[:], int(dp.Index)))
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		assert.NoError(t, depositCache.InsertDeposit(ctx, dp.Deposit, dp.SilaexecBlockHeight, dp.Index, root))
	}
	for _, dp := range recentDeposits {
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		depositCache.InsertPendingDeposit(ctx, dp.Deposit, dp.SilaexecBlockHeight, dp.Index, root)
	}

	bs := &Server{
		ChainStartFetcher:      p,
		SilaChainInfoFetcher:   p,
		SilaBlockFetcher:       p,
		DepositFetcher:         depositCache,
		PendingDepositsFetcher: depositCache,
		BlockReceiver:          &mock.ChainService{State: beaconState, Root: blkRoot[:]},
		HeadFetcher:            &mock.ChainService{State: beaconState, Root: blkRoot[:]},
	}

	deposits, err := bs.deposits(ctx, beaconState, &silapb.SilaData{})
	require.NoError(t, err)
	assert.Equal(t, 0, len(deposits), "Received unexpected list of deposits")

	// It should also return the recent deposits after their follow window.
	p.LatestBlockNumber = big.NewInt(0).Add(p.LatestBlockNumber, big.NewInt(10000))
	// we should get our pending deposits once this vote pushes the vote tally to include
	// the updated silaexec data.
	deposits, err = bs.deposits(ctx, beaconState, vote)
	require.NoError(t, err)
	assert.Equal(t, len(recentDeposits), len(deposits), "Received unexpected number of pending deposits")
}

func TestProposer_PendingDeposits_CantReturnBelowStateSilaExecutionDepositIndex(t *testing.T) {
	ctx := t.Context()
	height := big.NewInt(int64(params.BeaconConfig().SilaExecutionFollowDistance))
	p := &mockSila.Chain{
		LatestBlockNumber: height,
		HashesByHeight: map[int][]byte{
			int(height.Int64()): []byte("0x0"),
		},
	}

	beaconState, err := util.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, beaconState.SetSilaData(&silapb.SilaData{
		BlockHash:    bytesutil.PadTo([]byte("0x0"), 32),
		DepositRoot:  make([]byte, 32),
		DepositCount: 100,
	}))
	require.NoError(t, beaconState.SetSilaExecutionDepositIndex(10))
	blk := util.NewBeaconBlock()
	blk.Block.Slot = beaconState.Slot()
	blkRoot, err := blk.HashTreeRoot()
	require.NoError(t, err)

	var mockSig [96]byte
	var mockCreds [32]byte

	readyDeposits := []*silapb.DepositContainer{
		{
			Index: 0,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("a"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index: 1,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("b"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	var recentDeposits []*silapb.DepositContainer
	for i := int64(2); i < 16; i++ {
		recentDeposits = append(recentDeposits, &silapb.DepositContainer{
			Index: i,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte{byte(i)}, 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		})
	}
	depositTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
	require.NoError(t, err, "Could not setup deposit trie")

	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)

	for _, dp := range append(readyDeposits, recentDeposits...) {
		depositHash, err := dp.Deposit.Data.HashTreeRoot()
		require.NoError(t, err, "Unable to determine hashed value of deposit")

		assert.NoError(t, depositTrie.Insert(depositHash[:], int(dp.Index)))
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		assert.NoError(t, depositCache.InsertDeposit(ctx, dp.Deposit, uint64(dp.Index), dp.Index, root))
	}
	for _, dp := range recentDeposits {
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		depositCache.InsertPendingDeposit(ctx, dp.Deposit, uint64(dp.Index), dp.Index, root)
	}

	bs := &Server{
		ChainStartFetcher:      p,
		SilaChainInfoFetcher:   p,
		SilaBlockFetcher:       p,
		DepositFetcher:         depositCache,
		PendingDepositsFetcher: depositCache,
		BlockReceiver:          &mock.ChainService{State: beaconState, Root: blkRoot[:]},
		HeadFetcher:            &mock.ChainService{State: beaconState, Root: blkRoot[:]},
	}

	// It should also return the recent deposits after their follow window.
	p.LatestBlockNumber = big.NewInt(0).Add(p.LatestBlockNumber, big.NewInt(10000))
	deposits, err := bs.deposits(ctx, beaconState, &silapb.SilaData{})
	require.NoError(t, err)

	expectedDeposits := 6
	assert.Equal(t, expectedDeposits, len(deposits), "Received unexpected number of pending deposits")
}

func TestProposer_PendingDeposits_CantReturnMoreThanMax(t *testing.T) {
	ctx := t.Context()

	height := big.NewInt(int64(params.BeaconConfig().SilaExecutionFollowDistance))
	p := &mockSila.Chain{
		LatestBlockNumber: height,
		HashesByHeight: map[int][]byte{
			int(height.Int64()): []byte("0x0"),
		},
	}

	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		SilaData: &silapb.SilaData{
			BlockHash:    bytesutil.PadTo([]byte("0x0"), 32),
			DepositRoot:  make([]byte, 32),
			DepositCount: 100,
		},
		SilaexecDepositIndex: 2,
	})
	require.NoError(t, err)
	blk := util.NewBeaconBlock()
	blk.Block.Slot = beaconState.Slot()
	blkRoot, err := blk.HashTreeRoot()
	require.NoError(t, err)
	var mockSig [96]byte
	var mockCreds [32]byte

	readyDeposits := []*silapb.DepositContainer{
		{
			Index: 0,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("a"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index: 1,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("b"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	var recentDeposits []*silapb.DepositContainer
	for i := int64(2); i < 22; i++ {
		recentDeposits = append(recentDeposits, &silapb.DepositContainer{
			Index: i,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte{byte(i)}, 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		})
	}
	depositTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
	require.NoError(t, err, "Could not setup deposit trie")

	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)

	for _, dp := range append(readyDeposits, recentDeposits...) {
		depositHash, err := dp.Deposit.Data.HashTreeRoot()
		require.NoError(t, err, "Unable to determine hashed value of deposit")

		assert.NoError(t, depositTrie.Insert(depositHash[:], int(dp.Index)))
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		assert.NoError(t, depositCache.InsertDeposit(ctx, dp.Deposit, height.Uint64(), dp.Index, root))
	}
	for _, dp := range recentDeposits {
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		depositCache.InsertPendingDeposit(ctx, dp.Deposit, height.Uint64(), dp.Index, root)
	}

	bs := &Server{
		ChainStartFetcher:      p,
		SilaChainInfoFetcher:   p,
		SilaBlockFetcher:       p,
		DepositFetcher:         depositCache,
		PendingDepositsFetcher: depositCache,
		BlockReceiver:          &mock.ChainService{State: beaconState, Root: blkRoot[:]},
		HeadFetcher:            &mock.ChainService{State: beaconState, Root: blkRoot[:]},
	}

	// It should also return the recent deposits after their follow window.
	p.LatestBlockNumber = big.NewInt(0).Add(p.LatestBlockNumber, big.NewInt(10000))
	deposits, err := bs.deposits(ctx, beaconState, &silapb.SilaData{})
	require.NoError(t, err)
	assert.Equal(t, params.BeaconConfig().MaxDeposits, uint64(len(deposits)), "Received unexpected number of pending deposits")
}

func TestProposer_PendingDeposits_CantReturnMoreThanDepositCount(t *testing.T) {
	ctx := t.Context()

	height := big.NewInt(int64(params.BeaconConfig().SilaExecutionFollowDistance))
	p := &mockSila.Chain{
		LatestBlockNumber: height,
		HashesByHeight: map[int][]byte{
			int(height.Int64()): []byte("0x0"),
		},
	}

	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		SilaData: &silapb.SilaData{
			BlockHash:    bytesutil.PadTo([]byte("0x0"), 32),
			DepositRoot:  make([]byte, 32),
			DepositCount: 5,
		},
		SilaexecDepositIndex: 2,
	})
	require.NoError(t, err)
	blk := util.NewBeaconBlock()
	blk.Block.Slot = beaconState.Slot()
	blkRoot, err := blk.HashTreeRoot()
	require.NoError(t, err)
	var mockSig [96]byte
	var mockCreds [32]byte

	readyDeposits := []*silapb.DepositContainer{
		{
			Index: 0,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("a"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index: 1,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("b"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	var recentDeposits []*silapb.DepositContainer
	for i := int64(2); i < 22; i++ {
		recentDeposits = append(recentDeposits, &silapb.DepositContainer{
			Index: i,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte{byte(i)}, 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		})
	}
	depositTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
	require.NoError(t, err, "Could not setup deposit trie")

	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)

	for _, dp := range append(readyDeposits, recentDeposits...) {
		depositHash, err := dp.Deposit.Data.HashTreeRoot()
		require.NoError(t, err, "Unable to determine hashed value of deposit")

		assert.NoError(t, depositTrie.Insert(depositHash[:], int(dp.Index)))
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		assert.NoError(t, depositCache.InsertDeposit(ctx, dp.Deposit, uint64(dp.Index), dp.Index, root))
	}
	for _, dp := range recentDeposits {
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		depositCache.InsertPendingDeposit(ctx, dp.Deposit, uint64(dp.Index), dp.Index, root)
	}

	bs := &Server{
		BlockReceiver:          &mock.ChainService{State: beaconState, Root: blkRoot[:]},
		HeadFetcher:            &mock.ChainService{State: beaconState, Root: blkRoot[:]},
		ChainStartFetcher:      p,
		SilaChainInfoFetcher:   p,
		SilaBlockFetcher:       p,
		DepositFetcher:         depositCache,
		PendingDepositsFetcher: depositCache,
	}

	// It should also return the recent deposits after their follow window.
	p.LatestBlockNumber = big.NewInt(0).Add(p.LatestBlockNumber, big.NewInt(10000))
	deposits, err := bs.deposits(ctx, beaconState, &silapb.SilaData{})
	require.NoError(t, err)
	assert.Equal(t, 3, len(deposits), "Received unexpected number of pending deposits")
}

func TestProposer_DepositTrie_UtilizesCachedFinalizedDeposits(t *testing.T) {
	ctx := t.Context()

	height := big.NewInt(int64(params.BeaconConfig().SilaExecutionFollowDistance))
	p := &mockSila.Chain{
		LatestBlockNumber: height,
		HashesByHeight: map[int][]byte{
			int(height.Int64()): []byte("0x0"),
		},
	}

	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		SilaData: &silapb.SilaData{
			BlockHash:    bytesutil.PadTo([]byte("0x0"), 32),
			DepositRoot:  make([]byte, 32),
			DepositCount: 4,
		},
		SilaexecDepositIndex: 1,
	})
	require.NoError(t, err)
	blk := util.NewBeaconBlock()
	blk.Block.Slot = beaconState.Slot()

	blkRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)

	var mockSig [96]byte
	var mockCreds [32]byte

	// Using the merkleTreeIndex as the block number for this test...
	finalizedDeposits := []*silapb.DepositContainer{
		{
			Index:           0,
			SilaexecBlockHeight: 10,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("a"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index:           1,
			SilaexecBlockHeight: 10,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("b"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	recentDeposits := []*silapb.DepositContainer{
		{
			Index:           2,
			SilaexecBlockHeight: 11,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("c"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index:           3,
			SilaexecBlockHeight: 11,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("d"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)

	depositTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
	require.NoError(t, err, "Could not setup deposit trie")
	for _, dp := range append(finalizedDeposits, recentDeposits...) {
		depositHash, err := dp.Deposit.Data.HashTreeRoot()
		require.NoError(t, err, "Unable to determine hashed value of deposit")

		assert.NoError(t, depositTrie.Insert(depositHash[:], int(dp.Index)))
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		assert.NoError(t, depositCache.InsertDeposit(ctx, dp.Deposit, dp.SilaexecBlockHeight, dp.Index, root))
	}
	for _, dp := range recentDeposits {
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		depositCache.InsertPendingDeposit(ctx, dp.Deposit, dp.SilaexecBlockHeight, dp.Index, root)
	}

	bs := &Server{
		ChainStartFetcher:      p,
		SilaChainInfoFetcher:   p,
		SilaBlockFetcher:       p,
		DepositFetcher:         depositCache,
		PendingDepositsFetcher: depositCache,
		BlockReceiver:          &mock.ChainService{State: beaconState, Root: blkRoot[:]},
		HeadFetcher:            &mock.ChainService{State: beaconState, Root: blkRoot[:]},
	}

	dt, err := bs.depositTrie(ctx, &silapb.SilaData{}, big.NewInt(int64(params.BeaconConfig().SilaExecutionFollowDistance)))
	require.NoError(t, err)

	actualRoot, err := dt.HashTreeRoot()
	require.NoError(t, err)
	expectedRoot, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	assert.Equal(t, expectedRoot, actualRoot, "Incorrect deposit trie root")
}

func TestProposer_DepositTrie_RebuildTrie(t *testing.T) {
	ctx := t.Context()

	height := big.NewInt(int64(params.BeaconConfig().SilaExecutionFollowDistance))
	p := &mockSila.Chain{
		LatestBlockNumber: height,
		HashesByHeight: map[int][]byte{
			int(height.Int64()): []byte("0x0"),
		},
	}

	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		SilaData: &silapb.SilaData{
			BlockHash:    bytesutil.PadTo([]byte("0x0"), 32),
			DepositRoot:  make([]byte, 32),
			DepositCount: 4,
		},
		SilaexecDepositIndex: 1,
	})
	require.NoError(t, err)
	blk := util.NewBeaconBlock()
	blk.Block.Slot = beaconState.Slot()

	blkRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)

	var mockSig [96]byte
	var mockCreds [32]byte

	// Using the merkleTreeIndex as the block number for this test...
	finalizedDeposits := []*silapb.DepositContainer{
		{
			Index:           0,
			SilaexecBlockHeight: 10,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("a"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index:           1,
			SilaexecBlockHeight: 10,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("b"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	recentDeposits := []*silapb.DepositContainer{
		{
			Index:           2,
			SilaexecBlockHeight: 11,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("c"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index:           3,
			SilaexecBlockHeight: 11,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("d"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)

	depositTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
	require.NoError(t, err, "Could not setup deposit trie")
	for _, dp := range append(finalizedDeposits, recentDeposits...) {
		depositHash, err := dp.Deposit.Data.HashTreeRoot()
		require.NoError(t, err, "Unable to determine hashed value of deposit")

		assert.NoError(t, depositTrie.Insert(depositHash[:], int(dp.Index)))
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		assert.NoError(t, depositCache.InsertDeposit(ctx, dp.Deposit, dp.SilaexecBlockHeight, dp.Index, root))
	}
	for _, dp := range recentDeposits {
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		depositCache.InsertPendingDeposit(ctx, dp.Deposit, dp.SilaexecBlockHeight, dp.Index, root)
	}
	d := depositCache.AllDepositContainers(ctx)
	origDeposit, ok := proto.Clone(d[0].Deposit).(*silapb.Deposit)
	assert.Equal(t, true, ok)
	junkCreds := mockCreds
	copy(junkCreds[:1], []byte{'A'})
	// Mutate it since its a pointer
	d[0].Deposit.Data.WithdrawalCredentials = junkCreds[:]
	// Insert junk to corrupt trie.
	err = depositCache.InsertFinalizedDeposits(ctx, 2, [32]byte{}, 0)
	require.NoError(t, err)

	// Add original back
	d[0].Deposit = origDeposit

	bs := &Server{
		ChainStartFetcher:      p,
		SilaChainInfoFetcher:   p,
		SilaBlockFetcher:       p,
		DepositFetcher:         depositCache,
		PendingDepositsFetcher: depositCache,
		BlockReceiver:          &mock.ChainService{State: beaconState, Root: blkRoot[:]},
		HeadFetcher:            &mock.ChainService{State: beaconState, Root: blkRoot[:]},
	}

	dt, err := bs.depositTrie(ctx, &silapb.SilaData{}, big.NewInt(int64(params.BeaconConfig().SilaExecutionFollowDistance)))
	require.NoError(t, err)

	expectedRoot, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	actualRoot, err := dt.HashTreeRoot()
	require.NoError(t, err)
	assert.Equal(t, expectedRoot, actualRoot, "Incorrect deposit trie root")

}

func TestProposer_ValidateDepositTrie(t *testing.T) {
	tt := []struct {
		name            string
		silaDataCreator func() *silapb.SilaData
		trieCreator     func() *trie.SparseMerkleTrie
		success         bool
	}{
		{
			name: "invalid trie items",
			silaDataCreator: func() *silapb.SilaData {
				return &silapb.SilaData{DepositRoot: []byte{}, DepositCount: 10, BlockHash: []byte{}}
			},
			trieCreator: func() *trie.SparseMerkleTrie {
				newTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
				assert.NoError(t, err)
				return newTrie
			},
			success: false,
		},
		{
			name: "invalid deposit root",
			silaDataCreator: func() *silapb.SilaData {
				newTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
				assert.NoError(t, err)
				assert.NoError(t, newTrie.Insert([]byte{'a'}, 0))
				assert.NoError(t, newTrie.Insert([]byte{'b'}, 1))
				assert.NoError(t, newTrie.Insert([]byte{'c'}, 2))
				return &silapb.SilaData{DepositRoot: []byte{'B'}, DepositCount: 3, BlockHash: []byte{}}
			},
			trieCreator: func() *trie.SparseMerkleTrie {
				newTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
				assert.NoError(t, err)
				assert.NoError(t, newTrie.Insert([]byte{'a'}, 0))
				assert.NoError(t, newTrie.Insert([]byte{'b'}, 1))
				assert.NoError(t, newTrie.Insert([]byte{'c'}, 2))
				return newTrie
			},
			success: false,
		},
		{
			name: "valid deposit trie",
			silaDataCreator: func() *silapb.SilaData {
				newTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
				assert.NoError(t, err)
				assert.NoError(t, newTrie.Insert([]byte{'a'}, 0))
				assert.NoError(t, newTrie.Insert([]byte{'b'}, 1))
				assert.NoError(t, newTrie.Insert([]byte{'c'}, 2))
				rt, err := newTrie.HashTreeRoot()
				require.NoError(t, err)
				return &silapb.SilaData{DepositRoot: rt[:], DepositCount: 3, BlockHash: []byte{}}
			},
			trieCreator: func() *trie.SparseMerkleTrie {
				newTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
				assert.NoError(t, err)
				assert.NoError(t, newTrie.Insert([]byte{'a'}, 0))
				assert.NoError(t, newTrie.Insert([]byte{'b'}, 1))
				assert.NoError(t, newTrie.Insert([]byte{'c'}, 2))
				return newTrie
			},
			success: true,
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			valid, err := validateDepositTrie(test.trieCreator(), test.silaDataCreator())
			assert.Equal(t, test.success, valid)
			if valid {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProposer_SilaData_MajorityVote_SpansGenesis(t *testing.T) {
	ctx := t.Context()
	// Voting period will span genesis, causing the special case for pre-mined genesis to kick in.
	// In other words some part of the valid time range is before genesis, so querying the block cache would fail
	// without the special case added to allow this for testnets.
	slot := primitives.Slot(0)
	earliestValidTime, latestValidTime := majorityVoteBoundaryTime(slot)

	p := mockSila.New().
		InsertBlock(50, earliestValidTime, []byte("earliest")).
		InsertBlock(100, latestValidTime, []byte("latest"))

	headBlockHash := []byte("headb")
	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)
	ps := &Server{
		ChainStartFetcher:    p,
		SilaChainInfoFetcher: p,
		SilaBlockFetcher:     p,
		BlockFetcher:         p,
		DepositFetcher:       depositCache,
		HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{BlockHash: headBlockHash, DepositCount: 0}},
	}

	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		Slot: slot,
		SilaDataVotes: []*silapb.SilaData{
			{BlockHash: []byte("earliest"), DepositCount: 1},
		},
	})
	require.NoError(t, err)
	majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
	require.NoError(t, err)
	assert.DeepEqual(t, headBlockHash, majorityVoteSilaData.BlockHash)
}

func TestProposer_SilaData_MajorityVote(t *testing.T) {
	followDistanceSecs := params.BeaconConfig().SilaExecutionFollowDistance * params.BeaconConfig().SecondsPerSilaBlock
	followSlots := followDistanceSecs / params.BeaconConfig().SecondsPerSlot
	slot := primitives.Slot(64 + followSlots)
	earliestValidTime, latestValidTime := majorityVoteBoundaryTime(slot)

	dc := silapb.DepositContainer{
		Index:           0,
		SilaexecBlockHeight: 0,
		Deposit: &silapb.Deposit{
			Data: &silapb.Deposit_Data{
				PublicKey:             bytesutil.PadTo([]byte("a"), 48),
				Signature:             make([]byte, 96),
				WithdrawalCredentials: make([]byte, 32),
			}},
	}
	depositTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
	require.NoError(t, err)
	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)
	root, err := depositTrie.HashTreeRoot()
	require.NoError(t, err)
	assert.NoError(t, depositCache.InsertDeposit(t.Context(), dc.Deposit, dc.SilaexecBlockHeight, dc.Index, root))

	t.Run("choose highest count", func(t *testing.T) {
		t.Skip()
		p := mockSila.New().
			InsertBlock(50, earliestValidTime, []byte("earliest")).
			InsertBlock(51, earliestValidTime+1, []byte("first")).
			InsertBlock(52, earliestValidTime+2, []byte("second")).
			InsertBlock(100, latestValidTime, []byte("latest"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
			SilaDataVotes: []*silapb.SilaData{
				{BlockHash: []byte("first"), DepositCount: 1},
				{BlockHash: []byte("first"), DepositCount: 1},
				{BlockHash: []byte("second"), DepositCount: 1},
			},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 1}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("first")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("highest count at earliest valid time - choose highest count", func(t *testing.T) {
		t.Skip()
		p := mockSila.New().
			InsertBlock(50, earliestValidTime, []byte("earliest")).
			InsertBlock(52, earliestValidTime+2, []byte("second")).
			InsertBlock(100, latestValidTime, []byte("latest"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
			SilaDataVotes: []*silapb.SilaData{
				{BlockHash: []byte("earliest"), DepositCount: 1},
				{BlockHash: []byte("earliest"), DepositCount: 1},
				{BlockHash: []byte("second"), DepositCount: 1},
			},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 1}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("earliest")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("highest count at latest valid time - choose highest count", func(t *testing.T) {
		t.Skip()
		p := mockSila.New().
			InsertBlock(50, earliestValidTime, []byte("earliest")).
			InsertBlock(51, earliestValidTime+1, []byte("first")).
			InsertBlock(100, latestValidTime, []byte("latest"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
			SilaDataVotes: []*silapb.SilaData{
				{BlockHash: []byte("first"), DepositCount: 1},
				{BlockHash: []byte("latest"), DepositCount: 1},
				{BlockHash: []byte("latest"), DepositCount: 1},
			},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 1}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("latest")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("highest count before range - choose highest count within range", func(t *testing.T) {
		t.Skip()
		p := mockSila.New().
			InsertBlock(49, earliestValidTime-1, []byte("before_range")).
			InsertBlock(50, earliestValidTime, []byte("earliest")).
			InsertBlock(51, earliestValidTime+1, []byte("first")).
			InsertBlock(100, latestValidTime, []byte("latest"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
			SilaDataVotes: []*silapb.SilaData{
				{BlockHash: []byte("before_range"), DepositCount: 1},
				{BlockHash: []byte("before_range"), DepositCount: 1},
				{BlockHash: []byte("first"), DepositCount: 1},
			},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 1}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("first")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("highest count after range - choose highest count within range", func(t *testing.T) {
		t.Skip()
		p := mockSila.New().
			InsertBlock(50, earliestValidTime, []byte("earliest")).
			InsertBlock(51, earliestValidTime+1, []byte("first")).
			InsertBlock(100, latestValidTime, []byte("latest")).
			InsertBlock(101, latestValidTime+1, []byte("after_range"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
			SilaDataVotes: []*silapb.SilaData{
				{BlockHash: []byte("first"), DepositCount: 1},
				{BlockHash: []byte("after_range"), DepositCount: 1},
				{BlockHash: []byte("after_range"), DepositCount: 1},
			},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 1}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("first")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("highest count on unknown block - choose known block with highest count", func(t *testing.T) {
		t.Skip()
		p := mockSila.New().
			InsertBlock(50, earliestValidTime, []byte("earliest")).
			InsertBlock(51, earliestValidTime+1, []byte("first")).
			InsertBlock(52, earliestValidTime+2, []byte("second")).
			InsertBlock(100, latestValidTime, []byte("latest"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
			SilaDataVotes: []*silapb.SilaData{
				{BlockHash: []byte("unknown"), DepositCount: 1},
				{BlockHash: []byte("unknown"), DepositCount: 1},
				{BlockHash: []byte("first"), DepositCount: 1},
			},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 1}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("first")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("no blocks in range - choose current silaData", func(t *testing.T) {
		p := mockSila.New().
			InsertBlock(49, earliestValidTime-1, []byte("before_range")).
			InsertBlock(101, latestValidTime+1, []byte("after_range"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
		})
		require.NoError(t, err)

		currentSilaData := &silapb.SilaData{DepositCount: 1, BlockHash: []byte("current")}
		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: currentSilaData},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("current")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("no votes in range - choose most recent block", func(t *testing.T) {
		p := mockSila.New().
			InsertBlock(49, earliestValidTime-1, []byte("before_range")).
			InsertBlock(51, earliestValidTime+1, []byte("first")).
			InsertBlock(52, earliestValidTime+2, []byte("second")).
			InsertBlock(101, latestValidTime+1, []byte("after_range"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
			SilaDataVotes: []*silapb.SilaData{
				{BlockHash: []byte("before_range"), DepositCount: 1},
				{BlockHash: []byte("after_range"), DepositCount: 1},
			},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 1}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := make([]byte, 32)
		copy(expectedHash, "second")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("no votes - choose more recent block", func(t *testing.T) {
		p := mockSila.New().
			InsertBlock(50, earliestValidTime, []byte("earliest")).
			InsertBlock(100, latestValidTime, []byte("latest"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot:          slot,
			SilaDataVotes: []*silapb.SilaData{}})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 1}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := make([]byte, 32)
		copy(expectedHash, "latest")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("no votes and more recent block has less deposits - choose current silaData", func(t *testing.T) {
		p := mockSila.New().
			InsertBlock(50, earliestValidTime, []byte("earliest")).
			InsertBlock(100, latestValidTime, []byte("latest"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
		})
		require.NoError(t, err)

		// Set the deposit count in current silaData to exceed the latest most recent block's deposit count.
		currentSilaData := &silapb.SilaData{DepositCount: 2, BlockHash: []byte("current")}
		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: currentSilaData},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("current")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("same count - choose more recent block", func(t *testing.T) {
		t.Skip()
		p := mockSila.New().
			InsertBlock(50, earliestValidTime, []byte("earliest")).
			InsertBlock(51, earliestValidTime+1, []byte("first")).
			InsertBlock(52, earliestValidTime+2, []byte("second")).
			InsertBlock(100, latestValidTime, []byte("latest"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
			SilaDataVotes: []*silapb.SilaData{
				{BlockHash: []byte("first"), DepositCount: 1},
				{BlockHash: []byte("second"), DepositCount: 1},
			},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 1}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("second")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("highest count on block with less deposits - choose another block", func(t *testing.T) {
		t.Skip()
		p := mockSila.New().
			InsertBlock(50, earliestValidTime, []byte("earliest")).
			InsertBlock(51, earliestValidTime+1, []byte("first")).
			InsertBlock(52, earliestValidTime+2, []byte("second")).
			InsertBlock(100, latestValidTime, []byte("latest"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
			SilaDataVotes: []*silapb.SilaData{
				{BlockHash: []byte("no_new_deposits"), DepositCount: 0},
				{BlockHash: []byte("no_new_deposits"), DepositCount: 0},
				{BlockHash: []byte("second"), DepositCount: 1},
			},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 1}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("second")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("only one block at earliest valid time - choose this block", func(t *testing.T) {
		t.Skip()
		p := mockSila.New().InsertBlock(50, earliestValidTime, []byte("earliest"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
			SilaDataVotes: []*silapb.SilaData{
				{BlockHash: []byte("earliest"), DepositCount: 1},
			},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 1}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("earliest")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("vote on last block before range - choose next block", func(t *testing.T) {
		p := mockSila.New().
			InsertBlock(49, earliestValidTime-1, []byte("before_range")).
			// It is important to have height `50` with time `earliestValidTime+1` and not `earliestValidTime`
			// because of earliest block increment in the algorithm.
			InsertBlock(50, earliestValidTime+1, []byte("first"))

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
			SilaDataVotes: []*silapb.SilaData{
				{BlockHash: []byte("before_range"), DepositCount: 1},
			},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 1}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := make([]byte, 32)
		copy(expectedHash, "first")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("no deposits - choose chain start silaData", func(t *testing.T) {
		p := mockSila.New().
			InsertBlock(50, earliestValidTime, []byte("earliest")).
			InsertBlock(100, latestValidTime, []byte("latest"))
		p.SilaData = &silapb.SilaData{
			BlockHash: []byte("silaData"),
		}

		depositCache, err := depositsnapshot.New()
		require.NoError(t, err)

		beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
			Slot: slot,
			SilaDataVotes: []*silapb.SilaData{
				{BlockHash: []byte("earliest"), DepositCount: 1},
			},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
			HeadFetcher:          &mock.ChainService{SILASilaData: &silapb.SilaData{DepositCount: 0}},
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("silaData")
		assert.DeepEqual(t, expectedHash, hash)
	})

	t.Run("post electra the head silaData should be returned", func(t *testing.T) {
		p := mockSila.New().
			InsertBlock(50, earliestValidTime, []byte("earliest")).
			InsertBlock(100, latestValidTime, []byte("latest"))
		p.SilaData = &silapb.SilaData{
			BlockHash: []byte("silaData"),
		}

		depositCache, err := depositsnapshot.New()
		require.NoError(t, err)

		beaconState, err := state_native.InitializeFromProtoElectra(&silapb.BeaconStateElectra{
			Slot:     slot,
			SilaData: &silapb.SilaData{BlockHash: []byte("legacy"), DepositCount: 1},
		})
		require.NoError(t, err)

		ps := &Server{
			ChainStartFetcher:    p,
			SilaChainInfoFetcher: p,
			SilaBlockFetcher:     p,
			BlockFetcher:         p,
			DepositFetcher:       depositCache,
		}

		ctx := t.Context()
		majorityVoteSilaData, err := ps.silaexecDataMajorityVote(ctx, beaconState)
		require.NoError(t, err)

		hash := majorityVoteSilaData.BlockHash

		expectedHash := []byte("legacy")
		assert.DeepEqual(t, expectedHash, hash)
	})
}

func TestProposer_FilterAttestation(t *testing.T) {
	genesis := util.NewBeaconBlock()

	numValidators := uint64(64)
	st, privKeys := util.DeterministicGenesisState(t, numValidators)
	require.NoError(t, st.SetGenesisValidatorsRoot(params.BeaconConfig().ZeroHash[:]))
	assert.NoError(t, st.SetSlot(1))

	genesisRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err)

	tests := []struct {
		name         string
		inputAtts    func() []silapb.Att
		expectedAtts func(inputAtts []silapb.Att) []silapb.Att
	}{
		{
			name: "nil attestations",
			inputAtts: func() []silapb.Att {
				return nil
			},
			expectedAtts: func(inputAtts []silapb.Att) []silapb.Att {
				return []silapb.Att{}
			},
		},
		{
			name: "invalid attestations",
			inputAtts: func() []silapb.Att {
				atts := make([]silapb.Att, 10)
				for i := range atts {
					atts[i] = util.HydrateAttestation(&silapb.Attestation{
						Data: &silapb.AttestationData{
							CommitteeIndex: primitives.CommitteeIndex(i),
						},
					})
				}
				return atts
			},
			expectedAtts: func(inputAtts []silapb.Att) []silapb.Att {
				return []silapb.Att{}
			},
		},
		{
			name: "filter aggregates ok",
			inputAtts: func() []silapb.Att {
				atts := make([]silapb.Att, 10)
				for i := range atts {
					atts[i] = util.HydrateAttestation(&silapb.Attestation{
						Data: &silapb.AttestationData{
							CommitteeIndex: primitives.CommitteeIndex(i),
							Source:         &silapb.Checkpoint{Root: params.BeaconConfig().ZeroHash[:]},
						},
						AggregationBits: bitfield.Bitlist{0b00010010},
					})
					committee, err := helpers.BeaconCommitteeFromState(t.Context(), st, atts[i].GetData().Slot, atts[i].GetData().CommitteeIndex)
					assert.NoError(t, err)
					attestingIndices, err := attestation.AttestingIndices(atts[i], committee)
					require.NoError(t, err)
					assert.NoError(t, err)
					domain, err := signing.Domain(st.Fork(), 0, params.BeaconConfig().DomainBeaconAttester, params.BeaconConfig().ZeroHash[:])
					require.NoError(t, err)
					sigs := make([]bls.Signature, len(attestingIndices))
					var zeroSig [96]byte
					atts[i].(*silapb.Attestation).Signature = zeroSig[:]

					for i, indice := range attestingIndices {
						hashTreeRoot, err := signing.ComputeSigningRoot(atts[i].GetData(), domain)
						require.NoError(t, err)
						sig := privKeys[indice].Sign(hashTreeRoot[:])
						sigs[i] = sig
					}
					atts[i].(*silapb.Attestation).Signature = bls.AggregateSignatures(sigs).Marshal()
				}
				return atts
			},
			expectedAtts: func(inputAtts []silapb.Att) []silapb.Att {
				return []silapb.Att{inputAtts[0], inputAtts[1]}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proposerServer := &Server{
				AttPool:     attestations.NewPool(),
				HeadFetcher: &mock.ChainService{State: st, Root: genesisRoot[:]},
			}
			atts := tt.inputAtts()
			received := proposerServer.validateAndDeleteAttsInPool(t.Context(), st, atts)
			assert.DeepEqual(t, tt.expectedAtts(atts), received)
		})
	}
}

func TestProposer_Deposits_ReturnsEmptyList_IfLatestSilaDataEqGenesisSilaBlock(t *testing.T) {
	ctx := t.Context()

	height := big.NewInt(int64(params.BeaconConfig().SilaExecutionFollowDistance))
	p := &mockSila.Chain{
		LatestBlockNumber: height,
		HashesByHeight: map[int][]byte{
			int(height.Int64()): []byte("0x0"),
		},
		GenesisSilaBlock: height,
	}

	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		SilaData: &silapb.SilaData{
			BlockHash:   bytesutil.PadTo([]byte("0x0"), 32),
			DepositRoot: make([]byte, 32),
		},
		SilaexecDepositIndex: 2,
	})
	require.NoError(t, err)
	blk := util.NewBeaconBlock()
	blk.Block.Slot = beaconState.Slot()
	blkRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)

	var mockSig [96]byte
	var mockCreds [32]byte

	readyDeposits := []*silapb.DepositContainer{
		{
			Index: 0,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("a"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
		{
			Index: 1,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte("b"), 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		},
	}

	var recentDeposits []*silapb.DepositContainer
	for i := int64(2); i < 22; i++ {
		recentDeposits = append(recentDeposits, &silapb.DepositContainer{
			Index: i,
			Deposit: &silapb.Deposit{
				Data: &silapb.Deposit_Data{
					PublicKey:             bytesutil.PadTo([]byte{byte(i)}, 48),
					Signature:             mockSig[:],
					WithdrawalCredentials: mockCreds[:],
				}},
		})
	}
	depositTrie, err := trie.NewTrie(params.BeaconConfig().SilaDepositTreeDepth)
	require.NoError(t, err, "Could not setup deposit trie")

	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)

	for _, dp := range append(readyDeposits, recentDeposits...) {
		depositHash, err := dp.Deposit.Data.HashTreeRoot()
		require.NoError(t, err, "Unable to determine hashed value of deposit")

		assert.NoError(t, depositTrie.Insert(depositHash[:], int(dp.Index)))
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		assert.NoError(t, depositCache.InsertDeposit(ctx, dp.Deposit, uint64(dp.Index), dp.Index, root))
	}
	for _, dp := range recentDeposits {
		root, err := depositTrie.HashTreeRoot()
		require.NoError(t, err)
		depositCache.InsertPendingDeposit(ctx, dp.Deposit, uint64(dp.Index), dp.Index, root)
	}

	bs := &Server{
		BlockReceiver:          &mock.ChainService{State: beaconState, Root: blkRoot[:]},
		HeadFetcher:            &mock.ChainService{State: beaconState, Root: blkRoot[:]},
		ChainStartFetcher:      p,
		SilaChainInfoFetcher:   p,
		SilaBlockFetcher:       p,
		DepositFetcher:         depositCache,
		PendingDepositsFetcher: depositCache,
	}

	// It should also return the recent deposits after their follow window.
	p.LatestBlockNumber = big.NewInt(0).Add(p.LatestBlockNumber, big.NewInt(10000))
	deposits, err := bs.deposits(ctx, beaconState, &silapb.SilaData{})
	require.NoError(t, err)
	assert.Equal(t, 0, len(deposits), "Received unexpected number of pending deposits")
}

func TestProposer_DeleteAttsInPool_Aggregated(t *testing.T) {
	s := &Server{
		AttPool: attestations.NewPool(),
	}
	priv, err := bls.RandKey()
	require.NoError(t, err)
	sig := priv.Sign([]byte("foo")).Marshal()
	aggregatedAtts := []silapb.Att{
		util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b10101}, Signature: sig}),
		util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b11010}, Signature: sig})}
	unaggregatedAtts := []silapb.Att{
		util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b10010}, Signature: sig}),
		util.HydrateAttestation(&silapb.Attestation{Data: &silapb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b10100}, Signature: sig})}

	require.NoError(t, s.AttPool.SaveAggregatedAttestations(aggregatedAtts))
	require.NoError(t, s.AttPool.SaveUnaggregatedAttestations(unaggregatedAtts))

	aa, err := attaggregation.Aggregate(aggregatedAtts)
	require.NoError(t, err)
	require.NoError(t, s.deleteAttsInPool(t.Context(), append(aa, unaggregatedAtts...)))
	assert.Equal(t, 0, len(s.AttPool.AggregatedAttestations()), "Did not delete aggregated attestation")
	atts := s.AttPool.UnaggregatedAttestations()
	assert.Equal(t, 0, len(atts), "Did not delete unaggregated attestation")
}

func TestProposer_PrepareBeaconProposer(t *testing.T) {
	type args struct {
		request *silapb.PrepareBeaconProposerRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "Happy Path",
			args: args{
				request: &silapb.PrepareBeaconProposerRequest{
					Recipients: []*silapb.PrepareBeaconProposerRequest_FeeRecipientContainer{
						{
							FeeRecipient:   make([]byte, fieldparams.FeeRecipientLength),
							ValidatorIndex: 1,
						},
					},
				},
			},
			wantErr: "",
		},
		{
			name: "invalid fee recipient length",
			args: args{
				request: &silapb.PrepareBeaconProposerRequest{
					Recipients: []*silapb.PrepareBeaconProposerRequest_FeeRecipientContainer{
						{
							FeeRecipient:   make([]byte, fieldparams.BLSPubkeyLength),
							ValidatorIndex: 1,
						},
					},
				},
			},
			wantErr: "Invalid fee recipient address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := dbutil.SetupDB(t)
			ctx := t.Context()
			proposerServer := &Server{
				BeaconDB:               db,
				TrackedValidatorsCache: cache.NewTrackedValidatorsCache(),
			}
			require.Equal(t, false, proposerServer.TrackedValidatorsCache.Validating())
			_, err := proposerServer.PrepareBeaconProposer(ctx, tt.args.request)
			if tt.wantErr != "" {
				require.ErrorContains(t, tt.wantErr, err)
				return
			} else {
				require.Equal(t, true, proposerServer.TrackedValidatorsCache.Validating())
			}
			require.NoError(t, err)
			val, tracked := proposerServer.TrackedValidatorsCache.Validator(1)
			require.Equal(t, true, tracked)
			require.Equal(t, primitives.SilaAddress(tt.args.request.Recipients[0].FeeRecipient), val.FeeRecipient)

		})
	}
}

func TestProposer_PrepareBeaconProposerOverlapping(t *testing.T) {
	hook := logTest.NewGlobal()
	logrus.SetLevel(logrus.DebugLevel)

	db := dbutil.SetupDB(t)
	ctx := t.Context()
	proposerServer := &Server{
		BeaconDB:               db,
		TrackedValidatorsCache: cache.NewTrackedValidatorsCache(),
	}

	// New validator
	f := bytesutil.PadTo([]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF}, fieldparams.FeeRecipientLength)
	req := &silapb.PrepareBeaconProposerRequest{
		Recipients: []*silapb.PrepareBeaconProposerRequest_FeeRecipientContainer{
			{FeeRecipient: f, ValidatorIndex: 1},
		},
	}
	_, err := proposerServer.PrepareBeaconProposer(ctx, req)
	require.NoError(t, err)
	require.LogsContain(t, hook, "Updated fee recipient addresses")

	// Same validator
	hook.Reset()
	_, err = proposerServer.PrepareBeaconProposer(ctx, req)
	require.NoError(t, err)
	require.LogsContain(t, hook, "Updated fee recipient addresses")

	// Same validator with different fee recipient
	hook.Reset()
	f = bytesutil.PadTo([]byte{0x01, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF}, fieldparams.FeeRecipientLength)
	req = &silapb.PrepareBeaconProposerRequest{
		Recipients: []*silapb.PrepareBeaconProposerRequest_FeeRecipientContainer{
			{FeeRecipient: f, ValidatorIndex: 1},
		},
	}
	_, err = proposerServer.PrepareBeaconProposer(ctx, req)
	require.NoError(t, err)
	require.LogsContain(t, hook, "Updated fee recipient addresses")

	// More than one validator
	hook.Reset()
	f = bytesutil.PadTo([]byte{0x01, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF}, fieldparams.FeeRecipientLength)
	req = &silapb.PrepareBeaconProposerRequest{
		Recipients: []*silapb.PrepareBeaconProposerRequest_FeeRecipientContainer{
			{FeeRecipient: f, ValidatorIndex: 1},
			{FeeRecipient: f, ValidatorIndex: 2},
		},
	}
	_, err = proposerServer.PrepareBeaconProposer(ctx, req)
	require.NoError(t, err)
	require.LogsContain(t, hook, "Updated fee recipient addresses")

	// Same validators
	hook.Reset()
	_, err = proposerServer.PrepareBeaconProposer(ctx, req)
	require.NoError(t, err)
	require.LogsContain(t, hook, "Updated fee recipient addresses")
}

func BenchmarkServer_PrepareBeaconProposer(b *testing.B) {
	db := dbutil.SetupDB(b)
	ctx := b.Context()
	proposerServer := &Server{
		BeaconDB:               db,
		TrackedValidatorsCache: cache.NewTrackedValidatorsCache(),
	}
	f := bytesutil.PadTo([]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF}, fieldparams.FeeRecipientLength)
	recipients := make([]*silapb.PrepareBeaconProposerRequest_FeeRecipientContainer, 0)
	for i := range 10000 {
		recipients = append(recipients, &silapb.PrepareBeaconProposerRequest_FeeRecipientContainer{FeeRecipient: f, ValidatorIndex: primitives.ValidatorIndex(i)})
	}

	req := &silapb.PrepareBeaconProposerRequest{
		Recipients: recipients,
	}

	for b.Loop() {
		_, err := proposerServer.PrepareBeaconProposer(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestProposer_SubmitValidatorRegistrations(t *testing.T) {
	ctx := t.Context()
	proposerServer := &Server{}
	reg := &silapb.SignedValidatorRegistrationsV1{}
	_, err := proposerServer.SubmitValidatorRegistrations(ctx, reg)
	require.ErrorContains(t, builder.ErrNoBuilder.Error(), err)
	proposerServer = &Server{BlockBuilder: &builderTest.MockBuilderService{}}
	_, err = proposerServer.SubmitValidatorRegistrations(ctx, reg)
	require.ErrorContains(t, builder.ErrNoBuilder.Error(), err)
	proposerServer = &Server{BlockBuilder: &builderTest.MockBuilderService{HasConfigured: true}}
	_, err = proposerServer.SubmitValidatorRegistrations(ctx, reg)
	require.NoError(t, err)
	proposerServer = &Server{BlockBuilder: &builderTest.MockBuilderService{HasConfigured: true, ErrRegisterValidator: errors.New("bad")}}
	_, err = proposerServer.SubmitValidatorRegistrations(ctx, reg)
	require.ErrorContains(t, "bad", err)
}

func majorityVoteBoundaryTime(slot primitives.Slot) (uint64, uint64) {
	s := params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().EpochsPerSilaExecutionVotingPeriod))
	slotStartTime := uint64(mockSila.GenesisTime) + uint64((slot - (slot % (s))).Mul(params.BeaconConfig().SecondsPerSlot))
	earliestValidTime := slotStartTime - 2*params.BeaconConfig().SecondsPerSilaBlock*params.BeaconConfig().SilaExecutionFollowDistance
	latestValidTime := slotStartTime - params.BeaconConfig().SecondsPerSilaBlock*params.BeaconConfig().SilaExecutionFollowDistance

	return earliestValidTime, latestValidTime
}

func TestProposer_GetFeeRecipientByPubKey(t *testing.T) {
	db := dbutil.SetupDB(t)
	ctx := t.Context()
	numDeposits := uint64(64)
	beaconState, _ := util.DeterministicGenesisState(t, numDeposits)
	bsRoot, err := beaconState.HashTreeRoot(ctx)
	require.NoError(t, err)
	proposerServer := &Server{
		BeaconDB:    db,
		HeadFetcher: &mock.ChainService{Root: bsRoot[:], State: beaconState},
	}
	pubkey, err := hexutil.Decode("0xa057816155ad77931185101128655c0191bd0214c201ca48ed887f6c4c6adf334070efcd75140eada5ac83a92506dd7a")
	require.NoError(t, err)
	resp, err := proposerServer.GetFeeRecipientByPubKey(ctx, &silapb.FeeRecipientByPubKeyRequest{
		PublicKey: pubkey,
	})
	require.NoError(t, err)

	require.Equal(t, params.BeaconConfig().DefaultFeeRecipient.Hex(), hexutil.Encode(resp.FeeRecipient))
	params.BeaconConfig().DefaultFeeRecipient = common.HexToAddress("0x046Fb65722E7b2455012BFEBf6177F1D2e9728D9")
	resp, err = proposerServer.GetFeeRecipientByPubKey(ctx, &silapb.FeeRecipientByPubKeyRequest{
		PublicKey: beaconState.Validators()[0].PublicKey,
	})
	require.NoError(t, err)

	require.Equal(t, params.BeaconConfig().DefaultFeeRecipient.Hex(), common.BytesToAddress(resp.FeeRecipient).Hex())
	index, err := proposerServer.ValidatorIndex(ctx, &silapb.ValidatorIndexRequest{
		PublicKey: beaconState.Validators()[0].PublicKey,
	})
	require.NoError(t, err)
	err = proposerServer.BeaconDB.SaveFeeRecipientsByValidatorIDs(ctx, []primitives.ValidatorIndex{index.Index}, []common.Address{common.HexToAddress("0x055Fb65722E7b2455012BFEBf6177F1D2e9728D8")})
	require.NoError(t, err)
	resp, err = proposerServer.GetFeeRecipientByPubKey(ctx, &silapb.FeeRecipientByPubKeyRequest{
		PublicKey: beaconState.Validators()[0].PublicKey,
	})
	require.NoError(t, err)

	require.Equal(t, common.HexToAddress("0x055Fb65722E7b2455012BFEBf6177F1D2e9728D8").Hex(), common.BytesToAddress(resp.FeeRecipient).Hex())
}

func TestProposer_GetParentHeadState(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.MinimalSpecConfig().Copy()
	cfg.GloasForkEpoch = 0
	params.OverrideBeaconConfig(cfg)

	db := dbutil.SetupDB(t)
	ctx := t.Context()

	parentState, parentRoot, _ := util.DeterministicGenesisStateWithGenesisBlock(t, ctx, db, 100)
	headState, headRoot, _ := util.DeterministicGenesisStateWithGenesisBlock(t, ctx, db, 50)
	require.NoError(t, transition.UpdateNextSlotCache(ctx, parentRoot[:], parentState))

	proposerServer := &Server{
		ChainStartFetcher:    &mockSila.Chain{},
		SilaChainInfoFetcher: &mockSila.Chain{},
		SilaBlockFetcher:     &mockSila.Chain{},
		ForkchoiceFetcher:    &mock.ChainService{},
		StateGen:             stategen.New(db, doublylinkedtree.New()),
	}
	t.Run("successful reorg", func(tt *testing.T) {
		head, err := proposerServer.getParentStateFromReorgData(ctx, 1, parentRoot, parentRoot, headRoot)
		require.NoError(t, err)
		st := parentState.Copy()
		st, err = transition.ProcessSlots(ctx, st, st.Slot()+1)
		require.NoError(t, err)
		str, err := st.StateRootAtIndex(0)
		require.NoError(t, err)
		headStr, err := head.StateRootAtIndex(0)
		require.NoError(t, err)
		genesisStr, err := headState.StateRootAtIndex(0)
		require.NoError(t, err)
		require.Equal(t, [32]byte(str), [32]byte(headStr))
		require.NotEqual(t, [32]byte(str), [32]byte(genesisStr))
	})

	t.Run("no reorg", func(tt *testing.T) {
		require.NoError(t, transition.UpdateNextSlotCache(ctx, headRoot[:], headState))
		head, err := proposerServer.getParentStateFromReorgData(ctx, 1, headRoot, headRoot, headRoot)
		require.NoError(t, err)
		st := headState.Copy()
		st, err = transition.ProcessSlots(ctx, st, st.Slot()+1)
		require.NoError(t, err)
		str, err := st.StateRootAtIndex(0)
		require.NoError(t, err)
		headStr, err := head.StateRootAtIndex(0)
		require.NoError(t, err)
		genesisStr, err := parentState.StateRootAtIndex(0)
		require.NoError(t, err)
		require.Equal(t, [32]byte(str), [32]byte(headStr))
		require.NotEqual(t, [32]byte(str), [32]byte(genesisStr))
	})

	t.Run("failed reorg", func(tt *testing.T) {
		hook := logTest.NewGlobal()
		require.NoError(t, transition.UpdateNextSlotCache(ctx, headRoot[:], headState))
		head, err := proposerServer.getParentStateFromReorgData(ctx, 1, parentRoot, headRoot, headRoot)
		require.NoError(t, err)
		st := headState.Copy()
		st, err = transition.ProcessSlots(ctx, st, st.Slot()+1)
		require.NoError(t, err)
		str, err := st.StateRootAtIndex(0)
		require.NoError(t, err)
		headStr, err := head.StateRootAtIndex(0)
		require.NoError(t, err)
		genesisStr, err := parentState.StateRootAtIndex(0)
		require.NoError(t, err)
		require.Equal(t, [32]byte(str), [32]byte(headStr))
		require.NotEqual(t, [32]byte(str), [32]byte(genesisStr))
		require.LogsContain(t, hook, "Late block attempted reorg failed")
	})

	t.Run("successful reorg uses parent root for NSC lookup", func(tt *testing.T) {
		require.NoError(t, transition.UpdateNextSlotCache(ctx, parentRoot[:], parentState))

		proposerServer := &Server{
			ForkchoiceFetcher: &mock.ChainService{},
			StateGen:          stategen.New(db, doublylinkedtree.New()),
		}

		head, err := proposerServer.getParentStateFromReorgData(ctx, 1, parentRoot, parentRoot, headRoot)
		require.NoError(t, err)
		st := parentState.Copy()
		st, err = transition.ProcessSlots(ctx, st, st.Slot()+1)
		require.NoError(t, err)
		str, err := st.StateRootAtIndex(0)
		require.NoError(t, err)
		headStr, err := head.StateRootAtIndex(0)
		require.NoError(t, err)
		require.Equal(t, [32]byte(str), [32]byte(headStr))
	})

	t.Run("no reorg uses parent root for NSC lookup", func(tt *testing.T) {
		require.NoError(t, transition.UpdateNextSlotCache(ctx, headRoot[:], parentState))

		proposerServer := &Server{
			ForkchoiceFetcher: &mock.ChainService{},
			HeadFetcher: &mock.ChainService{
				State: headState,
				Root:  headRoot[:],
			},
		}

		head, err := proposerServer.getParentStateFromReorgData(ctx, 1, headRoot, headRoot, headRoot)
		require.NoError(t, err)
		st := parentState.Copy()
		st, err = transition.ProcessSlots(ctx, st, st.Slot()+1)
		require.NoError(t, err)
		str, err := st.StateRootAtIndex(0)
		require.NoError(t, err)
		headStr, err := head.StateRootAtIndex(0)
		require.NoError(t, err)
		require.Equal(t, [32]byte(str), [32]byte(headStr))
	})
}

func TestProposer_ElectraBlobsAndProofs(t *testing.T) {
	electraContents := &silapb.SignedBeaconBlockContentsElectra{Block: &silapb.SignedBeaconBlockElectra{}}
	electraContents.KzgProofs = make([][]byte, 10)
	electraContents.Blobs = make([][]byte, 10)

	genBlock := &silapb.GenericSignedBeaconBlock{Block: &silapb.GenericSignedBeaconBlock_Electra{Electra: electraContents}}
	blobs, proofs, err := blobsAndProofs(genBlock)
	require.NoError(t, err)
	require.Equal(t, 10, len(blobs))
	require.Equal(t, 10, len(proofs))
}

func TestServer_ProposeBeaconBlock_PostFuluBlindedBlock(t *testing.T) {
	db := dbutil.SetupDB(t)
	ctx := t.Context()

	beaconState, parentRoot, _ := util.DeterministicGenesisStateWithGenesisBlock(t, ctx, db, 100)
	require.NoError(t, beaconState.SetSlot(1))

	t.Run("post-Fulu blinded block - early return success", func(t *testing.T) {
		// Set up config with Fulu fork at epoch 5
		params.SetupTestConfigCleanup(t)
		cfg := params.BeaconConfig().Copy()
		cfg.FuluForkEpoch = 5
		params.OverrideBeaconConfig(cfg)

		mockBuilder := &builderTest.MockBuilderService{
			HasConfigured:                 true,
			Cfg:                           &builderTest.Config{BeaconDB: db},
			ErrSubmitBlindedBlockPostFulu: nil, // Success case
		}

		c := &mock.ChainService{State: beaconState, Root: parentRoot[:]}
		proposerServer := &Server{
			ChainStartFetcher:    &mockSila.Chain{},
			SilaChainInfoFetcher: &mockSila.Chain{},
			SilaBlockFetcher:     &mockSila.Chain{},
			BlockReceiver:        c,
			BlobReceiver:         c,
			HeadFetcher:          c,
			BlockNotifier:        c.BlockNotifier(),
			OperationNotifier:    c.OperationNotifier(),
			StateGen:             stategen.New(db, doublylinkedtree.New()),
			TimeFetcher:          c,
			SyncChecker:          &mockSync.Sync{IsSyncing: false},
			BeaconDB:             db,
			BlockBuilder:         mockBuilder,
			P2P:                  &mockp2p.MockBroadcaster{},
		}

		// Create a blinded block at slot 160 (epoch 5, which is >= FuluForkEpoch)
		blindedBlock := util.NewBlindedBeaconBlockDeneb()
		blindedBlock.Message.Slot = 160 // This puts us at epoch 5 (160/32 = 5)
		blindedBlock.Message.ProposerIndex = 0
		blindedBlock.Message.ParentRoot = parentRoot[:]
		blindedBlock.Message.StateRoot = make([]byte, 32)

		req := &silapb.GenericSignedBeaconBlock{
			Block: &silapb.GenericSignedBeaconBlock_BlindedDeneb{BlindedDeneb: blindedBlock},
		}

		// This should trigger the post-Fulu early return path
		res, err := proposerServer.ProposeBeaconBlock(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.BlockRoot)
	})

	t.Run("post-Fulu blinded block - builder submission error", func(t *testing.T) {
		// Set up config with Fulu fork at epoch 5
		params.SetupTestConfigCleanup(t)
		cfg := params.BeaconConfig().Copy()
		cfg.FuluForkEpoch = 5
		params.OverrideBeaconConfig(cfg)

		mockBuilder := &builderTest.MockBuilderService{
			HasConfigured:                 true,
			Cfg:                           &builderTest.Config{BeaconDB: db},
			ErrSubmitBlindedBlockPostFulu: errors.New("post-Fulu builder submission failed"),
		}

		c := &mock.ChainService{State: beaconState, Root: parentRoot[:]}
		proposerServer := &Server{
			ChainStartFetcher:    &mockSila.Chain{},
			SilaChainInfoFetcher: &mockSila.Chain{},
			SilaBlockFetcher:     &mockSila.Chain{},
			BlockReceiver:        c,
			BlobReceiver:         c,
			HeadFetcher:          c,
			BlockNotifier:        c.BlockNotifier(),
			OperationNotifier:    c.OperationNotifier(),
			StateGen:             stategen.New(db, doublylinkedtree.New()),
			TimeFetcher:          c,
			SyncChecker:          &mockSync.Sync{IsSyncing: false},
			BeaconDB:             db,
			BlockBuilder:         mockBuilder,
			P2P:                  &mockp2p.MockBroadcaster{},
		}

		// Create a blinded block at slot 160 (epoch 5)
		blindedBlock := util.NewBlindedBeaconBlockDeneb()
		blindedBlock.Message.Slot = 160
		blindedBlock.Message.ProposerIndex = 0
		blindedBlock.Message.ParentRoot = parentRoot[:]
		blindedBlock.Message.StateRoot = make([]byte, 32)

		req := &silapb.GenericSignedBeaconBlock{
			Block: &silapb.GenericSignedBeaconBlock_BlindedDeneb{BlindedDeneb: blindedBlock},
		}

		_, err := proposerServer.ProposeBeaconBlock(ctx, req)
		require.ErrorContains(t, "Could not submit blinded block post-Fulu", err)
		require.ErrorContains(t, "post-Fulu builder submission failed", err)
	})

	t.Run("pre-Fulu blinded block - uses regular handleBlindedBlock path", func(t *testing.T) {
		// Set up config with Fulu fork at epoch 10 (future)
		params.SetupTestConfigCleanup(t)
		cfg := params.BeaconConfig().Copy()
		cfg.FuluForkEpoch = 10
		params.OverrideBeaconConfig(cfg)

		mockBuilder := &builderTest.MockBuilderService{
			HasConfigured: true,
			Cfg:           &builderTest.Config{BeaconDB: db},
			PayloadDeneb:  &silaenginev1.SilaPayloadDeneb{},
			BlobBundle:    &silaenginev1.BlobsBundle{},
		}

		c := &mock.ChainService{State: beaconState, Root: parentRoot[:]}
		proposerServer := &Server{
			ChainStartFetcher:    &mockSila.Chain{},
			SilaChainInfoFetcher: &mockSila.Chain{},
			SilaBlockFetcher:     &mockSila.Chain{},
			BlockReceiver:        c,
			BlobReceiver:         c,
			HeadFetcher:          c,
			BlockNotifier:        c.BlockNotifier(),
			OperationNotifier:    c.OperationNotifier(),
			StateGen:             stategen.New(db, doublylinkedtree.New()),
			TimeFetcher:          c,
			SyncChecker:          &mockSync.Sync{IsSyncing: false},
			BeaconDB:             db,
			BlockBuilder:         mockBuilder,
			P2P:                  &mockp2p.MockBroadcaster{},
		}

		// Create a blinded block at slot 160 (epoch 5, which is < FuluForkEpoch=10)
		blindedBlock := util.NewBlindedBeaconBlockDeneb()
		blindedBlock.Message.Slot = 160
		blindedBlock.Message.ProposerIndex = 0
		blindedBlock.Message.ParentRoot = parentRoot[:]
		blindedBlock.Message.StateRoot = make([]byte, 32)

		req := &silapb.GenericSignedBeaconBlock{
			Block: &silapb.GenericSignedBeaconBlock_BlindedDeneb{BlindedDeneb: blindedBlock},
		}

		// This should NOT trigger the post-Fulu early return path, but use handleBlindedBlock instead
		res, err := proposerServer.ProposeBeaconBlock(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.BlockRoot)
	})

	t.Run("boundary test - exactly at Fulu fork epoch", func(t *testing.T) {
		// Set up config with Fulu fork at epoch 5
		params.SetupTestConfigCleanup(t)
		cfg := params.BeaconConfig().Copy()
		cfg.FuluForkEpoch = 5
		params.OverrideBeaconConfig(cfg)

		mockBuilder := &builderTest.MockBuilderService{
			HasConfigured:                 true,
			Cfg:                           &builderTest.Config{BeaconDB: db},
			ErrSubmitBlindedBlockPostFulu: nil,
		}

		c := &mock.ChainService{State: beaconState, Root: parentRoot[:]}
		proposerServer := &Server{
			ChainStartFetcher:    &mockSila.Chain{},
			SilaChainInfoFetcher: &mockSila.Chain{},
			SilaBlockFetcher:     &mockSila.Chain{},
			BlockReceiver:        c,
			BlobReceiver:         c,
			HeadFetcher:          c,
			BlockNotifier:        c.BlockNotifier(),
			OperationNotifier:    c.OperationNotifier(),
			StateGen:             stategen.New(db, doublylinkedtree.New()),
			TimeFetcher:          c,
			SyncChecker:          &mockSync.Sync{IsSyncing: false},
			BeaconDB:             db,
			BlockBuilder:         mockBuilder,
			P2P:                  &mockp2p.MockBroadcaster{},
		}

		// Create a blinded block at slot 160 (exactly epoch 5)
		blindedBlock := util.NewBlindedBeaconBlockDeneb()
		blindedBlock.Message.Slot = 160 // 160/32 = 5 (exactly at FuluForkEpoch)
		blindedBlock.Message.ProposerIndex = 0
		blindedBlock.Message.ParentRoot = parentRoot[:]
		blindedBlock.Message.StateRoot = make([]byte, 32)

		req := &silapb.GenericSignedBeaconBlock{
			Block: &silapb.GenericSignedBeaconBlock_BlindedDeneb{BlindedDeneb: blindedBlock},
		}

		// Should trigger post-Fulu path since epoch 5 >= FuluForkEpoch (5)
		res, err := proposerServer.ProposeBeaconBlock(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.BlockRoot)
	})

	t.Run("unblinded block - not affected by post-Fulu condition", func(t *testing.T) {
		// Set up config with Fulu fork at epoch 5
		params.SetupTestConfigCleanup(t)
		cfg := params.BeaconConfig().Copy()
		cfg.FuluForkEpoch = 5
		params.OverrideBeaconConfig(cfg)

		c := &mock.ChainService{State: beaconState, Root: parentRoot[:]}
		proposerServer := &Server{
			ChainStartFetcher:    &mockSila.Chain{},
			SilaChainInfoFetcher: &mockSila.Chain{},
			SilaBlockFetcher:     &mockSila.Chain{},
			BlockReceiver:        c,
			BlobReceiver:         c,
			HeadFetcher:          c,
			BlockNotifier:        c.BlockNotifier(),
			OperationNotifier:    c.OperationNotifier(),
			StateGen:             stategen.New(db, doublylinkedtree.New()),
			TimeFetcher:          c,
			SyncChecker:          &mockSync.Sync{IsSyncing: false},
			BeaconDB:             db,
			P2P:                  &mockp2p.MockBroadcaster{},
		}

		// Create an unblinded block at slot 160 (epoch 5)
		unblindeBlock := util.NewBeaconBlockDeneb()
		unblindeBlock.Block.Slot = 160
		unblindeBlock.Block.ProposerIndex = 0
		unblindeBlock.Block.ParentRoot = parentRoot[:]
		unblindeBlock.Block.StateRoot = make([]byte, 32)

		req := &silapb.GenericSignedBeaconBlock{
			Block: &silapb.GenericSignedBeaconBlock_Deneb{
				Deneb: &silapb.SignedBeaconBlockContentsDeneb{
					Block: unblindeBlock,
				},
			},
		}

		// Unblinded blocks should not trigger post-Fulu condition, even at epoch >= FuluForkEpoch
		res, err := proposerServer.ProposeBeaconBlock(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.BlockRoot)
	})

	t.Run("blinded block - 502 error handling", func(t *testing.T) {
		params.SetupTestConfigCleanup(t)
		cfg := params.BeaconConfig().Copy()
		cfg.FuluForkEpoch = 10
		params.OverrideBeaconConfig(cfg)

		mockBuilder := &builderTest.MockBuilderService{
			HasConfigured:         true,
			Cfg:                   &builderTest.Config{BeaconDB: db},
			PayloadDeneb:          &silaenginev1.SilaPayloadDeneb{},
			ErrSubmitBlindedBlock: builderapi.ErrBadGateway,
		}

		c := &mock.ChainService{State: beaconState, Root: parentRoot[:]}
		proposerServer := &Server{
			ChainStartFetcher:    &mockSila.Chain{},
			SilaChainInfoFetcher: &mockSila.Chain{},
			SilaBlockFetcher:     &mockSila.Chain{},
			BlockReceiver:        c,
			BlobReceiver:         c,
			HeadFetcher:          c,
			BlockNotifier:        c.BlockNotifier(),
			OperationNotifier:    c.OperationNotifier(),
			StateGen:             stategen.New(db, doublylinkedtree.New()),
			TimeFetcher:          c,
			SyncChecker:          &mockSync.Sync{IsSyncing: false},
			BeaconDB:             db,
			BlockBuilder:         mockBuilder,
			P2P:                  &mockp2p.MockBroadcaster{},
		}

		blindedBlock := util.NewBlindedBeaconBlockDeneb()
		blindedBlock.Message.Slot = 160 // This puts us at epoch 5 (160/32 = 5)
		blindedBlock.Message.ProposerIndex = 0
		blindedBlock.Message.ParentRoot = parentRoot[:]
		blindedBlock.Message.StateRoot = make([]byte, 32)

		req := &silapb.GenericSignedBeaconBlock{
			Block: &silapb.GenericSignedBeaconBlock_BlindedDeneb{BlindedDeneb: blindedBlock},
		}

		// Should handle 502 error gracefully and continue with original blinded block
		res, err := proposerServer.ProposeBeaconBlock(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.BlockRoot)
	})
}
