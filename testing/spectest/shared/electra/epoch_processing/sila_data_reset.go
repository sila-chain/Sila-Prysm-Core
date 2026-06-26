package epoch_processing

import (
	"path"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/electra"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/utils"
)

// RunSilaDataResetTests executes "epoch_processing/sila_data_reset" tests.
func RunSilaDataResetTests(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))

	testFolders, testsFolderPath := utils.TestFolders(t, config, "electra", "epoch_processing/sila_data_reset/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			RunEpochOperationTest(t, folderPath, processSilaDataResetWrapper)
		})
	}
}

func processSilaDataResetWrapper(t *testing.T, st state.BeaconState) (state.BeaconState, error) {
	st, err := electra.ProcessSilaDataReset(st)
	require.NoError(t, err, "Could not process final updates")
	return st, nil
}
