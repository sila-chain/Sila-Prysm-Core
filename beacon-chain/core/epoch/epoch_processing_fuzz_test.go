package epoch

import (
	"testing"

	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/fuzz"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	gofuzz "github.com/google/gofuzz"
)

func TestFuzzFinalUpdates_10000(t *testing.T) {
	fuzzer := gofuzz.NewWithSeed(0)
	base := &silapb.BeaconState{}

	for i := range 10000 {
		fuzzer.Fuzz(base)
		s, err := state_native.InitializeFromProtoUnsafePhase0(base)
		require.NoError(t, err)
		_, err = ProcessFinalUpdates(s)
		_ = err
		fuzz.FreeMemory(i)
	}
}
