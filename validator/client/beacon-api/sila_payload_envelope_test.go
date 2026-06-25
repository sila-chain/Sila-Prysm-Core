package beacon_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api"
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/network/httputil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/beacon-api/mock"
	"github.com/sila-chain/Sila/common/hexutil"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
)

func testProtoEnvelope() *silapb.SilaPayloadEnvelope {
	return &silapb.SilaPayloadEnvelope{
		Payload: &silaenginev1.SilaPayloadGloas{
			ParentHash:    bytesutil.PadTo([]byte("parent"), 32),
			FeeRecipient:  bytesutil.PadTo([]byte("fee"), 20),
			StateRoot:     bytesutil.PadTo([]byte("state"), 32),
			ReceiptsRoot:  bytesutil.PadTo([]byte("receipts"), 32),
			LogsBloom:     make([]byte, 256),
			PrevRandao:    bytesutil.PadTo([]byte("randao"), 32),
			BaseFeePerGas: bytesutil.PadTo([]byte{1}, 32),
			BlockHash:     bytesutil.PadTo([]byte("blockhash"), 32),
			Transactions:  [][]byte{},
			Withdrawals:   []*silaenginev1.Withdrawal{},
			SlotNumber:    primitives.Slot(100),
		},
		ExecutionRequests:     &silaenginev1.ExecutionRequests{},
		BuilderIndex:          primitives.BuilderIndex(42),
		BeaconBlockRoot:       bytesutil.PadTo([]byte("beacon-root"), 32),
		ParentBeaconBlockRoot: bytesutil.PadTo([]byte("parent-beacon-root"), 32),
	}
}

func TestGetSilaPayloadEnvelope_CachedHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := mock.NewMockJsonRestHandler(ctrl)
	// No Get expectation: cache hit must skip the HTTP call.

	envelope := testProtoEnvelope()
	client := &beaconApiValidatorClient{
		handler:       handler,
		envelopeCache: newSilaPayloadEnvelopeCache(),
	}
	client.envelopeCache.Add(100, envelope, nil, nil)

	full, blinded, err := client.getSilaPayloadEnvelope(t.Context(), 100, [32]byte{})
	require.NoError(t, err)
	require.NotNil(t, full)
	require.IsNil(t, blinded)
	assert.Equal(t, primitives.BuilderIndex(42), full.BuilderIndex)

	// Peek must leave the entry in the cache so the publish path can read blob data.
	cached, _, _ := client.envelopeCache.peek(100)
	require.NotNil(t, cached)
}

// Stateful: on a local cache miss the VC fetches the blinded envelope from the BN.
func TestGetSilaPayloadEnvelope_StatefulFetchesBlinded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	envelope := testProtoEnvelope()
	blinded, err := structs.WireBlindedFromFull(envelope)
	require.NoError(t, err)
	body, err := blinded.MarshalSSZ()
	require.NoError(t, err)

	root := bytesutil.ToBytes32(envelope.BeaconBlockRoot)
	respHeader := http.Header{}
	respHeader.Set("Content-Type", api.OctetStreamMediaType)

	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().GetSSZ(
		gomock.Any(),
		fmt.Sprintf("/sila/v1/validator/sila_payload_envelopes/100/%s", hexutil.Encode(root[:])),
	).Return(body, respHeader, nil)

	client := &beaconApiValidatorClient{handler: handler, envelopeCache: newSilaPayloadEnvelopeCache()}
	full, gotBlinded, err := client.getSilaPayloadEnvelope(t.Context(), 100, root)
	require.NoError(t, err)
	require.IsNil(t, full)
	require.NotNil(t, gotBlinded)
	assert.Equal(t, primitives.BuilderIndex(42), gotBlinded.BuilderIndex)
}

// Stateful publish sends the blinded envelope with Sila-Payload-Blinded: true.
func TestPublishBlindedSilaPayloadEnvelope(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signed := &silapb.SignedSilaPayloadEnvelope{Message: testProtoEnvelope(), Signature: bytesutil.PadTo([]byte("sig"), 96)}
	signedBlinded, err := structs.SignedWireBlindedFromFull(signed)
	require.NoError(t, err)
	expectedBody, err := signedBlinded.MarshalSSZ()
	require.NoError(t, err)

	expectedHeaders := map[string]string{
		api.VersionHeader:                 version.String(version.Gloas),
		api.SilaPayloadBlindedHeader: "true",
	}
	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().PostSSZ(
		gomock.Any(),
		"/sila/v1/beacon/sila_payload_envelopes",
		expectedHeaders,
		bytes.NewBuffer(expectedBody),
	).Return(nil, nil, nil)

	client := &beaconApiValidatorClient{handler: handler}
	resp, err := client.publishBlindedSilaPayloadEnvelope(t.Context(), signedBlinded)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestPublishBlindedSilaPayloadEnvelope_JSONFallbackOn406(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signed := &silapb.SignedSilaPayloadEnvelope{Message: testProtoEnvelope(), Signature: bytesutil.PadTo([]byte("sig"), 96)}
	signedBlinded, err := structs.SignedWireBlindedFromFull(signed)
	require.NoError(t, err)
	msg, err := structs.BlindedSilaPayloadEnvelopeFromConsensus(signedBlinded.Message)
	require.NoError(t, err)
	expectedJSON, err := json.Marshal(&structs.SignedBlindedSilaPayloadEnvelope{
		Message:   msg,
		Signature: hexutil.Encode(signedBlinded.Signature),
	})
	require.NoError(t, err)

	expectedHeaders := map[string]string{
		api.VersionHeader:                 version.String(version.Gloas),
		api.SilaPayloadBlindedHeader: "true",
	}
	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().PostSSZ(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(nil, nil, &httputil.DefaultJsonError{Code: http.StatusNotAcceptable, Message: "not acceptable"})
	handler.EXPECT().Post(
		gomock.Any(),
		"/sila/v1/beacon/sila_payload_envelopes",
		expectedHeaders,
		bytes.NewBuffer(expectedJSON),
		nil,
	).Return(nil)

	client := &beaconApiValidatorClient{handler: handler}
	resp, err := client.publishBlindedSilaPayloadEnvelope(t.Context(), signedBlinded)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestPublishSilaPayloadEnvelope_StatelessSendsContents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	envelope := testProtoEnvelope()
	signed := &silapb.SignedSilaPayloadEnvelope{
		Message:   envelope,
		Signature: bytesutil.PadTo([]byte("sig"), 96),
	}
	blob := bytesutil.PadTo([]byte("blob"), 131072)
	proof := bytesutil.PadTo([]byte("proof"), 48)

	expectedBody, err := (&silapb.SignedSilaPayloadEnvelopeContents{
		SignedSilaPayloadEnvelope: signed,
		KzgProofs:                      [][]byte{proof},
		Blobs:                          [][]byte{blob},
	}).MarshalSSZ()
	require.NoError(t, err)

	expectedHeaders := map[string]string{
		api.VersionHeader:                 version.String(version.Gloas),
		api.SilaPayloadBlindedHeader: "false",
	}
	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().PostSSZ(
		gomock.Any(),
		"/sila/v1/beacon/sila_payload_envelopes",
		expectedHeaders,
		bytes.NewBuffer(expectedBody),
	).Return(nil, nil, nil)

	client := &beaconApiValidatorClient{
		handler:       handler,
		stateless:     true,
		envelopeCache: newSilaPayloadEnvelopeCache(),
	}
	client.envelopeCache.Add(primitives.Slot(envelope.Payload.SlotNumber), envelope, [][]byte{blob}, [][]byte{proof})

	resp, err := client.publishSilaPayloadEnvelope(t.Context(), signed)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Cache must be drained after publish.
	cached, _, _ := client.envelopeCache.peek(primitives.Slot(envelope.Payload.SlotNumber))
	assert.Equal(t, (*silapb.SilaPayloadEnvelope)(nil), cached)
}

func TestPublishSilaPayloadEnvelope_StatelessCacheMissErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signed := &silapb.SignedSilaPayloadEnvelope{
		Message:   testProtoEnvelope(),
		Signature: bytesutil.PadTo([]byte("sig"), 96),
	}

	// No PostSSZ/Post expectation — must error before any HTTP call.
	handler := mock.NewMockJsonRestHandler(ctrl)
	client := &beaconApiValidatorClient{
		handler:       handler,
		envelopeCache: newSilaPayloadEnvelopeCache(),
	}

	_, err := client.publishSilaPayloadEnvelope(t.Context(), signed)
	assert.ErrorContains(t, "stateless publish: envelope cache miss", err)
}

func TestPublishSilaPayloadEnvelope_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	envelope := testProtoEnvelope()
	signed := &silapb.SignedSilaPayloadEnvelope{
		Message:   envelope,
		Signature: bytesutil.PadTo([]byte("sig"), 96),
	}

	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().PostSSZ(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(nil, nil, errors.New("server error"))

	client := &beaconApiValidatorClient{handler: handler, envelopeCache: newSilaPayloadEnvelopeCache()}
	client.envelopeCache.Add(primitives.Slot(envelope.Payload.SlotNumber), envelope, nil, nil)

	_, err := client.publishSilaPayloadEnvelope(t.Context(), signed)
	assert.ErrorContains(t, "server error", err)
}

func TestEnvelopeRoundTrip(t *testing.T) {
	envelope := testProtoEnvelope()
	jsonEnvelope, err := structs.SilaPayloadEnvelopeFromConsensus(envelope)
	require.NoError(t, err)

	roundTripped, err := jsonEnvelope.ToConsensus()
	require.NoError(t, err)

	assert.Equal(t, envelope.BuilderIndex, roundTripped.BuilderIndex)
	assert.Equal(t, envelope.Payload.SlotNumber, roundTripped.Payload.SlotNumber)
	assert.DeepEqual(t, envelope.BeaconBlockRoot, roundTripped.BeaconBlockRoot)
	assert.Equal(t, hexutil.Encode(envelope.Payload.BlockHash), hexutil.Encode(roundTripped.Payload.BlockHash))
}
