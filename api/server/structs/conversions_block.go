package structs

import (
	"fmt"
	"strconv"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/container/slice"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	eth "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila/common"
	"github.com/sila-chain/Sila/common/hexutil"
	"github.com/pkg/errors"
)

var ErrUnsupportedConversion = errors.New("Could not determine api struct type to use for value")

// ----------------------------------------------------------------------------
// Phase 0
// ----------------------------------------------------------------------------

func (h *SignedBeaconBlockHeader) ToConsensus() (*eth.SignedBeaconBlockHeader, error) {
	if h == nil {
		return nil, errNilValue
	}
	msg, err := h.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	sig, err := bytesutil.DecodeHexWithLength(h.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}

	return &eth.SignedBeaconBlockHeader{
		Header:    msg,
		Signature: sig,
	}, nil
}

func (h *BeaconBlockHeader) ToConsensus() (*eth.BeaconBlockHeader, error) {
	if h == nil {
		return nil, errNilValue
	}
	s, err := strconv.ParseUint(h.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	pi, err := strconv.ParseUint(h.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	pr, err := bytesutil.DecodeHexWithLength(h.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	sr, err := bytesutil.DecodeHexWithLength(h.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	br, err := bytesutil.DecodeHexWithLength(h.BodyRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "BodyRoot")
	}

	return &eth.BeaconBlockHeader{
		Slot:          primitives.Slot(s),
		ProposerIndex: primitives.ValidatorIndex(pi),
		ParentRoot:    pr,
		StateRoot:     sr,
		BodyRoot:      br,
	}, nil
}

func SignedBeaconBlockHeaderFromConsensus(src *eth.SignedBeaconBlockHeader) *SignedBeaconBlockHeader {
	return &SignedBeaconBlockHeader{
		Message: &BeaconBlockHeader{
			Slot:          fmt.Sprintf("%d", src.Header.Slot),
			ProposerIndex: fmt.Sprintf("%d", src.Header.ProposerIndex),
			ParentRoot:    hexutil.Encode(src.Header.ParentRoot),
			StateRoot:     hexutil.Encode(src.Header.StateRoot),
			BodyRoot:      hexutil.Encode(src.Header.BodyRoot),
		},
		Signature: hexutil.Encode(src.Signature),
	}
}

func (b *SignedBeaconBlock) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}

	bl, err := b.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}

	block := &eth.SignedBeaconBlock{
		Block:     bl,
		Signature: sig,
	}
	return &eth.GenericSignedBeaconBlock{Block: &eth.GenericSignedBeaconBlock_Phase0{Phase0: block}}, nil
}

func (b *BeaconBlock) ToGeneric() (*eth.GenericBeaconBlock, error) {
	block, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_Phase0{Phase0: block}}, nil
}

func (b *BeaconBlock) ToConsensus() (*eth.BeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}
	if b.Body == nil {
		return nil, server.NewDecodeError(errNilValue, "Body")
	}
	if b.Body.SilaData == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaData")
	}

	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	proposerIndex, err := strconv.ParseUint(b.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	parentRoot, err := bytesutil.DecodeHexWithLength(b.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	stateRoot, err := bytesutil.DecodeHexWithLength(b.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	randaoReveal, err := bytesutil.DecodeHexWithLength(b.Body.RandaoReveal, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.RandaoReveal")
	}
	depositRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.DepositRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositRoot")
	}
	depositCount, err := strconv.ParseUint(b.Body.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositCount")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.BlockHash")
	}
	graffiti, err := bytesutil.DecodeHexWithLength(b.Body.Graffiti, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Graffiti")
	}
	proposerSlashings, err := ProposerSlashingsToConsensus(b.Body.ProposerSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.ProposerSlashings")
	}
	attesterSlashings, err := AttesterSlashingsToConsensus(b.Body.AttesterSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.AttesterSlashings")
	}
	atts, err := AttsToConsensus(b.Body.Attestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Attestations")
	}
	deposits, err := DepositsToConsensus(b.Body.Deposits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Deposits")
	}
	exits, err := SignedExitsToConsensus(b.Body.VoluntaryExits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.VoluntaryExits")
	}

	return &eth.BeaconBlock{
		Slot:          primitives.Slot(slot),
		ProposerIndex: primitives.ValidatorIndex(proposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body: &eth.BeaconBlockBody{
			RandaoReveal: randaoReveal,
			SilaData: &eth.SilaData{
				DepositRoot:  depositRoot,
				DepositCount: depositCount,
				BlockHash:    blockHash,
			},
			Graffiti:          graffiti,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Attestations:      atts,
			Deposits:          deposits,
			VoluntaryExits:    exits,
		},
	}, nil
}

func BeaconBlockHeaderFromConsensus(h *eth.BeaconBlockHeader) *BeaconBlockHeader {
	return &BeaconBlockHeader{
		Slot:          fmt.Sprintf("%d", h.Slot),
		ProposerIndex: fmt.Sprintf("%d", h.ProposerIndex),
		ParentRoot:    hexutil.Encode(h.ParentRoot),
		StateRoot:     hexutil.Encode(h.StateRoot),
		BodyRoot:      hexutil.Encode(h.BodyRoot),
	}
}

func BeaconBlockFromConsensus(b *eth.BeaconBlock) *BeaconBlock {
	return &BeaconBlock{
		Slot:          fmt.Sprintf("%d", b.Slot),
		ProposerIndex: fmt.Sprintf("%d", b.ProposerIndex),
		ParentRoot:    hexutil.Encode(b.ParentRoot),
		StateRoot:     hexutil.Encode(b.StateRoot),
		Body: &BeaconBlockBody{
			RandaoReveal:      hexutil.Encode(b.Body.RandaoReveal),
			SilaData:          SilaDataFromConsensus(b.Body.SilaData),
			Graffiti:          hexutil.Encode(b.Body.Graffiti),
			ProposerSlashings: ProposerSlashingsFromConsensus(b.Body.ProposerSlashings),
			AttesterSlashings: AttesterSlashingsFromConsensus(b.Body.AttesterSlashings),
			Attestations:      AttsFromConsensus(b.Body.Attestations),
			Deposits:          DepositsFromConsensus(b.Body.Deposits),
			VoluntaryExits:    SignedExitsFromConsensus(b.Body.VoluntaryExits),
		},
	}
}

func SignedBeaconBlockMessageJsoner(block interfaces.ReadOnlySignedBeaconBlock) (SignedMessageJsoner, error) {
	pb, err := block.Proto()
	if err != nil {
		return nil, err
	}
	switch pbStruct := pb.(type) {
	case *eth.SignedBeaconBlock:
		return SignedBeaconBlockPhase0FromConsensus(pbStruct), nil
	case *eth.SignedBeaconBlockAltair:
		return SignedBeaconBlockAltairFromConsensus(pbStruct), nil
	case *eth.SignedBlindedBeaconBlockBellatrix:
		return SignedBlindedBeaconBlockBellatrixFromConsensus(pbStruct)
	case *eth.SignedBeaconBlockBellatrix:
		return SignedBeaconBlockBellatrixFromConsensus(pbStruct)
	case *eth.SignedBlindedBeaconBlockCapella:
		return SignedBlindedBeaconBlockCapellaFromConsensus(pbStruct)
	case *eth.SignedBeaconBlockCapella:
		return SignedBeaconBlockCapellaFromConsensus(pbStruct)
	case *eth.SignedBlindedBeaconBlockDeneb:
		return SignedBlindedBeaconBlockDenebFromConsensus(pbStruct)
	case *eth.SignedBeaconBlockDeneb:
		return SignedBeaconBlockDenebFromConsensus(pbStruct)
	case *eth.SignedBlindedBeaconBlockElectra:
		return SignedBlindedBeaconBlockElectraFromConsensus(pbStruct)
	case *eth.SignedBeaconBlockElectra:
		return SignedBeaconBlockElectraFromConsensus(pbStruct)
	case *eth.SignedBlindedBeaconBlockFulu:
		return SignedBlindedBeaconBlockFuluFromConsensus(pbStruct)
	case *eth.SignedBeaconBlockFulu:
		return SignedBeaconBlockFuluFromConsensus(pbStruct)
	case *eth.SignedBeaconBlockGloas:
		return SignedBeaconBlockGloasFromConsensus(pbStruct)
	default:
		return nil, ErrUnsupportedConversion
	}
}

func SignedBeaconBlockPhase0FromConsensus(b *eth.SignedBeaconBlock) *SignedBeaconBlock {
	return &SignedBeaconBlock{
		Message:   BeaconBlockFromConsensus(b.Block),
		Signature: hexutil.Encode(b.Signature),
	}
}

// ----------------------------------------------------------------------------
// Altair
// ----------------------------------------------------------------------------

func (b *SignedBeaconBlockAltair) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	bl, err := b.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	block := &eth.SignedBeaconBlockAltair{
		Block:     bl,
		Signature: sig,
	}
	return &eth.GenericSignedBeaconBlock{Block: &eth.GenericSignedBeaconBlock_Altair{Altair: block}}, nil
}

func (b *BeaconBlockAltair) ToGeneric() (*eth.GenericBeaconBlock, error) {
	block, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_Altair{Altair: block}}, nil
}

func (b *BeaconBlockAltair) ToConsensus() (*eth.BeaconBlockAltair, error) {
	if b == nil {
		return nil, errNilValue
	}
	if b.Body == nil {
		return nil, server.NewDecodeError(errNilValue, "Body")
	}
	if b.Body.SilaData == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaData")
	}
	if b.Body.SyncAggregate == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SyncAggregate")
	}

	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	proposerIndex, err := strconv.ParseUint(b.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	parentRoot, err := bytesutil.DecodeHexWithLength(b.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	stateRoot, err := bytesutil.DecodeHexWithLength(b.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	randaoReveal, err := bytesutil.DecodeHexWithLength(b.Body.RandaoReveal, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.RandaoReveal")
	}
	depositRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.DepositRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositRoot")
	}
	depositCount, err := strconv.ParseUint(b.Body.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositCount")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.BlockHash")
	}
	graffiti, err := bytesutil.DecodeHexWithLength(b.Body.Graffiti, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Graffiti")
	}
	proposerSlashings, err := ProposerSlashingsToConsensus(b.Body.ProposerSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.ProposerSlashings")
	}
	attesterSlashings, err := AttesterSlashingsToConsensus(b.Body.AttesterSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.AttesterSlashings")
	}
	atts, err := AttsToConsensus(b.Body.Attestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Attestations")
	}
	deposits, err := DepositsToConsensus(b.Body.Deposits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Deposits")
	}
	exits, err := SignedExitsToConsensus(b.Body.VoluntaryExits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.VoluntaryExits")
	}
	syncCommitteeBits, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeBits, fieldparams.SyncAggregateSyncCommitteeBytesLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeBits")
	}
	syncCommitteeSig, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeSignature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeSignature")
	}
	return &eth.BeaconBlockAltair{
		Slot:          primitives.Slot(slot),
		ProposerIndex: primitives.ValidatorIndex(proposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body: &eth.BeaconBlockBodyAltair{
			RandaoReveal: randaoReveal,
			SilaData: &eth.SilaData{
				DepositRoot:  depositRoot,
				DepositCount: depositCount,
				BlockHash:    blockHash,
			},
			Graffiti:          graffiti,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Attestations:      atts,
			Deposits:          deposits,
			VoluntaryExits:    exits,
			SyncAggregate: &eth.SyncAggregate{
				SyncCommitteeBits:      syncCommitteeBits,
				SyncCommitteeSignature: syncCommitteeSig,
			},
		},
	}, nil
}

func BeaconBlockAltairFromConsensus(b *eth.BeaconBlockAltair) *BeaconBlockAltair {
	return &BeaconBlockAltair{
		Slot:          fmt.Sprintf("%d", b.Slot),
		ProposerIndex: fmt.Sprintf("%d", b.ProposerIndex),
		ParentRoot:    hexutil.Encode(b.ParentRoot),
		StateRoot:     hexutil.Encode(b.StateRoot),
		Body: &BeaconBlockBodyAltair{
			RandaoReveal:      hexutil.Encode(b.Body.RandaoReveal),
			SilaData:          SilaDataFromConsensus(b.Body.SilaData),
			Graffiti:          hexutil.Encode(b.Body.Graffiti),
			ProposerSlashings: ProposerSlashingsFromConsensus(b.Body.ProposerSlashings),
			AttesterSlashings: AttesterSlashingsFromConsensus(b.Body.AttesterSlashings),
			Attestations:      AttsFromConsensus(b.Body.Attestations),
			Deposits:          DepositsFromConsensus(b.Body.Deposits),
			VoluntaryExits:    SignedExitsFromConsensus(b.Body.VoluntaryExits),
			SyncAggregate: &SyncAggregate{
				SyncCommitteeBits:      hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeBits),
				SyncCommitteeSignature: hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeSignature),
			},
		},
	}
}

func SignedBeaconBlockAltairFromConsensus(b *eth.SignedBeaconBlockAltair) *SignedBeaconBlockAltair {
	return &SignedBeaconBlockAltair{
		Message:   BeaconBlockAltairFromConsensus(b.Block),
		Signature: hexutil.Encode(b.Signature),
	}
}

// ----------------------------------------------------------------------------
// Bellatrix
// ----------------------------------------------------------------------------

func (b *SignedBeaconBlockBellatrix) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	bl, err := b.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	block := &eth.SignedBeaconBlockBellatrix{
		Block:     bl,
		Signature: sig,
	}
	return &eth.GenericSignedBeaconBlock{Block: &eth.GenericSignedBeaconBlock_Bellatrix{Bellatrix: block}}, nil
}

func (b *BeaconBlockBellatrix) ToGeneric() (*eth.GenericBeaconBlock, error) {
	block, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_Bellatrix{Bellatrix: block}}, nil
}

func (b *BeaconBlockBellatrix) ToConsensus() (*eth.BeaconBlockBellatrix, error) {
	if b == nil {
		return nil, errNilValue
	}
	if b.Body == nil {
		return nil, server.NewDecodeError(errNilValue, "Body")
	}
	if b.Body.SilaData == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaData")
	}
	if b.Body.SyncAggregate == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SyncAggregate")
	}
	if b.Body.SilaPayload == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaPayload")
	}

	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	proposerIndex, err := strconv.ParseUint(b.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	parentRoot, err := bytesutil.DecodeHexWithLength(b.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	stateRoot, err := bytesutil.DecodeHexWithLength(b.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	randaoReveal, err := bytesutil.DecodeHexWithLength(b.Body.RandaoReveal, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.RandaoReveal")
	}
	depositRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.DepositRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositRoot")
	}
	depositCount, err := strconv.ParseUint(b.Body.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositCount")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.BlockHash")
	}
	graffiti, err := bytesutil.DecodeHexWithLength(b.Body.Graffiti, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Graffiti")
	}
	proposerSlashings, err := ProposerSlashingsToConsensus(b.Body.ProposerSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.ProposerSlashings")
	}
	attesterSlashings, err := AttesterSlashingsToConsensus(b.Body.AttesterSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.AttesterSlashings")
	}
	atts, err := AttsToConsensus(b.Body.Attestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Attestations")
	}
	deposits, err := DepositsToConsensus(b.Body.Deposits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Deposits")
	}
	exits, err := SignedExitsToConsensus(b.Body.VoluntaryExits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.VoluntaryExits")
	}
	syncCommitteeBits, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeBits, fieldparams.SyncAggregateSyncCommitteeBytesLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeBits")
	}
	syncCommitteeSig, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeSignature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeSignature")
	}
	payload, err := b.Body.SilaPayload.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayload")
	}

	return &eth.BeaconBlockBellatrix{
		Slot:          primitives.Slot(slot),
		ProposerIndex: primitives.ValidatorIndex(proposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body: &eth.BeaconBlockBodyBellatrix{
			RandaoReveal: randaoReveal,
			SilaData: &eth.SilaData{
				DepositRoot:  depositRoot,
				DepositCount: depositCount,
				BlockHash:    blockHash,
			},
			Graffiti:          graffiti,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Attestations:      atts,
			Deposits:          deposits,
			VoluntaryExits:    exits,
			SyncAggregate: &eth.SyncAggregate{
				SyncCommitteeBits:      syncCommitteeBits,
				SyncCommitteeSignature: syncCommitteeSig,
			},
			SilaPayload: payload,
		},
	}, nil
}

func (b *SignedBlindedBeaconBlockBellatrix) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	bl, err := b.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	block := &eth.SignedBlindedBeaconBlockBellatrix{
		Block:     bl,
		Signature: sig,
	}
	return &eth.GenericSignedBeaconBlock{Block: &eth.GenericSignedBeaconBlock_BlindedBellatrix{BlindedBellatrix: block}, IsBlinded: true}, nil
}

func (b *BlindedBeaconBlockBellatrix) ToGeneric() (*eth.GenericBeaconBlock, error) {
	block, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_BlindedBellatrix{BlindedBellatrix: block}, IsBlinded: true}, nil
}

func (b *BlindedBeaconBlockBellatrix) ToConsensus() (*eth.BlindedBeaconBlockBellatrix, error) {
	if b == nil {
		return nil, errNilValue
	}
	if b.Body == nil {
		return nil, server.NewDecodeError(errNilValue, "Body")
	}
	if b.Body.SilaData == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaData")
	}
	if b.Body.SyncAggregate == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SyncAggregate")
	}
	if b.Body.SilaPayloadHeader == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaPayloadHeader")
	}

	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	proposerIndex, err := strconv.ParseUint(b.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	parentRoot, err := bytesutil.DecodeHexWithLength(b.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	stateRoot, err := bytesutil.DecodeHexWithLength(b.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	randaoReveal, err := bytesutil.DecodeHexWithLength(b.Body.RandaoReveal, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.RandaoReveal")
	}
	depositRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.DepositRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositRoot")
	}
	depositCount, err := strconv.ParseUint(b.Body.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositCount")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.BlockHash")
	}
	graffiti, err := bytesutil.DecodeHexWithLength(b.Body.Graffiti, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Graffiti")
	}
	proposerSlashings, err := ProposerSlashingsToConsensus(b.Body.ProposerSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.ProposerSlashings")
	}
	attesterSlashings, err := AttesterSlashingsToConsensus(b.Body.AttesterSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.AttesterSlashings")
	}
	atts, err := AttsToConsensus(b.Body.Attestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Attestations")
	}
	deposits, err := DepositsToConsensus(b.Body.Deposits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Deposits")
	}
	exits, err := SignedExitsToConsensus(b.Body.VoluntaryExits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.VoluntaryExits")
	}
	syncCommitteeBits, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeBits, fieldparams.SyncAggregateSyncCommitteeBytesLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeBits")
	}
	syncCommitteeSig, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeSignature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeSignature")
	}
	payload, err := b.Body.SilaPayloadHeader.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader")
	}
	return &eth.BlindedBeaconBlockBellatrix{
		Slot:          primitives.Slot(slot),
		ProposerIndex: primitives.ValidatorIndex(proposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body: &eth.BlindedBeaconBlockBodyBellatrix{
			RandaoReveal: randaoReveal,
			SilaData: &eth.SilaData{
				DepositRoot:  depositRoot,
				DepositCount: depositCount,
				BlockHash:    blockHash,
			},
			Graffiti:          graffiti,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Attestations:      atts,
			Deposits:          deposits,
			VoluntaryExits:    exits,
			SyncAggregate: &eth.SyncAggregate{
				SyncCommitteeBits:      syncCommitteeBits,
				SyncCommitteeSignature: syncCommitteeSig,
			},
			SilaPayloadHeader: payload,
		},
	}, nil
}

func BlindedBeaconBlockBellatrixFromConsensus(b *eth.BlindedBeaconBlockBellatrix) (*BlindedBeaconBlockBellatrix, error) {
	payload, err := SilaPayloadHeaderFromConsensus(b.Body.SilaPayloadHeader)
	if err != nil {
		return nil, err
	}

	return &BlindedBeaconBlockBellatrix{
		Slot:          fmt.Sprintf("%d", b.Slot),
		ProposerIndex: fmt.Sprintf("%d", b.ProposerIndex),
		ParentRoot:    hexutil.Encode(b.ParentRoot),
		StateRoot:     hexutil.Encode(b.StateRoot),
		Body: &BlindedBeaconBlockBodyBellatrix{
			RandaoReveal:      hexutil.Encode(b.Body.RandaoReveal),
			SilaData:          SilaDataFromConsensus(b.Body.SilaData),
			Graffiti:          hexutil.Encode(b.Body.Graffiti),
			ProposerSlashings: ProposerSlashingsFromConsensus(b.Body.ProposerSlashings),
			AttesterSlashings: AttesterSlashingsFromConsensus(b.Body.AttesterSlashings),
			Attestations:      AttsFromConsensus(b.Body.Attestations),
			Deposits:          DepositsFromConsensus(b.Body.Deposits),
			VoluntaryExits:    SignedExitsFromConsensus(b.Body.VoluntaryExits),
			SyncAggregate: &SyncAggregate{
				SyncCommitteeBits:      hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeBits),
				SyncCommitteeSignature: hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeSignature),
			},
			SilaPayloadHeader: payload,
		},
	}, nil
}

func SignedBlindedBeaconBlockBellatrixFromConsensus(b *eth.SignedBlindedBeaconBlockBellatrix) (*SignedBlindedBeaconBlockBellatrix, error) {
	blindedBlock, err := BlindedBeaconBlockBellatrixFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}
	return &SignedBlindedBeaconBlockBellatrix{
		Message:   blindedBlock,
		Signature: hexutil.Encode(b.Signature),
	}, nil
}

func BeaconBlockBellatrixFromConsensus(b *eth.BeaconBlockBellatrix) (*BeaconBlockBellatrix, error) {
	payload, err := SilaPayloadFromConsensus(b.Body.SilaPayload)
	if err != nil {
		return nil, err
	}

	return &BeaconBlockBellatrix{
		Slot:          fmt.Sprintf("%d", b.Slot),
		ProposerIndex: fmt.Sprintf("%d", b.ProposerIndex),
		ParentRoot:    hexutil.Encode(b.ParentRoot),
		StateRoot:     hexutil.Encode(b.StateRoot),
		Body: &BeaconBlockBodyBellatrix{
			RandaoReveal:      hexutil.Encode(b.Body.RandaoReveal),
			SilaData:          SilaDataFromConsensus(b.Body.SilaData),
			Graffiti:          hexutil.Encode(b.Body.Graffiti),
			ProposerSlashings: ProposerSlashingsFromConsensus(b.Body.ProposerSlashings),
			AttesterSlashings: AttesterSlashingsFromConsensus(b.Body.AttesterSlashings),
			Attestations:      AttsFromConsensus(b.Body.Attestations),
			Deposits:          DepositsFromConsensus(b.Body.Deposits),
			VoluntaryExits:    SignedExitsFromConsensus(b.Body.VoluntaryExits),
			SyncAggregate: &SyncAggregate{
				SyncCommitteeBits:      hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeBits),
				SyncCommitteeSignature: hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeSignature),
			},
			SilaPayload: payload,
		},
	}, nil
}

func SignedBeaconBlockBellatrixFromConsensus(b *eth.SignedBeaconBlockBellatrix) (*SignedBeaconBlockBellatrix, error) {
	block, err := BeaconBlockBellatrixFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}
	return &SignedBeaconBlockBellatrix{
		Message:   block,
		Signature: hexutil.Encode(b.Signature),
	}, nil
}

// ----------------------------------------------------------------------------
// Capella
// ----------------------------------------------------------------------------

func (b *SignedBeaconBlockCapella) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	bl, err := b.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	block := &eth.SignedBeaconBlockCapella{
		Block:     bl,
		Signature: sig,
	}
	return &eth.GenericSignedBeaconBlock{Block: &eth.GenericSignedBeaconBlock_Capella{Capella: block}}, nil
}

func (b *BeaconBlockCapella) ToGeneric() (*eth.GenericBeaconBlock, error) {
	block, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_Capella{Capella: block}}, nil
}

func (b *BeaconBlockCapella) ToConsensus() (*eth.BeaconBlockCapella, error) {
	if b == nil {
		return nil, errNilValue
	}
	if b.Body == nil {
		return nil, server.NewDecodeError(errNilValue, "Body")
	}
	if b.Body.SilaData == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaData")
	}
	if b.Body.SyncAggregate == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SyncAggregate")
	}
	if b.Body.SilaPayload == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaPayload")
	}

	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	proposerIndex, err := strconv.ParseUint(b.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	parentRoot, err := bytesutil.DecodeHexWithLength(b.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	stateRoot, err := bytesutil.DecodeHexWithLength(b.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	randaoReveal, err := bytesutil.DecodeHexWithLength(b.Body.RandaoReveal, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.RandaoReveal")
	}
	depositRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.DepositRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositRoot")
	}
	depositCount, err := strconv.ParseUint(b.Body.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositCount")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.BlockHash")
	}
	graffiti, err := bytesutil.DecodeHexWithLength(b.Body.Graffiti, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Graffiti")
	}
	proposerSlashings, err := ProposerSlashingsToConsensus(b.Body.ProposerSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.ProposerSlashings")
	}
	attesterSlashings, err := AttesterSlashingsToConsensus(b.Body.AttesterSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.AttesterSlashings")
	}
	atts, err := AttsToConsensus(b.Body.Attestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Attestations")
	}
	deposits, err := DepositsToConsensus(b.Body.Deposits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Deposits")
	}
	exits, err := SignedExitsToConsensus(b.Body.VoluntaryExits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.VoluntaryExits")
	}
	syncCommitteeBits, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeBits, fieldparams.SyncAggregateSyncCommitteeBytesLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeBits")
	}
	syncCommitteeSig, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeSignature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeSignature")
	}

	payload, err := b.Body.SilaPayload.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayload")
	}

	blsChanges, err := SignedBLSChangesToConsensus(b.Body.BLSToSilaChanges)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.BLSToSilaChanges")
	}

	return &eth.BeaconBlockCapella{
		Slot:          primitives.Slot(slot),
		ProposerIndex: primitives.ValidatorIndex(proposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body: &eth.BeaconBlockBodyCapella{
			RandaoReveal: randaoReveal,
			SilaData: &eth.SilaData{
				DepositRoot:  depositRoot,
				DepositCount: depositCount,
				BlockHash:    blockHash,
			},
			Graffiti:          graffiti,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Attestations:      atts,
			Deposits:          deposits,
			VoluntaryExits:    exits,
			SyncAggregate: &eth.SyncAggregate{
				SyncCommitteeBits:      syncCommitteeBits,
				SyncCommitteeSignature: syncCommitteeSig,
			},
			SilaPayload:      payload,
			BlsToSilaChanges: blsChanges,
		},
	}, nil
}

func (b *SignedBlindedBeaconBlockCapella) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	bl, err := b.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	block := &eth.SignedBlindedBeaconBlockCapella{
		Block:     bl,
		Signature: sig,
	}
	return &eth.GenericSignedBeaconBlock{Block: &eth.GenericSignedBeaconBlock_BlindedCapella{BlindedCapella: block}, IsBlinded: true}, nil
}

func (b *BlindedBeaconBlockCapella) ToGeneric() (*eth.GenericBeaconBlock, error) {
	block, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_BlindedCapella{BlindedCapella: block}, IsBlinded: true}, nil
}

func (b *BlindedBeaconBlockCapella) ToConsensus() (*eth.BlindedBeaconBlockCapella, error) {
	if b == nil {
		return nil, errNilValue
	}
	if b.Body == nil {
		return nil, server.NewDecodeError(errNilValue, "Body")
	}
	if b.Body.SilaData == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaData")
	}
	if b.Body.SyncAggregate == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SyncAggregate")
	}
	if b.Body.SilaPayloadHeader == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaPayloadHeader")
	}

	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	proposerIndex, err := strconv.ParseUint(b.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	parentRoot, err := bytesutil.DecodeHexWithLength(b.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	stateRoot, err := bytesutil.DecodeHexWithLength(b.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	randaoReveal, err := bytesutil.DecodeHexWithLength(b.Body.RandaoReveal, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.RandaoReveal")
	}
	depositRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.DepositRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositRoot")
	}
	depositCount, err := strconv.ParseUint(b.Body.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositCount")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.BlockHash")
	}
	graffiti, err := bytesutil.DecodeHexWithLength(b.Body.Graffiti, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Graffiti")
	}
	proposerSlashings, err := ProposerSlashingsToConsensus(b.Body.ProposerSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.ProposerSlashings")
	}
	attesterSlashings, err := AttesterSlashingsToConsensus(b.Body.AttesterSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.AttesterSlashings")
	}
	atts, err := AttsToConsensus(b.Body.Attestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Attestations")
	}
	deposits, err := DepositsToConsensus(b.Body.Deposits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Deposits")
	}
	exits, err := SignedExitsToConsensus(b.Body.VoluntaryExits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.VoluntaryExits")
	}
	syncCommitteeBits, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeBits, fieldparams.SyncAggregateSyncCommitteeBytesLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeBits")
	}
	syncCommitteeSig, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeSignature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeSignature")
	}

	payload, err := b.Body.SilaPayloadHeader.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader")
	}

	blsChanges, err := SignedBLSChangesToConsensus(b.Body.BLSToSilaChanges)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.BLSToSilaChanges")
	}

	return &eth.BlindedBeaconBlockCapella{
		Slot:          primitives.Slot(slot),
		ProposerIndex: primitives.ValidatorIndex(proposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body: &eth.BlindedBeaconBlockBodyCapella{
			RandaoReveal: randaoReveal,
			SilaData: &eth.SilaData{
				DepositRoot:  depositRoot,
				DepositCount: depositCount,
				BlockHash:    blockHash,
			},
			Graffiti:          graffiti,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Attestations:      atts,
			Deposits:          deposits,
			VoluntaryExits:    exits,
			SyncAggregate: &eth.SyncAggregate{
				SyncCommitteeBits:      syncCommitteeBits,
				SyncCommitteeSignature: syncCommitteeSig,
			},
			SilaPayloadHeader: payload,
			BlsToSilaChanges:  blsChanges,
		},
	}, nil
}

func BlindedBeaconBlockCapellaFromConsensus(b *eth.BlindedBeaconBlockCapella) (*BlindedBeaconBlockCapella, error) {
	payload, err := SilaPayloadHeaderCapellaFromConsensus(b.Body.SilaPayloadHeader)
	if err != nil {
		return nil, err
	}

	return &BlindedBeaconBlockCapella{
		Slot:          fmt.Sprintf("%d", b.Slot),
		ProposerIndex: fmt.Sprintf("%d", b.ProposerIndex),
		ParentRoot:    hexutil.Encode(b.ParentRoot),
		StateRoot:     hexutil.Encode(b.StateRoot),
		Body: &BlindedBeaconBlockBodyCapella{
			RandaoReveal:      hexutil.Encode(b.Body.RandaoReveal),
			SilaData:          SilaDataFromConsensus(b.Body.SilaData),
			Graffiti:          hexutil.Encode(b.Body.Graffiti),
			ProposerSlashings: ProposerSlashingsFromConsensus(b.Body.ProposerSlashings),
			AttesterSlashings: AttesterSlashingsFromConsensus(b.Body.AttesterSlashings),
			Attestations:      AttsFromConsensus(b.Body.Attestations),
			Deposits:          DepositsFromConsensus(b.Body.Deposits),
			VoluntaryExits:    SignedExitsFromConsensus(b.Body.VoluntaryExits),
			SyncAggregate: &SyncAggregate{
				SyncCommitteeBits:      hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeBits),
				SyncCommitteeSignature: hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeSignature),
			},
			SilaPayloadHeader: payload,
			BLSToSilaChanges:  SignedBLSChangesFromConsensus(b.Body.BlsToSilaChanges),
		},
	}, nil
}

func SignedBlindedBeaconBlockCapellaFromConsensus(b *eth.SignedBlindedBeaconBlockCapella) (*SignedBlindedBeaconBlockCapella, error) {
	blindedBlock, err := BlindedBeaconBlockCapellaFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}
	return &SignedBlindedBeaconBlockCapella{
		Message:   blindedBlock,
		Signature: hexutil.Encode(b.Signature),
	}, nil
}

func BeaconBlockCapellaFromConsensus(b *eth.BeaconBlockCapella) (*BeaconBlockCapella, error) {
	payload, err := SilaPayloadCapellaFromConsensus(b.Body.SilaPayload)
	if err != nil {
		return nil, err
	}

	return &BeaconBlockCapella{
		Slot:          fmt.Sprintf("%d", b.Slot),
		ProposerIndex: fmt.Sprintf("%d", b.ProposerIndex),
		ParentRoot:    hexutil.Encode(b.ParentRoot),
		StateRoot:     hexutil.Encode(b.StateRoot),
		Body: &BeaconBlockBodyCapella{
			RandaoReveal:      hexutil.Encode(b.Body.RandaoReveal),
			SilaData:          SilaDataFromConsensus(b.Body.SilaData),
			Graffiti:          hexutil.Encode(b.Body.Graffiti),
			ProposerSlashings: ProposerSlashingsFromConsensus(b.Body.ProposerSlashings),
			AttesterSlashings: AttesterSlashingsFromConsensus(b.Body.AttesterSlashings),
			Attestations:      AttsFromConsensus(b.Body.Attestations),
			Deposits:          DepositsFromConsensus(b.Body.Deposits),
			VoluntaryExits:    SignedExitsFromConsensus(b.Body.VoluntaryExits),
			SyncAggregate: &SyncAggregate{
				SyncCommitteeBits:      hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeBits),
				SyncCommitteeSignature: hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeSignature),
			},
			SilaPayload:      payload,
			BLSToSilaChanges: SignedBLSChangesFromConsensus(b.Body.BlsToSilaChanges),
		},
	}, nil
}

func SignedBeaconBlockCapellaFromConsensus(b *eth.SignedBeaconBlockCapella) (*SignedBeaconBlockCapella, error) {
	block, err := BeaconBlockCapellaFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}
	return &SignedBeaconBlockCapella{
		Message:   block,
		Signature: hexutil.Encode(b.Signature),
	}, nil
}

// ----------------------------------------------------------------------------
// Deneb
// ----------------------------------------------------------------------------

func (b *SignedBeaconBlockContentsDeneb) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}

	signedDenebBlock, err := b.SignedBlock.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "SignedBlock")
	}
	proofs := make([][]byte, len(b.KzgProofs))
	for i, proof := range b.KzgProofs {
		proofs[i], err = bytesutil.DecodeHexWithLength(proof, fieldparams.BLSPubkeyLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("KzgProofs[%d]", i))
		}
	}
	blbs := make([][]byte, len(b.Blobs))
	for i, blob := range b.Blobs {
		blbs[i], err = bytesutil.DecodeHexWithLength(blob, fieldparams.BlobLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Blobs[%d]", i))
		}
	}
	blk := &eth.SignedBeaconBlockContentsDeneb{
		Block:     signedDenebBlock,
		KzgProofs: proofs,
		Blobs:     blbs,
	}
	return &eth.GenericSignedBeaconBlock{Block: &eth.GenericSignedBeaconBlock_Deneb{Deneb: blk}}, nil
}

func (b *SignedBeaconBlockContentsDeneb) ToUnsigned() *BeaconBlockContentsDeneb {
	return &BeaconBlockContentsDeneb{
		Block:     b.SignedBlock.Message,
		KzgProofs: b.KzgProofs,
		Blobs:     b.Blobs,
	}
}

func (b *BeaconBlockContentsDeneb) ToGeneric() (*eth.GenericBeaconBlock, error) {
	block, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}

	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_Deneb{Deneb: block}}, nil
}

func (b *BeaconBlockContentsDeneb) ToConsensus() (*eth.BeaconBlockContentsDeneb, error) {
	if b == nil {
		return nil, errNilValue
	}

	denebBlock, err := b.Block.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Block")
	}
	proofs := make([][]byte, len(b.KzgProofs))
	for i, proof := range b.KzgProofs {
		proofs[i], err = bytesutil.DecodeHexWithLength(proof, fieldparams.BLSPubkeyLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("KzgProofs[%d]", i))
		}
	}
	blbs := make([][]byte, len(b.Blobs))
	for i, blob := range b.Blobs {
		blbs[i], err = bytesutil.DecodeHexWithLength(blob, fieldparams.BlobLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Blobs[%d]", i))
		}
	}
	return &eth.BeaconBlockContentsDeneb{
		Block:     denebBlock,
		KzgProofs: proofs,
		Blobs:     blbs,
	}, nil
}

func (b *BeaconBlockDeneb) ToConsensus() (*eth.BeaconBlockDeneb, error) {
	if b == nil {
		return nil, errNilValue
	}
	if b.Body == nil {
		return nil, server.NewDecodeError(errNilValue, "Body")
	}
	if b.Body.SilaData == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaData")
	}
	if b.Body.SyncAggregate == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SyncAggregate")
	}
	if b.Body.SilaPayload == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaPayload")
	}

	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	proposerIndex, err := strconv.ParseUint(b.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	parentRoot, err := bytesutil.DecodeHexWithLength(b.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	stateRoot, err := bytesutil.DecodeHexWithLength(b.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	randaoReveal, err := bytesutil.DecodeHexWithLength(b.Body.RandaoReveal, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.RandaoReveal")
	}
	depositRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.DepositRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositRoot")
	}
	depositCount, err := strconv.ParseUint(b.Body.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositCount")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.BlockHash")
	}
	graffiti, err := bytesutil.DecodeHexWithLength(b.Body.Graffiti, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Graffiti")
	}
	proposerSlashings, err := ProposerSlashingsToConsensus(b.Body.ProposerSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.ProposerSlashings")
	}
	attesterSlashings, err := AttesterSlashingsToConsensus(b.Body.AttesterSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.AttesterSlashings")
	}
	atts, err := AttsToConsensus(b.Body.Attestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Attestations")
	}
	deposits, err := DepositsToConsensus(b.Body.Deposits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Deposits")
	}
	exits, err := SignedExitsToConsensus(b.Body.VoluntaryExits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.VoluntaryExits")
	}
	syncCommitteeBits, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeBits, fieldparams.SyncAggregateSyncCommitteeBytesLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeBits")
	}
	syncCommitteeSig, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeSignature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeSignature")
	}
	payload, err := b.Body.SilaPayload.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayload")
	}
	blsChanges, err := SignedBLSChangesToConsensus(b.Body.BLSToSilaChanges)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.BLSToSilaChanges")
	}
	err = slice.VerifyMaxLength(b.Body.BlobKzgCommitments, fieldparams.MaxBlobCommitmentsPerBlock)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.BlobKzgCommitments")
	}
	blobKzgCommitments := make([][]byte, len(b.Body.BlobKzgCommitments))
	for i, b := range b.Body.BlobKzgCommitments {
		kzg, err := bytesutil.DecodeHexWithLength(b, fieldparams.BLSPubkeyLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Body.BlobKzgCommitments[%d]", i))
		}
		blobKzgCommitments[i] = kzg
	}
	return &eth.BeaconBlockDeneb{
		Slot:          primitives.Slot(slot),
		ProposerIndex: primitives.ValidatorIndex(proposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body: &eth.BeaconBlockBodyDeneb{
			RandaoReveal: randaoReveal,
			SilaData: &eth.SilaData{
				DepositRoot:  depositRoot,
				DepositCount: depositCount,
				BlockHash:    blockHash,
			},
			Graffiti:          graffiti,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Attestations:      atts,
			Deposits:          deposits,
			VoluntaryExits:    exits,
			SyncAggregate: &eth.SyncAggregate{
				SyncCommitteeBits:      syncCommitteeBits,
				SyncCommitteeSignature: syncCommitteeSig,
			},
			SilaPayload:      payload,
			BlsToSilaChanges: blsChanges,
			BlobKzgCommitments:    blobKzgCommitments,
		},
	}, nil
}

func (b *SignedBeaconBlockDeneb) ToConsensus() (*eth.SignedBeaconBlockDeneb, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	block, err := b.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	return &eth.SignedBeaconBlockDeneb{
		Block:     block,
		Signature: sig,
	}, nil
}

func (b *SignedBlindedBeaconBlockDeneb) ToConsensus() (*eth.SignedBlindedBeaconBlockDeneb, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	blindedBlock, err := b.Message.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.SignedBlindedBeaconBlockDeneb{
		Message:   blindedBlock,
		Signature: sig,
	}, nil
}

func (b *SignedBlindedBeaconBlockDeneb) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}
	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	blindedBlock, err := b.Message.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericSignedBeaconBlock{Block: &eth.GenericSignedBeaconBlock_BlindedDeneb{BlindedDeneb: &eth.SignedBlindedBeaconBlockDeneb{
		Message:   blindedBlock,
		Signature: sig,
	}}, IsBlinded: true}, nil
}

func (b *BlindedBeaconBlockDeneb) ToConsensus() (*eth.BlindedBeaconBlockDeneb, error) {
	if b == nil {
		return nil, errNilValue
	}
	if b.Body == nil {
		return nil, server.NewDecodeError(errNilValue, "Body")
	}
	if b.Body.SilaData == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaData")
	}
	if b.Body.SyncAggregate == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SyncAggregate")
	}
	if b.Body.SilaPayloadHeader == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaPayloadHeader")
	}

	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	proposerIndex, err := strconv.ParseUint(b.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	parentRoot, err := bytesutil.DecodeHexWithLength(b.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	stateRoot, err := bytesutil.DecodeHexWithLength(b.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	randaoReveal, err := bytesutil.DecodeHexWithLength(b.Body.RandaoReveal, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.RandaoReveal")
	}
	depositRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.DepositRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositRoot")
	}
	depositCount, err := strconv.ParseUint(b.Body.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositCount")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.BlockHash, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.BlockHash")
	}
	graffiti, err := bytesutil.DecodeHexWithLength(b.Body.Graffiti, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Graffiti")
	}
	proposerSlashings, err := ProposerSlashingsToConsensus(b.Body.ProposerSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.ProposerSlashings")
	}
	attesterSlashings, err := AttesterSlashingsToConsensus(b.Body.AttesterSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.AttesterSlashings")
	}
	atts, err := AttsToConsensus(b.Body.Attestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Attestations")
	}
	deposits, err := DepositsToConsensus(b.Body.Deposits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Deposits")
	}
	exits, err := SignedExitsToConsensus(b.Body.VoluntaryExits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.VoluntaryExits")
	}
	syncCommitteeBits, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeBits, fieldparams.SyncAggregateSyncCommitteeBytesLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeBits")
	}
	syncCommitteeSig, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeSignature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeSignature")
	}
	payload, err := b.Body.SilaPayloadHeader.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader")
	}
	blsChanges, err := SignedBLSChangesToConsensus(b.Body.BLSToSilaChanges)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.BLSToSilaChanges")
	}
	err = slice.VerifyMaxLength(b.Body.BlobKzgCommitments, fieldparams.MaxBlobCommitmentsPerBlock)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.BlobKzgCommitments")
	}
	blobKzgCommitments := make([][]byte, len(b.Body.BlobKzgCommitments))
	for i, b := range b.Body.BlobKzgCommitments {
		kzg, err := bytesutil.DecodeHexWithLength(b, fieldparams.BLSPubkeyLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Body.BlobKzgCommitments[%d]", i))
		}
		blobKzgCommitments[i] = kzg
	}

	return &eth.BlindedBeaconBlockDeneb{
		Slot:          primitives.Slot(slot),
		ProposerIndex: primitives.ValidatorIndex(proposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body: &eth.BlindedBeaconBlockBodyDeneb{
			RandaoReveal: randaoReveal,
			SilaData: &eth.SilaData{
				DepositRoot:  depositRoot,
				DepositCount: depositCount,
				BlockHash:    blockHash,
			},
			Graffiti:          graffiti,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Attestations:      atts,
			Deposits:          deposits,
			VoluntaryExits:    exits,
			SyncAggregate: &eth.SyncAggregate{
				SyncCommitteeBits:      syncCommitteeBits,
				SyncCommitteeSignature: syncCommitteeSig,
			},
			SilaPayloadHeader: payload,
			BlsToSilaChanges:  blsChanges,
			BlobKzgCommitments:     blobKzgCommitments,
		},
	}, nil
}

func (b *BlindedBeaconBlockDeneb) ToGeneric() (*eth.GenericBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}

	blindedBlock, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_BlindedDeneb{BlindedDeneb: blindedBlock}, IsBlinded: true}, nil
}

func BeaconBlockContentsDenebFromConsensus(b *eth.BeaconBlockContentsDeneb) (*BeaconBlockContentsDeneb, error) {
	block, err := BeaconBlockDenebFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}
	proofs := make([]string, len(b.KzgProofs))
	for i, proof := range b.KzgProofs {
		proofs[i] = hexutil.Encode(proof)
	}
	blbs := make([]string, len(b.Blobs))
	for i, blob := range b.Blobs {
		blbs[i] = hexutil.Encode(blob)
	}
	return &BeaconBlockContentsDeneb{
		Block:     block,
		KzgProofs: proofs,
		Blobs:     blbs,
	}, nil
}

func SignedBeaconBlockContentsDenebFromConsensus(b *eth.SignedBeaconBlockContentsDeneb) (*SignedBeaconBlockContentsDeneb, error) {
	block, err := SignedBeaconBlockDenebFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}

	proofs := make([]string, len(b.KzgProofs))
	for i, proof := range b.KzgProofs {
		proofs[i] = hexutil.Encode(proof)
	}

	blbs := make([]string, len(b.Blobs))
	for i, blob := range b.Blobs {
		blbs[i] = hexutil.Encode(blob)
	}

	return &SignedBeaconBlockContentsDeneb{
		SignedBlock: block,
		KzgProofs:   proofs,
		Blobs:       blbs,
	}, nil
}

func BlindedBeaconBlockDenebFromConsensus(b *eth.BlindedBeaconBlockDeneb) (*BlindedBeaconBlockDeneb, error) {
	blobKzgCommitments := make([]string, len(b.Body.BlobKzgCommitments))
	for i := range b.Body.BlobKzgCommitments {
		blobKzgCommitments[i] = hexutil.Encode(b.Body.BlobKzgCommitments[i])
	}
	payload, err := SilaPayloadHeaderDenebFromConsensus(b.Body.SilaPayloadHeader)
	if err != nil {
		return nil, err
	}

	return &BlindedBeaconBlockDeneb{
		Slot:          fmt.Sprintf("%d", b.Slot),
		ProposerIndex: fmt.Sprintf("%d", b.ProposerIndex),
		ParentRoot:    hexutil.Encode(b.ParentRoot),
		StateRoot:     hexutil.Encode(b.StateRoot),
		Body: &BlindedBeaconBlockBodyDeneb{
			RandaoReveal:      hexutil.Encode(b.Body.RandaoReveal),
			SilaData:          SilaDataFromConsensus(b.Body.SilaData),
			Graffiti:          hexutil.Encode(b.Body.Graffiti),
			ProposerSlashings: ProposerSlashingsFromConsensus(b.Body.ProposerSlashings),
			AttesterSlashings: AttesterSlashingsFromConsensus(b.Body.AttesterSlashings),
			Attestations:      AttsFromConsensus(b.Body.Attestations),
			Deposits:          DepositsFromConsensus(b.Body.Deposits),
			VoluntaryExits:    SignedExitsFromConsensus(b.Body.VoluntaryExits),
			SyncAggregate: &SyncAggregate{
				SyncCommitteeBits:      hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeBits),
				SyncCommitteeSignature: hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeSignature),
			},
			SilaPayloadHeader: payload,
			BLSToSilaChanges:  SignedBLSChangesFromConsensus(b.Body.BlsToSilaChanges),
			BlobKzgCommitments:     blobKzgCommitments,
		},
	}, nil
}

func SignedBlindedBeaconBlockDenebFromConsensus(b *eth.SignedBlindedBeaconBlockDeneb) (*SignedBlindedBeaconBlockDeneb, error) {
	block, err := BlindedBeaconBlockDenebFromConsensus(b.Message)
	if err != nil {
		return nil, err
	}
	return &SignedBlindedBeaconBlockDeneb{
		Message:   block,
		Signature: hexutil.Encode(b.Signature),
	}, nil
}

func BeaconBlockDenebFromConsensus(b *eth.BeaconBlockDeneb) (*BeaconBlockDeneb, error) {
	blobKzgCommitments := make([]string, len(b.Body.BlobKzgCommitments))
	for i := range b.Body.BlobKzgCommitments {
		blobKzgCommitments[i] = hexutil.Encode(b.Body.BlobKzgCommitments[i])
	}
	payload, err := SilaPayloadDenebFromConsensus(b.Body.SilaPayload)
	if err != nil {
		return nil, err
	}
	return &BeaconBlockDeneb{
		Slot:          fmt.Sprintf("%d", b.Slot),
		ProposerIndex: fmt.Sprintf("%d", b.ProposerIndex),
		ParentRoot:    hexutil.Encode(b.ParentRoot),
		StateRoot:     hexutil.Encode(b.StateRoot),
		Body: &BeaconBlockBodyDeneb{
			RandaoReveal:      hexutil.Encode(b.Body.RandaoReveal),
			SilaData:          SilaDataFromConsensus(b.Body.SilaData),
			Graffiti:          hexutil.Encode(b.Body.Graffiti),
			ProposerSlashings: ProposerSlashingsFromConsensus(b.Body.ProposerSlashings),
			AttesterSlashings: AttesterSlashingsFromConsensus(b.Body.AttesterSlashings),
			Attestations:      AttsFromConsensus(b.Body.Attestations),
			Deposits:          DepositsFromConsensus(b.Body.Deposits),
			VoluntaryExits:    SignedExitsFromConsensus(b.Body.VoluntaryExits),
			SyncAggregate: &SyncAggregate{
				SyncCommitteeBits:      hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeBits),
				SyncCommitteeSignature: hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeSignature),
			},
			SilaPayload:      payload,
			BLSToSilaChanges: SignedBLSChangesFromConsensus(b.Body.BlsToSilaChanges),
			BlobKzgCommitments:    blobKzgCommitments,
		},
	}, nil
}

func SignedBeaconBlockDenebFromConsensus(b *eth.SignedBeaconBlockDeneb) (*SignedBeaconBlockDeneb, error) {
	block, err := BeaconBlockDenebFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}
	return &SignedBeaconBlockDeneb{
		Message:   block,
		Signature: hexutil.Encode(b.Signature),
	}, nil
}

// ----------------------------------------------------------------------------
// Electra
// ----------------------------------------------------------------------------

func (b *SignedBeaconBlockContentsElectra) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}

	signedElectraBlock, err := b.SignedBlock.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "SignedBlock")
	}
	proofs := make([][]byte, len(b.KzgProofs))
	for i, proof := range b.KzgProofs {
		proofs[i], err = bytesutil.DecodeHexWithLength(proof, fieldparams.BLSPubkeyLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("KzgProofs[%d]", i))
		}
	}
	blbs := make([][]byte, len(b.Blobs))
	for i, blob := range b.Blobs {
		blbs[i], err = bytesutil.DecodeHexWithLength(blob, fieldparams.BlobLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Blobs[%d]", i))
		}
	}
	blk := &eth.SignedBeaconBlockContentsElectra{
		Block:     signedElectraBlock,
		KzgProofs: proofs,
		Blobs:     blbs,
	}
	return &eth.GenericSignedBeaconBlock{Block: &eth.GenericSignedBeaconBlock_Electra{Electra: blk}}, nil
}

func (b *SignedBeaconBlockContentsElectra) ToUnsigned() *BeaconBlockContentsElectra {
	return &BeaconBlockContentsElectra{
		Block:     b.SignedBlock.Message,
		KzgProofs: b.KzgProofs,
		Blobs:     b.Blobs,
	}
}

func (b *BeaconBlockContentsElectra) ToGeneric() (*eth.GenericBeaconBlock, error) {
	block, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}

	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_Electra{Electra: block}}, nil
}

func (b *BeaconBlockContentsElectra) ToConsensus() (*eth.BeaconBlockContentsElectra, error) {
	if b == nil {
		return nil, errNilValue
	}

	electraBlock, err := b.Block.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Block")
	}
	proofs := make([][]byte, len(b.KzgProofs))
	for i, proof := range b.KzgProofs {
		proofs[i], err = bytesutil.DecodeHexWithLength(proof, fieldparams.BLSPubkeyLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("KzgProofs[%d]", i))
		}
	}
	blbs := make([][]byte, len(b.Blobs))
	for i, blob := range b.Blobs {
		blbs[i], err = bytesutil.DecodeHexWithLength(blob, fieldparams.BlobLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Blobs[%d]", i))
		}
	}
	return &eth.BeaconBlockContentsElectra{
		Block:     electraBlock,
		KzgProofs: proofs,
		Blobs:     blbs,
	}, nil
}

func (b *BeaconBlockElectra) ToConsensus() (*eth.BeaconBlockElectra, error) {
	if b == nil {
		return nil, errNilValue
	}
	if b.Body == nil {
		return nil, server.NewDecodeError(errNilValue, "Body")
	}
	if b.Body.SilaData == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaData")
	}
	if b.Body.SyncAggregate == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SyncAggregate")
	}
	if b.Body.SilaPayload == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaPayload")
	}
	if b.Body.SilaRequests == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaRequests")
	}

	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	proposerIndex, err := strconv.ParseUint(b.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	parentRoot, err := bytesutil.DecodeHexWithLength(b.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	stateRoot, err := bytesutil.DecodeHexWithLength(b.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	randaoReveal, err := bytesutil.DecodeHexWithLength(b.Body.RandaoReveal, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.RandaoReveal")
	}
	depositRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.DepositRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositRoot")
	}
	depositCount, err := strconv.ParseUint(b.Body.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositCount")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.BlockHash")
	}
	graffiti, err := bytesutil.DecodeHexWithLength(b.Body.Graffiti, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Graffiti")
	}
	proposerSlashings, err := ProposerSlashingsToConsensus(b.Body.ProposerSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.ProposerSlashings")
	}
	attesterSlashings, err := AttesterSlashingsElectraToConsensus(b.Body.AttesterSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.AttesterSlashings")
	}
	atts, err := AttsElectraToConsensus(b.Body.Attestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Attestations")
	}
	deposits, err := DepositsToConsensus(b.Body.Deposits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Deposits")
	}
	exits, err := SignedExitsToConsensus(b.Body.VoluntaryExits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.VoluntaryExits")
	}
	syncCommitteeBits, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeBits, fieldparams.SyncAggregateSyncCommitteeBytesLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeBits")
	}
	syncCommitteeSig, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeSignature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeSignature")
	}

	payload, err := b.Body.SilaPayload.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayload")
	}

	blsChanges, err := SignedBLSChangesToConsensus(b.Body.BLSToSilaChanges)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.BLSToSilaChanges")
	}
	err = slice.VerifyMaxLength(b.Body.BlobKzgCommitments, fieldparams.MaxBlobCommitmentsPerBlock)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.BlobKzgCommitments")
	}
	blobKzgCommitments := make([][]byte, len(b.Body.BlobKzgCommitments))
	for i, b := range b.Body.BlobKzgCommitments {
		kzg, err := bytesutil.DecodeHexWithLength(b, fieldparams.BLSPubkeyLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Body.BlobKzgCommitments[%d]", i))
		}
		blobKzgCommitments[i] = kzg
	}

	requests, err := b.Body.SilaRequests.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaRequests")
	}

	return &eth.BeaconBlockElectra{
		Slot:          primitives.Slot(slot),
		ProposerIndex: primitives.ValidatorIndex(proposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body: &eth.BeaconBlockBodyElectra{
			RandaoReveal: randaoReveal,
			SilaData: &eth.SilaData{
				DepositRoot:  depositRoot,
				DepositCount: depositCount,
				BlockHash:    blockHash,
			},
			Graffiti:          graffiti,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Attestations:      atts,
			Deposits:          deposits,
			VoluntaryExits:    exits,
			SyncAggregate: &eth.SyncAggregate{
				SyncCommitteeBits:      syncCommitteeBits,
				SyncCommitteeSignature: syncCommitteeSig,
			},
			SilaPayload:      payload,
			BlsToSilaChanges: blsChanges,
			BlobKzgCommitments:    blobKzgCommitments,
			SilaRequests:     requests,
		},
	}, nil
}

func (b *SignedBeaconBlockElectra) ToConsensus() (*eth.SignedBeaconBlockElectra, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	block, err := b.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	return &eth.SignedBeaconBlockElectra{
		Block:     block,
		Signature: sig,
	}, nil
}

func (b *SignedBlindedBeaconBlockElectra) ToConsensus() (*eth.SignedBlindedBeaconBlockElectra, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	blindedBlock, err := b.Message.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.SignedBlindedBeaconBlockElectra{
		Message:   blindedBlock,
		Signature: sig,
	}, nil
}

func (b *SignedBlindedBeaconBlockElectra) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}
	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	blindedBlock, err := b.Message.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericSignedBeaconBlock{Block: &eth.GenericSignedBeaconBlock_BlindedElectra{BlindedElectra: &eth.SignedBlindedBeaconBlockElectra{
		Message:   blindedBlock,
		Signature: sig,
	}}, IsBlinded: true}, nil
}

func (b *BlindedBeaconBlockElectra) ToConsensus() (*eth.BlindedBeaconBlockElectra, error) {
	if b == nil {
		return nil, errNilValue
	}
	if b.Body == nil {
		return nil, server.NewDecodeError(errNilValue, "Body")
	}
	if b.Body.SilaData == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaData")
	}
	if b.Body.SyncAggregate == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SyncAggregate")
	}
	if b.Body.SilaPayloadHeader == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaPayloadHeader")
	}
	if b.Body.SilaRequests == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaRequests")
	}

	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	proposerIndex, err := strconv.ParseUint(b.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	parentRoot, err := bytesutil.DecodeHexWithLength(b.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	stateRoot, err := bytesutil.DecodeHexWithLength(b.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	randaoReveal, err := bytesutil.DecodeHexWithLength(b.Body.RandaoReveal, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.RandaoReveal")
	}
	depositRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.DepositRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositRoot")
	}
	depositCount, err := strconv.ParseUint(b.Body.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositCount")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.BlockHash, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.BlockHash")
	}
	graffiti, err := bytesutil.DecodeHexWithLength(b.Body.Graffiti, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Graffiti")
	}
	proposerSlashings, err := ProposerSlashingsToConsensus(b.Body.ProposerSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.ProposerSlashings")
	}
	attesterSlashings, err := AttesterSlashingsElectraToConsensus(b.Body.AttesterSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.AttesterSlashings")
	}
	atts, err := AttsElectraToConsensus(b.Body.Attestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Attestations")
	}
	deposits, err := DepositsToConsensus(b.Body.Deposits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Deposits")
	}
	exits, err := SignedExitsToConsensus(b.Body.VoluntaryExits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.VoluntaryExits")
	}
	syncCommitteeBits, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeBits, fieldparams.SyncAggregateSyncCommitteeBytesLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeBits")
	}
	syncCommitteeSig, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeSignature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeSignature")
	}
	payloadParentHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.ParentHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.ParentHash")
	}
	payloadFeeRecipient, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.FeeRecipient, fieldparams.FeeRecipientLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.FeeRecipient")
	}
	payloadStateRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.StateRoot")
	}
	payloadReceiptsRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.ReceiptsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.ReceiptsRoot")
	}
	payloadLogsBloom, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.LogsBloom, fieldparams.LogsBloomLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.LogsBloom")
	}
	payloadPrevRandao, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.PrevRandao, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.PrevRandao")
	}
	payloadBlockNumber, err := strconv.ParseUint(b.Body.SilaPayloadHeader.BlockNumber, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.BlockNumber")
	}
	payloadGasLimit, err := strconv.ParseUint(b.Body.SilaPayloadHeader.GasLimit, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.GasLimit")
	}
	payloadGasUsed, err := strconv.ParseUint(b.Body.SilaPayloadHeader.GasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.GasUsed")
	}
	payloadTimestamp, err := strconv.ParseUint(b.Body.SilaPayloadHeader.Timestamp, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.Timestamp")
	}
	payloadExtraData, err := bytesutil.DecodeHexWithMaxLength(b.Body.SilaPayloadHeader.ExtraData, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.ExtraData")
	}
	payloadBaseFeePerGas, err := bytesutil.Uint256ToSSZBytes(b.Body.SilaPayloadHeader.BaseFeePerGas)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.BaseFeePerGas")
	}
	payloadBlockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.BlockHash")
	}
	payloadTxsRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.TransactionsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.TransactionsRoot")
	}
	payloadWithdrawalsRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.WithdrawalsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.WithdrawalsRoot")
	}
	payloadBlobGasUsed, err := strconv.ParseUint(b.Body.SilaPayloadHeader.BlobGasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayload.BlobGasUsed")
	}
	payloadExcessBlobGas, err := strconv.ParseUint(b.Body.SilaPayloadHeader.ExcessBlobGas, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayload.ExcessBlobGas")
	}

	blsChanges, err := SignedBLSChangesToConsensus(b.Body.BLSToSilaChanges)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.BLSToSilaChanges")
	}
	err = slice.VerifyMaxLength(b.Body.BlobKzgCommitments, fieldparams.MaxBlobCommitmentsPerBlock)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.BlobKzgCommitments")
	}
	blobKzgCommitments := make([][]byte, len(b.Body.BlobKzgCommitments))
	for i, b := range b.Body.BlobKzgCommitments {
		kzg, err := bytesutil.DecodeHexWithLength(b, fieldparams.BLSPubkeyLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Body.BlobKzgCommitments[%d]", i))
		}
		blobKzgCommitments[i] = kzg
	}

	requests, err := b.Body.SilaRequests.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaRequests")
	}

	return &eth.BlindedBeaconBlockElectra{
		Slot:          primitives.Slot(slot),
		ProposerIndex: primitives.ValidatorIndex(proposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body: &eth.BlindedBeaconBlockBodyElectra{
			RandaoReveal: randaoReveal,
			SilaData: &eth.SilaData{
				DepositRoot:  depositRoot,
				DepositCount: depositCount,
				BlockHash:    blockHash,
			},
			Graffiti:          graffiti,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Attestations:      atts,
			Deposits:          deposits,
			VoluntaryExits:    exits,
			SyncAggregate: &eth.SyncAggregate{
				SyncCommitteeBits:      syncCommitteeBits,
				SyncCommitteeSignature: syncCommitteeSig,
			},
			SilaPayloadHeader: &silaenginev1.SilaPayloadHeaderDeneb{
				ParentHash:       payloadParentHash,
				FeeRecipient:     payloadFeeRecipient,
				StateRoot:        payloadStateRoot,
				ReceiptsRoot:     payloadReceiptsRoot,
				LogsBloom:        payloadLogsBloom,
				PrevRandao:       payloadPrevRandao,
				BlockNumber:      payloadBlockNumber,
				GasLimit:         payloadGasLimit,
				GasUsed:          payloadGasUsed,
				Timestamp:        payloadTimestamp,
				ExtraData:        payloadExtraData,
				BaseFeePerGas:    payloadBaseFeePerGas,
				BlockHash:        payloadBlockHash,
				TransactionsRoot: payloadTxsRoot,
				WithdrawalsRoot:  payloadWithdrawalsRoot,
				BlobGasUsed:      payloadBlobGasUsed,
				ExcessBlobGas:    payloadExcessBlobGas,
			},
			BlsToSilaChanges: blsChanges,
			BlobKzgCommitments:    blobKzgCommitments,
			SilaRequests:     requests,
		},
	}, nil
}

func (b *BlindedBeaconBlockElectra) ToGeneric() (*eth.GenericBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}

	blindedBlock, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_BlindedElectra{BlindedElectra: blindedBlock}, IsBlinded: true}, nil
}

func SignedBeaconBlockContentsElectraFromConsensus(b *eth.SignedBeaconBlockContentsElectra) (*SignedBeaconBlockContentsElectra, error) {
	block, err := SignedBeaconBlockElectraFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}

	proofs := make([]string, len(b.KzgProofs))
	for i, proof := range b.KzgProofs {
		proofs[i] = hexutil.Encode(proof)
	}

	blbs := make([]string, len(b.Blobs))
	for i, blob := range b.Blobs {
		blbs[i] = hexutil.Encode(blob)
	}

	return &SignedBeaconBlockContentsElectra{
		SignedBlock: block,
		KzgProofs:   proofs,
		Blobs:       blbs,
	}, nil
}

func BlindedBeaconBlockElectraFromConsensus(b *eth.BlindedBeaconBlockElectra) (*BlindedBeaconBlockElectra, error) {
	blobKzgCommitments := make([]string, len(b.Body.BlobKzgCommitments))
	for i := range b.Body.BlobKzgCommitments {
		blobKzgCommitments[i] = hexutil.Encode(b.Body.BlobKzgCommitments[i])
	}
	payload, err := SilaPayloadHeaderElectraFromConsensus(b.Body.SilaPayloadHeader)
	if err != nil {
		return nil, err
	}

	return &BlindedBeaconBlockElectra{
		Slot:          fmt.Sprintf("%d", b.Slot),
		ProposerIndex: fmt.Sprintf("%d", b.ProposerIndex),
		ParentRoot:    hexutil.Encode(b.ParentRoot),
		StateRoot:     hexutil.Encode(b.StateRoot),
		Body: &BlindedBeaconBlockBodyElectra{
			RandaoReveal:      hexutil.Encode(b.Body.RandaoReveal),
			SilaData:          SilaDataFromConsensus(b.Body.SilaData),
			Graffiti:          hexutil.Encode(b.Body.Graffiti),
			ProposerSlashings: ProposerSlashingsFromConsensus(b.Body.ProposerSlashings),
			AttesterSlashings: AttesterSlashingsElectraFromConsensus(b.Body.AttesterSlashings),
			Attestations:      AttsElectraFromConsensus(b.Body.Attestations),
			Deposits:          DepositsFromConsensus(b.Body.Deposits),
			VoluntaryExits:    SignedExitsFromConsensus(b.Body.VoluntaryExits),
			SyncAggregate: &SyncAggregate{
				SyncCommitteeBits:      hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeBits),
				SyncCommitteeSignature: hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeSignature),
			},
			SilaPayloadHeader: payload,
			BLSToSilaChanges:  SignedBLSChangesFromConsensus(b.Body.BlsToSilaChanges),
			BlobKzgCommitments:     blobKzgCommitments,
			SilaRequests:      SilaRequestsFromConsensus(b.Body.SilaRequests),
		},
	}, nil
}

func BeaconBlockContentsElectraFromConsensus(b *eth.BeaconBlockContentsElectra) (*BeaconBlockContentsElectra, error) {
	block, err := BeaconBlockElectraFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}
	proofs := make([]string, len(b.KzgProofs))
	for i, proof := range b.KzgProofs {
		proofs[i] = hexutil.Encode(proof)
	}
	blbs := make([]string, len(b.Blobs))
	for i, blob := range b.Blobs {
		blbs[i] = hexutil.Encode(blob)
	}
	return &BeaconBlockContentsElectra{
		Block:     block,
		KzgProofs: proofs,
		Blobs:     blbs,
	}, nil
}

func SignedBlindedBeaconBlockElectraFromConsensus(b *eth.SignedBlindedBeaconBlockElectra) (*SignedBlindedBeaconBlockElectra, error) {
	block, err := BlindedBeaconBlockElectraFromConsensus(b.Message)
	if err != nil {
		return nil, err
	}
	return &SignedBlindedBeaconBlockElectra{
		Message:   block,
		Signature: hexutil.Encode(b.Signature),
	}, nil
}

func BeaconBlockElectraFromConsensus(b *eth.BeaconBlockElectra) (*BeaconBlockElectra, error) {
	payload, err := SilaPayloadElectraFromConsensus(b.Body.SilaPayload)
	if err != nil {
		return nil, err
	}
	blobKzgCommitments := make([]string, len(b.Body.BlobKzgCommitments))
	for i := range b.Body.BlobKzgCommitments {
		blobKzgCommitments[i] = hexutil.Encode(b.Body.BlobKzgCommitments[i])
	}

	return &BeaconBlockElectra{
		Slot:          fmt.Sprintf("%d", b.Slot),
		ProposerIndex: fmt.Sprintf("%d", b.ProposerIndex),
		ParentRoot:    hexutil.Encode(b.ParentRoot),
		StateRoot:     hexutil.Encode(b.StateRoot),
		Body: &BeaconBlockBodyElectra{
			RandaoReveal:      hexutil.Encode(b.Body.RandaoReveal),
			SilaData:          SilaDataFromConsensus(b.Body.SilaData),
			Graffiti:          hexutil.Encode(b.Body.Graffiti),
			ProposerSlashings: ProposerSlashingsFromConsensus(b.Body.ProposerSlashings),
			AttesterSlashings: AttesterSlashingsElectraFromConsensus(b.Body.AttesterSlashings),
			Attestations:      AttsElectraFromConsensus(b.Body.Attestations),
			Deposits:          DepositsFromConsensus(b.Body.Deposits),
			VoluntaryExits:    SignedExitsFromConsensus(b.Body.VoluntaryExits),
			SyncAggregate: &SyncAggregate{
				SyncCommitteeBits:      hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeBits),
				SyncCommitteeSignature: hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeSignature),
			},
			SilaPayload:      payload,
			BLSToSilaChanges: SignedBLSChangesFromConsensus(b.Body.BlsToSilaChanges),
			BlobKzgCommitments:    blobKzgCommitments,
			SilaRequests:     SilaRequestsFromConsensus(b.Body.SilaRequests),
		},
	}, nil
}

func SignedBeaconBlockElectraFromConsensus(b *eth.SignedBeaconBlockElectra) (*SignedBeaconBlockElectra, error) {
	block, err := BeaconBlockElectraFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}
	return &SignedBeaconBlockElectra{
		Message:   block,
		Signature: hexutil.Encode(b.Signature),
	}, nil
}

// ----------------------------------------------------------------------------
// Fulu
// ----------------------------------------------------------------------------

func (b *SignedBeaconBlockContentsFulu) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}

	signedFuluBlock, err := b.SignedBlock.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "SignedBlock")
	}
	proofs := make([][]byte, len(b.KzgProofs))
	for i, proof := range b.KzgProofs {
		proofs[i], err = bytesutil.DecodeHexWithLength(proof, fieldparams.BLSPubkeyLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("KzgProofs[%d]", i))
		}
	}
	blbs := make([][]byte, len(b.Blobs))
	for i, blob := range b.Blobs {
		blbs[i], err = bytesutil.DecodeHexWithLength(blob, fieldparams.BlobLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Blobs[%d]", i))
		}
	}
	blk := &eth.SignedBeaconBlockContentsFulu{
		Block:     signedFuluBlock,
		KzgProofs: proofs,
		Blobs:     blbs,
	}
	return &eth.GenericSignedBeaconBlock{Block: &eth.GenericSignedBeaconBlock_Fulu{Fulu: blk}}, nil
}

func (b *SignedBeaconBlockContentsFulu) ToUnsigned() *BeaconBlockContentsFulu {
	return &BeaconBlockContentsFulu{
		Block:     b.SignedBlock.Message,
		KzgProofs: b.KzgProofs,
		Blobs:     b.Blobs,
	}
}

func (b *BeaconBlockContentsFulu) ToGeneric() (*eth.GenericBeaconBlock, error) {
	block, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}

	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_Fulu{Fulu: block}}, nil
}

func (b *BeaconBlockContentsFulu) ToConsensus() (*eth.BeaconBlockContentsFulu, error) {
	if b == nil {
		return nil, errNilValue
	}

	fuluBlock, err := b.Block.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Block")
	}
	proofs := make([][]byte, len(b.KzgProofs))
	for i, proof := range b.KzgProofs {
		proofs[i], err = bytesutil.DecodeHexWithLength(proof, fieldparams.BLSPubkeyLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("KzgProofs[%d]", i))
		}
	}
	blbs := make([][]byte, len(b.Blobs))
	for i, blob := range b.Blobs {
		blbs[i], err = bytesutil.DecodeHexWithLength(blob, fieldparams.BlobLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Blobs[%d]", i))
		}
	}
	return &eth.BeaconBlockContentsFulu{
		Block:     fuluBlock,
		KzgProofs: proofs,
		Blobs:     blbs,
	}, nil
}

func (b *SignedBeaconBlockFulu) ToConsensus() (*eth.SignedBeaconBlockFulu, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	block, err := b.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	return &eth.SignedBeaconBlockFulu{
		Block:     block,
		Signature: sig,
	}, nil
}

func (b *SignedBlindedBeaconBlockFulu) ToConsensus() (*eth.SignedBlindedBeaconBlockFulu, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	blindedBlock, err := b.Message.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.SignedBlindedBeaconBlockFulu{
		Message:   blindedBlock,
		Signature: sig,
	}, nil
}

func (b *SignedBlindedBeaconBlockFulu) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}
	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	blindedBlock, err := b.Message.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericSignedBeaconBlock{Block: &eth.GenericSignedBeaconBlock_BlindedFulu{BlindedFulu: &eth.SignedBlindedBeaconBlockFulu{
		Message:   blindedBlock,
		Signature: sig,
	}}, IsBlinded: true}, nil
}

func (b *BlindedBeaconBlockFulu) ToConsensus() (*eth.BlindedBeaconBlockFulu, error) {
	if b == nil {
		return nil, errNilValue
	}
	if b.Body == nil {
		return nil, server.NewDecodeError(errNilValue, "Body")
	}
	if b.Body.SilaData == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaData")
	}
	if b.Body.SyncAggregate == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SyncAggregate")
	}
	if b.Body.SilaPayloadHeader == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaPayloadHeader")
	}

	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	proposerIndex, err := strconv.ParseUint(b.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	parentRoot, err := bytesutil.DecodeHexWithLength(b.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	stateRoot, err := bytesutil.DecodeHexWithLength(b.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	randaoReveal, err := bytesutil.DecodeHexWithLength(b.Body.RandaoReveal, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.RandaoReveal")
	}
	depositRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.DepositRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositRoot")
	}
	depositCount, err := strconv.ParseUint(b.Body.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.DepositCount")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaData.BlockHash, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaData.BlockHash")
	}
	graffiti, err := bytesutil.DecodeHexWithLength(b.Body.Graffiti, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Graffiti")
	}
	proposerSlashings, err := ProposerSlashingsToConsensus(b.Body.ProposerSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.ProposerSlashings")
	}
	attesterSlashings, err := AttesterSlashingsElectraToConsensus(b.Body.AttesterSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.AttesterSlashings")
	}
	atts, err := AttsElectraToConsensus(b.Body.Attestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Attestations")
	}
	deposits, err := DepositsToConsensus(b.Body.Deposits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.Deposits")
	}
	exits, err := SignedExitsToConsensus(b.Body.VoluntaryExits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.VoluntaryExits")
	}
	syncCommitteeBits, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeBits, fieldparams.SyncAggregateSyncCommitteeBytesLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeBits")
	}
	syncCommitteeSig, err := bytesutil.DecodeHexWithLength(b.Body.SyncAggregate.SyncCommitteeSignature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SyncAggregate.SyncCommitteeSignature")
	}
	payloadParentHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.ParentHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.ParentHash")
	}
	payloadFeeRecipient, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.FeeRecipient, fieldparams.FeeRecipientLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.FeeRecipient")
	}
	payloadStateRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.StateRoot")
	}
	payloadReceiptsRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.ReceiptsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.ReceiptsRoot")
	}
	payloadLogsBloom, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.LogsBloom, fieldparams.LogsBloomLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.LogsBloom")
	}
	payloadPrevRandao, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.PrevRandao, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.PrevRandao")
	}
	payloadBlockNumber, err := strconv.ParseUint(b.Body.SilaPayloadHeader.BlockNumber, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.BlockNumber")
	}
	payloadGasLimit, err := strconv.ParseUint(b.Body.SilaPayloadHeader.GasLimit, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.GasLimit")
	}
	payloadGasUsed, err := strconv.ParseUint(b.Body.SilaPayloadHeader.GasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.GasUsed")
	}
	payloadTimestamp, err := strconv.ParseUint(b.Body.SilaPayloadHeader.Timestamp, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.Timestamp")
	}
	payloadExtraData, err := bytesutil.DecodeHexWithMaxLength(b.Body.SilaPayloadHeader.ExtraData, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.ExtraData")
	}
	payloadBaseFeePerGas, err := bytesutil.Uint256ToSSZBytes(b.Body.SilaPayloadHeader.BaseFeePerGas)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.BaseFeePerGas")
	}
	payloadBlockHash, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.BlockHash")
	}
	payloadTxsRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.TransactionsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.TransactionsRoot")
	}
	payloadWithdrawalsRoot, err := bytesutil.DecodeHexWithLength(b.Body.SilaPayloadHeader.WithdrawalsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayloadHeader.WithdrawalsRoot")
	}
	payloadBlobGasUsed, err := strconv.ParseUint(b.Body.SilaPayloadHeader.BlobGasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayload.BlobGasUsed")
	}
	payloadExcessBlobGas, err := strconv.ParseUint(b.Body.SilaPayloadHeader.ExcessBlobGas, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.SilaPayload.ExcessBlobGas")
	}
	if b.Body.SilaRequests == nil {
		return nil, server.NewDecodeError(errors.New("nil sila requests"), "Body.SilaRequests")
	}
	depositRequests := make([]*silaenginev1.DepositRequest, len(b.Body.SilaRequests.Deposits))
	for i, d := range b.Body.SilaRequests.Deposits {
		depositRequests[i], err = d.ToConsensus()
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Body.SilaRequests.Deposits[%d]", i))
		}
	}

	withdrawalRequests := make([]*silaenginev1.WithdrawalRequest, len(b.Body.SilaRequests.Withdrawals))
	for i, w := range b.Body.SilaRequests.Withdrawals {
		withdrawalRequests[i], err = w.ToConsensus()
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Body.SilaRequests.Withdrawals[%d]", i))
		}
	}

	consolidationRequests := make([]*silaenginev1.ConsolidationRequest, len(b.Body.SilaRequests.Consolidations))
	for i, c := range b.Body.SilaRequests.Consolidations {
		consolidationRequests[i], err = c.ToConsensus()
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Body.SilaRequests.Consolidations[%d]", i))
		}
	}

	blsChanges, err := SignedBLSChangesToConsensus(b.Body.BLSToSilaChanges)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.BLSToSilaChanges")
	}
	err = slice.VerifyMaxLength(b.Body.BlobKzgCommitments, fieldparams.MaxBlobCommitmentsPerBlock)
	if err != nil {
		return nil, server.NewDecodeError(err, "Body.BlobKzgCommitments")
	}
	blobKzgCommitments := make([][]byte, len(b.Body.BlobKzgCommitments))
	for i, b := range b.Body.BlobKzgCommitments {
		kzg, err := bytesutil.DecodeHexWithLength(b, fieldparams.BLSPubkeyLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Body.BlobKzgCommitments[%d]", i))
		}
		blobKzgCommitments[i] = kzg
	}

	return &eth.BlindedBeaconBlockFulu{
		Slot:          primitives.Slot(slot),
		ProposerIndex: primitives.ValidatorIndex(proposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body: &eth.BlindedBeaconBlockBodyElectra{
			RandaoReveal: randaoReveal,
			SilaData: &eth.SilaData{
				DepositRoot:  depositRoot,
				DepositCount: depositCount,
				BlockHash:    blockHash,
			},
			Graffiti:          graffiti,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Attestations:      atts,
			Deposits:          deposits,
			VoluntaryExits:    exits,
			SyncAggregate: &eth.SyncAggregate{
				SyncCommitteeBits:      syncCommitteeBits,
				SyncCommitteeSignature: syncCommitteeSig,
			},
			SilaPayloadHeader: &silaenginev1.SilaPayloadHeaderDeneb{
				ParentHash:       payloadParentHash,
				FeeRecipient:     payloadFeeRecipient,
				StateRoot:        payloadStateRoot,
				ReceiptsRoot:     payloadReceiptsRoot,
				LogsBloom:        payloadLogsBloom,
				PrevRandao:       payloadPrevRandao,
				BlockNumber:      payloadBlockNumber,
				GasLimit:         payloadGasLimit,
				GasUsed:          payloadGasUsed,
				Timestamp:        payloadTimestamp,
				ExtraData:        payloadExtraData,
				BaseFeePerGas:    payloadBaseFeePerGas,
				BlockHash:        payloadBlockHash,
				TransactionsRoot: payloadTxsRoot,
				WithdrawalsRoot:  payloadWithdrawalsRoot,
				BlobGasUsed:      payloadBlobGasUsed,
				ExcessBlobGas:    payloadExcessBlobGas,
			},
			BlsToSilaChanges: blsChanges,
			BlobKzgCommitments:    blobKzgCommitments,
			SilaRequests: &silaenginev1.SilaRequests{
				Deposits:       depositRequests,
				Withdrawals:    withdrawalRequests,
				Consolidations: consolidationRequests,
			},
		},
	}, nil
}

func (b *BlindedBeaconBlockFulu) ToGeneric() (*eth.GenericBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}

	blindedBlock, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_BlindedFulu{BlindedFulu: blindedBlock}, IsBlinded: true}, nil
}

func BeaconBlockContentsFuluFromConsensus(b *eth.BeaconBlockContentsFulu) (*BeaconBlockContentsFulu, error) {
	block, err := BeaconBlockFuluFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}
	proofs := make([]string, len(b.KzgProofs))
	for i, proof := range b.KzgProofs {
		proofs[i] = hexutil.Encode(proof)
	}
	blbs := make([]string, len(b.Blobs))
	for i, blob := range b.Blobs {
		blbs[i] = hexutil.Encode(blob)
	}
	return &BeaconBlockContentsFulu{
		Block:     block,
		KzgProofs: proofs,
		Blobs:     blbs,
	}, nil
}

func SignedBeaconBlockContentsFuluFromConsensus(b *eth.SignedBeaconBlockContentsFulu) (*SignedBeaconBlockContentsFulu, error) {
	block, err := SignedBeaconBlockFuluFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}

	proofs := make([]string, len(b.KzgProofs))
	for i, proof := range b.KzgProofs {
		proofs[i] = hexutil.Encode(proof)
	}

	blbs := make([]string, len(b.Blobs))
	for i, blob := range b.Blobs {
		blbs[i] = hexutil.Encode(blob)
	}

	return &SignedBeaconBlockContentsFulu{
		SignedBlock: block,
		KzgProofs:   proofs,
		Blobs:       blbs,
	}, nil
}

func BlindedBeaconBlockFuluFromConsensus(b *eth.BlindedBeaconBlockFulu) (*BlindedBeaconBlockFulu, error) {
	blobKzgCommitments := make([]string, len(b.Body.BlobKzgCommitments))
	for i := range b.Body.BlobKzgCommitments {
		blobKzgCommitments[i] = hexutil.Encode(b.Body.BlobKzgCommitments[i])
	}
	payload, err := SilaPayloadHeaderFuluFromConsensus(b.Body.SilaPayloadHeader)
	if err != nil {
		return nil, err
	}

	return &BlindedBeaconBlockFulu{
		Slot:          fmt.Sprintf("%d", b.Slot),
		ProposerIndex: fmt.Sprintf("%d", b.ProposerIndex),
		ParentRoot:    hexutil.Encode(b.ParentRoot),
		StateRoot:     hexutil.Encode(b.StateRoot),
		Body: &BlindedBeaconBlockBodyElectra{
			RandaoReveal:      hexutil.Encode(b.Body.RandaoReveal),
			SilaData:          SilaDataFromConsensus(b.Body.SilaData),
			Graffiti:          hexutil.Encode(b.Body.Graffiti),
			ProposerSlashings: ProposerSlashingsFromConsensus(b.Body.ProposerSlashings),
			AttesterSlashings: AttesterSlashingsElectraFromConsensus(b.Body.AttesterSlashings),
			Attestations:      AttsElectraFromConsensus(b.Body.Attestations),
			Deposits:          DepositsFromConsensus(b.Body.Deposits),
			VoluntaryExits:    SignedExitsFromConsensus(b.Body.VoluntaryExits),
			SyncAggregate: &SyncAggregate{
				SyncCommitteeBits:      hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeBits),
				SyncCommitteeSignature: hexutil.Encode(b.Body.SyncAggregate.SyncCommitteeSignature),
			},
			SilaPayloadHeader: payload,
			BLSToSilaChanges:  SignedBLSChangesFromConsensus(b.Body.BlsToSilaChanges),
			BlobKzgCommitments:     blobKzgCommitments,
			SilaRequests:      SilaRequestsFromConsensus(b.Body.SilaRequests),
		},
	}, nil
}

func SignedBlindedBeaconBlockFuluFromConsensus(b *eth.SignedBlindedBeaconBlockFulu) (*SignedBlindedBeaconBlockFulu, error) {
	block, err := BlindedBeaconBlockFuluFromConsensus(b.Message)
	if err != nil {
		return nil, err
	}
	return &SignedBlindedBeaconBlockFulu{
		Message:   block,
		Signature: hexutil.Encode(b.Signature),
	}, nil
}

func SignedBeaconBlockFuluFromConsensus(b *eth.SignedBeaconBlockFulu) (*SignedBeaconBlockFulu, error) {
	block, err := BeaconBlockFuluFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}
	return &SignedBeaconBlockFulu{
		Message:   block,
		Signature: hexutil.Encode(b.Signature),
	}, nil
}

// ----------------------------------------------------------------------------
// Gloas
// ----------------------------------------------------------------------------

func SignedBeaconBlockGloasFromConsensus(b *eth.SignedBeaconBlockGloas) (*SignedBeaconBlockGloas, error) {
	block, err := BeaconBlockGloasFromConsensus(b.Block)
	if err != nil {
		return nil, err
	}
	return &SignedBeaconBlockGloas{
		Message:   block,
		Signature: hexutil.Encode(b.Signature),
	}, nil
}

func BeaconBlockGloasFromConsensus(b *eth.BeaconBlockGloas) (*BeaconBlockGloas, error) {
	payloadAttestations := make([]*PayloadAttestation, len(b.Body.PayloadAttestations))
	for i, pa := range b.Body.PayloadAttestations {
		payloadAttestations[i] = PayloadAttestationFromConsensus(pa)
	}

	return &BeaconBlockGloas{
		Slot:          fmt.Sprintf("%d", b.Slot),
		ProposerIndex: fmt.Sprintf("%d", b.ProposerIndex),
		ParentRoot:    hexutil.Encode(b.ParentRoot),
		StateRoot:     hexutil.Encode(b.StateRoot),
		Body: &BeaconBlockBodyGloas{
			RandaoReveal:              hexutil.Encode(b.Body.RandaoReveal),
			SilaData:                  SilaDataFromConsensus(b.Body.SilaData),
			Graffiti:                  hexutil.Encode(b.Body.Graffiti),
			ProposerSlashings:         ProposerSlashingsFromConsensus(b.Body.ProposerSlashings),
			AttesterSlashings:         AttesterSlashingsElectraFromConsensus(b.Body.AttesterSlashings),
			Attestations:              AttsElectraFromConsensus(b.Body.Attestations),
			Deposits:                  DepositsFromConsensus(b.Body.Deposits),
			VoluntaryExits:            SignedExitsFromConsensus(b.Body.VoluntaryExits),
			SyncAggregate:             SyncAggregateFromConsensus(b.Body.SyncAggregate),
			BLSToSilaChanges:     SignedBLSChangesFromConsensus(b.Body.BlsToSilaChanges),
			SignedSilaPayloadBid: SignedSilaPayloadBidFromConsensus(b.Body.SignedSilaPayloadBid),
			PayloadAttestations:       payloadAttestations,
			ParentSilaRequests:   SilaRequestsFromConsensus(b.Body.ParentSilaRequests),
		},
	}, nil
}

func SignedSilaPayloadBidFromConsensus(b *eth.SignedSilaPayloadBid) *SignedSilaPayloadBid {
	return &SignedSilaPayloadBid{
		Message:   SilaPayloadBidFromConsensus(b.Message),
		Signature: hexutil.Encode(b.Signature),
	}
}

func SilaPayloadBidFromConsensus(b *eth.SilaPayloadBid) *SilaPayloadBid {
	blobKzgCommitments := make([]string, len(b.BlobKzgCommitments))
	for i := range b.BlobKzgCommitments {
		blobKzgCommitments[i] = hexutil.Encode(b.BlobKzgCommitments[i])
	}
	return &SilaPayloadBid{
		ParentBlockHash:       hexutil.Encode(b.ParentBlockHash),
		ParentBlockRoot:       hexutil.Encode(b.ParentBlockRoot),
		BlockHash:             hexutil.Encode(b.BlockHash),
		PrevRandao:            hexutil.Encode(b.PrevRandao),
		FeeRecipient:          hexutil.Encode(b.FeeRecipient),
		GasLimit:              fmt.Sprintf("%d", b.GasLimit),
		BuilderIndex:          fmt.Sprintf("%d", b.BuilderIndex),
		Slot:                  fmt.Sprintf("%d", b.Slot),
		Value:                 fmt.Sprintf("%d", b.Value),
		ExecutionPayment:      fmt.Sprintf("%d", b.ExecutionPayment),
		BlobKzgCommitments:    blobKzgCommitments,
		SilaRequestsRoot: hexutil.Encode(b.SilaRequestsRoot),
	}
}

func PayloadAttestationFromConsensus(pa *eth.PayloadAttestation) *PayloadAttestation {
	return &PayloadAttestation{
		AggregationBits: hexutil.Encode(pa.AggregationBits),
		Data:            PayloadAttestationDataFromConsensus(pa.Data),
		Signature:       hexutil.Encode(pa.Signature),
	}
}

func PayloadAttestationMessageFromConsensus(m *eth.PayloadAttestationMessage) *PayloadAttestationMessage {
	return &PayloadAttestationMessage{
		ValidatorIndex: fmt.Sprintf("%d", m.ValidatorIndex),
		Data:           PayloadAttestationDataFromConsensus(m.Data),
		Signature:      hexutil.Encode(m.Signature),
	}
}

func PayloadAttestationDataFromConsensus(d *eth.PayloadAttestationData) *PayloadAttestationData {
	return &PayloadAttestationData{
		BeaconBlockRoot:   hexutil.Encode(d.BeaconBlockRoot),
		Slot:              fmt.Sprintf("%d", d.Slot),
		PayloadPresent:    d.PayloadPresent,
		BlobDataAvailable: d.BlobDataAvailable,
	}
}

func (b *SignedBeaconBlockGloas) ToGeneric() (*eth.GenericSignedBeaconBlock, error) {
	if b == nil {
		return nil, errNilValue
	}
	signed, err := b.ToConsensus()
	if err != nil {
		return nil, err
	}
	return &eth.GenericSignedBeaconBlock{
		Block: &eth.GenericSignedBeaconBlock_Gloas{Gloas: signed},
	}, nil
}

func (b *SignedBeaconBlockGloas) ToConsensus() (*eth.SignedBeaconBlockGloas, error) {
	if b == nil {
		return nil, errNilValue
	}

	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	block, err := b.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	return &eth.SignedBeaconBlockGloas{
		Block:     block,
		Signature: sig,
	}, nil
}

func (b *BeaconBlockGloas) ToConsensus() (*eth.BeaconBlockGloas, error) {
	if b == nil {
		return nil, errNilValue
	}
	if b.Body == nil {
		return nil, server.NewDecodeError(errNilValue, "Body")
	}
	if b.Body.SilaData == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SilaData")
	}
	if b.Body.SyncAggregate == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SyncAggregate")
	}
	if b.Body.SignedSilaPayloadBid == nil {
		return nil, server.NewDecodeError(errNilValue, "Body.SignedSilaPayloadBid")
	}

	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	proposerIndex, err := strconv.ParseUint(b.ProposerIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerIndex")
	}
	parentRoot, err := bytesutil.DecodeHexWithLength(b.ParentRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentRoot")
	}
	stateRoot, err := bytesutil.DecodeHexWithLength(b.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	body, err := b.Body.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Body")
	}
	return &eth.BeaconBlockGloas{
		Slot:          primitives.Slot(slot),
		ProposerIndex: primitives.ValidatorIndex(proposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body:          body,
	}, nil
}

func (b *BeaconBlockBodyGloas) ToConsensus() (*eth.BeaconBlockBodyGloas, error) {
	if b == nil {
		return nil, errNilValue
	}

	randaoReveal, err := bytesutil.DecodeHexWithLength(b.RandaoReveal, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "RandaoReveal")
	}
	depositRoot, err := bytesutil.DecodeHexWithLength(b.SilaData.DepositRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaData.DepositRoot")
	}
	depositCount, err := strconv.ParseUint(b.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaData.DepositCount")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.SilaData.BlockHash, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaData.BlockHash")
	}
	graffiti, err := bytesutil.DecodeHexWithLength(b.Graffiti, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Graffiti")
	}
	proposerSlashings, err := ProposerSlashingsToConsensus(b.ProposerSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "ProposerSlashings")
	}
	attesterSlashings, err := AttesterSlashingsElectraToConsensus(b.AttesterSlashings)
	if err != nil {
		return nil, server.NewDecodeError(err, "AttesterSlashings")
	}
	atts, err := AttsElectraToConsensus(b.Attestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "Attestations")
	}
	deposits, err := DepositsToConsensus(b.Deposits)
	if err != nil {
		return nil, server.NewDecodeError(err, "Deposits")
	}
	exits, err := SignedExitsToConsensus(b.VoluntaryExits)
	if err != nil {
		return nil, server.NewDecodeError(err, "VoluntaryExits")
	}
	syncCommitteeBits, err := bytesutil.DecodeHexWithLength(b.SyncAggregate.SyncCommitteeBits, fieldparams.SyncAggregateSyncCommitteeBytesLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SyncAggregate.SyncCommitteeBits")
	}
	syncCommitteeSig, err := bytesutil.DecodeHexWithLength(b.SyncAggregate.SyncCommitteeSignature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SyncAggregate.SyncCommitteeSignature")
	}
	blsChanges, err := SignedBLSChangesToConsensus(b.BLSToSilaChanges)
	if err != nil {
		return nil, server.NewDecodeError(err, "BLSToSilaChanges")
	}
	signedBid, err := b.SignedSilaPayloadBid.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "SignedSilaPayloadBid")
	}
	payloadAttestations, err := PayloadAttestationsToConsensus(b.PayloadAttestations)
	if err != nil {
		return nil, server.NewDecodeError(err, "PayloadAttestations")
	}
	var parentSilaRequests *silaenginev1.SilaRequests
	if b.ParentSilaRequests != nil {
		parentSilaRequests, err = b.ParentSilaRequests.ToConsensus()
		if err != nil {
			return nil, server.NewDecodeError(err, "ParentSilaRequests")
		}
	}

	return &eth.BeaconBlockBodyGloas{
		RandaoReveal: randaoReveal,
		SilaData: &eth.SilaData{
			DepositRoot:  depositRoot,
			DepositCount: depositCount,
			BlockHash:    blockHash,
		},
		Graffiti:          graffiti,
		ProposerSlashings: proposerSlashings,
		AttesterSlashings: attesterSlashings,
		Attestations:      atts,
		Deposits:          deposits,
		VoluntaryExits:    exits,
		SyncAggregate: &eth.SyncAggregate{
			SyncCommitteeBits:      syncCommitteeBits,
			SyncCommitteeSignature: syncCommitteeSig,
		},
		BlsToSilaChanges:     blsChanges,
		SignedSilaPayloadBid: signedBid,
		PayloadAttestations:       payloadAttestations,
		ParentSilaRequests:   parentSilaRequests,
	}, nil
}

func (b *BeaconBlockGloas) ToGeneric() (*eth.GenericBeaconBlock, error) {
	block, err := b.ToConsensus()
	if err != nil {
		return nil, errors.Wrap(err, "could not convert gloas block to consensus")
	}
	return &eth.GenericBeaconBlock{Block: &eth.GenericBeaconBlock_Gloas{Gloas: block}}, nil
}

func (b *SignedSilaPayloadBid) ToConsensus() (*eth.SignedSilaPayloadBid, error) {
	if b == nil {
		return nil, errNilValue
	}
	sig, err := bytesutil.DecodeHexWithLength(b.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	message, err := b.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	return &eth.SignedSilaPayloadBid{
		Message:   message,
		Signature: sig,
	}, nil
}

func (b *SilaPayloadBid) ToConsensus() (*eth.SilaPayloadBid, error) {
	if b == nil {
		return nil, errNilValue
	}
	parentBlockHash, err := bytesutil.DecodeHexWithLength(b.ParentBlockHash, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentBlockHash")
	}
	parentBlockRoot, err := bytesutil.DecodeHexWithLength(b.ParentBlockRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentBlockRoot")
	}
	blockHash, err := bytesutil.DecodeHexWithLength(b.BlockHash, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "BlockHash")
	}
	prevRandao, err := bytesutil.DecodeHexWithLength(b.PrevRandao, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "PrevRandao")
	}
	feeRecipient, err := bytesutil.DecodeHexWithLength(b.FeeRecipient, fieldparams.FeeRecipientLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "FeeRecipient")
	}
	gasLimit, err := strconv.ParseUint(b.GasLimit, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "GasLimit")
	}
	builderIndex, err := strconv.ParseUint(b.BuilderIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "BuilderIndex")
	}
	slot, err := strconv.ParseUint(b.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	value, err := strconv.ParseUint(b.Value, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Value")
	}
	executionPayment, err := strconv.ParseUint(b.ExecutionPayment, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ExecutionPayment")
	}
	err = slice.VerifyMaxLength(b.BlobKzgCommitments, fieldparams.MaxBlobCommitmentsPerBlock)
	if err != nil {
		return nil, server.NewDecodeError(err, "BlobKzgCommitments")
	}
	blobKzgCommitments := make([][]byte, len(b.BlobKzgCommitments))
	for i, commitment := range b.BlobKzgCommitments {
		kzg, err := bytesutil.DecodeHexWithLength(commitment, fieldparams.BLSPubkeyLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("BlobKzgCommitments[%d]", i))
		}
		blobKzgCommitments[i] = kzg
	}
	silaRequestsRoot, err := bytesutil.DecodeHexWithLength(b.SilaRequestsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaRequestsRoot")
	}
	return &eth.SilaPayloadBid{
		ParentBlockHash:       parentBlockHash,
		ParentBlockRoot:       parentBlockRoot,
		BlockHash:             blockHash,
		PrevRandao:            prevRandao,
		FeeRecipient:          feeRecipient,
		GasLimit:              gasLimit,
		BuilderIndex:          primitives.BuilderIndex(builderIndex),
		Slot:                  primitives.Slot(slot),
		Value:                 primitives.Gwei(value),
		ExecutionPayment:      primitives.Gwei(executionPayment),
		BlobKzgCommitments:    blobKzgCommitments,
		SilaRequestsRoot: silaRequestsRoot,
	}, nil
}

func PayloadAttestationsToConsensus(pa []*PayloadAttestation) ([]*eth.PayloadAttestation, error) {
	if pa == nil {
		return nil, errNilValue
	}
	result := make([]*eth.PayloadAttestation, len(pa))
	for i, p := range pa {
		converted, err := p.ToConsensus()
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("[%d]", i))
		}
		result[i] = converted
	}
	return result, nil
}

func (p *PayloadAttestation) ToConsensus() (*eth.PayloadAttestation, error) {
	if p == nil {
		return nil, errNilValue
	}
	aggregationBits, err := hexutil.Decode(p.AggregationBits)
	if err != nil {
		return nil, server.NewDecodeError(err, "AggregationBits")
	}
	data, err := p.Data.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Data")
	}
	sig, err := bytesutil.DecodeHexWithLength(p.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	return &eth.PayloadAttestation{
		AggregationBits: aggregationBits,
		Data:            data,
		Signature:       sig,
	}, nil
}

func (d *PayloadAttestationData) ToConsensus() (*eth.PayloadAttestationData, error) {
	if d == nil {
		return nil, errNilValue
	}
	beaconBlockRoot, err := bytesutil.DecodeHexWithLength(d.BeaconBlockRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "BeaconBlockRoot")
	}
	slot, err := strconv.ParseUint(d.Slot, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Slot")
	}
	return &eth.PayloadAttestationData{
		BeaconBlockRoot:   beaconBlockRoot,
		Slot:              primitives.Slot(slot),
		PayloadPresent:    d.PayloadPresent,
		BlobDataAvailable: d.BlobDataAvailable,
	}, nil
}

func (p *PayloadAttestationMessage) ToConsensus() (*eth.PayloadAttestationMessage, error) {
	if p == nil {
		return nil, errNilValue
	}
	validatorIndex, err := strconv.ParseUint(p.ValidatorIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ValidatorIndex")
	}
	data, err := p.Data.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Data")
	}
	sig, err := bytesutil.DecodeHexWithLength(p.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	return &eth.PayloadAttestationMessage{
		ValidatorIndex: primitives.ValidatorIndex(validatorIndex),
		Data:           data,
		Signature:      sig,
	}, nil
}
