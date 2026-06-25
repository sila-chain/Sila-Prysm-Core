package blocks

import (
	"bytes"

	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	consensus_types "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/ssz"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/pkg/errors"
	fastssz "github.com/sila-chain/fastssz"
	"google.golang.org/protobuf/proto"
)

// silaPayload is a convenience wrapper around a beacon block body's sila payload data structure
// This wrapper allows us to conform to a common interface so that beacon
// blocks for future forks can also be applied across Sila without issues.
type silaPayload struct {
	p *silaenginev1.SilaPayload
}

// NewWrappedExecutionData creates an appropriate sila payload wrapper based on the incoming type.
func NewWrappedExecutionData(v proto.Message) (interfaces.ExecutionData, error) {
	if v == nil {
		return nil, consensus_types.ErrNilObjectWrapped
	}
	switch pbStruct := v.(type) {
	case *silaenginev1.SilaPayload:
		return WrappedSilaPayload(pbStruct)
	case *silaenginev1.SilaPayloadCapella:
		return WrappedSilaPayloadCapella(pbStruct)
	case *silaenginev1.SilaPayloadCapellaWithValue:
		return WrappedSilaPayloadCapella(pbStruct.Payload)
	case *silaenginev1.SilaPayloadDeneb:
		return WrappedSilaPayloadDeneb(pbStruct)
	case *silaenginev1.SilaPayloadDenebWithValueAndBlobsBundle:
		return WrappedSilaPayloadDeneb(pbStruct.Payload)
	case *silaenginev1.ExecutionBundleElectra:
		// note: no payload changes in electra so using deneb
		return WrappedSilaPayloadDeneb(pbStruct.Payload)
	case *silaenginev1.ExecutionBundleFulu:
		return WrappedSilaPayloadDeneb(pbStruct.Payload)
	case *silaenginev1.SilaPayloadGloas:
		return WrappedSilaPayloadGloas(pbStruct)
	case *silaenginev1.ExecutionBundleGloas:
		return WrappedSilaPayloadGloas(pbStruct.Payload)
	case *silaenginev1.SilaPayloadHeader:
		return WrappedSilaPayloadHeader(pbStruct)
	case *silaenginev1.SilaPayloadHeaderCapella:
		return WrappedSilaPayloadHeaderCapella(pbStruct)
	case *silaenginev1.SilaPayloadHeaderDeneb:
		return WrappedSilaPayloadHeaderDeneb(pbStruct)
	default:
		return nil, errors.Wrapf(ErrUnsupportedVersion, "type %T", pbStruct)
	}
}

var _ interfaces.ExecutionData = &silaPayload{}

// WrappedSilaPayload is a constructor which wraps a protobuf sila payload into an interface.
func WrappedSilaPayload(p *silaenginev1.SilaPayload) (interfaces.ExecutionData, error) {
	w := silaPayload{p: p}
	if w.IsNil() {
		return nil, consensus_types.ErrNilObjectWrapped
	}
	return w, nil
}

// IsNil checks if the underlying data is nil.
func (e silaPayload) IsNil() bool {
	return e.p == nil
}

// IsBlinded returns true if the underlying data is blinded.
func (silaPayload) IsBlinded() bool {
	return false
}

// MarshalSSZ --
func (e silaPayload) MarshalSSZ() ([]byte, error) {
	return e.p.MarshalSSZ()
}

// MarshalSSZTo --
func (e silaPayload) MarshalSSZTo(dst []byte) ([]byte, error) {
	return e.p.MarshalSSZTo(dst)
}

// SizeSSZ --
func (e silaPayload) SizeSSZ() int {
	return e.p.SizeSSZ()
}

// UnmarshalSSZ --
func (e silaPayload) UnmarshalSSZ(buf []byte) error {
	return e.p.UnmarshalSSZ(buf)
}

// HashTreeRoot --
func (e silaPayload) HashTreeRoot() ([32]byte, error) {
	return e.p.HashTreeRoot()
}

// HashTreeRootWith --
func (e silaPayload) HashTreeRootWith(hh *fastssz.Hasher) error {
	return e.p.HashTreeRootWith(hh)
}

// Proto --
func (e silaPayload) Proto() proto.Message {
	return e.p
}

// ParentHash --
func (e silaPayload) ParentHash() []byte {
	return e.p.ParentHash
}

// FeeRecipient --
func (e silaPayload) FeeRecipient() []byte {
	return e.p.FeeRecipient
}

// StateRoot --
func (e silaPayload) StateRoot() []byte {
	return e.p.StateRoot
}

// ReceiptsRoot --
func (e silaPayload) ReceiptsRoot() []byte {
	return e.p.ReceiptsRoot
}

// LogsBloom --
func (e silaPayload) LogsBloom() []byte {
	return e.p.LogsBloom
}

// PrevRandao --
func (e silaPayload) PrevRandao() []byte {
	return e.p.PrevRandao
}

// BlockNumber --
func (e silaPayload) BlockNumber() uint64 {
	return e.p.BlockNumber
}

// GasLimit --
func (e silaPayload) GasLimit() uint64 {
	return e.p.GasLimit
}

// GasUsed --
func (e silaPayload) GasUsed() uint64 {
	return e.p.GasUsed
}

// Timestamp --
func (e silaPayload) Timestamp() uint64 {
	return e.p.Timestamp
}

// ExtraData --
func (e silaPayload) ExtraData() []byte {
	return e.p.ExtraData
}

// BaseFeePerGas --
func (e silaPayload) BaseFeePerGas() []byte {
	return e.p.BaseFeePerGas
}

// BlockHash --
func (e silaPayload) BlockHash() []byte {
	return e.p.BlockHash
}

// Transactions --
func (e silaPayload) Transactions() ([][]byte, error) {
	return e.p.Transactions, nil
}

// TransactionsRoot --
func (silaPayload) TransactionsRoot() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// Withdrawals --
func (silaPayload) Withdrawals() ([]*silaenginev1.Withdrawal, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// WithdrawalsRoot --
func (silaPayload) WithdrawalsRoot() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// BlockAccessList --
func (silaPayload) BlockAccessList() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// BlobGasUsed --
func (e silaPayload) BlobGasUsed() (uint64, error) {
	return 0, consensus_types.ErrUnsupportedField
}

// ExcessBlobGas --
func (e silaPayload) ExcessBlobGas() (uint64, error) {
	return 0, consensus_types.ErrUnsupportedField
}

// silaPayloadHeader is a convenience wrapper around a blinded beacon block body's execution header data structure
// This wrapper allows us to conform to a common interface so that beacon
// blocks for future forks can also be applied across Sila without issues.
type silaPayloadHeader struct {
	p *silaenginev1.SilaPayloadHeader
}

var _ interfaces.ExecutionData = &silaPayloadHeader{}

// WrappedSilaPayloadHeader is a constructor which wraps a protobuf execution header into an interface.
func WrappedSilaPayloadHeader(p *silaenginev1.SilaPayloadHeader) (interfaces.ExecutionData, error) {
	w := silaPayloadHeader{p: p}
	if w.IsNil() {
		return nil, consensus_types.ErrNilObjectWrapped
	}
	return w, nil
}

// IsNil checks if the underlying data is nil.
func (e silaPayloadHeader) IsNil() bool {
	return e.p == nil
}

// IsBlinded returns true if the underlying data is a header.
func (silaPayloadHeader) IsBlinded() bool {
	return true
}

// MarshalSSZ --
func (e silaPayloadHeader) MarshalSSZ() ([]byte, error) {
	return e.p.MarshalSSZ()
}

// MarshalSSZTo --
func (e silaPayloadHeader) MarshalSSZTo(dst []byte) ([]byte, error) {
	return e.p.MarshalSSZTo(dst)
}

// SizeSSZ --
func (e silaPayloadHeader) SizeSSZ() int {
	return e.p.SizeSSZ()
}

// UnmarshalSSZ --
func (e silaPayloadHeader) UnmarshalSSZ(buf []byte) error {
	return e.p.UnmarshalSSZ(buf)
}

// HashTreeRoot --
func (e silaPayloadHeader) HashTreeRoot() ([32]byte, error) {
	return e.p.HashTreeRoot()
}

// HashTreeRootWith --
func (e silaPayloadHeader) HashTreeRootWith(hh *fastssz.Hasher) error {
	return e.p.HashTreeRootWith(hh)
}

// Proto --
func (e silaPayloadHeader) Proto() proto.Message {
	return e.p
}

// ParentHash --
func (e silaPayloadHeader) ParentHash() []byte {
	return e.p.ParentHash
}

// FeeRecipient --
func (e silaPayloadHeader) FeeRecipient() []byte {
	return e.p.FeeRecipient
}

// StateRoot --
func (e silaPayloadHeader) StateRoot() []byte {
	return e.p.StateRoot
}

// ReceiptsRoot --
func (e silaPayloadHeader) ReceiptsRoot() []byte {
	return e.p.ReceiptsRoot
}

// LogsBloom --
func (e silaPayloadHeader) LogsBloom() []byte {
	return e.p.LogsBloom
}

// PrevRandao --
func (e silaPayloadHeader) PrevRandao() []byte {
	return e.p.PrevRandao
}

// BlockNumber --
func (e silaPayloadHeader) BlockNumber() uint64 {
	return e.p.BlockNumber
}

// GasLimit --
func (e silaPayloadHeader) GasLimit() uint64 {
	return e.p.GasLimit
}

// GasUsed --
func (e silaPayloadHeader) GasUsed() uint64 {
	return e.p.GasUsed
}

// Timestamp --
func (e silaPayloadHeader) Timestamp() uint64 {
	return e.p.Timestamp
}

// ExtraData --
func (e silaPayloadHeader) ExtraData() []byte {
	return e.p.ExtraData
}

// BaseFeePerGas --
func (e silaPayloadHeader) BaseFeePerGas() []byte {
	return e.p.BaseFeePerGas
}

// BlockHash --
func (e silaPayloadHeader) BlockHash() []byte {
	return e.p.BlockHash
}

// Transactions --
func (silaPayloadHeader) Transactions() ([][]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// TransactionsRoot --
func (e silaPayloadHeader) TransactionsRoot() ([]byte, error) {
	return e.p.TransactionsRoot, nil
}

// Withdrawals --
func (silaPayloadHeader) Withdrawals() ([]*silaenginev1.Withdrawal, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// WithdrawalsRoot --
func (silaPayloadHeader) WithdrawalsRoot() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// BlockAccessList --
func (silaPayloadHeader) BlockAccessList() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// BlobGasUsed --
func (e silaPayloadHeader) BlobGasUsed() (uint64, error) {
	return 0, consensus_types.ErrUnsupportedField
}

// ExcessBlobGas --
func (e silaPayloadHeader) ExcessBlobGas() (uint64, error) {
	return 0, consensus_types.ErrUnsupportedField
}

// PayloadToHeader converts `payload` into sila payload header format.
func PayloadToHeader(payload interfaces.ExecutionData) (*silaenginev1.SilaPayloadHeader, error) {
	txs, err := payload.Transactions()
	if err != nil {
		return nil, err
	}
	txRoot, err := ssz.TransactionsRoot(txs)
	if err != nil {
		return nil, err
	}
	return &silaenginev1.SilaPayloadHeader{
		ParentHash:       bytesutil.SafeCopyBytes(payload.ParentHash()),
		FeeRecipient:     bytesutil.SafeCopyBytes(payload.FeeRecipient()),
		StateRoot:        bytesutil.SafeCopyBytes(payload.StateRoot()),
		ReceiptsRoot:     bytesutil.SafeCopyBytes(payload.ReceiptsRoot()),
		LogsBloom:        bytesutil.SafeCopyBytes(payload.LogsBloom()),
		PrevRandao:       bytesutil.SafeCopyBytes(payload.PrevRandao()),
		BlockNumber:      payload.BlockNumber(),
		GasLimit:         payload.GasLimit(),
		GasUsed:          payload.GasUsed(),
		Timestamp:        payload.Timestamp(),
		ExtraData:        bytesutil.SafeCopyBytes(payload.ExtraData()),
		BaseFeePerGas:    bytesutil.SafeCopyBytes(payload.BaseFeePerGas()),
		BlockHash:        bytesutil.SafeCopyBytes(payload.BlockHash()),
		TransactionsRoot: txRoot[:],
	}, nil
}

// silaPayloadCapella is a convenience wrapper around a beacon block body's sila payload data structure
// This wrapper allows us to conform to a common interface so that beacon
// blocks for future forks can also be applied across Sila without issues.
type silaPayloadCapella struct {
	p *silaenginev1.SilaPayloadCapella
}

var _ interfaces.ExecutionData = &silaPayloadCapella{}

// WrappedSilaPayloadCapella is a constructor which wraps a protobuf sila payload into an interface.
func WrappedSilaPayloadCapella(p *silaenginev1.SilaPayloadCapella) (interfaces.ExecutionData, error) {
	w := silaPayloadCapella{p: p}
	if w.IsNil() {
		return nil, consensus_types.ErrNilObjectWrapped
	}
	return w, nil
}

// IsNil checks if the underlying data is nil.
func (e silaPayloadCapella) IsNil() bool {
	return e.p == nil
}

// IsBlinded returns true if the underlying data is blinded.
func (silaPayloadCapella) IsBlinded() bool {
	return false
}

// MarshalSSZ --
func (e silaPayloadCapella) MarshalSSZ() ([]byte, error) {
	return e.p.MarshalSSZ()
}

// MarshalSSZTo --
func (e silaPayloadCapella) MarshalSSZTo(dst []byte) ([]byte, error) {
	return e.p.MarshalSSZTo(dst)
}

// SizeSSZ --
func (e silaPayloadCapella) SizeSSZ() int {
	return e.p.SizeSSZ()
}

// UnmarshalSSZ --
func (e silaPayloadCapella) UnmarshalSSZ(buf []byte) error {
	return e.p.UnmarshalSSZ(buf)
}

// HashTreeRoot --
func (e silaPayloadCapella) HashTreeRoot() ([32]byte, error) {
	return e.p.HashTreeRoot()
}

// HashTreeRootWith --
func (e silaPayloadCapella) HashTreeRootWith(hh *fastssz.Hasher) error {
	return e.p.HashTreeRootWith(hh)
}

// Proto --
func (e silaPayloadCapella) Proto() proto.Message {
	return e.p
}

// ParentHash --
func (e silaPayloadCapella) ParentHash() []byte {
	return e.p.ParentHash
}

// FeeRecipient --
func (e silaPayloadCapella) FeeRecipient() []byte {
	return e.p.FeeRecipient
}

// StateRoot --
func (e silaPayloadCapella) StateRoot() []byte {
	return e.p.StateRoot
}

// ReceiptsRoot --
func (e silaPayloadCapella) ReceiptsRoot() []byte {
	return e.p.ReceiptsRoot
}

// LogsBloom --
func (e silaPayloadCapella) LogsBloom() []byte {
	return e.p.LogsBloom
}

// PrevRandao --
func (e silaPayloadCapella) PrevRandao() []byte {
	return e.p.PrevRandao
}

// BlockNumber --
func (e silaPayloadCapella) BlockNumber() uint64 {
	return e.p.BlockNumber
}

// GasLimit --
func (e silaPayloadCapella) GasLimit() uint64 {
	return e.p.GasLimit
}

// GasUsed --
func (e silaPayloadCapella) GasUsed() uint64 {
	return e.p.GasUsed
}

// Timestamp --
func (e silaPayloadCapella) Timestamp() uint64 {
	return e.p.Timestamp
}

// ExtraData --
func (e silaPayloadCapella) ExtraData() []byte {
	return e.p.ExtraData
}

// BaseFeePerGas --
func (e silaPayloadCapella) BaseFeePerGas() []byte {
	return e.p.BaseFeePerGas
}

// BlockHash --
func (e silaPayloadCapella) BlockHash() []byte {
	return e.p.BlockHash
}

// Transactions --
func (e silaPayloadCapella) Transactions() ([][]byte, error) {
	return e.p.Transactions, nil
}

// TransactionsRoot --
func (silaPayloadCapella) TransactionsRoot() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// Withdrawals --
func (e silaPayloadCapella) Withdrawals() ([]*silaenginev1.Withdrawal, error) {
	return e.p.Withdrawals, nil
}

// WithdrawalsRoot --
func (silaPayloadCapella) WithdrawalsRoot() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// BlockAccessList --
func (silaPayloadCapella) BlockAccessList() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// BlobGasUsed --
func (e silaPayloadCapella) BlobGasUsed() (uint64, error) {
	return 0, consensus_types.ErrUnsupportedField
}

// ExcessBlobGas --
func (e silaPayloadCapella) ExcessBlobGas() (uint64, error) {
	return 0, consensus_types.ErrUnsupportedField
}

// silaPayloadHeaderCapella is a convenience wrapper around a blinded beacon block body's execution header data structure
// This wrapper allows us to conform to a common interface so that beacon
// blocks for future forks can also be applied across Sila without issues.
type silaPayloadHeaderCapella struct {
	p *silaenginev1.SilaPayloadHeaderCapella
}

var _ interfaces.ExecutionData = &silaPayloadHeaderCapella{}

// WrappedSilaPayloadHeaderCapella is a constructor which wraps a protobuf execution header into an interface.
func WrappedSilaPayloadHeaderCapella(p *silaenginev1.SilaPayloadHeaderCapella) (interfaces.ExecutionData, error) {
	w := silaPayloadHeaderCapella{p: p}
	if w.IsNil() {
		return nil, consensus_types.ErrNilObjectWrapped
	}
	return w, nil
}

// IsNil checks if the underlying data is nil.
func (e silaPayloadHeaderCapella) IsNil() bool {
	return e.p == nil
}

// IsBlinded returns true if the underlying data is blinded.
func (silaPayloadHeaderCapella) IsBlinded() bool {
	return true
}

// MarshalSSZ --
func (e silaPayloadHeaderCapella) MarshalSSZ() ([]byte, error) {
	return e.p.MarshalSSZ()
}

// MarshalSSZTo --
func (e silaPayloadHeaderCapella) MarshalSSZTo(dst []byte) ([]byte, error) {
	return e.p.MarshalSSZTo(dst)
}

// SizeSSZ --
func (e silaPayloadHeaderCapella) SizeSSZ() int {
	return e.p.SizeSSZ()
}

// UnmarshalSSZ --
func (e silaPayloadHeaderCapella) UnmarshalSSZ(buf []byte) error {
	return e.p.UnmarshalSSZ(buf)
}

// HashTreeRoot --
func (e silaPayloadHeaderCapella) HashTreeRoot() ([32]byte, error) {
	return e.p.HashTreeRoot()
}

// HashTreeRootWith --
func (e silaPayloadHeaderCapella) HashTreeRootWith(hh *fastssz.Hasher) error {
	return e.p.HashTreeRootWith(hh)
}

// Proto --
func (e silaPayloadHeaderCapella) Proto() proto.Message {
	return e.p
}

// ParentHash --
func (e silaPayloadHeaderCapella) ParentHash() []byte {
	return e.p.ParentHash
}

// FeeRecipient --
func (e silaPayloadHeaderCapella) FeeRecipient() []byte {
	return e.p.FeeRecipient
}

// StateRoot --
func (e silaPayloadHeaderCapella) StateRoot() []byte {
	return e.p.StateRoot
}

// ReceiptsRoot --
func (e silaPayloadHeaderCapella) ReceiptsRoot() []byte {
	return e.p.ReceiptsRoot
}

// LogsBloom --
func (e silaPayloadHeaderCapella) LogsBloom() []byte {
	return e.p.LogsBloom
}

// PrevRandao --
func (e silaPayloadHeaderCapella) PrevRandao() []byte {
	return e.p.PrevRandao
}

// BlockNumber --
func (e silaPayloadHeaderCapella) BlockNumber() uint64 {
	return e.p.BlockNumber
}

// GasLimit --
func (e silaPayloadHeaderCapella) GasLimit() uint64 {
	return e.p.GasLimit
}

// GasUsed --
func (e silaPayloadHeaderCapella) GasUsed() uint64 {
	return e.p.GasUsed
}

// Timestamp --
func (e silaPayloadHeaderCapella) Timestamp() uint64 {
	return e.p.Timestamp
}

// ExtraData --
func (e silaPayloadHeaderCapella) ExtraData() []byte {
	return e.p.ExtraData
}

// BaseFeePerGas --
func (e silaPayloadHeaderCapella) BaseFeePerGas() []byte {
	return e.p.BaseFeePerGas
}

// BlockHash --
func (e silaPayloadHeaderCapella) BlockHash() []byte {
	return e.p.BlockHash
}

// Transactions --
func (silaPayloadHeaderCapella) Transactions() ([][]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// TransactionsRoot --
func (e silaPayloadHeaderCapella) TransactionsRoot() ([]byte, error) {
	return e.p.TransactionsRoot, nil
}

// Withdrawals --
func (silaPayloadHeaderCapella) Withdrawals() ([]*silaenginev1.Withdrawal, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// WithdrawalsRoot --
func (e silaPayloadHeaderCapella) WithdrawalsRoot() ([]byte, error) {
	return e.p.WithdrawalsRoot, nil
}

// BlockAccessList --
func (silaPayloadHeaderCapella) BlockAccessList() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// BlobGasUsed --
func (e silaPayloadHeaderCapella) BlobGasUsed() (uint64, error) {
	return 0, consensus_types.ErrUnsupportedField
}

// ExcessBlobGas --
func (e silaPayloadHeaderCapella) ExcessBlobGas() (uint64, error) {
	return 0, consensus_types.ErrUnsupportedField
}

// PayloadToHeaderCapella converts `payload` into sila payload header format.
func PayloadToHeaderCapella(payload interfaces.ExecutionData) (*silaenginev1.SilaPayloadHeaderCapella, error) {
	txs, err := payload.Transactions()
	if err != nil {
		return nil, err
	}
	txRoot, err := ssz.TransactionsRoot(txs)
	if err != nil {
		return nil, err
	}
	withdrawals, err := payload.Withdrawals()
	if err != nil {
		return nil, err
	}
	withdrawalsRoot, err := ssz.WithdrawalSliceRoot(withdrawals, fieldparams.MaxWithdrawalsPerPayload)
	if err != nil {
		return nil, err
	}

	return &silaenginev1.SilaPayloadHeaderCapella{
		ParentHash:       bytesutil.SafeCopyBytes(payload.ParentHash()),
		FeeRecipient:     bytesutil.SafeCopyBytes(payload.FeeRecipient()),
		StateRoot:        bytesutil.SafeCopyBytes(payload.StateRoot()),
		ReceiptsRoot:     bytesutil.SafeCopyBytes(payload.ReceiptsRoot()),
		LogsBloom:        bytesutil.SafeCopyBytes(payload.LogsBloom()),
		PrevRandao:       bytesutil.SafeCopyBytes(payload.PrevRandao()),
		BlockNumber:      payload.BlockNumber(),
		GasLimit:         payload.GasLimit(),
		GasUsed:          payload.GasUsed(),
		Timestamp:        payload.Timestamp(),
		ExtraData:        bytesutil.SafeCopyBytes(payload.ExtraData()),
		BaseFeePerGas:    bytesutil.SafeCopyBytes(payload.BaseFeePerGas()),
		BlockHash:        bytesutil.SafeCopyBytes(payload.BlockHash()),
		TransactionsRoot: txRoot[:],
		WithdrawalsRoot:  withdrawalsRoot[:],
	}, nil
}

// PayloadToHeaderDeneb converts `payload` into sila payload header format.
func PayloadToHeaderDeneb(payload interfaces.ExecutionData) (*silaenginev1.SilaPayloadHeaderDeneb, error) {
	txs, err := payload.Transactions()
	if err != nil {
		return nil, err
	}
	txRoot, err := ssz.TransactionsRoot(txs)
	if err != nil {
		return nil, err
	}
	withdrawals, err := payload.Withdrawals()
	if err != nil {
		return nil, err
	}
	withdrawalsRoot, err := ssz.WithdrawalSliceRoot(withdrawals, fieldparams.MaxWithdrawalsPerPayload)
	if err != nil {
		return nil, err
	}
	blobGasUsed, err := payload.BlobGasUsed()
	if err != nil {
		return nil, err
	}
	excessBlobGas, err := payload.ExcessBlobGas()
	if err != nil {
		return nil, err
	}

	return &silaenginev1.SilaPayloadHeaderDeneb{
		ParentHash:       bytesutil.SafeCopyBytes(payload.ParentHash()),
		FeeRecipient:     bytesutil.SafeCopyBytes(payload.FeeRecipient()),
		StateRoot:        bytesutil.SafeCopyBytes(payload.StateRoot()),
		ReceiptsRoot:     bytesutil.SafeCopyBytes(payload.ReceiptsRoot()),
		LogsBloom:        bytesutil.SafeCopyBytes(payload.LogsBloom()),
		PrevRandao:       bytesutil.SafeCopyBytes(payload.PrevRandao()),
		BlockNumber:      payload.BlockNumber(),
		GasLimit:         payload.GasLimit(),
		GasUsed:          payload.GasUsed(),
		Timestamp:        payload.Timestamp(),
		ExtraData:        bytesutil.SafeCopyBytes(payload.ExtraData()),
		BaseFeePerGas:    bytesutil.SafeCopyBytes(payload.BaseFeePerGas()),
		BlockHash:        bytesutil.SafeCopyBytes(payload.BlockHash()),
		TransactionsRoot: txRoot[:],
		WithdrawalsRoot:  withdrawalsRoot[:],
		BlobGasUsed:      blobGasUsed,
		ExcessBlobGas:    excessBlobGas,
	}, nil
}

var (
	PayloadToHeaderElectra = PayloadToHeaderDeneb
	PayloadToHeaderFulu    = PayloadToHeaderDeneb
)

// IsEmptyExecutionData checks if an execution data is empty underneath. If a single field has
// a non-zero value, this function will return false.
func IsEmptyExecutionData(data interfaces.ExecutionData) (bool, error) {
	if data == nil {
		return true, nil
	}
	if !bytes.Equal(data.ParentHash(), make([]byte, fieldparams.RootLength)) {
		return false, nil
	}
	if !bytes.Equal(data.FeeRecipient(), make([]byte, fieldparams.FeeRecipientLength)) {
		return false, nil
	}
	if !bytes.Equal(data.StateRoot(), make([]byte, fieldparams.RootLength)) {
		return false, nil
	}
	if !bytes.Equal(data.ReceiptsRoot(), make([]byte, fieldparams.RootLength)) {
		return false, nil
	}
	if !bytes.Equal(data.LogsBloom(), make([]byte, fieldparams.LogsBloomLength)) {
		return false, nil
	}
	if !bytes.Equal(data.PrevRandao(), make([]byte, fieldparams.RootLength)) {
		return false, nil
	}
	if !bytes.Equal(data.BaseFeePerGas(), make([]byte, fieldparams.RootLength)) {
		return false, nil
	}
	if !bytes.Equal(data.BlockHash(), make([]byte, fieldparams.RootLength)) {
		return false, nil
	}

	txs, err := data.Transactions()
	switch {
	case errors.Is(err, consensus_types.ErrUnsupportedField):
	case err != nil:
		return false, err
	default:
		if len(txs) != 0 {
			return false, nil
		}
	}

	if len(data.ExtraData()) != 0 {
		return false, nil
	}
	if data.BlockNumber() != 0 {
		return false, nil
	}
	if data.GasLimit() != 0 {
		return false, nil
	}
	if data.GasUsed() != 0 {
		return false, nil
	}
	if data.Timestamp() != 0 {
		return false, nil
	}
	return true, nil
}

// silaPayloadHeaderDeneb is a convenience wrapper around a blinded beacon block body's execution header data structure.
// This wrapper allows us to conform to a common interface so that beacon
// blocks for future forks can also be applied across Sila without issues.
type silaPayloadHeaderDeneb struct {
	p *silaenginev1.SilaPayloadHeaderDeneb
}

var _ interfaces.ExecutionData = &silaPayloadHeaderDeneb{}

// WrappedSilaPayloadHeaderDeneb is a constructor which wraps a protobuf execution header into an interface.
func WrappedSilaPayloadHeaderDeneb(p *silaenginev1.SilaPayloadHeaderDeneb) (interfaces.ExecutionData, error) {
	w := silaPayloadHeaderDeneb{p: p}
	if w.IsNil() {
		return nil, consensus_types.ErrNilObjectWrapped
	}
	return w, nil
}

// IsNil checks if the underlying data is nil.
func (e silaPayloadHeaderDeneb) IsNil() bool {
	return e.p == nil
}

// MarshalSSZ --
func (e silaPayloadHeaderDeneb) MarshalSSZ() ([]byte, error) {
	return e.p.MarshalSSZ()
}

// MarshalSSZTo --
func (e silaPayloadHeaderDeneb) MarshalSSZTo(dst []byte) ([]byte, error) {
	return e.p.MarshalSSZTo(dst)
}

// SizeSSZ --
func (e silaPayloadHeaderDeneb) SizeSSZ() int {
	return e.p.SizeSSZ()
}

// UnmarshalSSZ --
func (e silaPayloadHeaderDeneb) UnmarshalSSZ(buf []byte) error {
	return e.p.UnmarshalSSZ(buf)
}

// HashTreeRoot --
func (e silaPayloadHeaderDeneb) HashTreeRoot() ([32]byte, error) {
	return e.p.HashTreeRoot()
}

// HashTreeRootWith --
func (e silaPayloadHeaderDeneb) HashTreeRootWith(hh *fastssz.Hasher) error {
	return e.p.HashTreeRootWith(hh)
}

// Proto --
func (e silaPayloadHeaderDeneb) Proto() proto.Message {
	return e.p
}

// ParentHash --
func (e silaPayloadHeaderDeneb) ParentHash() []byte {
	return e.p.ParentHash
}

// FeeRecipient --
func (e silaPayloadHeaderDeneb) FeeRecipient() []byte {
	return e.p.FeeRecipient
}

// StateRoot --
func (e silaPayloadHeaderDeneb) StateRoot() []byte {
	return e.p.StateRoot
}

// ReceiptsRoot --
func (e silaPayloadHeaderDeneb) ReceiptsRoot() []byte {
	return e.p.ReceiptsRoot
}

// LogsBloom --
func (e silaPayloadHeaderDeneb) LogsBloom() []byte {
	return e.p.LogsBloom
}

// PrevRandao --
func (e silaPayloadHeaderDeneb) PrevRandao() []byte {
	return e.p.PrevRandao
}

// BlockNumber --
func (e silaPayloadHeaderDeneb) BlockNumber() uint64 {
	return e.p.BlockNumber
}

// GasLimit --
func (e silaPayloadHeaderDeneb) GasLimit() uint64 {
	return e.p.GasLimit
}

// GasUsed --
func (e silaPayloadHeaderDeneb) GasUsed() uint64 {
	return e.p.GasUsed
}

// Timestamp --
func (e silaPayloadHeaderDeneb) Timestamp() uint64 {
	return e.p.Timestamp
}

// ExtraData --
func (e silaPayloadHeaderDeneb) ExtraData() []byte {
	return e.p.ExtraData
}

// BaseFeePerGas --
func (e silaPayloadHeaderDeneb) BaseFeePerGas() []byte {
	return e.p.BaseFeePerGas
}

// BlockHash --
func (e silaPayloadHeaderDeneb) BlockHash() []byte {
	return e.p.BlockHash
}

// Transactions --
func (silaPayloadHeaderDeneb) Transactions() ([][]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// TransactionsRoot --
func (e silaPayloadHeaderDeneb) TransactionsRoot() ([]byte, error) {
	return e.p.TransactionsRoot, nil
}

// Withdrawals --
func (e silaPayloadHeaderDeneb) Withdrawals() ([]*silaenginev1.Withdrawal, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// WithdrawalsRoot --
func (e silaPayloadHeaderDeneb) WithdrawalsRoot() ([]byte, error) {
	return e.p.WithdrawalsRoot, nil
}

// BlockAccessList --
func (silaPayloadHeaderDeneb) BlockAccessList() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// BlobGasUsed --
func (e silaPayloadHeaderDeneb) BlobGasUsed() (uint64, error) {
	return e.p.BlobGasUsed, nil
}

// ExcessBlobGas --
func (e silaPayloadHeaderDeneb) ExcessBlobGas() (uint64, error) {
	return e.p.ExcessBlobGas, nil
}

// IsBlinded returns true if the underlying data is blinded.
func (e silaPayloadHeaderDeneb) IsBlinded() bool {
	return true
}

// silaPayloadDeneb is a convenience wrapper around a beacon block body's sila payload data structure
// This wrapper allows us to conform to a common interface so that beacon
// blocks for future forks can also be applied across Sila without issues.
type silaPayloadDeneb struct {
	p *silaenginev1.SilaPayloadDeneb
}

var _ interfaces.ExecutionData = &silaPayloadDeneb{}

// WrappedSilaPayloadDeneb is a constructor which wraps a protobuf sila payload into an interface.
func WrappedSilaPayloadDeneb(p *silaenginev1.SilaPayloadDeneb) (interfaces.ExecutionData, error) {
	w := silaPayloadDeneb{p: p}
	if w.IsNil() {
		return nil, consensus_types.ErrNilObjectWrapped
	}
	return w, nil
}

// IsNil checks if the underlying data is nil.
func (e silaPayloadDeneb) IsNil() bool {
	return e.p == nil
}

// MarshalSSZ --
func (e silaPayloadDeneb) MarshalSSZ() ([]byte, error) {
	return e.p.MarshalSSZ()
}

// MarshalSSZTo --
func (e silaPayloadDeneb) MarshalSSZTo(dst []byte) ([]byte, error) {
	return e.p.MarshalSSZTo(dst)
}

// SizeSSZ --
func (e silaPayloadDeneb) SizeSSZ() int {
	return e.p.SizeSSZ()
}

// UnmarshalSSZ --
func (e silaPayloadDeneb) UnmarshalSSZ(buf []byte) error {
	return e.p.UnmarshalSSZ(buf)
}

// HashTreeRoot --
func (e silaPayloadDeneb) HashTreeRoot() ([32]byte, error) {
	return e.p.HashTreeRoot()
}

// HashTreeRootWith --
func (e silaPayloadDeneb) HashTreeRootWith(hh *fastssz.Hasher) error {
	return e.p.HashTreeRootWith(hh)
}

// Proto --
func (e silaPayloadDeneb) Proto() proto.Message {
	return e.p
}

// ParentHash --
func (e silaPayloadDeneb) ParentHash() []byte {
	return e.p.ParentHash
}

// FeeRecipient --
func (e silaPayloadDeneb) FeeRecipient() []byte {
	return e.p.FeeRecipient
}

// StateRoot --
func (e silaPayloadDeneb) StateRoot() []byte {
	return e.p.StateRoot
}

// ReceiptsRoot --
func (e silaPayloadDeneb) ReceiptsRoot() []byte {
	return e.p.ReceiptsRoot
}

// LogsBloom --
func (e silaPayloadDeneb) LogsBloom() []byte {
	return e.p.LogsBloom
}

// PrevRandao --
func (e silaPayloadDeneb) PrevRandao() []byte {
	return e.p.PrevRandao
}

// BlockNumber --
func (e silaPayloadDeneb) BlockNumber() uint64 {
	return e.p.BlockNumber
}

// GasLimit --
func (e silaPayloadDeneb) GasLimit() uint64 {
	return e.p.GasLimit
}

// GasUsed --
func (e silaPayloadDeneb) GasUsed() uint64 {
	return e.p.GasUsed
}

// Timestamp --
func (e silaPayloadDeneb) Timestamp() uint64 {
	return e.p.Timestamp
}

// ExtraData --
func (e silaPayloadDeneb) ExtraData() []byte {
	return e.p.ExtraData
}

// BaseFeePerGas --
func (e silaPayloadDeneb) BaseFeePerGas() []byte {
	return e.p.BaseFeePerGas
}

// BlockHash --
func (e silaPayloadDeneb) BlockHash() []byte {
	return e.p.BlockHash
}

// Transactions --
func (e silaPayloadDeneb) Transactions() ([][]byte, error) {
	return e.p.Transactions, nil
}

// TransactionsRoot --
func (e silaPayloadDeneb) TransactionsRoot() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// Withdrawals --
func (e silaPayloadDeneb) Withdrawals() ([]*silaenginev1.Withdrawal, error) {
	return e.p.Withdrawals, nil
}

// WithdrawalsRoot --
func (e silaPayloadDeneb) WithdrawalsRoot() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// BlockAccessList --
func (silaPayloadDeneb) BlockAccessList() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

func (e silaPayloadDeneb) BlobGasUsed() (uint64, error) {
	return e.p.BlobGasUsed, nil
}

func (e silaPayloadDeneb) ExcessBlobGas() (uint64, error) {
	return e.p.ExcessBlobGas, nil
}

// IsBlinded returns true if the underlying data is blinded.
func (e silaPayloadDeneb) IsBlinded() bool {
	return false
}

// silaPayloadGloas is a convenience wrapper around a beacon block body's sila payload data structure
// This wrapper allows us to conform to a common interface so that beacon
// blocks for future forks can also be applied across Sila without issues.
type silaPayloadGloas struct {
	p *silaenginev1.SilaPayloadGloas
}

var _ interfaces.ExecutionData = &silaPayloadGloas{}

// WrappedSilaPayloadGloas is a constructor which wraps a protobuf sila payload into an interface.
func WrappedSilaPayloadGloas(p *silaenginev1.SilaPayloadGloas) (interfaces.ExecutionData, error) {
	w := silaPayloadGloas{p: p}
	if w.IsNil() {
		return nil, consensus_types.ErrNilObjectWrapped
	}
	return w, nil
}

// IsNil checks if the underlying data is nil.
func (e silaPayloadGloas) IsNil() bool {
	return e.p == nil
}

// MarshalSSZ --
func (e silaPayloadGloas) MarshalSSZ() ([]byte, error) {
	return e.p.MarshalSSZ()
}

// MarshalSSZTo --
func (e silaPayloadGloas) MarshalSSZTo(dst []byte) ([]byte, error) {
	return e.p.MarshalSSZTo(dst)
}

// SizeSSZ --
func (e silaPayloadGloas) SizeSSZ() int {
	return e.p.SizeSSZ()
}

// UnmarshalSSZ --
func (e silaPayloadGloas) UnmarshalSSZ(buf []byte) error {
	return e.p.UnmarshalSSZ(buf)
}

// HashTreeRoot --
func (e silaPayloadGloas) HashTreeRoot() ([32]byte, error) {
	return e.p.HashTreeRoot()
}

// HashTreeRootWith --
func (e silaPayloadGloas) HashTreeRootWith(hh *fastssz.Hasher) error {
	return e.p.HashTreeRootWith(hh)
}

// Proto --
func (e silaPayloadGloas) Proto() proto.Message {
	return e.p
}

// ParentHash --
func (e silaPayloadGloas) ParentHash() []byte {
	return e.p.ParentHash
}

// FeeRecipient --
func (e silaPayloadGloas) FeeRecipient() []byte {
	return e.p.FeeRecipient
}

// StateRoot --
func (e silaPayloadGloas) StateRoot() []byte {
	return e.p.StateRoot
}

// ReceiptsRoot --
func (e silaPayloadGloas) ReceiptsRoot() []byte {
	return e.p.ReceiptsRoot
}

// LogsBloom --
func (e silaPayloadGloas) LogsBloom() []byte {
	return e.p.LogsBloom
}

// PrevRandao --
func (e silaPayloadGloas) PrevRandao() []byte {
	return e.p.PrevRandao
}

// BlockNumber --
func (e silaPayloadGloas) BlockNumber() uint64 {
	return e.p.BlockNumber
}

// GasLimit --
func (e silaPayloadGloas) GasLimit() uint64 {
	return e.p.GasLimit
}

// GasUsed --
func (e silaPayloadGloas) GasUsed() uint64 {
	return e.p.GasUsed
}

// Timestamp --
func (e silaPayloadGloas) Timestamp() uint64 {
	return e.p.Timestamp
}

// ExtraData --
func (e silaPayloadGloas) ExtraData() []byte {
	return e.p.ExtraData
}

// BaseFeePerGas --
func (e silaPayloadGloas) BaseFeePerGas() []byte {
	return e.p.BaseFeePerGas
}

// BlockHash --
func (e silaPayloadGloas) BlockHash() []byte {
	return e.p.BlockHash
}

// Transactions --
func (e silaPayloadGloas) Transactions() ([][]byte, error) {
	return e.p.Transactions, nil
}

// TransactionsRoot --
func (silaPayloadGloas) TransactionsRoot() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// Withdrawals --
func (e silaPayloadGloas) Withdrawals() ([]*silaenginev1.Withdrawal, error) {
	return e.p.Withdrawals, nil
}

// WithdrawalsRoot --
func (silaPayloadGloas) WithdrawalsRoot() ([]byte, error) {
	return nil, consensus_types.ErrUnsupportedField
}

// BlockAccessList --
func (e silaPayloadGloas) BlockAccessList() ([]byte, error) {
	return e.p.BlockAccessList, nil
}

// BlobGasUsed --
func (e silaPayloadGloas) BlobGasUsed() (uint64, error) {
	return e.p.BlobGasUsed, nil
}

// ExcessBlobGas --
func (e silaPayloadGloas) ExcessBlobGas() (uint64, error) {
	return e.p.ExcessBlobGas, nil
}

// IsBlinded returns true if the underlying data is blinded.
func (e silaPayloadGloas) IsBlinded() bool {
	return false
}
