package validator

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
)

func Test_getEmptyBlock(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	config := params.BeaconConfig()
	config.AltairForkEpoch = 1
	config.BellatrixForkEpoch = 2
	config.CapellaForkEpoch = 3
	config.DenebForkEpoch = 4
	config.ElectraForkEpoch = 5
	config.FuluForkEpoch = 6
	params.OverrideBeaconConfig(config)

	tests := []struct {
		name string
		slot primitives.Slot
		want func() interfaces.ReadOnlySignedBeaconBlock
	}{
		{
			name: "altair",
			slot: primitives.Slot(params.BeaconConfig().AltairForkEpoch) * params.BeaconConfig().SlotsPerEpoch,
			want: func() interfaces.ReadOnlySignedBeaconBlock {
				b, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockAltair{Block: &silapb.BeaconBlockAltair{Body: &silapb.BeaconBlockBodyAltair{}}})
				require.NoError(t, err)
				return b
			},
		},
		{
			name: "bellatrix",
			slot: primitives.Slot(params.BeaconConfig().BellatrixForkEpoch) * params.BeaconConfig().SlotsPerEpoch,
			want: func() interfaces.ReadOnlySignedBeaconBlock {
				b, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockBellatrix{Block: &silapb.BeaconBlockBellatrix{Body: &silapb.BeaconBlockBodyBellatrix{}}})
				require.NoError(t, err)
				return b
			},
		},
		{
			name: "capella",
			slot: primitives.Slot(params.BeaconConfig().CapellaForkEpoch) * params.BeaconConfig().SlotsPerEpoch,
			want: func() interfaces.ReadOnlySignedBeaconBlock {
				b, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockCapella{Block: &silapb.BeaconBlockCapella{Body: &silapb.BeaconBlockBodyCapella{}}})
				require.NoError(t, err)
				return b
			},
		},
		{
			name: "deneb",
			slot: primitives.Slot(params.BeaconConfig().DenebForkEpoch) * params.BeaconConfig().SlotsPerEpoch,
			want: func() interfaces.ReadOnlySignedBeaconBlock {
				b, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockDeneb{Block: &silapb.BeaconBlockDeneb{Body: &silapb.BeaconBlockBodyDeneb{}}})
				require.NoError(t, err)
				return b
			},
		},
		{
			name: "electra",
			slot: primitives.Slot(params.BeaconConfig().ElectraForkEpoch) * params.BeaconConfig().SlotsPerEpoch,
			want: func() interfaces.ReadOnlySignedBeaconBlock {
				b, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockElectra{Block: &silapb.BeaconBlockElectra{Body: &silapb.BeaconBlockBodyElectra{}}})
				require.NoError(t, err)
				return b
			},
		},
		{
			name: "fulu",
			slot: primitives.Slot(params.BeaconConfig().FuluForkEpoch) * params.BeaconConfig().SlotsPerEpoch,
			want: func() interfaces.ReadOnlySignedBeaconBlock {
				b, err := blocks.NewSignedBeaconBlock(&silapb.SignedBeaconBlockFulu{Block: &silapb.BeaconBlockElectra{Body: &silapb.BeaconBlockBodyElectra{}}})
				require.NoError(t, err)
				return b
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getEmptyBlock(tt.slot)
			require.NoError(t, err)
			require.DeepEqual(t, tt.want(), got, "getEmptyBlock() = %v, want %v", got, tt.want())
		})
	}
}
