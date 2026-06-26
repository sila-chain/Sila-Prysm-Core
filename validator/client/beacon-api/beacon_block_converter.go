package beacon_api

import (
	"math/big"
	"strconv"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila/common/hexutil"
	"github.com/pkg/errors"
)

type BeaconBlockConverter interface {
	ConvertRESTPhase0BlockToProto(block *structs.BeaconBlock) (*silapb.BeaconBlock, error)
	ConvertRESTAltairBlockToProto(block *structs.BeaconBlockAltair) (*silapb.BeaconBlockAltair, error)
	ConvertRESTBellatrixBlockToProto(block *structs.BeaconBlockBellatrix) (*silapb.BeaconBlockBellatrix, error)
	ConvertRESTCapellaBlockToProto(block *structs.BeaconBlockCapella) (*silapb.BeaconBlockCapella, error)
}

type beaconApiBeaconBlockConverter struct{}

// ConvertRESTPhase0BlockToProto converts a Phase0 JSON beacon block to its protobuf equivalent
func (c beaconApiBeaconBlockConverter) ConvertRESTPhase0BlockToProto(block *structs.BeaconBlock) (*silapb.BeaconBlock, error) {
	blockSlot, err := strconv.ParseUint(block.Slot, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse slot `%s`", block.Slot)
	}

	blockProposerIndex, err := strconv.ParseUint(block.ProposerIndex, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse proposer index `%s`", block.ProposerIndex)
	}

	parentRoot, err := hexutil.Decode(block.ParentRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode parent root `%s`", block.ParentRoot)
	}

	stateRoot, err := hexutil.Decode(block.StateRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode state root `%s`", block.StateRoot)
	}

	if block.Body == nil {
		return nil, errors.New("block body is nil")
	}

	randaoReveal, err := hexutil.Decode(block.Body.RandaoReveal)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode randao reveal `%s`", block.Body.RandaoReveal)
	}

	if block.Body.SilaData == nil {
		return nil, errors.New("silaexec data is nil")
	}

	depositRoot, err := hexutil.Decode(block.Body.SilaData.DepositRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode deposit root `%s`", block.Body.SilaData.DepositRoot)
	}

	depositCount, err := strconv.ParseUint(block.Body.SilaData.DepositCount, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse deposit count `%s`", block.Body.SilaData.DepositCount)
	}

	blockHash, err := hexutil.Decode(block.Body.SilaData.BlockHash)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode block hash `%s`", block.Body.SilaData.BlockHash)
	}

	graffiti, err := hexutil.Decode(block.Body.Graffiti)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode graffiti `%s`", block.Body.Graffiti)
	}

	proposerSlashings, err := convertProposerSlashingsToProto(block.Body.ProposerSlashings)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get proposer slashings")
	}

	attesterSlashings, err := convertAttesterSlashingsToProto(block.Body.AttesterSlashings)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get attester slashings")
	}

	attestations, err := convertAttestationsToProto(block.Body.Attestations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get attestations")
	}

	deposits, err := convertDepositsToProto(block.Body.Deposits)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get deposits")
	}

	voluntaryExits, err := convertVoluntaryExitsToProto(block.Body.VoluntaryExits)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get voluntary exits")
	}

	return &silapb.BeaconBlock{
		Slot:          primitives.Slot(blockSlot),
		ProposerIndex: primitives.ValidatorIndex(blockProposerIndex),
		ParentRoot:    parentRoot,
		StateRoot:     stateRoot,
		Body: &silapb.BeaconBlockBody{
			RandaoReveal: randaoReveal,
			SilaData: &silapb.SilaData{
				DepositRoot:  depositRoot,
				DepositCount: depositCount,
				BlockHash:    blockHash,
			},
			Graffiti:          graffiti,
			ProposerSlashings: proposerSlashings,
			AttesterSlashings: attesterSlashings,
			Attestations:      attestations,
			Deposits:          deposits,
			VoluntaryExits:    voluntaryExits,
		},
	}, nil
}

// ConvertRESTAltairBlockToProto converts an Altair JSON beacon block to its protobuf equivalent
func (c beaconApiBeaconBlockConverter) ConvertRESTAltairBlockToProto(block *structs.BeaconBlockAltair) (*silapb.BeaconBlockAltair, error) {
	if block.Body == nil {
		return nil, errors.New("block body is nil")
	}

	// Call convertRESTPhase0BlockToProto to set the phase0 fields because all the error handling and the heavy lifting
	// has already been done
	phase0Block, err := c.ConvertRESTPhase0BlockToProto(&structs.BeaconBlock{
		Slot:          block.Slot,
		ProposerIndex: block.ProposerIndex,
		ParentRoot:    block.ParentRoot,
		StateRoot:     block.StateRoot,
		Body: &structs.BeaconBlockBody{
			RandaoReveal:      block.Body.RandaoReveal,
			SilaData:          block.Body.SilaData,
			Graffiti:          block.Body.Graffiti,
			ProposerSlashings: block.Body.ProposerSlashings,
			AttesterSlashings: block.Body.AttesterSlashings,
			Attestations:      block.Body.Attestations,
			Deposits:          block.Body.Deposits,
			VoluntaryExits:    block.Body.VoluntaryExits,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the phase0 fields of the altair block")
	}

	if block.Body.SyncAggregate == nil {
		return nil, errors.New("sync aggregate is nil")
	}

	syncCommitteeBits, err := hexutil.Decode(block.Body.SyncAggregate.SyncCommitteeBits)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode sync committee bits `%s`", block.Body.SyncAggregate.SyncCommitteeBits)
	}

	syncCommitteeSignature, err := hexutil.Decode(block.Body.SyncAggregate.SyncCommitteeSignature)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode sync committee signature `%s`", block.Body.SyncAggregate.SyncCommitteeSignature)
	}

	return &silapb.BeaconBlockAltair{
		Slot:          phase0Block.Slot,
		ProposerIndex: phase0Block.ProposerIndex,
		ParentRoot:    phase0Block.ParentRoot,
		StateRoot:     phase0Block.StateRoot,
		Body: &silapb.BeaconBlockBodyAltair{
			RandaoReveal:      phase0Block.Body.RandaoReveal,
			SilaData:          phase0Block.Body.SilaData,
			Graffiti:          phase0Block.Body.Graffiti,
			ProposerSlashings: phase0Block.Body.ProposerSlashings,
			AttesterSlashings: phase0Block.Body.AttesterSlashings,
			Attestations:      phase0Block.Body.Attestations,
			Deposits:          phase0Block.Body.Deposits,
			VoluntaryExits:    phase0Block.Body.VoluntaryExits,
			SyncAggregate: &silapb.SyncAggregate{
				SyncCommitteeBits:      syncCommitteeBits,
				SyncCommitteeSignature: syncCommitteeSignature,
			},
		},
	}, nil
}

// ConvertRESTBellatrixBlockToProto converts a Bellatrix JSON beacon block to its protobuf equivalent
func (c beaconApiBeaconBlockConverter) ConvertRESTBellatrixBlockToProto(block *structs.BeaconBlockBellatrix) (*silapb.BeaconBlockBellatrix, error) {
	if block.Body == nil {
		return nil, errors.New("block body is nil")
	}

	// Call convertRESTAltairBlockToProto to set the altair fields because all the error handling and the heavy lifting
	// has already been done
	altairBlock, err := c.ConvertRESTAltairBlockToProto(&structs.BeaconBlockAltair{
		Slot:          block.Slot,
		ProposerIndex: block.ProposerIndex,
		ParentRoot:    block.ParentRoot,
		StateRoot:     block.StateRoot,
		Body: &structs.BeaconBlockBodyAltair{
			RandaoReveal:      block.Body.RandaoReveal,
			SilaData:          block.Body.SilaData,
			Graffiti:          block.Body.Graffiti,
			ProposerSlashings: block.Body.ProposerSlashings,
			AttesterSlashings: block.Body.AttesterSlashings,
			Attestations:      block.Body.Attestations,
			Deposits:          block.Body.Deposits,
			VoluntaryExits:    block.Body.VoluntaryExits,
			SyncAggregate:     block.Body.SyncAggregate,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the altair fields of the bellatrix block")
	}

	if block.Body.SilaPayload == nil {
		return nil, errors.New("sila payload is nil")
	}

	parentHash, err := hexutil.Decode(block.Body.SilaPayload.ParentHash)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode sila payload parent hash `%s`", block.Body.SilaPayload.ParentHash)
	}

	feeRecipient, err := hexutil.Decode(block.Body.SilaPayload.FeeRecipient)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode sila payload fee recipient `%s`", block.Body.SilaPayload.FeeRecipient)
	}

	stateRoot, err := hexutil.Decode(block.Body.SilaPayload.StateRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode sila payload state root `%s`", block.Body.SilaPayload.StateRoot)
	}

	receiptsRoot, err := hexutil.Decode(block.Body.SilaPayload.ReceiptsRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode sila payload receipts root `%s`", block.Body.SilaPayload.ReceiptsRoot)
	}

	logsBloom, err := hexutil.Decode(block.Body.SilaPayload.LogsBloom)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode sila payload logs bloom `%s`", block.Body.SilaPayload.LogsBloom)
	}

	prevRandao, err := hexutil.Decode(block.Body.SilaPayload.PrevRandao)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode sila payload prev randao `%s`", block.Body.SilaPayload.PrevRandao)
	}

	blockNumber, err := strconv.ParseUint(block.Body.SilaPayload.BlockNumber, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse sila payload block number `%s`", block.Body.SilaPayload.BlockNumber)
	}

	gasLimit, err := strconv.ParseUint(block.Body.SilaPayload.GasLimit, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse sila payload gas limit `%s`", block.Body.SilaPayload.GasLimit)
	}

	gasUsed, err := strconv.ParseUint(block.Body.SilaPayload.GasUsed, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse sila payload gas used `%s`", block.Body.SilaPayload.GasUsed)
	}

	timestamp, err := strconv.ParseUint(block.Body.SilaPayload.Timestamp, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse sila payload timestamp `%s`", block.Body.SilaPayload.Timestamp)
	}

	extraData, err := hexutil.Decode(block.Body.SilaPayload.ExtraData)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode sila payload extra data `%s`", block.Body.SilaPayload.ExtraData)
	}

	baseFeePerGas := new(big.Int)
	if _, ok := baseFeePerGas.SetString(block.Body.SilaPayload.BaseFeePerGas, 10); !ok {
		return nil, errors.Errorf("failed to parse sila payload base fee per gas `%s`", block.Body.SilaPayload.BaseFeePerGas)
	}

	blockHash, err := hexutil.Decode(block.Body.SilaPayload.BlockHash)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode sila payload block hash `%s`", block.Body.SilaPayload.BlockHash)
	}

	transactions, err := convertTransactionsToProto(block.Body.SilaPayload.Transactions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get sila payload transactions")
	}

	return &silapb.BeaconBlockBellatrix{
		Slot:          altairBlock.Slot,
		ProposerIndex: altairBlock.ProposerIndex,
		ParentRoot:    altairBlock.ParentRoot,
		StateRoot:     altairBlock.StateRoot,
		Body: &silapb.BeaconBlockBodyBellatrix{
			RandaoReveal:      altairBlock.Body.RandaoReveal,
			SilaData:          altairBlock.Body.SilaData,
			Graffiti:          altairBlock.Body.Graffiti,
			ProposerSlashings: altairBlock.Body.ProposerSlashings,
			AttesterSlashings: altairBlock.Body.AttesterSlashings,
			Attestations:      altairBlock.Body.Attestations,
			Deposits:          altairBlock.Body.Deposits,
			VoluntaryExits:    altairBlock.Body.VoluntaryExits,
			SyncAggregate:     altairBlock.Body.SyncAggregate,
			SilaPayload: &silaenginev1.SilaPayload{
				ParentHash:    parentHash,
				FeeRecipient:  feeRecipient,
				StateRoot:     stateRoot,
				ReceiptsRoot:  receiptsRoot,
				LogsBloom:     logsBloom,
				PrevRandao:    prevRandao,
				BlockNumber:   blockNumber,
				GasLimit:      gasLimit,
				GasUsed:       gasUsed,
				Timestamp:     timestamp,
				ExtraData:     extraData,
				BaseFeePerGas: bytesutil.PadTo(bytesutil.BigIntToLittleEndianBytes(baseFeePerGas), 32),
				BlockHash:     blockHash,
				Transactions:  transactions,
			},
		},
	}, nil
}

// ConvertRESTCapellaBlockToProto converts a Capella JSON beacon block to its protobuf equivalent
func (c beaconApiBeaconBlockConverter) ConvertRESTCapellaBlockToProto(block *structs.BeaconBlockCapella) (*silapb.BeaconBlockCapella, error) {
	if block.Body == nil {
		return nil, errors.New("block body is nil")
	}

	if block.Body.SilaPayload == nil {
		return nil, errors.New("sila payload is nil")
	}

	// Call convertRESTBellatrixBlockToProto to set the bellatrix fields because all the error handling and the heavy
	// lifting has already been done
	bellatrixBlock, err := c.ConvertRESTBellatrixBlockToProto(&structs.BeaconBlockBellatrix{
		Slot:          block.Slot,
		ProposerIndex: block.ProposerIndex,
		ParentRoot:    block.ParentRoot,
		StateRoot:     block.StateRoot,
		Body: &structs.BeaconBlockBodyBellatrix{
			RandaoReveal:      block.Body.RandaoReveal,
			SilaData:          block.Body.SilaData,
			Graffiti:          block.Body.Graffiti,
			ProposerSlashings: block.Body.ProposerSlashings,
			AttesterSlashings: block.Body.AttesterSlashings,
			Attestations:      block.Body.Attestations,
			Deposits:          block.Body.Deposits,
			VoluntaryExits:    block.Body.VoluntaryExits,
			SyncAggregate:     block.Body.SyncAggregate,
			SilaPayload: &structs.SilaPayload{
				ParentHash:    block.Body.SilaPayload.ParentHash,
				FeeRecipient:  block.Body.SilaPayload.FeeRecipient,
				StateRoot:     block.Body.SilaPayload.StateRoot,
				ReceiptsRoot:  block.Body.SilaPayload.ReceiptsRoot,
				LogsBloom:     block.Body.SilaPayload.LogsBloom,
				PrevRandao:    block.Body.SilaPayload.PrevRandao,
				BlockNumber:   block.Body.SilaPayload.BlockNumber,
				GasLimit:      block.Body.SilaPayload.GasLimit,
				GasUsed:       block.Body.SilaPayload.GasUsed,
				Timestamp:     block.Body.SilaPayload.Timestamp,
				ExtraData:     block.Body.SilaPayload.ExtraData,
				BaseFeePerGas: block.Body.SilaPayload.BaseFeePerGas,
				BlockHash:     block.Body.SilaPayload.BlockHash,
				Transactions:  block.Body.SilaPayload.Transactions,
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the bellatrix fields of the capella block")
	}

	withdrawals, err := convertWithdrawalsToProto(block.Body.SilaPayload.Withdrawals)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get withdrawals")
	}

	blsToSilaChanges, err := convertBlsToSilaChangesToProto(block.Body.BLSToSilaChanges)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bls to Sila changes")
	}

	return &silapb.BeaconBlockCapella{
		Slot:          bellatrixBlock.Slot,
		ProposerIndex: bellatrixBlock.ProposerIndex,
		ParentRoot:    bellatrixBlock.ParentRoot,
		StateRoot:     bellatrixBlock.StateRoot,
		Body: &silapb.BeaconBlockBodyCapella{
			RandaoReveal:      bellatrixBlock.Body.RandaoReveal,
			SilaData:          bellatrixBlock.Body.SilaData,
			Graffiti:          bellatrixBlock.Body.Graffiti,
			ProposerSlashings: bellatrixBlock.Body.ProposerSlashings,
			AttesterSlashings: bellatrixBlock.Body.AttesterSlashings,
			Attestations:      bellatrixBlock.Body.Attestations,
			Deposits:          bellatrixBlock.Body.Deposits,
			VoluntaryExits:    bellatrixBlock.Body.VoluntaryExits,
			SyncAggregate:     bellatrixBlock.Body.SyncAggregate,
			SilaPayload: &silaenginev1.SilaPayloadCapella{
				ParentHash:    bellatrixBlock.Body.SilaPayload.ParentHash,
				FeeRecipient:  bellatrixBlock.Body.SilaPayload.FeeRecipient,
				StateRoot:     bellatrixBlock.Body.SilaPayload.StateRoot,
				ReceiptsRoot:  bellatrixBlock.Body.SilaPayload.ReceiptsRoot,
				LogsBloom:     bellatrixBlock.Body.SilaPayload.LogsBloom,
				PrevRandao:    bellatrixBlock.Body.SilaPayload.PrevRandao,
				BlockNumber:   bellatrixBlock.Body.SilaPayload.BlockNumber,
				GasLimit:      bellatrixBlock.Body.SilaPayload.GasLimit,
				GasUsed:       bellatrixBlock.Body.SilaPayload.GasUsed,
				Timestamp:     bellatrixBlock.Body.SilaPayload.Timestamp,
				ExtraData:     bellatrixBlock.Body.SilaPayload.ExtraData,
				BaseFeePerGas: bellatrixBlock.Body.SilaPayload.BaseFeePerGas,
				BlockHash:     bellatrixBlock.Body.SilaPayload.BlockHash,
				Transactions:  bellatrixBlock.Body.SilaPayload.Transactions,
				Withdrawals:   withdrawals,
			},
			BlsToSilaChanges: blsToSilaChanges,
		},
	}, nil
}
