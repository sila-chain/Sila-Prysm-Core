package validator

import (
	"context"
	"math/big"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	mockp2p "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	consensusblocks "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func testGloasBlock(t *testing.T) (*consensusblocks.GetPayloadResponse, interfaces.SignedBeaconBlock) {
	t.Helper()

	payload := &silaenginev1.SilaPayloadGloas{
		ParentHash:    make([]byte, 32),
		FeeRecipient:  make([]byte, 20),
		StateRoot:     make([]byte, 32),
		ReceiptsRoot:  make([]byte, 32),
		LogsBloom:     make([]byte, 256),
		PrevRandao:    make([]byte, 32),
		BaseFeePerGas: make([]byte, 32),
		BlockHash:     make([]byte, 32),
		ExtraData:     make([]byte, 0),
	}
	ed, err := consensusblocks.WrappedSilaPayloadGloas(payload)
	require.NoError(t, err)

	local := &consensusblocks.GetPayloadResponse{
		SilaData:     ed,
		Bid:               big.NewInt(0),
		SilaRequests: &silaenginev1.SilaRequests{},
	}

	sBlk, err := consensusblocks.NewSignedBeaconBlock(util.NewBeaconBlockGloas())
	require.NoError(t, err)

	return local, sBlk
}

func TestStoreSilaPayloadEnvelope(t *testing.T) {
	local, sBlk := testGloasBlock(t)

	vs := &Server{SilaPayloadEnvelopeCache: cache.NewSilaPayloadEnvelopeCache()}
	err := vs.storeSilaPayloadEnvelope(sBlk, local)
	require.NoError(t, err)

	contents, ok := vs.SilaPayloadEnvelopeCache.Contents()
	require.Equal(t, true, ok)
	require.Equal(t, sBlk.Block().Slot(), contents.Envelope.Payload.SlotNumber)
}

func TestExtractSilaPayloadGloas(t *testing.T) {
	payload := &silaenginev1.SilaPayloadGloas{
		ParentHash:    make([]byte, 32),
		FeeRecipient:  make([]byte, 20),
		StateRoot:     make([]byte, 32),
		ReceiptsRoot:  make([]byte, 32),
		LogsBloom:     make([]byte, 256),
		PrevRandao:    make([]byte, 32),
		BaseFeePerGas: make([]byte, 32),
		BlockHash:     make([]byte, 32),
		ExtraData:     make([]byte, 0),
	}
	ed, err := consensusblocks.WrappedSilaPayloadGloas(payload)
	require.NoError(t, err)

	local := &consensusblocks.GetPayloadResponse{
		SilaData: ed,
		Bid:           big.NewInt(0),
	}

	result := extractSilaPayloadGloas(local)
	require.NotNil(t, result)
	require.DeepEqual(t, payload, result)
}

func TestExtractSilaPayloadGloas_Nil(t *testing.T) {
	require.Equal(t, true, extractSilaPayloadGloas(nil) == nil)
	require.Equal(t, true, extractSilaPayloadGloas(&consensusblocks.GetPayloadResponse{}) == nil)
}

func TestGetSilaPayloadEnvelopeRPC_NilRequest(t *testing.T) {
	vs := &Server{}
	_, err := vs.GetSilaPayloadEnvelope(t.Context(), nil)
	require.ErrorContains(t, "request cannot be nil", err)
}

func TestGetSilaPayloadEnvelopeRPC_PreFork(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.GloasForkEpoch = 10
	params.OverrideBeaconConfig(cfg)

	vs := &Server{}
	_, err := vs.GetSilaPayloadEnvelope(t.Context(), &silapb.SilaPayloadEnvelopeRequest{
		Slot: 0, // epoch 0, before GloasForkEpoch 10
	})
	require.ErrorContains(t, "not supported before Gloas fork", err)
}

func TestPublishSilaPayloadEnvelope_NilRequest(t *testing.T) {
	vs := &Server{}
	_, err := vs.PublishSilaPayloadEnvelope(t.Context(), nil)
	require.ErrorContains(t, "signed envelope or payload cannot be nil", err)

	_, err = vs.PublishSilaPayloadEnvelope(t.Context(), &silapb.SignedSilaPayloadEnvelope{})
	require.ErrorContains(t, "signed envelope or payload cannot be nil", err)
}

func TestPublishSilaPayloadEnvelope_PreFork(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.GloasForkEpoch = 10
	params.OverrideBeaconConfig(cfg)

	vs := &Server{}
	_, err := vs.PublishSilaPayloadEnvelope(t.Context(), &silapb.SignedSilaPayloadEnvelope{
		Message: &silapb.SilaPayloadEnvelope{
			Payload: &silaenginev1.SilaPayloadGloas{SlotNumber: 0}, // epoch 0, before GloasForkEpoch 10
		},
	})
	require.ErrorContains(t, "not supported before Gloas fork", err)
}

func TestGetSilaPayloadEnvelopeRPC_Success(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.GloasForkEpoch = 0
	params.OverrideBeaconConfig(cfg)

	envelope := &silapb.SilaPayloadEnvelope{
		Payload: &silaenginev1.SilaPayloadGloas{
			ParentHash:    make([]byte, 32),
			FeeRecipient:  make([]byte, 20),
			StateRoot:     make([]byte, 32),
			ReceiptsRoot:  make([]byte, 32),
			LogsBloom:     make([]byte, 256),
			PrevRandao:    make([]byte, 32),
			BaseFeePerGas: make([]byte, 32),
			BlockHash:     make([]byte, 32),
			SlotNumber:    1,
		},
		BuilderIndex:    primitives.BuilderIndex(0),
		BeaconBlockRoot: make([]byte, 32),
	}

	vs := &Server{SilaPayloadEnvelopeCache: cache.NewSilaPayloadEnvelopeCache()}
	vs.SilaPayloadEnvelopeCache.Set(&cache.SilaPayloadContents{Envelope: envelope})

	resp, err := vs.GetSilaPayloadEnvelope(t.Context(), &silapb.SilaPayloadEnvelopeRequest{
		Slot: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.DeepEqual(t, envelope, resp.Envelope)
}

func TestPublishSilaPayloadEnvelope_Success(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.GloasForkEpoch = 0
	params.OverrideBeaconConfig(cfg)

	broadcaster := &mockp2p.MockBroadcaster{}
	receiver := &mockSilaPayloadEnvelopeReceiver{}
	vs := &Server{
		P2P:                              broadcaster,
		SilaPayloadEnvelopeReceiver: receiver,
	}

	req := &silapb.SignedSilaPayloadEnvelope{
		Message: &silapb.SilaPayloadEnvelope{
			Payload: &silaenginev1.SilaPayloadGloas{
				ParentHash:    make([]byte, 32),
				FeeRecipient:  make([]byte, 20),
				StateRoot:     make([]byte, 32),
				ReceiptsRoot:  make([]byte, 32),
				LogsBloom:     make([]byte, 256),
				PrevRandao:    make([]byte, 32),
				BaseFeePerGas: make([]byte, 32),
				BlockHash:     make([]byte, 32),
				ExtraData:     make([]byte, 0),
				SlotNumber:    1,
			},
			SilaRequests:     &silaenginev1.SilaRequests{},
			BuilderIndex:          0,
			BeaconBlockRoot:       make([]byte, 32),
			ParentBeaconBlockRoot: make([]byte, 32),
		},
		Signature: make([]byte, 96),
	}

	resp, err := vs.PublishSilaPayloadEnvelope(t.Context(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, true, broadcaster.BroadcastCalled.Load())
	require.Equal(t, 1, len(broadcaster.BroadcastMessages))
	require.Equal(t, 1, receiver.calls)
}

func TestPublishSilaPayloadEnvelope_ImportFailureIsAborted(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.GloasForkEpoch = 0
	params.OverrideBeaconConfig(cfg)

	broadcaster := &mockp2p.MockBroadcaster{}
	receiver := &mockSilaPayloadEnvelopeReceiver{err: errors.New("import failed")}
	vs := &Server{
		P2P:                              broadcaster,
		SilaPayloadEnvelopeReceiver: receiver,
	}

	req := &silapb.SignedSilaPayloadEnvelope{
		Message: &silapb.SilaPayloadEnvelope{
			Payload: &silaenginev1.SilaPayloadGloas{
				ParentHash:    make([]byte, 32),
				FeeRecipient:  make([]byte, 20),
				StateRoot:     make([]byte, 32),
				ReceiptsRoot:  make([]byte, 32),
				LogsBloom:     make([]byte, 256),
				PrevRandao:    make([]byte, 32),
				BaseFeePerGas: make([]byte, 32),
				BlockHash:     make([]byte, 32),
				ExtraData:     make([]byte, 0),
				SlotNumber:    1,
			},
			SilaRequests:     &silaenginev1.SilaRequests{},
			BeaconBlockRoot:       make([]byte, 32),
			ParentBeaconBlockRoot: make([]byte, 32),
		},
		Signature: make([]byte, 96),
	}

	_, err := vs.PublishSilaPayloadEnvelope(t.Context(), req)
	require.NotNil(t, err)
	// Broadcast must have happened before the import failure (spec 202).
	require.Equal(t, true, broadcaster.BroadcastCalled.Load())
	require.Equal(t, codes.Aborted, status.Code(err))
}

type mockSilaPayloadEnvelopeReceiver struct {
	calls int
	err   error
}

func (m *mockSilaPayloadEnvelopeReceiver) ReceiveSilaPayloadEnvelope(_ context.Context, _ interfaces.ROSignedSilaPayloadEnvelope) error {
	m.calls++
	return m.err
}
