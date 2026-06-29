// Package testing provides useful mocks for a silaexec powchain
// service as needed by unit tests for the beacon node.
package testing

import (
	"context"
	"fmt"
	"math/big"
	"net/http/httptest"
	"time"

	"github.com/pkg/errors"
	"github.com/sila-chain/Sila-Consensus-Core/v7/async/event"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec/types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila/common"
	"github.com/sila-chain/Sila/common/hexutil"
	silaTypes "github.com/sila-chain/Sila/core/types"
	"github.com/sila-chain/Sila/rpc"
	"github.com/sila-chain/Sila/silaclient/simulated"
)

// Chain defines a properly functioning mock for the powchain service.
type Chain struct {
	ChainFeed         *event.Feed
	LatestBlockNumber *big.Int
	HashesByHeight    map[int][]byte
	TimesByHeight     map[int]uint64
	BlockNumberByTime map[uint64]*big.Int
	SilaData          *silapb.SilaData
	GenesisSilaBlock  *big.Int
	GenesisState      state.BeaconState
	CurrEndpoint      string
	CurrError         error
	Endpoints         []string
	Errors            []error
}

// GenesisTime represents a static past date - JAN 01 2000.
var GenesisTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

// New creates a new mock chain with empty block info.
func New() *Chain {
	return &Chain{
		HashesByHeight:    make(map[int][]byte),
		TimesByHeight:     make(map[int]uint64),
		BlockNumberByTime: make(map[uint64]*big.Int),
	}
}

// GenesisSilaChainInfo --
func (m *Chain) GenesisSilaChainInfo() (uint64, *big.Int) {
	blk := m.GenesisSilaBlock
	if blk == nil {
		blk = big.NewInt(GenesisTime)
	}
	return uint64(GenesisTime), blk
}

// BlockExists --
func (m *Chain) BlockExists(_ context.Context, hash common.Hash) (bool, *big.Int, error) {
	// Reverse the map of heights by hash.
	heightsByHash := make(map[[32]byte]int, len(m.HashesByHeight))
	for k, v := range m.HashesByHeight {
		h := bytesutil.ToBytes32(v)
		heightsByHash[h] = k
	}
	val, ok := heightsByHash[hash]
	if !ok {
		return false, nil, fmt.Errorf("could not fetch height for hash: %#x", hash)
	}
	return true, big.NewInt(int64(val)), nil
}

// BlockHashByHeight --
func (m *Chain) BlockHashByHeight(_ context.Context, height *big.Int) (common.Hash, error) {
	k := int(height.Int64())
	val, ok := m.HashesByHeight[k]
	if !ok {
		return [32]byte{}, fmt.Errorf("could not fetch hash for height: %v", height)
	}
	return bytesutil.ToBytes32(val), nil
}

// BlockTimeByHeight --
func (m *Chain) BlockTimeByHeight(_ context.Context, height *big.Int) (uint64, error) {
	h := int(height.Int64())
	return m.TimesByHeight[h], nil
}

// BlockByTimestamp --
func (m *Chain) BlockByTimestamp(_ context.Context, time uint64) (*types.HeaderInfo, error) {
	var chosenTime uint64
	var chosenNumber *big.Int
	for t, num := range m.BlockNumberByTime {
		if t > chosenTime && t <= time {
			chosenNumber = num
			chosenTime = t
		}
	}
	return &types.HeaderInfo{Number: chosenNumber, Time: chosenTime}, nil
}

// ChainStartSilaData --
func (m *Chain) ChainStartSilaData() *silapb.SilaData {
	return m.SilaData
}

// PreGenesisState --
func (m *Chain) PreGenesisState() state.BeaconState {
	return m.GenesisState
}

// ClearPreGenesisData --
func (*Chain) ClearPreGenesisData() {
	// no-op
}

func (*Chain) SilaClientConnected() bool {
	return true
}

func (m *Chain) SilaClientEndpoint() string {
	return m.CurrEndpoint
}

func (m *Chain) SilaClientConnectionErr() error {
	return m.CurrError
}

func (m *Chain) SilaEndpoints() []string {
	return m.Endpoints
}

func (m *Chain) SilaExecutionConnectionErrors() []error {
	return m.Errors
}

// RPCClient defines the mock rpc client.
type RPCClient struct {
	Backend     *simulated.Backend
	BlockNumMap map[uint64]*types.HeaderInfo
}

func (*RPCClient) Close() {}

func (r *RPCClient) CallContext(ctx context.Context, obj any, methodName string, args ...any) error {
	if methodName == "sila_chainId" {
		chainID, ok := obj.(*string)
		if !ok {
			return errors.Errorf("wrong argument type provided: %T", obj)
		}
		*chainID = "0x7ea"
		return nil
	}
	if r.BlockNumMap != nil && methodName == "sila_getBlockByNumber" {
		val, ok := args[0].(string)
		if !ok {
			return errors.Errorf("wrong argument type provided: %T", args[0])
		}
		num, err := hexutil.DecodeBig(val)
		if err != nil {
			return err
		}
		b := r.BlockNumMap[num.Uint64()]
		assertedObj, ok := obj.(**types.HeaderInfo)
		if !ok {
			return errors.Errorf("wrong argument type provided: %T", obj)
		}
		*assertedObj = b
		return nil
	}
	if r.Backend == nil && methodName == "sila_getBlockByNumber" {
		h := &silaTypes.Header{
			Number: big.NewInt(15),
			Time:   150,
		}
		assertedObj, ok := obj.(**types.HeaderInfo)
		if !ok {
			return errors.Errorf("wrong argument type provided: %T", obj)
		}
		*assertedObj = &types.HeaderInfo{
			Hash:   h.Hash(),
			Number: h.Number,
			Time:   h.Time,
		}
		return nil
	}
	switch methodName {
	case "sila_getBlockByNumber":
		val, ok := args[0].(string)
		if !ok {
			return errors.Errorf("wrong argument type provided: %T", args[0])
		}
		var num *big.Int
		var err error
		if val != "latest" {
			num, err = hexutil.DecodeBig(val)
			if err != nil {
				return err
			}
		}
		h, err := r.Backend.Client().HeaderByNumber(ctx, num)
		if err != nil {
			return err
		}
		assertedObj, ok := obj.(**types.HeaderInfo)
		if !ok {
			return errors.Errorf("wrong argument type provided: %T", obj)
		}
		*assertedObj = &types.HeaderInfo{
			Hash:   h.Hash(),
			Number: h.Number,
			Time:   h.Time,
		}
	case "sila_getBlockByHash":
		val, ok := args[0].(common.Hash)
		if !ok {
			return errors.Errorf("wrong argument type provided: %T", args[0])
		}
		h, err := r.Backend.Client().HeaderByHash(ctx, val)
		if err != nil {
			return err
		}
		assertedObj, ok := obj.(**types.HeaderInfo)
		if !ok {
			return errors.Errorf("wrong argument type provided: %T", obj)
		}
		*assertedObj = &types.HeaderInfo{
			Hash:   h.Hash(),
			Number: h.Number,
			Time:   h.Time,
		}
	}
	return nil
}

// BatchCall --
func (r *RPCClient) BatchCall(b []rpc.BatchElem) error {
	if r.Backend == nil {
		return nil
	}

	for _, e := range b {
		num, err := hexutil.DecodeBig(e.Args[0].(string))
		if err != nil {
			return err
		}
		h, err := r.Backend.Client().HeaderByNumber(context.Background(), num)
		if err != nil {
			return err
		}
		*e.Result.(*types.HeaderInfo) = types.HeaderInfo{Number: h.Number, Time: h.Time, Hash: h.Hash()}
	}
	return nil
}

// InsertBlock adds provided block info into the chain.
func (m *Chain) InsertBlock(height int, time uint64, hash []byte) *Chain {
	m.HashesByHeight[height] = hash
	m.TimesByHeight[height] = time
	m.BlockNumberByTime[time] = big.NewInt(int64(height))
	return m
}

func SetupRPCServer() (*rpc.Server, string, error) {
	srv := rpc.NewServer()
	if err := srv.RegisterName("eth", &testSilaRPC{}); err != nil {
		return nil, "", err
	}
	if err := srv.RegisterName("sila", &testSilaRPC{}); err != nil {
		return nil, "", err
	}
	if err := srv.RegisterName("net", &testSilaRPC{}); err != nil {
		return nil, "", err
	}
	hs := httptest.NewUnstartedServer(srv)
	hs.Start()
	return srv, hs.URL, nil
}

type testSilaRPC struct{}

func (*testSilaRPC) NoArgsRets() {}

func (*testSilaRPC) ChainId(_ context.Context) *hexutil.Big {
	return (*hexutil.Big)(big.NewInt(int64(params.BeaconConfig().DepositChainID)))
}

func (*testSilaRPC) Version(_ context.Context) string {
	return fmt.Sprintf("%d", params.BeaconConfig().DepositNetworkID)
}
