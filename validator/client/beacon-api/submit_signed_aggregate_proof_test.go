package beacon_api

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/beacon-api/mock"
	testhelpers "github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/beacon-api/test-helpers"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
)

func TestSubmitSignedAggregateSelectionProof_Valid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signedAggregateAndProof := generateSignedAggregateAndProofJson()
	marshalledSignedAggregateSignedAndProof, err := json.Marshal([]*structs.SignedAggregateAttestationAndProof{jsonifySignedAggregateAndProof(signedAggregateAndProof)})
	require.NoError(t, err)

	ctx := t.Context()
	headers := map[string]string{"Eth-Consensus-Version": version.String(signedAggregateAndProof.Message.Version())}
	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().Post(
		gomock.Any(),
		"/sila/v2/validator/aggregate_and_proofs",
		headers,
		bytes.NewBuffer(marshalledSignedAggregateSignedAndProof),
		nil,
	).Return(
		nil,
	).Times(1)

	attestationDataRoot, err := signedAggregateAndProof.Message.Aggregate.Data.HashTreeRoot()
	require.NoError(t, err)

	validatorClient := &beaconApiValidatorClient{handler: handler}
	resp, err := validatorClient.submitSignedAggregateSelectionProof(ctx, &silapb.SignedAggregateSubmitRequest{
		SignedAggregateAndProof: signedAggregateAndProof,
	})
	require.NoError(t, err)
	assert.DeepEqual(t, attestationDataRoot[:], resp.AttestationDataRoot)
}

func TestSubmitSignedAggregateSelectionProof_BadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signedAggregateAndProof := generateSignedAggregateAndProofJson()
	marshalledSignedAggregateSignedAndProof, err := json.Marshal([]*structs.SignedAggregateAttestationAndProof{jsonifySignedAggregateAndProof(signedAggregateAndProof)})
	require.NoError(t, err)

	ctx := t.Context()
	headers := map[string]string{"Eth-Consensus-Version": version.String(signedAggregateAndProof.Message.Version())}
	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().Post(
		gomock.Any(),
		"/sila/v2/validator/aggregate_and_proofs",
		headers,
		bytes.NewBuffer(marshalledSignedAggregateSignedAndProof),
		nil,
	).Return(
		errors.New("bad request"),
	).Times(1)

	validatorClient := &beaconApiValidatorClient{handler: handler}
	_, err = validatorClient.submitSignedAggregateSelectionProof(ctx, &silapb.SignedAggregateSubmitRequest{
		SignedAggregateAndProof: signedAggregateAndProof,
	})
	assert.ErrorContains(t, "bad request", err)
}

func TestSubmitSignedAggregateSelectionProofElectra_Valid(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	params.BeaconConfig().ElectraForkEpoch = 0
	params.BeaconConfig().FuluForkEpoch = 100

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signedAggregateAndProofElectra := generateSignedAggregateAndProofElectraJson()
	marshalledSignedAggregateSignedAndProofElectra, err := json.Marshal([]*structs.SignedAggregateAttestationAndProofElectra{jsonifySignedAggregateAndProofElectra(signedAggregateAndProofElectra)})
	require.NoError(t, err)

	ctx := t.Context()
	expectedVersion := version.String(slots.ToForkVersion(signedAggregateAndProofElectra.Message.Aggregate.Data.Slot))
	headers := map[string]string{"Eth-Consensus-Version": expectedVersion}
	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().Post(
		gomock.Any(),
		"/sila/v2/validator/aggregate_and_proofs",
		headers,
		bytes.NewBuffer(marshalledSignedAggregateSignedAndProofElectra),
		nil,
	).Return(
		nil,
	).Times(1)

	attestationDataRoot, err := signedAggregateAndProofElectra.Message.Aggregate.Data.HashTreeRoot()
	require.NoError(t, err)

	validatorClient := &beaconApiValidatorClient{handler: handler}
	resp, err := validatorClient.submitSignedAggregateSelectionProofElectra(ctx, &silapb.SignedAggregateSubmitElectraRequest{
		SignedAggregateAndProof: signedAggregateAndProofElectra,
	})
	require.NoError(t, err)
	assert.DeepEqual(t, attestationDataRoot[:], resp.AttestationDataRoot)
}

func TestSubmitSignedAggregateSelectionProofElectra_BadRequest(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	params.BeaconConfig().ElectraForkEpoch = 0
	params.BeaconConfig().FuluForkEpoch = 100

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signedAggregateAndProofElectra := generateSignedAggregateAndProofElectraJson()
	marshalledSignedAggregateSignedAndProofElectra, err := json.Marshal([]*structs.SignedAggregateAttestationAndProofElectra{jsonifySignedAggregateAndProofElectra(signedAggregateAndProofElectra)})
	require.NoError(t, err)

	ctx := t.Context()
	expectedVersion := version.String(slots.ToForkVersion(signedAggregateAndProofElectra.Message.Aggregate.Data.Slot))
	headers := map[string]string{"Eth-Consensus-Version": expectedVersion}
	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().Post(
		gomock.Any(),
		"/sila/v2/validator/aggregate_and_proofs",
		headers,
		bytes.NewBuffer(marshalledSignedAggregateSignedAndProofElectra),
		nil,
	).Return(
		errors.New("bad request"),
	).Times(1)

	validatorClient := &beaconApiValidatorClient{handler: handler}
	_, err = validatorClient.submitSignedAggregateSelectionProofElectra(ctx, &silapb.SignedAggregateSubmitElectraRequest{
		SignedAggregateAndProof: signedAggregateAndProofElectra,
	})
	assert.ErrorContains(t, "bad request", err)
}

func TestSubmitSignedAggregateSelectionProofElectra_FuluVersion(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	params.BeaconConfig().ElectraForkEpoch = 0
	params.BeaconConfig().FuluForkEpoch = 1

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signedAggregateAndProofElectra := generateSignedAggregateAndProofElectraJson()
	marshalledSignedAggregateSignedAndProofElectra, err := json.Marshal([]*structs.SignedAggregateAttestationAndProofElectra{jsonifySignedAggregateAndProofElectra(signedAggregateAndProofElectra)})
	require.NoError(t, err)

	ctx := t.Context()
	expectedVersion := version.String(slots.ToForkVersion(signedAggregateAndProofElectra.Message.Aggregate.Data.Slot))
	headers := map[string]string{"Eth-Consensus-Version": expectedVersion}
	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().Post(
		gomock.Any(),
		"/sila/v2/validator/aggregate_and_proofs",
		headers,
		bytes.NewBuffer(marshalledSignedAggregateSignedAndProofElectra),
		nil,
	).Return(
		nil,
	).Times(1)

	attestationDataRoot, err := signedAggregateAndProofElectra.Message.Aggregate.Data.HashTreeRoot()
	require.NoError(t, err)

	validatorClient := &beaconApiValidatorClient{handler: handler}
	resp, err := validatorClient.submitSignedAggregateSelectionProofElectra(ctx, &silapb.SignedAggregateSubmitElectraRequest{
		SignedAggregateAndProof: signedAggregateAndProofElectra,
	})
	require.NoError(t, err)
	assert.DeepEqual(t, attestationDataRoot[:], resp.AttestationDataRoot)
}

func generateSignedAggregateAndProofJson() *silapb.SignedAggregateAttestationAndProof {
	return &silapb.SignedAggregateAttestationAndProof{
		Message: &silapb.AggregateAttestationAndProof{
			AggregatorIndex: 72,
			Aggregate: &silapb.Attestation{
				AggregationBits: testhelpers.FillByteSlice(4, 74),
				Data: &silapb.AttestationData{
					Slot:            75,
					CommitteeIndex:  76,
					BeaconBlockRoot: testhelpers.FillByteSlice(32, 38),
					Source: &silapb.Checkpoint{
						Epoch: 78,
						Root:  testhelpers.FillByteSlice(32, 79),
					},
					Target: &silapb.Checkpoint{
						Epoch: 80,
						Root:  testhelpers.FillByteSlice(32, 81),
					},
				},
				Signature: testhelpers.FillByteSlice(96, 82),
			},
			SelectionProof: testhelpers.FillByteSlice(96, 82),
		},
		Signature: testhelpers.FillByteSlice(96, 82),
	}
}

func generateSignedAggregateAndProofElectraJson() *silapb.SignedAggregateAttestationAndProofElectra {
	return &silapb.SignedAggregateAttestationAndProofElectra{
		Message: &silapb.AggregateAttestationAndProofElectra{
			AggregatorIndex: 72,
			Aggregate: &silapb.AttestationElectra{
				AggregationBits: testhelpers.FillByteSlice(4, 74),
				Data: &silapb.AttestationData{
					Slot:            75,
					CommitteeIndex:  76,
					BeaconBlockRoot: testhelpers.FillByteSlice(32, 38),
					Source: &silapb.Checkpoint{
						Epoch: 78,
						Root:  testhelpers.FillByteSlice(32, 79),
					},
					Target: &silapb.Checkpoint{
						Epoch: 80,
						Root:  testhelpers.FillByteSlice(32, 81),
					},
				},
				Signature:     testhelpers.FillByteSlice(96, 82),
				CommitteeBits: testhelpers.FillByteSlice(8, 83),
			},
			SelectionProof: testhelpers.FillByteSlice(96, 84),
		},
		Signature: testhelpers.FillByteSlice(96, 85),
	}
}
