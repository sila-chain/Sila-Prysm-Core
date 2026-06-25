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

func FakeDeposits(n uint64) []*silapb.Eth1Data {
	deposits := make([]*silapb.Eth1Data, n)
	for i := range n {
		deposits[i] = &silapb.Eth1Data{
			DepositCount: 1,
			DepositRoot:  bytesutil.PadTo([]byte("root"), 32),
		}
	}
	return deposits
}

func TestEth1DataHasEnoughSupport(t *testing.T) {
	tests := []struct {
		stateVotes         []*silapb.Eth1Data
		data               *silapb.Eth1Data
		hasSupport         bool
		votingPeriodLength primitives.Epoch
	}{
		{
			stateVotes: FakeDeposits(uint64(params.BeaconConfig().SlotsPerEpoch.Mul(4))),
			data: &silapb.Eth1Data{
				DepositCount: 1,
				DepositRoot:  bytesutil.PadTo([]byte("root"), 32),
			},
			hasSupport:         true,
			votingPeriodLength: 7,
		}, {
			stateVotes: FakeDeposits(uint64(params.BeaconConfig().SlotsPerEpoch.Mul(4))),
			data: &silapb.Eth1Data{
				DepositCount: 1,
				DepositRoot:  bytesutil.PadTo([]byte("root"), 32),
			},
			hasSupport:         false,
			votingPeriodLength: 8,
		}, {
			stateVotes: FakeDeposits(uint64(params.BeaconConfig().SlotsPerEpoch.Mul(4))),
			data: &silapb.Eth1Data{
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
			c.EpochsPerEth1VotingPeriod = tt.votingPeriodLength
			params.OverrideBeaconConfig(c)

			s, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
				Eth1DataVotes: tt.stateVotes,
			})
			require.NoError(t, err)
			result, err := blocks.Eth1DataHasEnoughSupport(s, tt.data)
			require.NoError(t, err)

			if result != tt.hasSupport {
				t.Errorf(
					"blocks.Eth1DataHasEnoughSupport(%+v) = %t, wanted %t",
					tt.data,
					result,
					tt.hasSupport,
				)
			}
		})
	}
}

func TestAreEth1DataEqual(t *testing.T) {
	type args struct {
		a *silapb.Eth1Data
		b *silapb.Eth1Data
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
				b: &silapb.Eth1Data{
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
				a: &silapb.Eth1Data{
					DepositRoot:  make([]byte, 32),
					DepositCount: 0,
					BlockHash:    make([]byte, 32),
				},
				b: &silapb.Eth1Data{
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
				a: &silapb.Eth1Data{
					DepositRoot:  make([]byte, 32),
					DepositCount: 0,
					BlockHash:    make([]byte, 32),
				},
				b: &silapb.Eth1Data{
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
			assert.Equal(t, tt.want, blocks.AreEth1DataEqual(tt.args.a, tt.args.b))
		})
	}
}

func TestProcessEth1Data_SetsCorrectly(t *testing.T) {
	beaconState, err := state_native.InitializeFromProtoPhase0(&silapb.BeaconState{
		Eth1DataVotes: []*silapb.Eth1Data{},
	})
	require.NoError(t, err)

	b := util.NewBeaconBlock()
	b.Block = &silapb.BeaconBlock{
		Body: &silapb.BeaconBlockBody{
			Eth1Data: &silapb.Eth1Data{
				DepositRoot: []byte{2},
				BlockHash:   []byte{3},
			},
		},
	}

	period := uint64(params.BeaconConfig().SlotsPerEpoch.Mul(uint64(params.BeaconConfig().EpochsPerEth1VotingPeriod)))
	for range period {
		processedState, err := blocks.ProcessEth1DataInBlock(t.Context(), beaconState, b.Block.Body.Eth1Data)
		require.NoError(t, err)
		require.Equal(t, true, processedState.Version() == version.Phase0)
	}

	newETH1DataVotes := beaconState.Eth1DataVotes()
	if len(newETH1DataVotes) <= 1 {
		t.Error("Expected new ETH1 data votes to have length > 1")
	}
	if !proto.Equal(beaconState.Eth1Data(), b.Block.Body.Eth1Data.Copy()) {
		t.Errorf(
			"Expected latest eth1 data to have been set to %v, received %v",
			b.Block.Body.Eth1Data,
			beaconState.Eth1Data(),
		)
	}
}
