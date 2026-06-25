package validator

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/async/event"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed"
	blockfeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/block"
	statefeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// StreamBlocksAltair to clients every single time a block is received by the beacon node.
func (vs *Server) StreamBlocksAltair(req *silapb.StreamBlocksRequest, stream silapb.BeaconNodeValidator_StreamBlocksAltairServer) error {
	blocksChannel := make(chan *feed.Event, 1)
	var blockSub event.Subscription
	if req.VerifiedOnly {
		blockSub = vs.StateNotifier.StateFeed().Subscribe(blocksChannel)
	} else {
		blockSub = vs.BlockNotifier.BlockFeed().Subscribe(blocksChannel)
	}
	defer blockSub.Unsubscribe()

	for {
		select {
		case blockEvent := <-blocksChannel:
			if req.VerifiedOnly {
				if err := sendVerifiedBlocks(stream, blockEvent); err != nil {
					return err
				}
			} else {
				if err := vs.sendBlocks(stream, blockEvent); err != nil {
					return err
				}
			}
		case <-blockSub.Err():
			return status.Error(codes.Aborted, "Subscriber closed, exiting goroutine")
		case <-vs.Ctx.Done():
			return status.Error(codes.Canceled, "Context canceled")
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Context canceled")
		}
	}
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// StreamSlots sends a the block's slot and dependent roots to clients every single time a block is received by the beacon node.
func (vs *Server) StreamSlots(req *silapb.StreamSlotsRequest, stream silapb.BeaconNodeValidator_StreamSlotsServer) error {
	ch := make(chan *feed.Event, 1)
	var sub event.Subscription
	if req.VerifiedOnly {
		sub = vs.StateNotifier.StateFeed().Subscribe(ch)
	} else {
		sub = vs.BlockNotifier.BlockFeed().Subscribe(ch)
	}
	defer sub.Unsubscribe()

	for {
		select {
		case ev := <-ch:
			var s primitives.Slot
			var currDependentRoot, prevDependentRoot [32]byte
			if req.VerifiedOnly {
				if ev.Type != statefeed.BlockProcessed {
					continue
				}
				data, ok := ev.Data.(*statefeed.BlockProcessedData)
				if !ok || data == nil {
					continue
				}
				s = data.Slot
				currDependentRoot = data.CurrDependentRoot
				prevDependentRoot = data.PrevDependentRoot
			} else {
				if ev.Type != blockfeed.ReceivedBlock {
					continue
				}
				data, ok := ev.Data.(*blockfeed.ReceivedBlockData)
				if !ok || data == nil {
					continue
				}
				s = data.SignedBlock.Block().Slot()
				currDependentRoot = data.CurrDependentRoot
				prevDependentRoot = data.PrevDependentRoot
			}
			if err := stream.Send(
				&silapb.StreamSlotsResponse{
					Slot:                      s,
					PreviousDutyDependentRoot: prevDependentRoot[:],
					CurrentDutyDependentRoot:  currDependentRoot[:],
				}); err != nil {
				return status.Errorf(codes.Unavailable, "Could not send over stream: %v", err)
			}
		case <-sub.Err():
			return status.Error(codes.Aborted, "Subscriber closed, exiting goroutine")
		case <-vs.Ctx.Done():
			return status.Error(codes.Canceled, "Context canceled")
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Context canceled")
		}
	}
}

func sendVerifiedBlocks(stream silapb.BeaconNodeValidator_StreamBlocksAltairServer, blockEvent *feed.Event) error {
	if blockEvent.Type != statefeed.BlockProcessed {
		return nil
	}
	data, ok := blockEvent.Data.(*statefeed.BlockProcessedData)
	if !ok || data == nil {
		return nil
	}
	b := &silapb.StreamBlocksResponse{}
	switch data.SignedBlock.Version() {
	case version.Phase0:
		pb, err := data.SignedBlock.Proto()
		if err != nil {
			return errors.Wrap(err, "could not get protobuf block")
		}
		phBlk, ok := pb.(*silapb.SignedBeaconBlock)
		if !ok {
			log.Warn("Mismatch between version and block type, was expecting SignedBeaconBlock")
			return nil
		}
		b.Block = &silapb.StreamBlocksResponse_Phase0Block{Phase0Block: phBlk}
	case version.Altair:
		pb, err := data.SignedBlock.Proto()
		if err != nil {
			return errors.Wrap(err, "could not get protobuf block")
		}
		phBlk, ok := pb.(*silapb.SignedBeaconBlockAltair)
		if !ok {
			log.Warn("Mismatch between version and block type, was expecting SignedBeaconBlockAltair")
			return nil
		}
		b.Block = &silapb.StreamBlocksResponse_AltairBlock{AltairBlock: phBlk}
	case version.Bellatrix:
		pb, err := data.SignedBlock.Proto()
		if err != nil {
			return errors.Wrap(err, "could not get protobuf block")
		}
		phBlk, ok := pb.(*silapb.SignedBeaconBlockBellatrix)
		if !ok {
			log.Warn("Mismatch between version and block type, was expecting SignedBeaconBlockBellatrix")
			return nil
		}
		b.Block = &silapb.StreamBlocksResponse_BellatrixBlock{BellatrixBlock: phBlk}
	case version.Capella:
		pb, err := data.SignedBlock.Proto()
		if err != nil {
			return errors.Wrap(err, "could not get protobuf block")
		}
		phBlk, ok := pb.(*silapb.SignedBeaconBlockCapella)
		if !ok {
			log.Warn("Mismatch between version and block type, was expecting SignedBeaconBlockCapella")
			return nil
		}
		b.Block = &silapb.StreamBlocksResponse_CapellaBlock{CapellaBlock: phBlk}
	case version.Deneb:
		pb, err := data.SignedBlock.Proto()
		if err != nil {
			return errors.Wrap(err, "could not get protobuf block")
		}
		phBlk, ok := pb.(*silapb.SignedBeaconBlockDeneb)
		if !ok {
			log.Warn("Mismatch between version and block type, was expecting SignedBeaconBlockDeneb")
			return nil
		}
		b.Block = &silapb.StreamBlocksResponse_DenebBlock{DenebBlock: phBlk}
	case version.Electra:
		pb, err := data.SignedBlock.Proto()
		if err != nil {
			return errors.Wrap(err, "could not get protobuf block")
		}
		phBlk, ok := pb.(*silapb.SignedBeaconBlockElectra)
		if !ok {
			log.Warn("Mismatch between version and block type, was expecting SignedBeaconBlockElectra")
			return nil
		}
		b.Block = &silapb.StreamBlocksResponse_ElectraBlock{ElectraBlock: phBlk}
	case version.Fulu:
		pb, err := data.SignedBlock.Proto()
		if err != nil {
			return errors.Wrap(err, "could not get protobuf block")
		}
		phBlk, ok := pb.(*silapb.SignedBeaconBlockFulu)
		if !ok {
			log.Warn("Mismatch between version and block type, was expecting SignedBeaconBlockFulu")
			return nil
		}
		b.Block = &silapb.StreamBlocksResponse_FuluBlock{FuluBlock: phBlk}
	}

	if err := stream.Send(b); err != nil {
		return status.Errorf(codes.Unavailable, "Could not send over stream: %v", err)
	}

	return nil
}

func (vs *Server) sendBlocks(stream silapb.BeaconNodeValidator_StreamBlocksAltairServer, blockEvent *feed.Event) error {
	if blockEvent.Type != blockfeed.ReceivedBlock {
		return nil
	}

	data, ok := blockEvent.Data.(*blockfeed.ReceivedBlockData)
	if !ok || data == nil {
		// Got bad data over the stream.
		return nil
	}
	if data.SignedBlock == nil {
		// One nil block shouldn't stop the stream.
		return nil
	}
	log := log.WithField("blockSlot", data.SignedBlock.Block().Slot())
	headState, err := vs.HeadFetcher.HeadStateReadOnly(vs.Ctx)
	if err != nil {
		log.WithError(err).Error("Could not get head state")
		return nil
	}
	signed := data.SignedBlock
	sig := signed.Signature()
	if err := blocks.VerifyBlockSignature(headState, signed.Block().ProposerIndex(), sig[:], signed.Block().HashTreeRoot); err != nil {
		log.WithError(err).Error("Could not verify block signature")
		return nil
	}
	b := &silapb.StreamBlocksResponse{}
	pb, err := data.SignedBlock.Proto()
	if err != nil {
		return errors.Wrap(err, "could not get protobuf block")
	}
	switch p := pb.(type) {
	case *silapb.SignedBeaconBlock:
		b.Block = &silapb.StreamBlocksResponse_Phase0Block{Phase0Block: p}
	case *silapb.SignedBeaconBlockAltair:
		b.Block = &silapb.StreamBlocksResponse_AltairBlock{AltairBlock: p}
	case *silapb.SignedBeaconBlockBellatrix:
		b.Block = &silapb.StreamBlocksResponse_BellatrixBlock{BellatrixBlock: p}
	case *silapb.SignedBeaconBlockCapella:
		b.Block = &silapb.StreamBlocksResponse_CapellaBlock{CapellaBlock: p}
	case *silapb.SignedBeaconBlockDeneb:
		b.Block = &silapb.StreamBlocksResponse_DenebBlock{DenebBlock: p}
	case *silapb.SignedBeaconBlockElectra:
		b.Block = &silapb.StreamBlocksResponse_ElectraBlock{ElectraBlock: p}
	case *silapb.SignedBeaconBlockFulu:
		b.Block = &silapb.StreamBlocksResponse_FuluBlock{FuluBlock: p}
	default:
		log.Errorf("Unknown block type %T", p)
	}
	if err := stream.Send(b); err != nil {
		return status.Errorf(codes.Unavailable, "Could not send over stream: %v", err)
	}

	return nil
}
