package interfaces

import (
	field_params "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"google.golang.org/protobuf/proto"
)

type ROSignedSilaPayloadEnvelope interface {
	Envelope() (ROSilaPayloadEnvelope, error)
	Signature() [field_params.BLSSignatureLength]byte
	SigningRoot([]byte) ([32]byte, error)
	IsNil() bool
	Proto() proto.Message
}

// ROBlindedSilaPayloadEnvelope contains the fields common to both
// full and blinded sila payload envelopes.
type ROBlindedSilaPayloadEnvelope interface {
	ExecutionRequests() *silaenginev1.ExecutionRequests
	BuilderIndex() primitives.BuilderIndex
	BeaconBlockRoot() [field_params.RootLength]byte
	ParentBeaconBlockRoot() [field_params.RootLength]byte
	Slot() primitives.Slot
	BlockHash() [field_params.RootLength]byte
	IsBlinded() bool
	IsNil() bool
}

type ROSilaPayloadEnvelope interface {
	ROBlindedSilaPayloadEnvelope
	Execution() (ExecutionData, error)
}
