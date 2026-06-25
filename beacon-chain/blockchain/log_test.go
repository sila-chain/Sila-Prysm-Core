package blockchain

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

func Test_logStateTransitionData(t *testing.T) {
	payloadBlk := &silapb.BeaconBlockBellatrix{
		Body: &silapb.BeaconBlockBodyBellatrix{
			SyncAggregate: &silapb.SyncAggregate{},
			SilaPayload: &silaenginev1.SilaPayload{
				BlockHash:    []byte{1, 2, 3},
				Transactions: [][]byte{{}, {}},
			},
		},
	}
	wrappedPayloadBlk, err := blocks.NewBeaconBlock(payloadBlk)
	require.NoError(t, err)
	tests := []struct {
		name string
		b    func() interfaces.ReadOnlyBeaconBlock
		want string
	}{
		{name: "empty block body",
			b: func() interfaces.ReadOnlyBeaconBlock {
				wb, err := blocks.NewBeaconBlock(&silapb.BeaconBlock{Body: &silapb.BeaconBlockBody{}})
				require.NoError(t, err)
				return wb
			},
			want: "\"Finished applying state transition\" package=beacon-chain/blockchain slot=0",
		},
		{name: "has attestation",
			b: func() interfaces.ReadOnlyBeaconBlock {
				wb, err := blocks.NewBeaconBlock(&silapb.BeaconBlock{Body: &silapb.BeaconBlockBody{Attestations: []*silapb.Attestation{{}}}})
				require.NoError(t, err)
				return wb
			},
			want: "\"Finished applying state transition\" attestations=1 package=beacon-chain/blockchain slot=0",
		},
		{name: "has deposit",
			b: func() interfaces.ReadOnlyBeaconBlock {
				wb, err := blocks.NewBeaconBlock(
					&silapb.BeaconBlock{Body: &silapb.BeaconBlockBody{
						Attestations: []*silapb.Attestation{{}},
						Deposits:     []*silapb.Deposit{{}}}})
				require.NoError(t, err)
				return wb
			},
			want: "\"Finished applying state transition\" attestations=1 package=beacon-chain/blockchain slot=0",
		},
		{name: "has attester slashing",
			b: func() interfaces.ReadOnlyBeaconBlock {
				wb, err := blocks.NewBeaconBlock(&silapb.BeaconBlock{Body: &silapb.BeaconBlockBody{
					AttesterSlashings: []*silapb.AttesterSlashing{{}}}})
				require.NoError(t, err)
				return wb
			},
			want: "\"Finished applying state transition\" attesterSlashings=1 package=beacon-chain/blockchain slot=0",
		},
		{name: "has proposer slashing",
			b: func() interfaces.ReadOnlyBeaconBlock {
				wb, err := blocks.NewBeaconBlock(&silapb.BeaconBlock{Body: &silapb.BeaconBlockBody{
					ProposerSlashings: []*silapb.ProposerSlashing{{}}}})
				require.NoError(t, err)
				return wb
			},
			want: "\"Finished applying state transition\" package=beacon-chain/blockchain proposerSlashings=1 slot=0",
		},
		{name: "has exit",
			b: func() interfaces.ReadOnlyBeaconBlock {
				wb, err := blocks.NewBeaconBlock(&silapb.BeaconBlock{Body: &silapb.BeaconBlockBody{
					VoluntaryExits: []*silapb.SignedVoluntaryExit{{}}}})
				require.NoError(t, err)
				return wb
			},
			want: "\"Finished applying state transition\" package=beacon-chain/blockchain slot=0 voluntaryExits=1",
		},
		{name: "has everything",
			b: func() interfaces.ReadOnlyBeaconBlock {
				wb, err := blocks.NewBeaconBlock(&silapb.BeaconBlock{Body: &silapb.BeaconBlockBody{
					Attestations:      []*silapb.Attestation{{}},
					Deposits:          []*silapb.Deposit{{}},
					AttesterSlashings: []*silapb.AttesterSlashing{{}},
					ProposerSlashings: []*silapb.ProposerSlashing{{}},
					VoluntaryExits:    []*silapb.SignedVoluntaryExit{{}}}})
				require.NoError(t, err)
				return wb
			},
			want: "\"Finished applying state transition\" attestations=1 attesterSlashings=1 package=beacon-chain/blockchain proposerSlashings=1 slot=0 voluntaryExits=1",
		},
		{name: "has payload",
			b:    func() interfaces.ReadOnlyBeaconBlock { return wrappedPayloadBlk },
			want: "\"Finished applying state transition\" package=beacon-chain/blockchain payloadHash=0x010203 slot=0 syncBitsCount=0 txCount=2",
		},
	}
	for _, tt := range tests {
		hook := logTest.NewGlobal()
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, logStateTransitionData(tt.b()))
			require.LogsContain(t, hook, tt.want)
		})
	}
}
