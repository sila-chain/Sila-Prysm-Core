package blockchain

import (
	"bytes"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed"
	statefeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec"
	mockSila "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/crypto/bls"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
)

func gloasEnvelopeFixture(t *testing.T, blockRoot [32]byte) (*silapb.BeaconStateGloas, *silapb.SignedBeaconBlockGloas, *silapb.SignedSilaPayloadEnvelope) {
	t.Helper()

	cfg := params.BeaconConfig()
	slot := primitives.Slot(5)
	parentBeaconRoot := bytes.Repeat([]byte{0x11}, 32)
	blockHash := bytesutil.ToBytes32([]byte("payload-hash"))

	sk, err := bls.RandKey()
	require.NoError(t, err)
	pk := sk.PublicKey().Marshal()

	// Get base state and patch the state to be consistent with the payload we will build and sign.
	base, blk := testGloasState(t, slot, bytesutil.ToBytes32(parentBeaconRoot), blockHash)
	base.Fork = &silapb.Fork{
		CurrentVersion:  bytes.Repeat([]byte{0x01}, 4),
		PreviousVersion: bytes.Repeat([]byte{0x01}, 4),
		Epoch:           0,
	}
	base.GenesisValidatorsRoot = make([]byte, 32)
	base.Builders = []*silapb.Builder{{
		Pubkey:      pk,
		Version:     []byte{0},
		SilaAddress: make([]byte, 20),
	}}

	emptyRequestsRoot, err := silaenginev1.EmptySilaRequestsHashTreeRoot()
	require.NoError(t, err)

	base.LatestSilaPayloadBid.SilaRequestsRoot = emptyRequestsRoot[:]
	base.LatestSilaPayloadBid.BlobKzgCommitments = nil

	// Build a payload that is consistent with the committed bid and the state.
	bid := base.LatestSilaPayloadBid
	payload := &silaenginev1.SilaPayloadGloas{
		ParentHash:    base.LatestBlockHash,
		FeeRecipient:  make([]byte, 20),
		StateRoot:     make([]byte, 32),
		ReceiptsRoot:  make([]byte, 32),
		LogsBloom:     make([]byte, 256),
		PrevRandao:    bid.PrevRandao,
		BlockNumber:   1,
		GasLimit:      bid.GasLimit,
		Timestamp:     uint64(slot) * cfg.SecondsPerSlot,
		ExtraData:     []byte{},
		BaseFeePerGas: make([]byte, 32),
		BlockHash:     bid.BlockHash,
		Transactions:  [][]byte{},
		Withdrawals:   []*silaenginev1.Withdrawal{},
		SlotNumber:    slot,
	}

	// Build and sign the envelope.
	envelope := &silapb.SilaPayloadEnvelope{
		BuilderIndex:          0,
		BeaconBlockRoot:       blockRoot[:],
		ParentBeaconBlockRoot: parentBeaconRoot,
		Payload:               payload,
		SilaRequests:          &silaenginev1.SilaRequests{},
	}

	domain, err := signing.Domain(base.Fork, slots.ToEpoch(slot), cfg.DomainBeaconBuilder, base.GenesisValidatorsRoot)
	require.NoError(t, err)
	signingRoot, err := signing.ComputeSigningRoot(envelope, domain)
	require.NoError(t, err)
	signedProto := &silapb.SignedSilaPayloadEnvelope{
		Message:   envelope,
		Signature: sk.Sign(signingRoot[:]).Marshal(),
	}

	return base, blk, signedProto
}

// TestReceiveSilaPayloadEnvelope_EmitEvents verifies the event(`sila_payload`
// and `sila_payload_available`) emission behavior of receiver.
// Key regression: Independent of EL validation, `sila_payload_available`
// must be emitted as soon as the payload data is available,
// while `sila_payload` must only be emitted if the payload is imported successfully.
func TestReceiveSilaPayloadEnvelope_EmitEvents(t *testing.T) {
	tests := []struct {
		name          string
		engine        *mockSila.SilaEngineClient
		wantErr       bool
		wantAvailable int
		wantProcessed int
	}{
		{
			name:          "valid payload emits available and processed",
			engine:        &mockSila.SilaEngineClient{},
			wantErr:       false,
			wantAvailable: 1,
			wantProcessed: 1,
		},
		{
			name:          "EL-invalid still emits available but not processed",
			engine:        &mockSila.SilaEngineClient{ErrNewPayload: silaexec.ErrInvalidPayloadStatus},
			wantErr:       true,
			wantAvailable: 1,
			wantProcessed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := setupGloasService(t, tt.engine)
			ctx := t.Context()

			blockRoot := bytesutil.ToBytes32([]byte("envelope-root"))
			base, blk, signedProto := gloasEnvelopeFixture(t, blockRoot)
			insertGloasBlock(t, s, base, blk, blockRoot)

			events := make(chan *feed.Event, 10)
			sub := s.cfg.StateNotifier.StateFeed().Subscribe(events)
			defer sub.Unsubscribe()

			signed, err := blocks.WrappedROSignedSilaPayloadEnvelope(signedProto)
			require.NoError(t, err)

			err = s.ReceiveSilaPayloadEnvelope(ctx, signed)
			if tt.wantErr {
				require.NotNil(t, err)
				require.Equal(t, true, IsInvalidBlock(err))
			} else {
				require.NoError(t, err)
			}

			got := countStateEventsByType(events)
			require.Equal(t, tt.wantAvailable, got[statefeed.SilaPayloadAvailable])
			require.Equal(t, tt.wantProcessed, got[statefeed.SilaPayloadProcessed])
		})
	}
}

// countStateEventsByType is a helper function for counting the number of events
// of each type received on a channel.
func countStateEventsByType(ch chan *feed.Event) map[feed.EventType]int {
	got := make(map[feed.EventType]int)
	for {
		select {
		case e := <-ch:
			got[e.Type]++
		default:
			return got
		}
	}
}
