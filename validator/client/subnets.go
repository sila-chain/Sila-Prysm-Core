package client

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

// subscribeToSubnets iterates through each validator duty, signs each slot, and asks beacon node
// to eagerly subscribe to subnets so that the aggregator has attestations to aggregate.
func (v *validator) subscribeToSubnets(ctx context.Context, duties *silapb.ValidatorDutiesContainer) error {
	ctx, span := trace.StartSpan(ctx, "validator.subscribeToSubnets")
	defer span.End()

	subscribeSlots := make([]primitives.Slot, 0, len(duties.CurrentEpochDuties)+len(duties.NextEpochDuties))
	subscribeCommitteeIndices := make([]primitives.CommitteeIndex, 0, len(duties.CurrentEpochDuties)+len(duties.NextEpochDuties))
	subscribeIsAggregator := make([]bool, 0, len(duties.CurrentEpochDuties)+len(duties.NextEpochDuties))
	activeDuties := make([]*silapb.ValidatorDuty, 0, len(duties.CurrentEpochDuties)+len(duties.NextEpochDuties))
	alreadySubscribed := make(map[[64]byte]bool)

	if err := v.aggSelector.RefreshSelectionProofs(ctx); err != nil {
		return errors.Wrap(err, "could not prepare aggregated selection proofs")
	}

	for _, duty := range duties.CurrentEpochDuties {
		pk := bytesutil.ToBytes48(duty.PublicKey)
		if duty.Status == silapb.ValidatorStatus_ACTIVE || duty.Status == silapb.ValidatorStatus_EXITING {
			attesterSlot := duty.AttesterSlot
			committeeIndex := duty.CommitteeIndex
			alreadySubscribedKey := validatorSubnetSubscriptionKey(attesterSlot, committeeIndex)
			if _, ok := alreadySubscribed[alreadySubscribedKey]; ok {
				continue
			}

			aggregator, err := v.isAggregator(ctx, duty.CommitteeLength, attesterSlot, pk)
			if err != nil {
				return errors.Wrap(err, "could not check if a validator is an aggregator")
			}
			if aggregator {
				alreadySubscribed[alreadySubscribedKey] = true
			}

			subscribeSlots = append(subscribeSlots, attesterSlot)
			subscribeCommitteeIndices = append(subscribeCommitteeIndices, committeeIndex)
			subscribeIsAggregator = append(subscribeIsAggregator, aggregator)
			activeDuties = append(activeDuties, duty)
		}
	}

	for _, duty := range duties.NextEpochDuties {
		if duty.Status == silapb.ValidatorStatus_ACTIVE || duty.Status == silapb.ValidatorStatus_EXITING {
			attesterSlot := duty.AttesterSlot
			committeeIndex := duty.CommitteeIndex
			alreadySubscribedKey := validatorSubnetSubscriptionKey(attesterSlot, committeeIndex)
			if _, ok := alreadySubscribed[alreadySubscribedKey]; ok {
				continue
			}

			aggregator, err := v.isAggregator(ctx, duty.CommitteeLength, attesterSlot, bytesutil.ToBytes48(duty.PublicKey))
			if err != nil {
				return errors.Wrap(err, "could not check if a validator is an aggregator")
			}
			if aggregator {
				alreadySubscribed[alreadySubscribedKey] = true
			}

			subscribeSlots = append(subscribeSlots, attesterSlot)
			subscribeCommitteeIndices = append(subscribeCommitteeIndices, committeeIndex)
			subscribeIsAggregator = append(subscribeIsAggregator, aggregator)
			activeDuties = append(activeDuties, duty)
		}
	}

	_, err := v.validatorClient.SubscribeCommitteeSubnets(ctx,
		&silapb.CommitteeSubnetsSubscribeRequest{
			Slots:        subscribeSlots,
			CommitteeIds: subscribeCommitteeIndices,
			IsAggregator: subscribeIsAggregator,
		},
		activeDuties,
	)

	return err
}

func validatorSubnetSubscriptionKey(slot primitives.Slot, committeeIndex primitives.CommitteeIndex) [64]byte {
	return bytesutil.ToBytes64(append(bytesutil.Bytes32(uint64(slot)), bytesutil.Bytes32(uint64(committeeIndex))...))
}
