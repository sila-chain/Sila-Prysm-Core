package util

import (
	"testing"
	"time"

	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func TestDeterministicGenesisStateBellatrix(t *testing.T) {
	st, k := DeterministicGenesisStateBellatrix(t, params.BeaconConfig().MaxCommitteesPerSlot)
	require.Equal(t, params.BeaconConfig().MaxCommitteesPerSlot, uint64(len(k)))
	require.Equal(t, params.BeaconConfig().MaxCommitteesPerSlot, uint64(st.NumValidators()))
}

func TestGenesisBeaconStateBellatrix(t *testing.T) {
	ctx := t.Context()
	deposits, _, err := DeterministicDepositsAndKeys(params.BeaconConfig().MaxCommitteesPerSlot)
	require.NoError(t, err)
	silaexecData, err := DeterministicSilaData(len(deposits))
	require.NoError(t, err)
	gt := time.Now()
	st, err := genesisBeaconStateBellatrix(ctx, deposits, gt, silaexecData)
	require.NoError(t, err)
	require.Equal(t, gt.Truncate(time.Second), st.GenesisTime()) // Beacon state only keeps time precision of 1s, so we truncate.
	require.Equal(t, params.BeaconConfig().MaxCommitteesPerSlot, uint64(st.NumValidators()))
}
