package silaenginev1

import "github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"

type copier[T any] interface {
	Copy() T
}

func copySlice[T any, C copier[T]](original []C) []T {
	// Create a new slice with the same length as the original
	newSlice := make([]T, len(original))
	for i := range newSlice {
		newSlice[i] = original[i].Copy()
	}
	return newSlice
}

// Copy --
func (w *Withdrawal) Copy() *Withdrawal {
	if w == nil {
		return nil
	}

	return &Withdrawal{
		Index:          w.Index,
		ValidatorIndex: w.ValidatorIndex,
		Address:        bytesutil.SafeCopyBytes(w.Address),
		Amount:         w.Amount,
	}
}

// Copy --
func (d *DepositRequest) Copy() *DepositRequest {
	if d == nil {
		return nil
	}
	return &DepositRequest{
		Pubkey:                bytesutil.SafeCopyBytes(d.Pubkey),
		WithdrawalCredentials: bytesutil.SafeCopyBytes(d.WithdrawalCredentials),
		Amount:                d.Amount,
		Signature:             bytesutil.SafeCopyBytes(d.Signature),
		Index:                 d.Index,
	}
}

// Copy --
func (wr *WithdrawalRequest) Copy() *WithdrawalRequest {
	if wr == nil {
		return nil
	}
	return &WithdrawalRequest{
		SourceAddress:   bytesutil.SafeCopyBytes(wr.SourceAddress),
		ValidatorPubkey: bytesutil.SafeCopyBytes(wr.ValidatorPubkey),
		Amount:          wr.Amount,
	}
}

// Copy --
func (cr *ConsolidationRequest) Copy() *ConsolidationRequest {
	if cr == nil {
		return nil
	}
	return &ConsolidationRequest{
		SourceAddress: bytesutil.SafeCopyBytes(cr.SourceAddress),
		SourcePubkey:  bytesutil.SafeCopyBytes(cr.SourcePubkey),
		TargetPubkey:  bytesutil.SafeCopyBytes(cr.TargetPubkey),
	}
}

// Copy -- Deneb
func (payload *SilaPayloadDeneb) Copy() *SilaPayloadDeneb {
	if payload == nil {
		return nil
	}
	return &SilaPayloadDeneb{
		ParentHash:    bytesutil.SafeCopyBytes(payload.ParentHash),
		FeeRecipient:  bytesutil.SafeCopyBytes(payload.FeeRecipient),
		StateRoot:     bytesutil.SafeCopyBytes(payload.StateRoot),
		ReceiptsRoot:  bytesutil.SafeCopyBytes(payload.ReceiptsRoot),
		LogsBloom:     bytesutil.SafeCopyBytes(payload.LogsBloom),
		PrevRandao:    bytesutil.SafeCopyBytes(payload.PrevRandao),
		BlockNumber:   payload.BlockNumber,
		GasLimit:      payload.GasLimit,
		GasUsed:       payload.GasUsed,
		Timestamp:     payload.Timestamp,
		ExtraData:     bytesutil.SafeCopyBytes(payload.ExtraData),
		BaseFeePerGas: bytesutil.SafeCopyBytes(payload.BaseFeePerGas),
		BlockHash:     bytesutil.SafeCopyBytes(payload.BlockHash),
		Transactions:  bytesutil.SafeCopy2dBytes(payload.Transactions),
		Withdrawals:   copySlice(payload.Withdrawals),
		BlobGasUsed:   payload.BlobGasUsed,
		ExcessBlobGas: payload.ExcessBlobGas,
	}
}

// Copy -- Capella
func (payload *SilaPayloadCapella) Copy() *SilaPayloadCapella {
	if payload == nil {
		return nil
	}

	return &SilaPayloadCapella{
		ParentHash:    bytesutil.SafeCopyBytes(payload.ParentHash),
		FeeRecipient:  bytesutil.SafeCopyBytes(payload.FeeRecipient),
		StateRoot:     bytesutil.SafeCopyBytes(payload.StateRoot),
		ReceiptsRoot:  bytesutil.SafeCopyBytes(payload.ReceiptsRoot),
		LogsBloom:     bytesutil.SafeCopyBytes(payload.LogsBloom),
		PrevRandao:    bytesutil.SafeCopyBytes(payload.PrevRandao),
		BlockNumber:   payload.BlockNumber,
		GasLimit:      payload.GasLimit,
		GasUsed:       payload.GasUsed,
		Timestamp:     payload.Timestamp,
		ExtraData:     bytesutil.SafeCopyBytes(payload.ExtraData),
		BaseFeePerGas: bytesutil.SafeCopyBytes(payload.BaseFeePerGas),
		BlockHash:     bytesutil.SafeCopyBytes(payload.BlockHash),
		Transactions:  bytesutil.SafeCopy2dBytes(payload.Transactions),
		Withdrawals:   copySlice(payload.Withdrawals),
	}
}

// Copy -- Bellatrix
func (payload *SilaPayload) Copy() *SilaPayload {
	if payload == nil {
		return nil
	}

	return &SilaPayload{
		ParentHash:    bytesutil.SafeCopyBytes(payload.ParentHash),
		FeeRecipient:  bytesutil.SafeCopyBytes(payload.FeeRecipient),
		StateRoot:     bytesutil.SafeCopyBytes(payload.StateRoot),
		ReceiptsRoot:  bytesutil.SafeCopyBytes(payload.ReceiptsRoot),
		LogsBloom:     bytesutil.SafeCopyBytes(payload.LogsBloom),
		PrevRandao:    bytesutil.SafeCopyBytes(payload.PrevRandao),
		BlockNumber:   payload.BlockNumber,
		GasLimit:      payload.GasLimit,
		GasUsed:       payload.GasUsed,
		Timestamp:     payload.Timestamp,
		ExtraData:     bytesutil.SafeCopyBytes(payload.ExtraData),
		BaseFeePerGas: bytesutil.SafeCopyBytes(payload.BaseFeePerGas),
		BlockHash:     bytesutil.SafeCopyBytes(payload.BlockHash),
		Transactions:  bytesutil.SafeCopy2dBytes(payload.Transactions),
	}
}

// Copy -- Deneb
func (payload *SilaPayloadHeaderDeneb) Copy() *SilaPayloadHeaderDeneb {
	if payload == nil {
		return nil
	}
	return &SilaPayloadHeaderDeneb{
		ParentHash:       bytesutil.SafeCopyBytes(payload.ParentHash),
		FeeRecipient:     bytesutil.SafeCopyBytes(payload.FeeRecipient),
		StateRoot:        bytesutil.SafeCopyBytes(payload.StateRoot),
		ReceiptsRoot:     bytesutil.SafeCopyBytes(payload.ReceiptsRoot),
		LogsBloom:        bytesutil.SafeCopyBytes(payload.LogsBloom),
		PrevRandao:       bytesutil.SafeCopyBytes(payload.PrevRandao),
		BlockNumber:      payload.BlockNumber,
		GasLimit:         payload.GasLimit,
		GasUsed:          payload.GasUsed,
		Timestamp:        payload.Timestamp,
		ExtraData:        bytesutil.SafeCopyBytes(payload.ExtraData),
		BaseFeePerGas:    bytesutil.SafeCopyBytes(payload.BaseFeePerGas),
		BlockHash:        bytesutil.SafeCopyBytes(payload.BlockHash),
		TransactionsRoot: bytesutil.SafeCopyBytes(payload.TransactionsRoot),
		WithdrawalsRoot:  bytesutil.SafeCopyBytes(payload.WithdrawalsRoot),
		BlobGasUsed:      payload.BlobGasUsed,
		ExcessBlobGas:    payload.ExcessBlobGas,
	}
}

// Copy -- Capella
func (payload *SilaPayloadHeaderCapella) Copy() *SilaPayloadHeaderCapella {
	if payload == nil {
		return nil
	}
	return &SilaPayloadHeaderCapella{
		ParentHash:       bytesutil.SafeCopyBytes(payload.ParentHash),
		FeeRecipient:     bytesutil.SafeCopyBytes(payload.FeeRecipient),
		StateRoot:        bytesutil.SafeCopyBytes(payload.StateRoot),
		ReceiptsRoot:     bytesutil.SafeCopyBytes(payload.ReceiptsRoot),
		LogsBloom:        bytesutil.SafeCopyBytes(payload.LogsBloom),
		PrevRandao:       bytesutil.SafeCopyBytes(payload.PrevRandao),
		BlockNumber:      payload.BlockNumber,
		GasLimit:         payload.GasLimit,
		GasUsed:          payload.GasUsed,
		Timestamp:        payload.Timestamp,
		ExtraData:        bytesutil.SafeCopyBytes(payload.ExtraData),
		BaseFeePerGas:    bytesutil.SafeCopyBytes(payload.BaseFeePerGas),
		BlockHash:        bytesutil.SafeCopyBytes(payload.BlockHash),
		TransactionsRoot: bytesutil.SafeCopyBytes(payload.TransactionsRoot),
		WithdrawalsRoot:  bytesutil.SafeCopyBytes(payload.WithdrawalsRoot),
	}
}

// Copy -- Bellatrix
func (payload *SilaPayloadHeader) Copy() *SilaPayloadHeader {
	if payload == nil {
		return nil
	}
	return &SilaPayloadHeader{
		ParentHash:       bytesutil.SafeCopyBytes(payload.ParentHash),
		FeeRecipient:     bytesutil.SafeCopyBytes(payload.FeeRecipient),
		StateRoot:        bytesutil.SafeCopyBytes(payload.StateRoot),
		ReceiptsRoot:     bytesutil.SafeCopyBytes(payload.ReceiptsRoot),
		LogsBloom:        bytesutil.SafeCopyBytes(payload.LogsBloom),
		PrevRandao:       bytesutil.SafeCopyBytes(payload.PrevRandao),
		BlockNumber:      payload.BlockNumber,
		GasLimit:         payload.GasLimit,
		GasUsed:          payload.GasUsed,
		Timestamp:        payload.Timestamp,
		ExtraData:        bytesutil.SafeCopyBytes(payload.ExtraData),
		BaseFeePerGas:    bytesutil.SafeCopyBytes(payload.BaseFeePerGas),
		BlockHash:        bytesutil.SafeCopyBytes(payload.BlockHash),
		TransactionsRoot: bytesutil.SafeCopyBytes(payload.TransactionsRoot),
	}
}
