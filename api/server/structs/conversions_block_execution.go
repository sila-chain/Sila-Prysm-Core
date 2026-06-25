package structs

import (
	"fmt"
	"strconv"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	consensusblocks "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/container/slice"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila/common"
	"github.com/sila-chain/Sila/common/hexutil"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// ----------------------------------------------------------------------------
// Bellatrix
// ----------------------------------------------------------------------------

func SilaPayloadFromConsensus(payload *silaenginev1.SilaPayload) (*SilaPayload, error) {
	baseFeePerGas, err := sszBytesToUint256String(payload.BaseFeePerGas)
	if err != nil {
		return nil, err
	}
	transactions := make([]string, len(payload.Transactions))
	for i, tx := range payload.Transactions {
		transactions[i] = hexutil.Encode(tx)
	}

	return &SilaPayload{
		ParentHash:    hexutil.Encode(payload.ParentHash),
		FeeRecipient:  hexutil.Encode(payload.FeeRecipient),
		StateRoot:     hexutil.Encode(payload.StateRoot),
		ReceiptsRoot:  hexutil.Encode(payload.ReceiptsRoot),
		LogsBloom:     hexutil.Encode(payload.LogsBloom),
		PrevRandao:    hexutil.Encode(payload.PrevRandao),
		BlockNumber:   fmt.Sprintf("%d", payload.BlockNumber),
		GasLimit:      fmt.Sprintf("%d", payload.GasLimit),
		GasUsed:       fmt.Sprintf("%d", payload.GasUsed),
		Timestamp:     fmt.Sprintf("%d", payload.Timestamp),
		ExtraData:     hexutil.Encode(payload.ExtraData),
		BaseFeePerGas: baseFeePerGas,
		BlockHash:     hexutil.Encode(payload.BlockHash),
		Transactions:  transactions,
	}, nil
}

func (e *SilaPayload) ToConsensus() (*silaenginev1.SilaPayload, error) {
	if e == nil {
		return nil, server.NewDecodeError(errNilValue, "SilaPayload")
	}
	payloadParentHash, err := bytesutil.DecodeHexWithLength(e.ParentHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.ParentHash")
	}
	payloadFeeRecipient, err := bytesutil.DecodeHexWithLength(e.FeeRecipient, fieldparams.FeeRecipientLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.FeeRecipient")
	}
	payloadStateRoot, err := bytesutil.DecodeHexWithLength(e.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.StateRoot")
	}
	payloadReceiptsRoot, err := bytesutil.DecodeHexWithLength(e.ReceiptsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.ReceiptsRoot")
	}
	payloadLogsBloom, err := bytesutil.DecodeHexWithLength(e.LogsBloom, fieldparams.LogsBloomLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.LogsBloom")
	}
	payloadPrevRandao, err := bytesutil.DecodeHexWithLength(e.PrevRandao, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.PrevRandao")
	}
	payloadBlockNumber, err := strconv.ParseUint(e.BlockNumber, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.BlockNumber")
	}
	payloadGasLimit, err := strconv.ParseUint(e.GasLimit, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.GasLimit")
	}
	payloadGasUsed, err := strconv.ParseUint(e.GasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.GasUsed")
	}
	payloadTimestamp, err := strconv.ParseUint(e.Timestamp, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.Timestamp")
	}
	payloadExtraData, err := bytesutil.DecodeHexWithMaxLength(e.ExtraData, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.ExtraData")
	}
	payloadBaseFeePerGas, err := bytesutil.Uint256ToSSZBytes(e.BaseFeePerGas)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.BaseFeePerGas")
	}
	payloadBlockHash, err := bytesutil.DecodeHexWithLength(e.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.BlockHash")
	}
	err = slice.VerifyMaxLength(e.Transactions, fieldparams.MaxTxsPerPayloadLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.Transactions")
	}
	payloadTxs := make([][]byte, len(e.Transactions))
	for i, tx := range e.Transactions {
		payloadTxs[i], err = bytesutil.DecodeHexWithMaxLength(tx, fieldparams.MaxBytesPerTxLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("SilaPayload.Transactions[%d]", i))
		}
	}
	return &silaenginev1.SilaPayload{
		ParentHash:    payloadParentHash,
		FeeRecipient:  payloadFeeRecipient,
		StateRoot:     payloadStateRoot,
		ReceiptsRoot:  payloadReceiptsRoot,
		LogsBloom:     payloadLogsBloom,
		PrevRandao:    payloadPrevRandao,
		BlockNumber:   payloadBlockNumber,
		GasLimit:      payloadGasLimit,
		GasUsed:       payloadGasUsed,
		Timestamp:     payloadTimestamp,
		ExtraData:     payloadExtraData,
		BaseFeePerGas: payloadBaseFeePerGas,
		BlockHash:     payloadBlockHash,
		Transactions:  payloadTxs,
	}, nil
}

func (r *SilaPayload) PayloadProto() (proto.Message, error) {
	if r == nil {
		return nil, errors.Wrap(consensusblocks.ErrNilObject, "nil sila payload")
	}
	return r.ToConsensus()
}

func SilaPayloadHeaderFromConsensus(payload *silaenginev1.SilaPayloadHeader) (*SilaPayloadHeader, error) {
	baseFeePerGas, err := sszBytesToUint256String(payload.BaseFeePerGas)
	if err != nil {
		return nil, err
	}

	return &SilaPayloadHeader{
		ParentHash:       hexutil.Encode(payload.ParentHash),
		FeeRecipient:     hexutil.Encode(payload.FeeRecipient),
		StateRoot:        hexutil.Encode(payload.StateRoot),
		ReceiptsRoot:     hexutil.Encode(payload.ReceiptsRoot),
		LogsBloom:        hexutil.Encode(payload.LogsBloom),
		PrevRandao:       hexutil.Encode(payload.PrevRandao),
		BlockNumber:      fmt.Sprintf("%d", payload.BlockNumber),
		GasLimit:         fmt.Sprintf("%d", payload.GasLimit),
		GasUsed:          fmt.Sprintf("%d", payload.GasUsed),
		Timestamp:        fmt.Sprintf("%d", payload.Timestamp),
		ExtraData:        hexutil.Encode(payload.ExtraData),
		BaseFeePerGas:    baseFeePerGas,
		BlockHash:        hexutil.Encode(payload.BlockHash),
		TransactionsRoot: hexutil.Encode(payload.TransactionsRoot),
	}, nil
}

func (e *SilaPayloadHeader) ToConsensus() (*silaenginev1.SilaPayloadHeader, error) {
	if e == nil {
		return nil, server.NewDecodeError(errNilValue, "SilaPayloadHeader")
	}
	payloadParentHash, err := bytesutil.DecodeHexWithLength(e.ParentHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.ParentHash")
	}
	payloadFeeRecipient, err := bytesutil.DecodeHexWithLength(e.FeeRecipient, fieldparams.FeeRecipientLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.FeeRecipient")
	}
	payloadStateRoot, err := bytesutil.DecodeHexWithLength(e.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.StateRoot")
	}
	payloadReceiptsRoot, err := bytesutil.DecodeHexWithLength(e.ReceiptsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.ReceiptsRoot")
	}
	payloadLogsBloom, err := bytesutil.DecodeHexWithLength(e.LogsBloom, fieldparams.LogsBloomLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.LogsBloom")
	}
	payloadPrevRandao, err := bytesutil.DecodeHexWithLength(e.PrevRandao, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.PrevRandao")
	}
	payloadBlockNumber, err := strconv.ParseUint(e.BlockNumber, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.BlockNumber")
	}
	payloadGasLimit, err := strconv.ParseUint(e.GasLimit, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.GasLimit")
	}
	payloadGasUsed, err := strconv.ParseUint(e.GasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.GasUsed")
	}
	payloadTimestamp, err := strconv.ParseUint(e.Timestamp, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.Timestamp")
	}
	payloadExtraData, err := bytesutil.DecodeHexWithMaxLength(e.ExtraData, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.ExtraData")
	}
	payloadBaseFeePerGas, err := bytesutil.Uint256ToSSZBytes(e.BaseFeePerGas)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.BaseFeePerGas")
	}
	payloadBlockHash, err := bytesutil.DecodeHexWithLength(e.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.BlockHash")
	}
	payloadTxsRoot, err := bytesutil.DecodeHexWithLength(e.TransactionsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.TransactionsRoot")
	}

	return &silaenginev1.SilaPayloadHeader{
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
	}, nil
}

// ----------------------------------------------------------------------------
// Capella
// ----------------------------------------------------------------------------

func SilaPayloadCapellaFromConsensus(payload *silaenginev1.SilaPayloadCapella) (*SilaPayloadCapella, error) {
	baseFeePerGas, err := sszBytesToUint256String(payload.BaseFeePerGas)
	if err != nil {
		return nil, err
	}
	transactions := make([]string, len(payload.Transactions))
	for i, tx := range payload.Transactions {
		transactions[i] = hexutil.Encode(tx)
	}

	return &SilaPayloadCapella{
		ParentHash:    hexutil.Encode(payload.ParentHash),
		FeeRecipient:  hexutil.Encode(payload.FeeRecipient),
		StateRoot:     hexutil.Encode(payload.StateRoot),
		ReceiptsRoot:  hexutil.Encode(payload.ReceiptsRoot),
		LogsBloom:     hexutil.Encode(payload.LogsBloom),
		PrevRandao:    hexutil.Encode(payload.PrevRandao),
		BlockNumber:   fmt.Sprintf("%d", payload.BlockNumber),
		GasLimit:      fmt.Sprintf("%d", payload.GasLimit),
		GasUsed:       fmt.Sprintf("%d", payload.GasUsed),
		Timestamp:     fmt.Sprintf("%d", payload.Timestamp),
		ExtraData:     hexutil.Encode(payload.ExtraData),
		BaseFeePerGas: baseFeePerGas,
		BlockHash:     hexutil.Encode(payload.BlockHash),
		Transactions:  transactions,
		Withdrawals:   WithdrawalsFromConsensus(payload.Withdrawals),
	}, nil
}

func (e *SilaPayloadCapella) ToConsensus() (*silaenginev1.SilaPayloadCapella, error) {
	if e == nil {
		return nil, server.NewDecodeError(errNilValue, "SilaPayload")
	}
	payloadParentHash, err := bytesutil.DecodeHexWithLength(e.ParentHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.ParentHash")
	}
	payloadFeeRecipient, err := bytesutil.DecodeHexWithLength(e.FeeRecipient, fieldparams.FeeRecipientLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.FeeRecipient")
	}
	payloadStateRoot, err := bytesutil.DecodeHexWithLength(e.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.StateRoot")
	}
	payloadReceiptsRoot, err := bytesutil.DecodeHexWithLength(e.ReceiptsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.ReceiptsRoot")
	}
	payloadLogsBloom, err := bytesutil.DecodeHexWithLength(e.LogsBloom, fieldparams.LogsBloomLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.LogsBloom")
	}
	payloadPrevRandao, err := bytesutil.DecodeHexWithLength(e.PrevRandao, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.PrevRandao")
	}
	payloadBlockNumber, err := strconv.ParseUint(e.BlockNumber, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.BlockNumber")
	}
	payloadGasLimit, err := strconv.ParseUint(e.GasLimit, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.GasLimit")
	}
	payloadGasUsed, err := strconv.ParseUint(e.GasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.GasUsed")
	}
	payloadTimestamp, err := strconv.ParseUint(e.Timestamp, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.Timestamp")
	}
	payloadExtraData, err := bytesutil.DecodeHexWithMaxLength(e.ExtraData, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.ExtraData")
	}
	payloadBaseFeePerGas, err := bytesutil.Uint256ToSSZBytes(e.BaseFeePerGas)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.BaseFeePerGas")
	}
	payloadBlockHash, err := bytesutil.DecodeHexWithLength(e.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.BlockHash")
	}
	err = slice.VerifyMaxLength(e.Transactions, fieldparams.MaxTxsPerPayloadLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.Transactions")
	}
	payloadTxs := make([][]byte, len(e.Transactions))
	for i, tx := range e.Transactions {
		payloadTxs[i], err = bytesutil.DecodeHexWithMaxLength(tx, fieldparams.MaxBytesPerTxLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("SilaPayload.Transactions[%d]", i))
		}
	}
	err = slice.VerifyMaxLength(e.Withdrawals, fieldparams.MaxWithdrawalsPerPayload)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.Withdrawals")
	}
	withdrawals := make([]*silaenginev1.Withdrawal, len(e.Withdrawals))
	for i, w := range e.Withdrawals {
		withdrawalIndex, err := strconv.ParseUint(w.WithdrawalIndex, 10, 64)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("SilaPayload.Withdrawals[%d].WithdrawalIndex", i))
		}
		validatorIndex, err := strconv.ParseUint(w.ValidatorIndex, 10, 64)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("SilaPayload.Withdrawals[%d].ValidatorIndex", i))
		}
		address, err := bytesutil.DecodeHexWithLength(w.ExecutionAddress, common.AddressLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("SilaPayload.Withdrawals[%d].ExecutionAddress", i))
		}
		amount, err := strconv.ParseUint(w.Amount, 10, 64)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("SilaPayload.Withdrawals[%d].Amount", i))
		}
		withdrawals[i] = &silaenginev1.Withdrawal{
			Index:          withdrawalIndex,
			ValidatorIndex: primitives.ValidatorIndex(validatorIndex),
			Address:        address,
			Amount:         amount,
		}
	}
	return &silaenginev1.SilaPayloadCapella{
		ParentHash:    payloadParentHash,
		FeeRecipient:  payloadFeeRecipient,
		StateRoot:     payloadStateRoot,
		ReceiptsRoot:  payloadReceiptsRoot,
		LogsBloom:     payloadLogsBloom,
		PrevRandao:    payloadPrevRandao,
		BlockNumber:   payloadBlockNumber,
		GasLimit:      payloadGasLimit,
		GasUsed:       payloadGasUsed,
		Timestamp:     payloadTimestamp,
		ExtraData:     payloadExtraData,
		BaseFeePerGas: payloadBaseFeePerGas,
		BlockHash:     payloadBlockHash,
		Transactions:  payloadTxs,
		Withdrawals:   withdrawals,
	}, nil
}

func (p *SilaPayloadCapella) PayloadProto() (proto.Message, error) {
	if p == nil {
		return nil, errors.Wrap(consensusblocks.ErrNilObject, "nil capella sila payload")
	}
	return p.ToConsensus()
}

func SilaPayloadHeaderCapellaFromConsensus(payload *silaenginev1.SilaPayloadHeaderCapella) (*SilaPayloadHeaderCapella, error) {
	baseFeePerGas, err := sszBytesToUint256String(payload.BaseFeePerGas)
	if err != nil {
		return nil, err
	}

	return &SilaPayloadHeaderCapella{
		ParentHash:       hexutil.Encode(payload.ParentHash),
		FeeRecipient:     hexutil.Encode(payload.FeeRecipient),
		StateRoot:        hexutil.Encode(payload.StateRoot),
		ReceiptsRoot:     hexutil.Encode(payload.ReceiptsRoot),
		LogsBloom:        hexutil.Encode(payload.LogsBloom),
		PrevRandao:       hexutil.Encode(payload.PrevRandao),
		BlockNumber:      fmt.Sprintf("%d", payload.BlockNumber),
		GasLimit:         fmt.Sprintf("%d", payload.GasLimit),
		GasUsed:          fmt.Sprintf("%d", payload.GasUsed),
		Timestamp:        fmt.Sprintf("%d", payload.Timestamp),
		ExtraData:        hexutil.Encode(payload.ExtraData),
		BaseFeePerGas:    baseFeePerGas,
		BlockHash:        hexutil.Encode(payload.BlockHash),
		TransactionsRoot: hexutil.Encode(payload.TransactionsRoot),
		WithdrawalsRoot:  hexutil.Encode(payload.WithdrawalsRoot),
	}, nil
}

func (e *SilaPayloadHeaderCapella) ToConsensus() (*silaenginev1.SilaPayloadHeaderCapella, error) {
	if e == nil {
		return nil, server.NewDecodeError(errNilValue, "SilaPayloadHeader")
	}
	payloadParentHash, err := bytesutil.DecodeHexWithLength(e.ParentHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.ParentHash")
	}
	payloadFeeRecipient, err := bytesutil.DecodeHexWithLength(e.FeeRecipient, fieldparams.FeeRecipientLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.FeeRecipient")
	}
	payloadStateRoot, err := bytesutil.DecodeHexWithLength(e.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.StateRoot")
	}
	payloadReceiptsRoot, err := bytesutil.DecodeHexWithLength(e.ReceiptsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.ReceiptsRoot")
	}
	payloadLogsBloom, err := bytesutil.DecodeHexWithLength(e.LogsBloom, fieldparams.LogsBloomLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.LogsBloom")
	}
	payloadPrevRandao, err := bytesutil.DecodeHexWithLength(e.PrevRandao, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.PrevRandao")
	}
	payloadBlockNumber, err := strconv.ParseUint(e.BlockNumber, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.BlockNumber")
	}
	payloadGasLimit, err := strconv.ParseUint(e.GasLimit, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.GasLimit")
	}
	payloadGasUsed, err := strconv.ParseUint(e.GasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.GasUsed")
	}
	payloadTimestamp, err := strconv.ParseUint(e.Timestamp, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.Timestamp")
	}
	payloadExtraData, err := bytesutil.DecodeHexWithMaxLength(e.ExtraData, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.ExtraData")
	}
	payloadBaseFeePerGas, err := bytesutil.Uint256ToSSZBytes(e.BaseFeePerGas)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.BaseFeePerGas")
	}
	payloadBlockHash, err := bytesutil.DecodeHexWithMaxLength(e.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.BlockHash")
	}
	payloadTxsRoot, err := bytesutil.DecodeHexWithMaxLength(e.TransactionsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.TransactionsRoot")
	}
	payloadWithdrawalsRoot, err := bytesutil.DecodeHexWithMaxLength(e.WithdrawalsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.WithdrawalsRoot")
	}
	return &silaenginev1.SilaPayloadHeaderCapella{
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
	}, nil
}

// ----------------------------------------------------------------------------
// Deneb
// ----------------------------------------------------------------------------

func SilaPayloadDenebFromConsensus(payload *silaenginev1.SilaPayloadDeneb) (*SilaPayloadDeneb, error) {
	baseFeePerGas, err := sszBytesToUint256String(payload.BaseFeePerGas)
	if err != nil {
		return nil, err
	}
	transactions := make([]string, len(payload.Transactions))
	for i, tx := range payload.Transactions {
		transactions[i] = hexutil.Encode(tx)
	}

	return &SilaPayloadDeneb{
		ParentHash:    hexutil.Encode(payload.ParentHash),
		FeeRecipient:  hexutil.Encode(payload.FeeRecipient),
		StateRoot:     hexutil.Encode(payload.StateRoot),
		ReceiptsRoot:  hexutil.Encode(payload.ReceiptsRoot),
		LogsBloom:     hexutil.Encode(payload.LogsBloom),
		PrevRandao:    hexutil.Encode(payload.PrevRandao),
		BlockNumber:   fmt.Sprintf("%d", payload.BlockNumber),
		GasLimit:      fmt.Sprintf("%d", payload.GasLimit),
		GasUsed:       fmt.Sprintf("%d", payload.GasUsed),
		Timestamp:     fmt.Sprintf("%d", payload.Timestamp),
		ExtraData:     hexutil.Encode(payload.ExtraData),
		BaseFeePerGas: baseFeePerGas,
		BlockHash:     hexutil.Encode(payload.BlockHash),
		Transactions:  transactions,
		Withdrawals:   WithdrawalsFromConsensus(payload.Withdrawals),
		BlobGasUsed:   fmt.Sprintf("%d", payload.BlobGasUsed),
		ExcessBlobGas: fmt.Sprintf("%d", payload.ExcessBlobGas),
	}, nil
}

func (e *SilaPayloadDeneb) ToConsensus() (*silaenginev1.SilaPayloadDeneb, error) {
	if e == nil {
		return nil, server.NewDecodeError(errNilValue, "SilaPayload")
	}
	payloadParentHash, err := bytesutil.DecodeHexWithLength(e.ParentHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.ParentHash")
	}
	payloadFeeRecipient, err := bytesutil.DecodeHexWithLength(e.FeeRecipient, fieldparams.FeeRecipientLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.FeeRecipient")
	}
	payloadStateRoot, err := bytesutil.DecodeHexWithLength(e.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.StateRoot")
	}
	payloadReceiptsRoot, err := bytesutil.DecodeHexWithLength(e.ReceiptsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.ReceiptsRoot")
	}
	payloadLogsBloom, err := bytesutil.DecodeHexWithLength(e.LogsBloom, fieldparams.LogsBloomLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.LogsBloom")
	}
	payloadPrevRandao, err := bytesutil.DecodeHexWithLength(e.PrevRandao, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.PrevRandao")
	}
	payloadBlockNumber, err := strconv.ParseUint(e.BlockNumber, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.BlockNumber")
	}
	payloadGasLimit, err := strconv.ParseUint(e.GasLimit, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.GasLimit")
	}
	payloadGasUsed, err := strconv.ParseUint(e.GasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.GasUsed")
	}
	payloadTimestamp, err := strconv.ParseUint(e.Timestamp, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.Timestamp")
	}
	payloadExtraData, err := bytesutil.DecodeHexWithMaxLength(e.ExtraData, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.ExtraData")
	}
	payloadBaseFeePerGas, err := bytesutil.Uint256ToSSZBytes(e.BaseFeePerGas)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.BaseFeePerGas")
	}
	payloadBlockHash, err := bytesutil.DecodeHexWithLength(e.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.BlockHash")
	}
	err = slice.VerifyMaxLength(e.Transactions, fieldparams.MaxTxsPerPayloadLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.Transactions")
	}
	txs := make([][]byte, len(e.Transactions))
	for i, tx := range e.Transactions {
		txs[i], err = bytesutil.DecodeHexWithMaxLength(tx, fieldparams.MaxBytesPerTxLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("SilaPayload.Transactions[%d]", i))
		}
	}
	err = slice.VerifyMaxLength(e.Withdrawals, fieldparams.MaxWithdrawalsPerPayload)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.Withdrawals")
	}
	withdrawals := make([]*silaenginev1.Withdrawal, len(e.Withdrawals))
	for i, w := range e.Withdrawals {
		withdrawalIndex, err := strconv.ParseUint(w.WithdrawalIndex, 10, 64)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("SilaPayload.Withdrawals[%d].WithdrawalIndex", i))
		}
		validatorIndex, err := strconv.ParseUint(w.ValidatorIndex, 10, 64)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("SilaPayload.Withdrawals[%d].ValidatorIndex", i))
		}
		address, err := bytesutil.DecodeHexWithLength(w.ExecutionAddress, common.AddressLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("SilaPayload.Withdrawals[%d].ExecutionAddress", i))
		}
		amount, err := strconv.ParseUint(w.Amount, 10, 64)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("SilaPayload.Withdrawals[%d].Amount", i))
		}
		withdrawals[i] = &silaenginev1.Withdrawal{
			Index:          withdrawalIndex,
			ValidatorIndex: primitives.ValidatorIndex(validatorIndex),
			Address:        address,
			Amount:         amount,
		}
	}

	payloadBlobGasUsed, err := strconv.ParseUint(e.BlobGasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.BlobGasUsed")
	}
	payloadExcessBlobGas, err := strconv.ParseUint(e.ExcessBlobGas, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.ExcessBlobGas")
	}
	return &silaenginev1.SilaPayloadDeneb{
		ParentHash:    payloadParentHash,
		FeeRecipient:  payloadFeeRecipient,
		StateRoot:     payloadStateRoot,
		ReceiptsRoot:  payloadReceiptsRoot,
		LogsBloom:     payloadLogsBloom,
		PrevRandao:    payloadPrevRandao,
		BlockNumber:   payloadBlockNumber,
		GasLimit:      payloadGasLimit,
		GasUsed:       payloadGasUsed,
		Timestamp:     payloadTimestamp,
		ExtraData:     payloadExtraData,
		BaseFeePerGas: payloadBaseFeePerGas,
		BlockHash:     payloadBlockHash,
		Transactions:  txs,
		Withdrawals:   withdrawals,
		BlobGasUsed:   payloadBlobGasUsed,
		ExcessBlobGas: payloadExcessBlobGas,
	}, nil
}

func SilaPayloadHeaderDenebFromConsensus(payload *silaenginev1.SilaPayloadHeaderDeneb) (*SilaPayloadHeaderDeneb, error) {
	baseFeePerGas, err := sszBytesToUint256String(payload.BaseFeePerGas)
	if err != nil {
		return nil, err
	}

	return &SilaPayloadHeaderDeneb{
		ParentHash:       hexutil.Encode(payload.ParentHash),
		FeeRecipient:     hexutil.Encode(payload.FeeRecipient),
		StateRoot:        hexutil.Encode(payload.StateRoot),
		ReceiptsRoot:     hexutil.Encode(payload.ReceiptsRoot),
		LogsBloom:        hexutil.Encode(payload.LogsBloom),
		PrevRandao:       hexutil.Encode(payload.PrevRandao),
		BlockNumber:      fmt.Sprintf("%d", payload.BlockNumber),
		GasLimit:         fmt.Sprintf("%d", payload.GasLimit),
		GasUsed:          fmt.Sprintf("%d", payload.GasUsed),
		Timestamp:        fmt.Sprintf("%d", payload.Timestamp),
		ExtraData:        hexutil.Encode(payload.ExtraData),
		BaseFeePerGas:    baseFeePerGas,
		BlockHash:        hexutil.Encode(payload.BlockHash),
		TransactionsRoot: hexutil.Encode(payload.TransactionsRoot),
		WithdrawalsRoot:  hexutil.Encode(payload.WithdrawalsRoot),
		BlobGasUsed:      fmt.Sprintf("%d", payload.BlobGasUsed),
		ExcessBlobGas:    fmt.Sprintf("%d", payload.ExcessBlobGas),
	}, nil
}

func (e *SilaPayloadHeaderDeneb) ToConsensus() (*silaenginev1.SilaPayloadHeaderDeneb, error) {
	if e == nil {
		return nil, server.NewDecodeError(errNilValue, "SilaPayloadHeader")
	}
	payloadParentHash, err := bytesutil.DecodeHexWithLength(e.ParentHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.ParentHash")
	}
	payloadFeeRecipient, err := bytesutil.DecodeHexWithLength(e.FeeRecipient, fieldparams.FeeRecipientLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.FeeRecipient")
	}
	payloadStateRoot, err := bytesutil.DecodeHexWithLength(e.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.StateRoot")
	}
	payloadReceiptsRoot, err := bytesutil.DecodeHexWithLength(e.ReceiptsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.ReceiptsRoot")
	}
	payloadLogsBloom, err := bytesutil.DecodeHexWithLength(e.LogsBloom, fieldparams.LogsBloomLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.LogsBloom")
	}
	payloadPrevRandao, err := bytesutil.DecodeHexWithLength(e.PrevRandao, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.PrevRandao")
	}
	payloadBlockNumber, err := strconv.ParseUint(e.BlockNumber, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.BlockNumber")
	}
	payloadGasLimit, err := strconv.ParseUint(e.GasLimit, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.GasLimit")
	}
	payloadGasUsed, err := strconv.ParseUint(e.GasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.GasUsed")
	}
	payloadTimestamp, err := strconv.ParseUint(e.Timestamp, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.Timestamp")
	}
	payloadExtraData, err := bytesutil.DecodeHexWithMaxLength(e.ExtraData, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.ExtraData")
	}
	payloadBaseFeePerGas, err := bytesutil.Uint256ToSSZBytes(e.BaseFeePerGas)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.BaseFeePerGas")
	}
	payloadBlockHash, err := bytesutil.DecodeHexWithLength(e.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.BlockHash")
	}
	payloadTxsRoot, err := bytesutil.DecodeHexWithLength(e.TransactionsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.TransactionsRoot")
	}
	payloadWithdrawalsRoot, err := bytesutil.DecodeHexWithLength(e.WithdrawalsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayloadHeader.WithdrawalsRoot")
	}
	payloadBlobGasUsed, err := strconv.ParseUint(e.BlobGasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.BlobGasUsed")
	}
	payloadExcessBlobGas, err := strconv.ParseUint(e.ExcessBlobGas, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SilaPayload.ExcessBlobGas")
	}
	return &silaenginev1.SilaPayloadHeaderDeneb{
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
	}, nil
}

// ----------------------------------------------------------------------------
// Electra
// ----------------------------------------------------------------------------

var (
	SilaPayloadElectraFromConsensus       = SilaPayloadDenebFromConsensus
	SilaPayloadHeaderElectraFromConsensus = SilaPayloadHeaderDenebFromConsensus
)

func WithdrawalRequestsFromConsensus(ws []*silaenginev1.WithdrawalRequest) []*WithdrawalRequest {
	result := make([]*WithdrawalRequest, len(ws))
	for i, w := range ws {
		result[i] = WithdrawalRequestFromConsensus(w)
	}
	return result
}

func WithdrawalRequestFromConsensus(w *silaenginev1.WithdrawalRequest) *WithdrawalRequest {
	return &WithdrawalRequest{
		SourceAddress:   hexutil.Encode(w.SourceAddress),
		ValidatorPubkey: hexutil.Encode(w.ValidatorPubkey),
		Amount:          fmt.Sprintf("%d", w.Amount),
	}
}

func (w *WithdrawalRequest) ToConsensus() (*silaenginev1.WithdrawalRequest, error) {
	if w == nil {
		return nil, server.NewDecodeError(errNilValue, "WithdrawalRequest")
	}
	src, err := bytesutil.DecodeHexWithLength(w.SourceAddress, common.AddressLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SourceAddress")
	}
	pubkey, err := bytesutil.DecodeHexWithLength(w.ValidatorPubkey, fieldparams.BLSPubkeyLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ValidatorPubkey")
	}
	amount, err := strconv.ParseUint(w.Amount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Amount")
	}
	return &silaenginev1.WithdrawalRequest{
		SourceAddress:   src,
		ValidatorPubkey: pubkey,
		Amount:          amount,
	}, nil
}

func ConsolidationRequestsFromConsensus(cs []*silaenginev1.ConsolidationRequest) []*ConsolidationRequest {
	result := make([]*ConsolidationRequest, len(cs))
	for i, c := range cs {
		result[i] = ConsolidationRequestFromConsensus(c)
	}
	return result
}

func ConsolidationRequestFromConsensus(c *silaenginev1.ConsolidationRequest) *ConsolidationRequest {
	return &ConsolidationRequest{
		SourceAddress: hexutil.Encode(c.SourceAddress),
		SourcePubkey:  hexutil.Encode(c.SourcePubkey),
		TargetPubkey:  hexutil.Encode(c.TargetPubkey),
	}
}

func (c *ConsolidationRequest) ToConsensus() (*silaenginev1.ConsolidationRequest, error) {
	if c == nil {
		return nil, server.NewDecodeError(errNilValue, "ConsolidationRequest")
	}
	srcAddress, err := bytesutil.DecodeHexWithLength(c.SourceAddress, common.AddressLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SourceAddress")
	}
	srcPubkey, err := bytesutil.DecodeHexWithLength(c.SourcePubkey, fieldparams.BLSPubkeyLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "SourcePubkey")
	}
	targetPubkey, err := bytesutil.DecodeHexWithLength(c.TargetPubkey, fieldparams.BLSPubkeyLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "TargetPubkey")
	}
	return &silaenginev1.ConsolidationRequest{
		SourceAddress: srcAddress,
		SourcePubkey:  srcPubkey,
		TargetPubkey:  targetPubkey,
	}, nil
}

func DepositRequestsFromConsensus(ds []*silaenginev1.DepositRequest) []*DepositRequest {
	result := make([]*DepositRequest, len(ds))
	for i, d := range ds {
		result[i] = DepositRequestFromConsensus(d)
	}
	return result
}

func DepositRequestFromConsensus(d *silaenginev1.DepositRequest) *DepositRequest {
	return &DepositRequest{
		Pubkey:                hexutil.Encode(d.Pubkey),
		WithdrawalCredentials: hexutil.Encode(d.WithdrawalCredentials),
		Amount:                fmt.Sprintf("%d", d.Amount),
		Signature:             hexutil.Encode(d.Signature),
		Index:                 fmt.Sprintf("%d", d.Index),
	}
}

func (d *DepositRequest) ToConsensus() (*silaenginev1.DepositRequest, error) {
	if d == nil {
		return nil, server.NewDecodeError(errNilValue, "DepositRequest")
	}
	pubkey, err := bytesutil.DecodeHexWithLength(d.Pubkey, fieldparams.BLSPubkeyLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Pubkey")
	}
	withdrawalCredentials, err := bytesutil.DecodeHexWithLength(d.WithdrawalCredentials, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "WithdrawalCredentials")
	}
	amount, err := strconv.ParseUint(d.Amount, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Amount")
	}
	sig, err := bytesutil.DecodeHexWithLength(d.Signature, fieldparams.BLSSignatureLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Signature")
	}
	index, err := strconv.ParseUint(d.Index, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Index")
	}
	return &silaenginev1.DepositRequest{
		Pubkey:                pubkey,
		WithdrawalCredentials: withdrawalCredentials,
		Amount:                amount,
		Signature:             sig,
		Index:                 index,
	}, nil
}

func ExecutionRequestsFromConsensus(er *silaenginev1.ExecutionRequests) *ExecutionRequests {
	return &ExecutionRequests{
		Deposits:       DepositRequestsFromConsensus(er.Deposits),
		Withdrawals:    WithdrawalRequestsFromConsensus(er.Withdrawals),
		Consolidations: ConsolidationRequestsFromConsensus(er.Consolidations),
	}
}

func (e *ExecutionRequests) ToConsensus() (*silaenginev1.ExecutionRequests, error) {
	if e == nil {
		return nil, server.NewDecodeError(errNilValue, "ExecutionRequests")
	}
	var err error
	if err = slice.VerifyMaxLength(e.Deposits, params.BeaconConfig().MaxDepositRequestsPerPayload); err != nil {
		return nil, err
	}
	depositRequests := make([]*silaenginev1.DepositRequest, len(e.Deposits))
	for i, d := range e.Deposits {
		depositRequests[i], err = d.ToConsensus()
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("ExecutionRequests.Deposits[%d]", i))
		}
	}

	if err = slice.VerifyMaxLength(e.Withdrawals, params.BeaconConfig().MaxWithdrawalRequestsPerPayload); err != nil {
		return nil, err
	}
	withdrawalRequests := make([]*silaenginev1.WithdrawalRequest, len(e.Withdrawals))
	for i, w := range e.Withdrawals {
		withdrawalRequests[i], err = w.ToConsensus()
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("ExecutionRequests.Withdrawals[%d]", i))
		}
	}

	if err = slice.VerifyMaxLength(e.Consolidations, params.BeaconConfig().MaxConsolidationsRequestsPerPayload); err != nil {
		return nil, err
	}
	consolidationRequests := make([]*silaenginev1.ConsolidationRequest, len(e.Consolidations))
	for i, c := range e.Consolidations {
		consolidationRequests[i], err = c.ToConsensus()
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("ExecutionRequests.Consolidations[%d]", i))
		}
	}
	return &silaenginev1.ExecutionRequests{
		Deposits:       depositRequests,
		Withdrawals:    withdrawalRequests,
		Consolidations: consolidationRequests,
	}, nil
}

// ----------------------------------------------------------------------------
// Fulu
// ----------------------------------------------------------------------------

var (
	SilaPayloadFuluFromConsensus       = SilaPayloadDenebFromConsensus
	SilaPayloadHeaderFuluFromConsensus = SilaPayloadHeaderDenebFromConsensus
	BeaconBlockFuluFromConsensus            = BeaconBlockElectraFromConsensus
)

func SilaPayloadGloasFromConsensus(payload *silaenginev1.SilaPayloadGloas) (*SilaPayloadGloas, error) {
	baseFeePerGas, err := sszBytesToUint256String(payload.BaseFeePerGas)
	if err != nil {
		return nil, err
	}
	transactions := make([]string, len(payload.Transactions))
	for i, tx := range payload.Transactions {
		transactions[i] = hexutil.Encode(tx)
	}

	return &SilaPayloadGloas{
		ParentHash:      hexutil.Encode(payload.ParentHash),
		FeeRecipient:    hexutil.Encode(payload.FeeRecipient),
		StateRoot:       hexutil.Encode(payload.StateRoot),
		ReceiptsRoot:    hexutil.Encode(payload.ReceiptsRoot),
		LogsBloom:       hexutil.Encode(payload.LogsBloom),
		PrevRandao:      hexutil.Encode(payload.PrevRandao),
		BlockNumber:     fmt.Sprintf("%d", payload.BlockNumber),
		GasLimit:        fmt.Sprintf("%d", payload.GasLimit),
		GasUsed:         fmt.Sprintf("%d", payload.GasUsed),
		Timestamp:       fmt.Sprintf("%d", payload.Timestamp),
		ExtraData:       hexutil.Encode(payload.ExtraData),
		BaseFeePerGas:   baseFeePerGas,
		BlockHash:       hexutil.Encode(payload.BlockHash),
		Transactions:    transactions,
		Withdrawals:     WithdrawalsFromConsensus(payload.Withdrawals),
		BlobGasUsed:     fmt.Sprintf("%d", payload.BlobGasUsed),
		ExcessBlobGas:   fmt.Sprintf("%d", payload.ExcessBlobGas),
		BlockAccessList: hexutil.Encode(payload.BlockAccessList),
		SlotNumber:      fmt.Sprintf("%d", payload.SlotNumber),
	}, nil
}

func (e *SilaPayloadGloas) ToConsensus() (*silaenginev1.SilaPayloadGloas, error) {
	if e == nil {
		return nil, server.NewDecodeError(errNilValue, "SilaPayloadGloas")
	}
	payloadParentHash, err := bytesutil.DecodeHexWithLength(e.ParentHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ParentHash")
	}
	payloadFeeRecipient, err := bytesutil.DecodeHexWithLength(e.FeeRecipient, fieldparams.FeeRecipientLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "FeeRecipient")
	}
	payloadStateRoot, err := bytesutil.DecodeHexWithLength(e.StateRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "StateRoot")
	}
	payloadReceiptsRoot, err := bytesutil.DecodeHexWithLength(e.ReceiptsRoot, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ReceiptsRoot")
	}
	payloadLogsBloom, err := bytesutil.DecodeHexWithLength(e.LogsBloom, fieldparams.LogsBloomLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "LogsBloom")
	}
	payloadPrevRandao, err := bytesutil.DecodeHexWithLength(e.PrevRandao, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "PrevRandao")
	}
	payloadBlockNumber, err := strconv.ParseUint(e.BlockNumber, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "BlockNumber")
	}
	payloadGasLimit, err := strconv.ParseUint(e.GasLimit, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "GasLimit")
	}
	payloadGasUsed, err := strconv.ParseUint(e.GasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "GasUsed")
	}
	payloadTimestamp, err := strconv.ParseUint(e.Timestamp, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "Timestamp")
	}
	payloadExtraData, err := bytesutil.DecodeHexWithMaxLength(e.ExtraData, fieldparams.RootLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "ExtraData")
	}
	payloadBaseFeePerGas, err := bytesutil.Uint256ToSSZBytes(e.BaseFeePerGas)
	if err != nil {
		return nil, server.NewDecodeError(err, "BaseFeePerGas")
	}
	payloadBlockHash, err := bytesutil.DecodeHexWithLength(e.BlockHash, common.HashLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "BlockHash")
	}
	err = slice.VerifyMaxLength(e.Transactions, fieldparams.MaxTxsPerPayloadLength)
	if err != nil {
		return nil, server.NewDecodeError(err, "Transactions")
	}
	txs := make([][]byte, len(e.Transactions))
	for i, tx := range e.Transactions {
		txs[i], err = bytesutil.DecodeHexWithMaxLength(tx, fieldparams.MaxBytesPerTxLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Transactions[%d]", i))
		}
	}
	err = slice.VerifyMaxLength(e.Withdrawals, fieldparams.MaxWithdrawalsPerPayload)
	if err != nil {
		return nil, server.NewDecodeError(err, "Withdrawals")
	}
	withdrawals := make([]*silaenginev1.Withdrawal, len(e.Withdrawals))
	for i, w := range e.Withdrawals {
		withdrawalIndex, err := strconv.ParseUint(w.WithdrawalIndex, 10, 64)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Withdrawals[%d].WithdrawalIndex", i))
		}
		validatorIndex, err := strconv.ParseUint(w.ValidatorIndex, 10, 64)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Withdrawals[%d].ValidatorIndex", i))
		}
		address, err := bytesutil.DecodeHexWithLength(w.ExecutionAddress, common.AddressLength)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Withdrawals[%d].ExecutionAddress", i))
		}
		amount, err := strconv.ParseUint(w.Amount, 10, 64)
		if err != nil {
			return nil, server.NewDecodeError(err, fmt.Sprintf("Withdrawals[%d].Amount", i))
		}
		withdrawals[i] = &silaenginev1.Withdrawal{
			Index:          withdrawalIndex,
			ValidatorIndex: primitives.ValidatorIndex(validatorIndex),
			Address:        address,
			Amount:         amount,
		}
	}
	payloadBlobGasUsed, err := strconv.ParseUint(e.BlobGasUsed, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "BlobGasUsed")
	}
	payloadExcessBlobGas, err := strconv.ParseUint(e.ExcessBlobGas, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "ExcessBlobGas")
	}
	var bal []byte
	if e.BlockAccessList != "" {
		bal, err = hexutil.Decode(e.BlockAccessList)
		if err != nil {
			return nil, server.NewDecodeError(err, "BlockAccessList")
		}
	}
	payloadSlotNumber, err := strconv.ParseUint(e.SlotNumber, 10, 64)
	if err != nil {
		return nil, server.NewDecodeError(err, "SlotNumber")
	}
	return &silaenginev1.SilaPayloadGloas{
		ParentHash:      payloadParentHash,
		FeeRecipient:    payloadFeeRecipient,
		StateRoot:       payloadStateRoot,
		ReceiptsRoot:    payloadReceiptsRoot,
		LogsBloom:       payloadLogsBloom,
		PrevRandao:      payloadPrevRandao,
		BlockNumber:     payloadBlockNumber,
		GasLimit:        payloadGasLimit,
		GasUsed:         payloadGasUsed,
		Timestamp:       payloadTimestamp,
		ExtraData:       payloadExtraData,
		BaseFeePerGas:   payloadBaseFeePerGas,
		BlockHash:       payloadBlockHash,
		Transactions:    txs,
		Withdrawals:     withdrawals,
		BlobGasUsed:     payloadBlobGasUsed,
		ExcessBlobGas:   payloadExcessBlobGas,
		BlockAccessList: bal,
		SlotNumber:      primitives.Slot(payloadSlotNumber),
	}, nil
}
