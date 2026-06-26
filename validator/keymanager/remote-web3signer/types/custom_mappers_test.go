package types_test

import (
	"reflect"
	"testing"

	"github.com/sila-chain/go-bitfield"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/keymanager/remote-web3signer/types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/keymanager/remote-web3signer/types/mock"
)

func TestMapAggregateAndProof(t *testing.T) {
	type args struct {
		from *silapb.AggregateAttestationAndProof
	}
	tests := []struct {
		name    string
		args    args
		want    *types.AggregateAndProof
		wantErr bool
	}{
		{
			name: "HappyPathTest",
			args: args{
				from: &silapb.AggregateAttestationAndProof{
					AggregatorIndex: 0,
					Aggregate: &silapb.Attestation{
						AggregationBits: bitfield.Bitlist{0b1101},
						Data: &silapb.AttestationData{
							BeaconBlockRoot: make([]byte, fieldparams.RootLength),
							Source: &silapb.Checkpoint{
								Root: make([]byte, fieldparams.RootLength),
							},
							Target: &silapb.Checkpoint{
								Root: make([]byte, fieldparams.RootLength),
							},
						},
						Signature: make([]byte, 96),
					},
					SelectionProof: make([]byte, fieldparams.BLSSignatureLength),
				},
			},
			want: &types.AggregateAndProof{
				AggregatorIndex: "0",
				Aggregate:       mock.Attestation(),
				SelectionProof:  make([]byte, fieldparams.BLSSignatureLength),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.MapAggregateAndProof(tt.args.from)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapAggregateAndProof() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Aggregate, tt.want.Aggregate) {
				t.Errorf("MapAggregateAndProof() got = %v, want %v", got.Aggregate, tt.want.Aggregate)
			}
		})
	}
}

func TestMapAggregateAndProofElectra(t *testing.T) {
	type args struct {
		from *silapb.AggregateAttestationAndProofElectra
	}
	tests := []struct {
		name    string
		args    args
		want    *types.AggregateAndProofElectra
		wantErr bool
	}{
		{
			name: "HappyPathTest",
			args: args{
				from: &silapb.AggregateAttestationAndProofElectra{
					AggregatorIndex: 0,
					Aggregate: &silapb.AttestationElectra{
						AggregationBits: bitfield.Bitlist{0b1101},
						Data: &silapb.AttestationData{
							BeaconBlockRoot: make([]byte, fieldparams.RootLength),
							Source: &silapb.Checkpoint{
								Root: make([]byte, fieldparams.RootLength),
							},
							Target: &silapb.Checkpoint{
								Root: make([]byte, fieldparams.RootLength),
							},
						},
						Signature: make([]byte, 96),
						CommitteeBits: func() bitfield.Bitvector64 {
							committeeBits := bitfield.NewBitvector64()
							committeeBits.SetBitAt(0, true)
							return committeeBits
						}(),
					},
					SelectionProof: make([]byte, fieldparams.BLSSignatureLength),
				},
			},
			want: &types.AggregateAndProofElectra{
				AggregatorIndex: "0",
				Aggregate:       mock.AttestationElectra(),
				SelectionProof:  make([]byte, fieldparams.BLSSignatureLength),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.MapAggregateAndProofElectra(tt.args.from)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapAggregateAndProof() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Aggregate, tt.want.Aggregate) {
				t.Errorf("MapAggregateAndProof() got = %v, want %v", got.Aggregate, tt.want.Aggregate)
			}
		})
	}
}

func TestMapAttestation(t *testing.T) {
	type args struct {
		attestation *silapb.Attestation
	}
	tests := []struct {
		name    string
		args    args
		want    *types.Attestation
		wantErr bool
	}{
		{
			name: "HappyPathTest",
			args: args{
				attestation: &silapb.Attestation{
					AggregationBits: bitfield.Bitlist{0b1101},
					Data: &silapb.AttestationData{
						BeaconBlockRoot: make([]byte, fieldparams.RootLength),
						Source: &silapb.Checkpoint{
							Root: make([]byte, fieldparams.RootLength),
						},
						Target: &silapb.Checkpoint{
							Root: make([]byte, fieldparams.RootLength),
						},
					},
					Signature: make([]byte, 96),
				},
			},
			want:    mock.Attestation(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.MapAttestation(tt.args.attestation)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapAttestation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapAttestation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapAttestationElectra(t *testing.T) {
	type args struct {
		attestation *silapb.AttestationElectra
	}
	tests := []struct {
		name    string
		args    args
		want    *types.AttestationElectra
		wantErr bool
	}{
		{
			name: "HappyPathTest",
			args: args{
				attestation: &silapb.AttestationElectra{
					AggregationBits: bitfield.Bitlist{0b1101},
					Data: &silapb.AttestationData{
						BeaconBlockRoot: make([]byte, fieldparams.RootLength),
						Source: &silapb.Checkpoint{
							Root: make([]byte, fieldparams.RootLength),
						},
						Target: &silapb.Checkpoint{
							Root: make([]byte, fieldparams.RootLength),
						},
					},
					CommitteeBits: func() bitfield.Bitvector64 {
						committeeBits := bitfield.NewBitvector64()
						committeeBits.SetBitAt(0, true)
						return committeeBits
					}(),
					Signature: make([]byte, 96),
				},
			},
			want:    mock.AttestationElectra(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.MapAttestationElectra(tt.args.attestation)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapAttestation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapAttestation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapAttestationData(t *testing.T) {
	type args struct {
		data *silapb.AttestationData
	}
	tests := []struct {
		name    string
		args    args
		want    *types.AttestationData
		wantErr bool
	}{
		{
			name: "HappyPathTest",
			args: args{
				data: &silapb.AttestationData{
					BeaconBlockRoot: make([]byte, fieldparams.RootLength),
					Source: &silapb.Checkpoint{
						Root: make([]byte, fieldparams.RootLength),
					},
					Target: &silapb.Checkpoint{
						Root: make([]byte, fieldparams.RootLength),
					},
				},
			},
			want:    mock.Attestation().Data,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.MapAttestationData(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapAttestationData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapAttestationData() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapAttesterSlashing(t *testing.T) {
	type args struct {
		slashing *silapb.AttesterSlashing
	}
	tests := []struct {
		name    string
		args    args
		want    *types.AttesterSlashing
		wantErr bool
	}{
		{
			name: "HappyPathTest",
			args: args{
				slashing: &silapb.AttesterSlashing{
					Attestation_1: &silapb.IndexedAttestation{
						AttestingIndices: []uint64{0, 1, 2},
						Data: &silapb.AttestationData{
							BeaconBlockRoot: make([]byte, fieldparams.RootLength),
							Source: &silapb.Checkpoint{
								Root: make([]byte, fieldparams.RootLength),
							},
							Target: &silapb.Checkpoint{
								Root: make([]byte, fieldparams.RootLength),
							},
						},
						Signature: make([]byte, fieldparams.BLSSignatureLength),
					},
					Attestation_2: &silapb.IndexedAttestation{
						AttestingIndices: []uint64{0, 1, 2},
						Data: &silapb.AttestationData{
							BeaconBlockRoot: make([]byte, fieldparams.RootLength),
							Source: &silapb.Checkpoint{
								Root: make([]byte, fieldparams.RootLength),
							},
							Target: &silapb.Checkpoint{
								Root: make([]byte, fieldparams.RootLength),
							},
						},
						Signature: make([]byte, fieldparams.BLSSignatureLength),
					},
				},
			},
			want: &types.AttesterSlashing{
				Attestation1: mock.IndexedAttestation(),
				Attestation2: mock.IndexedAttestation(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.MapAttesterSlashing(tt.args.slashing)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapAttesterSlashing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Attestation1, tt.want.Attestation1) {
				t.Errorf("MapAttesterSlashing() got = %v, want %v", got.Attestation1, tt.want.Attestation1)
			}
		})
	}
}

func TestMapBeaconBlockAltair(t *testing.T) {
	type args struct {
		block *silapb.BeaconBlockAltair
	}
	tests := []struct {
		name    string
		args    args
		want    *types.BeaconBlockAltair
		wantErr bool
	}{
		{
			name: "Happy Path Test",
			args: args{
				block: &silapb.BeaconBlockAltair{
					Slot:          0,
					ProposerIndex: 0,
					ParentRoot:    make([]byte, fieldparams.RootLength),
					StateRoot:     make([]byte, fieldparams.RootLength),
					Body: &silapb.BeaconBlockBodyAltair{
						RandaoReveal: make([]byte, 32),
						SilaData: &silapb.SilaData{
							DepositRoot:  make([]byte, fieldparams.RootLength),
							DepositCount: 0,
							BlockHash:    make([]byte, 32),
						},
						Graffiti: make([]byte, 32),
						ProposerSlashings: []*silapb.ProposerSlashing{
							{
								Header_1: &silapb.SignedBeaconBlockHeader{
									Header: &silapb.BeaconBlockHeader{
										Slot:          0,
										ProposerIndex: 0,
										ParentRoot:    make([]byte, fieldparams.RootLength),
										StateRoot:     make([]byte, fieldparams.RootLength),
										BodyRoot:      make([]byte, fieldparams.RootLength),
									},
									Signature: make([]byte, fieldparams.BLSSignatureLength),
								},
								Header_2: &silapb.SignedBeaconBlockHeader{
									Header: &silapb.BeaconBlockHeader{
										Slot:          0,
										ProposerIndex: 0,
										ParentRoot:    make([]byte, fieldparams.RootLength),
										StateRoot:     make([]byte, fieldparams.RootLength),
										BodyRoot:      make([]byte, fieldparams.RootLength),
									},
									Signature: make([]byte, fieldparams.BLSSignatureLength),
								},
							},
						},
						AttesterSlashings: []*silapb.AttesterSlashing{
							{
								Attestation_1: &silapb.IndexedAttestation{
									AttestingIndices: []uint64{0, 1, 2},
									Data: &silapb.AttestationData{
										BeaconBlockRoot: make([]byte, fieldparams.RootLength),
										Source: &silapb.Checkpoint{
											Root: make([]byte, fieldparams.RootLength),
										},
										Target: &silapb.Checkpoint{
											Root: make([]byte, fieldparams.RootLength),
										},
									},
									Signature: make([]byte, fieldparams.BLSSignatureLength),
								},
								Attestation_2: &silapb.IndexedAttestation{
									AttestingIndices: []uint64{0, 1, 2},
									Data: &silapb.AttestationData{
										BeaconBlockRoot: make([]byte, fieldparams.RootLength),
										Source: &silapb.Checkpoint{
											Root: make([]byte, fieldparams.RootLength),
										},
										Target: &silapb.Checkpoint{
											Root: make([]byte, fieldparams.RootLength),
										},
									},
									Signature: make([]byte, fieldparams.BLSSignatureLength),
								},
							},
						},
						Attestations: []*silapb.Attestation{
							{
								AggregationBits: bitfield.Bitlist{0b1101},
								Data: &silapb.AttestationData{
									BeaconBlockRoot: make([]byte, fieldparams.RootLength),
									Source: &silapb.Checkpoint{
										Root: make([]byte, fieldparams.RootLength),
									},
									Target: &silapb.Checkpoint{
										Root: make([]byte, fieldparams.RootLength),
									},
								},
								Signature: make([]byte, 96),
							},
						},
						Deposits: []*silapb.Deposit{
							{
								Proof: [][]byte{[]byte("A")},
								Data: &silapb.Deposit_Data{
									PublicKey:             make([]byte, fieldparams.BLSPubkeyLength),
									WithdrawalCredentials: make([]byte, 32),
									Amount:                0,
									Signature:             make([]byte, fieldparams.BLSSignatureLength),
								},
							},
						},
						VoluntaryExits: []*silapb.SignedVoluntaryExit{
							{
								Exit: &silapb.VoluntaryExit{
									Epoch:          0,
									ValidatorIndex: 0,
								},
								Signature: make([]byte, fieldparams.BLSSignatureLength),
							},
						},
						SyncAggregate: &silapb.SyncAggregate{
							SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
							SyncCommitteeBits:      mock.SyncComitteeBits(),
						},
					},
				},
			},
			want:    mock.BeaconBlockAltair(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.MapBeaconBlockAltair(tt.args.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapBeaconBlockAltair() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Body, tt.want.Body) {
				t.Errorf("MapBeaconBlockAltair() got = %v, want %v", got.Body.SyncAggregate, tt.want.Body.SyncAggregate)
			}
		})
	}
}

func TestMapBeaconBlockBody(t *testing.T) {
	type args struct {
		body *silapb.BeaconBlockBody
	}
	tests := []struct {
		name    string
		args    args
		want    *types.BeaconBlockBody
		wantErr bool
	}{
		{
			name: "Happy Path Test",
			args: args{
				body: &silapb.BeaconBlockBody{
					RandaoReveal: make([]byte, 32),
					SilaData: &silapb.SilaData{
						DepositRoot:  make([]byte, fieldparams.RootLength),
						DepositCount: 0,
						BlockHash:    make([]byte, 32),
					},
					Graffiti: make([]byte, 32),
					ProposerSlashings: []*silapb.ProposerSlashing{
						{
							Header_1: &silapb.SignedBeaconBlockHeader{
								Header: &silapb.BeaconBlockHeader{
									Slot:          0,
									ProposerIndex: 0,
									ParentRoot:    make([]byte, fieldparams.RootLength),
									StateRoot:     make([]byte, fieldparams.RootLength),
									BodyRoot:      make([]byte, fieldparams.RootLength),
								},
								Signature: make([]byte, fieldparams.BLSSignatureLength),
							},
							Header_2: &silapb.SignedBeaconBlockHeader{
								Header: &silapb.BeaconBlockHeader{
									Slot:          0,
									ProposerIndex: 0,
									ParentRoot:    make([]byte, fieldparams.RootLength),
									StateRoot:     make([]byte, fieldparams.RootLength),
									BodyRoot:      make([]byte, fieldparams.RootLength),
								},
								Signature: make([]byte, fieldparams.BLSSignatureLength),
							},
						},
					},
					AttesterSlashings: []*silapb.AttesterSlashing{
						{
							Attestation_1: &silapb.IndexedAttestation{
								AttestingIndices: []uint64{0, 1, 2},
								Data: &silapb.AttestationData{
									BeaconBlockRoot: make([]byte, fieldparams.RootLength),
									Source: &silapb.Checkpoint{
										Root: make([]byte, fieldparams.RootLength),
									},
									Target: &silapb.Checkpoint{
										Root: make([]byte, fieldparams.RootLength),
									},
								},
								Signature: make([]byte, fieldparams.BLSSignatureLength),
							},
							Attestation_2: &silapb.IndexedAttestation{
								AttestingIndices: []uint64{0, 1, 2},
								Data: &silapb.AttestationData{
									BeaconBlockRoot: make([]byte, fieldparams.RootLength),
									Source: &silapb.Checkpoint{
										Root: make([]byte, fieldparams.RootLength),
									},
									Target: &silapb.Checkpoint{
										Root: make([]byte, fieldparams.RootLength),
									},
								},
								Signature: make([]byte, fieldparams.BLSSignatureLength),
							},
						},
					},
					Attestations: []*silapb.Attestation{
						{
							AggregationBits: bitfield.Bitlist{0b1101},
							Data: &silapb.AttestationData{
								BeaconBlockRoot: make([]byte, fieldparams.RootLength),
								Source: &silapb.Checkpoint{
									Root: make([]byte, fieldparams.RootLength),
								},
								Target: &silapb.Checkpoint{
									Root: make([]byte, fieldparams.RootLength),
								},
							},
							Signature: make([]byte, 96),
						},
					},
					Deposits: []*silapb.Deposit{
						{
							Proof: [][]byte{[]byte("A")},
							Data: &silapb.Deposit_Data{
								PublicKey:             make([]byte, fieldparams.BLSPubkeyLength),
								WithdrawalCredentials: make([]byte, 32),
								Amount:                0,
								Signature:             make([]byte, fieldparams.BLSSignatureLength),
							},
						},
					},
					VoluntaryExits: []*silapb.SignedVoluntaryExit{
						{
							Exit: &silapb.VoluntaryExit{
								Epoch:          0,
								ValidatorIndex: 0,
							},
							Signature: make([]byte, fieldparams.BLSSignatureLength),
						},
					},
				},
			},
			want:    mock.BeaconBlockBody(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.MapBeaconBlockBody(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapBeaconBlockBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapBeaconBlockBody() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapContributionAndProof(t *testing.T) {
	type args struct {
		contribution *silapb.ContributionAndProof
	}
	tests := []struct {
		name    string
		args    args
		want    *types.ContributionAndProof
		wantErr bool
	}{
		{
			name: "Happy Path Test",
			args: args{
				contribution: &silapb.ContributionAndProof{
					AggregatorIndex: 0,
					Contribution: &silapb.SyncCommitteeContribution{
						Slot:              0,
						BlockRoot:         make([]byte, fieldparams.RootLength),
						SubcommitteeIndex: 0,
						AggregationBits:   mock.AggregationBits(),
						Signature:         make([]byte, fieldparams.BLSSignatureLength),
					},
					SelectionProof: make([]byte, fieldparams.BLSSignatureLength),
				},
			},
			want: mock.ContributionAndProof(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.MapContributionAndProof(tt.args.contribution)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapContributionAndProof() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapContributionAndProof() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapForkInfo(t *testing.T) {
	type args struct {
		slot                  primitives.Slot
		genesisValidatorsRoot []byte
	}

	tests := []struct {
		name    string
		args    args
		want    *types.ForkInfo
		wantErr bool
	}{
		{
			name: "Happy Path Test",
			args: args{
				slot:                  0,
				genesisValidatorsRoot: make([]byte, fieldparams.RootLength),
			},
			want:    mock.ForkInfo(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.MapForkInfo(tt.args.slot, tt.args.genesisValidatorsRoot)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapForkInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapForkInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapSyncAggregatorSelectionData(t *testing.T) {
	type args struct {
		data *silapb.SyncAggregatorSelectionData
	}
	tests := []struct {
		name    string
		args    args
		want    *types.SyncAggregatorSelectionData
		wantErr bool
	}{
		{
			name: "Happy Path Test",
			args: args{
				data: &silapb.SyncAggregatorSelectionData{
					Slot:              0,
					SubcommitteeIndex: 0,
				},
			},
			want: &types.SyncAggregatorSelectionData{
				Slot:              "0",
				SubcommitteeIndex: "0",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := types.MapSyncAggregatorSelectionData(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapSyncAggregatorSelectionData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapSyncAggregatorSelectionData() got = %v, want %v", got, tt.want)
			}
		})
	}
}
