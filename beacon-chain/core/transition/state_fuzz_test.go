package transition

import (
	"testing"

	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	fuzz "github.com/google/gofuzz"
)

func TestGenesisBeaconState_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	deposits := make([]*silapb.Deposit, 300000)
	var genesisTime uint64
	silaexecData := &silapb.SilaData{}
	for range 1000 {
		fuzzer.Fuzz(&deposits)
		fuzzer.Fuzz(&genesisTime)
		fuzzer.Fuzz(silaexecData)
		gs, err := GenesisBeaconState(t.Context(), deposits, genesisTime, silaexecData)
		if err != nil {
			if gs != nil {
				t.Fatalf("Genesis state should be nil on err. found: %v on error: %v for inputs deposit: %v "+
					"genesis time: %v silaData: %v", gs, err, deposits, genesisTime, silaexecData)
			}
		}
	}
}

func TestOptimizedGenesisBeaconState_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	var genesisTime uint64
	preState, err := state_native.InitializeFromProtoUnsafePhase0(&silapb.BeaconState{})
	require.NoError(t, err)
	silaexecData := &silapb.SilaData{}
	for range 1000 {
		fuzzer.Fuzz(&genesisTime)
		fuzzer.Fuzz(silaexecData)
		fuzzer.Fuzz(preState)
		gs, err := OptimizedGenesisBeaconState(genesisTime, preState, silaexecData)
		if err != nil {
			if gs != nil {
				t.Fatalf("Genesis state should be nil on err. found: %v on error: %v for inputs genesis time: %v "+
					"pre state: %v silaData: %v", gs, err, genesisTime, preState, silaexecData)
			}
		}
	}
}

func TestIsValidGenesisState_100000(_ *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	var chainStartDepositCount, currentTime uint64
	for range 100000 {
		fuzzer.Fuzz(&chainStartDepositCount)
		fuzzer.Fuzz(&currentTime)
		IsValidGenesisState(chainStartDepositCount, currentTime)
	}
}
