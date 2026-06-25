package beacon

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/pagination"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db/filters"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/attestations"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stategen"
	"github.com/sila-chain/Sila-Consensus-Core/v7/cmd"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/features"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1/attestation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// sortableAttestations implements the Sort interface to sort attestations
// by slot as the canonical sorting attribute.
type sortableAttestations []silapb.Att

// Len is the number of elements in the collection.
func (s sortableAttestations) Len() int { return len(s) }

// Swap swaps the elements with indexes i and j.
func (s sortableAttestations) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Less reports whether the element with index i must sort before the element with index j.
func (s sortableAttestations) Less(i, j int) bool {
	return s[i].GetData().Slot < s[j].GetData().Slot
}

func mapAttestationsByTargetRoot(atts []silapb.Att) map[[32]byte][]silapb.Att {
	attsMap := make(map[[32]byte][]silapb.Att, len(atts))
	if len(atts) == 0 {
		return attsMap
	}
	for _, att := range atts {
		attsMap[bytesutil.ToBytes32(att.GetData().Target.Root)] = append(attsMap[bytesutil.ToBytes32(att.GetData().Target.Root)], att)
	}
	return attsMap
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// ListAttestations retrieves attestations by block root, slot, or epoch.
// Attestations are sorted by data slot by default.
//
// The server may return an empty list when no attestations match the given
// filter criteria. This RPC should not return NOT_FOUND. Only one filter
// criteria should be used.
func (bs *Server) ListAttestations(
	ctx context.Context, req *silapb.ListAttestationsRequest,
) (*silapb.ListAttestationsResponse, error) {
	if int(req.PageSize) > cmd.Get().MaxRPCPageSize {
		return nil, status.Errorf(codes.InvalidArgument, "Requested page size %d can not be greater than max size %d",
			req.PageSize, cmd.Get().MaxRPCPageSize)
	}
	var blocks []interfaces.ReadOnlySignedBeaconBlock
	var err error
	switch q := req.QueryFilter.(type) {
	case *silapb.ListAttestationsRequest_GenesisEpoch:
		blocks, _, err = bs.BeaconDB.Blocks(ctx, filters.NewFilter().SetStartEpoch(0).SetEndEpoch(0))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not fetch attestations: %v", err)
		}
	case *silapb.ListAttestationsRequest_Epoch:
		if q.Epoch >= params.BeaconConfig().ElectraForkEpoch {
			return &silapb.ListAttestationsResponse{
				Attestations:  make([]*silapb.Attestation, 0),
				TotalSize:     int32(0),
				NextPageToken: strconv.Itoa(0),
			}, nil
		}
		blocks, _, err = bs.BeaconDB.Blocks(ctx, filters.NewFilter().SetStartEpoch(q.Epoch).SetEndEpoch(q.Epoch))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not fetch attestations: %v", err)
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "Must specify a filter criteria for fetching attestations")
	}

	atts, err := blockAttestations[*silapb.Attestation](blocks)
	if err != nil {
		return nil, err
	}

	// If there are no attestations, we simply return a response specifying this.
	// Otherwise, attempting to paginate 0 attestations below would result in an error.
	if len(atts) == 0 {
		return &silapb.ListAttestationsResponse{
			Attestations:  make([]*silapb.Attestation, 0),
			TotalSize:     int32(0),
			NextPageToken: strconv.Itoa(0),
		}, nil
	}

	start, end, nextPageToken, err := pagination.StartAndEndPage(req.PageToken, int(req.PageSize), len(atts))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not paginate attestations: %v", err)
	}

	return &silapb.ListAttestationsResponse{
		Attestations:  atts[start:end],
		TotalSize:     int32(len(atts)),
		NextPageToken: nextPageToken,
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// ListAttestationsElectra retrieves attestations by block root, slot, or epoch.
// Attestations are sorted by data slot by default.
//
// The server may return an empty list when no attestations match the given
// filter criteria. This RPC should not return NOT_FOUND. Only one filter
// criteria should be used.
func (bs *Server) ListAttestationsElectra(ctx context.Context, req *silapb.ListAttestationsRequest) (*silapb.ListAttestationsElectraResponse, error) {
	if int(req.PageSize) > cmd.Get().MaxRPCPageSize {
		return nil, status.Errorf(codes.InvalidArgument, "Requested page size %d can not be greater than max size %d",
			req.PageSize, cmd.Get().MaxRPCPageSize)
	}
	var blocks []interfaces.ReadOnlySignedBeaconBlock
	var err error
	switch q := req.QueryFilter.(type) {
	case *silapb.ListAttestationsRequest_GenesisEpoch:
		return &silapb.ListAttestationsElectraResponse{
			Attestations:  make([]*silapb.AttestationElectra, 0),
			TotalSize:     int32(0),
			NextPageToken: strconv.Itoa(0),
		}, nil
	case *silapb.ListAttestationsRequest_Epoch:
		if q.Epoch < params.BeaconConfig().ElectraForkEpoch {
			return &silapb.ListAttestationsElectraResponse{
				Attestations:  make([]*silapb.AttestationElectra, 0),
				TotalSize:     int32(0),
				NextPageToken: strconv.Itoa(0),
			}, nil
		}
		blocks, _, err = bs.BeaconDB.Blocks(ctx, filters.NewFilter().SetStartEpoch(q.Epoch).SetEndEpoch(q.Epoch))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not fetch attestations: %v", err)
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "Must specify a filter criteria for fetching attestations")
	}

	atts, err := blockAttestations[*silapb.AttestationElectra](blocks)
	if err != nil {
		return nil, err
	}

	// If there are no attestations, we simply return a response specifying this.
	// Otherwise, attempting to paginate 0 attestations below would result in an error.
	if len(atts) == 0 {
		return &silapb.ListAttestationsElectraResponse{
			Attestations:  make([]*silapb.AttestationElectra, 0),
			TotalSize:     int32(0),
			NextPageToken: strconv.Itoa(0),
		}, nil
	}

	start, end, nextPageToken, err := pagination.StartAndEndPage(req.PageToken, int(req.PageSize), len(atts))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not paginate attestations: %v", err)
	}

	return &silapb.ListAttestationsElectraResponse{
		Attestations:  atts[start:end],
		TotalSize:     int32(len(atts)),
		NextPageToken: nextPageToken,
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// ListIndexedAttestations retrieves indexed attestations by block root.
// IndexedAttestationsForEpoch are sorted by data slot by default. Start-end epoch
// filter is used to retrieve blocks with.
//
// The server may return an empty list when no attestations match the given
// filter criteria. This RPC should not return NOT_FOUND.
func (bs *Server) ListIndexedAttestations(
	ctx context.Context, req *silapb.ListIndexedAttestationsRequest,
) (*silapb.ListIndexedAttestationsResponse, error) {
	var blocks []interfaces.ReadOnlySignedBeaconBlock
	var err error
	switch q := req.QueryFilter.(type) {
	case *silapb.ListIndexedAttestationsRequest_GenesisEpoch:
		blocks, _, err = bs.BeaconDB.Blocks(ctx, filters.NewFilter().SetStartEpoch(0).SetEndEpoch(0))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not fetch attestations: %v", err)
		}
	case *silapb.ListIndexedAttestationsRequest_Epoch:
		if q.Epoch >= params.BeaconConfig().ElectraForkEpoch {
			return &silapb.ListIndexedAttestationsResponse{
				IndexedAttestations: make([]*silapb.IndexedAttestation, 0),
				TotalSize:           int32(0),
				NextPageToken:       strconv.Itoa(0),
			}, nil
		}
		blocks, _, err = bs.BeaconDB.Blocks(ctx, filters.NewFilter().SetStartEpoch(q.Epoch).SetEndEpoch(q.Epoch))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not fetch attestations: %v", err)
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "Must specify a filter criteria for fetching attestations")
	}

	indexedAtts, err := blockIndexedAttestations[*silapb.IndexedAttestation](ctx, blocks, bs.StateGen)
	if err != nil {
		return nil, err
	}

	// If there are no attestations, we simply return a response specifying this.
	// Otherwise, attempting to paginate 0 attestations below would result in an error.
	if len(indexedAtts) == 0 {
		return &silapb.ListIndexedAttestationsResponse{
			IndexedAttestations: make([]*silapb.IndexedAttestation, 0),
			TotalSize:           int32(0),
			NextPageToken:       strconv.Itoa(0),
		}, nil
	}

	start, end, nextPageToken, err := pagination.StartAndEndPage(req.PageToken, int(req.PageSize), len(indexedAtts))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not paginate attestations: %v", err)
	}

	return &silapb.ListIndexedAttestationsResponse{
		IndexedAttestations: indexedAtts[start:end],
		TotalSize:           int32(len(indexedAtts)),
		NextPageToken:       nextPageToken,
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// ListIndexedAttestationsElectra retrieves indexed attestations by block root.
// IndexedAttestationsForEpoch are sorted by data slot by default. Start-end epoch
// filter is used to retrieve blocks with.
//
// The server may return an empty list when no attestations match the given
// filter criteria. This RPC should not return NOT_FOUND.
func (bs *Server) ListIndexedAttestationsElectra(
	ctx context.Context,
	req *silapb.ListIndexedAttestationsRequest,
) (*silapb.ListIndexedAttestationsElectraResponse, error) {
	var blocks []interfaces.ReadOnlySignedBeaconBlock
	var err error
	switch q := req.QueryFilter.(type) {
	case *silapb.ListIndexedAttestationsRequest_GenesisEpoch:
		return &silapb.ListIndexedAttestationsElectraResponse{
			IndexedAttestations: make([]*silapb.IndexedAttestationElectra, 0),
			TotalSize:           int32(0),
			NextPageToken:       strconv.Itoa(0),
		}, nil
	case *silapb.ListIndexedAttestationsRequest_Epoch:
		if q.Epoch < params.BeaconConfig().ElectraForkEpoch {
			return &silapb.ListIndexedAttestationsElectraResponse{
				IndexedAttestations: make([]*silapb.IndexedAttestationElectra, 0),
				TotalSize:           int32(0),
				NextPageToken:       strconv.Itoa(0),
			}, nil
		}
		blocks, _, err = bs.BeaconDB.Blocks(ctx, filters.NewFilter().SetStartEpoch(q.Epoch).SetEndEpoch(q.Epoch))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not fetch attestations: %v", err)
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "Must specify a filter criteria for fetching attestations")
	}

	indexedAtts, err := blockIndexedAttestations[*silapb.IndexedAttestationElectra](ctx, blocks, bs.StateGen)
	if err != nil {
		return nil, err
	}
	// If there are no attestations, we simply return a response specifying this.
	// Otherwise, attempting to paginate 0 attestations below would result in an error.
	if len(indexedAtts) == 0 {
		return &silapb.ListIndexedAttestationsElectraResponse{
			IndexedAttestations: make([]*silapb.IndexedAttestationElectra, 0),
			TotalSize:           int32(0),
			NextPageToken:       strconv.Itoa(0),
		}, nil
	}

	start, end, nextPageToken, err := pagination.StartAndEndPage(req.PageToken, int(req.PageSize), len(indexedAtts))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not paginate attestations: %v", err)
	}

	return &silapb.ListIndexedAttestationsElectraResponse{
		IndexedAttestations: indexedAtts[start:end],
		TotalSize:           int32(len(indexedAtts)),
		NextPageToken:       nextPageToken,
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
//
// AttestationPool retrieves pending attestations.
//
// The server returns a list of attestations that have been seen but not
// yet processed. Pool attestations eventually expire as the slot
// advances, so an attestation missing from this request does not imply
// that it was included in a block. The attestation may have expired.
// Refer to the Sila consensus specification for more details on how
// attestations are processed and when they are no longer valid.
// https://github.com/sila-chain/Sila-Consensus-Specs/blob/master/specs/phase0/beacon-chain.md#attestations
func (bs *Server) AttestationPool(_ context.Context, req *silapb.AttestationPoolRequest) (*silapb.AttestationPoolResponse, error) {
	var atts []*silapb.Attestation
	var err error

	if features.Get().EnableExperimentalAttestationPool {
		atts, err = attestationsFromCache[*silapb.Attestation](req.PageSize, bs.AttestationCache)
	} else {
		atts, err = attestationsFromPool[*silapb.Attestation](req.PageSize, bs.AttestationsPool)
	}
	if err != nil {
		return nil, err
	}
	// If there are no attestations, we simply return a response specifying this.
	// Otherwise, attempting to paginate 0 attestations below would result in an error.
	if len(atts) == 0 {
		return &silapb.AttestationPoolResponse{
			Attestations:  make([]*silapb.Attestation, 0),
			TotalSize:     int32(0),
			NextPageToken: strconv.Itoa(0),
		}, nil
	}

	start, end, nextPageToken, err := pagination.StartAndEndPage(req.PageToken, int(req.PageSize), len(atts))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not paginate attestations: %v", err)
	}

	return &silapb.AttestationPoolResponse{
		Attestations:  atts[start:end],
		TotalSize:     int32(len(atts)),
		NextPageToken: nextPageToken,
	}, nil
}

// Deprecated: The gRPC API will remain the default and fully supported through v8 (expected in 2026) but will be eventually removed in favor of REST API.
func (bs *Server) AttestationPoolElectra(_ context.Context, req *silapb.AttestationPoolRequest) (*silapb.AttestationPoolElectraResponse, error) {
	var atts []*silapb.AttestationElectra
	var err error

	if features.Get().EnableExperimentalAttestationPool {
		atts, err = attestationsFromCache[*silapb.AttestationElectra](req.PageSize, bs.AttestationCache)
	} else {
		atts, err = attestationsFromPool[*silapb.AttestationElectra](req.PageSize, bs.AttestationsPool)
	}
	if err != nil {
		return nil, err
	}

	// If there are no attestations, we simply return a response specifying this.
	// Otherwise, attempting to paginate 0 attestations below would result in an error.
	if len(atts) == 0 {
		return &silapb.AttestationPoolElectraResponse{
			Attestations:  make([]*silapb.AttestationElectra, 0),
			TotalSize:     int32(0),
			NextPageToken: strconv.Itoa(0),
		}, nil
	}

	start, end, nextPageToken, err := pagination.StartAndEndPage(req.PageToken, int(req.PageSize), len(atts))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not paginate attestations: %v", err)
	}

	return &silapb.AttestationPoolElectraResponse{
		Attestations:  atts[start:end],
		TotalSize:     int32(len(atts)),
		NextPageToken: nextPageToken,
	}, nil
}

func blockAttestations[T silapb.Att](blocks []interfaces.ReadOnlySignedBeaconBlock) ([]T, error) {
	blockAtts := make([]silapb.Att, 0, params.BeaconConfig().MaxAttestations*uint64(len(blocks)))
	for _, blk := range blocks {
		blockAtts = append(blockAtts, blk.Block().Body().Attestations()...)
	}
	// We sort attestations according to the Sortable interface.
	sort.Sort(sortableAttestations(blockAtts))
	numAttestations := len(blockAtts)
	if numAttestations == 0 {
		return []T{}, nil
	}

	atts := make([]T, 0, len(blockAtts))
	for _, att := range blockAtts {
		a, ok := att.(T)
		if !ok {
			var expected T
			return nil, status.Errorf(codes.Internal, "Attestation is of the wrong type (expected %T, got %T)", expected, att)
		}
		atts = append(atts, a)
	}

	return atts, nil
}

func blockIndexedAttestations[T silapb.IndexedAtt](
	ctx context.Context,
	blocks []interfaces.ReadOnlySignedBeaconBlock,
	stateGen stategen.StateManager,
) ([]T, error) {
	attsArray := make([]silapb.Att, 0, params.BeaconConfig().MaxAttestations*uint64(len(blocks)))
	for _, b := range blocks {
		attsArray = append(attsArray, b.Block().Body().Attestations()...)
	}

	// We sort attestations according to the Sortable interface.
	sort.Sort(sortableAttestations(attsArray))
	numAttestations := len(attsArray)
	if numAttestations == 0 {
		return []T{}, nil
	}

	// We use the retrieved committees for the b root to convert all attestations
	// into indexed form effectively.
	mappedAttestations := mapAttestationsByTargetRoot(attsArray)
	indexed := make([]T, 0, numAttestations)
	for targetRoot, atts := range mappedAttestations {
		attState, err := stateGen.StateByRootNoCopy(ctx, targetRoot)
		if err != nil && strings.Contains(err.Error(), "unknown state summary") {
			// We shouldn't stop the request if we encounter an attestation we don't have the state for.
			log.Debugf("Could not get state for attestation target root %#x", targetRoot)
			continue
		} else if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"Could not retrieve state for attestation target root %#x: %v",
				targetRoot,
				err,
			)
		}
		for i := range atts {
			att := atts[i]
			committee, err := helpers.BeaconCommitteeFromState(ctx, attState, att.GetData().Slot, att.GetData().CommitteeIndex)
			if err != nil {
				return nil, status.Errorf(
					codes.Internal,
					"Could not retrieve committee from state %v",
					err,
				)
			}
			idxAtt, err := attestation.ConvertToIndexed(ctx, att, committee)
			if err != nil {
				return nil, err
			}
			a, ok := idxAtt.(T)
			if !ok {
				var expected T
				return nil, status.Errorf(codes.Internal, "Indexed attestation is of the wrong type (expected %T, got %T)", expected, idxAtt)
			}
			indexed = append(indexed, a)
		}
	}

	return indexed, nil
}

func attestationsFromPool[T silapb.Att](pageSize int32, pool attestations.Pool) ([]T, error) {
	if int(pageSize) > cmd.Get().MaxRPCPageSize {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Requested page size %d can not be greater than max size %d",
			pageSize,
			cmd.Get().MaxRPCPageSize,
		)
	}
	poolAtts := pool.AggregatedAttestations()
	atts := make([]T, 0, len(poolAtts))
	for _, att := range poolAtts {
		a, ok := att.(T)
		if !ok {
			var expected T
			return nil, status.Errorf(codes.Internal, "Attestation is of the wrong type (expected %T, got %T)", expected, att)
		}
		atts = append(atts, a)
	}
	return atts, nil
}

func attestationsFromCache[T silapb.Att](pageSize int32, c *cache.AttestationCache) ([]T, error) {
	if int(pageSize) > cmd.Get().MaxRPCPageSize {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Requested page size %d can not be greater than max size %d",
			pageSize,
			cmd.Get().MaxRPCPageSize,
		)
	}
	cacheAtts := c.GetAll()
	atts := make([]T, 0, len(cacheAtts))
	for _, att := range cacheAtts {
		a, ok := att.(T)
		if !ok {
			var expected T
			return nil, status.Errorf(codes.Internal, "Attestation is of the wrong type (expected %T, got %T)", expected, att)
		}
		atts = append(atts, a)
	}
	return atts, nil
}
