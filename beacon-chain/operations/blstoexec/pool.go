package blstoexec

import (
	"maps"
	"sync"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	doublylinkedlist "github.com/sila-chain/Sila-Consensus-Core/v7/container/doubly-linked-list"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

// We recycle the BLS changes pool to avoid the backing map growing without
// bound. The cycling operation is expensive because it copies all elements, so
// we only do it when the map is smaller than this upper bound.
const blsChangesPoolThreshold = 2000

var (
	blsToExecMessageInPoolTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "bls_to_exec_message_pool_total",
		Help: "The number of saved bls to exec messages in the operation pool.",
	})
)

// PoolManager maintains pending and seen BLS-to-Sila-change objects.
// This pool is used by proposers to insert BLS-to-Sila-change objects into new blocks.
type PoolManager interface {
	PendingBLSToExecChanges() ([]*silapb.SignedBLSToSilaChange, error)
	BLSToExecChangesForInclusion(beaconState state.ReadOnlyBeaconState) ([]*silapb.SignedBLSToSilaChange, error)
	InsertBLSToExecChange(change *silapb.SignedBLSToSilaChange)
	MarkIncluded(change *silapb.SignedBLSToSilaChange)
	ValidatorExists(idx primitives.ValidatorIndex) bool
}

// Pool is a concrete implementation of PoolManager.
type Pool struct {
	lock    sync.RWMutex
	pending doublylinkedlist.List[*silapb.SignedBLSToSilaChange]
	m       map[primitives.ValidatorIndex]*doublylinkedlist.Node[*silapb.SignedBLSToSilaChange]
}

// NewPool returns an initialized pool.
func NewPool() *Pool {
	return &Pool{
		pending: doublylinkedlist.List[*silapb.SignedBLSToSilaChange]{},
		m:       make(map[primitives.ValidatorIndex]*doublylinkedlist.Node[*silapb.SignedBLSToSilaChange]),
	}
}

// Copies the internal map and returns a new one.
func (p *Pool) cycleMap() {
	newMap := make(map[primitives.ValidatorIndex]*doublylinkedlist.Node[*silapb.SignedBLSToSilaChange])
	maps.Copy(newMap, p.m)
	p.m = newMap
}

// PendingBLSToExecChanges returns all objects from the pool.
func (p *Pool) PendingBLSToExecChanges() ([]*silapb.SignedBLSToSilaChange, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	result := make([]*silapb.SignedBLSToSilaChange, p.pending.Len())
	node := p.pending.First()
	var err error
	for i := 0; node != nil; i++ {
		result[i], err = node.Value()
		if err != nil {
			return nil, err
		}
		node, err = node.Next()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// BLSToExecChangesForInclusion returns objects that are ready for inclusion.
// This method will not return more than the block enforced MaxBlsToSilaChanges.
func (p *Pool) BLSToExecChangesForInclusion(st state.ReadOnlyBeaconState) ([]*silapb.SignedBLSToSilaChange, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	length := int(min(float64(params.BeaconConfig().MaxBlsToSilaChanges), float64(p.pending.Len())))
	result := make([]*silapb.SignedBLSToSilaChange, 0, length)
	node := p.pending.Last()
	for node != nil && len(result) < length {
		change, err := node.Value()
		if err != nil {
			return nil, err
		}
		_, err = blocks.ValidateBLSToSilaChange(st, change)
		if err != nil {
			logrus.WithError(err).Warning("removing invalid BLSToSilaChange from pool")
			// MarkIncluded removes the invalid change from the pool
			p.lock.RUnlock()
			p.MarkIncluded(change)
			p.lock.RLock()
		} else {
			result = append(result, change)
		}
		node, err = node.Prev()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// InsertBLSToExecChange inserts an object into the pool.
func (p *Pool) InsertBLSToExecChange(change *silapb.SignedBLSToSilaChange) {
	p.lock.Lock()
	defer p.lock.Unlock()

	_, exists := p.m[change.Message.ValidatorIndex]
	if exists {
		return
	}

	p.pending.Append(doublylinkedlist.NewNode(change))
	p.m[change.Message.ValidatorIndex] = p.pending.Last()

	blsToExecMessageInPoolTotal.Inc()
}

// MarkIncluded is used when an object has been included in a beacon block. Every block seen by this
// node should call this method to include the object. This will remove the object from the pool.
func (p *Pool) MarkIncluded(change *silapb.SignedBLSToSilaChange) {
	p.lock.Lock()
	defer p.lock.Unlock()

	node := p.m[change.Message.ValidatorIndex]
	if node == nil {
		return
	}

	delete(p.m, change.Message.ValidatorIndex)
	p.pending.Remove(node)
	if p.numPending() == blsChangesPoolThreshold {
		p.cycleMap()
	}

	blsToExecMessageInPoolTotal.Dec()
}

// ValidatorExists checks if the bls to Sila change object exists
// for that particular validator.
func (p *Pool) ValidatorExists(idx primitives.ValidatorIndex) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	node := p.m[idx]

	return node != nil
}

// numPending returns the number of pending bls to Sila changes in the pool
func (p *Pool) numPending() int {
	return p.pending.Len()
}
