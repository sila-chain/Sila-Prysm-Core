package blockchain

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed"
	statefeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/gloas"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	payloadattribute "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/payload-attribute"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/sila-chain/Sila/common"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// SilaPayloadEnvelopeReceiver defines the methods for receiving sila payload envelopes.
type SilaPayloadEnvelopeReceiver interface {
	ReceiveSilaPayloadEnvelope(context.Context, interfaces.ROSignedSilaPayloadEnvelope) error
}

// ReceiveSilaPayloadEnvelope processes a signed sila payload envelope for the Gloas fork.
func (s *Service) ReceiveSilaPayloadEnvelope(ctx context.Context, signed interfaces.ROSignedSilaPayloadEnvelope) (err error) {
	ctx, span := trace.StartSpan(ctx, "blockChain.ReceiveSilaPayloadEnvelope")
	defer span.End()
	start := time.Now()
	defer func() {
		beaconSilaPayloadEnvelopeProcessingDurationSeconds.Observe(time.Since(start).Seconds())
		if err != nil {
			beaconSilaPayloadEnvelopeInvalidTotal.Inc()
			return
		}
		beaconSilaPayloadEnvelopeValidTotal.Inc()
	}()

	envelope, err := signed.Envelope()
	if err != nil {
		return errors.Wrap(err, "could not get envelope")
	}
	root := envelope.BeaconBlockRoot()

	err = s.payloadBeingSynced.set(root)
	if errors.Is(err, errBlockBeingSynced) {
		log.WithField("blockRoot", fmt.Sprintf("%#x", root)).Debug("Ignoring payload envelope currently being synced")
		return nil
	}
	defer s.payloadBeingSynced.unset(root)

	blockState, err := s.getPayloadEnvelopePrestate(ctx, envelope)
	if err != nil {
		return err
	}

	var isValidPayload bool

	// EL Validation runs separately from consensus validation in different errgroup.
	elCtx, cancelEL := context.WithCancel(ctx)
	defer cancelEL()

	var elGroup errgroup.Group
	elGroup.Go(func() error {
		var elErr error
		isValidPayload, elErr = s.validateSilaPayloadOnEnvelope(elCtx, blockState, envelope)
		return elErr
	})

	// Check data availability with consensus verification.
	availGroup, availCtx := errgroup.WithContext(ctx)
	availGroup.Go(func() error {
		if err := gloas.VerifySilaPayloadEnvelope(availCtx, blockState, signed); err != nil {
			return err
		}
		s.recordPayloadArrival(root, envelope.Slot(), start)
		return nil
	})
	availGroup.Go(func() error {
		bid, err := blockState.LatestSilaPayloadBid()
		if err != nil {
			return errors.Wrap(err, "could not get latest sila payload bid")
		}
		if bid == nil || len(bid.BlobKzgCommitments()) == 0 {
			return nil
		}
		if err := s.areDataColumnsAvailable(availCtx, root, envelope.Slot()); err != nil {
			return errors.Wrap(err, "data availability check failed for payload envelope")
		}
		return nil
	})

	if err := availGroup.Wait(); err != nil {
		cancelEL()
		_ = elGroup.Wait()
		return err
	}

	// sila_payload_available is emitted when a Sila payload
	// and all data are available for payload attestation
	// without verifying the sila payload itself
	s.cfg.StateNotifier.StateFeed().Send(&feed.Event{
		Type: statefeed.SilaPayloadAvailable,
		Data: &statefeed.SilaPayloadAvailableData{
			Slot:      envelope.Slot(),
			BlockRoot: root,
		},
	})

	// Join EL validation group after firing availability event.
	if err := elGroup.Wait(); err != nil {
		return err
	}

	if err := s.savePostPayload(ctx, signed); err != nil {
		return err
	}
	if err := s.InsertPayload(envelope); err != nil {
		return errors.Wrap(err, "could not insert payload into forkchoice")
	}

	if isValidPayload {
		s.cfg.ForkChoiceStore.Lock()
		if err := s.cfg.ForkChoiceStore.SetOptimisticToValid(ctx, root); err != nil {
			log.WithError(err).Error("Could not set optimistic to valid")
		}
		s.cfg.ForkChoiceStore.Unlock()
	}

	headRootSlice, err := s.HeadRoot(ctx)
	if err != nil {
		log.WithError(err).Error("Could not get head root")
		return nil
	}
	headRoot := bytesutil.ToBytes32(headRootSlice)
	if err := s.postPayloadTasks(ctx, envelope, blockState, root, headRoot); err != nil {
		return err
	}

	// sila_payload is emitted when a Sila payload is successfully imported.
	isOptimistic, err := s.cfg.ForkChoiceStore.IsOptimistic(root)
	if err != nil {
		log.WithError(err).Error("Could not get optimistic status of block root")
		isOptimistic = false
	}

	s.cfg.StateNotifier.StateFeed().Send(&feed.Event{
		Type: statefeed.SilaPayloadProcessed,
		Data: &statefeed.SilaPayloadProcessedData{
			Slot:         envelope.Slot(),
			BuilderIndex: envelope.BuilderIndex(),
			BlockHash:    envelope.BlockHash(),
			BlockRoot:    root,
			Optimistic:   isOptimistic,
		},
	})

	execution, err := envelope.Execution()
	if err != nil {
		log.WithError(err).Error("Could not get sila payload from envelope for logging")
		return nil
	}

	log.WithFields(logrus.Fields{
		"slot":       envelope.Slot(),
		"blockRoot":  fmt.Sprintf("%#x", bytesutil.Trunc(root[:])),
		"blockHash":  fmt.Sprintf("%#x", bytesutil.Trunc(silaexec.BlockHash())),
		"parentHash": fmt.Sprintf("%#x", bytesutil.Trunc(silaexec.ParentHash())),
	}).Info("Processed sila payload envelope")
	return nil
}

func (s *Service) postPayloadTasks(ctx context.Context, envelope interfaces.ROSilaPayloadEnvelope, st state.BeaconState, root, headRoot [32]byte) error {
	if headRoot != root {
		return nil
	}
	payload, err := envelope.Execution()
	if err != nil {
		return errors.Wrap(err, "could not get sila payload from envelope")
	}
	blockHash := bytesutil.ToBytes32(payload.BlockHash())

	s.headLock.Lock()
	if s.head != nil && s.head.root == root {
		s.head.full = true
	}
	s.headLock.Unlock()

	proposingSlot := s.CurrentSlot() + 1
	attr := s.getPayloadAttribute(ctx, st, proposingSlot, headRoot[:], true)
	if !s.inRegularSync() {
		return nil
	}
	go func() {
		pid, err := s.notifyForkchoiceUpdateGloas(s.ctx, blockHash, attr)
		if err != nil {
			log.WithError(err).Error("Could not notify forkchoice update")
			return
		}
		if !attr.IsEmpty() && pid != nil {
			var pId [8]byte
			copy(pId[:], pid[:])
			s.cfg.PayloadIDCache.Set(proposingSlot, root, pId)
		}
	}()
	if requests := envelope.SilaRequests(); requests != nil && len(requests.Deposits) > 0 {
		s.prefetchDepositSignatures(requests)
	}
	return nil
}

func (s *Service) prefetchDepositSignatures(requests *silaenginev1.SilaRequests) {
	invalidIdx, err := helpers.BatchVerifyDepositRequestSignatures(s.ctx, requests.Deposits)
	if err != nil {
		log.WithError(err).Debug("Could not batch verify deposit signatures for prefetch")
		return
	}
	root, err := requests.HashTreeRoot()
	if err != nil {
		log.WithError(err).Debug("Could not hash sila requests for deposit sig prefetch")
		return
	}
	cache.DepositSig.Put(root, invalidIdx)
}

func (s *Service) getPayloadEnvelopePrestate(ctx context.Context, envelope interfaces.ROSilaPayloadEnvelope) (state.BeaconState, error) {
	ctx, span := trace.StartSpan(ctx, "blockChain.getPayloadEnvelopePrestate")
	defer span.End()

	root := envelope.BeaconBlockRoot()
	if !s.InForkchoice(root) {
		return nil, fmt.Errorf("beacon block root %#x not found in forkchoice", root)
	}
	if err := s.verifyBlkPreState(ctx, root); err != nil {
		return nil, errors.Wrap(err, "could not verify pre-state")
	}
	preState, err := s.cfg.StateGen.StateByRoot(ctx, root)
	if err != nil {
		return nil, errors.Wrap(err, "could not get pre-state by root")
	}
	if preState == nil || preState.IsNil() {
		return nil, fmt.Errorf("nil pre-state for beacon block root %#x", root)
	}
	return preState, nil
}

func (s *Service) callNewPayload(
	ctx context.Context,
	payload interfaces.SilaData,
	versionedHashes []common.Hash,
	parentRoot common.Hash,
	requests *silaenginev1.SilaRequests,
	slot primitives.Slot,
) (bool, error) {
	_, err := s.cfg.SilaEngineCaller.NewPayload(ctx, payload, versionedHashes, &parentRoot, requests)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, silaexec.ErrAcceptedSyncingPayloadStatus) {
		log.WithFields(logrus.Fields{
			"slot":             slot,
			"payloadBlockHash": fmt.Sprintf("%#x", bytesutil.Trunc(payload.BlockHash())),
		}).Info("Called new payload with optimistic envelope")
		return false, nil
	}
	if errors.Is(err, silaexec.ErrInvalidPayloadStatus) {
		return false, invalidBlock{error: ErrInvalidPayload}
	}
	return false, errors.WithMessage(ErrUndefinedSilaEngineError, err.Error())
}

func (s *Service) notifyNewEnvelopeFromBlock(ctx context.Context, b blocks.ROBlock, envelope interfaces.ROSilaPayloadEnvelope) (bool, error) {
	ctx, span := trace.StartSpan(ctx, "blockChain.notifyNewEnvelopeFromBlock")
	defer span.End()

	payload, err := envelope.Execution()
	if err != nil {
		return false, errors.Wrap(err, "could not get sila payload from envelope")
	}
	sbid, err := b.Block().Body().SignedSilaPayloadBid()
	if err != nil {
		return false, errors.Wrap(err, "could not get signed sila payload bid from block")
	}
	versionedHashes := make([]common.Hash, len(sbid.Message.BlobKzgCommitments))
	for i, c := range sbid.Message.BlobKzgCommitments {
		versionedHashes[i] = primitives.ConvertKzgCommitmentToVersionedHash(c)
	}
	return s.callNewPayload(ctx, payload, versionedHashes, common.Hash(envelope.ParentBeaconBlockRoot()), envelope.SilaRequests(), envelope.Slot())
}

// The returned boolean indicates whether the payload was valid or if it was accepted as syncing (optimistic).
func (s *Service) notifyNewEnvelope(ctx context.Context, st state.BeaconState, envelope interfaces.ROSilaPayloadEnvelope) (bool, error) {
	ctx, span := trace.StartSpan(ctx, "blockChain.notifyNewEnvelope")
	defer span.End()

	payload, err := envelope.Execution()
	if err != nil {
		return false, errors.Wrap(err, "could not get sila payload from envelope")
	}
	latestBid, err := st.LatestSilaPayloadBid()
	if err != nil {
		return false, errors.Wrap(err, "could not get latest sila payload bid")
	}
	var versionedHashes []common.Hash
	if latestBid != nil {
		commitments := latestBid.BlobKzgCommitments()
		versionedHashes = make([]common.Hash, len(commitments))
		for i, c := range commitments {
			versionedHashes[i] = primitives.ConvertKzgCommitmentToVersionedHash(c)
		}
	}
	return s.callNewPayload(ctx, payload, versionedHashes, common.Hash(envelope.ParentBeaconBlockRoot()), envelope.SilaRequests(), envelope.Slot())
}

func (s *Service) validateSilaPayloadOnEnvelope(ctx context.Context, st state.BeaconState, envelope interfaces.ROSilaPayloadEnvelope) (bool, error) {
	isValid, err := s.notifyNewEnvelope(ctx, st, envelope)
	if err == nil {
		return isValid, nil
	}

	blockRoot := envelope.BeaconBlockRoot()
	parentRoot := bytesutil.ToBytes32(st.LatestBlockHeader().ParentRoot)
	payload, payloadErr := envelope.Execution()
	if payloadErr != nil {
		return false, errors.Wrap(payloadErr, "could not get sila payload from envelope")
	}
	parentHash := bytesutil.ToBytes32(payload.ParentHash())

	s.cfg.ForkChoiceStore.Lock()
	defer s.cfg.ForkChoiceStore.Unlock()
	return false, s.handleInvalidSilaPayloadError(ctx, err, blockRoot, parentRoot, parentHash)
}

func (s *Service) savePostPayload(ctx context.Context, signed interfaces.ROSignedSilaPayloadEnvelope) error {
	ctx, span := trace.StartSpan(ctx, "blockChain.savePostPayload")
	defer span.End()

	protoEnv, ok := signed.Proto().(*silapb.SignedSilaPayloadEnvelope)
	if !ok {
		return errors.New("could not type assert signed envelope to proto")
	}
	return s.cfg.BeaconDB.SaveSilaPayloadEnvelope(ctx, protoEnv)
}

func (s *Service) recordPayloadArrival(root [32]byte, slot primitives.Slot, arrivedAt time.Time) {
	slotStart, err := slots.StartTime(s.genesisTime, slot)
	if err != nil {
		return
	}
	cfg := params.BeaconConfig()
	due := slotStart.Add(cfg.SlotComponentDuration(cfg.PayloadDueBPS))
	s.payloadArrivals.record(root, slot, arrivedAt.Before(due))
}

// PayloadEarly reports whether the payload for root arrived early; second return is false when unknown.
func (s *Service) PayloadEarly(root [32]byte) (bool, bool) {
	return s.payloadArrivals.isEarly(root)
}

// notifyForkchoiceUpdateGloas takes the block hash directly because Gloas
// blocks don't carry a Sila payload in the body.
func (s *Service) notifyForkchoiceUpdateGloas(ctx context.Context, blockHash [32]byte, attributes payloadattribute.Attributer) (*silaenginev1.PayloadIDBytes, error) {
	ctx, span := trace.StartSpan(ctx, "blockChain.notifyForkchoiceUpdateGloas")
	defer span.End()

	s.cfg.ForkChoiceStore.RLock()
	finalizedHash := s.cfg.ForkChoiceStore.FinalizedPayloadBlockHash()
	justifiedHash := s.cfg.ForkChoiceStore.UnrealizedJustifiedPayloadBlockHash()
	s.cfg.ForkChoiceStore.RUnlock()
	fcs := &silaenginev1.ForkchoiceState{
		HeadBlockHash:      blockHash[:],
		SafeBlockHash:      justifiedHash[:],
		FinalizedBlockHash: finalizedHash[:],
	}
	if attributes == nil {
		attributes = payloadattribute.EmptyWithVersion(version.Gloas)
	}

	payloadID, lastValidHash, err := s.cfg.SilaEngineCaller.ForkchoiceUpdated(ctx, fcs, attributes)
	if err == nil {
		return payloadID, nil
	}

	switch {
	case errors.Is(err, silaexec.ErrAcceptedSyncingPayloadStatus):
		log.WithFields(logrus.Fields{
			"headBlockHash":             fmt.Sprintf("%#x", bytesutil.Trunc(blockHash[:])),
			"finalizedPayloadBlockHash": fmt.Sprintf("%#x", bytesutil.Trunc(finalizedHash[:])),
		}).Info("Called forkchoice updated with optimistic block (Gloas)")
		return payloadID, nil
	case errors.Is(err, silaexec.ErrInvalidPayloadStatus):
		if len(lastValidHash) == 0 {
			lastValidHash = defaultLatestValidHash
		}
		return nil, invalidBlock{
			error:         ErrInvalidPayload,
			lastValidHash: bytesutil.ToBytes32(lastValidHash),
		}
	default:
		log.WithError(err).Error(ErrUndefinedSilaEngineError)
		return nil, nil
	}
}
