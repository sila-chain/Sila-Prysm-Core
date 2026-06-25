package cache

import (
	"sync"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

type executionPayloadBidKey struct {
	slot       primitives.Slot
	parentHash [32]byte
	parentRoot [32]byte
}

// HighestExecutionPayloadBidCache stores the highest bid for each
// (slot, parent_block_hash, parent_block_root) tuple.
type HighestExecutionPayloadBidCache struct {
	bids map[executionPayloadBidKey]*silapb.SignedExecutionPayloadBid
	lock sync.RWMutex
}

// NewHighestExecutionPayloadBidCache initializes a highest-bid cache.
func NewHighestExecutionPayloadBidCache() *HighestExecutionPayloadBidCache {
	return &HighestExecutionPayloadBidCache{
		bids: make(map[executionPayloadBidKey]*silapb.SignedExecutionPayloadBid),
	}
}

// Get returns the highest cached bid for the given tuple.
func (c *HighestExecutionPayloadBidCache) Get(
	slot primitives.Slot,
	parentHash [32]byte,
	parentRoot [32]byte,
) (*silapb.SignedExecutionPayloadBid, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	bid, ok := c.bids[executionPayloadBidKey{
		slot:       slot,
		parentHash: parentHash,
		parentRoot: parentRoot,
	}]
	return bid, ok
}

// SetIfHigher inserts the bid if absent, or replaces the cached bid only if
// the incoming value is strictly greater.
func (c *HighestExecutionPayloadBidCache) SetIfHigher(bid *silapb.SignedExecutionPayloadBid) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := executionPayloadBidKey{
		slot:       bid.Message.Slot,
		parentHash: [32]byte(bid.Message.ParentBlockHash),
		parentRoot: [32]byte(bid.Message.ParentBlockRoot),
	}
	cached, ok := c.bids[key]
	if !ok || bid.Message.Value > cached.Message.Value {
		c.bids[key] = bid
		return true
	}
	return false
}

// PruneBefore removes all cached bids for slots before the provided slot.
func (c *HighestExecutionPayloadBidCache) PruneBefore(slot primitives.Slot) {
	c.lock.Lock()
	defer c.lock.Unlock()

	for key := range c.bids {
		if key.slot < slot {
			delete(c.bids, key)
		}
	}
}
