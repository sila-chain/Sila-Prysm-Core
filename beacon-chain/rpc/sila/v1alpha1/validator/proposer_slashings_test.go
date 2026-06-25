package validator

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/slashings"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
)

func TestServer_getSlashings(t *testing.T) {
	beaconState, privKeys := util.DeterministicGenesisState(t, 64)

	proposerServer := &Server{
		SlashingsPool: slashings.NewPool(),
	}

	proposerSlashings := make([]*silapb.ProposerSlashing, params.BeaconConfig().MaxProposerSlashings)
	for i := primitives.ValidatorIndex(0); uint64(i) < params.BeaconConfig().MaxProposerSlashings; i++ {
		proposerSlashing, err := util.GenerateProposerSlashingForValidator(beaconState, privKeys[i], i)
		require.NoError(t, err)
		proposerSlashings[i] = proposerSlashing
		err = proposerServer.SlashingsPool.InsertProposerSlashing(t.Context(), beaconState, proposerSlashing)
		require.NoError(t, err)
	}

	attSlashings := make([]silapb.AttSlashing, params.BeaconConfig().MaxAttesterSlashings)
	for i := uint64(0); i < params.BeaconConfig().MaxAttesterSlashings; i++ {
		attesterSlashing, err := util.GenerateAttesterSlashingForValidator(
			beaconState,
			privKeys[i+params.BeaconConfig().MaxProposerSlashings],
			primitives.ValidatorIndex(i+params.BeaconConfig().MaxProposerSlashings),
		)
		require.NoError(t, err)
		attSlashings[i] = attesterSlashing
		err = proposerServer.SlashingsPool.InsertAttesterSlashing(t.Context(), beaconState, attesterSlashing)
		require.NoError(t, err)
	}

	p, a := proposerServer.getSlashings(t.Context(), beaconState)
	require.Equal(t, len(p), int(params.BeaconConfig().MaxProposerSlashings))
	require.Equal(t, len(a), int(params.BeaconConfig().MaxAttesterSlashings))
	require.DeepEqual(t, p, proposerSlashings)
	require.DeepEqual(t, a, attSlashings)
}
