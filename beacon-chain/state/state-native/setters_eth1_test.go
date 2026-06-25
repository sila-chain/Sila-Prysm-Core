package state_native_test

import (
	"testing"

	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func BenchmarkAppendEth1DataVotes(b *testing.B) {
	st, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{})
	require.NoError(b, err)

	max := params.BeaconConfig().Eth1DataVotesLength()

	if max < 2 {
		b.Fatalf("Eth1DataVotesLength is less than 2")
	}

	for i := uint64(0); i < max-2; i++ {
		err := st.AppendEth1DataVotes(&silapb.Eth1Data{
			DepositCount: i,
			DepositRoot:  make([]byte, 64),
			BlockHash:    make([]byte, 64),
		})
		require.NoError(b, err)
	}

	ref := st.Copy()

	for i := 0; b.Loop(); i++ {
		err := ref.AppendEth1DataVotes(&silapb.Eth1Data{DepositCount: uint64(i)})
		require.NoError(b, err)
		ref = st.Copy()
	}
}
