package test_helpers

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

func GenerateProtoAltairBeaconBlock() *silapb.BeaconBlockAltair {
	return &silapb.BeaconBlockAltair{
		Slot:          1,
		ProposerIndex: 2,
		ParentRoot:    FillByteSlice(32, 3),
		StateRoot:     FillByteSlice(32, 4),
		Body: &silapb.BeaconBlockBodyAltair{
			RandaoReveal: FillByteSlice(96, 5),
			SilaData: &silapb.SilaData{
				DepositRoot:  FillByteSlice(32, 6),
				DepositCount: 7,
				BlockHash:    FillByteSlice(32, 8),
			},
			Graffiti: FillByteSlice(32, 9),
			ProposerSlashings: []*silapb.ProposerSlashing{
				{
					Header_1: &silapb.SignedBeaconBlockHeader{
						Header: &silapb.BeaconBlockHeader{
							Slot:          10,
							ProposerIndex: 11,
							ParentRoot:    FillByteSlice(32, 12),
							StateRoot:     FillByteSlice(32, 13),
							BodyRoot:      FillByteSlice(32, 14),
						},
						Signature: FillByteSlice(96, 15),
					},
					Header_2: &silapb.SignedBeaconBlockHeader{
						Header: &silapb.BeaconBlockHeader{
							Slot:          16,
							ProposerIndex: 17,
							ParentRoot:    FillByteSlice(32, 18),
							StateRoot:     FillByteSlice(32, 19),
							BodyRoot:      FillByteSlice(32, 20),
						},
						Signature: FillByteSlice(96, 21),
					},
				},
				{
					Header_1: &silapb.SignedBeaconBlockHeader{
						Header: &silapb.BeaconBlockHeader{
							Slot:          22,
							ProposerIndex: 23,
							ParentRoot:    FillByteSlice(32, 24),
							StateRoot:     FillByteSlice(32, 25),
							BodyRoot:      FillByteSlice(32, 26),
						},
						Signature: FillByteSlice(96, 27),
					},
					Header_2: &silapb.SignedBeaconBlockHeader{
						Header: &silapb.BeaconBlockHeader{
							Slot:          28,
							ProposerIndex: 29,
							ParentRoot:    FillByteSlice(32, 30),
							StateRoot:     FillByteSlice(32, 31),
							BodyRoot:      FillByteSlice(32, 32),
						},
						Signature: FillByteSlice(96, 33),
					},
				},
			},
			AttesterSlashings: []*silapb.AttesterSlashing{
				{
					Attestation_1: &silapb.IndexedAttestation{
						AttestingIndices: []uint64{34, 35},
						Data: &silapb.AttestationData{
							Slot:            36,
							CommitteeIndex:  37,
							BeaconBlockRoot: FillByteSlice(32, 38),
							Source: &silapb.Checkpoint{
								Epoch: 39,
								Root:  FillByteSlice(32, 40),
							},
							Target: &silapb.Checkpoint{
								Epoch: 41,
								Root:  FillByteSlice(32, 42),
							},
						},
						Signature: FillByteSlice(96, 43),
					},
					Attestation_2: &silapb.IndexedAttestation{
						AttestingIndices: []uint64{44, 45},
						Data: &silapb.AttestationData{
							Slot:            46,
							CommitteeIndex:  47,
							BeaconBlockRoot: FillByteSlice(32, 38),
							Source: &silapb.Checkpoint{
								Epoch: 49,
								Root:  FillByteSlice(32, 50),
							},
							Target: &silapb.Checkpoint{
								Epoch: 51,
								Root:  FillByteSlice(32, 52),
							},
						},
						Signature: FillByteSlice(96, 53),
					},
				},
				{
					Attestation_1: &silapb.IndexedAttestation{
						AttestingIndices: []uint64{54, 55},
						Data: &silapb.AttestationData{
							Slot:            56,
							CommitteeIndex:  57,
							BeaconBlockRoot: FillByteSlice(32, 38),
							Source: &silapb.Checkpoint{
								Epoch: 59,
								Root:  FillByteSlice(32, 60),
							},
							Target: &silapb.Checkpoint{
								Epoch: 61,
								Root:  FillByteSlice(32, 62),
							},
						},
						Signature: FillByteSlice(96, 63),
					},
					Attestation_2: &silapb.IndexedAttestation{
						AttestingIndices: []uint64{64, 65},
						Data: &silapb.AttestationData{
							Slot:            66,
							CommitteeIndex:  67,
							BeaconBlockRoot: FillByteSlice(32, 38),
							Source: &silapb.Checkpoint{
								Epoch: 69,
								Root:  FillByteSlice(32, 70),
							},
							Target: &silapb.Checkpoint{
								Epoch: 71,
								Root:  FillByteSlice(32, 72),
							},
						},
						Signature: FillByteSlice(96, 73),
					},
				},
			},
			Attestations: []*silapb.Attestation{
				{
					AggregationBits: FillByteSlice(4, 74),
					Data: &silapb.AttestationData{
						Slot:            75,
						CommitteeIndex:  76,
						BeaconBlockRoot: FillByteSlice(32, 38),
						Source: &silapb.Checkpoint{
							Epoch: 78,
							Root:  FillByteSlice(32, 79),
						},
						Target: &silapb.Checkpoint{
							Epoch: 80,
							Root:  FillByteSlice(32, 81),
						},
					},
					Signature: FillByteSlice(96, 82),
				},
				{
					AggregationBits: FillByteSlice(4, 83),
					Data: &silapb.AttestationData{
						Slot:            84,
						CommitteeIndex:  85,
						BeaconBlockRoot: FillByteSlice(32, 38),
						Source: &silapb.Checkpoint{
							Epoch: 87,
							Root:  FillByteSlice(32, 88),
						},
						Target: &silapb.Checkpoint{
							Epoch: 89,
							Root:  FillByteSlice(32, 90),
						},
					},
					Signature: FillByteSlice(96, 91),
				},
			},
			Deposits: []*silapb.Deposit{
				{
					Proof: FillByteArraySlice(33, FillByteSlice(32, 92)),
					Data: &silapb.Deposit_Data{
						PublicKey:             FillByteSlice(48, 94),
						WithdrawalCredentials: FillByteSlice(32, 95),
						Amount:                96,
						Signature:             FillByteSlice(96, 97),
					},
				},
				{
					Proof: FillByteArraySlice(33, FillByteSlice(32, 98)),
					Data: &silapb.Deposit_Data{
						PublicKey:             FillByteSlice(48, 100),
						WithdrawalCredentials: FillByteSlice(32, 101),
						Amount:                102,
						Signature:             FillByteSlice(96, 103),
					},
				},
			},
			VoluntaryExits: []*silapb.SignedVoluntaryExit{
				{
					Exit: &silapb.VoluntaryExit{
						Epoch:          104,
						ValidatorIndex: 105,
					},
					Signature: FillByteSlice(96, 106),
				},
				{
					Exit: &silapb.VoluntaryExit{
						Epoch:          107,
						ValidatorIndex: 108,
					},
					Signature: FillByteSlice(96, 109),
				},
			},
			SyncAggregate: &silapb.SyncAggregate{
				SyncCommitteeBits:      FillByteSlice(64, 110),
				SyncCommitteeSignature: FillByteSlice(96, 111),
			},
		},
	}
}

func GenerateJsonAltairBeaconBlock() *structs.BeaconBlockAltair {
	return &structs.BeaconBlockAltair{
		Slot:          "1",
		ProposerIndex: "2",
		ParentRoot:    FillEncodedByteSlice(32, 3),
		StateRoot:     FillEncodedByteSlice(32, 4),
		Body: &structs.BeaconBlockBodyAltair{
			RandaoReveal: FillEncodedByteSlice(96, 5),
			SilaData: &structs.SilaData{
				DepositRoot:  FillEncodedByteSlice(32, 6),
				DepositCount: "7",
				BlockHash:    FillEncodedByteSlice(32, 8),
			},
			Graffiti: FillEncodedByteSlice(32, 9),
			ProposerSlashings: []*structs.ProposerSlashing{
				{
					SignedHeader1: &structs.SignedBeaconBlockHeader{
						Message: &structs.BeaconBlockHeader{
							Slot:          "10",
							ProposerIndex: "11",
							ParentRoot:    FillEncodedByteSlice(32, 12),
							StateRoot:     FillEncodedByteSlice(32, 13),
							BodyRoot:      FillEncodedByteSlice(32, 14),
						},
						Signature: FillEncodedByteSlice(96, 15),
					},
					SignedHeader2: &structs.SignedBeaconBlockHeader{
						Message: &structs.BeaconBlockHeader{
							Slot:          "16",
							ProposerIndex: "17",
							ParentRoot:    FillEncodedByteSlice(32, 18),
							StateRoot:     FillEncodedByteSlice(32, 19),
							BodyRoot:      FillEncodedByteSlice(32, 20),
						},
						Signature: FillEncodedByteSlice(96, 21),
					},
				},
				{
					SignedHeader1: &structs.SignedBeaconBlockHeader{
						Message: &structs.BeaconBlockHeader{
							Slot:          "22",
							ProposerIndex: "23",
							ParentRoot:    FillEncodedByteSlice(32, 24),
							StateRoot:     FillEncodedByteSlice(32, 25),
							BodyRoot:      FillEncodedByteSlice(32, 26),
						},
						Signature: FillEncodedByteSlice(96, 27),
					},
					SignedHeader2: &structs.SignedBeaconBlockHeader{
						Message: &structs.BeaconBlockHeader{
							Slot:          "28",
							ProposerIndex: "29",
							ParentRoot:    FillEncodedByteSlice(32, 30),
							StateRoot:     FillEncodedByteSlice(32, 31),
							BodyRoot:      FillEncodedByteSlice(32, 32),
						},
						Signature: FillEncodedByteSlice(96, 33),
					},
				},
			},
			AttesterSlashings: []*structs.AttesterSlashing{
				{
					Attestation1: &structs.IndexedAttestation{
						AttestingIndices: []string{"34", "35"},
						Data: &structs.AttestationData{
							Slot:            "36",
							CommitteeIndex:  "37",
							BeaconBlockRoot: FillEncodedByteSlice(32, 38),
							Source: &structs.Checkpoint{
								Epoch: "39",
								Root:  FillEncodedByteSlice(32, 40),
							},
							Target: &structs.Checkpoint{
								Epoch: "41",
								Root:  FillEncodedByteSlice(32, 42),
							},
						},
						Signature: FillEncodedByteSlice(96, 43),
					},
					Attestation2: &structs.IndexedAttestation{
						AttestingIndices: []string{"44", "45"},
						Data: &structs.AttestationData{
							Slot:            "46",
							CommitteeIndex:  "47",
							BeaconBlockRoot: FillEncodedByteSlice(32, 38),
							Source: &structs.Checkpoint{
								Epoch: "49",
								Root:  FillEncodedByteSlice(32, 50),
							},
							Target: &structs.Checkpoint{
								Epoch: "51",
								Root:  FillEncodedByteSlice(32, 52),
							},
						},
						Signature: FillEncodedByteSlice(96, 53),
					},
				},
				{
					Attestation1: &structs.IndexedAttestation{
						AttestingIndices: []string{"54", "55"},
						Data: &structs.AttestationData{
							Slot:            "56",
							CommitteeIndex:  "57",
							BeaconBlockRoot: FillEncodedByteSlice(32, 38),
							Source: &structs.Checkpoint{
								Epoch: "59",
								Root:  FillEncodedByteSlice(32, 60),
							},
							Target: &structs.Checkpoint{
								Epoch: "61",
								Root:  FillEncodedByteSlice(32, 62),
							},
						},
						Signature: FillEncodedByteSlice(96, 63),
					},
					Attestation2: &structs.IndexedAttestation{
						AttestingIndices: []string{"64", "65"},
						Data: &structs.AttestationData{
							Slot:            "66",
							CommitteeIndex:  "67",
							BeaconBlockRoot: FillEncodedByteSlice(32, 38),
							Source: &structs.Checkpoint{
								Epoch: "69",
								Root:  FillEncodedByteSlice(32, 70),
							},
							Target: &structs.Checkpoint{
								Epoch: "71",
								Root:  FillEncodedByteSlice(32, 72),
							},
						},
						Signature: FillEncodedByteSlice(96, 73),
					},
				},
			},
			Attestations: []*structs.Attestation{
				{
					AggregationBits: FillEncodedByteSlice(4, 74),
					Data: &structs.AttestationData{
						Slot:            "75",
						CommitteeIndex:  "76",
						BeaconBlockRoot: FillEncodedByteSlice(32, 38),
						Source: &structs.Checkpoint{
							Epoch: "78",
							Root:  FillEncodedByteSlice(32, 79),
						},
						Target: &structs.Checkpoint{
							Epoch: "80",
							Root:  FillEncodedByteSlice(32, 81),
						},
					},
					Signature: FillEncodedByteSlice(96, 82),
				},
				{
					AggregationBits: FillEncodedByteSlice(4, 83),
					Data: &structs.AttestationData{
						Slot:            "84",
						CommitteeIndex:  "85",
						BeaconBlockRoot: FillEncodedByteSlice(32, 38),
						Source: &structs.Checkpoint{
							Epoch: "87",
							Root:  FillEncodedByteSlice(32, 88),
						},
						Target: &structs.Checkpoint{
							Epoch: "89",
							Root:  FillEncodedByteSlice(32, 90),
						},
					},
					Signature: FillEncodedByteSlice(96, 91),
				},
			},
			Deposits: []*structs.Deposit{
				{
					Proof: FillEncodedByteArraySlice(33, FillEncodedByteSlice(32, 92)),
					Data: &structs.DepositData{
						Pubkey:                FillEncodedByteSlice(48, 94),
						WithdrawalCredentials: FillEncodedByteSlice(32, 95),
						Amount:                "96",
						Signature:             FillEncodedByteSlice(96, 97),
					},
				},
				{
					Proof: FillEncodedByteArraySlice(33, FillEncodedByteSlice(32, 98)),
					Data: &structs.DepositData{
						Pubkey:                FillEncodedByteSlice(48, 100),
						WithdrawalCredentials: FillEncodedByteSlice(32, 101),
						Amount:                "102",
						Signature:             FillEncodedByteSlice(96, 103),
					},
				},
			},
			VoluntaryExits: []*structs.SignedVoluntaryExit{
				{
					Message: &structs.VoluntaryExit{
						Epoch:          "104",
						ValidatorIndex: "105",
					},
					Signature: FillEncodedByteSlice(96, 106),
				},
				{
					Message: &structs.VoluntaryExit{
						Epoch:          "107",
						ValidatorIndex: "108",
					},
					Signature: FillEncodedByteSlice(96, 109),
				},
			},
			SyncAggregate: &structs.SyncAggregate{
				SyncCommitteeBits:      FillEncodedByteSlice(64, 110),
				SyncCommitteeSignature: FillEncodedByteSlice(96, 111),
			},
		},
	}
}
