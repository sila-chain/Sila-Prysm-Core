package structs

import (
	"fmt"
	"testing"

	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila/common"
	"github.com/sila-chain/Sila/common/hexutil"
)

func fillByteSlice(sliceLength int, value byte) []byte {
	bytes := make([]byte, sliceLength)

	for index := range bytes {
		bytes[index] = value
	}

	return bytes
}

// TestSilaPayloadFromConsensus_HappyPath checks the
// SilaPayloadFromConsensus function under normal conditions.
func TestSilaPayloadFromConsensus_HappyPath(t *testing.T) {
	consensusPayload := &silaenginev1.SilaPayload{
		ParentHash:    fillByteSlice(common.HashLength, 0xaa),
		FeeRecipient:  fillByteSlice(20, 0xbb),
		StateRoot:     fillByteSlice(32, 0xcc),
		ReceiptsRoot:  fillByteSlice(32, 0xdd),
		LogsBloom:     fillByteSlice(256, 0xee),
		PrevRandao:    fillByteSlice(32, 0xff),
		BlockNumber:   12345,
		GasLimit:      15000000,
		GasUsed:       8000000,
		Timestamp:     1680000000,
		ExtraData:     fillByteSlice(8, 0x11),
		BaseFeePerGas: fillByteSlice(32, 0x01),
		BlockHash:     fillByteSlice(common.HashLength, 0x22),
		Transactions: [][]byte{
			fillByteSlice(10, 0x33),
			fillByteSlice(10, 0x44),
		},
	}

	result, err := SilaPayloadFromConsensus(consensusPayload)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, hexutil.Encode(consensusPayload.ParentHash), result.ParentHash)
	require.Equal(t, hexutil.Encode(consensusPayload.FeeRecipient), result.FeeRecipient)
	require.Equal(t, hexutil.Encode(consensusPayload.StateRoot), result.StateRoot)
	require.Equal(t, hexutil.Encode(consensusPayload.ReceiptsRoot), result.ReceiptsRoot)
	require.Equal(t, fmt.Sprintf("%d", consensusPayload.BlockNumber), result.BlockNumber)
}

// TestSilaPayload_ToConsensus_HappyPath checks the
// (*SilaPayload).ToConsensus function under normal conditions.
func TestSilaPayload_ToConsensus_HappyPath(t *testing.T) {
	payload := &SilaPayload{
		ParentHash:    hexutil.Encode(fillByteSlice(common.HashLength, 0xaa)),
		FeeRecipient:  hexutil.Encode(fillByteSlice(20, 0xbb)),
		StateRoot:     hexutil.Encode(fillByteSlice(32, 0xcc)),
		ReceiptsRoot:  hexutil.Encode(fillByteSlice(32, 0xdd)),
		LogsBloom:     hexutil.Encode(fillByteSlice(256, 0xee)),
		PrevRandao:    hexutil.Encode(fillByteSlice(32, 0xff)),
		BlockNumber:   "12345",
		GasLimit:      "15000000",
		GasUsed:       "8000000",
		Timestamp:     "1680000000",
		ExtraData:     "0x11111111",
		BaseFeePerGas: "1234",
		BlockHash:     hexutil.Encode(fillByteSlice(common.HashLength, 0x22)),
		Transactions: []string{
			hexutil.Encode(fillByteSlice(10, 0x33)),
			hexutil.Encode(fillByteSlice(10, 0x44)),
		},
	}

	result, err := payload.ToConsensus()
	require.NoError(t, err)
	require.DeepEqual(t, result.ParentHash, fillByteSlice(common.HashLength, 0xaa))
	require.DeepEqual(t, result.FeeRecipient, fillByteSlice(20, 0xbb))
	require.DeepEqual(t, result.StateRoot, fillByteSlice(32, 0xcc))
}

// TestSilaPayloadHeaderFromConsensus_HappyPath checks the
// SilaPayloadHeaderFromConsensus function under normal conditions.
func TestSilaPayloadHeaderFromConsensus_HappyPath(t *testing.T) {
	consensusHeader := &silaenginev1.SilaPayloadHeader{
		ParentHash:       fillByteSlice(common.HashLength, 0xaa),
		FeeRecipient:     fillByteSlice(20, 0xbb),
		StateRoot:        fillByteSlice(32, 0xcc),
		ReceiptsRoot:     fillByteSlice(32, 0xdd),
		LogsBloom:        fillByteSlice(256, 0xee),
		PrevRandao:       fillByteSlice(32, 0xff),
		BlockNumber:      9999,
		GasLimit:         5000000,
		GasUsed:          2500000,
		Timestamp:        1111111111,
		ExtraData:        fillByteSlice(4, 0x12),
		BaseFeePerGas:    fillByteSlice(32, 0x34),
		BlockHash:        fillByteSlice(common.HashLength, 0x56),
		TransactionsRoot: fillByteSlice(32, 0x78),
	}

	result, err := SilaPayloadHeaderFromConsensus(consensusHeader)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, hexutil.Encode(consensusHeader.ParentHash), result.ParentHash)
	require.Equal(t, fmt.Sprintf("%d", consensusHeader.BlockNumber), result.BlockNumber)
}

// TestSilaPayloadHeader_ToConsensus_HappyPath checks the
// (*SilaPayloadHeader).ToConsensus function under normal conditions.
func TestSilaPayloadHeader_ToConsensus_HappyPath(t *testing.T) {
	header := &SilaPayloadHeader{
		ParentHash:       hexutil.Encode(fillByteSlice(common.HashLength, 0xaa)),
		FeeRecipient:     hexutil.Encode(fillByteSlice(20, 0xbb)),
		StateRoot:        hexutil.Encode(fillByteSlice(32, 0xcc)),
		ReceiptsRoot:     hexutil.Encode(fillByteSlice(32, 0xdd)),
		LogsBloom:        hexutil.Encode(fillByteSlice(256, 0xee)),
		PrevRandao:       hexutil.Encode(fillByteSlice(32, 0xff)),
		BlockNumber:      "9999",
		GasLimit:         "5000000",
		GasUsed:          "2500000",
		Timestamp:        "1111111111",
		ExtraData:        "0x1234abcd",
		BaseFeePerGas:    "1234",
		BlockHash:        hexutil.Encode(fillByteSlice(common.HashLength, 0x56)),
		TransactionsRoot: hexutil.Encode(fillByteSlice(32, 0x78)),
	}

	result, err := header.ToConsensus()
	require.NoError(t, err)
	require.DeepEqual(t, hexutil.Encode(result.ParentHash), header.ParentHash)
	require.DeepEqual(t, hexutil.Encode(result.FeeRecipient), header.FeeRecipient)
	require.DeepEqual(t, hexutil.Encode(result.StateRoot), header.StateRoot)
}

// TestSilaPayloadCapellaFromConsensus_HappyPath checks the
// SilaPayloadCapellaFromConsensus function under normal conditions.
func TestSilaPayloadCapellaFromConsensus_HappyPath(t *testing.T) {
	capellaPayload := &silaenginev1.SilaPayloadCapella{
		ParentHash:    fillByteSlice(common.HashLength, 0xaa),
		FeeRecipient:  fillByteSlice(20, 0xbb),
		StateRoot:     fillByteSlice(32, 0xcc),
		ReceiptsRoot:  fillByteSlice(32, 0xdd),
		LogsBloom:     fillByteSlice(256, 0xee),
		PrevRandao:    fillByteSlice(32, 0xff),
		BlockNumber:   123,
		GasLimit:      9876543,
		GasUsed:       1234567,
		Timestamp:     5555555,
		ExtraData:     fillByteSlice(6, 0x11),
		BaseFeePerGas: fillByteSlice(32, 0x22),
		BlockHash:     fillByteSlice(common.HashLength, 0x33),
		Transactions: [][]byte{
			fillByteSlice(5, 0x44),
		},
		Withdrawals: []*silaenginev1.Withdrawal{
			{
				Index:          1,
				ValidatorIndex: 2,
				Address:        fillByteSlice(20, 0xaa),
				Amount:         100,
			},
		},
	}

	result, err := SilaPayloadCapellaFromConsensus(capellaPayload)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, hexutil.Encode(capellaPayload.ParentHash), result.ParentHash)
	require.Equal(t, len(capellaPayload.Transactions), len(result.Transactions))
	require.Equal(t, len(capellaPayload.Withdrawals), len(result.Withdrawals))
}

// TestSilaPayloadCapella_ToConsensus_HappyPath checks the
// (*SilaPayloadCapella).ToConsensus function under normal conditions.
func TestSilaPayloadCapella_ToConsensus_HappyPath(t *testing.T) {
	capella := &SilaPayloadCapella{
		ParentHash:    hexutil.Encode(fillByteSlice(common.HashLength, 0xaa)),
		FeeRecipient:  hexutil.Encode(fillByteSlice(20, 0xbb)),
		StateRoot:     hexutil.Encode(fillByteSlice(32, 0xcc)),
		ReceiptsRoot:  hexutil.Encode(fillByteSlice(32, 0xdd)),
		LogsBloom:     hexutil.Encode(fillByteSlice(256, 0xee)),
		PrevRandao:    hexutil.Encode(fillByteSlice(32, 0xff)),
		BlockNumber:   "123",
		GasLimit:      "9876543",
		GasUsed:       "1234567",
		Timestamp:     "5555555",
		ExtraData:     hexutil.Encode(fillByteSlice(6, 0x11)),
		BaseFeePerGas: "1234",
		BlockHash:     hexutil.Encode(fillByteSlice(common.HashLength, 0x33)),
		Transactions: []string{
			hexutil.Encode(fillByteSlice(5, 0x44)),
		},
		Withdrawals: []*Withdrawal{
			{
				WithdrawalIndex:  "1",
				ValidatorIndex:   "2",
				ExecutionAddress: hexutil.Encode(fillByteSlice(20, 0xaa)),
				Amount:           "100",
			},
		},
	}

	result, err := capella.ToConsensus()
	require.NoError(t, err)
	require.DeepEqual(t, hexutil.Encode(result.ParentHash), capella.ParentHash)
	require.DeepEqual(t, hexutil.Encode(result.FeeRecipient), capella.FeeRecipient)
	require.DeepEqual(t, hexutil.Encode(result.StateRoot), capella.StateRoot)
}

// TestSilaPayloadDenebFromConsensus_HappyPath checks the
// SilaPayloadDenebFromConsensus function under normal conditions.
func TestSilaPayloadDenebFromConsensus_HappyPath(t *testing.T) {
	denebPayload := &silaenginev1.SilaPayloadDeneb{
		ParentHash:    fillByteSlice(common.HashLength, 0xaa),
		FeeRecipient:  fillByteSlice(20, 0xbb),
		StateRoot:     fillByteSlice(32, 0xcc),
		ReceiptsRoot:  fillByteSlice(32, 0xdd),
		LogsBloom:     fillByteSlice(256, 0xee),
		PrevRandao:    fillByteSlice(32, 0xff),
		BlockNumber:   999,
		GasLimit:      2222222,
		GasUsed:       1111111,
		Timestamp:     666666,
		ExtraData:     fillByteSlice(6, 0x11),
		BaseFeePerGas: fillByteSlice(32, 0x22),
		BlockHash:     fillByteSlice(common.HashLength, 0x33),
		Transactions: [][]byte{
			fillByteSlice(5, 0x44),
		},
		Withdrawals: []*silaenginev1.Withdrawal{
			{
				Index:          1,
				ValidatorIndex: 2,
				Address:        fillByteSlice(20, 0xaa),
				Amount:         100,
			},
		},
		BlobGasUsed:   1234,
		ExcessBlobGas: 5678,
	}

	result, err := SilaPayloadDenebFromConsensus(denebPayload)
	require.NoError(t, err)
	require.Equal(t, hexutil.Encode(denebPayload.ParentHash), result.ParentHash)
	require.Equal(t, len(denebPayload.Transactions), len(result.Transactions))
	require.Equal(t, len(denebPayload.Withdrawals), len(result.Withdrawals))
	require.Equal(t, "1234", result.BlobGasUsed)
	require.Equal(t, fmt.Sprintf("%d", denebPayload.BlockNumber), result.BlockNumber)
}

// TestSilaPayloadDeneb_ToConsensus_HappyPath checks the
// (*SilaPayloadDeneb).ToConsensus function under normal conditions.
func TestSilaPayloadDeneb_ToConsensus_HappyPath(t *testing.T) {
	deneb := &SilaPayloadDeneb{
		ParentHash:    hexutil.Encode(fillByteSlice(common.HashLength, 0xaa)),
		FeeRecipient:  hexutil.Encode(fillByteSlice(20, 0xbb)),
		StateRoot:     hexutil.Encode(fillByteSlice(32, 0xcc)),
		ReceiptsRoot:  hexutil.Encode(fillByteSlice(32, 0xdd)),
		LogsBloom:     hexutil.Encode(fillByteSlice(256, 0xee)),
		PrevRandao:    hexutil.Encode(fillByteSlice(32, 0xff)),
		BlockNumber:   "999",
		GasLimit:      "2222222",
		GasUsed:       "1111111",
		Timestamp:     "666666",
		ExtraData:     hexutil.Encode(fillByteSlice(6, 0x11)),
		BaseFeePerGas: "1234",
		BlockHash:     hexutil.Encode(fillByteSlice(common.HashLength, 0x33)),
		Transactions: []string{
			hexutil.Encode(fillByteSlice(5, 0x44)),
		},
		Withdrawals: []*Withdrawal{
			{
				WithdrawalIndex:  "1",
				ValidatorIndex:   "2",
				ExecutionAddress: hexutil.Encode(fillByteSlice(20, 0xaa)),
				Amount:           "100",
			},
		},
		BlobGasUsed:   "1234",
		ExcessBlobGas: "5678",
	}

	result, err := deneb.ToConsensus()
	require.NoError(t, err)
	require.DeepEqual(t, hexutil.Encode(result.ParentHash), deneb.ParentHash)
	require.DeepEqual(t, hexutil.Encode(result.FeeRecipient), deneb.FeeRecipient)
	require.Equal(t, result.BlockNumber, uint64(999))
}

func TestSilaPayloadHeaderCapellaFromConsensus_HappyPath(t *testing.T) {
	capellaHeader := &silaenginev1.SilaPayloadHeaderCapella{
		ParentHash:       fillByteSlice(common.HashLength, 0xaa),
		FeeRecipient:     fillByteSlice(20, 0xbb),
		StateRoot:        fillByteSlice(32, 0xcc),
		ReceiptsRoot:     fillByteSlice(32, 0xdd),
		LogsBloom:        fillByteSlice(256, 0xee),
		PrevRandao:       fillByteSlice(32, 0xff),
		BlockNumber:      555,
		GasLimit:         1111111,
		GasUsed:          222222,
		Timestamp:        3333333333,
		ExtraData:        fillByteSlice(4, 0x12),
		BaseFeePerGas:    fillByteSlice(32, 0x34),
		BlockHash:        fillByteSlice(common.HashLength, 0x56),
		TransactionsRoot: fillByteSlice(32, 0x78),
		WithdrawalsRoot:  fillByteSlice(32, 0x99),
	}

	result, err := SilaPayloadHeaderCapellaFromConsensus(capellaHeader)
	require.NoError(t, err)
	require.Equal(t, hexutil.Encode(capellaHeader.ParentHash), result.ParentHash)
	require.DeepEqual(t, hexutil.Encode(capellaHeader.WithdrawalsRoot), result.WithdrawalsRoot)
}

func TestSilaPayloadHeaderCapella_ToConsensus_HappyPath(t *testing.T) {
	header := &SilaPayloadHeaderCapella{
		ParentHash:       hexutil.Encode(fillByteSlice(common.HashLength, 0xaa)),
		FeeRecipient:     hexutil.Encode(fillByteSlice(20, 0xbb)),
		StateRoot:        hexutil.Encode(fillByteSlice(32, 0xcc)),
		ReceiptsRoot:     hexutil.Encode(fillByteSlice(32, 0xdd)),
		LogsBloom:        hexutil.Encode(fillByteSlice(256, 0xee)),
		PrevRandao:       hexutil.Encode(fillByteSlice(32, 0xff)),
		BlockNumber:      "555",
		GasLimit:         "1111111",
		GasUsed:          "222222",
		Timestamp:        "3333333333",
		ExtraData:        "0x1234abcd",
		BaseFeePerGas:    "1234",
		BlockHash:        hexutil.Encode(fillByteSlice(common.HashLength, 0x56)),
		TransactionsRoot: hexutil.Encode(fillByteSlice(32, 0x78)),
		WithdrawalsRoot:  hexutil.Encode(fillByteSlice(32, 0x99)),
	}

	result, err := header.ToConsensus()
	require.NoError(t, err)
	require.DeepEqual(t, hexutil.Encode(result.ParentHash), header.ParentHash)
	require.DeepEqual(t, hexutil.Encode(result.FeeRecipient), header.FeeRecipient)
	require.DeepEqual(t, hexutil.Encode(result.StateRoot), header.StateRoot)
	require.DeepEqual(t, hexutil.Encode(result.ReceiptsRoot), header.ReceiptsRoot)
	require.DeepEqual(t, hexutil.Encode(result.WithdrawalsRoot), header.WithdrawalsRoot)
}

func TestSilaPayloadHeaderDenebFromConsensus_HappyPath(t *testing.T) {
	denebHeader := &silaenginev1.SilaPayloadHeaderDeneb{
		ParentHash:       fillByteSlice(common.HashLength, 0xaa),
		FeeRecipient:     fillByteSlice(20, 0xbb),
		StateRoot:        fillByteSlice(32, 0xcc),
		ReceiptsRoot:     fillByteSlice(32, 0xdd),
		LogsBloom:        fillByteSlice(256, 0xee),
		PrevRandao:       fillByteSlice(32, 0xff),
		BlockNumber:      999,
		GasLimit:         5000000,
		GasUsed:          2500000,
		Timestamp:        4444444444,
		ExtraData:        fillByteSlice(4, 0x12),
		BaseFeePerGas:    fillByteSlice(32, 0x34),
		BlockHash:        fillByteSlice(common.HashLength, 0x56),
		TransactionsRoot: fillByteSlice(32, 0x78),
		WithdrawalsRoot:  fillByteSlice(32, 0x99),
		BlobGasUsed:      1234,
		ExcessBlobGas:    5678,
	}

	result, err := SilaPayloadHeaderDenebFromConsensus(denebHeader)
	require.NoError(t, err)
	require.Equal(t, hexutil.Encode(denebHeader.ParentHash), result.ParentHash)
	require.DeepEqual(t, hexutil.Encode(denebHeader.FeeRecipient), result.FeeRecipient)
	require.DeepEqual(t, hexutil.Encode(denebHeader.StateRoot), result.StateRoot)
	require.DeepEqual(t, fmt.Sprintf("%d", denebHeader.BlobGasUsed), result.BlobGasUsed)
}

func TestSilaPayloadHeaderDeneb_ToConsensus_HappyPath(t *testing.T) {
	header := &SilaPayloadHeaderDeneb{
		ParentHash:       hexutil.Encode(fillByteSlice(common.HashLength, 0xaa)),
		FeeRecipient:     hexutil.Encode(fillByteSlice(20, 0xbb)),
		StateRoot:        hexutil.Encode(fillByteSlice(32, 0xcc)),
		ReceiptsRoot:     hexutil.Encode(fillByteSlice(32, 0xdd)),
		LogsBloom:        hexutil.Encode(fillByteSlice(256, 0xee)),
		PrevRandao:       hexutil.Encode(fillByteSlice(32, 0xff)),
		BlockNumber:      "999",
		GasLimit:         "5000000",
		GasUsed:          "2500000",
		Timestamp:        "4444444444",
		ExtraData:        "0x1234abcd",
		BaseFeePerGas:    "1234",
		BlockHash:        hexutil.Encode(fillByteSlice(common.HashLength, 0x56)),
		TransactionsRoot: hexutil.Encode(fillByteSlice(32, 0x78)),
		WithdrawalsRoot:  hexutil.Encode(fillByteSlice(32, 0x99)),
		BlobGasUsed:      "1234",
		ExcessBlobGas:    "5678",
	}

	result, err := header.ToConsensus()
	require.NoError(t, err)
	require.DeepEqual(t, hexutil.Encode(result.ParentHash), header.ParentHash)
	require.DeepEqual(t, result.BlobGasUsed, uint64(1234))
	require.DeepEqual(t, result.ExcessBlobGas, uint64(5678))
	require.DeepEqual(t, result.BlockNumber, uint64(999))
}

func TestWithdrawalRequestsFromConsensus_HappyPath(t *testing.T) {
	consensusRequests := []*silaenginev1.WithdrawalRequest{
		{
			SourceAddress:   fillByteSlice(20, 0xbb),
			ValidatorPubkey: fillByteSlice(48, 0xbb),
			Amount:          12345,
		},
		{
			SourceAddress:   fillByteSlice(20, 0xcc),
			ValidatorPubkey: fillByteSlice(48, 0xcc),
			Amount:          54321,
		},
	}

	result := WithdrawalRequestsFromConsensus(consensusRequests)
	require.DeepEqual(t, len(result), len(consensusRequests))
	require.DeepEqual(t, result[0].Amount, fmt.Sprintf("%d", consensusRequests[0].Amount))
}

func TestWithdrawalRequestFromConsensus_HappyPath(t *testing.T) {
	req := &silaenginev1.WithdrawalRequest{
		SourceAddress:   fillByteSlice(20, 0xbb),
		ValidatorPubkey: fillByteSlice(48, 0xbb),
		Amount:          42,
	}
	result := WithdrawalRequestFromConsensus(req)
	require.NotNil(t, result)
	require.DeepEqual(t, result.SourceAddress, hexutil.Encode(fillByteSlice(20, 0xbb)))
}

func TestWithdrawalRequest_ToConsensus_HappyPath(t *testing.T) {
	withdrawalReq := &WithdrawalRequest{
		SourceAddress:   hexutil.Encode(fillByteSlice(20, 111)),
		ValidatorPubkey: hexutil.Encode(fillByteSlice(48, 123)),
		Amount:          "12345",
	}
	result, err := withdrawalReq.ToConsensus()
	require.NoError(t, err)
	require.DeepEqual(t, result.Amount, uint64(12345))
}

func TestConsolidationRequestsFromConsensus_HappyPath(t *testing.T) {
	consensusRequests := []*silaenginev1.ConsolidationRequest{
		{
			SourceAddress: fillByteSlice(20, 111),
			SourcePubkey:  fillByteSlice(48, 112),
			TargetPubkey:  fillByteSlice(48, 113),
		},
	}
	result := ConsolidationRequestsFromConsensus(consensusRequests)
	require.DeepEqual(t, len(result), len(consensusRequests))
	require.DeepEqual(t, result[0].SourceAddress, "0x6f6f6f6f6f6f6f6f6f6f6f6f6f6f6f6f6f6f6f6f")
}

func TestDepositRequestsFromConsensus_HappyPath(t *testing.T) {
	ds := []*silaenginev1.DepositRequest{
		{
			Pubkey:                fillByteSlice(48, 0xbb),
			WithdrawalCredentials: fillByteSlice(32, 0xdd),
			Amount:                98765,
			Signature:             fillByteSlice(96, 0xff),
			Index:                 111,
		},
	}
	result := DepositRequestsFromConsensus(ds)
	require.DeepEqual(t, len(result), len(ds))
	require.DeepEqual(t, result[0].Amount, "98765")
}

func TestDepositRequest_ToConsensus_HappyPath(t *testing.T) {
	req := &DepositRequest{
		Pubkey:                hexutil.Encode(fillByteSlice(48, 0xbb)),
		WithdrawalCredentials: hexutil.Encode(fillByteSlice(32, 0xaa)),
		Amount:                "123",
		Signature:             hexutil.Encode(fillByteSlice(96, 0xdd)),
		Index:                 "456",
	}

	result, err := req.ToConsensus()
	require.NoError(t, err)
	require.DeepEqual(t, result.Amount, uint64(123))
	require.DeepEqual(t, result.Signature, fillByteSlice(96, 0xdd))
}

func TestExecutionRequestsFromConsensus_HappyPath(t *testing.T) {
	er := &silaenginev1.ExecutionRequests{
		Deposits: []*silaenginev1.DepositRequest{
			{
				Pubkey:                fillByteSlice(48, 0xba),
				WithdrawalCredentials: fillByteSlice(32, 0xaa),
				Amount:                33,
				Signature:             fillByteSlice(96, 0xff),
				Index:                 44,
			},
		},
		Withdrawals: []*silaenginev1.WithdrawalRequest{
			{
				SourceAddress:   fillByteSlice(20, 0xaa),
				ValidatorPubkey: fillByteSlice(48, 0xba),
				Amount:          555,
			},
		},
		Consolidations: []*silaenginev1.ConsolidationRequest{
			{
				SourceAddress: fillByteSlice(20, 0xdd),
				SourcePubkey:  fillByteSlice(48, 0xdd),
				TargetPubkey:  fillByteSlice(48, 0xcc),
			},
		},
	}

	result := ExecutionRequestsFromConsensus(er)
	require.NotNil(t, result)
	require.Equal(t, 1, len(result.Deposits))
	require.Equal(t, "33", result.Deposits[0].Amount)
	require.Equal(t, 1, len(result.Withdrawals))
	require.Equal(t, "555", result.Withdrawals[0].Amount)
	require.Equal(t, 1, len(result.Consolidations))
	require.Equal(t, "0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", result.Consolidations[0].TargetPubkey)
}

func TestExecutionRequests_ToConsensus_HappyPath(t *testing.T) {
	execReq := &ExecutionRequests{
		Deposits: []*DepositRequest{
			{
				Pubkey:                hexutil.Encode(fillByteSlice(48, 0xbb)),
				WithdrawalCredentials: hexutil.Encode(fillByteSlice(32, 0xaa)),
				Amount:                "33",
				Signature:             hexutil.Encode(fillByteSlice(96, 0xff)),
				Index:                 "44",
			},
		},
		Withdrawals: []*WithdrawalRequest{
			{
				SourceAddress:   hexutil.Encode(fillByteSlice(20, 0xdd)),
				ValidatorPubkey: hexutil.Encode(fillByteSlice(48, 0xbb)),
				Amount:          "555",
			},
		},
		Consolidations: []*ConsolidationRequest{
			{
				SourceAddress: hexutil.Encode(fillByteSlice(20, 0xcc)),
				SourcePubkey:  hexutil.Encode(fillByteSlice(48, 0xbb)),
				TargetPubkey:  hexutil.Encode(fillByteSlice(48, 0xcc)),
			},
		},
	}

	result, err := execReq.ToConsensus()
	require.NoError(t, err)

	require.Equal(t, 1, len(result.Deposits))
	require.Equal(t, uint64(33), result.Deposits[0].Amount)
	require.Equal(t, 1, len(result.Withdrawals))
	require.Equal(t, uint64(555), result.Withdrawals[0].Amount)
	require.Equal(t, 1, len(result.Consolidations))
	require.DeepEqual(t, fillByteSlice(48, 0xcc), result.Consolidations[0].TargetPubkey)
}
