package blocks

import (
	"bytes"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/signing"
	field_params "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	consensus_types "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"google.golang.org/protobuf/proto"
)

type signedSilaPayloadEnvelope struct {
	s *silapb.SignedSilaPayloadEnvelope
}

type silaPayloadEnvelope struct {
	p *silapb.SilaPayloadEnvelope
}

// WrappedROSignedSilaPayloadEnvelope wraps a signed sila payload envelope proto in a read-only interface.
func WrappedROSignedSilaPayloadEnvelope(s *silapb.SignedSilaPayloadEnvelope) (interfaces.ROSignedSilaPayloadEnvelope, error) {
	w := signedSilaPayloadEnvelope{s: s}
	if w.IsNil() {
		return nil, consensus_types.ErrNilObjectWrapped
	}
	return w, nil
}

// WrappedROSilaPayloadEnvelope wraps an sila payload envelope proto in a read-only interface.
func WrappedROSilaPayloadEnvelope(p *silapb.SilaPayloadEnvelope) (interfaces.ROSilaPayloadEnvelope, error) {
	w := &silaPayloadEnvelope{p: p}
	if w.IsNil() {
		return nil, consensus_types.ErrNilObjectWrapped
	}
	return w, nil
}

// Envelope returns the sila payload envelope as a read-only interface.
func (s signedSilaPayloadEnvelope) Envelope() (interfaces.ROSilaPayloadEnvelope, error) {
	return WrappedROSilaPayloadEnvelope(s.s.Message)
}

// Signature returns the BLS signature as a 96-byte array.
func (s signedSilaPayloadEnvelope) Signature() [field_params.BLSSignatureLength]byte {
	return [field_params.BLSSignatureLength]byte(s.s.Signature)
}

// IsNil reports whether the signed envelope or its contents are invalid.
func (s signedSilaPayloadEnvelope) IsNil() bool {
	if s.s == nil {
		return true
	}
	if len(s.s.Signature) != field_params.BLSSignatureLength {
		return true
	}
	if s.s.Message == nil {
		return true
	}
	if len(s.s.Message.BeaconBlockRoot) != field_params.RootLength {
		return true
	}
	if len(s.s.Message.ParentBeaconBlockRoot) != field_params.RootLength {
		return true
	}
	if s.s.Message.SilaRequests == nil {
		return true
	}
	if s.s.Message.Payload == nil {
		return true
	}
	w := silaPayloadEnvelope{p: s.s.Message}
	return w.IsNil()
}

// SigningRoot computes the signing root for the signed envelope with the provided domain.
func (s signedSilaPayloadEnvelope) SigningRoot(domain []byte) (root [32]byte, err error) {
	return signing.ComputeSigningRoot(s.s.Message, domain)
}

// Proto returns the underlying protobuf message.
func (s signedSilaPayloadEnvelope) Proto() proto.Message {
	return s.s
}

// IsNil reports whether the envelope or its required fields are invalid.
func (p *silaPayloadEnvelope) IsNil() bool {
	if p.p == nil {
		return true
	}
	if p.p.Payload == nil {
		return true
	}
	if len(p.p.BeaconBlockRoot) != field_params.RootLength {
		return true
	}
	if len(p.p.ParentBeaconBlockRoot) != field_params.RootLength {
		return true
	}
	return false
}

// IsBlinded reports whether the envelope contains a blinded payload.
func (p *silaPayloadEnvelope) IsBlinded() bool {
	return false
}

// Execution returns the sila payload as a read-only interface.
func (p *silaPayloadEnvelope) Execution() (interfaces.SilaData, error) {
	return WrappedSilaPayloadGloas(p.p.Payload)
}

// SilaRequests returns the sila requests attached to the envelope.
func (p *silaPayloadEnvelope) SilaRequests() *silaenginev1.SilaRequests {
	return silapb.CopySilaRequests(p.p.SilaRequests)
}

// BuilderIndex returns the proposer/builder index for the envelope.
func (p *silaPayloadEnvelope) BuilderIndex() primitives.BuilderIndex {
	return p.p.BuilderIndex
}

// BeaconBlockRoot returns the beacon block root referenced by the envelope.
func (p *silaPayloadEnvelope) BeaconBlockRoot() [field_params.RootLength]byte {
	return [field_params.RootLength]byte(p.p.BeaconBlockRoot)
}

// ParentBeaconBlockRoot returns the parent beacon block root referenced by the envelope.
func (p *silaPayloadEnvelope) ParentBeaconBlockRoot() [field_params.RootLength]byte {
	return [field_params.RootLength]byte(p.p.ParentBeaconBlockRoot)
}

// Slot returns the slot derived from the payload's slot_number field.
func (p *silaPayloadEnvelope) Slot() primitives.Slot {
	return primitives.Slot(p.p.Payload.SlotNumber)
}

// BlockHash returns the block hash from the sila payload.
func (p *silaPayloadEnvelope) BlockHash() [field_params.RootLength]byte {
	return [field_params.RootLength]byte(p.p.Payload.BlockHash)
}

type blindedSilaPayloadEnvelope struct {
	p *silapb.BlindedSilaPayloadEnvelope
}

// WrappedROBlindedSilaPayloadEnvelope wraps a blinded sila payload envelope proto in a read-only interface.
func WrappedROBlindedSilaPayloadEnvelope(p *silapb.BlindedSilaPayloadEnvelope) (interfaces.ROBlindedSilaPayloadEnvelope, error) {
	w := &blindedSilaPayloadEnvelope{p: p}
	if w.IsNil() {
		return nil, consensus_types.ErrNilObjectWrapped
	}
	return w, nil
}

func (p *blindedSilaPayloadEnvelope) IsNil() bool {
	if p.p == nil {
		return true
	}
	if len(p.p.BeaconBlockRoot) != field_params.RootLength {
		return true
	}
	if len(p.p.ParentBeaconBlockRoot) != field_params.RootLength {
		return true
	}
	if len(p.p.BlockHash) != field_params.RootLength {
		return true
	}
	return false
}

func (p *blindedSilaPayloadEnvelope) IsBlinded() bool {
	return true
}

func (p *blindedSilaPayloadEnvelope) SilaRequests() *silaenginev1.SilaRequests {
	return silapb.CopySilaRequests(p.p.SilaRequests)
}

func (p *blindedSilaPayloadEnvelope) BuilderIndex() primitives.BuilderIndex {
	return p.p.BuilderIndex
}

func (p *blindedSilaPayloadEnvelope) BeaconBlockRoot() [field_params.RootLength]byte {
	return [field_params.RootLength]byte(p.p.BeaconBlockRoot)
}

func (p *blindedSilaPayloadEnvelope) ParentBeaconBlockRoot() [field_params.RootLength]byte {
	return [field_params.RootLength]byte(p.p.ParentBeaconBlockRoot)
}

func (p *blindedSilaPayloadEnvelope) Slot() primitives.Slot {
	return p.p.Slot
}

func (p *blindedSilaPayloadEnvelope) BlockHash() [field_params.RootLength]byte {
	return [field_params.RootLength]byte(p.p.BlockHash)
}

// BlockBuiltOnEnvelope checks if the block's parent hash matches the envelope's sila block hash.
func BlockBuiltOnEnvelope(env interfaces.ROSignedSilaPayloadEnvelope, blk ROBlock) (bool, error) {
	msg, err := env.Envelope()
	if err != nil {
		return false, err
	}
	ex, err := msg.Execution()
	if err != nil {
		return false, err
	}
	ph, err := blk.ParentHash()
	if err != nil {
		return false, err
	}
	return bytes.Equal(ex.BlockHash(), ph[:]), nil
}
