package validator

import (
	"fmt"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// constructGenericBeaconBlock constructs a `GenericBeaconBlock` based on the block version and other parameters.
func (vs *Server) constructGenericBeaconBlock(
	sBlk interfaces.SignedBeaconBlock,
	blobsBundler silaenginev1.BlobsBundler,
	winningBid primitives.Wei,
) (*silapb.GenericBeaconBlock, error) {
	if sBlk == nil || sBlk.Block() == nil {
		return nil, errors.New("block cannot be nil")
	}

	blockProto, err := sBlk.Block().Proto()
	if err != nil {
		return nil, err
	}

	isBlinded := sBlk.IsBlinded()
	bidStr := primitives.WeiToBigInt(winningBid).String()

	switch sBlk.Version() {
	case version.Phase0:
		return vs.constructPhase0Block(blockProto), nil
	case version.Altair:
		return vs.constructAltairBlock(blockProto), nil
	case version.Bellatrix:
		return vs.constructBellatrixBlock(blockProto, isBlinded, bidStr), nil
	case version.Capella:
		return vs.constructCapellaBlock(blockProto, isBlinded, bidStr), nil
	case version.Deneb:
		bundle, ok := blobsBundler.(*silaenginev1.BlobsBundle)
		if blobsBundler != nil && !ok {
			return nil, fmt.Errorf("expected *BlobsBundler, got %T", blobsBundler)
		}
		return vs.constructDenebBlock(blockProto, isBlinded, bidStr, bundle), nil
	case version.Electra:
		bundle, ok := blobsBundler.(*silaenginev1.BlobsBundle)
		if blobsBundler != nil && !ok {
			return nil, fmt.Errorf("expected *BlobsBundler, got %T", blobsBundler)
		}
		return vs.constructElectraBlock(blockProto, isBlinded, bidStr, bundle), nil
	case version.Fulu:
		bundle, ok := blobsBundler.(*silaenginev1.BlobsBundleV2)
		if blobsBundler != nil && !ok {
			return nil, fmt.Errorf("expected *BlobsBundleV2, got %T", blobsBundler)
		}
		return vs.constructFuluBlock(blockProto, isBlinded, bidStr, bundle), nil
	case version.Gloas:
		// Gloas blocks do not carry a separate payload value — the bid is part of the block body.
		return &silapb.GenericBeaconBlock{
			Block: &silapb.GenericBeaconBlock_Gloas{Gloas: blockProto.(*silapb.BeaconBlockGloas)},
		}, nil
	default:
		return nil, fmt.Errorf("unknown block version: %d", sBlk.Version())
	}
}

// Helper functions for constructing blocks for each version
func (vs *Server) constructPhase0Block(pb proto.Message) *silapb.GenericBeaconBlock {
	return &silapb.GenericBeaconBlock{Block: &silapb.GenericBeaconBlock_Phase0{Phase0: pb.(*silapb.BeaconBlock)}}
}

func (vs *Server) constructAltairBlock(pb proto.Message) *silapb.GenericBeaconBlock {
	return &silapb.GenericBeaconBlock{Block: &silapb.GenericBeaconBlock_Altair{Altair: pb.(*silapb.BeaconBlockAltair)}}
}

func (vs *Server) constructBellatrixBlock(pb proto.Message, isBlinded bool, payloadValue string) *silapb.GenericBeaconBlock {
	if isBlinded {
		return &silapb.GenericBeaconBlock{Block: &silapb.GenericBeaconBlock_BlindedBellatrix{BlindedBellatrix: pb.(*silapb.BlindedBeaconBlockBellatrix)}, IsBlinded: true, PayloadValue: payloadValue}
	}
	return &silapb.GenericBeaconBlock{Block: &silapb.GenericBeaconBlock_Bellatrix{Bellatrix: pb.(*silapb.BeaconBlockBellatrix)}, IsBlinded: false, PayloadValue: payloadValue}
}

func (vs *Server) constructCapellaBlock(pb proto.Message, isBlinded bool, payloadValue string) *silapb.GenericBeaconBlock {
	if isBlinded {
		return &silapb.GenericBeaconBlock{Block: &silapb.GenericBeaconBlock_BlindedCapella{BlindedCapella: pb.(*silapb.BlindedBeaconBlockCapella)}, IsBlinded: true, PayloadValue: payloadValue}
	}
	return &silapb.GenericBeaconBlock{Block: &silapb.GenericBeaconBlock_Capella{Capella: pb.(*silapb.BeaconBlockCapella)}, IsBlinded: false, PayloadValue: payloadValue}
}

func (vs *Server) constructDenebBlock(blockProto proto.Message, isBlinded bool, payloadValue string, bundle *silaenginev1.BlobsBundle) *silapb.GenericBeaconBlock {
	if isBlinded {
		return &silapb.GenericBeaconBlock{Block: &silapb.GenericBeaconBlock_BlindedDeneb{BlindedDeneb: blockProto.(*silapb.BlindedBeaconBlockDeneb)}, IsBlinded: true, PayloadValue: payloadValue}
	}
	denebContents := &silapb.BeaconBlockContentsDeneb{Block: blockProto.(*silapb.BeaconBlockDeneb)}
	if bundle != nil {
		denebContents.KzgProofs = bundle.Proofs
		denebContents.Blobs = bundle.Blobs
	}
	return &silapb.GenericBeaconBlock{Block: &silapb.GenericBeaconBlock_Deneb{Deneb: denebContents}, IsBlinded: false, PayloadValue: payloadValue}
}

func (vs *Server) constructElectraBlock(blockProto proto.Message, isBlinded bool, payloadValue string, bundle *silaenginev1.BlobsBundle) *silapb.GenericBeaconBlock {
	if isBlinded {
		return &silapb.GenericBeaconBlock{Block: &silapb.GenericBeaconBlock_BlindedElectra{BlindedElectra: blockProto.(*silapb.BlindedBeaconBlockElectra)}, IsBlinded: true, PayloadValue: payloadValue}
	}
	electraContents := &silapb.BeaconBlockContentsElectra{Block: blockProto.(*silapb.BeaconBlockElectra)}
	if bundle != nil {
		electraContents.KzgProofs = bundle.Proofs
		electraContents.Blobs = bundle.Blobs
	}
	return &silapb.GenericBeaconBlock{Block: &silapb.GenericBeaconBlock_Electra{Electra: electraContents}, IsBlinded: false, PayloadValue: payloadValue}
}

func (vs *Server) constructFuluBlock(blockProto proto.Message, isBlinded bool, payloadValue string, bundle *silaenginev1.BlobsBundleV2) *silapb.GenericBeaconBlock {
	if isBlinded {
		return &silapb.GenericBeaconBlock{Block: &silapb.GenericBeaconBlock_BlindedFulu{BlindedFulu: blockProto.(*silapb.BlindedBeaconBlockFulu)}, IsBlinded: true, PayloadValue: payloadValue}
	}
	fuluContents := &silapb.BeaconBlockContentsFulu{Block: blockProto.(*silapb.BeaconBlockElectra)}
	if bundle != nil {
		fuluContents.KzgProofs = bundle.Proofs
		fuluContents.Blobs = bundle.Blobs
	}
	return &silapb.GenericBeaconBlock{Block: &silapb.GenericBeaconBlock_Fulu{Fulu: fuluContents}, IsBlinded: false, PayloadValue: payloadValue}
}
