package test_helpers

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila/common/hexutil"
)

func GenerateProtoGloasBeaconBlock() *silapb.BeaconBlockGloas {
	return &silapb.BeaconBlockGloas{
		Slot:          1,
		ProposerIndex: 2,
		ParentRoot:    FillByteSlice(32, 3),
		StateRoot:     FillByteSlice(32, 4),
		Body: &silapb.BeaconBlockBodyGloas{
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
			},
			AttesterSlashings: []*silapb.AttesterSlashingElectra{},
			Attestations:      []*silapb.AttestationElectra{},
			Deposits:          []*silapb.Deposit{},
			VoluntaryExits:    []*silapb.SignedVoluntaryExit{},
			SyncAggregate: &silapb.SyncAggregate{
				SyncCommitteeBits:      FillByteSlice(64, 100),
				SyncCommitteeSignature: FillByteSlice(96, 101),
			},
			BlsToSilaChanges: []*silapb.SignedBLSToSilaChange{},
			SignedSilaPayloadBid: &silapb.SignedSilaPayloadBid{
				Message: &silapb.SilaPayloadBid{
					ParentBlockHash:       FillByteSlice(32, 110),
					ParentBlockRoot:       FillByteSlice(32, 111),
					BlockHash:             FillByteSlice(32, 112),
					PrevRandao:            FillByteSlice(32, 113),
					FeeRecipient:          FillByteSlice(20, 114),
					GasLimit:              120,
					BuilderIndex:          121,
					Slot:                  1,
					Value:                 123,
					ExecutionPayment:      124,
					BlobKzgCommitments:    [][]byte{},
					SilaRequestsRoot: FillByteSlice(32, 131),
				},
				Signature: FillByteSlice(96, 130),
			},
			PayloadAttestations: []*silapb.PayloadAttestation{},
		},
	}
}

func GenerateJsonGloasBeaconBlock() *structs.BeaconBlockGloas {
	return &structs.BeaconBlockGloas{
		Slot:          "1",
		ProposerIndex: "2",
		ParentRoot:    hexutil.Encode(FillByteSlice(32, 3)),
		StateRoot:     hexutil.Encode(FillByteSlice(32, 4)),
		Body: &structs.BeaconBlockBodyGloas{
			RandaoReveal: hexutil.Encode(FillByteSlice(96, 5)),
			SilaData: &structs.SilaData{
				DepositRoot:  hexutil.Encode(FillByteSlice(32, 6)),
				DepositCount: "7",
				BlockHash:    hexutil.Encode(FillByteSlice(32, 8)),
			},
			Graffiti: hexutil.Encode(FillByteSlice(32, 9)),
			ProposerSlashings: []*structs.ProposerSlashing{
				{
					SignedHeader1: &structs.SignedBeaconBlockHeader{
						Message: &structs.BeaconBlockHeader{
							Slot:          "10",
							ProposerIndex: "11",
							ParentRoot:    hexutil.Encode(FillByteSlice(32, 12)),
							StateRoot:     hexutil.Encode(FillByteSlice(32, 13)),
							BodyRoot:      hexutil.Encode(FillByteSlice(32, 14)),
						},
						Signature: hexutil.Encode(FillByteSlice(96, 15)),
					},
					SignedHeader2: &structs.SignedBeaconBlockHeader{
						Message: &structs.BeaconBlockHeader{
							Slot:          "16",
							ProposerIndex: "17",
							ParentRoot:    hexutil.Encode(FillByteSlice(32, 18)),
							StateRoot:     hexutil.Encode(FillByteSlice(32, 19)),
							BodyRoot:      hexutil.Encode(FillByteSlice(32, 20)),
						},
						Signature: hexutil.Encode(FillByteSlice(96, 21)),
					},
				},
			},
			AttesterSlashings: []*structs.AttesterSlashingElectra{},
			Attestations:      []*structs.AttestationElectra{},
			Deposits:          []*structs.Deposit{},
			VoluntaryExits:    []*structs.SignedVoluntaryExit{},
			SyncAggregate: &structs.SyncAggregate{
				SyncCommitteeBits:      hexutil.Encode(FillByteSlice(64, 100)),
				SyncCommitteeSignature: hexutil.Encode(FillByteSlice(96, 101)),
			},
			BLSToSilaChanges: []*structs.SignedBLSToSilaChange{},
			SignedSilaPayloadBid: &structs.SignedSilaPayloadBid{
				Message: &structs.SilaPayloadBid{
					ParentBlockHash:       hexutil.Encode(FillByteSlice(32, 110)),
					ParentBlockRoot:       hexutil.Encode(FillByteSlice(32, 111)),
					BlockHash:             hexutil.Encode(FillByteSlice(32, 112)),
					PrevRandao:            hexutil.Encode(FillByteSlice(32, 113)),
					FeeRecipient:          hexutil.Encode(FillByteSlice(20, 114)),
					GasLimit:              "120",
					BuilderIndex:          "121",
					Slot:                  "1",
					Value:                 "123",
					ExecutionPayment:      "124",
					BlobKzgCommitments:    []string{},
					SilaRequestsRoot: hexutil.Encode(FillByteSlice(32, 131)),
				},
				Signature: hexutil.Encode(FillByteSlice(96, 130)),
			},
			PayloadAttestations: []*structs.PayloadAttestation{},
		},
	}
}

func GenerateProtoSilaPayloadEnvelope() *silapb.SilaPayloadEnvelope {
	return &silapb.SilaPayloadEnvelope{
		Payload: &silaenginev1.SilaPayloadGloas{
			ParentHash:    FillByteSlice(32, 200),
			FeeRecipient:  FillByteSlice(20, 201),
			StateRoot:     FillByteSlice(32, 202),
			ReceiptsRoot:  FillByteSlice(32, 203),
			LogsBloom:     FillByteSlice(256, 204),
			PrevRandao:    FillByteSlice(32, 205),
			BaseFeePerGas: FillByteSlice(32, 206),
			BlockHash:     FillByteSlice(32, 207),
			ExtraData:     make([]byte, 0),
			SlotNumber:    1,
		},
		SilaRequests:     &silaenginev1.SilaRequests{},
		BuilderIndex:          121,
		BeaconBlockRoot:       FillByteSlice(32, 210),
		ParentBeaconBlockRoot: FillByteSlice(32, 211),
	}
}
