package blocks_test

import (
	"testing"

	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	consensus_types "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func TestWrapSilaPayload(t *testing.T) {
	data := &silaenginev1.SilaPayload{GasUsed: 54}
	wsb, err := blocks.WrappedSilaPayload(data)
	require.NoError(t, err)

	assert.DeepEqual(t, data, wsb.Proto())
}

func TestWrapSilaPayloadHeader(t *testing.T) {
	data := &silaenginev1.SilaPayloadHeader{GasUsed: 54}
	wsb, err := blocks.WrappedSilaPayloadHeader(data)
	require.NoError(t, err)

	assert.DeepEqual(t, data, wsb.Proto())
}

func TestWrapSilaPayload_IsNil(t *testing.T) {
	_, err := blocks.WrappedSilaPayload(nil)
	require.Equal(t, consensus_types.ErrNilObjectWrapped, err)

	data := &silaenginev1.SilaPayload{GasUsed: 54}
	wsb, err := blocks.WrappedSilaPayload(data)
	require.NoError(t, err)

	assert.Equal(t, false, wsb.IsNil())
}

func TestWrapSilaPayloadHeader_IsNil(t *testing.T) {
	_, err := blocks.WrappedSilaPayloadHeader(nil)
	require.Equal(t, consensus_types.ErrNilObjectWrapped, err)

	data := &silaenginev1.SilaPayloadHeader{GasUsed: 54}
	wsb, err := blocks.WrappedSilaPayloadHeader(data)
	require.NoError(t, err)

	assert.Equal(t, false, wsb.IsNil())
}

func TestWrapSilaPayload_SSZ(t *testing.T) {
	wsb := createWrappedPayload(t)
	rt, err := wsb.HashTreeRoot()
	assert.NoError(t, err)
	assert.NotEmpty(t, rt)

	var b []byte
	b, err = wsb.MarshalSSZTo(b)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(b))
	encoded, err := wsb.MarshalSSZ()
	require.NoError(t, err)
	assert.NotEqual(t, 0, wsb.SizeSSZ())
	assert.NoError(t, wsb.UnmarshalSSZ(encoded))
}

func TestWrapSilaPayloadHeader_SSZ(t *testing.T) {
	wsb := createWrappedPayloadHeader(t)
	rt, err := wsb.HashTreeRoot()
	assert.NoError(t, err)
	assert.NotEmpty(t, rt)

	var b []byte
	b, err = wsb.MarshalSSZTo(b)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(b))
	encoded, err := wsb.MarshalSSZ()
	require.NoError(t, err)
	assert.NotEqual(t, 0, wsb.SizeSSZ())
	assert.NoError(t, wsb.UnmarshalSSZ(encoded))
}

func TestWrapSilaPayloadCapella(t *testing.T) {
	data := &silaenginev1.SilaPayloadCapella{
		ParentHash:    []byte("parenthash"),
		FeeRecipient:  []byte("feerecipient"),
		StateRoot:     []byte("stateroot"),
		ReceiptsRoot:  []byte("receiptsroot"),
		LogsBloom:     []byte("logsbloom"),
		PrevRandao:    []byte("prevrandao"),
		BlockNumber:   11,
		GasLimit:      22,
		GasUsed:       33,
		Timestamp:     44,
		ExtraData:     []byte("extradata"),
		BaseFeePerGas: []byte("basefeepergas"),
		BlockHash:     []byte("blockhash"),
		Transactions:  [][]byte{[]byte("transaction")},
		Withdrawals: []*silaenginev1.Withdrawal{{
			Index:          55,
			ValidatorIndex: 66,
			Address:        []byte("executionaddress"),
			Amount:         77,
		}},
	}
	payload, err := blocks.WrappedSilaPayloadCapella(data)
	require.NoError(t, err)

	assert.DeepEqual(t, data, payload.Proto())
}

func TestWrapSilaPayloadHeaderCapella(t *testing.T) {
	data := &silaenginev1.SilaPayloadHeaderCapella{
		ParentHash:       []byte("parenthash"),
		FeeRecipient:     []byte("feerecipient"),
		StateRoot:        []byte("stateroot"),
		ReceiptsRoot:     []byte("receiptsroot"),
		LogsBloom:        []byte("logsbloom"),
		PrevRandao:       []byte("prevrandao"),
		BlockNumber:      11,
		GasLimit:         22,
		GasUsed:          33,
		Timestamp:        44,
		ExtraData:        []byte("extradata"),
		BaseFeePerGas:    []byte("basefeepergas"),
		BlockHash:        []byte("blockhash"),
		TransactionsRoot: []byte("transactionsroot"),
		WithdrawalsRoot:  []byte("withdrawalsroot"),
	}
	payload, err := blocks.WrappedSilaPayloadHeaderCapella(data)
	require.NoError(t, err)

	assert.DeepEqual(t, data, payload.Proto())

	txRoot, err := payload.TransactionsRoot()
	require.NoError(t, err)
	require.DeepEqual(t, txRoot, data.TransactionsRoot)

	wrRoot, err := payload.WithdrawalsRoot()
	require.NoError(t, err)
	require.DeepEqual(t, wrRoot, data.WithdrawalsRoot)
}

func TestWrapSilaPayloadCapella_IsNil(t *testing.T) {
	_, err := blocks.WrappedSilaPayloadCapella(nil)
	require.Equal(t, consensus_types.ErrNilObjectWrapped, err)

	data := &silaenginev1.SilaPayloadCapella{GasUsed: 54}
	payload, err := blocks.WrappedSilaPayloadCapella(data)
	require.NoError(t, err)

	assert.Equal(t, false, payload.IsNil())
}

func TestWrapSilaPayloadHeaderCapella_IsNil(t *testing.T) {
	_, err := blocks.WrappedSilaPayloadHeaderCapella(nil)
	require.Equal(t, consensus_types.ErrNilObjectWrapped, err)

	data := &silaenginev1.SilaPayloadHeaderCapella{GasUsed: 54}
	payload, err := blocks.WrappedSilaPayloadHeaderCapella(data)
	require.NoError(t, err)

	assert.Equal(t, false, payload.IsNil())
}

func TestWrapSilaPayloadCapella_SSZ(t *testing.T) {
	payload := createWrappedPayloadCapella(t)
	rt, err := payload.HashTreeRoot()
	assert.NoError(t, err)
	assert.NotEmpty(t, rt)

	var b []byte
	b, err = payload.MarshalSSZTo(b)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(b))
	encoded, err := payload.MarshalSSZ()
	require.NoError(t, err)
	assert.NotEqual(t, 0, payload.SizeSSZ())
	assert.NoError(t, payload.UnmarshalSSZ(encoded))
}

func TestWrapSilaPayloadHeaderCapella_SSZ(t *testing.T) {
	payload := createWrappedPayloadHeaderCapella(t)
	rt, err := payload.HashTreeRoot()
	assert.NoError(t, err)
	assert.NotEmpty(t, rt)

	var b []byte
	b, err = payload.MarshalSSZTo(b)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(b))
	encoded, err := payload.MarshalSSZ()
	require.NoError(t, err)
	assert.NotEqual(t, 0, payload.SizeSSZ())
	assert.NoError(t, payload.UnmarshalSSZ(encoded))
}

func TestWrapSilaPayloadDeneb(t *testing.T) {
	data := &silaenginev1.SilaPayloadDeneb{
		ParentHash:    []byte("parenthash"),
		FeeRecipient:  []byte("feerecipient"),
		StateRoot:     []byte("stateroot"),
		ReceiptsRoot:  []byte("receiptsroot"),
		LogsBloom:     []byte("logsbloom"),
		PrevRandao:    []byte("prevrandao"),
		BlockNumber:   11,
		GasLimit:      22,
		GasUsed:       33,
		Timestamp:     44,
		ExtraData:     []byte("extradata"),
		BaseFeePerGas: []byte("basefeepergas"),
		BlockHash:     []byte("blockhash"),
		Transactions:  [][]byte{[]byte("transaction")},
		Withdrawals: []*silaenginev1.Withdrawal{{
			Index:          55,
			ValidatorIndex: 66,
			Address:        []byte("executionaddress"),
			Amount:         77,
		}},
		BlobGasUsed:   88,
		ExcessBlobGas: 99,
	}
	payload, err := blocks.WrappedSilaPayloadDeneb(data)
	require.NoError(t, err)

	g, err := payload.BlobGasUsed()
	require.NoError(t, err)
	require.DeepEqual(t, uint64(88), g)

	g, err = payload.ExcessBlobGas()
	require.NoError(t, err)
	require.DeepEqual(t, uint64(99), g)
}

func TestWrapSilaPayloadHeaderDeneb(t *testing.T) {
	data := &silaenginev1.SilaPayloadHeaderDeneb{
		ParentHash:       []byte("parenthash"),
		FeeRecipient:     []byte("feerecipient"),
		StateRoot:        []byte("stateroot"),
		ReceiptsRoot:     []byte("receiptsroot"),
		LogsBloom:        []byte("logsbloom"),
		PrevRandao:       []byte("prevrandao"),
		BlockNumber:      11,
		GasLimit:         22,
		GasUsed:          33,
		Timestamp:        44,
		ExtraData:        []byte("extradata"),
		BaseFeePerGas:    []byte("basefeepergas"),
		BlockHash:        []byte("blockhash"),
		TransactionsRoot: []byte("transactionsroot"),
		WithdrawalsRoot:  []byte("withdrawalsroot"),
		BlobGasUsed:      88,
		ExcessBlobGas:    99,
	}
	payload, err := blocks.WrappedSilaPayloadHeaderDeneb(data)
	require.NoError(t, err)

	g, err := payload.BlobGasUsed()
	require.NoError(t, err)
	require.DeepEqual(t, uint64(88), g)

	g, err = payload.ExcessBlobGas()
	require.NoError(t, err)
	require.DeepEqual(t, uint64(99), g)
}

func TestWrapSilaPayloadDeneb_SSZ(t *testing.T) {
	payload := createWrappedPayloadDeneb(t)
	rt, err := payload.HashTreeRoot()
	assert.NoError(t, err)
	assert.NotEmpty(t, rt)

	var b []byte
	b, err = payload.MarshalSSZTo(b)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(b))
	encoded, err := payload.MarshalSSZ()
	require.NoError(t, err)
	assert.NotEqual(t, 0, payload.SizeSSZ())
	assert.NoError(t, payload.UnmarshalSSZ(encoded))
}

func TestWrapSilaPayloadHeaderDeneb_SSZ(t *testing.T) {
	payload := createWrappedPayloadHeaderDeneb(t)
	rt, err := payload.HashTreeRoot()
	assert.NoError(t, err)
	assert.NotEmpty(t, rt)

	var b []byte
	b, err = payload.MarshalSSZTo(b)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(b))
	encoded, err := payload.MarshalSSZ()
	require.NoError(t, err)
	assert.NotEqual(t, 0, payload.SizeSSZ())
	assert.NoError(t, payload.UnmarshalSSZ(encoded))
}

func createWrappedPayload(t testing.TB) interfaces.ExecutionData {
	wsb, err := blocks.WrappedSilaPayload(&silaenginev1.SilaPayload{
		ParentHash:    make([]byte, fieldparams.RootLength),
		FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:     make([]byte, fieldparams.RootLength),
		ReceiptsRoot:  make([]byte, fieldparams.RootLength),
		LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
		PrevRandao:    make([]byte, fieldparams.RootLength),
		BlockNumber:   0,
		GasLimit:      0,
		GasUsed:       0,
		Timestamp:     0,
		ExtraData:     make([]byte, 0),
		BaseFeePerGas: make([]byte, fieldparams.RootLength),
		BlockHash:     make([]byte, fieldparams.RootLength),
		Transactions:  make([][]byte, 0),
	})
	require.NoError(t, err)
	return wsb
}

func createWrappedPayloadHeader(t testing.TB) interfaces.ExecutionData {
	wsb, err := blocks.WrappedSilaPayloadHeader(&silaenginev1.SilaPayloadHeader{
		ParentHash:       make([]byte, fieldparams.RootLength),
		FeeRecipient:     make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:        make([]byte, fieldparams.RootLength),
		ReceiptsRoot:     make([]byte, fieldparams.RootLength),
		LogsBloom:        make([]byte, fieldparams.LogsBloomLength),
		PrevRandao:       make([]byte, fieldparams.RootLength),
		BlockNumber:      0,
		GasLimit:         0,
		GasUsed:          0,
		Timestamp:        0,
		ExtraData:        make([]byte, 0),
		BaseFeePerGas:    make([]byte, fieldparams.RootLength),
		BlockHash:        make([]byte, fieldparams.RootLength),
		TransactionsRoot: make([]byte, fieldparams.RootLength),
	})
	require.NoError(t, err)
	return wsb
}

func createWrappedPayloadCapella(t testing.TB) interfaces.ExecutionData {
	payload, err := blocks.WrappedSilaPayloadCapella(&silaenginev1.SilaPayloadCapella{
		ParentHash:    make([]byte, fieldparams.RootLength),
		FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:     make([]byte, fieldparams.RootLength),
		ReceiptsRoot:  make([]byte, fieldparams.RootLength),
		LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
		PrevRandao:    make([]byte, fieldparams.RootLength),
		BlockNumber:   0,
		GasLimit:      0,
		GasUsed:       0,
		Timestamp:     0,
		ExtraData:     make([]byte, 0),
		BaseFeePerGas: make([]byte, fieldparams.RootLength),
		BlockHash:     make([]byte, fieldparams.RootLength),
		Transactions:  make([][]byte, 0),
		Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
	})
	require.NoError(t, err)
	return payload
}

func createWrappedPayloadHeaderCapella(t testing.TB) interfaces.ExecutionData {
	payload, err := blocks.WrappedSilaPayloadHeaderCapella(&silaenginev1.SilaPayloadHeaderCapella{
		ParentHash:       make([]byte, fieldparams.RootLength),
		FeeRecipient:     make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:        make([]byte, fieldparams.RootLength),
		ReceiptsRoot:     make([]byte, fieldparams.RootLength),
		LogsBloom:        make([]byte, fieldparams.LogsBloomLength),
		PrevRandao:       make([]byte, fieldparams.RootLength),
		BlockNumber:      0,
		GasLimit:         0,
		GasUsed:          0,
		Timestamp:        0,
		ExtraData:        make([]byte, 0),
		BaseFeePerGas:    make([]byte, fieldparams.RootLength),
		BlockHash:        make([]byte, fieldparams.RootLength),
		TransactionsRoot: make([]byte, fieldparams.RootLength),
		WithdrawalsRoot:  make([]byte, fieldparams.RootLength),
	})
	require.NoError(t, err)
	return payload
}

func createWrappedPayloadDeneb(t testing.TB) interfaces.ExecutionData {
	payload, err := blocks.WrappedSilaPayloadDeneb(&silaenginev1.SilaPayloadDeneb{
		ParentHash:    make([]byte, fieldparams.RootLength),
		FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:     make([]byte, fieldparams.RootLength),
		ReceiptsRoot:  make([]byte, fieldparams.RootLength),
		LogsBloom:     make([]byte, fieldparams.LogsBloomLength),
		PrevRandao:    make([]byte, fieldparams.RootLength),
		BlockNumber:   0,
		GasLimit:      0,
		GasUsed:       0,
		Timestamp:     0,
		ExtraData:     make([]byte, 0),
		BaseFeePerGas: make([]byte, fieldparams.RootLength),
		BlockHash:     make([]byte, fieldparams.RootLength),
		Transactions:  make([][]byte, 0),
		Withdrawals:   make([]*silaenginev1.Withdrawal, 0),
		BlobGasUsed:   0,
		ExcessBlobGas: 0,
	})
	require.NoError(t, err)
	return payload
}

func createWrappedPayloadHeaderDeneb(t testing.TB) interfaces.ExecutionData {
	payload, err := blocks.WrappedSilaPayloadHeaderDeneb(&silaenginev1.SilaPayloadHeaderDeneb{
		ParentHash:       make([]byte, fieldparams.RootLength),
		FeeRecipient:     make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:        make([]byte, fieldparams.RootLength),
		ReceiptsRoot:     make([]byte, fieldparams.RootLength),
		LogsBloom:        make([]byte, fieldparams.LogsBloomLength),
		PrevRandao:       make([]byte, fieldparams.RootLength),
		BlockNumber:      0,
		GasLimit:         0,
		GasUsed:          0,
		Timestamp:        0,
		ExtraData:        make([]byte, 0),
		BaseFeePerGas:    make([]byte, fieldparams.RootLength),
		BlockHash:        make([]byte, fieldparams.RootLength),
		TransactionsRoot: make([]byte, fieldparams.RootLength),
		WithdrawalsRoot:  make([]byte, fieldparams.RootLength),
		BlobGasUsed:      0,
		ExcessBlobGas:    0,
	})
	require.NoError(t, err)
	return payload
}
