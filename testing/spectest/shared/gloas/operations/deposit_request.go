package operations

import (
	"context"
	"path"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/gloas"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	common "github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/shared/common/operations"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/spectest/utils"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	"github.com/golang/snappy"
)

func blockWithDepositRequest(ssz []byte) (interfaces.SignedBeaconBlock, error) {
	dr := &silaenginev1.DepositRequest{}
	if err := dr.UnmarshalSSZ(ssz); err != nil {
		return nil, err
	}
	er := &silaenginev1.ExecutionRequests{
		Deposits: []*silaenginev1.DepositRequest{dr},
	}
	b := util.NewBeaconBlockElectra()
	b.Block.Body = &silapb.BeaconBlockBodyElectra{ExecutionRequests: er}
	return blocks.NewSignedBeaconBlock(b)
}

func RunDepositRequestsTest(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))
	testFolders, testsFolderPath := utils.TestFolders(t, config, version.String(version.Gloas), "operations/deposit_request/pyspec_tests")
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			drFile, err := util.BazelFileBytes(folderPath, "deposit_request.ssz_snappy")
			require.NoError(t, err)
			drSSZ, err := snappy.Decode(nil /* dst */, drFile)
			require.NoError(t, err, "Failed to decompress")
			blk, err := blockWithDepositRequest(drSSZ)
			require.NoError(t, err)
			common.RunBlockOperationTest(t, folderPath, blk, sszToState, func(ctx context.Context, s state.BeaconState, b interfaces.ReadOnlySignedBeaconBlock) (state.BeaconState, error) {
				e, err := b.Block().Body().ExecutionRequests()
				if err != nil {
					return nil, err
				}
				return s, gloas.ProcessDepositRequests(ctx, s, e.Deposits, nil)
			})
		})
	}
}
