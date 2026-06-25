package operations

import (
	"context"
	"path"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/utils"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	"github.com/golang/snappy"
	"github.com/pkg/errors"
)

func RunBLSToSilaChangeTest(t *testing.T, config string, fork string, block blockWithSSZObject, sszToState SSZToState) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, testsFolderPath := utils.TestFolders(t, config, fork, "operations/bls_to_sila_change/pyspec_tests")
	if len(testFolders) == 0 {
		t.Fatalf("No test folders found for %s/%s/%s", config, fork, "operations/bls_to_sila_change/pyspec_tests")
	}
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			changeFile, err := util.BazelFileBytes(folderPath, "address_change.ssz_snappy")
			require.NoError(t, err)
			changeSSZ, err := snappy.Decode(nil /* dst */, changeFile)
			require.NoError(t, err, "Failed to decompress")
			blk, err := block(changeSSZ)
			require.NoError(t, err)
			RunBlockOperationTest(t, folderPath, blk, sszToState, func(ctx context.Context, s state.BeaconState, b interfaces.ReadOnlySignedBeaconBlock) (state.BeaconState, error) {
				st, err := blocks.ProcessBLSToSilaChanges(s, b.Block())
				if err != nil {
					return nil, err
				}
				changes, err := b.Block().Body().BLSToSilaChanges()
				if err != nil {
					return nil, err
				}
				cSet, err := blocks.BLSChangesSignatureBatch(st, changes)
				if err != nil {
					return nil, err
				}
				ok, err := cSet.Verify()
				if err != nil {
					return nil, err
				}
				if !ok {
					return nil, errors.New("signature did not verify")
				}
				return st, nil
			})
		})
	}
}
