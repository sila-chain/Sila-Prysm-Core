package structs

import (
	"fmt"
	"strconv"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	eth "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila/common/hexutil"
)

// SilaPayloadEnvelopeFromConsensus converts a proto envelope to the API struct.
func SilaPayloadEnvelopeFromConsensus(e *eth.SilaPayloadEnvelope) (*SilaPayloadEnvelope, error) {
	payload, err := SilaPayloadGloasFromConsensus(e.Payload)
	if err != nil {
		return nil, err
	}
	var requests *ExecutionRequests
	if e.ExecutionRequests != nil {
		requests = ExecutionRequestsFromConsensus(e.ExecutionRequests)
	}
	return &SilaPayloadEnvelope{
		Payload:               payload,
		ExecutionRequests:     requests,
		BuilderIndex:          fmt.Sprintf("%d", e.BuilderIndex),
		BeaconBlockRoot:       hexutil.Encode(e.BeaconBlockRoot),
		ParentBeaconBlockRoot: hexutil.Encode(e.ParentBeaconBlockRoot),
	}, nil
}

// SignedSilaPayloadEnvelopeFromConsensus converts a signed proto envelope to the API struct.
func SignedSilaPayloadEnvelopeFromConsensus(e *eth.SignedSilaPayloadEnvelope) (*SignedSilaPayloadEnvelope, error) {
	envelope, err := SilaPayloadEnvelopeFromConsensus(e.Message)
	if err != nil {
		return nil, err
	}
	return &SignedSilaPayloadEnvelope{
		Message:   envelope,
		Signature: hexutil.Encode(e.Signature),
	}, nil
}

// BlockContentsGloasFromConsensus converts a proto Gloas block, envelope, and
// blob data to the API struct.
func BlockContentsGloasFromConsensus(block *eth.BeaconBlockGloas, envelope *eth.SilaPayloadEnvelope, kzgProofs [][]byte, blobs [][]byte) (*BlockContentsGloas, error) {
	b, err := BeaconBlockGloasFromConsensus(block)
	if err != nil {
		return nil, err
	}
	env, err := SilaPayloadEnvelopeFromConsensus(envelope)
	if err != nil {
		return nil, err
	}
	encodedProofs := make([]string, len(kzgProofs))
	for i, p := range kzgProofs {
		encodedProofs[i] = hexutil.Encode(p)
	}
	encodedBlobs := make([]string, len(blobs))
	for i, b := range blobs {
		encodedBlobs[i] = hexutil.Encode(b)
	}
	return &BlockContentsGloas{
		Block:                    b,
		SilaPayloadEnvelope: env,
		KzgProofs:                encodedProofs,
		Blobs:                    encodedBlobs,
	}, nil
}

// ToConsensus converts the API struct to a proto SilaPayloadEnvelope.
func (e *SilaPayloadEnvelope) ToConsensus() (*eth.SilaPayloadEnvelope, error) {
	if e == nil {
		return nil, server.NewDecodeError(errNilValue, "SilaPayloadEnvelope")
	}
	payload, err := e.Payload.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Payload")
	}
	var requests *silaenginev1.ExecutionRequests
	if e.ExecutionRequests != nil {
		requests, err = e.ExecutionRequests.ToConsensus()
		if err != nil {
			return nil, server.NewDecodeError(err, "ExecutionRequests")
		}
	}
	builderIndex, err := strconv.ParseUint(e.BuilderIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "BuilderIndex")
	}
	beaconBlockRoot, err := bytesutil.DecodeHexWithLength(e.BeaconBlockRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "BeaconBlockRoot")
	}
	parentBeaconBlockRoot, err := bytesutil.DecodeHexWithLength(e.ParentBeaconBlockRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentBeaconBlockRoot")
	}
	return &eth.SilaPayloadEnvelope{
		Payload:               payload,
		ExecutionRequests:     requests,
		BuilderIndex:          primitives.BuilderIndex(builderIndex),
		BeaconBlockRoot:       beaconBlockRoot,
		ParentBeaconBlockRoot: parentBeaconBlockRoot,
	}, nil
}

// ToConsensus converts the API struct to a proto SignedSilaPayloadEnvelope.
func (e *SignedSilaPayloadEnvelope) ToConsensus() (*eth.SignedSilaPayloadEnvelope, error) {
	if e == nil {
		return nil, server.NewDecodeError(errNilValue, "SignedSilaPayloadEnvelope")
	}
	msg, err := e.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	sig, err := bytesutil.DecodeHexWithLength(e.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	return &eth.SignedSilaPayloadEnvelope{
		Message:   msg,
		Signature: sig,
	}, nil
}

func BlindedSilaPayloadEnvelopeFromConsensus(b *eth.WireBlindedSilaPayloadEnvelope) (*BlindedSilaPayloadEnvelope, error) {
	if b == nil {
		return nil, errNilValue
	}
	var requests *ExecutionRequests
	if b.ExecutionRequests != nil {
		requests = ExecutionRequestsFromConsensus(b.ExecutionRequests)
	}
	return &BlindedSilaPayloadEnvelope{
		PayloadRoot:           hexutil.Encode(b.PayloadRoot),
		ExecutionRequests:     requests,
		BuilderIndex:          fmt.Sprintf("%d", b.BuilderIndex),
		BeaconBlockRoot:       hexutil.Encode(b.BeaconBlockRoot),
		ParentBeaconBlockRoot: hexutil.Encode(b.ParentBeaconBlockRoot),
	}, nil
}

func (b *BlindedSilaPayloadEnvelope) ToConsensus() (*eth.WireBlindedSilaPayloadEnvelope, error) {
	if b == nil {
		return nil, server.NewDecodeError(errNilValue, "BlindedSilaPayloadEnvelope")
	}
	payloadRoot, err := bytesutil.DecodeHexWithLength(b.PayloadRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "PayloadRoot")
	}
	var requests *silaenginev1.ExecutionRequests
	if b.ExecutionRequests != nil {
		requests, err = b.ExecutionRequests.ToConsensus()
		if err != nil {
			return nil, server.NewDecodeError(err, "ExecutionRequests")
		}
	}
	builderIndex, err := strconv.ParseUint(b.BuilderIndex, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "BuilderIndex")
	}
	beaconBlockRoot, err := bytesutil.DecodeHexWithLength(b.BeaconBlockRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "BeaconBlockRoot")
	}
	parentBeaconBlockRoot, err := bytesutil.DecodeHexWithLength(b.ParentBeaconBlockRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentBeaconBlockRoot")
	}
	return &eth.WireBlindedSilaPayloadEnvelope{
		PayloadRoot:           payloadRoot,
		ExecutionRequests:     requests,
		BuilderIndex:          primitives.BuilderIndex(builderIndex),
		BeaconBlockRoot:       beaconBlockRoot,
		ParentBeaconBlockRoot: parentBeaconBlockRoot,
	}, nil
}

func (s *SignedBlindedSilaPayloadEnvelope) ToConsensus() (*eth.SignedWireBlindedSilaPayloadEnvelope, error) {
	if s == nil {
		return nil, server.NewDecodeError(errNilValue, "SignedBlindedSilaPayloadEnvelope")
	}
	msg, err := s.Message.ToConsensus()
	if err != nil {
		return nil, server.NewDecodeError(err, "Message")
	}
	sig, err := bytesutil.DecodeHexWithLength(s.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	return &eth.SignedWireBlindedSilaPayloadEnvelope{
		Message:   msg,
		Signature: sig,
	}, nil
}

// SignedSilaPayloadEnvelopeContentsFromConsensus builds the API struct
// used for stateless envelope publishing from native components.
func SignedSilaPayloadEnvelopeContentsFromConsensus(signed *eth.SignedSilaPayloadEnvelope, kzgProofs [][]byte, blobs [][]byte) (*SignedSilaPayloadEnvelopeContents, error) {
	signedJSON, err := SignedSilaPayloadEnvelopeFromConsensus(signed)
	if err != nil {
		return nil, err
	}
	encodedProofs := make([]string, len(kzgProofs))
	for i, p := range kzgProofs {
		encodedProofs[i] = hexutil.Encode(p)
	}
	encodedBlobs := make([]string, len(blobs))
	for i, b := range blobs {
		encodedBlobs[i] = hexutil.Encode(b)
	}
	return &SignedSilaPayloadEnvelopeContents{
		SignedSilaPayloadEnvelope: signedJSON,
		KzgProofs:                      encodedProofs,
		Blobs:                          encodedBlobs,
	}, nil
}

// ToConsensus decodes the API struct into the signed envelope plus raw blob and
// KZG proof bytes used by the stateless publish path.
func (c *SignedSilaPayloadEnvelopeContents) ToConsensus() (*eth.SignedSilaPayloadEnvelope, [][]byte, [][]byte, error) {
	if c == nil {
		return nil, nil, nil, server.NewDecodeError(errNilValue, "SignedSilaPayloadEnvelopeContents")
	}
	signed, err := c.SignedSilaPayloadEnvelope.ToConsensus()
	if err != nil {
		return nil, nil, nil, server.NewDecodeError(err, "SignedSilaPayloadEnvelope")
	}
	proofs := make([][]byte, len(c.KzgProofs))
	for i, p := range c.KzgProofs {
		proof, err := bytesutil.DecodeHexWithLength(p, 48)
		if err != nil {
			return nil, nil, nil, server.NewDecodeError(err, fmt.Sprintf("KzgProofs[%d]", i))
		}
		proofs[i] = proof
	}
	blobs := make([][]byte, len(c.Blobs))
	for i, b := range c.Blobs {
		blob, err := bytesutil.DecodeHexWithLength(b, fieldparams.BlobSize)
		if err != nil {
			return nil, nil, nil, server.NewDecodeError(err, fmt.Sprintf("Blobs[%d]", i))
		}
		blobs[i] = blob
	}
	return signed, proofs, blobs, nil
}

// WireBlindedFromFull derives the spec-wire blinded envelope from a full one: payload_root is
// HashTreeRoot(payload), so HashTreeRoot(blinded) == HashTreeRoot(full) and a validator signature
// over either form is valid against the other.
func WireBlindedFromFull(full *eth.SilaPayloadEnvelope) (*eth.WireBlindedSilaPayloadEnvelope, error) {
	if full == nil {
		return nil, nil
	}
	payloadRoot, err := full.Payload.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	return &eth.WireBlindedSilaPayloadEnvelope{
		PayloadRoot:           payloadRoot[:],
		ExecutionRequests:     full.ExecutionRequests,
		BuilderIndex:          full.BuilderIndex,
		BeaconBlockRoot:       bytesutil.SafeCopyBytes(full.BeaconBlockRoot),
		ParentBeaconBlockRoot: bytesutil.SafeCopyBytes(full.ParentBeaconBlockRoot),
	}, nil
}

// SignedWireBlindedFromFull lifts a signed envelope to its blinded form, preserving the signature.
func SignedWireBlindedFromFull(full *eth.SignedSilaPayloadEnvelope) (*eth.SignedWireBlindedSilaPayloadEnvelope, error) {
	if full == nil {
		return nil, nil
	}
	msg, err := WireBlindedFromFull(full.Message)
	if err != nil {
		return nil, err
	}
	return &eth.SignedWireBlindedSilaPayloadEnvelope{
		Message:   msg,
		Signature: bytesutil.SafeCopyBytes(full.Signature),
	}, nil
}
