package silaexec_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec"
	pb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	silaEngine "github.com/sila-chain/Sila/beacon/silaEngine"
	"github.com/sila-chain/Sila/common"
	"github.com/sila-chain/Sila/core/types"
)

func FuzzForkChoiceResponse(f *testing.F) {
	valHash := common.Hash([32]byte{0xFF, 0x01})
	payloadID := silaEngine.PayloadID([8]byte{0x01, 0xFF, 0xAA, 0x00, 0xEE, 0xFE, 0x00, 0x00})
	valErr := "asjajshjahsaj"
	seed := &silaEngine.ForkChoiceResponse{
		PayloadStatus: silaEngine.PayloadStatusV1{
			Status:          "INVALID_TERMINAL_BLOCK",
			LatestValidHash: &valHash,
			ValidationError: &valErr,
		},
		PayloadID: &payloadID,
	}
	output, err := json.Marshal(seed)
	assert.NoError(f, err)
	f.Add(output)
	f.Fuzz(func(t *testing.T, jsonBlob []byte) {
		upstreamResp := &silaEngine.ForkChoiceResponse{}
		silaResp := &silaexec.ForkchoiceUpdatedResponse{}
		upstreamErr := json.Unmarshal(jsonBlob, upstreamResp)
		silaErr := json.Unmarshal(jsonBlob, silaResp)
		assert.Equal(t, upstreamErr != nil, silaErr != nil, fmt.Sprintf("upstream and Sila unmarshaller return inconsistent errors. %v and %v", upstreamErr, silaErr))
		// Nothing to marshal if we have an error.
		if upstreamErr != nil {
			return
		}
		upstreamBlob, upstreamErr := json.Marshal(upstreamResp)
		silaBlob, silaErr := json.Marshal(silaResp)
		assert.Equal(t, upstreamErr != nil, silaErr != nil, "upstream and Sila unmarshaller return inconsistent errors")
		newUpstreamResp := &silaEngine.ForkChoiceResponse{}
		newUpstreamErr := json.Unmarshal(silaBlob, newUpstreamResp)
		assert.NoError(t, newUpstreamErr)
		if newUpstreamResp.PayloadStatus.Status == "UNKNOWN" {
			return
		}

		newUpstreamResp2 := &silaEngine.ForkChoiceResponse{}
		newUpstreamErr = json.Unmarshal(upstreamBlob, newUpstreamResp2)
		assert.NoError(t, newUpstreamErr)

		assert.DeepEqual(t, newUpstreamResp.PayloadID, newUpstreamResp2.PayloadID)
		assert.DeepEqual(t, newUpstreamResp.PayloadStatus.Status, newUpstreamResp2.PayloadStatus.Status)
		assert.DeepEqual(t, newUpstreamResp.PayloadStatus.LatestValidHash, newUpstreamResp2.PayloadStatus.LatestValidHash)
		isNilOrEmpty := newUpstreamResp.PayloadStatus.ValidationError == nil || (*newUpstreamResp.PayloadStatus.ValidationError == "")
		isNilOrEmpty2 := newUpstreamResp2.PayloadStatus.ValidationError == nil || (*newUpstreamResp2.PayloadStatus.ValidationError == "")
		assert.DeepEqual(t, isNilOrEmpty, isNilOrEmpty2)
		if !isNilOrEmpty {
			assert.DeepEqual(t, *newUpstreamResp.PayloadStatus.ValidationError, *newUpstreamResp2.PayloadStatus.ValidationError)
		}
	})
}

func FuzzSilaPayload(f *testing.F) {
	logsBloom := [256]byte{'j', 'u', 'n', 'k'}
	execData := &silaEngine.SilaExecutionPayloadEnvelope{
		SilaExecutionPayload: &silaEngine.ExecutableData{
			ParentHash:    common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
			FeeRecipient:  common.Address([20]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF}),
			StateRoot:     common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
			ReceiptsRoot:  common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
			LogsBloom:     logsBloom[:],
			Random:        common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
			Number:        math.MaxUint64,
			GasLimit:      math.MaxUint64,
			GasUsed:       math.MaxUint64,
			Timestamp:     100,
			ExtraData:     []byte{},
			BaseFeePerGas: big.NewInt(math.MaxInt),
			BlockHash:     common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
			Transactions:  [][]byte{{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}, {0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}, {0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}, {0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}},
			Withdrawals:   []*types.Withdrawal{},
		},
		BlockValue: nil,
	}
	output, err := json.Marshal(execData)
	assert.NoError(f, err)
	f.Add(output)
	f.Fuzz(func(t *testing.T, jsonBlob []byte) {
		upstreamResp := &silaEngine.SilaExecutionPayloadEnvelope{}
		silaResp := &pb.SilaPayloadCapellaWithValue{}
		upstreamErr := json.Unmarshal(jsonBlob, upstreamResp)
		silaErr := json.Unmarshal(jsonBlob, silaResp)
		assert.Equal(t, upstreamErr != nil, silaErr != nil, fmt.Sprintf("upstream and Sila unmarshaller return inconsistent errors. %v and %v", upstreamErr, silaErr))
		// Nothing to marshal if we have an error.
		if upstreamErr != nil {
			return
		}
		upstreamBlob, upstreamErr := json.Marshal(upstreamResp)
		silaBlob, silaErr := json.Marshal(silaResp)
		assert.Equal(t, upstreamErr != nil, silaErr != nil, "upstream and Sila unmarshaller return inconsistent errors")
		newUpstreamResp := &silaEngine.SilaExecutionPayloadEnvelope{}
		newUpstreamErr := json.Unmarshal(silaBlob, newUpstreamResp)
		assert.NoError(t, newUpstreamErr)
		newUpstreamResp2 := &silaEngine.SilaExecutionPayloadEnvelope{}
		newUpstreamErr = json.Unmarshal(upstreamBlob, newUpstreamResp2)
		assert.NoError(t, newUpstreamErr)

		assert.DeepEqual(t, newUpstreamResp, newUpstreamResp2)
	})
}

func FuzzSilaBlock(f *testing.F) {
	f.Skip("Is skipped until false positive rate can be resolved.")
	logsBloom := [256]byte{'j', 'u', 'n', 'k'}
	addr := common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87")
	innerData := &types.DynamicFeeTx{
		ChainID:   big.NewInt(math.MaxInt),
		Nonce:     math.MaxUint64,
		GasTipCap: big.NewInt(math.MaxInt),
		GasFeeCap: big.NewInt(math.MaxInt),
		Gas:       math.MaxUint64,
		To:        &addr,
		Value:     big.NewInt(math.MaxInt),
		Data:      []byte{'r', 'a', 'n', 'd', 'o', 'm'},

		// Signature values
		V: big.NewInt(0),
		R: big.NewInt(math.MaxInt),
		S: big.NewInt(math.MaxInt),
	}
	tx := types.NewTx(innerData)
	execBlock := &pb.SilaBlock{
		Header: types.Header{
			ParentHash:  common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
			Root:        common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
			ReceiptHash: common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
			Bloom:       types.Bloom(logsBloom),
			Number:      big.NewInt(math.MaxInt),
			GasLimit:    math.MaxUint64,
			GasUsed:     math.MaxUint64,
			Time:        100,
			Extra:       nil,
			BaseFee:     big.NewInt(math.MaxInt),
			Difficulty:  big.NewInt(math.MaxInt),
		},
		Hash:            common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
		TotalDifficulty: "999999999999999999999999999999999999999",
		Transactions:    []*types.Transaction{tx, tx},
	}
	output, err := json.Marshal(execBlock)
	assert.NoError(f, err)

	f.Add(output)

	f.Fuzz(func(t *testing.T, jsonBlob []byte) {
		upstreamResp := make(map[string]any)
		silaResp := &pb.SilaBlock{}
		upstreamErr := json.Unmarshal(jsonBlob, &upstreamResp)
		silaErr := json.Unmarshal(jsonBlob, silaResp)
		// Nothing to marshal if we have an error.
		if upstreamErr != nil || silaErr != nil {
			return
		}
		// Exit early if fuzzer is inserting bogus hashes in.
		if isBogusTransactionHash(silaResp, upstreamResp) {
			return
		}
		// Exit early if fuzzer provides bogus fields.
		valid, err := jsonFieldsAreValid(silaResp, upstreamResp)
		assert.NoError(t, err)
		if !valid {
			return
		}
		assert.NoError(t, validateBlockConsistency(silaResp, upstreamResp))

		upstreamBlob, upstreamErr := json.Marshal(upstreamResp)
		silaBlob, silaErr := json.Marshal(silaResp)
		assert.Equal(t, upstreamErr != nil, silaErr != nil, "upstream and Sila unmarshaller return inconsistent errors")
		newUpstreamResp := make(map[string]any)
		newUpstreamErr := json.Unmarshal(silaBlob, &newUpstreamResp)
		assert.NoError(t, newUpstreamErr)
		newUpstreamResp2 := make(map[string]any)
		newUpstreamErr = json.Unmarshal(upstreamBlob, &newUpstreamResp2)
		assert.NoError(t, newUpstreamErr)

		assert.DeepEqual(t, newUpstreamResp, newUpstreamResp2)
		compareHeaders(t, jsonBlob)
	})
}

func isBogusTransactionHash(blk *pb.SilaBlock, jsonMap map[string]any) bool {
	if blk.Transactions == nil {
		return false
	}

	for i, tx := range blk.Transactions {
		jsonTx, ok := jsonMap["transactions"].([]any)[i].(map[string]any)
		if !ok {
			return true
		}
		// Fuzzer removed hash field.
		if _, ok := jsonTx["hash"]; !ok {
			return true
		}
		if tx.Hash().String() != jsonTx["hash"].(string) {
			return true
		}
	}
	return false
}

func compareHeaders(t *testing.T, jsonBlob []byte) {
	upstreamResp := &types.Header{}
	silaResp := &pb.SilaBlock{}
	upstreamErr := json.Unmarshal(jsonBlob, upstreamResp)
	silaErr := json.Unmarshal(jsonBlob, silaResp)
	assert.Equal(t, upstreamErr != nil, silaErr != nil, fmt.Sprintf("upstream and Sila unmarshaller return inconsistent errors. %v and %v", upstreamErr, silaErr))
	// Nothing to marshal if we have an error.
	if upstreamErr != nil {
		return
	}

	upstreamBlob, upstreamErr := json.Marshal(upstreamResp)
	silaBlob, silaErr := json.Marshal(silaResp.Header)
	assert.Equal(t, upstreamErr != nil, silaErr != nil, "upstream and Sila unmarshaller return inconsistent errors")
	newUpstreamResp := &types.Header{}
	newUpstreamErr := json.Unmarshal(silaBlob, newUpstreamResp)
	assert.NoError(t, newUpstreamErr)
	newUpstreamResp2 := &types.Header{}
	newUpstreamErr = json.Unmarshal(upstreamBlob, newUpstreamResp2)
	assert.NoError(t, newUpstreamErr)

	assert.DeepEqual(t, newUpstreamResp, newUpstreamResp2)
}

func validateBlockConsistency(execBlock *pb.SilaBlock, jsonMap map[string]any) error {
	blockVal := reflect.ValueOf(execBlock).Elem()
	bType := reflect.TypeFor[pb.SilaBlock]()

	fieldnum := bType.NumField()

	for i := range fieldnum {
		field := bType.Field(i)
		fName := field.Tag.Get("json")
		if field.Name == "Header" {
			continue
		}
		if fName == "" {
			return errors.Errorf("Field %s had no json tag", field.Name)
		}
		fVal, ok := jsonMap[fName]
		if !ok {
			return errors.Errorf("%s doesn't exist in json map for field %s", fName, field.Name)
		}
		jsonVal := fVal
		bVal := blockVal.Field(i).Interface()
		if field.Name == "Hash" {
			jsonVal = common.HexToHash(jsonVal.(string))
		}
		if field.Name == "Transactions" {
			continue
		}
		if !reflect.DeepEqual(jsonVal, bVal) {
			return errors.Errorf("fields don't match, %v and %v are not equal for field %s", jsonVal, bVal, field.Name)
		}
	}
	return nil
}

func jsonFieldsAreValid(execBlock *pb.SilaBlock, jsonMap map[string]any) (bool, error) {
	bType := reflect.TypeFor[pb.SilaBlock]()

	fieldnum := bType.NumField()

	for i := range fieldnum {
		field := bType.Field(i)
		fName := field.Tag.Get("json")
		if field.Name == "Header" {
			continue
		}
		if fName == "" {
			return false, errors.Errorf("Field %s had no json tag", field.Name)
		}
		_, ok := jsonMap[fName]
		if !ok {
			return false, nil
		}
	}
	return true, nil
}
