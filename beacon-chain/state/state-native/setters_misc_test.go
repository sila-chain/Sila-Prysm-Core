package state_native_test

import (
	"testing"

	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func BenchmarkAppendHistoricalRoots(b *testing.B) {
	st, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{})
	require.NoError(b, err)

	max := params.BeaconConfig().HistoricalRootsLimit
	if max < 2 {
		b.Fatalf("HistoricalRootsLimit is less than 2: %d", max)
	}

	root := bytesutil.ToBytes32([]byte{0, 1, 2, 3, 4, 5})
	for i := uint64(0); i < max-2; i++ {
		err := st.AppendHistoricalRoots(root)
		require.NoError(b, err)
	}

	ref := st.Copy()

	for b.Loop() {
		err := ref.AppendHistoricalRoots(root)
		require.NoError(b, err)
		ref = st.Copy()
	}
}

func BenchmarkAppendHistoricalSummaries(b *testing.B) {
	st, err := state_native.InitializeFromProtoCapella(&silapb.BeaconStateCapella{})
	require.NoError(b, err)

	max := params.BeaconConfig().HistoricalRootsLimit
	if max < 2 {
		b.Fatalf("HistoricalRootsLimit is less than 2: %d", max)
	}

	for i := uint64(0); i < max-2; i++ {
		err := st.AppendHistoricalSummaries(&silapb.HistoricalSummary{})
		require.NoError(b, err)
	}

	ref := st.Copy()

	for b.Loop() {
		err := ref.AppendHistoricalSummaries(&silapb.HistoricalSummary{})
		require.NoError(b, err)
		ref = st.Copy()
	}
}
