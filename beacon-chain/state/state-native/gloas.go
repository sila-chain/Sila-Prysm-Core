package state_native

import (
	enginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/engine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// executionPayloadAvailabilityVal returns a copy of the execution payload availability.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) executionPayloadAvailabilityVal() []byte {
	if b.executionPayloadAvailability == nil {
		return nil
	}

	availability := make([]byte, len(b.executionPayloadAvailability))
	copy(availability, b.executionPayloadAvailability)

	return availability
}

// builderPendingPaymentsVal returns a copy of the builder pending payments.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) builderPendingPaymentsVal() []*silapb.BuilderPendingPayment {
	if b.builderPendingPayments == nil {
		return nil
	}

	payments := make([]*silapb.BuilderPendingPayment, len(b.builderPendingPayments))
	for i, payment := range b.builderPendingPayments {
		payments[i] = payment.Copy()
	}

	return payments
}

// builderPendingWithdrawalsVal returns a copy of the builder pending withdrawals.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) builderPendingWithdrawalsVal() []*silapb.BuilderPendingWithdrawal {
	if b.builderPendingWithdrawals == nil {
		return nil
	}

	withdrawals := make([]*silapb.BuilderPendingWithdrawal, len(b.builderPendingWithdrawals))
	for i, withdrawal := range b.builderPendingWithdrawals {
		withdrawals[i] = withdrawal.Copy()
	}

	return withdrawals
}

// buildersVal returns a copy of the builders registry.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) buildersVal() []*silapb.Builder {
	if b.builders == nil {
		return nil
	}

	builders := make([]*silapb.Builder, len(b.builders))
	for i := range builders {
		builder := b.builders[i]
		builders[i] = silapb.CopyBuilder(builder)
	}

	return builders
}

// latestBlockHashVal returns a copy of the latest block hash.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) latestBlockHashVal() []byte {
	if b.latestBlockHash == nil {
		return nil
	}

	hash := make([]byte, len(b.latestBlockHash))
	copy(hash, b.latestBlockHash)

	return hash
}

// payloadExpectedWithdrawalsVal returns a copy of the payload expected withdrawals.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) payloadExpectedWithdrawalsVal() []*enginev1.Withdrawal {
	if b.payloadExpectedWithdrawals == nil {
		return nil
	}

	withdrawals := make([]*enginev1.Withdrawal, len(b.payloadExpectedWithdrawals))
	for i, withdrawal := range b.payloadExpectedWithdrawals {
		withdrawals[i] = withdrawal.Copy()
	}

	return withdrawals
}

// ptcWindowVal returns a copy of the cached PTC window.
// This assumes that a lock is already held on BeaconState.
func (b *BeaconState) ptcWindowVal() []*silapb.PTCs {
	return silapb.CopyPTCWindow(b.ptcWindow)
}
