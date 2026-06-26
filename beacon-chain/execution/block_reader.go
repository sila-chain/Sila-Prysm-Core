package execution

import (
	"context"
	"fmt"
	"math/big"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/execution/types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	"github.com/sila-chain/Sila/common"
	"github.com/pkg/errors"
)

// searchThreshold to apply for when searching for blocks of a particular time. If the buffer
// is exceeded we recalibrate the search again.
const searchThreshold = 5

// amount of times we repeat a failed search till is satisfies the conditional.
const repeatedSearches = 2 * searchThreshold

var errBlockTimeTooLate = errors.New("provided time is later than the current silaexec head")

// BlockExists returns true if the block exists, its height and any possible error encountered.
func (s *Service) BlockExists(ctx context.Context, hash common.Hash) (bool, *big.Int, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.BlockExists")
	defer span.End()

	if exists, hdrInfo, err := s.headerCache.HeaderInfoByHash(hash); exists || err != nil {
		if err != nil {
			return false, nil, err
		}
		span.SetAttributes(trace.BoolAttribute("blockCacheHit", true))
		return true, hdrInfo.Number, nil
	}
	span.SetAttributes(trace.BoolAttribute("blockCacheHit", false))
	header, err := s.HeaderByHash(ctx, hash)
	if err != nil {
		return false, big.NewInt(0), errors.Wrap(err, "could not query block with given hash")
	}

	if err := s.headerCache.AddHeader(header); err != nil {
		return false, big.NewInt(0), err
	}

	return true, new(big.Int).Set(header.Number), nil
}

// BlockHashByHeight returns the block hash of the block at the given height.
func (s *Service) BlockHashByHeight(ctx context.Context, height *big.Int) (common.Hash, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.BlockHashByHeight")
	defer span.End()

	if exists, hInfo, err := s.headerCache.HeaderInfoByHeight(height); exists || err != nil {
		if err != nil {
			return [32]byte{}, err
		}
		span.SetAttributes(trace.BoolAttribute("headerCacheHit", true))
		return hInfo.Hash, nil
	}
	span.SetAttributes(trace.BoolAttribute("headerCacheHit", false))

	if s.rpcClient == nil {
		err := errors.New("nil rpc client")
		tracing.AnnotateError(span, err)
		return [32]byte{}, err
	}

	header, err := s.HeaderByNumber(ctx, height)
	if err != nil {
		return [32]byte{}, errors.Wrap(err, fmt.Sprintf("could not query header with height %d", height.Uint64()))
	}
	if err := s.headerCache.AddHeader(header); err != nil {
		return [32]byte{}, err
	}
	return header.Hash, nil
}

// BlockTimeByHeight fetches an silaexec block timestamp by its height.
func (s *Service) BlockTimeByHeight(ctx context.Context, height *big.Int) (uint64, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.BlockTimeByHeight")
	defer span.End()
	if s.rpcClient == nil {
		err := errors.New("nil rpc client")
		tracing.AnnotateError(span, err)
		return 0, err
	}

	header, err := s.HeaderByNumber(ctx, height)
	if err != nil {
		return 0, errors.Wrap(err, fmt.Sprintf("could not query block with height %d", height.Uint64()))
	}
	return header.Time, nil
}

// BlockByTimestamp returns the most recent block number up to a given timestamp.
// This is an optimized version with the worst case being O(2*repeatedSearches) number of calls
// while in best case search for the block is performed in O(1).
func (s *Service) BlockByTimestamp(ctx context.Context, time uint64) (*types.HeaderInfo, error) {
	ctx, span := trace.StartSpan(ctx, "powchain.BlockByTimestamp")
	defer span.End()

	s.latestSilaDataLock.RLock()
	latestBlkHeight := s.latestSilaData.BlockHeight
	latestBlkTime := s.latestSilaData.BlockTime
	s.latestSilaDataLock.RUnlock()

	if time > latestBlkTime {
		return nil, errors.Wrap(errBlockTimeTooLate, fmt.Sprintf("(%d > %d)", time, latestBlkTime))
	}
	// Initialize a pointer to silaexec chain's history to start our search from.
	cursorNum := new(big.Int).SetUint64(latestBlkHeight)
	cursorTime := latestBlkTime

	var numOfBlocks uint64
	estimatedBlk := cursorNum.Uint64()
	maxTimeBuffer := searchThreshold * params.BeaconConfig().SecondsPerSilaBlock
	// Terminate if we can't find an acceptable block after
	// repeated searches.
	for range repeatedSearches {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if time > cursorTime+maxTimeBuffer {
			numOfBlocks = (time - cursorTime) / params.BeaconConfig().SecondsPerSilaBlock
			// In the event we have an infeasible estimated block, this is a defensive
			// check to ensure it does not exceed rational bounds.
			if cursorNum.Uint64()+numOfBlocks > latestBlkHeight {
				break
			}
			estimatedBlk = cursorNum.Uint64() + numOfBlocks
		} else if time+maxTimeBuffer < cursorTime {
			numOfBlocks = (cursorTime - time) / params.BeaconConfig().SecondsPerSilaBlock
			// In the event we have an infeasible number of blocks
			// we exit early.
			if numOfBlocks >= cursorNum.Uint64() {
				break
			}
			estimatedBlk = cursorNum.Uint64() - numOfBlocks
		} else {
			// Exit if we are in the range of
			// time - buffer <= head.time <= time + buffer
			break
		}
		hInfo, err := s.retrieveHeaderInfo(ctx, estimatedBlk)
		if err != nil {
			return nil, err
		}
		cursorNum = hInfo.Number
		cursorTime = hInfo.Time
	}

	// Exit early if we get the desired block.
	if cursorTime == time {
		return s.retrieveHeaderInfo(ctx, cursorNum.Uint64())
	}
	if cursorTime > time {
		return s.findMaxTargetSilaBlock(ctx, new(big.Int).SetUint64(estimatedBlk), time)
	}
	return s.findMinTargetSilaBlock(ctx, new(big.Int).SetUint64(estimatedBlk), time)
}

// Performs a search to find a target silaexec block which is earlier than or equal to the
// target time. This method is used when head.time > targetTime
func (s *Service) findMaxTargetSilaBlock(ctx context.Context, upperBoundBlk *big.Int, targetTime uint64) (*types.HeaderInfo, error) {
	for bn := upperBoundBlk; ; bn = new(big.Int).Sub(bn, big.NewInt(1)) {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		info, err := s.retrieveHeaderInfo(ctx, bn.Uint64())
		if err != nil {
			return nil, err
		}
		if info.Time <= targetTime {
			return info, nil
		}
	}
}

// Performs a search to find a target silaexec block which is just earlier than or equal to the
// target time. This method is used when head.time < targetTime
func (s *Service) findMinTargetSilaBlock(ctx context.Context, lowerBoundBlk *big.Int, targetTime uint64) (*types.HeaderInfo, error) {
	for bn := lowerBoundBlk; ; bn = new(big.Int).Add(bn, big.NewInt(1)) {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		info, err := s.retrieveHeaderInfo(ctx, bn.Uint64())
		if err != nil {
			return nil, err
		}
		// Return the last block before we hit the threshold time.
		if info.Time > targetTime {
			return s.retrieveHeaderInfo(ctx, info.Number.Uint64()-1)
		}
		// If time is equal, this is our target block.
		if info.Time == targetTime {
			return info, nil
		}
	}
}

func (s *Service) retrieveHeaderInfo(ctx context.Context, bNum uint64) (*types.HeaderInfo, error) {
	bn := new(big.Int).SetUint64(bNum)
	exists, info, err := s.headerCache.HeaderInfoByHeight(bn)
	if err != nil {
		return nil, err
	}
	if !exists {
		hdr, err := s.HeaderByNumber(ctx, bn)
		if err != nil {
			return nil, err
		}
		if hdr == nil {
			return nil, errors.Errorf("header with the number %d does not exist", bNum)
		}
		if err := s.headerCache.AddHeader(hdr); err != nil {
			return nil, err
		}
		info = hdr
	}
	return info, nil
}
