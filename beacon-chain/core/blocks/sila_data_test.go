package blocks_test

import (
	"fmt"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/blocks"
	state_native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	"google.golang.org/protobuf/proto"
)

func FakeDeposits(n uint64) []*silapb.SilaData {
	deposits := make([]*silapb.SilaData, n)
	for i := range n {
		deposits[i] = &silapb.SilaData{
			DepositCount: 1,
			DepositRoot:  bytesutil.PadTo([]byte("root"), 32),
		}
	}
	return deposits
}

func TestSilaDataHasEnoughSupport(t *testing.T) {
	tests := []struct {
		stateVotes         []*silapb.SilaData
		data               *silapb.SilaData
		hasSupport         bool
		votingPeriodLength primitives.Epoch
	}{
		{
			stateVotes: FakeDeposits(uint64(params.BeaconConfig().SlotsPerEpoch.Mul(4))),
			data: &silapb.SilaData{
				DepositCount: 1,
				DepositRoot:  bytesutil.PadTo([]byte("root"), 32),
			},
			hasSupport:         true,
			votingPeriodLength: 7,
		}, {
			stateVotes: FakeDeposits(uint64(params.BeaconConfig().SlotsPerEpoch.Mul(4))),
			data: &silapb.SilaData{
				DepositCount: 1,
				DepositRoot:  bytesutil.PadTo([]byte("root"), 32),
			},
			hasSupport:         false,
			votingPeriodLength: 8,
		}, {
			stateVotes: FakeDeposits(uint64(params.BeaconConfig().SlotsPerEpoch.Mul(4))),
			data: &silapb.SilaData{
				DepositCount: 1,
				DepositRoot:  bytesutil.PadTo([]byte("root"), 32),
			},
			hasSupport:         false,
			votingPeriodLength: 10,
		},
	}

	params.SetupTestConfigCleanup(t)
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			c := params.BeaconConfig()
			c.EpochsPerSilaExecutionVotingPeriod = tt.votingPeriodLength
			params.OverrideBeaconConfig(c)

			s, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
				SilaDataVotes: tt.stateVotes,
			})
			require.NoError(t, err)
			result, err := blocks.SilaDataHasEnoughSupport(s, tt.data)
			require.NoError(t, err)

			if result != tt.hasSupport {
				t.Errorf(
					"blocks.SilaDataHasEnoughSupport(%+v) = %t, wanted %t",
					tt.data,
					result,
					tt.hasSupport,
				)
			}
		})
	}
}

func TestAreSilaDataEqual(t *testing.T) {
	type args struct {
		a *silapb.SilaData
		b *silapb.SilaData
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "true when both are nil",
			args: args{
				a: nil,
				b: nil,
			},
			want: true,
		},
		{
			name: "false when only one is nil",
			args: args{
				a: nil,
				b: &silapb.SilaData{
					DepositRoot:  make([]byte, 32),
					DepositCount: 0,
					BlockHash:    make([]byte, 32),
				},
			},
			want: false,
		},
		{
			name: "true when real equality",
			args: args{
				a: &silapb.SilaData{
					DepositRoot:  make([]byte, 32),
					DepositCount: 0,
					BlockHash:    make([]byte, 32),
				},
				b: &silapb.SilaData{
					DepositRoot:  make([]byte, 32),
					DepositCount: 0,
					BlockHash:    make([]byte, 32),
				},
			},
			want: true,
		},
		{
			name: "false is field value differs",
			args: args{
				a: &silapb.SilaData{
					DepositRoot:  make([]byte, 32),
					DepositCount: 0,
					BlockHash:    make([]byte, 32),
				},
				b: &silapb.SilaData{
					DepositRoot:  make([]byte, 32),
					DepositCount: 64,
					BlockHash:    make([]byte, 32),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, blocks.AreSilaDataEqual(tt.args.a, tt.args.b))
		})
	}
}

func TestProcessSilaData_SetsCorrectly(t *testing.T) {
	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		SilaDataVotes: []*silapb.SilaData{},
	})
	require.NoError(t, err)

	b := util.NewBeaconBlock()
	b.Block = &silapb.BeaconBlock{
		Body: &silapb.BeaconBlockBody{
			SilaData: &silapb.SilaData{
				DepositRoot: []byte{2},
				BlockHash:   []byte{3},
			},
		},
	}

	period := uint64(params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().EpochsPerSilaExecutionVotingPeriod)))
	for range period {
		processedState, err := blocks.ProcessSilaDataInBlock(t.Context(), beaconState, b.Block.Body.SilaData)
		require.NoError(t, err)
		require.Equal(t, true, processedState.Version() == version.Phase0)
	}

	newSilaDataVotes := beaconState.SilaDataVotes()
	if len(newSilaDataVotes) <= 1 {
		t.Error("Expected new SILAEXEC data votes to have length > 1")
	}
	if !proto.Equal(beaconState.SilaData(), b.Block.Body.SilaData.Copy()) {
		t.Errorf(
			"Expected latest silaexec data to have been set to %v, received %v",
			b.Block.Body.SilaData,
			beaconState.SilaData(),
		)
	}
}
