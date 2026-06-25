package state_native

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native/types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func TestBeaconState_RotateAttestations(t *testing.T) {
	st, err := InitializeFromProtoPhase0(&silapb.BeaconState{
		Slot:                      1,
		CurrentEpochAttestations:  []*silapb.PendingAttestation{{Data: &silapb.AttestationData{Slot: 456}}},
		PreviousEpochAttestations: []*silapb.PendingAttestation{{Data: &silapb.AttestationData{Slot: 123}}},
	})
	require.NoError(t, err)

	require.NoError(t, st.RotateAttestations())
	s, ok := st.(*BeaconState)
	require.Equal(t, true, ok)
	require.Equal(t, 0, len(s.currentEpochAttestationsVal()))
	require.Equal(t, primitives.Slot(456), s.previousEpochAttestationsVal()[0].Data.Slot)
}

func TestAppendBeyondIndicesLimit(t *testing.T) {
	zeroHash := params.BeaconConfig().ZeroHash
	mockblockRoots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := range mockblockRoots {
		mockblockRoots[i] = zeroHash[:]
	}

	mockstateRoots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := range mockstateRoots {
		mockstateRoots[i] = zeroHash[:]
	}
	mockrandaoMixes := make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector)
	for i := range mockrandaoMixes {
		mockrandaoMixes[i] = zeroHash[:]
	}
	st, err := InitializeFromProtoPhase0(&silapb.BeaconState{
		Slot:                      1,
		CurrentEpochAttestations:  []*silapb.PendingAttestation{{Data: &silapb.AttestationData{Slot: 456}}},
		PreviousEpochAttestations: []*silapb.PendingAttestation{{Data: &silapb.AttestationData{Slot: 123}}},
		Validators:                []*silapb.Validator{},
		Eth1Data:                  &silapb.Eth1Data{},
		BlockRoots:                mockblockRoots,
		StateRoots:                mockstateRoots,
		RandaoMixes:               mockrandaoMixes,
	})
	require.NoError(t, err)
	_, err = st.HashTreeRoot(t.Context())
	require.NoError(t, err)
	s, ok := st.(*BeaconState)
	require.Equal(t, true, ok)
	for i := types.FieldIndex(0); i < types.FieldIndex(params.BeaconConfig().BeaconStateFieldCount); i++ {
		s.dirtyFields[i] = true
	}
	_, err = st.HashTreeRoot(t.Context())
	require.NoError(t, err)
	for range 10 {
		assert.NoError(t, st.AppendValidator(&silapb.Validator{}))
	}
	assert.Equal(t, false, s.rebuildTrie[types.Validators])
	assert.NotEqual(t, len(s.dirtyIndices[types.Validators]), 0)

	for range indicesLimit {
		assert.NoError(t, st.AppendValidator(&silapb.Validator{}))
	}
	assert.Equal(t, true, s.rebuildTrie[types.Validators])
	assert.Equal(t, len(s.dirtyIndices[types.Validators]), 0)
}

func BenchmarkAppendPreviousEpochAttestations(b *testing.B) {
	st, err := InitializeFromProtoPhase0(&silapb.BeaconState{})
	require.NoError(b, err)

	max := params.BeaconConfig().PreviousEpochAttestationsLength()
	if max < 2 {
		b.Fatalf("previous epoch attestations length is less than 2: %d", max)
	}

	for i := uint64(0); i < max-2; i++ {
		err := st.AppendPreviousEpochAttestations(&silapb.PendingAttestation{Data: &silapb.AttestationData{Slot: primitives.Slot(i)}})
		require.NoError(b, err)
	}

	ref := st.Copy()
	for i := 0; b.Loop(); i++ {
		err := ref.AppendPreviousEpochAttestations(&silapb.PendingAttestation{Data: &silapb.AttestationData{Slot: primitives.Slot(i)}})
		require.NoError(b, err)
		ref = st.Copy()
	}
}
