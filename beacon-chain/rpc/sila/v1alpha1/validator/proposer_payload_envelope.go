package validator

import (
	"context"
	"fmt"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/peerdas"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	consensusblocks "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// storeSilaPayloadEnvelope creates and caches the sila payload envelope
// after the block is fully built (state root set). If postBlockState is non-nil,
func (vs *Server) storeSilaPayloadEnvelope(
	sBlk interfaces.SignedBeaconBlock,
	local *consensusblocks.GetPayloadResponse,
) error {
	blockRoot, err := sBlk.Block().HashTreeRoot()
	if err != nil {
		return errors.Wrap(err, "could not compute block hash tree root")
	}

	payload := extractSilaPayloadGloas(local)

	parentRoot := sBlk.Block().ParentRoot()
	envelope := &silapb.SilaPayloadEnvelope{
		Payload:               payload,
		ExecutionRequests:     local.ExecutionRequests,
		BuilderIndex:          params.BeaconConfig().BuilderIndexSelfBuild,
		BeaconBlockRoot:       blockRoot[:],
		ParentBeaconBlockRoot: parentRoot[:],
	}

	// Precompute sidecars here (during ProposeBeaconBlock slack) so publish stays fast.
	var roSidecars []consensusblocks.RODataColumn
	if bundle := local.BlobsBundler; bundle != nil && len(bundle.GetBlobs()) > 0 {
		cellsPerBlob, proofsPerBlob, err := peerdas.ComputeCellsAndProofsFromFlat(bundle.GetBlobs(), bundle.GetProofs())
		if err != nil {
			return errors.Wrap(err, "compute cells and proofs from blobs bundle")
		}
		roSidecars, err = peerdas.DataColumnSidecarsGloas(cellsPerBlob, proofsPerBlob, sBlk.Block().Slot(), blockRoot)
		if err != nil {
			return errors.Wrap(err, "build gloas data column sidecars")
		}
	}

	vs.SilaPayloadEnvelopeCache.Set(&cache.SilaPayloadContents{
		Envelope:    envelope,
		DataColumns: roSidecars,
	})
	return nil
}

func extractSilaPayloadGloas(local *consensusblocks.GetPayloadResponse) *silaenginev1.SilaPayloadGloas {
	if local == nil || local.ExecutionData == nil || local.ExecutionData.IsNil() {
		return nil
	}
	if p, ok := local.ExecutionData.Proto().(*silaenginev1.SilaPayloadGloas); ok {
		return p
	}
	return nil
}

// GetSilaPayloadEnvelope implements the gRPC endpoint:
// /sila/v1alpha1/validator/sila_payload_envelope/{slot}/{builder_index}
// It returns the stored sila payload envelope for a slot/builder and, for
// self-build envelopes, computes the post-payload state root on demand.
func (vs *Server) GetSilaPayloadEnvelope(
	ctx context.Context,
	req *silapb.SilaPayloadEnvelopeRequest,
) (*silapb.SilaPayloadEnvelopeResponse, error) {
	_, span := trace.StartSpan(ctx, "ProposerServer.GetSilaPayloadEnvelope")
	defer span.End()

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	span.SetAttributes(trace.StringAttribute("slot", fmt.Sprintf("%d", req.Slot)))

	if slots.ToEpoch(req.Slot) < params.BeaconConfig().GloasForkEpoch {
		return nil, status.Errorf(codes.InvalidArgument,
			"sila payload envelopes are not supported before Gloas fork (slot %d)", req.Slot)
	}

	contents, ok := vs.SilaPayloadEnvelopeCache.Contents()
	if !ok || contents.Envelope.Payload.SlotNumber != req.Slot {
		return nil, status.Errorf(codes.NotFound,
			"sila payload envelope not found for slot %d", req.Slot)
	}

	return &silapb.SilaPayloadEnvelopeResponse{
		Envelope: contents.Envelope,
	}, nil
}

// PublishSilaPayloadEnvelope validates and broadcasts a signed sila payload envelope.
// This is called by validators after signing the envelope retrieved from GetSilaPayloadEnvelope.
//
// gRPC endpoint: POST /sila/v1alpha1/validator/sila_payload_envelope
func (vs *Server) PublishSilaPayloadEnvelope(
	ctx context.Context,
	req *silapb.SignedSilaPayloadEnvelope,
) (*emptypb.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "ProposerServer.PublishSilaPayloadEnvelope")
	defer span.End()

	if req == nil || req.Message == nil || req.Message.Payload == nil {
		return nil, status.Error(codes.InvalidArgument, "signed envelope or payload cannot be nil")
	}

	envSlot := primitives.Slot(req.Message.Payload.SlotNumber)
	if slots.ToEpoch(envSlot) < params.BeaconConfig().GloasForkEpoch {
		return nil, status.Errorf(codes.InvalidArgument,
			"sila payload envelopes are not supported before Gloas fork (slot %d)", envSlot)
	}

	beaconBlockRoot := bytesutil.ToBytes32(req.Message.BeaconBlockRoot)
	span.SetAttributes(
		trace.StringAttribute("slot", fmt.Sprintf("%d", envSlot)),
		trace.StringAttribute("builderIndex", fmt.Sprintf("%d", req.Message.BuilderIndex)),
		trace.StringAttribute("beaconBlockRoot", fmt.Sprintf("%#x", beaconBlockRoot[:8])),
	)

	log := log.WithFields(logrus.Fields{
		"slot":            envSlot,
		"builderIndex":    req.Message.BuilderIndex,
		"beaconBlockRoot": fmt.Sprintf("%#x", beaconBlockRoot[:8]),
	})
	log.Info("Publishing signed sila payload envelope")

	// Broadcast sidecars BEFORE receiving the envelope so the DA check sees them.
	// Slot guard avoids broadcasting cached sidecars from an unrelated slot.
	if contents, ok := vs.SilaPayloadEnvelopeCache.Contents(); ok &&
		contents.Envelope.Payload.SlotNumber == envSlot && len(contents.DataColumns) > 0 {
		log.WithField("columns", len(contents.DataColumns)).Debug("Broadcasting Gloas data column sidecars")
		if err := vs.broadcastAndReceiveDataColumns(ctx, contents.DataColumns); err != nil {
			log.WithError(err).Error("Failed to broadcast Gloas data column sidecars")
		}
	}

	if err := vs.P2P.Broadcast(ctx, req); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to broadcast sila payload envelope: %v", err)
	}

	roSigned, err := consensusblocks.WrappedROSignedSilaPayloadEnvelope(req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not wrap signed envelope: %v", err)
	}
	if err := vs.SilaPayloadEnvelopeReceiver.ReceiveSilaPayloadEnvelope(ctx, roSigned); err != nil {
		// Broadcast already succeeded; import failed. REST maps Aborted -> 202 (beacon-APIs #580).
		return nil, status.Errorf(codes.Aborted, "failed to receive sila payload envelope: %v", err)
	}

	log.Info("Successfully published sila payload envelope")

	return &emptypb.Empty{}, nil
}

// setParentExecutionRequests populates the parent_execution_requests field
// in the block body based on the parent's sila payload envelope.
func (vs *Server) setParentExecutionRequests(ctx context.Context, sBlk interfaces.SignedBeaconBlock, head state.BeaconState, parentFull bool) error {
	if head.Version() < version.Gloas {
		return sBlk.SetParentExecutionRequests(&silaenginev1.ExecutionRequests{})
	}

	parentRoot := sBlk.Block().ParentRoot()
	parentSlot, err := vs.ForkchoiceFetcher.RecentBlockSlot(parentRoot)
	if err != nil {
		return errors.Wrap(err, "could not get parent block slot")
	}
	if slots.ToEpoch(parentSlot) < params.BeaconConfig().GloasForkEpoch || !parentFull {
		return sBlk.SetParentExecutionRequests(&silaenginev1.ExecutionRequests{})
	}

	// TODO: replace DB lookup with a single-entry cache (blockroot → envelope).
	signedEnvelope, err := vs.BeaconDB.SilaPayloadEnvelope(ctx, parentRoot)
	if err != nil {
		return errors.Wrap(err, "could not get parent sila payload envelope")
	}
	return sBlk.SetParentExecutionRequests(signedEnvelope.Message.ExecutionRequests)
}
