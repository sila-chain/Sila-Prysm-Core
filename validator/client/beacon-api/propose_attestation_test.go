package beacon_api

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/beacon-api/mock"
	testhelpers "github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/beacon-api/test-helpers"
	"go.uber.org/mock/gomock"
)

func TestProposeAttestation(t *testing.T) {
	attestation := &silapb.Attestation{
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
	}

	tests := []struct {
		name                 string
		attestation          *silapb.Attestation
		expectedErrorMessage string
		endpointError        error
		endpointCall         int
	}{
		{
			name:         "valid",
			attestation:  attestation,
			endpointCall: 1,
		},
		{
			name:                 "nil attestation",
			expectedErrorMessage: "attestation is nil",
		},
		{
			name: "nil attestation data",
			attestation: &silapb.Attestation{
				AggregationBits: testhelpers.FillByteSlice(4, 74),
				Signature:       testhelpers.FillByteSlice(96, 82),
			},
			expectedErrorMessage: "attestation is nil",
		},
		{
			name: "nil source checkpoint",
			attestation: &silapb.Attestation{
				AggregationBits: testhelpers.FillByteSlice(4, 74),
				Data: &silapb.AttestationData{
					Target: &silapb.Checkpoint{},
				},
				Signature: testhelpers.FillByteSlice(96, 82),
			},
			expectedErrorMessage: "attestation's source can't be nil",
		},
		{
			name: "nil target checkpoint",
			attestation: &silapb.Attestation{
				AggregationBits: testhelpers.FillByteSlice(4, 74),
				Data: &silapb.AttestationData{
					Source: &silapb.Checkpoint{},
				},
				Signature: testhelpers.FillByteSlice(96, 82),
			},
			expectedErrorMessage: "attestation's target can't be nil",
		},
		{
			name: "nil aggregation bits",
			attestation: &silapb.Attestation{
				Data: &silapb.AttestationData{
					Source: &silapb.Checkpoint{},
					Target: &silapb.Checkpoint{},
				},
				Signature: testhelpers.FillByteSlice(96, 82),
			},
			expectedErrorMessage: "attestation's bitfield can't be nil",
		},
		{
			name:                 "bad request",
			attestation:          attestation,
			expectedErrorMessage: "bad request",
			endpointError:        errors.New("bad request"),
			endpointCall:         1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			handler := mock.NewMockJsonRestHandler(ctrl)

			var marshalledAttestations []byte
			if helpers.ValidateNilAttestation(test.attestation) == nil {
				b, err := json.Marshal(jsonifyAttestations([]*silapb.Attestation{test.attestation}))
				require.NoError(t, err)
				marshalledAttestations = b
			}

			ctx := t.Context()

			headers := map[string]string{"Eth-Consensus-Version": version.String(test.attestation.Version())}
			handler.EXPECT().Post(
				gomock.Any(),
				"/sila/v2/beacon/pool/attestations",
				headers,
				bytes.NewBuffer(marshalledAttestations),
				nil,
			).Return(
				test.endpointError,
			).Times(test.endpointCall)

			validatorClient := &beaconApiValidatorClient{handler: handler}
			proposeResponse, err := validatorClient.proposeAttestation(ctx, test.attestation)
			if test.expectedErrorMessage != "" {
				require.ErrorContains(t, test.expectedErrorMessage, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, proposeResponse)

			expectedAttestationDataRoot, err := attestation.Data.HashTreeRoot()
			require.NoError(t, err)

			// Make sure that the attestation data root is set
			assert.DeepEqual(t, expectedAttestationDataRoot[:], proposeResponse.AttestationDataRoot)
		})
	}
}

func TestProposeAttestationElectra(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	params.BeaconConfig().ElectraForkEpoch = 0
	params.BeaconConfig().FuluForkEpoch = 1

	buildSingleAttestation := func(slot primitives.Slot) *silapb.SingleAttestation {
		targetEpoch := slots.ToEpoch(slot)
		sourceEpoch := targetEpoch
		if targetEpoch > 0 {
			sourceEpoch = targetEpoch - 1
		}
		return &silapb.SingleAttestation{
			AttesterIndex: 74,
			Data: &silapb.AttestationData{
				Slot:            slot,
				CommitteeIndex:  76,
				BeaconBlockRoot: testhelpers.FillByteSlice(32, 38),
				Source: &silapb.Checkpoint{
					Epoch: sourceEpoch,
					Root:  testhelpers.FillByteSlice(32, 79),
				},
				Target: &silapb.Checkpoint{
					Epoch: targetEpoch,
					Root:  testhelpers.FillByteSlice(32, 81),
				},
			},
			Signature:   testhelpers.FillByteSlice(96, 82),
			CommitteeId: 83,
		}
	}

	attestationElectra := buildSingleAttestation(0)
	attestationFulu := buildSingleAttestation(params.BeaconConfig().SlotsPerEpoch)

	tests := []struct {
		name                     string
		attestation              *silapb.SingleAttestation
		expectedConsensusVersion string
		expectedErrorMessage     string
		endpointError            error
		endpointCall             int
	}{
		{
			name:                     "valid electra",
			attestation:              attestationElectra,
			expectedConsensusVersion: version.String(slots.ToForkVersion(attestationElectra.GetData().GetSlot())),
			endpointCall:             1,
		},
		{
			name:                     "valid fulu consensus version",
			attestation:              attestationFulu,
			expectedConsensusVersion: version.String(slots.ToForkVersion(attestationFulu.GetData().GetSlot())),
			endpointCall:             1,
		},
		{
			name:                 "nil attestation",
			expectedErrorMessage: "attestation is nil",
		},
		{
			name: "nil attestation data",
			attestation: &silapb.SingleAttestation{
				AttesterIndex: 74,
				Signature:     testhelpers.FillByteSlice(96, 82),
				CommitteeId:   83,
			},
			expectedErrorMessage: "attestation is nil",
		},
		{
			name: "nil source checkpoint",
			attestation: &silapb.SingleAttestation{
				AttesterIndex: 74,
				Data: &silapb.AttestationData{
					Target: &silapb.Checkpoint{},
				},
				Signature:   testhelpers.FillByteSlice(96, 82),
				CommitteeId: 83,
			},
			expectedErrorMessage: "attestation's source can't be nil",
		},
		{
			name: "nil target checkpoint",
			attestation: &silapb.SingleAttestation{
				AttesterIndex: 74,
				Data: &silapb.AttestationData{
					Source: &silapb.Checkpoint{},
				},
				Signature:   testhelpers.FillByteSlice(96, 82),
				CommitteeId: 83,
			},
			expectedErrorMessage: "attestation's target can't be nil",
		},
		{
			name:        "bad request",
			attestation: attestationElectra,
			expectedConsensusVersion: version.String(
				slots.ToForkVersion(attestationElectra.GetData().GetSlot()),
			),
			expectedErrorMessage: "bad request",
			endpointError:        errors.New("bad request"),
			endpointCall:         1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			handler := mock.NewMockJsonRestHandler(ctrl)

			var marshalledAttestations []byte
			if helpers.ValidateNilAttestation(test.attestation) == nil {
				b, err := json.Marshal(jsonifySingleAttestations([]*silapb.SingleAttestation{test.attestation}))
				require.NoError(t, err)
				marshalledAttestations = b
			}

			ctx := t.Context()
			headerMatcher := gomock.Any()
			if test.expectedConsensusVersion != "" {
				headerMatcher = gomock.Eq(map[string]string{"Eth-Consensus-Version": test.expectedConsensusVersion})
			}
			handler.EXPECT().Post(
				gomock.Any(),
				"/sila/v2/beacon/pool/attestations",
				headerMatcher,
				bytes.NewBuffer(marshalledAttestations),
				nil,
			).Return(
				test.endpointError,
			).Times(test.endpointCall)

			validatorClient := &beaconApiValidatorClient{handler: handler}
			proposeResponse, err := validatorClient.proposeAttestationElectra(ctx, test.attestation)
			if test.expectedErrorMessage != "" {
				require.ErrorContains(t, test.expectedErrorMessage, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, proposeResponse)

			expectedAttestationDataRoot, err := test.attestation.Data.HashTreeRoot()
			require.NoError(t, err)

			// Make sure that the attestation data root is set
			assert.DeepEqual(t, expectedAttestationDataRoot[:], proposeResponse.AttestationDataRoot)
		})
	}
}
