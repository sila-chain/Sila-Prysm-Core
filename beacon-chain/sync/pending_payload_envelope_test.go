package sync

import (
	"bytes"
	"context"
	"testing"
	"time"

	mock "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain/testing"
	dbtest "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db/testing"
	doublylinkedtree "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/forkchoice/doubly-linked-tree"
	p2ptesting "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/startup"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stategen"
	lruwrpr "github.com/sila-chain/Sila-Consensus-Core/v7/cache/lru"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/pkg/errors"
)

func TestProcessPendingPayloadEnvelope_NoPendingEnvelope(t *testing.T) {
	s := &Service{
		pendingPayloadEnvelopes:  make(map[[32]byte]map[uint64]*silapb.SignedSilaPayloadEnvelope),
		seenPayloadEnvelopeCache: lruwrpr.New(10),
		badBlockCache:            lruwrpr.New(10),
		cfg:                      &config{chain: &mock.ChainService{}},
	}
	root := [32]byte{0x01}
	s.processPendingPayloadEnvelope(context.Background(), root)
}

func TestProcessPendingPayloadEnvelope_AlreadySeen(t *testing.T) {
	ctx := context.Background()
	db := dbtest.SetupDB(t)
	chainService := &mock.ChainService{
		Genesis:             time.Unix(time.Now().Unix()-int64(params.BeaconConfig().SecondsPerSlot), 0),
		FinalizedCheckPoint: &silapb.Checkpoint{},
		DB:                  db,
	}
	s := &Service{
		pendingPayloadEnvelopes:  make(map[[32]byte]map[uint64]*silapb.SignedSilaPayloadEnvelope),
		seenPayloadEnvelopeCache: lruwrpr.New(10),
		badBlockCache:            lruwrpr.New(10),
		cfg:                      &config{chain: chainService, beaconDB: db},
	}

	bid := util.GenerateTestSignedSilaPayloadBid(1)
	sb := util.NewBeaconBlockGloas()
	sb.Block.Slot = 1
	sb.Block.Body.SignedSilaPayloadBid = bid
	signedBlock, err := blocks.NewSignedBeaconBlock(sb)
	require.NoError(t, err)
	root, err := signedBlock.Block().HashTreeRoot()
	require.NoError(t, err)

	builderIdx := primitives.BuilderIndex(bid.Message.BuilderIndex)
	blockHash := bytesutil.ToBytes32(bid.Message.BlockHash)
	env := testSignedSilaPayloadEnvelope(t, 1, builderIdx, root, blockHash)
	s.pendingPayloadEnvelopes[root] = map[uint64]*silapb.SignedSilaPayloadEnvelope{uint64(builderIdx): env}

	s.setSeenPayloadEnvelope(root, builderIdx)
	s.processPendingPayloadEnvelope(ctx, root)
	require.Equal(t, 0, len(s.pendingPayloadEnvelopes))
}

func TestProcessPendingPayloadEnvelope_HappyPath(t *testing.T) {
	ctx := context.Background()
	db := dbtest.SetupDB(t)
	chainService := &mock.ChainService{
		Genesis:             time.Unix(time.Now().Unix()-int64(params.BeaconConfig().SecondsPerSlot), 0),
		FinalizedCheckPoint: &silapb.Checkpoint{},
		DB:                  db,
	}
	stateGen := stategen.New(db, doublylinkedtree.New())
	broadcaster := p2ptesting.NewTestP2P(t)
	s := &Service{
		pendingPayloadEnvelopes:  make(map[[32]byte]map[uint64]*silapb.SignedSilaPayloadEnvelope),
		seenPayloadEnvelopeCache: lruwrpr.New(10),
		badBlockCache:            lruwrpr.New(10),
		cfg: &config{
			chain:    chainService,
			beaconDB: db,
			stateGen: stateGen,
			clock:    startup.NewClock(chainService.Genesis, chainService.ValidatorsRoot),
			p2p:      broadcaster,
		},
	}

	bid := util.GenerateTestSignedSilaPayloadBid(1)
	sb := util.NewBeaconBlockGloas()
	sb.Block.Slot = 1
	sb.Block.Body.SignedSilaPayloadBid = bid
	signedBlock, err := blocks.NewSignedBeaconBlock(sb)
	require.NoError(t, err)
	root, err := signedBlock.Block().HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, db.SaveBlock(ctx, signedBlock))

	st, err := util.NewBeaconStateFulu()
	require.NoError(t, err)
	require.NoError(t, db.SaveState(ctx, st, root))

	builderIdx := primitives.BuilderIndex(bid.Message.BuilderIndex)
	blockHash := bytesutil.ToBytes32(bid.Message.BlockHash)
	env := testSignedSilaPayloadEnvelope(t, 1, builderIdx, root, blockHash)
	s.pendingPayloadEnvelopes[root] = map[uint64]*silapb.SignedSilaPayloadEnvelope{uint64(builderIdx): env}

	require.Equal(t, false, s.hasSeenPayloadEnvelope(root, builderIdx))
	s.processPendingPayloadEnvelope(ctx, root)
	require.Equal(t, 0, len(s.pendingPayloadEnvelopes))
	require.Equal(t, true, s.hasSeenPayloadEnvelope(root, builderIdx))
	require.Equal(t, true, broadcaster.BroadcastCalled.Load())
}

func TestProcessPendingPayloadEnvelope_DoesNotBroadcastOnReceiveError(t *testing.T) {
	ctx := context.Background()
	db := dbtest.SetupDB(t)
	chainService := &mock.ChainService{
		Genesis:                   time.Unix(time.Now().Unix()-int64(params.BeaconConfig().SecondsPerSlot), 0),
		FinalizedCheckPoint:       &silapb.Checkpoint{},
		DB:                        db,
		ReceivePayloadEnvelopeErr: errors.New("receive failed"),
	}
	stateGen := stategen.New(db, doublylinkedtree.New())
	broadcaster := p2ptesting.NewTestP2P(t)
	s := &Service{
		pendingPayloadEnvelopes:  make(map[[32]byte]map[uint64]*silapb.SignedSilaPayloadEnvelope),
		seenPayloadEnvelopeCache: lruwrpr.New(10),
		badBlockCache:            lruwrpr.New(10),
		cfg: &config{
			chain:    chainService,
			beaconDB: db,
			stateGen: stateGen,
			clock:    startup.NewClock(chainService.Genesis, chainService.ValidatorsRoot),
			p2p:      broadcaster,
		},
	}

	bid := util.GenerateTestSignedSilaPayloadBid(1)
	sb := util.NewBeaconBlockGloas()
	sb.Block.Slot = 1
	sb.Block.Body.SignedSilaPayloadBid = bid
	signedBlock, err := blocks.NewSignedBeaconBlock(sb)
	require.NoError(t, err)
	root, err := signedBlock.Block().HashTreeRoot()
	require.NoError(t, err)

	builderIdx := primitives.BuilderIndex(bid.Message.BuilderIndex)
	blockHash := bytesutil.ToBytes32(bid.Message.BlockHash)
	env := testSignedSilaPayloadEnvelope(t, 1, builderIdx, root, blockHash)
	s.pendingPayloadEnvelopes[root] = map[uint64]*silapb.SignedSilaPayloadEnvelope{uint64(builderIdx): env}

	s.processPendingPayloadEnvelope(ctx, root)
	require.Equal(t, false, broadcaster.BroadcastCalled.Load())
}

func TestProcessPendingPayloadEnvelopes_Sweep(t *testing.T) {
	ctx := context.Background()
	db := dbtest.SetupDB(t)
	chainService := &mock.ChainService{
		Genesis:             time.Unix(time.Now().Unix()-int64(params.BeaconConfig().SecondsPerSlot), 0),
		FinalizedCheckPoint: &silapb.Checkpoint{},
		DB:                  db,
	}
	stateGen := stategen.New(db, doublylinkedtree.New())
	s := &Service{
		pendingPayloadEnvelopes:  make(map[[32]byte]map[uint64]*silapb.SignedSilaPayloadEnvelope),
		seenPayloadEnvelopeCache: lruwrpr.New(10),
		badBlockCache:            lruwrpr.New(10),
		cfg: &config{
			chain:    chainService,
			beaconDB: db,
			stateGen: stateGen,
			clock:    startup.NewClock(chainService.Genesis, chainService.ValidatorsRoot),
			p2p:      p2ptesting.NewTestP2P(t),
		},
	}

	bid := util.GenerateTestSignedSilaPayloadBid(1)
	sb := util.NewBeaconBlockGloas()
	sb.Block.Slot = 1
	sb.Block.Body.SignedSilaPayloadBid = bid
	signedBlock, err := blocks.NewSignedBeaconBlock(sb)
	require.NoError(t, err)
	root, err := signedBlock.Block().HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, db.SaveBlock(ctx, signedBlock))

	st, err := util.NewBeaconStateFulu()
	require.NoError(t, err)
	require.NoError(t, db.SaveState(ctx, st, root))

	builderIdx := primitives.BuilderIndex(bid.Message.BuilderIndex)
	blockHash := bytesutil.ToBytes32(bid.Message.BlockHash)
	env := testSignedSilaPayloadEnvelope(t, 1, builderIdx, root, blockHash)
	s.pendingPayloadEnvelopes[root] = map[uint64]*silapb.SignedSilaPayloadEnvelope{uint64(builderIdx): env}
	s.newSilaPayloadEnvelopeVerifier = testNewSilaPayloadEnvelopeVerifier(mockSilaPayloadEnvelopeVerifier{})

	require.Equal(t, false, s.hasSeenPayloadEnvelope(root, builderIdx))

	s.processPendingPayloadEnvelopes(ctx)
	require.Equal(t, 0, len(s.pendingPayloadEnvelopes))
	require.Equal(t, true, s.hasSeenPayloadEnvelope(root, builderIdx))
}

func TestProcessPendingPayloadEnvelopes_SkipsUnknownRoot(t *testing.T) {
	ctx := context.Background()
	db := dbtest.SetupDB(t)
	chainService := &mock.ChainService{
		Genesis:             time.Unix(time.Now().Unix()-int64(params.BeaconConfig().SecondsPerSlot), 0),
		FinalizedCheckPoint: &silapb.Checkpoint{},
		DB:                  db,
		NotFinalized:        true, // InForkchoice returns false
	}
	s := &Service{
		pendingPayloadEnvelopes:  make(map[[32]byte]map[uint64]*silapb.SignedSilaPayloadEnvelope),
		seenPayloadEnvelopeCache: lruwrpr.New(10),
		badBlockCache:            lruwrpr.New(10),
		cfg:                      &config{chain: chainService, beaconDB: db},
	}

	root := [32]byte{0x01}
	blockHash := [32]byte{0x02}
	env := testSignedSilaPayloadEnvelope(t, 1, 1, root, blockHash)
	s.pendingPayloadEnvelopes[root] = map[uint64]*silapb.SignedSilaPayloadEnvelope{1: env}

	s.processPendingPayloadEnvelopes(ctx)
	require.Equal(t, 1, len(s.pendingPayloadEnvelopes))
}

func TestPrunePendingPayloadEnvelopes(t *testing.T) {
	finalizedEpoch := primitives.Epoch(3)
	slotsPerEpoch := params.BeaconConfig().SlotsPerEpoch
	s := &Service{
		pendingPayloadEnvelopes: make(map[[32]byte]map[uint64]*silapb.SignedSilaPayloadEnvelope),
		cfg: &config{
			chain: &mock.ChainService{
				FinalizedCheckPoint: &silapb.Checkpoint{Epoch: finalizedEpoch},
			},
		},
	}

	oldRoot := [32]byte{0x01}
	oldEnv := &silapb.SignedSilaPayloadEnvelope{
		Message: &silapb.SilaPayloadEnvelope{
			Payload:         &silaenginev1.SilaPayloadGloas{SlotNumber: primitives.Slot(finalizedEpoch-1) * slotsPerEpoch},
			BeaconBlockRoot: oldRoot[:],
		},
		Signature: bytes.Repeat([]byte{0xAA}, 96),
	}

	atFinalizedRoot := [32]byte{0x03}
	atFinalizedEnv := &silapb.SignedSilaPayloadEnvelope{
		Message: &silapb.SilaPayloadEnvelope{
			Payload:         &silaenginev1.SilaPayloadGloas{SlotNumber: primitives.Slot(finalizedEpoch) * slotsPerEpoch},
			BeaconBlockRoot: atFinalizedRoot[:],
		},
		Signature: bytes.Repeat([]byte{0xCC}, 96),
	}

	freshRoot := [32]byte{0x02}
	freshEnv := &silapb.SignedSilaPayloadEnvelope{
		Message: &silapb.SilaPayloadEnvelope{
			Payload:         &silaenginev1.SilaPayloadGloas{SlotNumber: primitives.Slot(finalizedEpoch+1) * slotsPerEpoch},
			BeaconBlockRoot: freshRoot[:],
		},
		Signature: bytes.Repeat([]byte{0xBB}, 96),
	}

	s.pendingPayloadEnvelopes[oldRoot] = map[uint64]*silapb.SignedSilaPayloadEnvelope{1: oldEnv}
	s.pendingPayloadEnvelopes[atFinalizedRoot] = map[uint64]*silapb.SignedSilaPayloadEnvelope{1: atFinalizedEnv}
	s.pendingPayloadEnvelopes[freshRoot] = map[uint64]*silapb.SignedSilaPayloadEnvelope{1: freshEnv}
	require.Equal(t, 3, len(s.pendingPayloadEnvelopes))

	s.prunePendingPayloadEnvelopes()

	require.Equal(t, 2, len(s.pendingPayloadEnvelopes))
	_, ok := s.pendingPayloadEnvelopes[oldRoot]
	require.Equal(t, false, ok)
	_, ok = s.pendingPayloadEnvelopes[atFinalizedRoot]
	require.Equal(t, true, ok)
	_, ok = s.pendingPayloadEnvelopes[freshRoot]
	require.Equal(t, true, ok)
}

func TestQueuePendingPayloadEnvelope_SelfBuildIgnoredOutsideLookahead(t *testing.T) {
	ctx := context.Background()
	cfg := params.BeaconConfig()
	selfBuild := cfg.BuilderIndexSelfBuild
	// Place the envelope in epoch 2 so the head state (epoch 0) is outside
	// the proposer lookahead window.
	envelopeSlot := primitives.Slot(2 * cfg.SlotsPerEpoch)

	db := dbtest.SetupDB(t)
	chainService := &mock.ChainService{
		Genesis:             time.Unix(time.Now().Unix()-int64(uint64(envelopeSlot)*cfg.SecondsPerSlot), 0),
		FinalizedCheckPoint: &silapb.Checkpoint{},
		DB:                  db,
	}
	st, err := util.NewBeaconStateFulu()
	require.NoError(t, err)
	chainService.State = st

	s := &Service{
		seenPayloadEnvelopeCache: lruwrpr.New(10),
		pendingPayloadEnvelopes:  make(map[[32]byte]map[uint64]*silapb.SignedSilaPayloadEnvelope),
		cfg: &config{
			chain: chainService,
			clock: startup.NewClock(chainService.Genesis, chainService.ValidatorsRoot),
		},
	}

	root := [32]byte{0x01}
	blockHash := [32]byte{0x02}
	signedEnv := testSignedSilaPayloadEnvelope(t, envelopeSlot, selfBuild, root, blockHash)
	e, err := blocks.WrappedROSignedSilaPayloadEnvelope(signedEnv)
	require.NoError(t, err)
	env, err := e.Envelope()
	require.NoError(t, err)

	// Signature verification would fail, but self-build outside the lookahead
	// should skip it and return Ignore without queuing.
	v := &mockSilaPayloadEnvelopeVerifier{errSignature: errors.New("bad signature")}
	result, err := s.queuePendingPayloadEnvelope(ctx, v, env, signedEnv)
	require.NoError(t, err)
	require.Equal(t, pubsub.ValidationIgnore, result)
	require.Equal(t, 0, len(s.pendingPayloadEnvelopes))
}

func TestQueuePendingPayloadEnvelope_SelfBuildInLookaheadVerifiesSignature(t *testing.T) {
	ctx := context.Background()
	s, _, _, root := setupSilaPayloadEnvelopeService(t, 1, 1)
	selfBuild := params.BeaconConfig().BuilderIndexSelfBuild

	blockHash := [32]byte{0x02}
	signedEnv := testSignedSilaPayloadEnvelope(t, 1, selfBuild, root, blockHash)
	e, err := blocks.WrappedROSignedSilaPayloadEnvelope(signedEnv)
	require.NoError(t, err)
	env, err := e.Envelope()
	require.NoError(t, err)

	// Self-build in the same epoch (lookahead) verifies the signature but ignores failures.
	v := &mockSilaPayloadEnvelopeVerifier{errSignature: errors.New("bad signature")}
	result, err := s.queuePendingPayloadEnvelope(ctx, v, env, signedEnv)
	require.NoError(t, err)
	require.Equal(t, pubsub.ValidationIgnore, result)
	require.Equal(t, 1, s.selfBuildSigFailures)

	// After maxSelfBuildSigFailures, skip the signature check entirely and queue the envelope.
	s.selfBuildSigFailures = maxSelfBuildSigFailures
	result, err = s.queuePendingPayloadEnvelope(ctx, v, env, signedEnv)
	require.NoError(t, err)
	require.Equal(t, pubsub.ValidationIgnore, result)
	require.Equal(t, maxSelfBuildSigFailures, s.selfBuildSigFailures)
}

func TestQueuePendingPayloadEnvelope_RejectBadSignature(t *testing.T) {
	ctx := context.Background()
	s, _, _, root := setupSilaPayloadEnvelopeService(t, 1, 1)

	blockHash := [32]byte{0x02}
	signedEnv := testSignedSilaPayloadEnvelope(t, 1, 1, root, blockHash)
	e, err := blocks.WrappedROSignedSilaPayloadEnvelope(signedEnv)
	require.NoError(t, err)
	env, err := e.Envelope()
	require.NoError(t, err)

	v := &mockSilaPayloadEnvelopeVerifier{errSignature: errors.New("bad signature")}
	result, err := s.queuePendingPayloadEnvelope(ctx, v, env, signedEnv)
	require.NotNil(t, err)
	require.Equal(t, pubsub.ValidationReject, result)
	require.Equal(t, 0, len(s.pendingPayloadEnvelopes))
}

func TestQueuePendingPayloadEnvelope_QueuesNewRoot(t *testing.T) {
	ctx := context.Background()
	s, _, _, root := setupSilaPayloadEnvelopeService(t, 1, 1)

	blockHash := [32]byte{0x02}
	signedEnv := testSignedSilaPayloadEnvelope(t, 1, 1, root, blockHash)
	e, err := blocks.WrappedROSignedSilaPayloadEnvelope(signedEnv)
	require.NoError(t, err)
	env, err := e.Envelope()
	require.NoError(t, err)

	v := &mockSilaPayloadEnvelopeVerifier{}
	result, err := s.queuePendingPayloadEnvelope(ctx, v, env, signedEnv)
	require.NoError(t, err)
	require.Equal(t, pubsub.ValidationIgnore, result)
	require.Equal(t, 1, len(s.pendingPayloadEnvelopes))
	_, ok := s.pendingPayloadEnvelopes[root]
	require.Equal(t, true, ok)
}

func TestQueuePendingPayloadEnvelope_DoesNotOverwrite(t *testing.T) {
	ctx := context.Background()
	s, _, _, root := setupSilaPayloadEnvelopeService(t, 1, 1)

	blockHash := [32]byte{0x02}
	first := testSignedSilaPayloadEnvelope(t, 1, 1, root, blockHash)
	s.pendingPayloadEnvelopes[root] = map[uint64]*silapb.SignedSilaPayloadEnvelope{1: first}

	second := testSignedSilaPayloadEnvelope(t, 1, 1, root, blockHash)
	e, err := blocks.WrappedROSignedSilaPayloadEnvelope(second)
	require.NoError(t, err)
	env, err := e.Envelope()
	require.NoError(t, err)

	v := &mockSilaPayloadEnvelopeVerifier{}
	result, err := s.queuePendingPayloadEnvelope(ctx, v, env, second)
	require.NoError(t, err)
	require.Equal(t, pubsub.ValidationIgnore, result)
	require.Equal(t, 1, len(s.pendingPayloadEnvelopes[root]))
	require.Equal(t, first, s.pendingPayloadEnvelopes[root][1])
}

func TestQueuePendingPayloadEnvelope_PrunesMalformedExistingEnvelope(t *testing.T) {
	ctx := context.Background()
	s, _, _, root := setupSilaPayloadEnvelopeService(t, 1, 1)

	s.pendingPayloadEnvelopes[root] = map[uint64]*silapb.SignedSilaPayloadEnvelope{
		1: {Signature: bytes.Repeat([]byte{0xAA}, 96)},
	}

	blockHash := [32]byte{0x02}
	next := testSignedSilaPayloadEnvelope(t, 1, 1, root, blockHash)
	e, err := blocks.WrappedROSignedSilaPayloadEnvelope(next)
	require.NoError(t, err)
	env, err := e.Envelope()
	require.NoError(t, err)

	v := &mockSilaPayloadEnvelopeVerifier{}
	result, err := s.queuePendingPayloadEnvelope(ctx, v, env, next)
	require.NoError(t, err)
	require.Equal(t, pubsub.ValidationIgnore, result)
	require.Equal(t, 1, len(s.pendingPayloadEnvelopes[root]))
	require.Equal(t, next, s.pendingPayloadEnvelopes[root][1])
}

func TestQueuePendingPayloadEnvelope_RootCountBound(t *testing.T) {
	ctx := context.Background()
	s, _, _, _ := setupSilaPayloadEnvelopeService(t, 1, 1)

	// Fill up to maxPendingPayloadRoots with non-self-build envelopes.
	for i := range maxPendingPayloadRoots {
		root := [32]byte{byte(i + 1)}
		env := &silapb.SignedSilaPayloadEnvelope{
			Message: &silapb.SilaPayloadEnvelope{Payload: &silaenginev1.SilaPayloadGloas{SlotNumber: 1}, BeaconBlockRoot: root[:]},
		}
		s.pendingPayloadEnvelopes[root] = map[uint64]*silapb.SignedSilaPayloadEnvelope{uint64(i): env}
	}
	require.Equal(t, maxPendingPayloadRoots, len(s.pendingPayloadEnvelopes))

	// Next non-self-build root should be rejected.
	newRoot := [32]byte{0xFF}
	signedEnv := testSignedSilaPayloadEnvelope(t, 1, 1, newRoot, [32]byte{0x02})
	e, err := blocks.WrappedROSignedSilaPayloadEnvelope(signedEnv)
	require.NoError(t, err)
	env, err := e.Envelope()
	require.NoError(t, err)

	v := &mockSilaPayloadEnvelopeVerifier{}
	result, err := s.queuePendingPayloadEnvelope(ctx, v, env, signedEnv)
	require.NoError(t, err)
	require.Equal(t, pubsub.ValidationIgnore, result)
	_, ok := s.pendingPayloadEnvelopes[newRoot]
	require.Equal(t, false, ok)
}

func TestQueuePendingPayloadEnvelope_SelfBuildBypassesRootBound(t *testing.T) {
	ctx := context.Background()
	s, _, _, _ := setupSilaPayloadEnvelopeService(t, 1, 1)
	selfBuild := params.BeaconConfig().BuilderIndexSelfBuild

	// Fill to the root limit.
	for i := range maxPendingPayloadRoots {
		root := [32]byte{byte(i + 1)}
		env := &silapb.SignedSilaPayloadEnvelope{
			Message: &silapb.SilaPayloadEnvelope{Payload: &silaenginev1.SilaPayloadGloas{SlotNumber: 1}, BeaconBlockRoot: root[:]},
		}
		s.pendingPayloadEnvelopes[root] = map[uint64]*silapb.SignedSilaPayloadEnvelope{uint64(i): env}
	}

	// Self-build for a new root should still be accepted.
	newRoot := [32]byte{0xFF}
	signedEnv := testSignedSilaPayloadEnvelope(t, 1, selfBuild, newRoot, [32]byte{0x02})
	e, err := blocks.WrappedROSignedSilaPayloadEnvelope(signedEnv)
	require.NoError(t, err)
	env, err := e.Envelope()
	require.NoError(t, err)

	v := &mockSilaPayloadEnvelopeVerifier{}
	result, err := s.queuePendingPayloadEnvelope(ctx, v, env, signedEnv)
	require.NoError(t, err)
	require.Equal(t, pubsub.ValidationIgnore, result)
	_, ok := s.pendingPayloadEnvelopes[newRoot]
	require.Equal(t, true, ok)
}

func TestQueuePendingPayloadEnvelope_PerRootBuilderBound(t *testing.T) {
	ctx := context.Background()
	s, _, _, root := setupSilaPayloadEnvelopeService(t, 1, 1)

	blockHash := [32]byte{0x02}
	// Insert two non-self-build builders for the same root.
	for i := range uint64(maxPendingBuildersPerRoot) {
		env := testSignedSilaPayloadEnvelope(t, 1, primitives.BuilderIndex(i+10), root, blockHash)
		e, err := blocks.WrappedROSignedSilaPayloadEnvelope(env)
		require.NoError(t, err)
		wrapped, err := e.Envelope()
		require.NoError(t, err)
		v := &mockSilaPayloadEnvelopeVerifier{}
		result, err := s.queuePendingPayloadEnvelope(ctx, v, wrapped, env)
		require.NoError(t, err)
		require.Equal(t, pubsub.ValidationIgnore, result)
	}
	require.Equal(t, int(maxPendingBuildersPerRoot), len(s.pendingPayloadEnvelopes[root]))

	// Third non-self-build builder should be rejected.
	third := testSignedSilaPayloadEnvelope(t, 1, 99, root, blockHash)
	e, err := blocks.WrappedROSignedSilaPayloadEnvelope(third)
	require.NoError(t, err)
	env, err := e.Envelope()
	require.NoError(t, err)

	v := &mockSilaPayloadEnvelopeVerifier{}
	result, err := s.queuePendingPayloadEnvelope(ctx, v, env, third)
	require.NoError(t, err)
	require.Equal(t, pubsub.ValidationIgnore, result)
	require.Equal(t, int(maxPendingBuildersPerRoot), len(s.pendingPayloadEnvelopes[root]))
}

func TestQueuePendingPayloadEnvelope_SelfBuildBypassesPerRootBound(t *testing.T) {
	ctx := context.Background()
	s, _, _, root := setupSilaPayloadEnvelopeService(t, 1, 1)
	selfBuild := params.BeaconConfig().BuilderIndexSelfBuild

	blockHash := [32]byte{0x02}
	// Fill with maxPendingBuildersPerRoot non-self-build builders.
	for i := range uint64(maxPendingBuildersPerRoot) {
		env := testSignedSilaPayloadEnvelope(t, 1, primitives.BuilderIndex(i+10), root, blockHash)
		e, err := blocks.WrappedROSignedSilaPayloadEnvelope(env)
		require.NoError(t, err)
		wrapped, err := e.Envelope()
		require.NoError(t, err)
		v := &mockSilaPayloadEnvelopeVerifier{}
		_, _ = s.queuePendingPayloadEnvelope(ctx, v, wrapped, env)
	}

	// Self-build should be accepted as the 3rd builder.
	selfEnv := testSignedSilaPayloadEnvelope(t, 1, selfBuild, root, blockHash)
	e, err := blocks.WrappedROSignedSilaPayloadEnvelope(selfEnv)
	require.NoError(t, err)
	env, err := e.Envelope()
	require.NoError(t, err)

	v := &mockSilaPayloadEnvelopeVerifier{}
	result, err := s.queuePendingPayloadEnvelope(ctx, v, env, selfEnv)
	require.NoError(t, err)
	require.Equal(t, pubsub.ValidationIgnore, result)
	require.Equal(t, int(maxPendingBuildersPerRoot)+1, len(s.pendingPayloadEnvelopes[root]))
	_, ok := s.pendingPayloadEnvelopes[root][uint64(selfBuild)]
	require.Equal(t, true, ok)
}

func TestValidateSilaPayloadEnvelope_RejectBadSignatureBeforeQueue(t *testing.T) {
	ctx := context.Background()
	s, msg, _, _ := setupSilaPayloadEnvelopeService(t, 1, 1)
	s.newSilaPayloadEnvelopeVerifier = testNewSilaPayloadEnvelopeVerifier(
		mockSilaPayloadEnvelopeVerifier{
			errBlockRootSeen: errors.New("not seen"),
			errSignature:     errors.New("bad signature"),
		},
	)

	result, err := s.validateSilaPayloadEnvelope(ctx, "", msg)
	require.NotNil(t, err)
	require.Equal(t, result, pubsub.ValidationReject)
	require.Equal(t, 0, len(s.pendingPayloadEnvelopes))
}

func TestValidateSilaPayloadEnvelope_QueueOnUnknownBlock(t *testing.T) {
	ctx := context.Background()
	s, msg, _, root := setupSilaPayloadEnvelopeService(t, 1, 1)
	s.newSilaPayloadEnvelopeVerifier = testNewSilaPayloadEnvelopeVerifier(
		mockSilaPayloadEnvelopeVerifier{errBlockRootSeen: errors.New("not seen")},
	)

	require.Equal(t, 0, len(s.pendingPayloadEnvelopes))
	result, err := s.validateSilaPayloadEnvelope(ctx, "", msg)
	require.NoError(t, err)
	require.Equal(t, result, pubsub.ValidationIgnore)
	require.Equal(t, 1, len(s.pendingPayloadEnvelopes))
	_, ok := s.pendingPayloadEnvelopes[root]
	require.Equal(t, true, ok)
}

func TestValidateSilaPayloadEnvelope_QueueKeepsFirst(t *testing.T) {
	ctx := context.Background()
	s, msg, _, root := setupSilaPayloadEnvelopeService(t, 1, 1)
	s.newSilaPayloadEnvelopeVerifier = testNewSilaPayloadEnvelopeVerifier(
		mockSilaPayloadEnvelopeVerifier{errBlockRootSeen: errors.New("not seen")},
	)

	// First envelope gets queued.
	_, _ = s.validateSilaPayloadEnvelope(ctx, "", msg)
	require.Equal(t, 1, len(s.pendingPayloadEnvelopes[root]))

	// Second envelope for the same root and same builder should be ignored (keep first).
	_, _ = s.validateSilaPayloadEnvelope(ctx, "", msg)
	require.Equal(t, 1, len(s.pendingPayloadEnvelopes))
	require.Equal(t, 1, len(s.pendingPayloadEnvelopes[root]))
}
