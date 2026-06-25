package altair_test

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/altair"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/fuzz"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	gofuzz "github.com/google/gofuzz"
)

func TestFuzzProcessDeposits_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconStateAltair{}
	deposits := make([]*silapb.Deposit, 100)
	ctx := t.Context()
	for i := range 10000 {
		fuzzer.Fuzz(state)
		for i := range deposits {
			fuzzer.Fuzz(deposits[i])
		}
		s, err := state_native.InitializeFromProtoUnsafeAltair(state)
		require.NoError(t, err)
		r, err := altair.ProcessDeposits(ctx, s, deposits)
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and block: %v", r, err, state, deposits)
		}
		fuzz.FreeMemory(i)
	}
}

func TestFuzzProcessPreGenesisDeposit_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconStateAltair{}
	deposit := &silapb.Deposit{}
	ctx := t.Context()

	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(deposit)
		s, err := state_native.InitializeFromProtoUnsafeAltair(state)
		require.NoError(t, err)
		r, err := altair.ProcessPreGenesisDeposits(ctx, s, []*silapb.Deposit{deposit})
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and block: %v", r, err, state, deposit)
		}
		fuzz.FreeMemory(i)
	}
}

func TestFuzzProcessPreGenesisDeposit_Phase0_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	deposit := &silapb.Deposit{}
	ctx := t.Context()

	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(deposit)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		r, err := altair.ProcessPreGenesisDeposits(ctx, s, []*silapb.Deposit{deposit})
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and block: %v", r, err, state, deposit)
		}
		fuzz.FreeMemory(i)
	}
}

func TestFuzzProcessDeposit_Phase0_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconState{}
	deposit := &silapb.Deposit{}

	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(deposit)
		s, err := state_native.InitializeFromProtoUnsafePhase0(state)
		require.NoError(t, err)
		r, err := altair.ProcessDeposit(s, deposit, true)
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and block: %v", r, err, state, deposit)
		}
		fuzz.FreeMemory(i)
	}
}

func TestFuzzProcessDeposit_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	state := &silapb.BeaconStateAltair{}
	deposit := &silapb.Deposit{}

	for i := range 10000 {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(deposit)
		s, err := state_native.InitializeFromProtoUnsafeAltair(state)
		require.NoError(t, err)
		r, err := altair.ProcessDeposit(s, deposit, true)
		if err != nil && r != nil {
			t.Fatalf("return value should be nil on err. found: %v on error: %v for state: %v and block: %v", r, err, state, deposit)
		}
		fuzz.FreeMemory(i)
	}
}
