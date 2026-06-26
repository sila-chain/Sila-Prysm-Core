package test_helpers

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

func GenerateProtoDenebBeaconBlockContents() *silapb.BeaconBlockContentsDeneb {
	return &silapb.BeaconBlockContentsDeneb{
		Block: &silapb.BeaconBlockDeneb{
			Slot:          1,
			ProposerIndex: 2,
			ParentRoot:    FillByteSlice(32, 3),
			StateRoot:     FillByteSlice(32, 4),
			Body: &silapb.BeaconBlockBodyDeneb{
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
				SilaPayload: &silaenginev1.SilaPayloadDeneb{
					ParentHash:    FillByteSlice(32, 112),
					FeeRecipient:  FillByteSlice(20, 113),
					StateRoot:     FillByteSlice(32, 114),
					ReceiptsRoot:  FillByteSlice(32, 115),
					LogsBloom:     FillByteSlice(256, 116),
					PrevRandao:    FillByteSlice(32, 117),
					BlockNumber:   118,
					GasLimit:      119,
					GasUsed:       120,
					Timestamp:     121,
					ExtraData:     FillByteSlice(32, 122),
					BaseFeePerGas: FillByteSlice(32, 123),
					BlockHash:     FillByteSlice(32, 124),
					Transactions: [][]byte{
						FillByteSlice(32, 125),
						FillByteSlice(32, 126),
					},
					Withdrawals: []*silaenginev1.Withdrawal{
						{
							Index:          127,
							ValidatorIndex: 128,
							Address:        FillByteSlice(20, 129),
							Amount:         130,
						},
						{
							Index:          131,
							ValidatorIndex: 132,
							Address:        FillByteSlice(20, 133),
							Amount:         134,
						},
					},
					BlobGasUsed:   135,
					ExcessBlobGas: 136,
				},
				BlsToSilaChanges: []*silapb.SignedBLSToSilaChange{
					{
						Message: &silapb.BLSToSilaChange{
							ValidatorIndex:     137,
							FromBlsPubkey:      FillByteSlice(48, 138),
							ToSilaAddress: FillByteSlice(20, 139),
						},
						Signature: FillByteSlice(96, 140),
					},
					{
						Message: &silapb.BLSToSilaChange{
							ValidatorIndex:     141,
							FromBlsPubkey:      FillByteSlice(48, 142),
							ToSilaAddress: FillByteSlice(20, 143),
						},
						Signature: FillByteSlice(96, 144),
					},
				},
				BlobKzgCommitments: [][]byte{FillByteSlice(48, 145), FillByteSlice(48, 146)},
			},
		},
		KzgProofs: [][]byte{FillByteSlice(48, 146)},
		Blobs:     [][]byte{FillByteSlice(131072, 147)},
	}
}

func GenerateProtoBlindedDenebBeaconBlock() *silapb.BlindedBeaconBlockDeneb {
	return &silapb.BlindedBeaconBlockDeneb{
		Slot:          1,
		ProposerIndex: 2,
		ParentRoot:    FillByteSlice(32, 3),
		StateRoot:     FillByteSlice(32, 4),
		Body: &silapb.BlindedBeaconBlockBodyDeneb{
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
			SilaPayloadHeader: &silaenginev1.SilaPayloadHeaderDeneb{
				ParentHash:       FillByteSlice(32, 112),
				FeeRecipient:     FillByteSlice(20, 113),
				StateRoot:        FillByteSlice(32, 114),
				ReceiptsRoot:     FillByteSlice(32, 115),
				LogsBloom:        FillByteSlice(256, 116),
				PrevRandao:       FillByteSlice(32, 117),
				BlockNumber:      118,
				GasLimit:         119,
				GasUsed:          120,
				Timestamp:        121,
				ExtraData:        FillByteSlice(32, 122),
				BaseFeePerGas:    FillByteSlice(32, 123),
				BlockHash:        FillByteSlice(32, 124),
				TransactionsRoot: FillByteSlice(32, 125),
				WithdrawalsRoot:  FillByteSlice(32, 126),
				BlobGasUsed:      127,
				ExcessBlobGas:    128,
			},
			BlsToSilaChanges: []*silapb.SignedBLSToSilaChange{
				{
					Message: &silapb.BLSToSilaChange{
						ValidatorIndex:     129,
						FromBlsPubkey:      FillByteSlice(48, 130),
						ToSilaAddress: FillByteSlice(20, 131),
					},
					Signature: FillByteSlice(96, 132),
				},
				{
					Message: &silapb.BLSToSilaChange{
						ValidatorIndex:     133,
						FromBlsPubkey:      FillByteSlice(48, 134),
						ToSilaAddress: FillByteSlice(20, 135),
					},
					Signature: FillByteSlice(96, 136),
				},
			},
			BlobKzgCommitments: [][]byte{FillByteSlice(48, 137), FillByteSlice(48, 138)},
		},
	}
}

func GenerateJsonDenebBeaconBlockContents() *structs.BeaconBlockContentsDeneb {
	return &structs.BeaconBlockContentsDeneb{
		Block: &structs.BeaconBlockDeneb{
			Slot:          "1",
			ProposerIndex: "2",
			ParentRoot:    FillEncodedByteSlice(32, 3),
			StateRoot:     FillEncodedByteSlice(32, 4),
			Body: &structs.BeaconBlockBodyDeneb{
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
				SilaPayload: &structs.SilaPayloadDeneb{
					ParentHash:    FillEncodedByteSlice(32, 112),
					FeeRecipient:  FillEncodedByteSlice(20, 113),
					StateRoot:     FillEncodedByteSlice(32, 114),
					ReceiptsRoot:  FillEncodedByteSlice(32, 115),
					LogsBloom:     FillEncodedByteSlice(256, 116),
					PrevRandao:    FillEncodedByteSlice(32, 117),
					BlockNumber:   "118",
					GasLimit:      "119",
					GasUsed:       "120",
					Timestamp:     "121",
					ExtraData:     FillEncodedByteSlice(32, 122),
					BaseFeePerGas: bytesutil.LittleEndianBytesToBigInt(FillByteSlice(32, 123)).String(),
					BlockHash:     FillEncodedByteSlice(32, 124),
					Transactions: []string{
						FillEncodedByteSlice(32, 125),
						FillEncodedByteSlice(32, 126),
					},
					Withdrawals: []*structs.Withdrawal{
						{
							WithdrawalIndex:  "127",
							ValidatorIndex:   "128",
							ExecutionAddress: FillEncodedByteSlice(20, 129),
							Amount:           "130",
						},
						{
							WithdrawalIndex:  "131",
							ValidatorIndex:   "132",
							ExecutionAddress: FillEncodedByteSlice(20, 133),
							Amount:           "134",
						},
					},
					BlobGasUsed:   "135",
					ExcessBlobGas: "136",
				},
				BLSToSilaChanges: []*structs.SignedBLSToSilaChange{
					{
						Message: &structs.BLSToSilaChange{
							ValidatorIndex:     "137",
							FromBLSPubkey:      FillEncodedByteSlice(48, 138),
							ToSilaAddress: FillEncodedByteSlice(20, 139),
						},
						Signature: FillEncodedByteSlice(96, 140),
					},
					{
						Message: &structs.BLSToSilaChange{
							ValidatorIndex:     "141",
							FromBLSPubkey:      FillEncodedByteSlice(48, 142),
							ToSilaAddress: FillEncodedByteSlice(20, 143),
						},
						Signature: FillEncodedByteSlice(96, 144),
					},
				},
				BlobKzgCommitments: []string{FillEncodedByteSlice(48, 145), FillEncodedByteSlice(48, 146)},
			},
		},
		KzgProofs: []string{FillEncodedByteSlice(48, 146)},
		Blobs:     []string{FillEncodedByteSlice(131072, 147)},
	}
}

func GenerateJsonBlindedDenebBeaconBlock() *structs.BlindedBeaconBlockDeneb {
	return &structs.BlindedBeaconBlockDeneb{
		Slot:          "1",
		ProposerIndex: "2",
		ParentRoot:    FillEncodedByteSlice(32, 3),
		StateRoot:     FillEncodedByteSlice(32, 4),
		Body: &structs.BlindedBeaconBlockBodyDeneb{
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
			SilaPayloadHeader: &structs.SilaPayloadHeaderDeneb{
				ParentHash:       FillEncodedByteSlice(32, 112),
				FeeRecipient:     FillEncodedByteSlice(20, 113),
				StateRoot:        FillEncodedByteSlice(32, 114),
				ReceiptsRoot:     FillEncodedByteSlice(32, 115),
				LogsBloom:        FillEncodedByteSlice(256, 116),
				PrevRandao:       FillEncodedByteSlice(32, 117),
				BlockNumber:      "118",
				GasLimit:         "119",
				GasUsed:          "120",
				Timestamp:        "121",
				ExtraData:        FillEncodedByteSlice(32, 122),
				BaseFeePerGas:    bytesutil.LittleEndianBytesToBigInt(FillByteSlice(32, 123)).String(),
				BlockHash:        FillEncodedByteSlice(32, 124),
				TransactionsRoot: FillEncodedByteSlice(32, 125),
				WithdrawalsRoot:  FillEncodedByteSlice(32, 126),
				BlobGasUsed:      "127",
				ExcessBlobGas:    "128",
			},
			BLSToSilaChanges: []*structs.SignedBLSToSilaChange{
				{
					Message: &structs.BLSToSilaChange{
						ValidatorIndex:     "129",
						FromBLSPubkey:      FillEncodedByteSlice(48, 130),
						ToSilaAddress: FillEncodedByteSlice(20, 131),
					},
					Signature: FillEncodedByteSlice(96, 132),
				},
				{
					Message: &structs.BLSToSilaChange{
						ValidatorIndex:     "133",
						FromBLSPubkey:      FillEncodedByteSlice(48, 134),
						ToSilaAddress: FillEncodedByteSlice(20, 135),
					},
					Signature: FillEncodedByteSlice(96, 136),
				},
			},
			BlobKzgCommitments: []string{FillEncodedByteSlice(48, 137), FillEncodedByteSlice(48, 138)},
		},
	}
}
