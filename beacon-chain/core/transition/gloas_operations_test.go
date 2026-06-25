package transition_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sila-chain/go-bitfield"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/transition"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func newGloasBlock(t *testing.T, body *silapb.BeaconBlockBodyGloas) interfaces.ReadOnlyBeaconBlock {
	t.Helper()
	hydrated := util.HydrateSignedBeaconBlockGloas(&silapb.SignedBeaconBlockGloas{
		Block: &silapb.BeaconBlockGloas{Body: body},
	})
	signed, err := blocks.NewSignedBeaconBlock(hydrated)
	require.NoError(t, err)
	return signed.Block()
}

func emptyGloasBody() *silapb.BeaconBlockBodyGloas {
	return util.HydrateBeaconBlockBodyGloas(nil)
}

func TestGloasOperations_HappyPath(t *testing.T) {
	st, _ := util.DeterministicGenesisStateElectra(t, 16)
	// A plain Electra state is fine here because we exercise zero operations.
	blk := newGloasBlock(t, emptyGloasBody())

	_, err := transition.GloasOperations(context.Background(), st, blk)
	require.NoError(t, err)
}

// TestGloasOperations_ProcessingErrors covers every sentinel error the
// function can return, one sub-test per operation step.
func TestGloasOperations_ProcessingErrors(t *testing.T) {
	tests := []struct {
		name        string
		modifyBlk   func(*silapb.BeaconBlockBodyGloas)
		errSentinel error
		errSubstr   string
	}{
		{
			name: "ErrProcessProposerSlashingsFailed – out-of-bounds proposer index",
			modifyBlk: func(b *silapb.BeaconBlockBodyGloas) {
				b.ProposerSlashings = []*silapb.ProposerSlashing{
					{
						Header_1: &silapb.SignedBeaconBlockHeader{
							Header: &silapb.BeaconBlockHeader{
								Slot:          1,
								ProposerIndex: 999999,
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
			errSentinel: transition.ErrProcessProposerSlashingsFailed,
			errSubstr:   "process proposer slashings failed",
		},
		{
			name: "ErrProcessAttesterSlashingsFailed – out-of-bounds attesting index",
			modifyBlk: func(b *silapb.BeaconBlockBodyGloas) {
				makeIndexed := func(root []byte) *silapb.IndexedAttestationElectra {
					return &silapb.IndexedAttestationElectra{
						AttestingIndices: []uint64{999999},
						Data: &silapb.AttestationData{
							Slot:            1,
							CommitteeIndex:  0,
							BeaconBlockRoot: root,
							Source:          &silapb.Checkpoint{Root: make([]byte, 32)},
							Target:          &silapb.Checkpoint{Root: make([]byte, 32)},
						},
						Signature: make([]byte, 96),
					}
				}
				root1 := make([]byte, 32)
				root2 := make([]byte, 32)
				root2[0] = 0xff // different roots → slashable
				b.AttesterSlashings = []*silapb.AttesterSlashingElectra{
					{
						Attestation_1: makeIndexed(root1),
						Attestation_2: makeIndexed(root2),
					},
				}
			},
			errSentinel: transition.ErrProcessAttesterSlashingsFailed,
			errSubstr:   "process attester slashings failed",
		},

		{
			name: "ErrProcessAttestationsFailed – invalid committee index",
			modifyBlk: func(b *silapb.BeaconBlockBodyGloas) {
				b.Attestations = []*silapb.AttestationElectra{
					{
						AggregationBits: []byte{0b00000001},
						Data: &silapb.AttestationData{
							Slot:            1,
							CommitteeIndex:  999999, // no such committee
							BeaconBlockRoot: make([]byte, 32),
							Source:          &silapb.Checkpoint{Root: make([]byte, 32)},
							Target:          &silapb.Checkpoint{Root: make([]byte, 32)},
						},
						CommitteeBits: []byte{0b00000001},
						Signature:     make([]byte, 96),
					},
				}
			},
			errSentinel: transition.ErrProcessAttestationsFailed,
			errSubstr:   "process attestations failed",
		},

		{
			name: "ErrProcessDepositsFailed – empty merkle proof",
			modifyBlk: func(b *silapb.BeaconBlockBodyGloas) {
				b.Deposits = []*silapb.Deposit{
					{
						Proof: [][]byte{}, // invalid: proof must not be empty
						Data: &silapb.Deposit_Data{
							PublicKey:             make([]byte, 48),
							WithdrawalCredentials: make([]byte, 32),
							Amount:                32_000_000_000,
							Signature:             make([]byte, 96),
						},
					},
				}
			},
			errSentinel: transition.ErrProcessDepositsFailed,
			errSubstr:   "process deposits failed",
		},

		{
			name: "ErrProcessVoluntaryExitsFailed – out-of-bounds validator index",
			modifyBlk: func(b *silapb.BeaconBlockBodyGloas) {
				b.VoluntaryExits = []*silapb.SignedVoluntaryExit{
					{
						Exit: &silapb.VoluntaryExit{
							Epoch:          0,
							ValidatorIndex: 999999,
						},
						Signature: make([]byte, 96),
					},
				}
			},
			errSentinel: transition.ErrProcessVoluntaryExitsFailed,
			errSubstr:   "process voluntary exits failed",
		},

		{
			name: "ErrProcessBLSChangesFailed – out-of-bounds validator index",
			modifyBlk: func(b *silapb.BeaconBlockBodyGloas) {
				b.BlsToSilaChanges = []*silapb.SignedBLSToSilaChange{
					{
						Message: &silapb.BLSToSilaChange{
							ValidatorIndex:     999999,
							FromBlsPubkey:      make([]byte, 48),
							ToSilaAddress: make([]byte, 20),
						},
						Signature: make([]byte, 96),
					},
				}
			},
			errSentinel: transition.ErrProcessBLSChangesFailed,
			errSubstr:   "process BLS to Sila changes failed",
		},

		{
			name: "ErrProcessPayloadAttestationsFailed – wrong beacon block root",
			modifyBlk: func(b *silapb.BeaconBlockBodyGloas) {
				b.PayloadAttestations = []*silapb.PayloadAttestation{
					{
						AggregationBits: bitfield.NewBitvector512(),
						Data: &silapb.PayloadAttestationData{
							BeaconBlockRoot: make([]byte, 32), // all-zeros ≠ header.parent_root
							Slot:            0,
						},
						Signature: make([]byte, 96),
					},
				}
			},
			errSentinel: transition.ErrProcessPayloadAttestationsFailed,
			errSubstr:   "process payload attestations failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			st, _ := util.DeterministicGenesisStateElectra(t, 128)

			// For the payload-attestation sub-test we need the state's latest block
			// header to have a non-zero parent root so the all-zeros root in the
			// attestation definitely mismatches.
			if tc.errSentinel == transition.ErrProcessPayloadAttestationsFailed {
				hdr := &silapb.BeaconBlockHeader{
					ParentRoot: make([]byte, 32),
					StateRoot:  make([]byte, 32),
					BodyRoot:   make([]byte, 32),
				}
				hdr.ParentRoot[0] = 0xde
				require.NoError(t, st.SetLatestBlockHeader(hdr))
			}

			body := emptyGloasBody()
			tc.modifyBlk(body)

			gloasBlk := newGloasBlock(t, body)

			_, err := transition.GloasOperations(ctx, st, gloasBlk)
			require.NotNil(t, err, "expected an error but got nil")
			require.ErrorContains(t, tc.errSubstr, err)
			require.Equal(t, true, errors.Is(err, tc.errSentinel))
		})
	}
}
