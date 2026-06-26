package blockchain

import (
	"context"
	"testing"
	"time"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec"
	mockSila "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/features"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	payloadattribute "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/payload-attribute"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

func prepareGloasForkchoiceState(
	_ context.Context,
	slot primitives.Slot,
	blockRoot [32]byte,
	parentRoot [32]byte,
	blockHash [32]byte,
	parentBlockHash [32]byte,
	justifiedEpoch primitives.Epoch,
	finalizedEpoch primitives.Epoch,
) (state.BeaconState, blocks.ROBlock, error) {
	blockHeader := &silapb.BeaconBlockHeader{
		ParentRoot: parentRoot[:],
	}

	justifiedCheckpoint := &silapb.Checkpoint{
		Epoch: justifiedEpoch,
	}

	finalizedCheckpoint := &silapb.Checkpoint{
		Epoch: finalizedEpoch,
	}

	builderPendingPayments := make([]*silapb.BuilderPendingPayment, 64)
	for i := range builderPendingPayments {
		builderPendingPayments[i] = &silapb.BuilderPendingPayment{
			Withdrawal: &silapb.BuilderPendingWithdrawal{
				FeeRecipient: make([]byte, 20),
			},
		}
	}

	base := &silapb.BeaconStateGloas{
		Slot:                       slot,
		RandaoMixes:                make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
		CurrentJustifiedCheckpoint: justifiedCheckpoint,
		FinalizedCheckpoint:        finalizedCheckpoint,
		LatestBlockHeader:          blockHeader,
		LatestSilaPayloadBid: &silapb.SilaPayloadBid{
			BlockHash:          blockHash[:],
			ParentBlockHash:    parentBlockHash[:],
			ParentBlockRoot:    make([]byte, 32),
			PrevRandao:         make([]byte, 32),
			FeeRecipient:       make([]byte, 20),
			BlobKzgCommitments: [][]byte{make([]byte, 48)},
			SilaRequestsRoot:   make([]byte, 32),
		},
		Builders:                   make([]*silapb.Builder, 0),
		BuilderPendingPayments:     builderPendingPayments,
		SilaPayloadAvailability:    make([]byte, 1024),
		LatestBlockHash:            make([]byte, 32),
		PayloadExpectedWithdrawals: make([]*silaenginev1.Withdrawal, 0),
		ProposerLookahead:          make([]primitives.ValidatorIndex, 64),
	}

	st, err := state_native.InitializeFromProtoUnsafeGloas(base)
	if err != nil {
		return nil, blocks.ROBlock{}, err
	}

	bid := util.HydrateSignedSilaPayloadBid(&silapb.SignedSilaPayloadBid{
		Message: &silapb.SilaPayloadBid{
			BlockHash:       blockHash[:],
			ParentBlockHash: parentBlockHash[:],
		},
	})

	blk := util.HydrateSignedBeaconBlockGloas(&silapb.SignedBeaconBlockGloas{
		Block: &silapb.BeaconBlockGloas{
			Slot:       slot,
			ParentRoot: parentRoot[:],
			Body: &silapb.BeaconBlockBodyGloas{
				SignedSilaPayloadBid: bid,
			},
		},
	})

	signed, err := blocks.NewSignedBeaconBlock(blk)
	if err != nil {
		return nil, blocks.ROBlock{}, err
	}
	roblock, err := blocks.NewROBlockWithRoot(signed, blockRoot)
	return st, roblock, err
}

func testGloasState(t *testing.T, slot primitives.Slot, parentRoot [32]byte, blockHash [32]byte) (*silapb.BeaconStateGloas, *silapb.SignedBeaconBlockGloas) {
	t.Helper()
	builderPendingPayments := make([]*silapb.BuilderPendingPayment, 64)
	for i := range builderPendingPayments {
		builderPendingPayments[i] = &silapb.BuilderPendingPayment{
			Withdrawal: &silapb.BuilderPendingWithdrawal{FeeRecipient: make([]byte, 20)},
		}
	}
	base := &silapb.BeaconStateGloas{
		Slot:                       slot,
		RandaoMixes:                make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
		BlockRoots:                 make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot),
		StateRoots:                 make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot),
		Slashings:                  make([]uint64, params.BeaconConfig().EpochsPerSlashingsVector),
		CurrentJustifiedCheckpoint: &silapb.Checkpoint{Root: make([]byte, 32)},
		FinalizedCheckpoint:        &silapb.Checkpoint{Root: make([]byte, 32)},
		LatestBlockHeader: &silapb.BeaconBlockHeader{
			ParentRoot: parentRoot[:],
			StateRoot:  make([]byte, 32),
			BodyRoot:   make([]byte, 32),
		},
		SilaData: &silapb.SilaData{
			DepositRoot: make([]byte, 32),
			BlockHash:   make([]byte, 32),
		},
		LatestSilaPayloadBid: &silapb.SilaPayloadBid{
			BlockHash:          blockHash[:],
			ParentBlockHash:    make([]byte, 32),
			ParentBlockRoot:    make([]byte, 32),
			PrevRandao:         make([]byte, 32),
			FeeRecipient:       make([]byte, 20),
			BlobKzgCommitments: [][]byte{make([]byte, 48)},
			SilaRequestsRoot:   make([]byte, 32),
		},
		Builders:                   make([]*silapb.Builder, 0),
		BuilderPendingPayments:     builderPendingPayments,
		SilaPayloadAvailability:    make([]byte, 1024),
		LatestBlockHash:            make([]byte, 32),
		PayloadExpectedWithdrawals: make([]*silaenginev1.Withdrawal, 0),
		ProposerLookahead:          make([]primitives.ValidatorIndex, 64),
	}

	bid := util.HydrateSignedSilaPayloadBid(&silapb.SignedSilaPayloadBid{
		Message: &silapb.SilaPayloadBid{
			BlockHash:       blockHash[:],
			ParentBlockHash: make([]byte, 32),
		},
	})

	blk := util.HydrateSignedBeaconBlockGloas(&silapb.SignedBeaconBlockGloas{
		Block: &silapb.BeaconBlockGloas{
			Slot:       slot,
			ParentRoot: parentRoot[:],
			Body:       &silapb.BeaconBlockBodyGloas{SignedSilaPayloadBid: bid},
		},
	})
	return base, blk
}

func testSignedEnvelope(t *testing.T, blockRoot [32]byte, slot primitives.Slot, blockHash []byte) *silapb.SignedSilaPayloadEnvelope {
	t.Helper()
	return &silapb.SignedSilaPayloadEnvelope{
		Message: &silapb.SilaPayloadEnvelope{
			Payload: &silaenginev1.SilaPayloadGloas{
				ParentHash:    make([]byte, 32),
				FeeRecipient:  make([]byte, 20),
				StateRoot:     make([]byte, 32),
				ReceiptsRoot:  make([]byte, 32),
				LogsBloom:     make([]byte, 256),
				PrevRandao:    make([]byte, 32),
				BaseFeePerGas: make([]byte, 32),
				BlockHash:     blockHash,
				Transactions:  [][]byte{},
				Withdrawals:   []*silaenginev1.Withdrawal{},
			},
			SilaRequests:          &silaenginev1.SilaRequests{},
			BuilderIndex:          0,
			BeaconBlockRoot:       blockRoot[:],
			ParentBeaconBlockRoot: make([]byte, 32),
		},
		Signature: make([]byte, 96),
	}
}

func setupGloasService(t *testing.T, silaEngineClient *mockSila.SilaEngineClient) (*Service, *testServiceRequirements) {
	t.Helper()
	return minimalTestService(t,
		WithPayloadIDCache(cache.NewPayloadIDCache()),
		WithSilaEngineCaller(silaEngineClient),
	)
}

func insertGloasBlock(t *testing.T, s *Service, base *silapb.BeaconStateGloas, blk *silapb.SignedBeaconBlockGloas, blockRoot [32]byte) {
	t.Helper()
	ctx := t.Context()
	st, err := state_native.InitializeFromProtoUnsafeGloas(base)
	require.NoError(t, err)
	signed, err := blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	roblock, err := blocks.NewROBlockWithRoot(signed, blockRoot)
	require.NoError(t, err)
	require.NoError(t, s.cfg.BeaconDB.SaveBlock(ctx, signed))
	require.NoError(t, s.cfg.BeaconDB.SaveStateSummary(ctx, &silapb.StateSummary{Root: blockRoot[:], Slot: blk.Block.Slot}))
	require.NoError(t, s.cfg.StateGen.SaveState(ctx, blockRoot, st))
	require.NoError(t, s.InsertNode(ctx, st, roblock))
}

func TestGetPayloadEnvelopePrestate_UnknownRoot(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})
	ctx := t.Context()
	unknownRoot := bytesutil.ToBytes32([]byte("unknown"))
	env := &silapb.SilaPayloadEnvelope{
		BeaconBlockRoot:       unknownRoot[:],
		ParentBeaconBlockRoot: make([]byte, 32),
		Payload:               &silaenginev1.SilaPayloadGloas{},
	}
	envelope, err := blocks.WrappedROSilaPayloadEnvelope(env)
	require.NoError(t, err)
	_, err = s.getPayloadEnvelopePrestate(ctx, envelope)
	require.ErrorContains(t, "not found in forkchoice", err)
}

func TestGetPayloadEnvelopePrestate_OK(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})
	ctx := t.Context()

	blockRoot := bytesutil.ToBytes32([]byte("root1"))
	parentRoot := params.BeaconConfig().ZeroHash
	blockHash := bytesutil.ToBytes32([]byte("hash1"))

	base, blk := testGloasState(t, 1, parentRoot, blockHash)
	insertGloasBlock(t, s, base, blk, blockRoot)

	env := &silapb.SilaPayloadEnvelope{
		BeaconBlockRoot:       blockRoot[:],
		ParentBeaconBlockRoot: make([]byte, 32),
		Payload:               &silaenginev1.SilaPayloadGloas{},
	}
	envelope, err := blocks.WrappedROSilaPayloadEnvelope(env)
	require.NoError(t, err)
	st, err := s.getPayloadEnvelopePrestate(ctx, envelope)
	require.NoError(t, err)
	require.Equal(t, primitives.Slot(1), st.Slot())
}

func TestNotifyNewEnvelope_Valid(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})
	ctx := t.Context()

	blockRoot := bytesutil.ToBytes32([]byte("root1"))
	parentRoot := params.BeaconConfig().ZeroHash
	blockHash := bytesutil.ToBytes32([]byte("hash1"))

	base, _ := testGloasState(t, 1, parentRoot, blockHash)
	st, err := state_native.InitializeFromProtoUnsafeGloas(base)
	require.NoError(t, err)

	env := &silapb.SilaPayloadEnvelope{
		BeaconBlockRoot:       blockRoot[:],
		ParentBeaconBlockRoot: make([]byte, 32),
		Payload:               &silaenginev1.SilaPayloadGloas{BlockHash: blockHash[:]},
		SilaRequests:          &silaenginev1.SilaRequests{},
	}
	envelope, err := blocks.WrappedROSilaPayloadEnvelope(env)
	require.NoError(t, err)

	isValid, err := s.notifyNewEnvelope(ctx, st, envelope)
	require.NoError(t, err)
	require.Equal(t, true, isValid)
}

func TestNotifyNewEnvelope_Syncing(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{
		ErrNewPayload: silaexec.ErrAcceptedSyncingPayloadStatus,
	})
	ctx := t.Context()

	blockRoot := bytesutil.ToBytes32([]byte("root1"))
	parentRoot := params.BeaconConfig().ZeroHash
	blockHash := bytesutil.ToBytes32([]byte("hash1"))

	base, _ := testGloasState(t, 1, parentRoot, blockHash)
	st, err := state_native.InitializeFromProtoUnsafeGloas(base)
	require.NoError(t, err)

	env := &silapb.SilaPayloadEnvelope{
		BeaconBlockRoot:       blockRoot[:],
		ParentBeaconBlockRoot: make([]byte, 32),
		Payload:               &silaenginev1.SilaPayloadGloas{BlockHash: blockHash[:]},
		SilaRequests:          &silaenginev1.SilaRequests{},
	}
	envelope, err := blocks.WrappedROSilaPayloadEnvelope(env)
	require.NoError(t, err)

	isValid, err := s.notifyNewEnvelope(ctx, st, envelope)
	require.NoError(t, err)
	require.Equal(t, false, isValid)
}

func TestNotifyNewEnvelope_Invalid(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{
		ErrNewPayload: silaexec.ErrInvalidPayloadStatus,
	})
	ctx := t.Context()

	blockRoot := bytesutil.ToBytes32([]byte("root1"))
	parentRoot := params.BeaconConfig().ZeroHash
	blockHash := bytesutil.ToBytes32([]byte("hash1"))

	base, _ := testGloasState(t, 1, parentRoot, blockHash)
	st, err := state_native.InitializeFromProtoUnsafeGloas(base)
	require.NoError(t, err)

	env := &silapb.SilaPayloadEnvelope{
		BeaconBlockRoot:       blockRoot[:],
		ParentBeaconBlockRoot: make([]byte, 32),
		Payload:               &silaenginev1.SilaPayloadGloas{BlockHash: blockHash[:]},
		SilaRequests:          &silaenginev1.SilaRequests{},
	}
	envelope, err := blocks.WrappedROSilaPayloadEnvelope(env)
	require.NoError(t, err)

	_, err = s.notifyNewEnvelope(ctx, st, envelope)
	require.Equal(t, true, IsInvalidBlock(err))
}

func TestNotifyForkchoiceUpdateGloas_Valid(t *testing.T) {
	pid := &silaenginev1.PayloadIDBytes{1, 2, 3, 4, 5, 6, 7, 8}
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{PayloadIDBytes: pid})
	ctx := t.Context()

	blockHash := bytesutil.ToBytes32([]byte("hash1"))
	attr := payloadattribute.EmptyWithVersion(version.Gloas)

	retPid, err := s.notifyForkchoiceUpdateGloas(ctx, blockHash, attr)
	require.NoError(t, err)
	require.DeepEqual(t, pid, retPid)
}

func TestNotifyForkchoiceUpdateGloas_Syncing(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{
		ErrForkchoiceUpdated: silaexec.ErrAcceptedSyncingPayloadStatus,
	})
	ctx := t.Context()

	blockHash := bytesutil.ToBytes32([]byte("hash1"))
	_, err := s.notifyForkchoiceUpdateGloas(ctx, blockHash, nil)
	require.NoError(t, err)
}

func TestNotifyForkchoiceUpdateGloas_Invalid(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{
		ErrForkchoiceUpdated: silaexec.ErrInvalidPayloadStatus,
	})
	ctx := t.Context()

	blockHash := bytesutil.ToBytes32([]byte("hash1"))
	_, err := s.notifyForkchoiceUpdateGloas(ctx, blockHash, nil)
	require.Equal(t, true, IsInvalidBlock(err))
}

func TestNotifyForkchoiceUpdateGloas_NilAttributes(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})
	ctx := t.Context()

	blockHash := bytesutil.ToBytes32([]byte("hash1"))
	_, err := s.notifyForkchoiceUpdateGloas(ctx, blockHash, nil)
	require.NoError(t, err)
}

func TestFcuFromReorgData_CachesPayloadID(t *testing.T) {
	logHook := logTest.NewGlobal()
	pid := &silaenginev1.PayloadIDBytes{1, 2, 3, 4, 5, 6, 7, 8}
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{PayloadIDBytes: pid})

	headRoot := bytesutil.ToBytes32([]byte("headroot"))
	headHash := bytesutil.ToBytes32([]byte("headhash"))
	proposingSlot := primitives.Slot(2)
	attr, err := payloadattribute.New(&silaenginev1.PayloadAttributesV4{
		Timestamp:             1,
		PrevRandao:            make([]byte, 32),
		SuggestedFeeRecipient: make([]byte, 20),
		Withdrawals:           []*silaenginev1.Withdrawal{},
		ParentBeaconBlockRoot: make([]byte, 32),
	})
	require.NoError(t, err)
	require.Equal(t, false, attr.IsEmpty())

	s.fcuFromReorgData(headRoot, headHash, attr, proposingSlot)

	require.LogsDoNotContain(t, logHook, "Could not update forkchoice with engine")
	cachedPid, has := s.cfg.PayloadIDCache.PayloadID(proposingSlot, headRoot)
	require.Equal(t, true, has)
	require.Equal(t, primitives.PayloadID(pid[:]), cachedPid)
}

func TestFcuFromReorgData_NilPayloadID_NoCache(t *testing.T) {
	// Engine returns no payload ID (nil), so nothing should be cached.
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})

	headRoot := bytesutil.ToBytes32([]byte("headroot"))
	headHash := bytesutil.ToBytes32([]byte("headhash"))
	proposingSlot := primitives.Slot(2)
	attr := payloadattribute.EmptyWithVersion(version.Gloas)

	s.fcuFromReorgData(headRoot, headHash, attr, proposingSlot)

	_, has := s.cfg.PayloadIDCache.PayloadID(proposingSlot, headRoot)
	require.Equal(t, false, has)
}

func TestFcuFromReorgData_EngineError(t *testing.T) {
	logHook := logTest.NewGlobal()
	// An invalid-payload status surfaces as an error from notifyForkchoiceUpdateGloas.
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{
		ErrForkchoiceUpdated: silaexec.ErrInvalidPayloadStatus,
	})

	headRoot := bytesutil.ToBytes32([]byte("headroot"))
	headHash := bytesutil.ToBytes32([]byte("headhash"))
	proposingSlot := primitives.Slot(2)
	attr := payloadattribute.EmptyWithVersion(version.Gloas)

	s.fcuFromReorgData(headRoot, headHash, attr, proposingSlot)

	require.LogsContain(t, logHook, "Could not update forkchoice with engine")
	_, has := s.cfg.PayloadIDCache.PayloadID(proposingSlot, headRoot)
	require.Equal(t, false, has)
}

func TestSavePostPayload(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})
	ctx := t.Context()

	blockRoot := bytesutil.ToBytes32([]byte("root1"))
	blockHash := bytesutil.ToBytes32([]byte("hash1"))

	protoEnv := testSignedEnvelope(t, blockRoot, 1, blockHash[:])
	signed, err := blocks.WrappedROSignedSilaPayloadEnvelope(protoEnv)
	require.NoError(t, err)

	require.NoError(t, s.savePostPayload(ctx, signed))

	// Verify the envelope was saved in the DB.
	require.Equal(t, true, s.cfg.BeaconDB.HasSilaPayloadEnvelope(ctx, blockRoot))
}

func TestValidateSilaPayloadOnEnvelope_Valid(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})
	ctx := t.Context()

	blockRoot := bytesutil.ToBytes32([]byte("root1"))
	parentRoot := params.BeaconConfig().ZeroHash
	blockHash := bytesutil.ToBytes32([]byte("hash1"))

	base, _ := testGloasState(t, 1, parentRoot, blockHash)
	st, err := state_native.InitializeFromProtoUnsafeGloas(base)
	require.NoError(t, err)

	env := &silapb.SilaPayloadEnvelope{
		BeaconBlockRoot:       blockRoot[:],
		ParentBeaconBlockRoot: make([]byte, 32),
		Payload:               &silaenginev1.SilaPayloadGloas{BlockHash: blockHash[:], ParentHash: make([]byte, 32)},
		SilaRequests:          &silaenginev1.SilaRequests{},
	}
	envelope, err := blocks.WrappedROSilaPayloadEnvelope(env)
	require.NoError(t, err)

	isValid, err := s.validateSilaPayloadOnEnvelope(ctx, st, envelope)
	require.NoError(t, err)
	require.Equal(t, true, isValid)
}

func TestPostPayloadTasks_NotHead(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})
	ctx := t.Context()

	root := bytesutil.ToBytes32([]byte("root1"))
	headRoot := bytesutil.ToBytes32([]byte("different"))
	blockHash := bytesutil.ToBytes32([]byte("hash1"))

	base, _ := testGloasState(t, 1, params.BeaconConfig().ZeroHash, blockHash)
	st, err := state_native.InitializeFromProtoUnsafeGloas(base)
	require.NoError(t, err)

	env := &silapb.SilaPayloadEnvelope{
		BeaconBlockRoot:       root[:],
		ParentBeaconBlockRoot: make([]byte, 32),
		Payload:               &silaenginev1.SilaPayloadGloas{BlockHash: blockHash[:]},
	}
	envelope, err := blocks.WrappedROSilaPayloadEnvelope(env)
	require.NoError(t, err)

	require.NoError(t, s.postPayloadTasks(ctx, envelope, st, root, headRoot))
}

func TestPostPayloadTasks_DoesNotMutateHead(t *testing.T) {
	s, _ := setupGloasService(t, &mockSila.SilaEngineClient{})
	ctx := t.Context()

	root := bytesutil.ToBytes32([]byte("root1"))
	blockHash := bytesutil.ToBytes32([]byte("hash1"))

	base, blk := testGloasState(t, 1, params.BeaconConfig().ZeroHash, blockHash)
	st, err := state_native.InitializeFromProtoUnsafeGloas(base)
	require.NoError(t, err)
	oldBase, _ := testGloasState(t, 0, params.BeaconConfig().ZeroHash, blockHash)
	oldSt, err := state_native.InitializeFromProtoUnsafeGloas(oldBase)
	require.NoError(t, err)
	signed, err := blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)

	s.head = &head{root: root, block: signed, state: st, slot: 1}
	s.head.state = oldSt

	env := &silapb.SilaPayloadEnvelope{
		BeaconBlockRoot:       root[:],
		ParentBeaconBlockRoot: make([]byte, 32),
		Payload:               &silaenginev1.SilaPayloadGloas{BlockHash: blockHash[:], ParentHash: make([]byte, 32)},
	}
	envelope, err := blocks.WrappedROSilaPayloadEnvelope(env)
	require.NoError(t, err)

	require.NoError(t, s.postPayloadTasks(ctx, envelope, st, root, root))

	s.headLock.RLock()
	require.Equal(t, root, s.head.root)
	require.Equal(t, primitives.Slot(0), s.head.state.Slot())
	s.headLock.RUnlock()
}

func TestLatePayloadTasks_ReturnsEarlyWhenBlockLate(t *testing.T) {
	logHook := logTest.NewGlobal()
	service, tr := setupGloasService(t, &mockSila.SilaEngineClient{})

	blockHash := bytesutil.ToBytes32([]byte("hash1"))
	base, _ := testGloasState(t, 1, params.BeaconConfig().ZeroHash, blockHash)
	base.LatestBlockHash = blockHash[:]
	st, err := state_native.InitializeFromProtoUnsafeGloas(base)
	require.NoError(t, err)

	headRoot := bytesutil.ToBytes32([]byte("headroot"))
	service.head = &head{
		root:  headRoot,
		state: st,
		slot:  1,
	}
	// Set genesis time so CurrentSlot > HeadSlot.
	service.SetGenesisTime(time.Now().Add(-2 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second))

	service.latePayloadTasks(tr.ctx)
	require.LogsDoNotContain(t, logHook, "Could not notify forkchoice update")
	// No payload ID should have been cached.
	_, has := service.cfg.PayloadIDCache.PayloadID(service.CurrentSlot()+1, headRoot)
	require.Equal(t, false, has)
}

func TestLatePayloadTasks_SendsFCU(t *testing.T) {
	logHook := logTest.NewGlobal()
	resetCfg := features.InitWithReset(&features.Flags{
		PrepareAllPayloads: true,
	})
	defer resetCfg()

	pid := &silaenginev1.PayloadIDBytes{1, 2, 3, 4, 5, 6, 7, 8}
	service, tr := setupGloasService(t, &mockSila.SilaEngineClient{PayloadIDBytes: pid})

	blockHash := bytesutil.ToBytes32([]byte("hash1"))
	base, blk := testGloasState(t, 1, params.BeaconConfig().ZeroHash, blockHash)
	base.LatestBlockHash = blockHash[:]
	st, err := state_native.InitializeFromProtoUnsafeGloas(base)
	require.NoError(t, err)

	signed, err := blocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)

	headRoot := bytesutil.ToBytes32([]byte("headroot"))
	insertGloasBlock(t, service, base, blk, headRoot)
	service.head = &head{
		root:  headRoot,
		block: signed,
		state: st,
		slot:  1,
	}
	// CurrentSlot == HeadSlot == 1: place genesis 1.5 slots ago so we're solidly in slot 1.
	service.SetGenesisTime(time.Now().Add(-3 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second / 2))
	service.SetForkChoiceGenesisTime(service.genesisTime)

	service.latePayloadTasks(tr.ctx)
	require.LogsDoNotContain(t, logHook, "Could not notify forkchoice update")
	require.LogsDoNotContain(t, logHook, "Could not get")
	// Payload ID should have been cached.
	cachedPid, has := service.cfg.PayloadIDCache.PayloadID(service.CurrentSlot()+1, headRoot)
	require.Equal(t, true, has)
	require.Equal(t, primitives.PayloadID(pid[:]), cachedPid)
}

func TestLateBlockTasks_GloasFCU(t *testing.T) {
	logHook := logTest.NewGlobal()
	resetCfg := features.InitWithReset(&features.Flags{
		PrepareAllPayloads: true,
	})
	defer resetCfg()

	pid := &silaenginev1.PayloadIDBytes{1, 2, 3, 4, 5, 6, 7, 8}
	service, tr := setupGloasService(t, &mockSila.SilaEngineClient{PayloadIDBytes: pid})

	blockHash := bytesutil.ToBytes32([]byte("hash1"))
	base, blk := testGloasState(t, 1, params.BeaconConfig().ZeroHash, blockHash)
	base.LatestBlockHash = blockHash[:]
	st, err := state_native.InitializeFromProtoUnsafeGloas(base)
	require.NoError(t, err)

	headRoot := bytesutil.ToBytes32([]byte("headroot"))
	insertGloasBlock(t, service, base, blk, headRoot)
	service.head = &head{
		root:  headRoot,
		state: st,
		slot:  1,
	}

	// Set genesis time so CurrentSlot > HeadSlot, triggering late block logic.
	service.SetGenesisTime(time.Now().Add(-2 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second))
	service.SetForkChoiceGenesisTime(service.genesisTime)

	service.lateBlockTasks(tr.ctx)
	require.LogsDoNotContain(t, logHook, "could not perform late block tasks")

	// Payload ID should have been cached by the Gloas FCU path.
	cachedPid, has := service.cfg.PayloadIDCache.PayloadID(service.CurrentSlot()+1, headRoot)
	require.Equal(t, true, has)
	require.Equal(t, primitives.PayloadID(pid[:]), cachedPid)
}

// TestLateBlockTasks_GloasForkBoundary_PreforkBidUsesHeadRoot verifies that lateBlockTasks
// uses headRoot for the next-slot cache lookup even at the fork boundary.
func TestLateBlockTasks_GloasForkBoundary_PreforkBidUsesHeadRoot(t *testing.T) {
	logHook := logTest.NewGlobal()
	resetCfg := features.InitWithReset(&features.Flags{
		PrepareAllPayloads: true,
	})
	defer resetCfg()

	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.GloasForkEpoch = 1
	cfg.InitializeForkSchedule()
	params.OverrideBeaconConfig(cfg)

	pid := &silaenginev1.PayloadIDBytes{1, 2, 3, 4, 5, 6, 7, 8}
	service, tr := setupGloasService(t, &mockSila.SilaEngineClient{PayloadIDBytes: pid})

	blockHash := bytesutil.ToBytes32([]byte("hash1"))
	base, blk := testGloasState(t, 1, params.BeaconConfig().ZeroHash, blockHash)
	// Make LatestBlockHashMatchesBidBlockHash() true: bid.BlockHash == LatestBlockHash.
	base.LatestBlockHash = blockHash[:]
	// bid.Slot is 0 (pre-fork epoch).

	st, err := state_native.InitializeFromProtoUnsafeGloas(base)
	require.NoError(t, err)

	headRoot := bytesutil.ToBytes32([]byte("headroot"))
	insertGloasBlock(t, service, base, blk, headRoot)
	service.head = &head{
		root:  headRoot,
		state: st,
		slot:  1,
	}

	// Trigger late block logic: CurrentSlot > HeadSlot.
	service.SetGenesisTime(time.Now().Add(-2 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second))
	service.SetForkChoiceGenesisTime(service.genesisTime)

	service.lateBlockTasks(tr.ctx)
	require.LogsDoNotContain(t, logHook, "could not perform late block tasks")
}
