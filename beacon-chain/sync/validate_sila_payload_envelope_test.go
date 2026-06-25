package sync

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/sila-chain/Sila-Consensus-Core/v7/async/abool"
	mock "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed"
	opfeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/operation"
	dbtest "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db/testing"
	doublylinkedtree "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/forkchoice/doubly-linked-tree"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p"
	p2ptest "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/startup"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stategen"
	mockSync "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/sync/initial-sync/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/verification"
	lruwrpr "github.com/sila-chain/Sila-Consensus-Core/v7/cache/lru"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/pkg/errors"
)

func TestValidateSilaPayloadEnvelope_InvalidTopic(t *testing.T) {
	ctx := context.Background()
	p := p2ptest.NewTestP2P(t)
	s := &Service{cfg: &config{p2p: p, initialSync: &mockSync.Sync{}}}

	result, err := s.validateSilaPayloadEnvelope(ctx, "", &pubsub.Message{
		Message: &pb.Message{},
	})
	require.ErrorIs(t, p2p.ErrInvalidTopic, err)
	require.Equal(t, result, pubsub.ValidationReject)
}

func TestValidateSilaPayloadEnvelope_AlreadySeen(t *testing.T) {
	ctx := context.Background()
	s, msg, builderIdx, root := setupSilaPayloadEnvelopeService(t, 1, 1)
	s.newSilaPayloadEnvelopeVerifier = testNewSilaPayloadEnvelopeVerifier(mockSilaPayloadEnvelopeVerifier{})

	s.setSeenPayloadEnvelope(root, builderIdx)
	result, err := s.validateSilaPayloadEnvelope(ctx, "", msg)
	require.NoError(t, err)
	require.Equal(t, result, pubsub.ValidationIgnore)
}

func TestValidateSilaPayloadEnvelope_ErrorPathsWithMock(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name      string
		verifier  mockSilaPayloadEnvelopeVerifier
		result    pubsub.ValidationResult
		wantError bool
	}{
		{
			name:     "block root not seen queues envelope",
			verifier: mockSilaPayloadEnvelopeVerifier{errBlockRootSeen: errors.New("not seen")},
			result:   pubsub.ValidationIgnore,
		},
		{
			name:      "slot below finalized",
			verifier:  mockSilaPayloadEnvelopeVerifier{errSlotAboveFinalized: errors.New("below finalized")},
			result:    pubsub.ValidationIgnore,
			wantError: true,
		},
		{
			name:      "block root invalid",
			verifier:  mockSilaPayloadEnvelopeVerifier{errBlockRootValid: errors.New("invalid block")},
			result:    pubsub.ValidationReject,
			wantError: true,
		},
		{
			name:      "slot mismatch",
			verifier:  mockSilaPayloadEnvelopeVerifier{errSlotMatchesBlock: errors.New("slot mismatch")},
			result:    pubsub.ValidationReject,
			wantError: true,
		},
		{
			name:      "builder mismatch",
			verifier:  mockSilaPayloadEnvelopeVerifier{errBuilderValid: errors.New("builder mismatch")},
			result:    pubsub.ValidationReject,
			wantError: true,
		},
		{
			name:      "payload hash mismatch",
			verifier:  mockSilaPayloadEnvelopeVerifier{errPayloadHash: errors.New("payload hash mismatch")},
			result:    pubsub.ValidationReject,
			wantError: true,
		},
		{
			name:      "signature invalid",
			verifier:  mockSilaPayloadEnvelopeVerifier{errSignature: errors.New("signature invalid")},
			result:    pubsub.ValidationReject,
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, msg, _, _ := setupSilaPayloadEnvelopeService(t, 1, 1)
			s.newSilaPayloadEnvelopeVerifier = testNewSilaPayloadEnvelopeVerifier(tc.verifier)

			result, err := s.validateSilaPayloadEnvelope(ctx, "", msg)
			if tc.wantError {
				require.NotNil(t, err)
			}
			require.Equal(t, result, tc.result)
		})
	}
}

func TestValidateSilaPayloadEnvelope_HappyPath(t *testing.T) {
	ctx := context.Background()
	s, msg, builderIdx, root := setupSilaPayloadEnvelopeService(t, 1, 1)
	s.newSilaPayloadEnvelopeVerifier = testNewSilaPayloadEnvelopeVerifier(mockSilaPayloadEnvelopeVerifier{})

	require.Equal(t, false, s.hasSeenPayloadEnvelope(root, builderIdx))
	result, err := s.validateSilaPayloadEnvelope(ctx, "", msg)
	require.NoError(t, err)
	require.Equal(t, result, pubsub.ValidationAccept)
	require.Equal(t, true, s.hasSeenPayloadEnvelope(root, builderIdx))
}

func TestValidateSilaPayloadEnvelope_GossipEvent(t *testing.T) {
	ctx := context.Background()
	s, msg, builderIdx, root := setupSilaPayloadEnvelopeService(t, 1, 1)
	s.newSilaPayloadEnvelopeVerifier = testNewSilaPayloadEnvelopeVerifier(mockSilaPayloadEnvelopeVerifier{})

	opChannel := make(chan *feed.Event, 1)
	opSub := s.cfg.operationNotifier.OperationFeed().Subscribe(opChannel)
	defer opSub.Unsubscribe()

	t.Run("gossip event emitted on valid envelope", func(t *testing.T) {
		result, err := s.validateSilaPayloadEnvelope(ctx, "", msg)
		require.NoError(t, err)
		require.Equal(t, pubsub.ValidationAccept, result)

		signed, ok := msg.ValidatorData.(*silapb.SignedSilaPayloadEnvelope)
		require.Equal(t, true, ok)

		select {
		case event := <-opChannel:
			require.Equal(t, feed.EventType(opfeed.SilaPayloadGossipReceived), event.Type)
			data, ok := event.Data.(*opfeed.SilaPayloadGossipReceivedData)
			require.Equal(t, true, ok)
			require.Equal(t, primitives.Slot(1), data.Slot)
			require.Equal(t, builderIdx, data.BuilderIndex)
			require.Equal(t, root, data.BlockRoot)
			require.Equal(t, bytesutil.ToBytes32(signed.Message.Payload.BlockHash), data.BlockHash)
		case <-time.After(time.Second):
			t.Fatal("expected sila_payload_gossip event was not received")
		}
	})

	t.Run("self-origin envelope does not emit gossip event", func(t *testing.T) {
		result, err := s.validateSilaPayloadEnvelope(ctx, s.cfg.p2p.PeerID(), msg)
		require.NoError(t, err)
		require.Equal(t, pubsub.ValidationAccept, result)

		select {
		case event := <-opChannel:
			t.Fatalf("did not expect a gossip event for a self-origin envelope, got type %v", event.Type)
		case <-time.After(50 * time.Millisecond):
			// No event received, as expected.
		}
	})
}

func TestSilaPayloadEnvelopeSubscriber_WrongMessage(t *testing.T) {
	s := &Service{cfg: &config{}}
	err := s.silaPayloadEnvelopeSubscriber(context.Background(), &silapb.BeaconBlock{})
	require.ErrorIs(t, errWrongMessage, err)
}

func TestSilaPayloadEnvelopeSubscriber_HappyPath(t *testing.T) {
	s := &Service{
		cfg:          &config{chain: &mock.ChainService{}},
		chainStarted: abool.New(),
	}
	root := [32]byte{0x01}
	blockHash := [32]byte{0x02}
	env := testSignedSilaPayloadEnvelope(t, 1, 2, root, blockHash)

	err := s.silaPayloadEnvelopeSubscriber(context.Background(), env)
	require.NoError(t, err)
}

type mockSilaPayloadEnvelopeVerifier struct {
	errBlockRootSeen      error
	errBlockRootValid     error
	errSlotAboveFinalized error
	errSlotMatchesBlock   error
	errBuilderValid       error
	errPayloadHash        error
	errSignature          error
}

var _ verification.SilaPayloadEnvelopeVerifier = &mockSilaPayloadEnvelopeVerifier{}

func (m *mockSilaPayloadEnvelopeVerifier) VerifyBlockRootSeen(_ func([32]byte) bool) error {
	return m.errBlockRootSeen
}

func (m *mockSilaPayloadEnvelopeVerifier) VerifyBlockRootValid(_ func([32]byte) bool) error {
	return m.errBlockRootValid
}

func (m *mockSilaPayloadEnvelopeVerifier) VerifySlotAboveFinalized(_ primitives.Epoch) error {
	return m.errSlotAboveFinalized
}

func (m *mockSilaPayloadEnvelopeVerifier) VerifySlotMatchesBlock(_ primitives.Slot) error {
	return m.errSlotMatchesBlock
}

func (m *mockSilaPayloadEnvelopeVerifier) VerifyBuilderValid(_ interfaces.ROSilaPayloadBid) error {
	return m.errBuilderValid
}

func (m *mockSilaPayloadEnvelopeVerifier) VerifyPayloadHash(_ interfaces.ROSilaPayloadBid) error {
	return m.errPayloadHash
}

func (m *mockSilaPayloadEnvelopeVerifier) VerifyExecutionRequestsRoot(_ interfaces.ROSilaPayloadBid) error {
	return nil
}

func (m *mockSilaPayloadEnvelopeVerifier) VerifySignature(_ context.Context, _ state.ReadOnlyBeaconState) error {
	return m.errSignature
}

func (*mockSilaPayloadEnvelopeVerifier) SatisfyRequirement(_ verification.Requirement) {}

func testNewSilaPayloadEnvelopeVerifier(m mockSilaPayloadEnvelopeVerifier) verification.NewSilaPayloadEnvelopeVerifier {
	return func(_ interfaces.ROSignedSilaPayloadEnvelope, _ []verification.Requirement) verification.SilaPayloadEnvelopeVerifier {
		clone := m
		return &clone
	}
}

func setupSilaPayloadEnvelopeService(t *testing.T, envelopeSlot, blockSlot primitives.Slot) (*Service, *pubsub.Message, primitives.BuilderIndex, [32]byte) {
	t.Helper()

	ctx := context.Background()
	db := dbtest.SetupDB(t)
	p := p2ptest.NewTestP2P(t)
	chainService := &mock.ChainService{
		Genesis:             time.Unix(time.Now().Unix()-int64(params.BeaconConfig().SecondsPerSlot), 0),
		FinalizedCheckPoint: &silapb.Checkpoint{},
		DB:                  db,
	}
	stateGen := stategen.New(db, doublylinkedtree.New())
	s := &Service{
		seenPayloadEnvelopeCache: lruwrpr.New(10),
		pendingPayloadEnvelopes:  make(map[[32]byte]map[uint64]*silapb.SignedSilaPayloadEnvelope),
		cfg: &config{
			p2p:               p,
			initialSync:       &mockSync.Sync{},
			chain:             chainService,
			beaconDB:          db,
			stateGen:          stateGen,
			clock:             startup.NewClock(chainService.Genesis, chainService.ValidatorsRoot),
			operationNotifier: chainService.OperationNotifier(),
		},
	}

	bid := util.GenerateTestSignedSilaPayloadBid(blockSlot)
	sb := util.NewBeaconBlockGloas()
	sb.Block.Slot = blockSlot
	sb.Block.Body.SignedSilaPayloadBid = bid
	signedBlock, err := blocks.NewSignedBeaconBlock(sb)
	require.NoError(t, err)
	root, err := signedBlock.Block().HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, db.SaveBlock(ctx, signedBlock))

	state, err := util.NewBeaconStateFulu()
	require.NoError(t, err)
	require.NoError(t, db.SaveState(ctx, state, root))
	chainService.State = state

	blockHash := bytesutil.ToBytes32(bid.Message.BlockHash)
	env := testSignedSilaPayloadEnvelope(t, envelopeSlot, primitives.BuilderIndex(bid.Message.BuilderIndex), root, blockHash)
	msg := envelopeToPubsub(t, s, p, env)

	return s, msg, primitives.BuilderIndex(bid.Message.BuilderIndex), root
}

func envelopeToPubsub(t *testing.T, s *Service, p p2p.P2P, env *silapb.SignedSilaPayloadEnvelope) *pubsub.Message {
	t.Helper()

	buf := new(bytes.Buffer)
	_, err := p.Encoding().EncodeGossip(buf, env)
	require.NoError(t, err)

	topic := p2p.GossipTypeMapping[reflect.TypeFor[*silapb.SignedSilaPayloadEnvelope]()]
	digest, err := s.currentForkDigest()
	require.NoError(t, err)
	topic = s.addDigestToTopic(topic, digest)

	return &pubsub.Message{
		Message: &pb.Message{
			Data:  buf.Bytes(),
			Topic: &topic,
		},
	}
}

func TestQueuePendingPayloadEnvelope_SelfBuildInvalidSignature(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name       string
		builderIdx primitives.BuilderIndex
		result     pubsub.ValidationResult
		wantError  bool
	}{
		{
			name:       "self-build with invalid signature is ignored",
			builderIdx: params.BeaconConfig().BuilderIndexSelfBuild,
			result:     pubsub.ValidationIgnore,
		},
		{
			name:       "non-self-build with invalid signature is rejected",
			builderIdx: 42,
			result:     pubsub.ValidationReject,
			wantError:  true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := p2ptest.NewTestP2P(t)
			chainService := &mock.ChainService{
				Genesis:             time.Unix(time.Now().Unix()-int64(params.BeaconConfig().SecondsPerSlot), 0),
				FinalizedCheckPoint: &silapb.Checkpoint{},
			}
			st, err := util.NewBeaconStateFulu()
			require.NoError(t, err)
			chainService.State = st

			s := &Service{
				seenPayloadEnvelopeCache: lruwrpr.New(10),
				pendingPayloadEnvelopes:  make(map[[32]byte]map[uint64]*silapb.SignedSilaPayloadEnvelope),
				cfg: &config{
					p2p:         p,
					initialSync: &mockSync.Sync{},
					chain:       chainService,
					clock:       startup.NewClock(chainService.Genesis, chainService.ValidatorsRoot),
				},
			}
			s.newSilaPayloadEnvelopeVerifier = testNewSilaPayloadEnvelopeVerifier(mockSilaPayloadEnvelopeVerifier{
				errBlockRootSeen: errors.New("not seen"),
				errSignature:     errors.New("bad signature"),
			})

			root := [32]byte{0x01}
			blockHash := [32]byte{0x02}
			env := testSignedSilaPayloadEnvelope(t, 1, tc.builderIdx, root, blockHash)
			msg := envelopeToPubsub(t, s, p, env)

			result, err := s.validateSilaPayloadEnvelope(ctx, "", msg)
			if tc.wantError {
				require.NotNil(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.result, result)
		})
	}
}

func testSignedSilaPayloadEnvelope(t *testing.T, slot primitives.Slot, builderIdx primitives.BuilderIndex, root, blockHash [32]byte) *silapb.SignedSilaPayloadEnvelope {
	t.Helper()

	payload := &silaenginev1.SilaPayloadGloas{
		ParentHash:    bytes.Repeat([]byte{0x01}, 32),
		FeeRecipient:  bytes.Repeat([]byte{0x02}, 20),
		StateRoot:     bytes.Repeat([]byte{0x03}, 32),
		ReceiptsRoot:  bytes.Repeat([]byte{0x04}, 32),
		LogsBloom:     bytes.Repeat([]byte{0x05}, 256),
		PrevRandao:    bytes.Repeat([]byte{0x06}, 32),
		BlockNumber:   1,
		GasLimit:      2,
		GasUsed:       3,
		Timestamp:     4,
		BaseFeePerGas: bytes.Repeat([]byte{0x07}, 32),
		BlockHash:     blockHash[:],
		Transactions:  [][]byte{},
		Withdrawals:   []*silaenginev1.Withdrawal{},
		BlobGasUsed:   0,
		ExcessBlobGas: 0,
		SlotNumber:    slot,
	}

	return &silapb.SignedSilaPayloadEnvelope{
		Message: &silapb.SilaPayloadEnvelope{
			Payload: payload,
			ExecutionRequests: &silaenginev1.ExecutionRequests{
				Deposits: []*silaenginev1.DepositRequest{},
			},
			BuilderIndex:          builderIdx,
			BeaconBlockRoot:       root[:],
			ParentBeaconBlockRoot: make([]byte, 32),
		},
		Signature: bytes.Repeat([]byte{0xAA}, 96),
	}
}
