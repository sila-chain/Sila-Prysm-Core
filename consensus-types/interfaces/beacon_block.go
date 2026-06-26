package interfaces

import (
	field_params "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	validatorpb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/validator-client"
	"github.com/pkg/errors"
	ssz "github.com/sila-chain/fastssz"
	"google.golang.org/protobuf/proto"
)

var ErrIncompatibleFork = errors.New("Can't convert to fork-specific interface")

// ReadOnlySignedBeaconBlock is an interface describing the method set of
// a signed beacon block.
type ReadOnlySignedBeaconBlock interface {
	Block() ReadOnlyBeaconBlock
	Signature() [field_params.BLSSignatureLength]byte
	IsNil() bool
	Copy() (SignedBeaconBlock, error)
	Proto() (proto.Message, error)
	PbGenericBlock() (*silapb.GenericSignedBeaconBlock, error)
	ToBlinded() (ReadOnlySignedBeaconBlock, error)
	ssz.Marshaler
	ssz.Unmarshaler
	Version() int
	IsBlinded() bool
	Header() (*silapb.SignedBeaconBlockHeader, error)
}

// ReadOnlyBeaconBlock describes an interface which states the methods
// employed by an object that is a beacon block.
type ReadOnlyBeaconBlock interface {
	Slot() primitives.Slot
	ProposerIndex() primitives.ValidatorIndex
	ParentRoot() [field_params.RootLength]byte
	StateRoot() [field_params.RootLength]byte
	Body() ReadOnlyBeaconBlockBody
	IsNil() bool
	IsBlinded() bool
	HashTreeRoot() ([field_params.RootLength]byte, error)
	Proto() (proto.Message, error)
	ssz.Marshaler
	ssz.Unmarshaler
	ssz.HashRoot
	Version() int
	AsSignRequestObject() (validatorpb.SignRequestObject, error)
}

// ReadOnlyBeaconBlockBody describes the method set employed by an object
// that is a beacon block body.
type ReadOnlyBeaconBlockBody interface {
	Version() int
	RandaoReveal() [field_params.BLSSignatureLength]byte
	SilaData() *silapb.SilaData
	Graffiti() [field_params.RootLength]byte
	ProposerSlashings() []*silapb.ProposerSlashing
	AttesterSlashings() []silapb.AttSlashing
	Attestations() []silapb.Att
	Deposits() []*silapb.Deposit
	VoluntaryExits() []*silapb.SignedVoluntaryExit
	SyncAggregate() (*silapb.SyncAggregate, error)
	IsNil() bool
	HashTreeRoot() ([field_params.RootLength]byte, error)
	Proto() (proto.Message, error)
	Execution() (SilaData, error)
	BLSToSilaChanges() ([]*silapb.SignedBLSToSilaChange, error)
	BlobKzgCommitments() ([][]byte, error)
	SilaRequests() (*silaenginev1.SilaRequests, error)
	PayloadAttestations() ([]*silapb.PayloadAttestation, error)
	SignedSilaPayloadBid() (*silapb.SignedSilaPayloadBid, error)
	ParentSilaRequests() (*silaenginev1.SilaRequests, error)
}

type SignedBeaconBlock interface {
	ReadOnlySignedBeaconBlock
	SetExecution(SilaData) error
	SetBLSToSilaChanges([]*silapb.SignedBLSToSilaChange) error
	SetBlobKzgCommitments(c [][]byte) error
	SetSyncAggregate(*silapb.SyncAggregate) error
	SetVoluntaryExits([]*silapb.SignedVoluntaryExit)
	SetDeposits([]*silapb.Deposit)
	SetAttestations([]silapb.Att) error
	SetAttesterSlashings([]silapb.AttSlashing) error
	SetProposerSlashings([]*silapb.ProposerSlashing)
	SetGraffiti([]byte)
	SetSilaData(*silapb.SilaData)
	SetRandaoReveal([]byte)
	SetStateRoot([]byte)
	SetParentRoot([]byte)
	SetProposerIndex(idx primitives.ValidatorIndex)
	SetSlot(slot primitives.Slot)
	SetSignature(sig []byte)
	SetSilaRequests(er *silaenginev1.SilaRequests) error
	SetPayloadAttestations(pa []*silapb.PayloadAttestation) error
	SetSignedSilaPayloadBid(header *silapb.SignedSilaPayloadBid) error
	SetParentSilaRequests(r *silaenginev1.SilaRequests) error
	Unblind(e SilaData) error
}

// SilaData represents execution layer information that is contained
// within post-Bellatrix beacon block bodies.
type SilaData interface {
	ssz.Marshaler
	ssz.Unmarshaler
	ssz.HashRoot
	IsNil() bool
	IsBlinded() bool
	Proto() proto.Message
	ParentHash() []byte
	FeeRecipient() []byte
	StateRoot() []byte
	ReceiptsRoot() []byte
	LogsBloom() []byte
	PrevRandao() []byte
	BlockNumber() uint64
	GasLimit() uint64
	GasUsed() uint64
	Timestamp() uint64
	ExtraData() []byte
	BaseFeePerGas() []byte
	BlobGasUsed() (uint64, error)
	ExcessBlobGas() (uint64, error)
	BlockHash() []byte
	Transactions() ([][]byte, error)
	TransactionsRoot() ([]byte, error)
	Withdrawals() ([]*silaenginev1.Withdrawal, error)
	WithdrawalsRoot() ([]byte, error)
	BlockAccessList() ([]byte, error)
}
