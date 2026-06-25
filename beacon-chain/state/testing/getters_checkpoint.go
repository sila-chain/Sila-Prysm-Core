package testing

import (
	"testing"

	"github.com/sila-chain/go-bitfield"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func VerifyBeaconStateJustificationBitsNil(t *testing.T, factory getState) {
	s, err := factory()
	require.NoError(t, err)
	require.DeepEqual(t, bitfield.Bitvector4{}.Bytes(), s.JustificationBits().Bytes())
}

type getStateWithJustificationBits = func(bitfield.Bitvector4) (state.BeaconState, error)

func VerifyBeaconStateJustificationBits(t *testing.T, factory getStateWithJustificationBits) {
	s, err := factory(bitfield.Bitvector4{1, 2, 3, 4})
	require.NoError(t, err)
	require.DeepEqual(t, bitfield.Bitvector4{1, 2, 3, 4}.Bytes(), s.JustificationBits().Bytes())
}

func VerifyBeaconStatePreviousJustifiedCheckpointNil(t *testing.T, factory getState) {
	s, err := factory()

	require.NoError(t, err)

	checkpoint := s.PreviousJustifiedCheckpoint()
	require.Equal(t, (*silapb.Checkpoint)(nil), checkpoint)
}

type getStateWithCheckpoint = func(checkpoint *silapb.Checkpoint) (state.BeaconState, error)

func VerifyBeaconStatePreviousJustifiedCheckpoint(t *testing.T, factory getStateWithCheckpoint) {
	orgCheckpoint := &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)}
	orgCheckpoint.Root[1] = 1
	orgCheckpoint.Root[2] = 2
	orgCheckpoint.Root[3] = 3
	s, err := factory(orgCheckpoint)

	require.NoError(t, err)

	checkpoint := s.PreviousJustifiedCheckpoint()
	require.DeepEqual(t, orgCheckpoint.Root, checkpoint.Root)
}

func VerifyBeaconStateCurrentJustifiedCheckpointNil(t *testing.T, factory getState) {
	s, err := factory()

	require.NoError(t, err)

	checkpoint := s.CurrentJustifiedCheckpoint()
	require.Equal(t, (*silapb.Checkpoint)(nil), checkpoint)
}

func VerifyBeaconStateCurrentJustifiedCheckpoint(t *testing.T, factory getStateWithCheckpoint) {
	orgCheckpoint := &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)}
	orgCheckpoint.Root[1] = 1
	orgCheckpoint.Root[2] = 2
	orgCheckpoint.Root[3] = 3
	s, err := factory(orgCheckpoint)

	require.NoError(t, err)

	checkpoint := s.CurrentJustifiedCheckpoint()
	require.DeepEqual(t, orgCheckpoint.Root, checkpoint.Root)
}

func VerifyBeaconStateFinalizedCheckpointNil(t *testing.T, factory getState) {
	s, err := factory()

	require.NoError(t, err)

	checkpoint := s.FinalizedCheckpoint()
	require.Equal(t, (*silapb.Checkpoint)(nil), checkpoint)
	epoch := s.FinalizedCheckpointEpoch()
	require.Equal(t, primitives.Epoch(0), epoch)
}

func VerifyBeaconStateFinalizedCheckpoint(t *testing.T, factory getStateWithCheckpoint) {
	orgCheckpoint := &silapb.Checkpoint{Root: make([]byte, fieldparams.RootLength)}
	orgCheckpoint.Root[1] = 1
	orgCheckpoint.Root[2] = 2
	orgCheckpoint.Root[3] = 3
	orgCheckpoint.Epoch = 123
	s, err := factory(orgCheckpoint)

	require.NoError(t, err)

	checkpoint := s.FinalizedCheckpoint()
	require.DeepEqual(t, orgCheckpoint.Root, checkpoint.Root)
	epoch := s.FinalizedCheckpointEpoch()
	require.Equal(t, orgCheckpoint.Epoch, epoch)
}
