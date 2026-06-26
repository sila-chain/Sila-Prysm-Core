package blockchain

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sila-chain/Sila-Consensus-Core/v7/async/event"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed"
	statefeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/time"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/transition"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/features"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	blocktypes "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	payloadattribute "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/payload-attribute"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/tracing/trace"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/sila-chain/Sila/common"
	"github.com/sirupsen/logrus"
)

var defaultLatestValidHash = bytesutil.PadTo([]byte{0xff}, 32)

// notifyForkchoiceUpdate signals SilaEngine the fork choice updates. SilaEngine should:
// 1. Re-organizes the sila payload chain and corresponding state to make head_block_hash the head.
// 2. Applies finality to the execution state: it irreversibly persists the chain of all sila payloads and corresponding state, up to and including finalized_block_hash.
func (s *Service) notifyForkchoiceUpdate(ctx context.Context, arg *fcuConfig) (*silaenginev1.PayloadIDBytes, error) {
	ctx, span := trace.StartSpan(ctx, "blockChain.notifyForkchoiceUpdate")
	defer span.End()

	if arg.headBlock == nil || arg.headBlock.IsNil() {
		log.Error("Head block is nil")
		return nil, nil
	}
	headBlk := arg.headBlock.Block()
	if headBlk == nil || headBlk.IsNil() || headBlk.Body().IsNil() {
		log.Error("Head block is nil")
		return nil, nil
	}
	// Must not call fork choice updated until the transition conditions are met on the Pow network.
	isSilaBlk, err := blocks.IsSilaBlock(headBlk.Body())
	if err != nil {
		log.WithError(err).Error("Could not determine if head block is sila block")
		return nil, nil
	}
	if !isSilaBlk {
		return nil, nil
	}
	headPayload, err := headBlk.Body().Execution()
	if err != nil {
		log.WithError(err).Error("Could not get sila payload for head block")
		return nil, nil
	}
	finalizedHash := s.cfg.ForkChoiceStore.FinalizedPayloadBlockHash()
	justifiedHash := s.cfg.ForkChoiceStore.UnrealizedJustifiedPayloadBlockHash()
	fcs := &silaenginev1.ForkchoiceState{
		HeadBlockHash:      headPayload.BlockHash(),
		SafeBlockHash:      justifiedHash[:],
		FinalizedBlockHash: finalizedHash[:],
	}
	if len(fcs.HeadBlockHash) != 32 || [32]byte(fcs.HeadBlockHash) == [32]byte{} {
		// check if we are sending FCU at genesis
		hash, err := s.hashForGenesisBlock(ctx, arg.headRoot)
		if errors.Is(err, errNotGenesisRoot) {
			log.Error("Sending nil head block hash to SilaEngine")
			return nil, nil
		}
		if err != nil {
			return nil, errors.Wrap(err, "could not get head block hash")
		}
		fcs.HeadBlockHash = hash
	}
	if arg.attributes == nil {
		arg.attributes = payloadattribute.EmptyWithVersion(headBlk.Version())
	}
	payloadID, lastValidHash, err := s.cfg.SilaEngineCaller.ForkchoiceUpdated(ctx, fcs, arg.attributes)
	if err != nil {
		switch {
		case errors.Is(err, silaexec.ErrAcceptedSyncingPayloadStatus):
			forkchoiceUpdatedOptimisticNodeCount.Inc()
			log.WithFields(logrus.Fields{
				"headSlot":                  headBlk.Slot(),
				"headPayloadBlockHash":      fmt.Sprintf("%#x", bytesutil.Trunc(headPayload.BlockHash())),
				"finalizedPayloadBlockHash": fmt.Sprintf("%#x", bytesutil.Trunc(finalizedHash[:])),
			}).Info("Called fork choice updated with optimistic block")
			return payloadID, nil
		case errors.Is(err, silaexec.ErrInvalidPayloadStatus):
			forkchoiceUpdatedInvalidNodeCount.Inc()
			headRoot := arg.headRoot
			if len(lastValidHash) == 0 {
				lastValidHash = defaultLatestValidHash
			}
			// this call has guaranteed to have the `headRoot` with its payload in forkchoice.
			invalidRoots, err := s.cfg.ForkChoiceStore.SetOptimisticToInvalid(ctx, headRoot, headBlk.ParentRoot(), bytesutil.ToBytes32(headPayload.ParentHash()), bytesutil.ToBytes32(lastValidHash))
			if err != nil {
				log.WithError(err).Error("Could not set head root to invalid")
				return nil, nil
			}
			// TODO: Gloas, we should not include the head root in this call
			if len(invalidRoots) == 0 || invalidRoots[0] != headRoot {
				invalidRoots = append([][32]byte{headRoot}, invalidRoots...)
			}
			if err := s.removeInvalidBlockAndState(ctx, invalidRoots); err != nil {
				log.WithError(err).Error("Could not remove invalid block and state")
				return nil, nil
			}

			r, _, full, err := s.cfg.ForkChoiceStore.FullHead(ctx)
			if err != nil {
				log.WithFields(logrus.Fields{
					"slot":                 headBlk.Slot(),
					"blockRoot":            fmt.Sprintf("%#x", bytesutil.Trunc(headRoot[:])),
					"invalidChildrenCount": len(invalidRoots),
				}).Warn("Pruned invalid blocks, could not update head root")
				return nil, invalidBlock{error: ErrInvalidPayload, root: arg.headRoot, invalidAncestorRoots: invalidRoots}
			}
			b, err := s.getBlock(ctx, r)
			if err != nil {
				log.WithError(err).Error("Could not get head block")
				return nil, nil
			}
			st, err := s.cfg.StateGen.StateByRoot(ctx, r)
			if err != nil {
				log.WithError(err).Error("Could not get head state")
				return nil, nil
			}
			pid, err := s.notifyForkchoiceUpdate(ctx, &fcuConfig{
				headState:  st,
				headRoot:   r,
				headBlock:  b,
				attributes: arg.attributes,
			})
			if err != nil {
				return nil, err // Returning err because it's recursive here.
			}

			if err := s.saveHead(ctx, r, b, st, full); err != nil {
				log.WithError(err).Error("Could not save head after pruning invalid blocks")
			}

			log.WithFields(logrus.Fields{
				"slot":                 headBlk.Slot(),
				"blockRoot":            fmt.Sprintf("%#x", bytesutil.Trunc(headRoot[:])),
				"invalidChildrenCount": len(invalidRoots),
				"newHeadRoot":          fmt.Sprintf("%#x", bytesutil.Trunc(r[:])),
			}).Warn("Pruned invalid blocks")
			return pid, invalidBlock{error: ErrInvalidPayload, root: arg.headRoot, invalidAncestorRoots: invalidRoots}
		default:
			log.WithError(err).Error(ErrUndefinedSilaEngineError)
			return nil, nil
		}
	}
	forkchoiceUpdatedValidNodeCount.Inc()
	if err := s.cfg.ForkChoiceStore.SetOptimisticToValid(ctx, arg.headRoot); err != nil {
		log.WithError(err).Error("Could not set head root to valid")
		return nil, nil
	}
	// If the forkchoice update call has an attribute, update the payload ID cache.
	hasAttr := arg.attributes != nil && !arg.attributes.IsEmpty()
	nextSlot := s.CurrentSlot() + 1
	if hasAttr && payloadID != nil {
		var pId [8]byte
		copy(pId[:], payloadID[:])
		log.WithFields(logrus.Fields{
			"blockRoot": fmt.Sprintf("%#x", bytesutil.Trunc(arg.headRoot[:])),
			"headSlot":  headBlk.Slot(),
			"nextSlot":  nextSlot,
			"payloadID": fmt.Sprintf("%#x", bytesutil.Trunc(payloadID[:])),
		}).Info("Forkchoice updated with payload attributes for proposal")
		s.cfg.PayloadIDCache.Set(nextSlot, arg.headRoot, pId)
		go s.firePayloadAttributesEvent(s.cfg.StateNotifier.StateFeed(), arg.headBlock, arg.headRoot, nextSlot)
	} else if hasAttr && payloadID == nil && !features.Get().PrepareAllPayloads {
		log.WithFields(logrus.Fields{
			"blockHash": fmt.Sprintf("%#x", headPayload.BlockHash()),
			"slot":      headBlk.Slot(),
			"nextSlot":  nextSlot,
		}).Error("Received nil payload ID on VALID engine response")
	}
	return payloadID, nil
}

func (s *Service) firePayloadAttributesEvent(f event.SubscriberSender, block interfaces.ReadOnlySignedBeaconBlock, root [32]byte, nextSlot primitives.Slot) {
	// If we're syncing a block in the past and init-sync is still running, we shouldn't fire this event.
	if !s.cfg.SyncChecker.Synced() {
		return
	}
	// the fcu args have differing amounts of completeness based on the code path,
	// and there is work we only want to do if a client is actually listening to the events beacon api endpoint.
	// temporary solution: just fire a blank event and fill in the details in the api handler.
	f.Send(&feed.Event{
		Type: statefeed.PayloadAttributes,
		Data: payloadattribute.EventData{HeadBlock: block, HeadRoot: root, ProposalSlot: nextSlot},
	})
}

// getPayloadHash returns the payload hash given the block root.
// if the block is before bellatrix fork epoch, it returns the zero hash.
func (s *Service) getPayloadHash(ctx context.Context, root []byte) ([32]byte, error) {
	blk, err := s.getBlock(ctx, s.ensureRootNotZeros(bytesutil.ToBytes32(root)))
	if err != nil {
		return [32]byte{}, err
	}
	if blocks.IsPreBellatrixVersion(blk.Block().Version()) {
		return params.BeaconConfig().ZeroHash, nil
	}
	payload, err := blk.Block().Body().Execution()
	if err != nil {
		return [32]byte{}, errors.Wrap(err, "could not get sila payload")
	}
	return bytesutil.ToBytes32(payload.BlockHash()), nil
}

// notifyNewPayload signals SilaEngine on a new payload.
// It returns true if the EL has returned VALID for the block
// stVersion should represent the version of the pre-state; header should also be from the pre-state.
func (s *Service) notifyNewPayload(ctx context.Context, stVersion int, header interfaces.SilaData, blk blocktypes.ROBlock) (bool, error) {
	ctx, span := trace.StartSpan(ctx, "blockChain.notifyNewPayload")
	defer span.End()

	// Sila payload is only supported in Bellatrix and beyond. Pre
	// merge blocks are never optimistic
	if stVersion < version.Bellatrix {
		return true, nil
	}
	if blk.Version() >= version.Gloas {
		return false, nil
	}
	body := blk.Block().Body()
	enabled, err := blocks.IsExecutionEnabledUsingHeader(header, body)
	if err != nil {
		return false, errors.Wrap(invalidBlock{error: err}, "could not determine if execution is enabled")
	}
	if !enabled {
		return true, nil
	}
	payload, err := body.Execution()
	if err != nil {
		return false, errors.Wrap(invalidBlock{error: err}, "could not get sila payload")
	}

	var lastValidHash []byte
	var parentRoot *common.Hash
	var versionedHashes []common.Hash
	var requests *silaenginev1.SilaRequests
	if blk.Version() >= version.Deneb {
		versionedHashes, err = kzgCommitmentsToVersionedHashes(blk.Block().Body())
		if err != nil {
			return false, errors.Wrap(err, "could not get versioned hashes to feed the engine")
		}
		prh := common.Hash(blk.Block().ParentRoot())
		parentRoot = &prh
	}
	if blk.Version() >= version.Electra {
		requests, err = blk.Block().Body().SilaRequests()
		if err != nil {
			return false, errors.Wrap(err, "could not get sila requests")
		}
		if requests == nil {
			return false, errors.New("nil sila requests")
		}
	}

	lastValidHash, err = s.cfg.SilaEngineCaller.NewPayload(ctx, payload, versionedHashes, parentRoot, requests)
	if err == nil {
		newPayloadValidNodeCount.Inc()
		return true, nil
	}
	logFields := logrus.Fields{
		"slot":             blk.Block().Slot(),
		"parentRoot":       fmt.Sprintf("%#x", parentRoot),
		"root":             fmt.Sprintf("%#x", blk.Root()),
		"payloadBlockHash": fmt.Sprintf("%#x", bytesutil.Trunc(payload.BlockHash())),
	}
	if errors.Is(err, silaexec.ErrAcceptedSyncingPayloadStatus) {
		newPayloadOptimisticNodeCount.Inc()
		log.WithFields(logFields).Info("Called new payload with optimistic block")
		return false, nil
	}
	if errors.Is(err, silaexec.ErrInvalidPayloadStatus) {
		log.WithFields(logFields).WithError(err).Error("Invalid payload status")
		return false, invalidBlock{
			error:         ErrInvalidPayload,
			lastValidHash: bytesutil.ToBytes32(lastValidHash),
		}
	}
	log.WithFields(logFields).WithError(err).Error("Unexpected SilaEngine error")
	return false, errors.WithMessage(ErrUndefinedSilaEngineError, err.Error())
}

// pruneInvalidBlock deals with the event that an invalid block was detected by the Sila layer
func (s *Service) pruneInvalidBlock(ctx context.Context, root, parentRoot, parentHash [32]byte, lvh [32]byte) error {
	newPayloadInvalidNodeCount.Inc()
	invalidRoots, err := s.cfg.ForkChoiceStore.SetOptimisticToInvalid(ctx, root, parentRoot, parentHash, lvh)
	if err != nil {
		return err
	}
	if err := s.removeInvalidBlockAndState(ctx, invalidRoots); err != nil {
		return err
	}
	log.WithFields(logrus.Fields{
		"blockRoot":            fmt.Sprintf("%#x", root),
		"invalidChildrenCount": len(invalidRoots),
	}).Warn("Pruned invalid blocks")
	return invalidBlock{
		invalidAncestorRoots: invalidRoots,
		error:                ErrInvalidPayload,
		lastValidHash:        lvh,
	}
}

// getPayloadAttributes returns the payload attributes for the given state and slot.
// The attribute is required to initiate a payload build process in the context of an `silaEngine_forkchoiceUpdated` call.
func (s *Service) getPayloadAttribute(ctx context.Context, st state.BeaconState, slot primitives.Slot, headRoot []byte, headFull bool) payloadattribute.Attributer {
	emptyAttri := payloadattribute.EmptyWithVersion(st.Version())

	// If it is an epoch boundary then process slots to get the right
	// shuffling before checking if the proposer is tracked. Otherwise
	// perform this check before. This is cheap as the NSC has already been updated.
	var val cache.TrackedValidator
	var ok bool
	e := slots.ToEpoch(slot)
	stateEpoch := slots.ToEpoch(st.Slot())
	fuluAndNextEpoch := st.Version() >= version.Fulu && e == stateEpoch+1
	if e == stateEpoch || fuluAndNextEpoch {
		val, ok = s.trackedProposer(st, slot)
		if !ok {
			return emptyAttri
		}
	}
	if slot > st.Slot() {
		// At this point either we know we are proposing on a future slot or we need to still compute the
		// right proposer index pre-Fulu, either way we need to copy the state to process it.
		st = st.Copy()
		var err error
		st, err = transition.ProcessSlotsUsingNextSlotCache(ctx, st, headRoot, slot)
		if err != nil {
			log.WithError(err).Error("Could not process slots to get payload attribute")
			return emptyAttri
		}
	}
	if e > stateEpoch && !fuluAndNextEpoch {
		emptyAttri := payloadattribute.EmptyWithVersion(st.Version())
		val, ok = s.trackedProposer(st, slot)
		if !ok {
			return emptyAttri
		}
	}
	// Get previous randao.
	prevRando, err := helpers.RandaoMix(st, time.CurrentEpoch(st))
	if err != nil {
		log.WithError(err).Error("Could not get randao mix to get payload attribute")
		return emptyAttri
	}

	// Get timestamp.
	t, err := slots.StartTime(s.genesisTime, slot)
	if err != nil {
		log.WithError(err).Error("Could not get timestamp to get payload attribute")
		return emptyAttri
	}

	v := st.Version()
	switch {
	case v >= version.Gloas:
		withdrawals, err := s.computePayloadWithdrawals(ctx, st, bytesutil.ToBytes32(headRoot), headFull)
		if err != nil {
			log.WithError(err).Error("Could not get withdrawals for payload attribute")
			return emptyAttri
		}
		targetGasLimit := val.GasLimit
		if targetGasLimit == 0 {
			targetGasLimit = params.BeaconConfig().DefaultBuilderGasLimit
		}
		return payloadAttributesGloas(uint64(t.Unix()), prevRando, val.FeeRecipient[:], headRoot, withdrawals, slot, targetGasLimit)
	case v >= version.Deneb:
		return payloadAttributesDeneb(st, uint64(t.Unix()), prevRando, val.FeeRecipient[:], headRoot)
	case v >= version.Capella:
		return payloadAttributesCapella(st, uint64(t.Unix()), prevRando, val.FeeRecipient[:])
	case v >= version.Bellatrix:
		return payloadAttributesBellatrix(uint64(t.Unix()), prevRando, val.FeeRecipient[:])
	default:
		log.WithField("version", version.String(v)).Error("Could not get payload attribute due to unknown state version")
		return payloadattribute.EmptyWithVersion(v)
	}
}

func payloadAttributesGloas(timestamp uint64, prevRandao, feeRecipient, parentBeaconBlockRoot []byte, withdrawals []*silaenginev1.Withdrawal, slot primitives.Slot, targetGasLimit uint64) payloadattribute.Attributer {
	attr, err := payloadattribute.New(&silaenginev1.PayloadAttributesV4{
		Timestamp:             timestamp,
		PrevRandao:            prevRandao,
		SuggestedFeeRecipient: feeRecipient,
		Withdrawals:           withdrawals,
		ParentBeaconBlockRoot: parentBeaconBlockRoot,
		SlotNumber:            uint64(slot),
		TargetGasLimit:        targetGasLimit,
	})
	if err != nil {
		log.WithError(err).Error("Could not get payload attribute")
		return payloadattribute.EmptyWithVersion(version.Gloas)
	}
	return attr
}

func payloadAttributesDeneb(st state.BeaconState, timestamp uint64, prevRandao, feeRecipient, parentBeaconBlockRoot []byte) payloadattribute.Attributer {
	withdrawals, _, err := st.ExpectedWithdrawals()
	if err != nil {
		log.WithError(err).Error("Could not get expected withdrawals to get payload attribute")
		return payloadattribute.EmptyWithVersion(st.Version())
	}
	attr, err := payloadattribute.New(&silaenginev1.PayloadAttributesV3{
		Timestamp:             timestamp,
		PrevRandao:            prevRandao,
		SuggestedFeeRecipient: feeRecipient,
		Withdrawals:           withdrawals,
		ParentBeaconBlockRoot: parentBeaconBlockRoot,
	})
	if err != nil {
		log.WithError(err).Error("Could not get payload attribute")
		return payloadattribute.EmptyWithVersion(st.Version())
	}
	return attr
}

func payloadAttributesCapella(st state.BeaconState, timestamp uint64, prevRandao, feeRecipient []byte) payloadattribute.Attributer {
	withdrawals, _, err := st.ExpectedWithdrawals()
	if err != nil {
		log.WithError(err).Error("Could not get expected withdrawals to get payload attribute")
		return payloadattribute.EmptyWithVersion(st.Version())
	}
	attr, err := payloadattribute.New(&silaenginev1.PayloadAttributesV2{
		Timestamp:             timestamp,
		PrevRandao:            prevRandao,
		SuggestedFeeRecipient: feeRecipient,
		Withdrawals:           withdrawals,
	})
	if err != nil {
		log.WithError(err).Error("Could not get payload attribute")
		return payloadattribute.EmptyWithVersion(st.Version())
	}
	return attr
}

func payloadAttributesBellatrix(timestamp uint64, prevRandao, feeRecipient []byte) payloadattribute.Attributer {
	attr, err := payloadattribute.New(&silaenginev1.PayloadAttributes{
		Timestamp:             timestamp,
		PrevRandao:            prevRandao,
		SuggestedFeeRecipient: feeRecipient,
	})
	if err != nil {
		log.WithError(err).Error("Could not get payload attribute")
		return payloadattribute.EmptyWithVersion(version.Bellatrix)
	}
	return attr
}

// removeInvalidBlockAndState removes the invalid block, blob and its corresponding state from the cache and DB.
func (s *Service) removeInvalidBlockAndState(ctx context.Context, blkRoots [][32]byte) error {
	for _, root := range blkRoots {
		if err := s.cfg.StateGen.DeleteStateFromCaches(ctx, root); err != nil {
			return err
		}
		// Delete block also deletes the state as well.
		if err := s.cfg.BeaconDB.DeleteBlock(ctx, root); err != nil {
			// TODO(10487): If a caller requests to delete a root that's justified and finalized. We should gracefully shutdown.
			// This is an irreparable condition, it would me a justified or finalized block has become invalid.
			return err
		}
		if err := s.blobStorage.Remove(root); err != nil {
			// Blobs may not exist for some blocks, leading to deletion failures. Log such errors at debug level.
			log.WithError(err).Debug("Could not remove blob from blob storage")
		}
		if err := s.dataColumnStorage.Remove(root); err != nil {
			log.WithError(err).Errorf("Could not remove data columns from data column storage for root %#x", root)
		}
	}
	return nil
}

func kzgCommitmentsToVersionedHashes(body interfaces.ReadOnlyBeaconBlockBody) ([]common.Hash, error) {
	commitments, err := body.BlobKzgCommitments()
	if err != nil {
		return nil, errors.Wrap(invalidBlock{error: err}, "could not get blob kzg commitments")
	}

	versionedHashes := make([]common.Hash, len(commitments))
	for i, commitment := range commitments {
		versionedHashes[i] = primitives.ConvertKzgCommitmentToVersionedHash(commitment)
	}
	return versionedHashes, nil
}
