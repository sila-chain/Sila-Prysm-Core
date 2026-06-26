package testing

import (
	"context"
	"math/big"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/peerdas"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	payloadattribute "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/payload-attribute"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	pb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila/common"
	"github.com/sila-chain/Sila/common/hexutil"
	"github.com/holiman/uint256"
	"github.com/pkg/errors"
)

// SilaEngineClient --
type SilaEngineClient struct {
	NewPayloadResp              []byte
	PayloadIDBytes              *pb.PayloadIDBytes
	ForkChoiceUpdatedResp       []byte
	SilaBlock              *pb.SilaBlock
	Err                         error
	ErrLatestExecBlock          error
	ErrExecBlockByHash          error
	ErrForkchoiceUpdated        error
	ErrNewPayload               error
	SilaPayloadByBlockHash map[[32]byte]*pb.SilaPayload
	SlotByBlockHash             map[[32]byte]primitives.Slot
	BlockByHashMap              map[[32]byte]*pb.SilaBlock
	NumReconstructedPayloads    uint64
	TerminalBlockHash           []byte
	TerminalBlockHashExists     bool
	OverrideValidHash           [32]byte
	GetPayloadResponse          *blocks.GetPayloadResponse
	ErrGetPayload               error
	BlobSidecars                []blocks.VerifiedROBlob
	ErrorBlobSidecars           error
	DataColumnSidecars          []blocks.VerifiedRODataColumn
	ErrorDataColumnSidecars     error
	ClientVersion               []*structs.ClientVersionV1
	ErrorClientVersion          error
}

// NewPayload --
func (e *SilaEngineClient) NewPayload(_ context.Context, _ interfaces.SilaData, _ []common.Hash, _ *common.Hash, _ *pb.SilaRequests) ([]byte, error) {
	return e.NewPayloadResp, e.ErrNewPayload
}

// ForkchoiceUpdated --
func (e *SilaEngineClient) ForkchoiceUpdated(
	_ context.Context, fcs *pb.ForkchoiceState, _ payloadattribute.Attributer,
) (*pb.PayloadIDBytes, []byte, error) {
	if e.OverrideValidHash != [32]byte{} && bytesutil.ToBytes32(fcs.HeadBlockHash) == e.OverrideValidHash {
		return e.PayloadIDBytes, e.ForkChoiceUpdatedResp, nil
	}
	return e.PayloadIDBytes, e.ForkChoiceUpdatedResp, e.ErrForkchoiceUpdated
}

// GetPayload --
func (e *SilaEngineClient) GetPayload(_ context.Context, _ [8]byte, _ primitives.Slot) (*blocks.GetPayloadResponse, error) {
	return e.GetPayloadResponse, e.ErrGetPayload
}

// LatestSilaBlock --
func (e *SilaEngineClient) LatestSilaBlock(_ context.Context) (*pb.SilaBlock, error) {
	return e.SilaBlock, e.ErrLatestExecBlock
}

// SilaBlockByHash --
func (e *SilaEngineClient) SilaBlockByHash(_ context.Context, h common.Hash, _ bool) (*pb.SilaBlock, error) {
	b, ok := e.BlockByHashMap[h]
	if !ok {
		return nil, errors.New("block not found")
	}
	return b, e.ErrExecBlockByHash
}

// ReconstructFullBlock --
func (e *SilaEngineClient) ReconstructFullBlock(
	_ context.Context, blindedBlock interfaces.ReadOnlySignedBeaconBlock,
) (interfaces.SignedBeaconBlock, error) {
	if !blindedBlock.Block().IsBlinded() {
		return nil, errors.New("block must be blinded")
	}
	header, err := blindedBlock.Block().Body().Execution()
	if err != nil {
		return nil, err
	}
	payload, ok := e.SilaPayloadByBlockHash[bytesutil.ToBytes32(header.BlockHash())]
	if !ok {
		return nil, errors.New("block not found")
	}
	e.NumReconstructedPayloads++
	return blocks.BuildSignedBeaconBlockFromSilaPayload(blindedBlock, payload)
}

// ReconstructFullBellatrixBlockBatch --
func (e *SilaEngineClient) ReconstructFullBellatrixBlockBatch(
	ctx context.Context, blindedBlocks []interfaces.ReadOnlySignedBeaconBlock,
) ([]interfaces.SignedBeaconBlock, error) {
	fullBlocks := make([]interfaces.SignedBeaconBlock, 0, len(blindedBlocks))
	for _, b := range blindedBlocks {
		newBlock, err := e.ReconstructFullBlock(ctx, b)
		if err != nil {
			return nil, err
		}
		fullBlocks = append(fullBlocks, newBlock)
	}
	return fullBlocks, nil
}

// ReconstructFullGloasSilaPayloadsByHash --
func (e *SilaEngineClient) ReconstructFullGloasSilaPayloadsByHash(
	_ context.Context, blockHashes [][32]byte,
) (map[[32]byte]*pb.SilaPayloadGloas, error) {
	payloads := make(map[[32]byte]*pb.SilaPayloadGloas, len(blockHashes))
	for i := range blockHashes {
		blockHash := blockHashes[i]
		if p, ok := e.SilaPayloadByBlockHash[blockHash]; ok {
			payloads[blockHash] = &pb.SilaPayloadGloas{
				ParentHash:    p.ParentHash,
				FeeRecipient:  p.FeeRecipient,
				StateRoot:     p.StateRoot,
				ReceiptsRoot:  p.ReceiptsRoot,
				LogsBloom:     p.LogsBloom,
				PrevRandao:    p.PrevRandao,
				BlockNumber:   p.BlockNumber,
				GasLimit:      p.GasLimit,
				GasUsed:       p.GasUsed,
				Timestamp:     p.Timestamp,
				ExtraData:     p.ExtraData,
				BaseFeePerGas: p.BaseFeePerGas,
				BlockHash:     p.BlockHash,
				Transactions:  p.Transactions,
				Withdrawals:   []*pb.Withdrawal{},
				SlotNumber:    e.SlotByBlockHash[blockHash],
			}
			continue
		}
		if e.GetPayloadResponse != nil && e.GetPayloadResponse.SilaData != nil {
			if p, ok := e.GetPayloadResponse.SilaData.Proto().(*pb.SilaPayloadGloas); ok {
				payloads[blockHash] = p
				continue
			}
		}
		return nil, errors.New("payload not found")
	}
	return payloads, nil
}

// ReconstructBlobSidecars is a mock implementation of the ReconstructBlobSidecars method.
func (e *SilaEngineClient) ReconstructBlobSidecars(context.Context, interfaces.ReadOnlySignedBeaconBlock, [fieldparams.RootLength]byte, func(uint64) bool) ([]blocks.VerifiedROBlob, error) {
	return e.BlobSidecars, e.ErrorBlobSidecars
}

// ConstructDataColumnSidecars is a mock implementation of the ConstructDataColumnSidecars method.
func (e *SilaEngineClient) ConstructDataColumnSidecars(context.Context, peerdas.ConstructionPopulator) ([]blocks.VerifiedRODataColumn, error) {
	return e.DataColumnSidecars, e.ErrorDataColumnSidecars
}

// ReconstructSilaPayloadEnvelope --
func (e *SilaEngineClient) ReconstructSilaPayloadEnvelope(
	_ context.Context, envelope *silapb.SignedBlindedSilaPayloadEnvelope,
) (*silapb.SignedSilaPayloadEnvelope, error) {
	if e.Err != nil {
		return nil, e.Err
	}
	payload, ok := e.SilaPayloadByBlockHash[bytesutil.ToBytes32(envelope.Message.BlockHash)]
	if !ok {
		return nil, errors.New("sila payload not found for block hash")
	}
	p := payloadToPayloadGloas(payload)
	p.SlotNumber = envelope.Message.Slot
	return &silapb.SignedSilaPayloadEnvelope{
		Message: &silapb.SilaPayloadEnvelope{
			Payload:           p,
			SilaRequests: envelope.Message.SilaRequests,
			BuilderIndex:      envelope.Message.BuilderIndex,
			BeaconBlockRoot:   envelope.Message.BeaconBlockRoot,
		},
		Signature: envelope.Signature,
	}, nil
}

func payloadToPayloadGloas(p *pb.SilaPayload) *pb.SilaPayloadGloas {
	return &pb.SilaPayloadGloas{
		ParentHash:    p.ParentHash,
		FeeRecipient:  p.FeeRecipient,
		StateRoot:     p.StateRoot,
		ReceiptsRoot:  p.ReceiptsRoot,
		LogsBloom:     p.LogsBloom,
		PrevRandao:    p.PrevRandao,
		BlockNumber:   p.BlockNumber,
		GasLimit:      p.GasLimit,
		GasUsed:       p.GasUsed,
		Timestamp:     p.Timestamp,
		ExtraData:     p.ExtraData,
		BaseFeePerGas: p.BaseFeePerGas,
		BlockHash:     p.BlockHash,
		Transactions:  p.Transactions,
	}
}

// GetTerminalBlockHash --
func (e *SilaEngineClient) GetTerminalBlockHash(ctx context.Context, transitionTime uint64) ([]byte, bool, error) {
	ttd := new(big.Int)
	ttd.SetString(params.BeaconConfig().TerminalTotalDifficulty, 10)
	terminalTotalDifficulty, overflows := uint256.FromBig(ttd)
	if overflows {
		return nil, false, errors.New("could not convert terminal total difficulty to uint256")
	}
	blk, err := e.LatestSilaBlock(ctx)
	if err != nil {
		return nil, false, errors.Wrap(err, "could not get latest sila block")
	}
	if blk == nil {
		return nil, false, errors.New("latest sila block is nil")
	}

	for {
		b, err := hexutil.DecodeBig(blk.TotalDifficulty)
		if err != nil {
			return nil, false, errors.Wrap(err, "could not convert total difficulty to uint256")
		}
		currentTotalDifficulty, _ := uint256.FromBig(b)
		blockReachedTTD := currentTotalDifficulty.Cmp(terminalTotalDifficulty) >= 0

		parentHash := blk.ParentHash
		if parentHash == params.BeaconConfig().ZeroHash {
			return nil, false, nil
		}
		parentBlk, err := e.SilaBlockByHash(ctx, parentHash, false /* with txs */)
		if err != nil {
			return nil, false, errors.Wrap(err, "could not get parent sila block")
		}
		if blockReachedTTD {
			b, err := hexutil.DecodeBig(parentBlk.TotalDifficulty)
			if err != nil {
				return nil, false, errors.Wrap(err, "could not convert total difficulty to uint256")
			}
			parentTotalDifficulty, _ := uint256.FromBig(b)
			parentReachedTTD := parentTotalDifficulty.Cmp(terminalTotalDifficulty) >= 0
			if blk.Time >= transitionTime {
				return nil, false, nil
			}
			if !parentReachedTTD {
				return blk.Hash[:], true, nil
			}
		} else {
			return nil, false, nil
		}
		blk = parentBlk
	}
}

// GetClientVersionV1 --
func (e *SilaEngineClient) GetClientVersionV1(context.Context) ([]*structs.ClientVersionV1, error) {
	return e.ClientVersion, e.ErrorClientVersion
}
