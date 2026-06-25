package slashings

import (
	"context"
	"sync"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/startup"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// PoolInserter is capable of inserting new slashing objects into the operations pool.
type PoolInserter interface {
	InsertAttesterSlashing(
		ctx context.Context,
		state state.ReadOnlyBeaconState,
		slashing silapb.AttSlashing,
	) error
	InsertProposerSlashing(
		ctx context.Context,
		state state.ReadOnlyBeaconState,
		slashing *silapb.ProposerSlashing,
	) error
}

// PoolManager maintains a pool of pending and recently included attester and proposer slashings.
// This pool is used by proposers to insert data into new blocks.
type PoolManager interface {
	PoolInserter
	PendingAttesterSlashings(ctx context.Context, state state.ReadOnlyBeaconState, noLimit bool) []silapb.AttSlashing
	PendingProposerSlashings(ctx context.Context, state state.ReadOnlyBeaconState, noLimit bool) []*silapb.ProposerSlashing
	MarkIncludedAttesterSlashing(as silapb.AttSlashing)
	MarkIncludedProposerSlashing(ps *silapb.ProposerSlashing)
	ConvertToElectra()
}

// Option for pool service configuration.
type Option func(p *PoolService) error

// PoolService manages the Pool.
type PoolService struct {
	ctx             context.Context
	cancel          context.CancelFunc
	poolManager     PoolManager
	currentSlotFn   func() primitives.Slot
	cw              startup.ClockWaiter
	clock           *startup.Clock
	runElectraTimer bool
}

// Pool is a concrete implementation of PoolManager.
type Pool struct {
	lock                    sync.RWMutex
	pendingProposerSlashing []*silapb.ProposerSlashing
	pendingAttesterSlashing []*PendingAttesterSlashing
	included                map[primitives.ValidatorIndex]bool
}

// PendingAttesterSlashing represents an attester slashing in the operation pool.
// Allows for easy binary searching of included validator indexes.
type PendingAttesterSlashing struct {
	attesterSlashing silapb.AttSlashing
	validatorToSlash primitives.ValidatorIndex
}
