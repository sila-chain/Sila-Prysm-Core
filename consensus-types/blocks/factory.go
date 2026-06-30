package blocks

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	sila "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
)

var (
	// ErrUnsupportedSignedBeaconBlock is returned when the struct type is not a supported signed
	// beacon block type.
	ErrUnsupportedSignedBeaconBlock = errors.New("unsupported signed beacon block")
	// errUnsupportedBeaconBlock is returned when the struct type is not a supported beacon block
	// type.
	errUnsupportedBeaconBlock = errors.New("unsupported beacon block")
	// errUnsupportedBeaconBlockBody is returned when the struct type is not a supported beacon block body
	// type.
	errUnsupportedBeaconBlockBody = errors.New("unsupported beacon block body")
	// ErrNilObject is returned in a constructor when the underlying object is nil.
	ErrNilObject = errors.New("received nil object")
	// ErrNilSignedBeaconBlock is returned when a nil signed beacon block is received.
	ErrNilSignedBeaconBlock = errors.New("signed beacon block can't be nil")
	// ErrNilBeaconBlock is returned when a nil beacon block is received.
	ErrNilBeaconBlock              = errors.New("beacon block can't be nil")
	errNonBlindedSignedBeaconBlock = errors.New("can only build signed beacon block from blinded format")
)

// NewSignedBeaconBlock creates a signed beacon block from a protobuf signed beacon block.
func NewSignedBeaconBlock(i any) (interfaces.SignedBeaconBlock, error) {
	switch b := i.(type) {
	case nil:
		return nil, ErrNilObject
	case *sila.GenericSignedBeaconBlock_Phase0:
		return initSignedBlockFromProtoPhase0(b.Phase0)
	case *sila.SignedBeaconBlock:
		return initSignedBlockFromProtoPhase0(b)
	case *sila.GenericSignedBeaconBlock_Altair:
		return initSignedBlockFromProtoAltair(b.Altair)
	case *sila.SignedBeaconBlockAltair:
		return initSignedBlockFromProtoAltair(b)
	case *sila.GenericSignedBeaconBlock_Bellatrix:
		return initSignedBlockFromProtoBellatrix(b.Bellatrix)
	case *sila.SignedBeaconBlockBellatrix:
		return initSignedBlockFromProtoBellatrix(b)
	case *sila.GenericSignedBeaconBlock_BlindedBellatrix:
		return initBlindedSignedBlockFromProtoBellatrix(b.BlindedBellatrix)
	case *sila.SignedBlindedBeaconBlockBellatrix:
		return initBlindedSignedBlockFromProtoBellatrix(b)
	case *sila.GenericSignedBeaconBlock_Capella:
		return initSignedBlockFromProtoCapella(b.Capella)
	case *sila.SignedBeaconBlockCapella:
		return initSignedBlockFromProtoCapella(b)
	case *sila.GenericSignedBeaconBlock_BlindedCapella:
		return initBlindedSignedBlockFromProtoCapella(b.BlindedCapella)
	case *sila.SignedBlindedBeaconBlockCapella:
		return initBlindedSignedBlockFromProtoCapella(b)
	case *sila.GenericSignedBeaconBlock_Deneb:
		return initSignedBlockFromProtoDeneb(b.Deneb.Block)
	case *sila.SignedBeaconBlockDeneb:
		return initSignedBlockFromProtoDeneb(b)
	case *sila.SignedBlindedBeaconBlockDeneb:
		return initBlindedSignedBlockFromProtoDeneb(b)
	case *sila.GenericSignedBeaconBlock_BlindedDeneb:
		return initBlindedSignedBlockFromProtoDeneb(b.BlindedDeneb)
	case *sila.GenericSignedBeaconBlock_Electra:
		return initSignedBlockFromProtoElectra(b.Electra.Block)
	case *sila.SignedBeaconBlockElectra:
		return initSignedBlockFromProtoElectra(b)
	case *sila.SignedBlindedBeaconBlockElectra:
		return initBlindedSignedBlockFromProtoElectra(b)
	case *sila.GenericSignedBeaconBlock_BlindedElectra:
		return initBlindedSignedBlockFromProtoElectra(b.BlindedElectra)
	case *sila.GenericSignedBeaconBlock_Fulu:
		return initSignedBlockFromProtoFulu(b.Fulu.Block)
	case *sila.SignedBeaconBlockFulu:
		return initSignedBlockFromProtoFulu(b)
	case *sila.SignedBlindedBeaconBlockFulu:
		return initBlindedSignedBlockFromProtoFulu(b)
	case *sila.GenericSignedBeaconBlock_BlindedFulu:
		return initBlindedSignedBlockFromProtoFulu(b.BlindedFulu)
	case *sila.GenericSignedBeaconBlock_Gloas:
		return initSignedBlockFromProtoGloas(b.Gloas)
	case *sila.SignedBeaconBlockGloas:
		return initSignedBlockFromProtoGloas(b)
	default:
		return nil, errors.Wrapf(ErrUnsupportedSignedBeaconBlock, "unable to create block from type %T", i)
	}
}

// NewBeaconBlock creates a beacon block from a protobuf beacon block.
func NewBeaconBlock(i any) (interfaces.ReadOnlyBeaconBlock, error) {
	switch b := i.(type) {
	case nil:
		return nil, ErrNilObject
	case *sila.GenericBeaconBlock_Phase0:
		return initBlockFromProtoPhase0(b.Phase0)
	case *sila.BeaconBlock:
		return initBlockFromProtoPhase0(b)
	case *sila.GenericBeaconBlock_Altair:
		return initBlockFromProtoAltair(b.Altair)
	case *sila.BeaconBlockAltair:
		return initBlockFromProtoAltair(b)
	case *sila.GenericBeaconBlock_Bellatrix:
		return initBlockFromProtoBellatrix(b.Bellatrix)
	case *sila.BeaconBlockBellatrix:
		return initBlockFromProtoBellatrix(b)
	case *sila.GenericBeaconBlock_BlindedBellatrix:
		return initBlindedBlockFromProtoBellatrix(b.BlindedBellatrix)
	case *sila.BlindedBeaconBlockBellatrix:
		return initBlindedBlockFromProtoBellatrix(b)
	case *sila.GenericBeaconBlock_Capella:
		return initBlockFromProtoCapella(b.Capella)
	case *sila.BeaconBlockCapella:
		return initBlockFromProtoCapella(b)
	case *sila.GenericBeaconBlock_BlindedCapella:
		return initBlindedBlockFromProtoCapella(b.BlindedCapella)
	case *sila.BlindedBeaconBlockCapella:
		return initBlindedBlockFromProtoCapella(b)
	case *sila.GenericBeaconBlock_Deneb:
		return initBlockFromProtoDeneb(b.Deneb.Block)
	case *sila.BeaconBlockDeneb:
		return initBlockFromProtoDeneb(b)
	case *sila.BlindedBeaconBlockDeneb:
		return initBlindedBlockFromProtoDeneb(b)
	case *sila.GenericBeaconBlock_BlindedDeneb:
		return initBlindedBlockFromProtoDeneb(b.BlindedDeneb)
	case *sila.GenericBeaconBlock_Electra:
		return initBlockFromProtoElectra(b.Electra.Block)
	case *sila.BeaconBlockElectra:
		return initBlockFromProtoElectra(b)
	case *sila.BlindedBeaconBlockElectra:
		return initBlindedBlockFromProtoElectra(b)
	case *sila.GenericBeaconBlock_BlindedElectra:
		return initBlindedBlockFromProtoElectra(b.BlindedElectra)
	case *sila.GenericBeaconBlock_Fulu:
		return initBlockFromProtoFulu(b.Fulu.Block)
	case *sila.BlindedBeaconBlockFulu:
		return initBlindedBlockFromProtoFulu(b)
	case *sila.GenericBeaconBlock_BlindedFulu:
		return initBlindedBlockFromProtoFulu(b.BlindedFulu)
	case *sila.GenericBeaconBlock_Gloas:
		return initBlockFromProtoGloas(b.Gloas)
	case *sila.BeaconBlockGloas:
		return initBlockFromProtoGloas(b)
	default:
		return nil, errors.Wrapf(errUnsupportedBeaconBlock, "unable to create block from type %T", i)
	}
}

// NewBeaconBlockBody creates a beacon block body from a protobuf beacon block body.
func NewBeaconBlockBody(i any) (interfaces.ReadOnlyBeaconBlockBody, error) {
	switch b := i.(type) {
	case nil:
		return nil, ErrNilObject
	case *sila.BeaconBlockBody:
		return initBlockBodyFromProtoPhase0(b)
	case *sila.BeaconBlockBodyAltair:
		return initBlockBodyFromProtoAltair(b)
	case *sila.BeaconBlockBodyBellatrix:
		return initBlockBodyFromProtoBellatrix(b)
	case *sila.BlindedBeaconBlockBodyBellatrix:
		return initBlindedBlockBodyFromProtoBellatrix(b)
	case *sila.BeaconBlockBodyCapella:
		return initBlockBodyFromProtoCapella(b)
	case *sila.BlindedBeaconBlockBodyCapella:
		return initBlindedBlockBodyFromProtoCapella(b)
	case *sila.BeaconBlockBodyDeneb:
		return initBlockBodyFromProtoDeneb(b)
	case *sila.BlindedBeaconBlockBodyDeneb:
		return initBlindedBlockBodyFromProtoDeneb(b)
	case *sila.BeaconBlockBodyElectra:
		return initBlockBodyFromProtoElectra(b)
	case *sila.BlindedBeaconBlockBodyElectra:
		return initBlindedBlockBodyFromProtoElectra(b)
	case *sila.BeaconBlockBodyGloas:
		return initBlockBodyFromProtoGloas(b)
	default:
		return nil, errors.Wrapf(errUnsupportedBeaconBlockBody, "unable to create block body from type %T", i)
	}
}

// BuildSignedBeaconBlock assembles a block.ReadOnlySignedBeaconBlock interface compatible struct from a
// given beacon block and the appropriate signature. This method may be used to easily create a
// signed beacon block.
func BuildSignedBeaconBlock(blk interfaces.ReadOnlyBeaconBlock, signature []byte) (interfaces.SignedBeaconBlock, error) {
	pb, err := blk.Proto()
	if err != nil {
		return nil, err
	}

	switch blk.Version() {
	case version.Phase0:
		pb, ok := pb.(*sila.BeaconBlock)
		if !ok {
			return nil, errIncorrectBlockVersion
		}
		return NewSignedBeaconBlock(&sila.SignedBeaconBlock{Block: pb, Signature: signature})
	case version.Altair:
		pb, ok := pb.(*sila.BeaconBlockAltair)
		if !ok {
			return nil, errIncorrectBlockVersion
		}
		return NewSignedBeaconBlock(&sila.SignedBeaconBlockAltair{Block: pb, Signature: signature})
	case version.Bellatrix:
		if blk.IsBlinded() {
			pb, ok := pb.(*sila.BlindedBeaconBlockBellatrix)
			if !ok {
				return nil, errIncorrectBlockVersion
			}
			return NewSignedBeaconBlock(&sila.SignedBlindedBeaconBlockBellatrix{Block: pb, Signature: signature})
		}
		pb, ok := pb.(*sila.BeaconBlockBellatrix)
		if !ok {
			return nil, errIncorrectBlockVersion
		}
		return NewSignedBeaconBlock(&sila.SignedBeaconBlockBellatrix{Block: pb, Signature: signature})
	case version.Capella:
		if blk.IsBlinded() {
			pb, ok := pb.(*sila.BlindedBeaconBlockCapella)
			if !ok {
				return nil, errIncorrectBlockVersion
			}
			return NewSignedBeaconBlock(&sila.SignedBlindedBeaconBlockCapella{Block: pb, Signature: signature})
		}
		pb, ok := pb.(*sila.BeaconBlockCapella)
		if !ok {
			return nil, errIncorrectBlockVersion
		}
		return NewSignedBeaconBlock(&sila.SignedBeaconBlockCapella{Block: pb, Signature: signature})
	case version.Deneb:
		if blk.IsBlinded() {
			pb, ok := pb.(*sila.BlindedBeaconBlockDeneb)
			if !ok {
				return nil, errIncorrectBlockVersion
			}
			return NewSignedBeaconBlock(&sila.SignedBlindedBeaconBlockDeneb{Message: pb, Signature: signature})
		}
		pb, ok := pb.(*sila.BeaconBlockDeneb)
		if !ok {
			return nil, errIncorrectBlockVersion
		}
		return NewSignedBeaconBlock(&sila.SignedBeaconBlockDeneb{Block: pb, Signature: signature})
	case version.Electra:
		if blk.IsBlinded() {
			pb, ok := pb.(*sila.BlindedBeaconBlockElectra)
			if !ok {
				return nil, errIncorrectBlockVersion
			}
			return NewSignedBeaconBlock(&sila.SignedBlindedBeaconBlockElectra{Message: pb, Signature: signature})
		}
		pb, ok := pb.(*sila.BeaconBlockElectra)
		if !ok {
			return nil, errIncorrectBlockVersion
		}
		return NewSignedBeaconBlock(&sila.SignedBeaconBlockElectra{Block: pb, Signature: signature})
	case version.Fulu:
		if blk.IsBlinded() {
			pb, ok := pb.(*sila.BlindedBeaconBlockFulu)
			if !ok {
				return nil, errIncorrectBlockVersion
			}
			return NewSignedBeaconBlock(&sila.SignedBlindedBeaconBlockFulu{Message: pb, Signature: signature})
		}
		pb, ok := pb.(*sila.BeaconBlockElectra)
		if !ok {
			return nil, errIncorrectBlockVersion
		}
		return NewSignedBeaconBlock(&sila.SignedBeaconBlockFulu{Block: pb, Signature: signature})
	case version.Gloas:
		pb, ok := pb.(*sila.BeaconBlockGloas)
		if !ok {
			return nil, errIncorrectBlockVersion
		}
		return NewSignedBeaconBlock(&sila.SignedBeaconBlockGloas{Block: pb, Signature: signature})
	default:
		return nil, errUnsupportedBeaconBlock
	}
}

func getWrappedPayload(payload any) (wrappedPayload interfaces.SilaData, wrapErr error) {
	switch p := payload.(type) {
	case *silaenginev1.SilaPayload:
		wrappedPayload, wrapErr = WrappedSilaPayload(p)
	case *silaenginev1.SilaPayloadCapella:
		wrappedPayload, wrapErr = WrappedSilaPayloadCapella(p)
	case *silaenginev1.SilaPayloadDeneb:
		wrappedPayload, wrapErr = WrappedSilaPayloadDeneb(p)
	default:
		wrappedPayload, wrapErr = nil, fmt.Errorf("%T is not a type of sila payload", p)
	}
	return wrappedPayload, wrapErr
}

func checkPayloadAgainstHeader(wrappedPayload, payloadHeader interfaces.SilaData) error {
	empty, err := IsEmptySilaData(wrappedPayload)
	if err != nil {
		return err
	}
	if empty {
		return nil
	}
	payloadRoot, err := wrappedPayload.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "could not hash tree root sila payload")
	}
	payloadHeaderRoot, err := payloadHeader.HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "could not hash tree root payload header")
	}
	if payloadRoot != payloadHeaderRoot {
		return fmt.Errorf(
			"payload %#x and header %#x roots do not match",
			payloadRoot,
			payloadHeaderRoot,
		)
	}
	return nil
}

// BuildSignedBeaconBlockFromSilaPayload takes a signed, blinded beacon block and converts into
// a full, signed beacon block by specifying a sila payload.
// nolint:gocognit
func BuildSignedBeaconBlockFromSilaPayload(blk interfaces.ReadOnlySignedBeaconBlock, payload any) (interfaces.SignedBeaconBlock, error) {
	if err := BeaconBlockIsNil(blk); err != nil {
		return nil, err
	}
	if !blk.IsBlinded() {
		return nil, errNonBlindedSignedBeaconBlock
	}
	b := blk.Block()
	payloadHeader, err := b.Body().SilaData()
	if err != nil {
		return nil, errors.Wrap(err, "could not get sila payload header")
	}

	wrappedPayload, err := getWrappedPayload(payload)
	if err != nil {
		return nil, err
	}
	if err := checkPayloadAgainstHeader(wrappedPayload, payloadHeader); err != nil {
		return nil, err
	}
	syncAgg, err := b.Body().SyncAggregate()
	if err != nil {
		return nil, errors.Wrap(err, "could not get sync aggregate from block body")
	}
	parentRoot := b.ParentRoot()
	stateRoot := b.StateRoot()
	randaoReveal := b.Body().RandaoReveal()
	graffiti := b.Body().Graffiti()
	sig := blk.Signature()

	var fullBlock any
	switch blk.Version() {
	case version.Bellatrix:
		p, ok := payload.(*silaenginev1.SilaPayload)
		if !ok {
			return nil, fmt.Errorf("payload has wrong type (expected %T, got %T)", &silaenginev1.SilaPayload{}, payload)
		}
		var atts []*sila.Attestation
		if b.Body().Attestations() != nil {
			atts = make([]*sila.Attestation, len(b.Body().Attestations()))
			for i, att := range b.Body().Attestations() {
				a, ok := att.(*sila.Attestation)
				if !ok {
					return nil, fmt.Errorf("attestation has wrong type (expected %T, got %T)", &sila.Attestation{}, att)
				}
				atts[i] = a
			}
		}
		var attSlashings []*sila.AttesterSlashing
		if b.Body().AttesterSlashings() != nil {
			attSlashings = make([]*sila.AttesterSlashing, len(b.Body().AttesterSlashings()))
			for i, slashing := range b.Body().AttesterSlashings() {
				s, ok := slashing.(*sila.AttesterSlashing)
				if !ok {
					return nil, fmt.Errorf("attester slashing has wrong type (expected %T, got %T)", &sila.AttesterSlashing{}, slashing)
				}
				attSlashings[i] = s
			}
		}
		fullBlock = &sila.SignedBeaconBlockBellatrix{
			Block: &sila.BeaconBlockBellatrix{
				Slot:          b.Slot(),
				ProposerIndex: b.ProposerIndex(),
				ParentRoot:    parentRoot[:],
				StateRoot:     stateRoot[:],
				Body: &sila.BeaconBlockBodyBellatrix{
					RandaoReveal:      randaoReveal[:],
					SilaData:          b.Body().SilaChainData(),
					Graffiti:          graffiti[:],
					ProposerSlashings: b.Body().ProposerSlashings(),
					AttesterSlashings: attSlashings,
					Attestations:      atts,
					Deposits:          b.Body().Deposits(),
					VoluntaryExits:    b.Body().VoluntaryExits(),
					SyncAggregate:     syncAgg,
					SilaPayload:       p,
				},
			},
			Signature: sig[:],
		}
	case version.Capella:
		p, ok := payload.(*silaenginev1.SilaPayloadCapella)
		if !ok {
			return nil, fmt.Errorf("payload has wrong type (expected %T, got %T)", &silaenginev1.SilaPayloadCapella{}, payload)
		}
		blsToSilaChanges, err := b.Body().BLSToSilaChanges()
		if err != nil {
			return nil, err
		}
		var atts []*sila.Attestation
		if b.Body().Attestations() != nil {
			atts = make([]*sila.Attestation, len(b.Body().Attestations()))
			for i, att := range b.Body().Attestations() {
				a, ok := att.(*sila.Attestation)
				if !ok {
					return nil, fmt.Errorf("attestation has wrong type (expected %T, got %T)", &sila.Attestation{}, att)
				}
				atts[i] = a
			}
		}
		var attSlashings []*sila.AttesterSlashing
		if b.Body().AttesterSlashings() != nil {
			attSlashings = make([]*sila.AttesterSlashing, len(b.Body().AttesterSlashings()))
			for i, slashing := range b.Body().AttesterSlashings() {
				s, ok := slashing.(*sila.AttesterSlashing)
				if !ok {
					return nil, fmt.Errorf("attester slashing has wrong type (expected %T, got %T)", &sila.AttesterSlashing{}, slashing)
				}
				attSlashings[i] = s
			}
		}
		fullBlock = &sila.SignedBeaconBlockCapella{
			Block: &sila.BeaconBlockCapella{
				Slot:          b.Slot(),
				ProposerIndex: b.ProposerIndex(),
				ParentRoot:    parentRoot[:],
				StateRoot:     stateRoot[:],
				Body: &sila.BeaconBlockBodyCapella{
					RandaoReveal:      randaoReveal[:],
					SilaData:          b.Body().SilaChainData(),
					Graffiti:          graffiti[:],
					ProposerSlashings: b.Body().ProposerSlashings(),
					AttesterSlashings: attSlashings,
					Attestations:      atts,
					Deposits:          b.Body().Deposits(),
					VoluntaryExits:    b.Body().VoluntaryExits(),
					SyncAggregate:     syncAgg,
					SilaPayload:       p,
					BlsToSilaChanges:  blsToSilaChanges,
				},
			},
			Signature: sig[:],
		}
	case version.Deneb:
		p, ok := payload.(*silaenginev1.SilaPayloadDeneb)
		if !ok {
			return nil, fmt.Errorf("payload has wrong type (expected %T, got %T)", &silaenginev1.SilaPayloadDeneb{}, payload)
		}
		blsToSilaChanges, err := b.Body().BLSToSilaChanges()
		if err != nil {
			return nil, err
		}
		commitments, err := b.Body().BlobKzgCommitments()
		if err != nil {
			return nil, err
		}
		var atts []*sila.Attestation
		if b.Body().Attestations() != nil {
			atts = make([]*sila.Attestation, len(b.Body().Attestations()))
			for i, att := range b.Body().Attestations() {
				a, ok := att.(*sila.Attestation)
				if !ok {
					return nil, fmt.Errorf("attestation has wrong type (expected %T, got %T)", &sila.Attestation{}, att)
				}
				atts[i] = a
			}
		}
		var attSlashings []*sila.AttesterSlashing
		if b.Body().AttesterSlashings() != nil {
			attSlashings = make([]*sila.AttesterSlashing, len(b.Body().AttesterSlashings()))
			for i, slashing := range b.Body().AttesterSlashings() {
				s, ok := slashing.(*sila.AttesterSlashing)
				if !ok {
					return nil, fmt.Errorf("attester slashing has wrong type (expected %T, got %T)", &sila.AttesterSlashing{}, slashing)
				}
				attSlashings[i] = s
			}
		}
		fullBlock = &sila.SignedBeaconBlockDeneb{
			Block: &sila.BeaconBlockDeneb{
				Slot:          b.Slot(),
				ProposerIndex: b.ProposerIndex(),
				ParentRoot:    parentRoot[:],
				StateRoot:     stateRoot[:],
				Body: &sila.BeaconBlockBodyDeneb{
					RandaoReveal:       randaoReveal[:],
					SilaData:           b.Body().SilaChainData(),
					Graffiti:           graffiti[:],
					ProposerSlashings:  b.Body().ProposerSlashings(),
					AttesterSlashings:  attSlashings,
					Attestations:       atts,
					Deposits:           b.Body().Deposits(),
					VoluntaryExits:     b.Body().VoluntaryExits(),
					SyncAggregate:      syncAgg,
					SilaPayload:        p,
					BlsToSilaChanges:   blsToSilaChanges,
					BlobKzgCommitments: commitments,
				},
			},
			Signature: sig[:],
		}
	case version.Electra:
		p, ok := payload.(*silaenginev1.SilaPayloadDeneb)
		if !ok {
			return nil, fmt.Errorf("payload has wrong type (expected %T, got %T)", &silaenginev1.SilaPayloadDeneb{}, payload)
		}
		blsToSilaChanges, err := b.Body().BLSToSilaChanges()
		if err != nil {
			return nil, err
		}
		commitments, err := b.Body().BlobKzgCommitments()
		if err != nil {
			return nil, err
		}
		var atts []*sila.AttestationElectra
		if b.Body().Attestations() != nil {
			atts = make([]*sila.AttestationElectra, len(b.Body().Attestations()))
			for i, att := range b.Body().Attestations() {
				a, ok := att.(*sila.AttestationElectra)
				if !ok {
					return nil, fmt.Errorf("attestation has wrong type (expected %T, got %T)", &sila.AttestationElectra{}, att)
				}
				atts[i] = a
			}
		}
		var attSlashings []*sila.AttesterSlashingElectra
		if b.Body().AttesterSlashings() != nil {
			attSlashings = make([]*sila.AttesterSlashingElectra, len(b.Body().AttesterSlashings()))
			for i, slashing := range b.Body().AttesterSlashings() {
				s, ok := slashing.(*sila.AttesterSlashingElectra)
				if !ok {
					return nil, fmt.Errorf("attester slashing has wrong type (expected %T, got %T)", &sila.AttesterSlashingElectra{}, slashing)
				}
				attSlashings[i] = s
			}
		}

		er, err := b.Body().SilaRequests()
		if err != nil {
			return nil, err
		}

		fullBlock = &sila.SignedBeaconBlockElectra{
			Block: &sila.BeaconBlockElectra{
				Slot:          b.Slot(),
				ProposerIndex: b.ProposerIndex(),
				ParentRoot:    parentRoot[:],
				StateRoot:     stateRoot[:],
				Body: &sila.BeaconBlockBodyElectra{
					RandaoReveal:       randaoReveal[:],
					SilaData:           b.Body().SilaChainData(),
					Graffiti:           graffiti[:],
					ProposerSlashings:  b.Body().ProposerSlashings(),
					AttesterSlashings:  attSlashings,
					Attestations:       atts,
					Deposits:           b.Body().Deposits(),
					VoluntaryExits:     b.Body().VoluntaryExits(),
					SyncAggregate:      syncAgg,
					SilaPayload:        p,
					BlsToSilaChanges:   blsToSilaChanges,
					BlobKzgCommitments: commitments,
					SilaRequests:       er,
				},
			},
			Signature: sig[:],
		}
	case version.Fulu:
		p, ok := payload.(*silaenginev1.SilaPayloadDeneb)
		if !ok {
			return nil, fmt.Errorf("payload has wrong type (expected %T, got %T)", &silaenginev1.SilaPayloadDeneb{}, payload)
		}
		blsToSilaChanges, err := b.Body().BLSToSilaChanges()
		if err != nil {
			return nil, err
		}
		commitments, err := b.Body().BlobKzgCommitments()
		if err != nil {
			return nil, err
		}
		var atts []*sila.AttestationElectra
		if b.Body().Attestations() != nil {
			atts = make([]*sila.AttestationElectra, len(b.Body().Attestations()))
			for i, att := range b.Body().Attestations() {
				a, ok := att.(*sila.AttestationElectra)
				if !ok {
					return nil, fmt.Errorf("attestation has wrong type (expected %T, got %T)", &sila.Attestation{}, att)
				}
				atts[i] = a
			}
		}
		var attSlashings []*sila.AttesterSlashingElectra
		if b.Body().AttesterSlashings() != nil {
			attSlashings = make([]*sila.AttesterSlashingElectra, len(b.Body().AttesterSlashings()))
			for i, slashing := range b.Body().AttesterSlashings() {
				s, ok := slashing.(*sila.AttesterSlashingElectra)
				if !ok {
					return nil, fmt.Errorf("attester slashing has wrong type (expected %T, got %T)", &sila.AttesterSlashing{}, slashing)
				}
				attSlashings[i] = s
			}
		}

		er, err := b.Body().SilaRequests()
		if err != nil {
			return nil, err
		}

		fullBlock = &sila.SignedBeaconBlockFulu{
			Block: &sila.BeaconBlockElectra{
				Slot:          b.Slot(),
				ProposerIndex: b.ProposerIndex(),
				ParentRoot:    parentRoot[:],
				StateRoot:     stateRoot[:],
				Body: &sila.BeaconBlockBodyElectra{
					RandaoReveal:       randaoReveal[:],
					SilaData:           b.Body().SilaChainData(),
					Graffiti:           graffiti[:],
					ProposerSlashings:  b.Body().ProposerSlashings(),
					AttesterSlashings:  attSlashings,
					Attestations:       atts,
					Deposits:           b.Body().Deposits(),
					VoluntaryExits:     b.Body().VoluntaryExits(),
					SyncAggregate:      syncAgg,
					SilaPayload:        p,
					BlsToSilaChanges:   blsToSilaChanges,
					BlobKzgCommitments: commitments,
					SilaRequests:       er,
				},
			},
			Signature: sig[:],
		}
	case version.Gloas:
		return nil, errors.Wrap(errUnsupportedBeaconBlock, "gloas blocks are not supported in this function")
	default:
		return nil, errors.New("Block not of known type")
	}

	return NewSignedBeaconBlock(fullBlock)
}

// BeaconBlockContainerToSignedBeaconBlock converts BeaconBlockContainer (API response) to a SignedBeaconBlock.
// This is particularly useful for using the values from API calls.
func BeaconBlockContainerToSignedBeaconBlock(obj *sila.BeaconBlockContainer) (interfaces.ReadOnlySignedBeaconBlock, error) {
	switch obj.Block.(type) {
	case *sila.BeaconBlockContainer_BlindedFuluBlock:
		return NewSignedBeaconBlock(obj.GetBlindedFuluBlock())
	case *sila.BeaconBlockContainer_FuluBlock:
		return NewSignedBeaconBlock(obj.GetFuluBlock())
	case *sila.BeaconBlockContainer_BlindedElectraBlock:
		return NewSignedBeaconBlock(obj.GetBlindedElectraBlock())
	case *sila.BeaconBlockContainer_ElectraBlock:
		return NewSignedBeaconBlock(obj.GetElectraBlock())
	case *sila.BeaconBlockContainer_BlindedDenebBlock:
		return NewSignedBeaconBlock(obj.GetBlindedDenebBlock())
	case *sila.BeaconBlockContainer_DenebBlock:
		return NewSignedBeaconBlock(obj.GetDenebBlock())
	case *sila.BeaconBlockContainer_BlindedCapellaBlock:
		return NewSignedBeaconBlock(obj.GetBlindedCapellaBlock())
	case *sila.BeaconBlockContainer_CapellaBlock:
		return NewSignedBeaconBlock(obj.GetCapellaBlock())
	case *sila.BeaconBlockContainer_BlindedBellatrixBlock:
		return NewSignedBeaconBlock(obj.GetBlindedBellatrixBlock())
	case *sila.BeaconBlockContainer_BellatrixBlock:
		return NewSignedBeaconBlock(obj.GetBellatrixBlock())
	case *sila.BeaconBlockContainer_AltairBlock:
		return NewSignedBeaconBlock(obj.GetAltairBlock())
	case *sila.BeaconBlockContainer_Phase0Block:
		return NewSignedBeaconBlock(obj.GetPhase0Block())
	default:
		return nil, errors.New("container block type not recognized")
	}
}
