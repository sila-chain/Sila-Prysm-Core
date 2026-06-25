package beacon_api

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/beacon-api/mock"
	"github.com/sila-chain/Sila/common/hexutil"
	"go.uber.org/mock/gomock"
)

const submitSignedContributionAndProofTestEndpoint = "/sila/v1/validator/contribution_and_proofs"

func TestSubmitSignedContributionAndProof_Valid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jsonContributionAndProofs := []structs.SignedContributionAndProof{
		{
			Message: &structs.ContributionAndProof{
				AggregatorIndex: "1",
				Contribution: &structs.SyncCommitteeContribution{
					Slot:              "2",
					BeaconBlockRoot:   hexutil.Encode([]byte{3}),
					SubcommitteeIndex: "4",
					AggregationBits:   hexutil.Encode([]byte{5}),
					Signature:         hexutil.Encode([]byte{6}),
				},
				SelectionProof: hexutil.Encode([]byte{7}),
			},
			Signature: hexutil.Encode([]byte{8}),
		},
	}

	marshalledContributionAndProofs, err := json.Marshal(jsonContributionAndProofs)
	require.NoError(t, err)

	ctx := t.Context()

	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().Post(
		gomock.Any(),
		submitSignedContributionAndProofTestEndpoint,
		nil,
		bytes.NewBuffer(marshalledContributionAndProofs),
		nil,
	).Return(
		nil,
	).Times(1)

	contributionAndProof := &silapb.SignedContributionAndProof{
		Message: &silapb.ContributionAndProof{
			AggregatorIndex: 1,
			Contribution: &silapb.SyncCommitteeContribution{
				Slot:              2,
				BlockRoot:         []byte{3},
				SubcommitteeIndex: 4,
				AggregationBits:   []byte{5},
				Signature:         []byte{6},
			},
			SelectionProof: []byte{7},
		},
		Signature: []byte{8},
	}

	validatorClient := &beaconApiValidatorClient{handler: handler}
	err = validatorClient.submitSignedContributionAndProof(ctx, contributionAndProof)
	require.NoError(t, err)
}

func TestSubmitSignedContributionAndProof_Error(t *testing.T) {
	testCases := []struct {
		name                 string
		data                 *silapb.SignedContributionAndProof
		expectedErrorMessage string
		httpRequestExpected  bool
	}{
		{
			name:                 "nil signed contribution and proof",
			data:                 nil,
			expectedErrorMessage: "signed contribution and proof is nil",
		},
		{
			name:                 "nil message",
			data:                 &silapb.SignedContributionAndProof{},
			expectedErrorMessage: "signed contribution and proof message is nil",
		},
		{
			name: "nil contribution",
			data: &silapb.SignedContributionAndProof{
				Message: &silapb.ContributionAndProof{},
			},
			expectedErrorMessage: "signed contribution and proof contribution is nil",
		},
		{
			name: "bad request",
			data: &silapb.SignedContributionAndProof{
				Message: &silapb.ContributionAndProof{
					Contribution: &silapb.SyncCommitteeContribution{},
				},
			},
			httpRequestExpected:  true,
			expectedErrorMessage: "foo error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := t.Context()

			handler := mock.NewMockJsonRestHandler(ctrl)
			if testCase.httpRequestExpected {
				handler.EXPECT().Post(
					gomock.Any(),
					submitSignedContributionAndProofTestEndpoint,
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(
					errors.New("foo error"),
				).Times(1)
			}

			validatorClient := &beaconApiValidatorClient{handler: handler}
			err := validatorClient.submitSignedContributionAndProof(ctx, testCase.data)
			assert.ErrorContains(t, testCase.expectedErrorMessage, err)
		})
	}
}
