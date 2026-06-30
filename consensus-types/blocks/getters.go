package blocks

import (
	"fmt"

	"github.com/pkg/errors"
	field_params "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	consensus_types "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	sila "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	validatorpb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/validator-client"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	ssz "github.com/sila-chain/fastssz"
)

// BeaconBlockIsNil checks if any composite field of input signed beacon block is nil.
// Access to these nil fields will result in run time panic,
// it is recommended to run these checks as first line of defense.
func BeaconBlockIsNil(b interfaces.ReadOnlySignedBeaconBlock) error {
	if b == nil || b.IsNil() {
		return ErrNilSignedBeaconBlock
	}
	return nil
}

// Signature returns the respective block signature.
func (b *SignedBeaconBlock) Signature() [field_params.BLSSignatureLength]byte {
	return b.signature
}

// Block returns the underlying beacon block object.
func (b *SignedBeaconBlock) Block() interfaces.ReadOnlyBeaconBlock {
	return b.block
}

// IsNil checks if the underlying beacon block is nil.
func (b *SignedBeaconBlock) IsNil() bool {
	return b == nil || b.block.IsNil()
}

// Copy performs a deep copy of the signed beacon block object.
func (b *SignedBeaconBlock) Copy() (interfaces.SignedBeaconBlock, error) {
	if b == nil {
		return nil, nil
	}

	pb, err := b.Proto()
	if err != nil {
		return nil, err
	}
	switch b.version {
	case version.Phase0:
		return initSignedBlockFromProtoPhase0(pb.(*sila.SignedBeaconBlock).Copy())
	case version.Altair:
		return initSignedBlockFromProtoAltair(pb.(*sila.SignedBeaconBlockAltair).Copy())
	case version.Bellatrix:
		if b.IsBlinded() {
			return initBlindedSignedBlockFromProtoBellatrix(pb.(*sila.SignedBlindedBeaconBlockBellatrix).Copy())
		}
		return initSignedBlockFromProtoBellatrix(pb.(*sila.SignedBeaconBlockBellatrix).Copy())
	case version.Capella:
		if b.IsBlinded() {
			return initBlindedSignedBlockFromProtoCapella(pb.(*sila.SignedBlindedBeaconBlockCapella).Copy())
		}
		return initSignedBlockFromProtoCapella(pb.(*sila.SignedBeaconBlockCapella).Copy())
	case version.Deneb:
		if b.IsBlinded() {
			return initBlindedSignedBlockFromProtoDeneb(pb.(*sila.SignedBlindedBeaconBlockDeneb).Copy())
		}
		return initSignedBlockFromProtoDeneb(pb.(*sila.SignedBeaconBlockDeneb).Copy())
	case version.Electra:
		if b.IsBlinded() {
			return initBlindedSignedBlockFromProtoElectra(pb.(*sila.SignedBlindedBeaconBlockElectra).Copy())
		}
		return initSignedBlockFromProtoElectra(pb.(*sila.SignedBeaconBlockElectra).Copy())
	case version.Fulu:
		if b.IsBlinded() {
			return initBlindedSignedBlockFromProtoFulu(pb.(*sila.SignedBlindedBeaconBlockFulu).Copy())
		}
		return initSignedBlockFromProtoFulu(pb.(*sila.SignedBeaconBlockFulu).Copy())
	case version.Gloas:
		return initSignedBlockFromProtoGloas(sila.CopySignedBeaconBlockGloas(pb.(*sila.SignedBeaconBlockGloas)))
	default:
		return nil, errIncorrectBlockVersion
	}
}

// PbGenericBlock returns a generic signed beacon block.
func (b *SignedBeaconBlock) PbGenericBlock() (*sila.GenericSignedBeaconBlock, error) {
	pb, err := b.Proto()
	if err != nil {
		return nil, err
	}
	switch b.version {
	case version.Phase0:
		return &sila.GenericSignedBeaconBlock{
			Block: &sila.GenericSignedBeaconBlock_Phase0{Phase0: pb.(*sila.SignedBeaconBlock)},
		}, nil
	case version.Altair:
		return &sila.GenericSignedBeaconBlock{
			Block: &sila.GenericSignedBeaconBlock_Altair{Altair: pb.(*sila.SignedBeaconBlockAltair)},
		}, nil
	case version.Bellatrix:
		if b.IsBlinded() {
			return &sila.GenericSignedBeaconBlock{
				Block: &sila.GenericSignedBeaconBlock_BlindedBellatrix{BlindedBellatrix: pb.(*sila.SignedBlindedBeaconBlockBellatrix)},
			}, nil
		}
		return &sila.GenericSignedBeaconBlock{
			Block: &sila.GenericSignedBeaconBlock_Bellatrix{Bellatrix: pb.(*sila.SignedBeaconBlockBellatrix)},
		}, nil
	case version.Capella:
		if b.IsBlinded() {
			return &sila.GenericSignedBeaconBlock{
				Block: &sila.GenericSignedBeaconBlock_BlindedCapella{BlindedCapella: pb.(*sila.SignedBlindedBeaconBlockCapella)},
			}, nil
		}
		return &sila.GenericSignedBeaconBlock{
			Block: &sila.GenericSignedBeaconBlock_Capella{Capella: pb.(*sila.SignedBeaconBlockCapella)},
		}, nil
	case version.Deneb:
		if b.IsBlinded() {
			return &sila.GenericSignedBeaconBlock{
				Block: &sila.GenericSignedBeaconBlock_BlindedDeneb{BlindedDeneb: pb.(*sila.SignedBlindedBeaconBlockDeneb)},
			}, nil
		}
		bc, ok := pb.(*sila.SignedBeaconBlockContentsDeneb)
		if !ok {
			return nil, fmt.Errorf("PbGenericBlock() only supports block content type but got %T", pb)
		}
		return &sila.GenericSignedBeaconBlock{
			Block: &sila.GenericSignedBeaconBlock_Deneb{Deneb: bc},
		}, nil
	case version.Electra:
		if b.IsBlinded() {
			return &sila.GenericSignedBeaconBlock{
				Block: &sila.GenericSignedBeaconBlock_BlindedElectra{BlindedElectra: pb.(*sila.SignedBlindedBeaconBlockElectra)},
			}, nil
		}
		bc, ok := pb.(*sila.SignedBeaconBlockContentsElectra)
		if !ok {
			return nil, fmt.Errorf("PbGenericBlock() only supports block content type but got %T", pb)
		}
		return &sila.GenericSignedBeaconBlock{
			Block: &sila.GenericSignedBeaconBlock_Electra{Electra: bc},
		}, nil
	case version.Fulu:
		if b.IsBlinded() {
			return &sila.GenericSignedBeaconBlock{
				Block: &sila.GenericSignedBeaconBlock_BlindedFulu{BlindedFulu: pb.(*sila.SignedBlindedBeaconBlockFulu)},
			}, nil
		}
		bc, ok := pb.(*sila.SignedBeaconBlockContentsFulu)
		if !ok {
			return nil, fmt.Errorf("PbGenericBlock() only supports block content type but got %T", pb)
		}
		return &sila.GenericSignedBeaconBlock{
			Block: &sila.GenericSignedBeaconBlock_Fulu{Fulu: bc},
		}, nil
	case version.Gloas:
		return &sila.GenericSignedBeaconBlock{
			Block: &sila.GenericSignedBeaconBlock_Gloas{Gloas: pb.(*sila.SignedBeaconBlockGloas)},
		}, nil
	default:
		return nil, errIncorrectBlockVersion
	}
}

// ToBlinded converts a non-blinded block to its blinded equivalent.
func (b *SignedBeaconBlock) ToBlinded() (interfaces.ReadOnlySignedBeaconBlock, error) {
	if b.version < version.Bellatrix || b.version >= version.Gloas {
		return nil, ErrUnsupportedVersion
	}
	if b.IsBlinded() {
		return b, nil
	}
	if b.block.IsNil() {
		return nil, errors.New("cannot convert nil block to blinded format")
	}
	payload, err := b.block.Body().SilaData()
	if err != nil {
		return nil, err
	}

	if b.version >= version.Fulu {
		p, ok := payload.Proto().(*silaenginev1.SilaPayloadDeneb)
		if !ok {
			return nil, fmt.Errorf("%T is not a sila payload header of Deneb version", p)
		}
		header, err := PayloadToHeaderFulu(payload)
		if err != nil {
			return nil, errors.Wrap(err, "payload to header fulu")
		}

		return initBlindedSignedBlockFromProtoFulu(
			&sila.SignedBlindedBeaconBlockFulu{
				Message: &sila.BlindedBeaconBlockFulu{
					Slot:          b.block.slot,
					ProposerIndex: b.block.proposerIndex,
					ParentRoot:    b.block.parentRoot[:],
					StateRoot:     b.block.stateRoot[:],
					Body: &sila.BlindedBeaconBlockBodyElectra{
						RandaoReveal:       b.block.body.randaoReveal[:],
						SilaData:           b.block.body.silaexecData,
						Graffiti:           b.block.body.graffiti[:],
						ProposerSlashings:  b.block.body.proposerSlashings,
						AttesterSlashings:  b.block.body.attesterSlashingsElectra,
						Attestations:       b.block.body.attestationsElectra,
						Deposits:           b.block.body.deposits,
						VoluntaryExits:     b.block.body.voluntaryExits,
						SyncAggregate:      b.block.body.syncAggregate,
						SilaPayloadHeader:  header,
						BlsToSilaChanges:   b.block.body.blsToSilaChanges,
						BlobKzgCommitments: b.block.body.blobKzgCommitments,
						SilaRequests:       b.block.body.silaRequests,
					},
				},
				Signature: b.signature[:],
			})
	}

	if b.version >= version.Electra {
		p, ok := payload.Proto().(*silaenginev1.SilaPayloadDeneb)
		if !ok {
			return nil, fmt.Errorf("%T is not a sila payload header of Deneb version", p)
		}
		header, err := PayloadToHeaderElectra(payload)
		if err != nil {
			return nil, errors.Wrap(err, "payload to header electra")
		}
		return initBlindedSignedBlockFromProtoElectra(
			&sila.SignedBlindedBeaconBlockElectra{
				Message: &sila.BlindedBeaconBlockElectra{
					Slot:          b.block.slot,
					ProposerIndex: b.block.proposerIndex,
					ParentRoot:    b.block.parentRoot[:],
					StateRoot:     b.block.stateRoot[:],
					Body: &sila.BlindedBeaconBlockBodyElectra{
						RandaoReveal:       b.block.body.randaoReveal[:],
						SilaData:           b.block.body.silaexecData,
						Graffiti:           b.block.body.graffiti[:],
						ProposerSlashings:  b.block.body.proposerSlashings,
						AttesterSlashings:  b.block.body.attesterSlashingsElectra,
						Attestations:       b.block.body.attestationsElectra,
						Deposits:           b.block.body.deposits,
						VoluntaryExits:     b.block.body.voluntaryExits,
						SyncAggregate:      b.block.body.syncAggregate,
						SilaPayloadHeader:  header,
						BlsToSilaChanges:   b.block.body.blsToSilaChanges,
						BlobKzgCommitments: b.block.body.blobKzgCommitments,
						SilaRequests:       b.block.body.silaRequests,
					},
				},
				Signature: b.signature[:],
			})
	}

	switch p := payload.Proto().(type) {
	case *silaenginev1.SilaPayload:
		header, err := PayloadToHeader(payload)
		if err != nil {
			return nil, errors.Wrap(err, "payload to header")
		}
		return initBlindedSignedBlockFromProtoBellatrix(
			&sila.SignedBlindedBeaconBlockBellatrix{
				Block: &sila.BlindedBeaconBlockBellatrix{
					Slot:          b.block.slot,
					ProposerIndex: b.block.proposerIndex,
					ParentRoot:    b.block.parentRoot[:],
					StateRoot:     b.block.stateRoot[:],
					Body: &sila.BlindedBeaconBlockBodyBellatrix{
						RandaoReveal:      b.block.body.randaoReveal[:],
						SilaData:          b.block.body.silaexecData,
						Graffiti:          b.block.body.graffiti[:],
						ProposerSlashings: b.block.body.proposerSlashings,
						AttesterSlashings: b.block.body.attesterSlashings,
						Attestations:      b.block.body.attestations,
						Deposits:          b.block.body.deposits,
						VoluntaryExits:    b.block.body.voluntaryExits,
						SyncAggregate:     b.block.body.syncAggregate,
						SilaPayloadHeader: header,
					},
				},
				Signature: b.signature[:],
			})
	case *silaenginev1.SilaPayloadCapella:
		header, err := PayloadToHeaderCapella(payload)
		if err != nil {
			return nil, err
		}
		return initBlindedSignedBlockFromProtoCapella(
			&sila.SignedBlindedBeaconBlockCapella{
				Block: &sila.BlindedBeaconBlockCapella{
					Slot:          b.block.slot,
					ProposerIndex: b.block.proposerIndex,
					ParentRoot:    b.block.parentRoot[:],
					StateRoot:     b.block.stateRoot[:],
					Body: &sila.BlindedBeaconBlockBodyCapella{
						RandaoReveal:      b.block.body.randaoReveal[:],
						SilaData:          b.block.body.silaexecData,
						Graffiti:          b.block.body.graffiti[:],
						ProposerSlashings: b.block.body.proposerSlashings,
						AttesterSlashings: b.block.body.attesterSlashings,
						Attestations:      b.block.body.attestations,
						Deposits:          b.block.body.deposits,
						VoluntaryExits:    b.block.body.voluntaryExits,
						SyncAggregate:     b.block.body.syncAggregate,
						SilaPayloadHeader: header,
						BlsToSilaChanges:  b.block.body.blsToSilaChanges,
					},
				},
				Signature: b.signature[:],
			})
	case *silaenginev1.SilaPayloadDeneb:
		header, err := PayloadToHeaderDeneb(payload)
		if err != nil {
			return nil, errors.Wrap(err, "payload to header deneb")
		}
		return initBlindedSignedBlockFromProtoDeneb(
			&sila.SignedBlindedBeaconBlockDeneb{
				Message: &sila.BlindedBeaconBlockDeneb{
					Slot:          b.block.slot,
					ProposerIndex: b.block.proposerIndex,
					ParentRoot:    b.block.parentRoot[:],
					StateRoot:     b.block.stateRoot[:],
					Body: &sila.BlindedBeaconBlockBodyDeneb{
						RandaoReveal:       b.block.body.randaoReveal[:],
						SilaData:           b.block.body.silaexecData,
						Graffiti:           b.block.body.graffiti[:],
						ProposerSlashings:  b.block.body.proposerSlashings,
						AttesterSlashings:  b.block.body.attesterSlashings,
						Attestations:       b.block.body.attestations,
						Deposits:           b.block.body.deposits,
						VoluntaryExits:     b.block.body.voluntaryExits,
						SyncAggregate:      b.block.body.syncAggregate,
						SilaPayloadHeader:  header,
						BlsToSilaChanges:   b.block.body.blsToSilaChanges,
						BlobKzgCommitments: b.block.body.blobKzgCommitments,
					},
				},
				Signature: b.signature[:],
			})
	default:
		return nil, fmt.Errorf("%T is not a sila payload header", p)
	}
}

func (b *SignedBeaconBlock) Unblind(e interfaces.SilaData) error {
	if e == nil || e.IsNil() {
		return errors.New("cannot unblind with nil sila data")
	}
	if !b.IsBlinded() {
		return errors.New("cannot unblind if the block is already unblinded")
	}
	payloadRoot, err := e.HashTreeRoot()
	if err != nil {
		return err
	}
	header, err := b.Block().Body().SilaData()
	if err != nil {
		return err
	}
	headerRoot, err := header.HashTreeRoot()
	if err != nil {
		return err
	}
	if payloadRoot != headerRoot {
		return errors.New("cannot unblind with different sila data")
	}
	if err := b.SetSilaData(e); err != nil {
		return err
	}
	return nil
}

// Version of the underlying protobuf object.
func (b *SignedBeaconBlock) Version() int {
	return b.version
}

// IsBlinded metadata on whether a block is blinded
func (b *SignedBeaconBlock) IsBlinded() bool {
	return b.version < version.Gloas && b.version >= version.Bellatrix && b.block.body.silaPayload == nil
}

// Header converts the underlying protobuf object from blinded block to header format.
func (b *SignedBeaconBlock) Header() (*sila.SignedBeaconBlockHeader, error) {
	if b.IsNil() {
		return nil, errNilBlock
	}
	root, err := b.block.body.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrapf(err, "could not hash block body")
	}

	return &sila.SignedBeaconBlockHeader{
		Header: &sila.BeaconBlockHeader{
			Slot:          b.block.slot,
			ProposerIndex: b.block.proposerIndex,
			ParentRoot:    b.block.parentRoot[:],
			StateRoot:     b.block.stateRoot[:],
			BodyRoot:      root[:],
		},
		Signature: b.signature[:],
	}, nil
}

// MarshalSSZ marshals the signed beacon block to its relevant ssz form.
func (b *SignedBeaconBlock) MarshalSSZ() ([]byte, error) {
	pb, err := b.Proto()
	if err != nil {
		return []byte{}, err
	}
	switch b.version {
	case version.Phase0:
		return pb.(*sila.SignedBeaconBlock).MarshalSSZ()
	case version.Altair:
		return pb.(*sila.SignedBeaconBlockAltair).MarshalSSZ()
	case version.Bellatrix:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockBellatrix).MarshalSSZ()
		}
		return pb.(*sila.SignedBeaconBlockBellatrix).MarshalSSZ()
	case version.Capella:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockCapella).MarshalSSZ()
		}
		return pb.(*sila.SignedBeaconBlockCapella).MarshalSSZ()
	case version.Deneb:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockDeneb).MarshalSSZ()
		}
		return pb.(*sila.SignedBeaconBlockDeneb).MarshalSSZ()
	case version.Electra:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockElectra).MarshalSSZ()
		}
		return pb.(*sila.SignedBeaconBlockElectra).MarshalSSZ()
	case version.Fulu:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockFulu).MarshalSSZ()
		}
		return pb.(*sila.SignedBeaconBlockFulu).MarshalSSZ()
	case version.Gloas:
		return pb.(*sila.SignedBeaconBlockGloas).MarshalSSZ()
	default:
		return []byte{}, errIncorrectBlockVersion
	}
}

// MarshalSSZTo marshals the signed beacon block's ssz
// form to the provided byte buffer.
func (b *SignedBeaconBlock) MarshalSSZTo(dst []byte) ([]byte, error) {
	pb, err := b.Proto()
	if err != nil {
		return []byte{}, err
	}
	switch b.version {
	case version.Phase0:
		return pb.(*sila.SignedBeaconBlock).MarshalSSZTo(dst)
	case version.Altair:
		return pb.(*sila.SignedBeaconBlockAltair).MarshalSSZTo(dst)
	case version.Bellatrix:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockBellatrix).MarshalSSZTo(dst)
		}
		return pb.(*sila.SignedBeaconBlockBellatrix).MarshalSSZTo(dst)
	case version.Capella:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockCapella).MarshalSSZTo(dst)
		}
		return pb.(*sila.SignedBeaconBlockCapella).MarshalSSZTo(dst)
	case version.Deneb:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockDeneb).MarshalSSZTo(dst)
		}
		return pb.(*sila.SignedBeaconBlockDeneb).MarshalSSZTo(dst)
	case version.Electra:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockElectra).MarshalSSZTo(dst)
		}
		return pb.(*sila.SignedBeaconBlockElectra).MarshalSSZTo(dst)
	case version.Fulu:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockFulu).MarshalSSZTo(dst)
		}
		return pb.(*sila.SignedBeaconBlockFulu).MarshalSSZTo(dst)
	case version.Gloas:
		return pb.(*sila.SignedBeaconBlockGloas).MarshalSSZTo(dst)
	default:
		return []byte{}, errIncorrectBlockVersion
	}
}

// SizeSSZ returns the size of the serialized signed block
//
// WARNING: This function panics. It is required to change the signature
// of fastssz's SizeSSZ() interface function to avoid panicking.
// Changing the signature causes very problematic issues with Sila deps.
// For the time being panicking is preferable.
// lint:nopanic -- Panic warning is communicated in godoc commentary.
func (b *SignedBeaconBlock) SizeSSZ() int {
	pb, err := b.Proto()
	if err != nil {
		panic(err)
	}
	switch b.version {
	case version.Phase0:
		return pb.(*sila.SignedBeaconBlock).SizeSSZ()
	case version.Altair:
		return pb.(*sila.SignedBeaconBlockAltair).SizeSSZ()
	case version.Bellatrix:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockBellatrix).SizeSSZ()
		}
		return pb.(*sila.SignedBeaconBlockBellatrix).SizeSSZ()
	case version.Capella:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockCapella).SizeSSZ()
		}
		return pb.(*sila.SignedBeaconBlockCapella).SizeSSZ()
	case version.Deneb:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockDeneb).SizeSSZ()
		}
		return pb.(*sila.SignedBeaconBlockDeneb).SizeSSZ()
	case version.Electra:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockElectra).SizeSSZ()
		}
		return pb.(*sila.SignedBeaconBlockElectra).SizeSSZ()
	case version.Fulu:
		if b.IsBlinded() {
			return pb.(*sila.SignedBlindedBeaconBlockFulu).SizeSSZ()
		}
		return pb.(*sila.SignedBeaconBlockFulu).SizeSSZ()
	case version.Gloas:
		return pb.(*sila.SignedBeaconBlockGloas).SizeSSZ()
	default:
		panic(incorrectBlockVersion)
	}
}

// UnmarshalSSZ unmarshals the sitime/slots/slottime.gogned beacon block from its relevant ssz form.
// nolint:gocognit
func (b *SignedBeaconBlock) UnmarshalSSZ(buf []byte) error {
	var newBlock *SignedBeaconBlock
	switch b.version {
	case version.Phase0:
		pb := &sila.SignedBeaconBlock{}
		if err := pb.UnmarshalSSZ(buf); err != nil {
			return err
		}
		var err error
		newBlock, err = initSignedBlockFromProtoPhase0(pb)
		if err != nil {
			return err
		}
	case version.Altair:
		pb := &sila.SignedBeaconBlockAltair{}
		if err := pb.UnmarshalSSZ(buf); err != nil {
			return err
		}
		var err error
		newBlock, err = initSignedBlockFromProtoAltair(pb)
		if err != nil {
			return err
		}
	case version.Bellatrix:
		if b.IsBlinded() {
			pb := &sila.SignedBlindedBeaconBlockBellatrix{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlindedSignedBlockFromProtoBellatrix(pb)
			if err != nil {
				return err
			}
		} else {
			pb := &sila.SignedBeaconBlockBellatrix{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initSignedBlockFromProtoBellatrix(pb)
			if err != nil {
				return err
			}
		}
	case version.Capella:
		if b.IsBlinded() {
			pb := &sila.SignedBlindedBeaconBlockCapella{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlindedSignedBlockFromProtoCapella(pb)
			if err != nil {
				return err
			}
		} else {
			pb := &sila.SignedBeaconBlockCapella{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initSignedBlockFromProtoCapella(pb)
			if err != nil {
				return err
			}
		}
	case version.Deneb:
		if b.IsBlinded() {
			pb := &sila.SignedBlindedBeaconBlockDeneb{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlindedSignedBlockFromProtoDeneb(pb)
			if err != nil {
				return err
			}
		} else {
			pb := &sila.SignedBeaconBlockDeneb{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initSignedBlockFromProtoDeneb(pb)
			if err != nil {
				return err
			}
		}
	case version.Electra:
		if b.IsBlinded() {
			pb := &sila.SignedBlindedBeaconBlockElectra{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlindedSignedBlockFromProtoElectra(pb)
			if err != nil {
				return err
			}
		} else {
			pb := &sila.SignedBeaconBlockElectra{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initSignedBlockFromProtoElectra(pb)
			if err != nil {
				return err
			}
		}
	case version.Fulu:
		if b.IsBlinded() {
			pb := &sila.SignedBlindedBeaconBlockFulu{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlindedSignedBlockFromProtoFulu(pb)
			if err != nil {
				return err
			}
		} else {
			pb := &sila.SignedBeaconBlockFulu{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initSignedBlockFromProtoFulu(pb)
			if err != nil {
				return err
			}
		}
	case version.Gloas:
		pb := &sila.SignedBeaconBlockGloas{}
		err := pb.UnmarshalSSZ(buf)
		if err != nil {
			return err
		}
		newBlock, err = initSignedBlockFromProtoGloas(pb)
		if err != nil {
			return err
		}
	default:
		return errIncorrectBlockVersion
	}
	*b = *newBlock
	return nil
}

// Slot returns the respective slot of the block.
func (b *BeaconBlock) Slot() primitives.Slot {
	return b.slot
}

// ProposerIndex returns the proposer index of the beacon block.
func (b *BeaconBlock) ProposerIndex() primitives.ValidatorIndex {
	return b.proposerIndex
}

// ParentRoot returns the parent root of beacon block.
func (b *BeaconBlock) ParentRoot() [field_params.RootLength]byte {
	return b.parentRoot
}

// StateRoot returns the state root of the beacon block.
func (b *BeaconBlock) StateRoot() [field_params.RootLength]byte {
	return b.stateRoot
}

// Body returns the underlying block body.
func (b *BeaconBlock) Body() interfaces.ReadOnlyBeaconBlockBody {
	return b.body
}

// IsNil checks if the beacon block is nil.
func (b *BeaconBlock) IsNil() bool {
	return b == nil || b.Body().IsNil()
}

// IsBlinded checks if the beacon block is a blinded block.
func (b *BeaconBlock) IsBlinded() bool {
	return b.version < version.Gloas && b.version >= version.Bellatrix && b.body.silaPayload == nil
}

// Version of the underlying protobuf object.
func (b *BeaconBlock) Version() int {
	return b.version
}

// HashTreeRoot returns the ssz root of the block.
func (b *BeaconBlock) HashTreeRoot() ([field_params.RootLength]byte, error) {
	pb, err := b.Proto()
	if err != nil {
		return [field_params.RootLength]byte{}, err
	}
	switch b.version {
	case version.Phase0:
		return pb.(*sila.BeaconBlock).HashTreeRoot()
	case version.Altair:
		return pb.(*sila.BeaconBlockAltair).HashTreeRoot()
	case version.Bellatrix:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockBellatrix).HashTreeRoot()
		}
		return pb.(*sila.BeaconBlockBellatrix).HashTreeRoot()
	case version.Capella:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockCapella).HashTreeRoot()
		}
		return pb.(*sila.BeaconBlockCapella).HashTreeRoot()
	case version.Deneb:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockDeneb).HashTreeRoot()
		}
		return pb.(*sila.BeaconBlockDeneb).HashTreeRoot()
	case version.Electra:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockElectra).HashTreeRoot()
		}
		return pb.(*sila.BeaconBlockElectra).HashTreeRoot()
	case version.Fulu:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockFulu).HashTreeRoot()
		}
		return pb.(*sila.BeaconBlockElectra).HashTreeRoot()
	case version.Gloas:
		return pb.(*sila.BeaconBlockGloas).HashTreeRoot()

	default:
		return [field_params.RootLength]byte{}, errIncorrectBlockVersion
	}
}

// HashTreeRootWith ssz hashes the BeaconBlock object with a hasher.
func (b *BeaconBlock) HashTreeRootWith(h *ssz.Hasher) error {
	pb, err := b.Proto()
	if err != nil {
		return err
	}
	switch b.version {
	case version.Phase0:
		return pb.(*sila.BeaconBlock).HashTreeRootWith(h)
	case version.Altair:
		return pb.(*sila.BeaconBlockAltair).HashTreeRootWith(h)
	case version.Bellatrix:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockBellatrix).HashTreeRootWith(h)
		}
		return pb.(*sila.BeaconBlockBellatrix).HashTreeRootWith(h)
	case version.Capella:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockCapella).HashTreeRootWith(h)
		}
		return pb.(*sila.BeaconBlockCapella).HashTreeRootWith(h)
	case version.Deneb:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockDeneb).HashTreeRootWith(h)
		}
		return pb.(*sila.BeaconBlockDeneb).HashTreeRootWith(h)
	case version.Electra:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockElectra).HashTreeRootWith(h)
		}
		return pb.(*sila.BeaconBlockElectra).HashTreeRootWith(h)
	case version.Fulu:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockFulu).HashTreeRootWith(h)
		}
		return pb.(*sila.BeaconBlockElectra).HashTreeRootWith(h)
	case version.Gloas:
		return pb.(*sila.BeaconBlockGloas).HashTreeRootWith(h)
	default:
		return errIncorrectBlockVersion
	}
}

// MarshalSSZ marshals the block into its respective
// ssz form.
func (b *BeaconBlock) MarshalSSZ() ([]byte, error) {
	pb, err := b.Proto()
	if err != nil {
		return []byte{}, err
	}
	switch b.version {
	case version.Phase0:
		return pb.(*sila.BeaconBlock).MarshalSSZ()
	case version.Altair:
		return pb.(*sila.BeaconBlockAltair).MarshalSSZ()
	case version.Bellatrix:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockBellatrix).MarshalSSZ()
		}
		return pb.(*sila.BeaconBlockBellatrix).MarshalSSZ()
	case version.Capella:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockCapella).MarshalSSZ()
		}
		return pb.(*sila.BeaconBlockCapella).MarshalSSZ()
	case version.Deneb:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockDeneb).MarshalSSZ()
		}
		return pb.(*sila.BeaconBlockDeneb).MarshalSSZ()
	case version.Electra:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockElectra).MarshalSSZ()
		}
		return pb.(*sila.BeaconBlockElectra).MarshalSSZ()
	case version.Fulu:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockFulu).MarshalSSZ()
		}
		return pb.(*sila.BeaconBlockElectra).MarshalSSZ()
	case version.Gloas:
		return pb.(*sila.BeaconBlockGloas).MarshalSSZ()
	default:
		return []byte{}, errIncorrectBlockVersion
	}
}

// MarshalSSZTo marshals the beacon block's ssz
// form to the provided byte buffer.
func (b *BeaconBlock) MarshalSSZTo(dst []byte) ([]byte, error) {
	pb, err := b.Proto()
	if err != nil {
		return []byte{}, err
	}
	switch b.version {
	case version.Phase0:
		return pb.(*sila.BeaconBlock).MarshalSSZTo(dst)
	case version.Altair:
		return pb.(*sila.BeaconBlockAltair).MarshalSSZTo(dst)
	case version.Bellatrix:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockBellatrix).MarshalSSZTo(dst)
		}
		return pb.(*sila.BeaconBlockBellatrix).MarshalSSZTo(dst)
	case version.Capella:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockCapella).MarshalSSZTo(dst)
		}
		return pb.(*sila.BeaconBlockCapella).MarshalSSZTo(dst)
	case version.Deneb:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockDeneb).MarshalSSZTo(dst)
		}
		return pb.(*sila.BeaconBlockDeneb).MarshalSSZTo(dst)
	case version.Electra:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockElectra).MarshalSSZTo(dst)
		}
		return pb.(*sila.BeaconBlockElectra).MarshalSSZTo(dst)
	case version.Fulu:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockFulu).MarshalSSZTo(dst)
		}
		return pb.(*sila.BeaconBlockElectra).MarshalSSZTo(dst)
	case version.Gloas:
		return pb.(*sila.BeaconBlockGloas).MarshalSSZTo(dst)
	default:
		return []byte{}, errIncorrectBlockVersion
	}
}

// SizeSSZ returns the size of the serialized block.
//
// WARNING: This function panics. It is required to change the signature
// of fastssz's SizeSSZ() interface function to avoid panicking.
// Changing the signature causes very problematic issues with Sila deps.
// For the time being panicking is preferable.
// lint:nopanic -- Panic is communicated in godoc.
func (b *BeaconBlock) SizeSSZ() int {
	pb, err := b.Proto()
	if err != nil {
		panic(err)
	}
	switch b.version {
	case version.Phase0:
		return pb.(*sila.BeaconBlock).SizeSSZ()
	case version.Altair:
		return pb.(*sila.BeaconBlockAltair).SizeSSZ()
	case version.Bellatrix:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockBellatrix).SizeSSZ()
		}
		return pb.(*sila.BeaconBlockBellatrix).SizeSSZ()
	case version.Capella:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockCapella).SizeSSZ()
		}
		return pb.(*sila.BeaconBlockCapella).SizeSSZ()
	case version.Deneb:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockDeneb).SizeSSZ()
		}
		return pb.(*sila.BeaconBlockDeneb).SizeSSZ()
	case version.Electra:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockElectra).SizeSSZ()
		}
		return pb.(*sila.BeaconBlockElectra).SizeSSZ()
	case version.Fulu:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockFulu).SizeSSZ()
		}
		return pb.(*sila.BeaconBlockElectra).SizeSSZ()
	case version.Gloas:
		return pb.(*sila.BeaconBlockGloas).SizeSSZ()
	default:
		panic(incorrectBodyVersion)
	}
}

// UnmarshalSSZ unmarshals the beacon block from its relevant ssz form.
// nolint:gocognit
func (b *BeaconBlock) UnmarshalSSZ(buf []byte) error {
	var newBlock *BeaconBlock
	switch b.version {
	case version.Phase0:
		pb := &sila.BeaconBlock{}
		if err := pb.UnmarshalSSZ(buf); err != nil {
			return err
		}
		var err error
		newBlock, err = initBlockFromProtoPhase0(pb)
		if err != nil {
			return err
		}
	case version.Altair:
		pb := &sila.BeaconBlockAltair{}
		if err := pb.UnmarshalSSZ(buf); err != nil {
			return err
		}
		var err error
		newBlock, err = initBlockFromProtoAltair(pb)
		if err != nil {
			return err
		}
	case version.Bellatrix:
		if b.IsBlinded() {
			pb := &sila.BlindedBeaconBlockBellatrix{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlindedBlockFromProtoBellatrix(pb)
			if err != nil {
				return err
			}
		} else {
			pb := &sila.BeaconBlockBellatrix{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlockFromProtoBellatrix(pb)
			if err != nil {
				return err
			}
		}
	case version.Capella:
		if b.IsBlinded() {
			pb := &sila.BlindedBeaconBlockCapella{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlindedBlockFromProtoCapella(pb)
			if err != nil {
				return err
			}
		} else {
			pb := &sila.BeaconBlockCapella{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlockFromProtoCapella(pb)
			if err != nil {
				return err
			}
		}
	case version.Deneb:
		if b.IsBlinded() {
			pb := &sila.BlindedBeaconBlockDeneb{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlindedBlockFromProtoDeneb(pb)
			if err != nil {
				return err
			}
		} else {
			pb := &sila.BeaconBlockDeneb{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlockFromProtoDeneb(pb)
			if err != nil {
				return err
			}
		}
	case version.Electra:
		if b.IsBlinded() {
			pb := &sila.BlindedBeaconBlockElectra{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlindedBlockFromProtoElectra(pb)
			if err != nil {
				return err
			}
		} else {
			pb := &sila.BeaconBlockElectra{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlockFromProtoElectra(pb)
			if err != nil {
				return err
			}
		}
	case version.Fulu:
		if b.IsBlinded() {
			pb := &sila.BlindedBeaconBlockFulu{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlindedBlockFromProtoFulu(pb)
			if err != nil {
				return err
			}
		} else {
			pb := &sila.BeaconBlockElectra{}
			if err := pb.UnmarshalSSZ(buf); err != nil {
				return err
			}
			var err error
			newBlock, err = initBlockFromProtoFulu(pb)
			if err != nil {
				return err
			}
		}
	case version.Gloas:
		pb := &sila.BeaconBlockGloas{}
		if err := pb.UnmarshalSSZ(buf); err != nil {
			return err
		}
		var err error
		newBlock, err = initBlockFromProtoGloas(pb)
		if err != nil {
			return err
		}
	default:
		return errIncorrectBlockVersion
	}
	*b = *newBlock
	return nil
}

// AsSignRequestObject returns the underlying sign request object.
func (b *BeaconBlock) AsSignRequestObject() (validatorpb.SignRequestObject, error) {
	pb, err := b.Proto()
	if err != nil {
		return nil, err
	}
	switch b.version {
	case version.Phase0:
		return &validatorpb.SignRequest_Block{Block: pb.(*sila.BeaconBlock)}, nil
	case version.Altair:
		return &validatorpb.SignRequest_BlockAltair{BlockAltair: pb.(*sila.BeaconBlockAltair)}, nil
	case version.Bellatrix:
		if b.IsBlinded() {
			return &validatorpb.SignRequest_BlindedBlockBellatrix{BlindedBlockBellatrix: pb.(*sila.BlindedBeaconBlockBellatrix)}, nil
		}
		return &validatorpb.SignRequest_BlockBellatrix{BlockBellatrix: pb.(*sila.BeaconBlockBellatrix)}, nil
	case version.Capella:
		if b.IsBlinded() {
			return &validatorpb.SignRequest_BlindedBlockCapella{BlindedBlockCapella: pb.(*sila.BlindedBeaconBlockCapella)}, nil
		}
		return &validatorpb.SignRequest_BlockCapella{BlockCapella: pb.(*sila.BeaconBlockCapella)}, nil
	case version.Deneb:
		if b.IsBlinded() {
			return &validatorpb.SignRequest_BlindedBlockDeneb{BlindedBlockDeneb: pb.(*sila.BlindedBeaconBlockDeneb)}, nil
		}
		return &validatorpb.SignRequest_BlockDeneb{BlockDeneb: pb.(*sila.BeaconBlockDeneb)}, nil
	case version.Electra:
		if b.IsBlinded() {
			return &validatorpb.SignRequest_BlindedBlockElectra{BlindedBlockElectra: pb.(*sila.BlindedBeaconBlockElectra)}, nil
		}
		return &validatorpb.SignRequest_BlockElectra{BlockElectra: pb.(*sila.BeaconBlockElectra)}, nil
	case version.Fulu:
		if b.IsBlinded() {
			return &validatorpb.SignRequest_BlindedBlockFulu{BlindedBlockFulu: pb.(*sila.BlindedBeaconBlockFulu)}, nil
		}
		return &validatorpb.SignRequest_BlockFulu{BlockFulu: pb.(*sila.BeaconBlockElectra)}, nil
	case version.Gloas:
		return &validatorpb.SignRequest_BlockGloas{BlockGloas: pb.(*sila.BeaconBlockGloas)}, nil
	default:
		return nil, errIncorrectBlockVersion
	}
}

// IsNil checks if the block body is nil.
func (b *BeaconBlockBody) IsNil() bool {
	return b == nil
}

// RandaoReveal returns the randao reveal from the block body.
func (b *BeaconBlockBody) RandaoReveal() [field_params.BLSSignatureLength]byte {
	return b.randaoReveal
}

// SilaData returns the silaexec data in the block.
func (b *BeaconBlockBody) SilaChainData() *sila.SilaData {
	return b.silaexecData
}

// Graffiti returns the graffiti in the block.
func (b *BeaconBlockBody) Graffiti() [field_params.RootLength]byte {
	return b.graffiti
}

// ProposerSlashings returns the proposer slashings in the block.
func (b *BeaconBlockBody) ProposerSlashings() []*sila.ProposerSlashing {
	return b.proposerSlashings
}

// AttesterSlashings returns the attester slashings in the block.
func (b *BeaconBlockBody) AttesterSlashings() []sila.AttSlashing {
	var slashings []sila.AttSlashing
	if b.version < version.Electra {
		if b.attesterSlashings == nil {
			return nil
		}
		slashings = make([]sila.AttSlashing, len(b.attesterSlashings))
		for i, s := range b.attesterSlashings {
			slashings[i] = s
		}
	} else {
		if b.attesterSlashingsElectra == nil {
			return nil
		}
		slashings = make([]sila.AttSlashing, len(b.attesterSlashingsElectra))
		for i, s := range b.attesterSlashingsElectra {
			slashings[i] = s
		}
	}
	return slashings
}

// Attestations returns the stored attestations in the block.
func (b *BeaconBlockBody) Attestations() []sila.Att {
	var atts []sila.Att
	if b.version < version.Electra {
		if b.attestations == nil {
			return nil
		}
		atts = make([]sila.Att, len(b.attestations))
		for i, a := range b.attestations {
			atts[i] = a
		}
	} else {
		if b.attestationsElectra == nil {
			return nil
		}
		atts = make([]sila.Att, len(b.attestationsElectra))
		for i, a := range b.attestationsElectra {
			atts[i] = a
		}
	}
	return atts
}

// Deposits returns the stored deposits in the block.
func (b *BeaconBlockBody) Deposits() []*sila.Deposit {
	return b.deposits
}

// VoluntaryExits returns the voluntary exits in the block.
func (b *BeaconBlockBody) VoluntaryExits() []*sila.SignedVoluntaryExit {
	return b.voluntaryExits
}

// SyncAggregate returns the sync aggregate in the block.
func (b *BeaconBlockBody) SyncAggregate() (*sila.SyncAggregate, error) {
	if b.version == version.Phase0 {
		return nil, consensus_types.ErrNotSupported("SyncAggregate", b.version)
	}
	return b.syncAggregate, nil
}

// Execution returns the sila payload of the block body.
func (b *BeaconBlockBody) SilaData() (interfaces.SilaData, error) {
	if b.version <= version.Altair || b.version >= version.Gloas {
		return nil, consensus_types.ErrNotSupported("Execution", b.version)
	}
	if b.IsBlinded() {
		return b.silaPayloadHeader, nil
	}
	return b.silaPayload, nil
}

func (b *BeaconBlockBody) BLSToSilaChanges() ([]*sila.SignedBLSToSilaChange, error) {
	if b.version < version.Capella {
		return nil, consensus_types.ErrNotSupported("BLSToSilaChanges", b.version)
	}
	return b.blsToSilaChanges, nil
}

// BlobKzgCommitments returns the blob kzg commitments in the block.
func (b *BeaconBlockBody) BlobKzgCommitments() ([][]byte, error) {
	if b.version >= version.Gloas {
		signedBid, err := b.SignedSilaPayloadBid()
		if err != nil {
			return nil, err
		}
		return signedBid.Message.BlobKzgCommitments, nil
	}
	if b.version >= version.Deneb {
		return b.blobKzgCommitments, nil
	}
	if b.version >= version.Phase0 {
		return nil, consensus_types.ErrNotSupported("BlobKzgCommitments", b.version)
	}
	return nil, errIncorrectBlockVersion
}

// SilaRequests returns the sila requests
func (b *BeaconBlockBody) SilaRequests() (*silaenginev1.SilaRequests, error) {
	if b.version < version.Electra || b.version >= version.Gloas {
		return nil, consensus_types.ErrNotSupported("SilaRequests", b.version)
	}
	return b.silaRequests, nil
}

// PayloadAttestations returns the payload attestations in the block.
func (b *BeaconBlockBody) PayloadAttestations() ([]*sila.PayloadAttestation, error) {
	if b.version >= version.Gloas {
		return b.payloadAttestations, nil
	}
	return nil, consensus_types.ErrNotSupported("PayloadAttestations", b.version)
}

// SignedSilaPayloadBid returns the signed sila payload header in the block.
func (b *BeaconBlockBody) SignedSilaPayloadBid() (*sila.SignedSilaPayloadBid, error) {
	if b.version >= version.Gloas {
		return b.signedSilaPayloadBid, nil
	}
	return nil, consensus_types.ErrNotSupported("SignedSilaPayloadBid", b.version)
}

// ParentSilaRequests returns the parent's deferred sila requests.
func (b *BeaconBlockBody) ParentSilaRequests() (*silaenginev1.SilaRequests, error) {
	if b.version >= version.Gloas {
		return b.parentSilaRequests, nil
	}
	return nil, consensus_types.ErrNotSupported("ParentSilaRequests", b.version)
}

// Version returns the version of the beacon block body
func (b *BeaconBlockBody) Version() int {
	return b.version
}

// HashTreeRoot returns the ssz root of the block body.
func (b *BeaconBlockBody) HashTreeRoot() ([field_params.RootLength]byte, error) {
	pb, err := b.Proto()
	if err != nil {
		return [field_params.RootLength]byte{}, err
	}
	switch b.version {
	case version.Phase0:
		return pb.(*sila.BeaconBlockBody).HashTreeRoot()
	case version.Altair:
		return pb.(*sila.BeaconBlockBodyAltair).HashTreeRoot()
	case version.Bellatrix:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockBodyBellatrix).HashTreeRoot()
		}
		return pb.(*sila.BeaconBlockBodyBellatrix).HashTreeRoot()
	case version.Capella:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockBodyCapella).HashTreeRoot()
		}
		return pb.(*sila.BeaconBlockBodyCapella).HashTreeRoot()
	case version.Deneb:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockBodyDeneb).HashTreeRoot()
		}
		return pb.(*sila.BeaconBlockBodyDeneb).HashTreeRoot()
	case version.Electra:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockBodyElectra).HashTreeRoot()
		}
		return pb.(*sila.BeaconBlockBodyElectra).HashTreeRoot()
	case version.Fulu:
		if b.IsBlinded() {
			return pb.(*sila.BlindedBeaconBlockBodyElectra).HashTreeRoot()
		}
		return pb.(*sila.BeaconBlockBodyElectra).HashTreeRoot()
	case version.Gloas:
		return pb.(*sila.BeaconBlockBodyGloas).HashTreeRoot()
	default:
		return [field_params.RootLength]byte{}, errIncorrectBodyVersion
	}
}

// IsBlinded checks if the beacon block body is a blinded block body.
func (b *BeaconBlockBody) IsBlinded() bool {
	return b.version < version.Gloas && b.version >= version.Bellatrix && b.silaPayload == nil
}
