package interfaces

import (
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// SignedBeaconBlockHeaderFromBlock function to retrieve signed block header from block.
func SignedBeaconBlockHeaderFromBlock(block *silapb.SignedBeaconBlock) (*silapb.SignedBeaconBlockHeader, error) {
	if block.Block == nil || block.Block.Body == nil {
		return nil, errors.New("nil block")
	}

	bodyRoot, err := block.Block.Body.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get body root of block")
	}
	return &silapb.SignedBeaconBlockHeader{
		Header: &silapb.BeaconBlockHeader{
			Slot:          block.Block.Slot,
			ProposerIndex: block.Block.ProposerIndex,
			ParentRoot:    block.Block.ParentRoot,
			StateRoot:     block.Block.StateRoot,
			BodyRoot:      bodyRoot[:],
		},
		Signature: block.Signature,
	}, nil
}

// SignedBeaconBlockHeaderFromBlockInterface function to retrieve signed block header from block.
func SignedBeaconBlockHeaderFromBlockInterface(sb ReadOnlySignedBeaconBlock) (*silapb.SignedBeaconBlockHeader, error) {
	b := sb.Block()
	if b.IsNil() || b.Body().IsNil() {
		return nil, errors.New("nil block")
	}

	h, err := BeaconBlockHeaderFromBlockInterface(b)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get block header of block")
	}
	sig := sb.Signature()
	return &silapb.SignedBeaconBlockHeader{
		Header:    h,
		Signature: sig[:],
	}, nil
}

// BeaconBlockHeaderFromBlock function to retrieve block header from block.
func BeaconBlockHeaderFromBlock(block *silapb.BeaconBlock) (*silapb.BeaconBlockHeader, error) {
	if block.Body == nil {
		return nil, errors.New("nil block body")
	}

	bodyRoot, err := block.Body.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get body root of block")
	}
	return &silapb.BeaconBlockHeader{
		Slot:          block.Slot,
		ProposerIndex: block.ProposerIndex,
		ParentRoot:    block.ParentRoot,
		StateRoot:     block.StateRoot,
		BodyRoot:      bodyRoot[:],
	}, nil
}

// BeaconBlockHeaderFromBlockInterface function to retrieve block header from block.
func BeaconBlockHeaderFromBlockInterface(block ReadOnlyBeaconBlock) (*silapb.BeaconBlockHeader, error) {
	if block.Body().IsNil() {
		return nil, errors.New("nil block body")
	}

	bodyRoot, err := block.Body().HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get body root of block")
	}
	parentRoot := block.ParentRoot()
	stateRoot := block.StateRoot()
	return &silapb.BeaconBlockHeader{
		Slot:          block.Slot(),
		ProposerIndex: block.ProposerIndex(),
		ParentRoot:    parentRoot[:],
		StateRoot:     stateRoot[:],
		BodyRoot:      bodyRoot[:],
	}, nil
}
