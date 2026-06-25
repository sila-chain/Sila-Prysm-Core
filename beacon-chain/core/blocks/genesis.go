// Package blocks contains block processing libraries according to
// the Sila beacon chain spec.
package blocks

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	enginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/engine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// NewGenesisBlock returns the canonical, genesis block for the beacon chain protocol.
func NewGenesisBlock(stateRoot []byte) *silapb.SignedBeaconBlock {
	zeroHash := params.BeaconConfig().ZeroHash[:]
	block := &silapb.SignedBeaconBlock{
		Block: &silapb.BeaconBlock{
			ParentRoot: zeroHash,
			StateRoot:  bytesutil.PadTo(stateRoot, 32),
			Body: &silapb.BeaconBlockBody{
				RandaoReveal: make([]byte, fieldparams.BLSSignatureLength),
				Eth1Data: &silapb.Eth1Data{
					DepositRoot: make([]byte, 32),
					BlockHash:   make([]byte, 32),
				},
				Graffiti: make([]byte, 32),
			},
		},
		Signature: params.BeaconConfig().EmptySignature[:],
	}
	return block
}

var ErrUnrecognizedState = errors.New("unknown underlying type for state.BeaconState value")

func NewGenesisBlockForState(ctx context.Context, st state.BeaconState) (interfaces.ReadOnlySignedBeaconBlock, error) {
	root, err := st.HashTreeRoot(ctx)
	if err != nil {
		return nil, err
	}
	ps := st.ToProto()
	switch ps.(type) {
	case *silapb.BeaconState:
		return blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlock{
			Block: &silapb.BeaconBlock{
				ParentRoot: params.BeaconConfig().ZeroHash[:],
				StateRoot:  root[:],
				Body: &silapb.BeaconBlockBody{
					RandaoReveal: make([]byte, fieldparams.BLSSignatureLength),
					Eth1Data: &silapb.Eth1Data{
						DepositRoot: make([]byte, 32),
						BlockHash:   make([]byte, 32),
					},
					Graffiti: make([]byte, 32),
				},
			},
			Signature: params.BeaconConfig().EmptySignature[:],
		})
	case *silapb.BeaconStateAltair:
		return blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockAltair{
			Block: &silapb.BeaconBlockAltair{
				ParentRoot: params.BeaconConfig().ZeroHash[:],
				StateRoot:  root[:],
				Body: &silapb.BeaconBlockBodyAltair{
					RandaoReveal: make([]byte, fieldparams.BLSSignatureLength),
					Eth1Data: &silapb.Eth1Data{
						DepositRoot: make([]byte, 32),
						BlockHash:   make([]byte, 32),
					},
					Graffiti: make([]byte, 32),
					SyncAggregate: &silapb.SyncAggregate{
						SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
						SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
					},
				},
			},
			Signature: params.BeaconConfig().EmptySignature[:],
		})
	case *silapb.BeaconStateBellatrix:
		return blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockBellatrix{
			Block: &silapb.BeaconBlockBellatrix{
				ParentRoot: params.BeaconConfig().ZeroHash[:],
				StateRoot:  root[:],
				Body: &silapb.BeaconBlockBodyBellatrix{
					RandaoReveal: make([]byte, 96),
					Eth1Data: &silapb.Eth1Data{
						DepositRoot: make([]byte, 32),
						BlockHash:   make([]byte, 32),
					},
					Graffiti: make([]byte, 32),
					SyncAggregate: &silapb.SyncAggregate{
						SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
						SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
					},
					ExecutionPayload: &enginev1.ExecutionPayload{
						ParentHash:    make([]byte, 32),
						FeeRecipient:  make([]byte, 20),
						StateRoot:     make([]byte, 32),
						ReceiptsRoot:  make([]byte, 32),
						LogsBloom:     make([]byte, 256),
						PrevRandao:    make([]byte, 32),
						ExtraData:     make([]byte, 0),
						BaseFeePerGas: make([]byte, 32),
						BlockHash:     make([]byte, 32),
						Transactions:  make([][]byte, 0),
					},
				},
			},
			Signature: params.BeaconConfig().EmptySignature[:],
		})
	case *silapb.BeaconStateCapella:
		return blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockCapella{
			Block: &silapb.BeaconBlockCapella{
				ParentRoot: params.BeaconConfig().ZeroHash[:],
				StateRoot:  root[:],
				Body: &silapb.BeaconBlockBodyCapella{
					RandaoReveal: make([]byte, 96),
					Eth1Data: &silapb.Eth1Data{
						DepositRoot: make([]byte, 32),
						BlockHash:   make([]byte, 32),
					},
					Graffiti: make([]byte, 32),
					SyncAggregate: &silapb.SyncAggregate{
						SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
						SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
					},
					ExecutionPayload: &enginev1.ExecutionPayloadCapella{
						ParentHash:    make([]byte, 32),
						FeeRecipient:  make([]byte, 20),
						StateRoot:     make([]byte, 32),
						ReceiptsRoot:  make([]byte, 32),
						LogsBloom:     make([]byte, 256),
						PrevRandao:    make([]byte, 32),
						ExtraData:     make([]byte, 0),
						BaseFeePerGas: make([]byte, 32),
						BlockHash:     make([]byte, 32),
						Transactions:  make([][]byte, 0),
						Withdrawals:   make([]*enginev1.Withdrawal, 0),
					},
				},
			},
			Signature: params.BeaconConfig().EmptySignature[:],
		})
	case *silapb.BeaconStateDeneb:
		return blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockDeneb{
			Block: &silapb.BeaconBlockDeneb{
				ParentRoot: params.BeaconConfig().ZeroHash[:],
				StateRoot:  root[:],
				Body: &silapb.BeaconBlockBodyDeneb{
					RandaoReveal: make([]byte, 96),
					Eth1Data: &silapb.Eth1Data{
						DepositRoot: make([]byte, 32),
						BlockHash:   make([]byte, 32),
					},
					Graffiti: make([]byte, 32),
					SyncAggregate: &silapb.SyncAggregate{
						SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
						SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
					},
					ExecutionPayload: &enginev1.ExecutionPayloadDeneb{
						ParentHash:    make([]byte, 32),
						FeeRecipient:  make([]byte, 20),
						StateRoot:     make([]byte, 32),
						ReceiptsRoot:  make([]byte, 32),
						LogsBloom:     make([]byte, 256),
						PrevRandao:    make([]byte, 32),
						ExtraData:     make([]byte, 0),
						BaseFeePerGas: make([]byte, 32),
						BlockHash:     make([]byte, 32),
						Transactions:  make([][]byte, 0),
						Withdrawals:   make([]*enginev1.Withdrawal, 0),
					},
					BlsToExecutionChanges: make([]*silapb.SignedBLSToExecutionChange, 0),
					BlobKzgCommitments:    make([][]byte, 0),
				},
			},
			Signature: params.BeaconConfig().EmptySignature[:],
		})
	case *silapb.BeaconStateElectra:
		return blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockElectra{
			Block:     electraGenesisBlock(root),
			Signature: params.BeaconConfig().EmptySignature[:],
		})
	case *silapb.BeaconStateFulu:
		return blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockFulu{
			Block:     electraGenesisBlock(root),
			Signature: params.BeaconConfig().EmptySignature[:],
		})
	case *silapb.BeaconStateGloas:
		gs := ps.(*silapb.BeaconStateGloas)
		return blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockGloas{
			Block:     gloasGenesisBlock(root, gs.LatestExecutionPayloadBid),
			Signature: params.BeaconConfig().EmptySignature[:],
		})
	default:
		return nil, ErrUnrecognizedState
	}
}

func gloasGenesisBlock(root [fieldparams.RootLength]byte, latestBid *silapb.ExecutionPayloadBid) *silapb.BeaconBlockGloas {
	// The genesis block body's signed_execution_payload_bid mirrors the state's
	// latest_execution_payload_bid so the reconstructed block's body_root matches
	// state.latest_block_header.body_root (which the genesis distribution tool
	// commits to). Falling back to a zero bid is only useful in tests that
	// initialize a Gloas state without populating latest_execution_payload_bid.
	bidMessage := latestBid.Copy()
	if bidMessage == nil {
		bidMessage = &silapb.ExecutionPayloadBid{
			ParentBlockHash:       make([]byte, 32),
			ParentBlockRoot:       make([]byte, 32),
			BlockHash:             make([]byte, 32),
			PrevRandao:            make([]byte, 32),
			FeeRecipient:          make([]byte, 20),
			BlobKzgCommitments:    make([][]byte, 0),
			ExecutionRequestsRoot: make([]byte, 32),
		}
	}
	return &silapb.BeaconBlockGloas{
		ParentRoot: params.BeaconConfig().ZeroHash[:],
		StateRoot:  root[:],
		Body: &silapb.BeaconBlockBodyGloas{
			RandaoReveal: make([]byte, 96),
			Eth1Data: &silapb.Eth1Data{
				DepositRoot: make([]byte, 32),
				BlockHash:   make([]byte, 32),
			},
			Graffiti: make([]byte, 32),
			SyncAggregate: &silapb.SyncAggregate{
				SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
				SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
			},
			SignedExecutionPayloadBid: &silapb.SignedExecutionPayloadBid{
				Message:   bidMessage,
				Signature: make([]byte, fieldparams.BLSSignatureLength),
			},
			PayloadAttestations: make([]*silapb.PayloadAttestation, 0),
			ParentExecutionRequests: &enginev1.ExecutionRequests{
				Withdrawals:    make([]*enginev1.WithdrawalRequest, 0),
				Deposits:       make([]*enginev1.DepositRequest, 0),
				Consolidations: make([]*enginev1.ConsolidationRequest, 0),
			},
		},
	}
}

func electraGenesisBlock(root [fieldparams.RootLength]byte) *silapb.BeaconBlockElectra {
	return &silapb.BeaconBlockElectra{
		ParentRoot: params.BeaconConfig().ZeroHash[:],
		StateRoot:  root[:],
		Body: &silapb.BeaconBlockBodyElectra{
			RandaoReveal: make([]byte, 96),
			Eth1Data: &silapb.Eth1Data{
				DepositRoot: make([]byte, 32),
				BlockHash:   make([]byte, 32),
			},
			Graffiti: make([]byte, 32),
			SyncAggregate: &silapb.SyncAggregate{
				SyncCommitteeBits:      make([]byte, fieldparams.SyncCommitteeLength/8),
				SyncCommitteeSignature: make([]byte, fieldparams.BLSSignatureLength),
			},
			ExecutionPayload: &enginev1.ExecutionPayloadDeneb{
				ParentHash:    make([]byte, 32),
				FeeRecipient:  make([]byte, 20),
				StateRoot:     make([]byte, 32),
				ReceiptsRoot:  make([]byte, 32),
				LogsBloom:     make([]byte, 256),
				PrevRandao:    make([]byte, 32),
				ExtraData:     make([]byte, 0),
				BaseFeePerGas: make([]byte, 32),
				BlockHash:     make([]byte, 32),
				Transactions:  make([][]byte, 0),
				Withdrawals:   make([]*enginev1.Withdrawal, 0),
			},
			BlsToExecutionChanges: make([]*silapb.SignedBLSToExecutionChange, 0),
			BlobKzgCommitments:    make([][]byte, 0),
			ExecutionRequests: &enginev1.ExecutionRequests{
				Withdrawals:    make([]*enginev1.WithdrawalRequest, 0),
				Deposits:       make([]*enginev1.DepositRequest, 0),
				Consolidations: make([]*enginev1.ConsolidationRequest, 0),
			},
		},
	}
}
