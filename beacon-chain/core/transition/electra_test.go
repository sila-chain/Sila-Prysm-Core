package transition_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/transition"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func TestProcessOperationsWithNilRequests(t *testing.T) {
	tests := []struct {
		name      string
		modifyBlk func(blockElectra *silapb.SignedBeaconBlockElectra)
		errMsg    string
	}{
		{
			name: "Nil deposit request",
			modifyBlk: func(blk *silapb.SignedBeaconBlockElectra) {
				blk.Block.Body.SilaRequests.Deposits = []*silaenginev1.DepositRequest{nil}
			},
			errMsg: "nil deposit request",
		},
		{
			name: "Nil withdrawal request",
			modifyBlk: func(blk *silapb.SignedBeaconBlockElectra) {
				blk.Block.Body.SilaRequests.Withdrawals = []*silaenginev1.WithdrawalRequest{nil}
			},
			errMsg: "nil withdrawal request",
		},
		{
			name: "Nil consolidation request",
			modifyBlk: func(blk *silapb.SignedBeaconBlockElectra) {
				blk.Block.Body.SilaRequests.Consolidations = []*silaenginev1.ConsolidationRequest{nil}
			},
			errMsg: "nil consolidation request",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st, ks := util.DeterministicGenesisStateElectra(t, 128)
			blk, err := util.GenerateFullBlockElectra(st, ks, util.DefaultBlockGenConfig(), 1)
			require.NoError(t, err)

			tc.modifyBlk(blk)

			b, err := blocks.NewSignedBeaconBlock(blk)
			require.NoError(t, err)

			require.NoError(t, st.SetSlot(1))

			_, err = transition.ElectraOperations(t.Context(), st, b.Block())
			require.ErrorContains(t, tc.errMsg, err)
		})
	}
}

func TestElectraOperations_ProcessingErrors(t *testing.T) {
	tests := []struct {
		name      string
		modifyBlk func(blk *silapb.SignedBeaconBlockElectra)
		errCheck  func(t *testing.T, err error)
	}{
		{
			name: "ErrProcessProposerSlashingsFailed",
			modifyBlk: func(blk *silapb.SignedBeaconBlockElectra) {
				// Create invalid proposer slashing with out-of-bounds proposer index
				blk.Block.Body.ProposerSlashings = []*silapb.ProposerSlashing{
					{
						Header_1: &silapb.SignedBeaconBlockHeader{
							Header: &silapb.BeaconBlockHeader{
								Slot:          1,
								ProposerIndex: 999999, // Invalid index (out of bounds)
								ParentRoot:    make([]byte, 32),
								StateRoot:     make([]byte, 32),
								BodyRoot:      make([]byte, 32),
							},
							Signature: make([]byte, 96),
						},
						Header_2: &silapb.SignedBeaconBlockHeader{
							Header: &silapb.BeaconBlockHeader{
								Slot:          1,
								ProposerIndex: 999999,
								ParentRoot:    make([]byte, 32),
								StateRoot:     make([]byte, 32),
								BodyRoot:      make([]byte, 32),
							},
							Signature: make([]byte, 96),
						},
					},
				}
			},
			errCheck: func(t *testing.T, err error) {
				require.ErrorContains(t, "process proposer slashings failed", err)
				require.Equal(t, true, errors.Is(err, transition.ErrProcessProposerSlashingsFailed))
			},
		},
		{
			name: "ErrProcessAttestationsFailed",
			modifyBlk: func(blk *silapb.SignedBeaconBlockElectra) {
				// Create attestation with invalid committee index
				blk.Block.Body.Attestations = []*silapb.AttestationElectra{
					{
						AggregationBits: []byte{0b00000001},
						Data: &silapb.AttestationData{
							Slot:            1,
							CommitteeIndex:  999999, // Invalid committee index
							BeaconBlockRoot: make([]byte, 32),
							Source: &silapb.Checkpoint{
								Epoch: 0,
								Root:  make([]byte, 32),
							},
							Target: &silapb.Checkpoint{
								Epoch: 0,
								Root:  make([]byte, 32),
							},
						},
						CommitteeBits: []byte{0b00000001},
						Signature:     make([]byte, 96),
					},
				}
			},
			errCheck: func(t *testing.T, err error) {
				require.ErrorContains(t, "process attestations failed", err)
				require.Equal(t, true, errors.Is(err, transition.ErrProcessAttestationsFailed))
			},
		},
		{
			name: "ErrProcessDepositsFailed",
			modifyBlk: func(blk *silapb.SignedBeaconBlockElectra) {
				// Create deposit with invalid proof length
				blk.Block.Body.Deposits = []*silapb.Deposit{
					{
						Proof: [][]byte{}, // Invalid: empty proof
						Data: &silapb.Deposit_Data{
							PublicKey:             make([]byte, 48),
							WithdrawalCredentials: make([]byte, 32),
							Amount:                32000000000, // 32 ETH in Gwei
							Signature:             make([]byte, 96),
						},
					},
				}
			},
			errCheck: func(t *testing.T, err error) {
				require.ErrorContains(t, "process deposits failed", err)
				require.Equal(t, true, errors.Is(err, transition.ErrProcessDepositsFailed))
			},
		},
		{
			name: "ErrProcessVoluntaryExitsFailed",
			modifyBlk: func(blk *silapb.SignedBeaconBlockElectra) {
				// Create voluntary exit with invalid validator index
				blk.Block.Body.VoluntaryExits = []*silapb.SignedVoluntaryExit{
					{
						Exit: &silapb.VoluntaryExit{
							Epoch:          0,
							ValidatorIndex: 999999, // Invalid index (out of bounds)
						},
						Signature: make([]byte, 96),
					},
				}
			},
			errCheck: func(t *testing.T, err error) {
				require.ErrorContains(t, "process voluntary exits failed", err)
				require.Equal(t, true, errors.Is(err, transition.ErrProcessVoluntaryExitsFailed))
			},
		},
		{
			name: "ErrProcessBLSChangesFailed",
			modifyBlk: func(blk *silapb.SignedBeaconBlockElectra) {
				// Create BLS to Sila change with invalid validator index
				blk.Block.Body.BlsToSilaChanges = []*silapb.SignedBLSToSilaChange{
					{
						Message: &silapb.BLSToSilaChange{
							ValidatorIndex:     999999, // Invalid index (out of bounds)
							FromBlsPubkey:      make([]byte, 48),
							ToSilaAddress: make([]byte, 20),
						},
						Signature: make([]byte, 96),
					},
				}
			},
			errCheck: func(t *testing.T, err error) {
				require.ErrorContains(t, "process BLS to Sila changes failed", err)
				require.Equal(t, true, errors.Is(err, transition.ErrProcessBLSChangesFailed))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			st, ks := util.DeterministicGenesisStateElectra(t, 128)
			blk, err := util.GenerateFullBlockElectra(st, ks, util.DefaultBlockGenConfig(), 1)
			require.NoError(t, err)

			tc.modifyBlk(blk)

			b, err := blocks.NewSignedBeaconBlock(blk)
			require.NoError(t, err)

			require.NoError(t, st.SetSlot(primitives.Slot(1)))

			_, err = transition.ElectraOperations(ctx, st, b.Block())
			require.NotNil(t, err, "Expected an error but got nil")
			tc.errCheck(t, err)
		})
	}
}
