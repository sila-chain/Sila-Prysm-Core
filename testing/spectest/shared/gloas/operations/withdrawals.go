package operations

import (
	"context"
	"path"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/gloas"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	common "github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/common/operations"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/utils"
)

func emptyBlockGloas() (interfaces.SignedBeaconBlock, error) {
	b := &silapb.SignedBeaconBlockGloas{
		Block: &silapb.BeaconBlockGloas{
			Body: &silapb.BeaconBlockBodyGloas{},
		},
	}
	return blocks.NewSignedBeaconBlock(b)
}

func RunWithdrawalsTest(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, testsFolderPath := utils.TestFolders(t, config, version.String(version.Gloas), "operations/withdrawals/pyspec_tests")
	if len(testFolders) == 0 {
		t.Fatalf("No test folders found for %s/%s/%s", config, version.String(version.Gloas), "operations/withdrawals/pyspec_tests")
	}
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			blk, err := emptyBlockGloas()
			require.NoError(t, err)

			common.RunBlockOperationTest(t, folderPath, blk, sszToState, func(_ context.Context, s state.BeaconState, _ interfaces.ReadOnlySignedBeaconBlock) (state.BeaconState, error) {
				if err := gloas.ProcessWithdrawals(s); err != nil {
					return nil, err
				}
				return s, nil
			})
		})
	}
}
