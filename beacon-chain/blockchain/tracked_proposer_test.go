package blockchain

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/features"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	"github.com/sila-chain/Sila/common"
)

func TestTrackedProposer_NotTracked(t *testing.T) {
	service, _ := minimalTestService(t, WithPayloadIDCache(cache.NewPayloadIDCache()))
	st, _ := util.DeterministicGenesisStateBellatrix(t, 1)
	_, ok := service.trackedProposer(st, 0)
	require.Equal(t, false, ok)
}

func TestTrackedProposer_Tracked(t *testing.T) {
	service, _ := minimalTestService(t, WithPayloadIDCache(cache.NewPayloadIDCache()))
	st, _ := util.DeterministicGenesisStateBellatrix(t, 1)
	addr := common.HexToAddress("0x1234")
	service.cfg.TrackedValidatorsCache.Set(cache.TrackedValidator{Active: true, FeeRecipient: primitives.SilaAddress(addr), Index: 0})
	val, ok := service.trackedProposer(st, 0)
	require.Equal(t, true, ok)
	require.Equal(t, primitives.SilaAddress(addr), val.FeeRecipient)
}

func TestTrackedProposer_PrepareAllPayloads_Default(t *testing.T) {
	resetCfg := features.InitWithReset(&features.Flags{PrepareAllPayloads: true})
	defer resetCfg()

	service, _ := minimalTestService(t, WithPayloadIDCache(cache.NewPayloadIDCache()))
	st, _ := util.DeterministicGenesisStateBellatrix(t, 1)
	val, ok := service.trackedProposer(st, 0)
	require.Equal(t, true, ok)
	require.Equal(t, true, val.Active)
	require.Equal(t, params.BeaconConfig().SilaBurnAddressHex, common.BytesToAddress(val.FeeRecipient[:]).String())
}
