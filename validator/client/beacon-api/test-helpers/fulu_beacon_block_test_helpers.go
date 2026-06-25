package test_helpers

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

func GenerateProtoFuluBeaconBlockContents() *silapb.BeaconBlockContentsFulu {
	electra := GenerateProtoElectraBeaconBlockContents()
	return &silapb.BeaconBlockContentsFulu{
		Block:     electra.Block,
		KzgProofs: electra.KzgProofs,
		Blobs:     electra.Blobs,
	}
}

func GenerateProtoBlindedFuluBeaconBlock() *silapb.BlindedBeaconBlockFulu {
	electra := GenerateProtoBlindedElectraBeaconBlock()
	return &silapb.BlindedBeaconBlockFulu{
		Slot:          electra.Slot,
		ProposerIndex: electra.ProposerIndex,
		ParentRoot:    electra.ParentRoot,
		StateRoot:     electra.StateRoot,
		Body:          electra.Body,
	}
}

func GenerateJsonFuluBeaconBlockContents() *structs.BeaconBlockContentsFulu {
	electra := GenerateJsonElectraBeaconBlockContents()
	return &structs.BeaconBlockContentsFulu{
		Block:     electra.Block,
		KzgProofs: electra.KzgProofs,
		Blobs:     electra.Blobs,
	}
}

func GenerateJsonBlindedFuluBeaconBlock() *structs.BlindedBeaconBlockFulu {
	electra := GenerateJsonBlindedElectraBeaconBlock()
	return &structs.BlindedBeaconBlockFulu{
		Slot:          electra.Slot,
		ProposerIndex: electra.ProposerIndex,
		ParentRoot:    electra.ParentRoot,
		StateRoot:     electra.StateRoot,
		Body:          electra.Body,
	}
}
