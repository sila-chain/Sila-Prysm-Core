package state_native

import (
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
)

// DepositBalanceToConsume is a non-mutating call to the beacon state which returns the value of the
// deposit balance to consume field. This method requires access to the RLock on the state and only
// applies in electra or later.
func (b *BeaconState) DepositBalanceToConsume() (primitives.Gwei, error) {
	if b.version < version.Electra {
		return 0, errNotSupported("DepositBalanceToConsume", b.version)
	}
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.depositBalanceToConsume, nil
}

// PendingDeposits is a non-mutating call to the beacon state which returns a deep copy of
// the pending balance deposit slice. This method requires access to the RLock on the state and
// only applies in electra or later.
func (b *BeaconState) PendingDeposits() ([]*silapb.PendingDeposit, error) {
	if b.version < version.Electra {
		return nil, errNotSupported("PendingDeposits", b.version)
	}
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.pendingDepositsVal(), nil
}

// IsPendingValidator checks the state's pending_deposits queue under RLock; the underlying
// slice is not copied.
func (b *BeaconState) IsPendingValidator(pubkey []byte) (bool, error) {
	if b.version < version.Electra {
		return false, errNotSupported("IsPendingValidator", b.version)
	}
	b.lock.RLock()
	defer b.lock.RUnlock()
	return helpers.IsPendingValidator(b.pendingDeposits, pubkey)
}

func (b *BeaconState) pendingDepositsVal() []*silapb.PendingDeposit {
	if b.pendingDeposits == nil {
		return nil
	}

	return silapb.CopySlice(b.pendingDeposits)
}
