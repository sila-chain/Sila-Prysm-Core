package p2p

import (
	"context"
	"testing"

	iface "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db/iface"
	dbutil "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db/testing"
	mockstategen "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stategen/mock"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func TestCorrect_ActiveValidatorsCount(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.MainnetConfig()
	cfg.ConfigName = "test"

	params.OverrideBeaconConfig(cfg)

	db := dbutil.SetupDB(t)
	wrappedDB := &finalizedCheckpointDB{ReadOnlyDatabaseWithSeqNum: db}
	stateGen := mockstategen.NewService()
	s := &Service{
		ctx: t.Context(),
		cfg: &Config{DB: wrappedDB, StateGen: stateGen},
	}
	bState, err := util.NewBeaconState(func(state *silapb.BeaconState) error {
		validators := make([]*silapb.Validator, params.BeaconConfig().MinGenesisActiveValidatorCount)
		for i := range validators {
			validators[i] = &silapb.Validator{
				PublicKey:             make([]byte, 48),
				WithdrawalCredentials: make([]byte, 32),
				ExitEpoch:             params.BeaconConfig().FarFutureEpoch,
				Slashed:               false,
			}
		}
		state.Validators = validators
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, db.SaveGenesisData(s.ctx, bState))
	checkpoint, err := db.FinalizedCheckpoint(s.ctx)
	require.NoError(t, err)
	wrappedDB.finalized = checkpoint
	stateGen.AddStateForRoot(bState, bytesutil.ToBytes32(checkpoint.Root))

	vals, err := s.retrieveActiveValidators()
	assert.NoError(t, err, "genesis state not retrieved")
	assert.Equal(t, int(params.BeaconConfig().MinGenesisActiveValidatorCount), int(vals), "mainnet genesis active count isn't accurate")
	for range 100 {
		require.NoError(t, bState.AppendValidator(&silapb.Validator{
			PublicKey:             make([]byte, 48),
			WithdrawalCredentials: make([]byte, 32),
			ExitEpoch:             params.BeaconConfig().FarFutureEpoch,
			Slashed:               false,
		}))
	}
	require.NoError(t, bState.SetSlot(10000))
	rootA := [32]byte{'a'}
	require.NoError(t, db.SaveState(s.ctx, bState, rootA))
	wrappedDB.finalized = &silapb.Checkpoint{Root: rootA[:]}
	stateGen.AddStateForRoot(bState, rootA)
	// Reset count
	s.activeValidatorCount = 0

	// Retrieve last archived state.
	vals, err = s.retrieveActiveValidators()
	assert.NoError(t, err, "genesis state not retrieved")
	assert.Equal(t, int(params.BeaconConfig().MinGenesisActiveValidatorCount)+100, int(vals), "mainnet genesis active count isn't accurate")
}

func TestLoggingParameters(_ *testing.T) {
	logGossipParameters("testing", nil)
	logGossipParameters("testing", &pubsub.TopicScoreParams{})
	// Test out actual gossip parameters.
	logGossipParameters("testing", defaultBlockTopicParams())
	p := defaultAggregateSubnetTopicParams(10000)
	logGossipParameters("testing", p)
	p = defaultAggregateTopicParams(10000)
	logGossipParameters("testing", p)
	logGossipParameters("testing", defaultAttesterSlashingTopicParams())
	logGossipParameters("testing", defaultProposerSlashingTopicParams())
	logGossipParameters("testing", defaultVoluntaryExitTopicParams())
	logGossipParameters("testing", defaultLightClientOptimisticUpdateTopicParams())
	logGossipParameters("testing", defaultLightClientFinalityUpdateTopicParams())
}

type finalizedCheckpointDB struct {
	iface.ReadOnlyDatabaseWithSeqNum
	finalized *silapb.Checkpoint
}

func (f *finalizedCheckpointDB) FinalizedCheckpoint(ctx context.Context) (*silapb.Checkpoint, error) {
	if f.finalized != nil {
		return f.finalized, nil
	}
	return f.ReadOnlyDatabaseWithSeqNum.FinalizedCheckpoint(ctx)
}
