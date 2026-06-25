package state_native_test

import (
	"testing"

	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func TestSetDepositRequestsStartIndex(t *testing.T) {
	t.Run("previous fork returns expected error", func(t *testing.T) {
		dState, _ := util.DeterministicGenesisState(t, 1)
		require.ErrorContains(t, "is not supported", dState.SetDepositRequestsStartIndex(1))
	})
	t.Run("electra sets expected value", func(t *testing.T) {
		old := uint64(2)
		dState, err := state_native.InitializeFromProtoElectra(&silapb.BeaconStateElectra{DepositRequestsStartIndex: old})
		require.NoError(t, err)
		want := uint64(3)
		require.NoError(t, dState.SetDepositRequestsStartIndex(want))
		got, err := dState.DepositRequestsStartIndex()
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}
