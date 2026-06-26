package transition

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/time"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	fuzz "github.com/google/gofuzz"
)

func TestFuzzExecuteStateTransition_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := t.Context()
	state, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	sb := &silapb.SignedBeaconBlock{}
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for range 1000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(sb)
		if sb.Block == nil || sb.Block.Body == nil {
			continue
		}
		wsb, err := blocks.NewSignedBeaconBlock(sb)
		require.NoError(t, err)
		s, err := ExecuteStateTransition(ctx, state, wsb)
		if err != nil && s != nil {
			t.Fatalf("state should be nil on err. found: %v on error: %v for state: %v and signed block: %v", s, err, state, sb)
		}
	}
}

func TestFuzzCalculateStateRoot_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := t.Context()
	state, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	sb := &silapb.SignedBeaconBlock{}
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for range 1000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(sb)
		if sb.Block == nil || sb.Block.Body == nil {
			continue
		}
		wsb, err := blocks.NewSignedBeaconBlock(sb)
		require.NoError(t, err)
		stateRoot, err := CalculateStateRoot(ctx, state, wsb)
		if err != nil && stateRoot != [32]byte{} {
			t.Fatalf("state root should be empty on err. found: %v on error: %v for signed block: %v", stateRoot, err, sb)
		}
	}
}

func TestFuzzProcessSlot_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := t.Context()
	state, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for range 1000 {
		fuzzer.Fuzz(state)
		s, err := ProcessSlot(ctx, state)
		if err != nil && s != nil {
			t.Fatalf("state should be nil on err. found: %v on error: %v for state: %v", s, err, state)
		}
	}
}

func TestFuzzProcessSlots_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := t.Context()
	state, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	slot := primitives.Slot(0)
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for range 1000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(&slot)
		s, err := ProcessSlots(ctx, state, slot)
		if err != nil && s != nil {
			t.Fatalf("state should be nil on err. found: %v on error: %v for state: %v", s, err, state)
		}
	}
}

func TestFuzzprocessOperationsNoVerify_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := t.Context()
	state, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	bb := &silapb.BeaconBlock{}
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for range 1000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(bb)
		if bb.Body == nil {
			continue
		}
		wb, err := blocks.NewBeaconBlock(bb)
		require.NoError(t, err)
		s, err := ProcessOperationsNoVerifyAttsSigs(ctx, state, wb)
		if err != nil && s != nil {
			t.Fatalf("state should be nil on err. found: %v on error: %v for block body: %v", s, err, bb)
		}
	}
}

func TestFuzzverifyOperationLengths_10000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	state, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	bb := &silapb.BeaconBlock{}
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(bb)
		if bb.Body == nil {
			continue
		}
		wb, err := blocks.NewBeaconBlock(bb)
		require.NoError(t, err)
		_, err = VerifyOperationLengths(t.Context(), state, wb)
		_ = err
	}
}

func TestFuzzCanProcessEpoch_10000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	state, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for range 10000 {
		fuzzer.Fuzz(state)
		time.CanProcessEpoch(state)
	}
}

func TestFuzzProcessEpochPrecompute_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := t.Context()
	state, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for range 1000 {
		fuzzer.Fuzz(state)
		s, err := ProcessEpochPrecompute(ctx, state)
		if err != nil && s != nil {
			t.Fatalf("state should be nil on err. found: %v on error: %v for state: %v", s, err, state)
		}
	}
}

func TestFuzzProcessBlockForStateRoot_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := t.Context()
	state, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	sb := &silapb.SignedBeaconBlock{}
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for range 1000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(sb)
		if sb.Block == nil || sb.Block.Body == nil || sb.Block.Body.SilaData == nil {
			continue
		}
		wsb, err := blocks.NewSignedBeaconBlock(sb)
		require.NoError(t, err)
		s, err := ProcessBlockForStateRoot(ctx, state, wsb)
		if err != nil && s != nil {
			t.Fatalf("state should be nil on err. found: %v on error: %v for signed block: %v", s, err, sb)
		}
	}
}
