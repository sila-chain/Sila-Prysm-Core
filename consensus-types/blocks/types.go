package blocks

import (
	field_params "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	eth "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

var (
	_ = interfaces.ReadOnlySignedBeaconBlock(&SignedBeaconBlock{})
	_ = interfaces.ReadOnlyBeaconBlock(&BeaconBlock{})
	_ = interfaces.ReadOnlyBeaconBlockBody(&BeaconBlockBody{})
)

var (
	errPayloadWrongType       = errors.New("sila payload has wrong type")
	errPayloadHeaderWrongType = errors.New("sila payload header has wrong type")
)

const (
	incorrectBlockVersion = "incorrect beacon block version"
	incorrectBodyVersion  = "incorrect beacon block body version"
)

var (
	// ErrUnsupportedVersion for beacon block methods.
	ErrUnsupportedVersion    = errors.New("unsupported beacon block version")
	errNilBlob               = errors.New("received nil blob sidecar")
	errNilDataColumn         = errors.New("received nil data column sidecar")
	errNilBlock              = errors.New("received nil beacon block")
	errNilBlockBody          = errors.New("received nil beacon block body")
	errIncorrectBlockVersion = errors.New(incorrectBlockVersion)
	errIncorrectBodyVersion  = errors.New(incorrectBodyVersion)
	errNilBlockHeader        = errors.New("received nil beacon block header")
	errMissingBlockSignature = errors.New("received nil beacon block signature")
)

// BeaconBlockBody is the main beacon block body structure. It can represent any block type.
type BeaconBlockBody struct {
	version                   int
	randaoReveal              [field_params.BLSSignatureLength]byte
	silaexecData                  *eth.SilaExecutionData
	graffiti                  [field_params.RootLength]byte
	proposerSlashings         []*eth.ProposerSlashing
	attesterSlashings         []*eth.AttesterSlashing
	attesterSlashingsElectra  []*eth.AttesterSlashingElectra
	attestations              []*eth.Attestation
	attestationsElectra       []*eth.AttestationElectra
	deposits                  []*eth.Deposit
	voluntaryExits            []*eth.SignedVoluntaryExit
	syncAggregate             *eth.SyncAggregate
	silaPayload          interfaces.ExecutionData
	silaPayloadHeader    interfaces.ExecutionData
	blsToSilaChanges     []*eth.SignedBLSToSilaChange
	blobKzgCommitments        [][]byte
	silaRequests         *silaenginev1.SilaRequests
	signedSilaPayloadBid *eth.SignedSilaPayloadBid
	payloadAttestations       []*eth.PayloadAttestation
	parentSilaRequests   *silaenginev1.SilaRequests
}

var _ interfaces.ReadOnlyBeaconBlockBody = &BeaconBlockBody{}

// BeaconBlock is the main beacon block structure. It can represent any block type.
type BeaconBlock struct {
	version       int
	slot          primitives.Slot
	proposerIndex primitives.ValidatorIndex
	parentRoot    [field_params.RootLength]byte
	stateRoot     [field_params.RootLength]byte
	body          *BeaconBlockBody
}

// SignedBeaconBlock is the main signed beacon block structure. It can represent any block type.
type SignedBeaconBlock struct {
	version   int
	block     *BeaconBlock
	signature [field_params.BLSSignatureLength]byte
}
