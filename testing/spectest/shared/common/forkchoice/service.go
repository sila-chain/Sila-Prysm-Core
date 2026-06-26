package forkchoice

import (
	"context"
	"math/big"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain/kzg"
	mock "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/blockchain/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache/depositsnapshot"
	coreTime "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/time"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db/filesystem"
	testDB "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/forkchoice"
	doublylinkedtree "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/forkchoice/doubly-linked-tree"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/operations/attestations"
	p2pTesting "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/p2p/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/startup"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stategen"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	payloadattribute "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/payload-attribute"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	pb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila/common"
	"github.com/sila-chain/Sila/common/hexutil"
	silaTypes "github.com/sila-chain/Sila/core/types"
)

func startChainService(t testing.TB,
	st state.BeaconState,
	block interfaces.ReadOnlySignedBeaconBlock,
	engineMock *engineMock,
	clockSync *startup.ClockSynchronizer,
) (*blockchain.Service, *stategen.State, forkchoice.ForkChoicer) {
	ctx := context.Background()
	db := testDB.SetupDB(t)
	require.NoError(t, db.SaveBlock(ctx, block))
	r, err := block.Block().HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, db.SaveGenesisBlockRoot(ctx, r))

	cp := &silapb.Checkpoint{
		Epoch: coreTime.CurrentEpoch(st),
		Root:  r[:],
	}
	require.NoError(t, db.SaveState(ctx, st, r))
	require.NoError(t, db.SaveJustifiedCheckpoint(ctx, cp))
	require.NoError(t, db.SaveFinalizedCheckpoint(ctx, cp))
	attPool, err := attestations.NewService(ctx, &attestations.Config{
		Pool: attestations.NewPool(),
	})
	require.NoError(t, err)

	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)

	genesis := st.GenesisTime()

	fc := doublylinkedtree.New()
	fc.SetGenesisTime(genesis)
	sg := stategen.New(db, fc)
	opts := append([]blockchain.Option{},
		blockchain.WithSilaEngineCaller(engineMock),
		blockchain.WithFinalizedStateAtStartUp(st),
		blockchain.WithDatabase(db),
		blockchain.WithAttestationService(attPool),
		blockchain.WithForkChoiceStore(fc),
		blockchain.WithStateGen(sg),
		blockchain.WithStateNotifier(&mock.MockStateNotifier{}),
		blockchain.WithAttestationPool(attestations.NewPool()),
		blockchain.WithDepositCache(depositCache),
		blockchain.WithTrackedValidatorsCache(cache.NewTrackedValidatorsCache()),
		blockchain.WithPayloadIDCache(cache.NewPayloadIDCache()),
		blockchain.WithClockSynchronizer(clockSync),
		blockchain.WithBlobStorage(filesystem.NewEphemeralBlobStorage(t)),
		blockchain.WithDataColumnStorage(filesystem.NewEphemeralDataColumnStorage(t)),
		blockchain.WithSyncChecker(mock.MockChecker{}),
		blockchain.WithGenesisTime(genesis),
		blockchain.WithP2PBroadcaster(&p2pTesting.TestP2P{}),
	)
	service, err := blockchain.NewService(t.Context(), opts...)
	require.NoError(t, err)
	// force start kzg context here until Deneb fork epoch is decided
	require.NoError(t, kzg.Start())
	require.NoError(t, service.StartFromSavedState(st))
	return service, sg, fc
}

type engineMock struct {
	powBlocks       map[[32]byte]*silapb.PowBlock
	latestValidHash []byte
	payloadStatus   error
}

func (m *engineMock) GetPayload(context.Context, [8]byte, primitives.Slot) (*blocks.GetPayloadResponse, error) {
	return nil, nil
}
func (m *engineMock) GetPayloadV2(context.Context, [8]byte) (*pb.SilaPayloadCapella, error) {
	return nil, nil
}
func (m *engineMock) ForkchoiceUpdated(context.Context, *pb.ForkchoiceState, payloadattribute.Attributer) (*pb.PayloadIDBytes, []byte, error) {
	return nil, m.latestValidHash, m.payloadStatus
}

func (m *engineMock) NewPayload(context.Context, interfaces.SilaData, []common.Hash, *common.Hash, *pb.SilaRequests) ([]byte, error) {
	return m.latestValidHash, m.payloadStatus
}

func (m *engineMock) ForkchoiceUpdatedV2(context.Context, *pb.ForkchoiceState, payloadattribute.Attributer) (*pb.PayloadIDBytes, []byte, error) {
	return nil, m.latestValidHash, m.payloadStatus
}

func (m *engineMock) LatestSilaBlock(context.Context) (*pb.SilaBlock, error) {
	return nil, nil
}

func (m *engineMock) SilaBlockByHash(_ context.Context, hash common.Hash, _ bool) (*pb.SilaBlock, error) {
	b, ok := m.powBlocks[bytesutil.ToBytes32(hash.Bytes())]
	if !ok {
		return nil, nil
	}

	td := new(big.Int).SetBytes(bytesutil.ReverseByteOrder(b.TotalDifficulty))
	tdHex := hexutil.EncodeBig(td)
	return &pb.SilaBlock{
		Header: silaTypes.Header{
			ParentHash: common.BytesToHash(b.ParentHash),
		},
		TotalDifficulty: tdHex,
		Hash:            common.BytesToHash(b.BlockHash),
	}, nil
}

func (m *engineMock) GetTerminalBlockHash(context.Context, uint64) ([]byte, bool, error) {
	return nil, false, nil
}

func (m *engineMock) GetClientVersionV1(context.Context) ([]*structs.ClientVersionV1, error) {
	return nil, nil
}
