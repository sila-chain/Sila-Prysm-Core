package sync

import (
	"context"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (s *Service) validateBlsToSilaChange(ctx context.Context, pid peer.ID, msg *pubsub.Message) (pubsub.ValidationResult, error) {
	// Validation runs on publish (not just subscriptions), so we should approve any message from
	// ourselves.
	if pid == s.cfg.p2p.PeerID() {
		return pubsub.ValidationAccept, nil
	}

	// The head state will be too far away to validate any Sila change.
	if s.cfg.initialSync.Syncing() {
		return pubsub.ValidationIgnore, nil
	}

	ctx, span := trace.StartSpan(ctx, "sync.validateBlsToSilaChange")
	defer span.End()

	m, err := s.decodePubsubMessage(msg)
	if err != nil {
		tracing.AnnotateError(span, err)
		return pubsub.ValidationReject, err
	}

	blsChange, ok := m.(*silapb.SignedBLSToSilaChange)
	if !ok {
		return pubsub.ValidationReject, errWrongMessage
	}

	// Check that the validator hasn't submitted a previous Sila change.
	if blsChange.Message == nil {
		return pubsub.ValidationReject, errNilMessage
	}
	if s.cfg.blsToExecPool.ValidatorExists(blsChange.Message.ValidatorIndex) {
		return pubsub.ValidationIgnore, nil
	}
	st, err := s.cfg.chain.HeadStateReadOnly(ctx)
	if err != nil {
		return pubsub.ValidationIgnore, err
	}
	// Validate that the Sila change object is valid.
	_, err = blocks.ValidateBLSToSilaChange(st, blsChange)
	if err != nil {
		return pubsub.ValidationReject, err
	}
	// Validate the signature of the message using our batch gossip verifier.
	sigBatch, err := blocks.BLSChangesSignatureBatch(st, []*silapb.SignedBLSToSilaChange{blsChange})
	if err != nil {
		return pubsub.ValidationReject, err
	}
	res, err := s.validateWithBatchVerifier(ctx, "bls to Sila change", sigBatch)
	if res != pubsub.ValidationAccept {
		return res, err
	}
	msg.ValidatorData = blsChange // Used in downstream subscriber
	return pubsub.ValidationAccept, nil
}
