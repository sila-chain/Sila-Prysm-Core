package beacon_api

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/beacon-api/mock"
	"github.com/sila-chain/Sila/common/hexutil"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"
)

const proposerPreferencesEndpoint = "/sila/v1/validator/proposer_preferences"

func TestSubmitSignedProposerPreferences_Valid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dependentRoot := bytes.Repeat([]byte{0xcc}, 32)
	feeRecipient := bytes.Repeat([]byte{0xab}, 20)
	signature := bytes.Repeat([]byte{0x01}, 96)

	expected := []*structs.SignedProposerPreferences{{
		Message: &structs.ProposerPreferences{
			DependentRoot:  hexutil.Encode(dependentRoot),
			ProposalSlot:   "32",
			ValidatorIndex: "2",
			FeeRecipient:   hexutil.Encode(feeRecipient),
			TargetGasLimit: "30000000",
		},
		Signature: hexutil.Encode(signature),
	}}
	body, err := json.Marshal(expected)
	require.NoError(t, err)

	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().Post(
		gomock.Any(),
		proposerPreferencesEndpoint,
		map[string]string{"Eth-Consensus-Version": "gloas"},
		bytes.NewBuffer(body),
		nil,
	).Return(nil).Times(1)

	client := &beaconApiValidatorClient{handler: handler}
	err = client.submitSignedProposerPreferences(t.Context(), []*silapb.SignedProposerPreferences{{
		Message: &silapb.ProposerPreferences{
			DependentRoot:  dependentRoot,
			ProposalSlot:   32,
			ValidatorIndex: 2,
			FeeRecipient:   feeRecipient,
			TargetGasLimit: 30_000_000,
		},
		Signature: signature,
	}})
	require.NoError(t, err)
}

func TestSubmitSignedProposerPreferences_HandlerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := mock.NewMockJsonRestHandler(ctrl)
	handler.EXPECT().Post(
		gomock.Any(),
		proposerPreferencesEndpoint,
		map[string]string{"Eth-Consensus-Version": "gloas"},
		gomock.Any(),
		nil,
	).Return(errors.New("foo error")).Times(1)

	client := &beaconApiValidatorClient{handler: handler}
	err := client.submitSignedProposerPreferences(t.Context(), []*silapb.SignedProposerPreferences{{
		Message:   &silapb.ProposerPreferences{DependentRoot: bytes.Repeat([]byte{0xcc}, 32), FeeRecipient: bytes.Repeat([]byte{0xab}, 20)},
		Signature: bytes.Repeat([]byte{0x01}, 96),
	}})
	assert.ErrorContains(t, "foo error", err)
}

func TestSubmitSignedProposerPreferences_NilEntry(t *testing.T) {
	client := &beaconApiValidatorClient{}
	err := client.submitSignedProposerPreferences(t.Context(), []*silapb.SignedProposerPreferences{nil})
	assert.ErrorContains(t, "is nil", err)
}
