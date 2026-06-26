// Package execution defines a runtime service which is tasked with
// communicating with an silaexec endpoint, processing logs from a deposit
// contract, and the latest silaexec data headers for usage in the beacon node.
package execution

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"runtime/debug"
	"sort"
	"sync"
	"time"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache/depositsnapshot"
	statefeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/transition"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/execution/types"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state"
	native "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/state-native"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/state/stategen"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/verification"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/container/trie"
	contracts "github.com/sila-chain/Sila-Consensus-Core/v7/contracts/deposit"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/monitoring/clientstats"
	"github.com/sila-chain/Sila-Consensus-Core/v7/network"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	silaTime "github.com/sila-chain/Sila-Consensus-Core/v7/time"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/sila-chain/Sila/accounts/abi/bind"
	"github.com/sila-chain/Sila/common"
	"github.com/sila-chain/Sila/common/hexutil"
	gethRPC "github.com/sila-chain/Sila/rpc"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

var (
	validDepositsCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "powchain_valid_deposits_received",
		Help: "The number of valid deposits received in the sila deposit",
	})
	blockNumberGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "powchain_block_number",
		Help: "The current block number in the proof-of-work chain",
	})
	missedDepositLogsCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "powchain_missed_deposit_logs",
		Help: "The number of times a missed deposit log is detected",
	})
)

var (
	// time to wait before trying to reconnect with the silaexec node.
	backOffPeriod = 15 * time.Second
	// amount of times before we log the status of the silaexec dial attempt.
	logThreshold = 8
	// period to log chainstart related information
	logPeriod = 1 * time.Minute
)

// ChainStartFetcher retrieves information pertaining to the chain start event
// of the beacon chain for usage across various services.
type ChainStartFetcher interface {
	ChainStartSilaExecutionData() *silapb.SilaExecutionData
	PreGenesisState() state.BeaconState
	ClearPreGenesisData()
}

// ChainInfoFetcher retrieves information about silaexec metadata at the Sila consensus genesis time.
type ChainInfoFetcher interface {
	GenesisExecutionChainInfo() (uint64, *big.Int)
	ExecutionClientConnected() bool
	ExecutionClientEndpoint() string
	ExecutionClientConnectionErr() error
}

// POWBlockFetcher defines a struct that can retrieve mainchain blocks.
type POWBlockFetcher interface {
	BlockTimeByHeight(ctx context.Context, height *big.Int) (uint64, error)
	BlockByTimestamp(ctx context.Context, time uint64) (*types.HeaderInfo, error)
	BlockHashByHeight(ctx context.Context, height *big.Int) (common.Hash, error)
	BlockExists(ctx context.Context, hash common.Hash) (bool, *big.Int, error)
}

// Chain defines a standard interface for the powchain service in Sila.
type Chain interface {
	ChainStartFetcher
	ChainInfoFetcher
	POWBlockFetcher
}

// RPCClient defines the rpc methods required to interact with the silaexec node.
type RPCClient interface {
	Close()
	BatchCall(b []gethRPC.BatchElem) error
	CallContext(ctx context.Context, result any, method string, args ...any) error
}

type RPCClientEmpty struct {
}

func (RPCClientEmpty) Close() {}
func (RPCClientEmpty) BatchCall([]gethRPC.BatchElem) error {
	return errors.New("rpc client is not initialized")
}

func (RPCClientEmpty) CallContext(context.Context, any, string, ...any) error {
	return errors.New("rpc client is not initialized")
}

// config defines a config struct for dependencies into the service.
type config struct {
	silaDepositAddr     common.Address
	beaconDB                db.HeadAccessDatabase
	depositCache            cache.DepositCache
	stateNotifier           statefeed.Notifier
	stateGen                *stategen.State
	silaexecHeaderReqLimit      uint64
	beaconNodeStatsUpdater  BeaconNodeStatsUpdater
	currHttpEndpoint        network.Endpoint
	headers                 []string
	finalizedStateAtStartup state.BeaconState
	jwtId                   string
}

// Service fetches important information about the canonical
// silaexec chain via a web3 endpoint using an ethclient.
// The beacon chain requires synchronization with the silaexec chain's current
// block hash, block number, and access to logs within the
// Validator Registration Contract on the silaexec chain to kick off the beacon
// chain's validator registration process.
type Service struct {
	connectedSilaExecution           bool
	isRunning               bool
	depositRequestsStarted  bool
	processingLock          sync.RWMutex
	latestSilaExecutionDataLock      sync.RWMutex
	cfg                     *config
	ctx                     context.Context
	cancel                  context.CancelFunc
	silaexecHeadTicker          *time.Ticker
	httpLogger              bind.ContractFilterer
	rpcClient               RPCClient
	headerCache             *headerCache // cache to store block hash/block height.
	latestSilaExecutionData          *silapb.LatestSilaExecutionData
	silaDepositCaller   *contracts.SilaDepositCaller
	depositTrie             cache.MerkleTree
	chainStartData          *silapb.ChainStartData
	lastReceivedMerkleIndex int64 // Keeps track of the last received index to prevent log spam.
	runError                error
	preGenesisState         state.BeaconState
	verifierWaiter          *verification.InitializerWaiter
	blobVerifier            verification.NewBlobVerifier
	capabilityCache         *capabilityCache
	graffitiInfo            *GraffitiInfo
}

// NewService sets up a new instance with an ethclient when given a web3 endpoint as a string in the config.
func NewService(ctx context.Context, opts ...Option) (*Service, error) {
	ctx, cancel := context.WithCancel(ctx)
	_ = cancel // govet fix for lost cancel. Cancel is handled in service.Stop()
	var depositTrie cache.MerkleTree
	var err error
	depositTrie = depositsnapshot.NewDepositTree()
	genState, err := transition.EmptyGenesisState()
	if err != nil {
		return nil, errors.Wrap(err, "could not set up genesis state")
	}

	s := &Service{
		ctx:       ctx,
		cancel:    cancel,
		rpcClient: RPCClientEmpty{},
		cfg: &config{
			beaconNodeStatsUpdater: &NopBeaconNodeStatsUpdater{},
			silaexecHeaderReqLimit:     defaultSilaExecutionHeaderReqLimit,
		},
		latestSilaExecutionData: &silapb.LatestSilaExecutionData{
			BlockHeight:        0,
			BlockTime:          0,
			BlockHash:          []byte{},
			LastRequestedBlock: 0,
		},
		headerCache: newHeaderCache(),
		depositTrie: depositTrie,
		chainStartData: &silapb.ChainStartData{
			SilaExecutionData:           &silapb.SilaExecutionData{},
			ChainstartDeposits: make([]*silapb.Deposit, 0),
		},
		lastReceivedMerkleIndex: -1,
		preGenesisState:         genState,
		silaexecHeadTicker:          time.NewTicker(time.Duration(params.BeaconConfig().SecondsPerSilaBlock) * time.Second),
		capabilityCache:         &capabilityCache{},
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	s.initDepositRequests()
	silaexecData, err := s.validPowchainData(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to validate powchain data")
	}
	if err := s.initializeSilaExecutionData(ctx, silaexecData); err != nil {
		return nil, err
	}
	return s, nil
}

// Start the powchain service's main event loop.
func (s *Service) Start() {
	if err := s.setupExecutionClientConnections(s.ctx, s.cfg.currHttpEndpoint); err != nil {
		log.WithError(err).Error("Could not connect to execution endpoint")
	}
	// If the chain has not started already and we don't have access to silaexec nodes, we will not be
	// able to generate the genesis state.
	if !s.chainStartData.Chainstarted && s.cfg.currHttpEndpoint.Url == "" {
		// check for genesis state before shutting down the node,
		// if a genesis state exists, we can continue on.
		genState, err := s.cfg.beaconDB.GenesisState(s.ctx)
		if err != nil {
			log.Fatal(err)
		}
		if genState == nil || genState.IsNil() {
			log.Fatal("cannot create genesis state: no silaexec http endpoint defined")
		}
	}

	v, err := s.verifierWaiter.WaitForInitializer(s.ctx)
	if err != nil {
		log.WithError(err).Error("Could not get verification initializer")
		return
	}
	s.blobVerifier = newBlobVerifierFromInitializer(v)

	s.isRunning = true

	// Poll the execution client connection and fallback if errors occur.
	go s.pollConnectionStatus(s.ctx)

	go s.run(s.ctx.Done())
}

// Stop the web3 service's main event loop and associated goroutines.
func (s *Service) Stop() error {
	if s.cancel != nil {
		defer s.cancel()
	}
	if s.rpcClient != nil {
		s.rpcClient.Close()
	}
	return nil
}

// ClearPreGenesisData clears out the stored chainstart deposits and beacon state.
func (s *Service) ClearPreGenesisData() {
	s.chainStartData.ChainstartDeposits = []*silapb.Deposit{}
	s.preGenesisState = &native.BeaconState{}
}

// ChainStartSilaExecutionData returns the silaexec data at chainstart.
func (s *Service) ChainStartSilaExecutionData() *silapb.SilaExecutionData {
	return s.chainStartData.SilaExecutionData
}

// PreGenesisState returns a state that contains
// pre-chainstart deposits.
func (s *Service) PreGenesisState() state.BeaconState {
	return s.preGenesisState
}

// Status is service health checks. Return nil or error.
func (s *Service) Status() error {
	// Service don't start
	if !s.isRunning {
		return nil
	}
	// get error from run function
	return s.runError
}

// ExecutionClientConnected checks whether are connected via RPC.
func (s *Service) ExecutionClientConnected() bool {
	return s.connectedSilaExecution
}

// ExecutionClientEndpoint returns the URL of the current, connected execution client.
func (s *Service) ExecutionClientEndpoint() string {
	return s.cfg.currHttpEndpoint.Url
}

// ExecutionClientConnectionErr returns the error (if any) of the current connection.
func (s *Service) ExecutionClientConnectionErr() error {
	return s.runError
}

func (s *Service) updateBeaconNodeStats() {
	bs := clientstats.BeaconNodeStats{}
	if s.ExecutionClientConnected() {
		bs.SyncSilaExecutionConnected = true
	}
	s.cfg.beaconNodeStatsUpdater.Update(bs)
}

func (s *Service) updateConnectedSilaExecution(state bool) {
	s.connectedSilaExecution = state
	s.updateBeaconNodeStats()
}

// GraffitiInfo returns the GraffitiInfo struct for graffiti generation.
func (s *Service) GraffitiInfo() *GraffitiInfo {
	return s.graffitiInfo
}

// updateGraffitiInfo fetches EL client version and updates the graffiti info.
func (s *Service) updateGraffitiInfo() {
	if s.graffitiInfo == nil {
		return
	}
	ctx, cancel := context.WithTimeout(s.ctx, time.Second)
	defer cancel()
	versions, err := s.GetClientVersionV1(ctx)
	if err != nil {
		log.WithError(err).Debug("Could not get execution client version for graffiti")
		return
	}
	if len(versions) >= 1 {
		s.graffitiInfo.UpdateFromEngine(versions[0].Code, versions[0].Commit)
	}
}

// refers to the latest silaexec block which follows the condition: silaexec_timestamp +
// SECONDS_PER_SilaExecution_BLOCK * SilaExecution_FOLLOW_DISTANCE <= current_unix_time
func (s *Service) followedBlockHeight(ctx context.Context) (uint64, error) {
	followTime := params.BeaconConfig().SilaExecutionFollowDistance * params.BeaconConfig().SecondsPerSilaBlock
	latestBlockTime := uint64(0)
	if s.latestSilaExecutionData.BlockTime > followTime {
		latestBlockTime = s.latestSilaExecutionData.BlockTime - followTime
		// This should only come into play in testnets - when the chain hasn't advanced past the follow distance,
		// we don't want to consider any block before the genesis block.
		if s.latestSilaExecutionData.BlockHeight < params.BeaconConfig().SilaExecutionFollowDistance {
			latestBlockTime = s.latestSilaExecutionData.BlockTime
		}
	}
	blk, err := s.BlockByTimestamp(ctx, latestBlockTime)
	if err != nil {
		return 0, errors.Wrapf(err, "BlockByTimestamp=%d", latestBlockTime)
	}
	return blk.Number.Uint64(), nil
}

func (s *Service) initDepositCaches(ctx context.Context, ctrs []*silapb.DepositContainer) error {
	if len(ctrs) == 0 {
		return nil
	}
	s.cfg.depositCache.InsertDepositContainers(ctx, ctrs)
	if !s.chainStartData.Chainstarted {
		// Do not add to pending cache if no genesis state exists.
		validDepositsCount.Add(float64(s.preGenesisState.SilaExecutionDepositIndex()))
		return nil
	}
	genesisState, err := s.cfg.beaconDB.GenesisState(ctx)
	if err != nil {
		return err
	}
	// Default to all post-genesis deposits in
	// the event we cannot find a finalized state.
	currIndex := genesisState.SilaExecutionDepositIndex()
	chkPt, err := s.cfg.beaconDB.FinalizedCheckpoint(ctx)
	if err != nil {
		return err
	}
	rt := bytesutil.ToBytes32(chkPt.Root)
	if rt != [32]byte{} {
		fState := s.cfg.finalizedStateAtStartup
		if fState == nil || fState.IsNil() {
			return errors.Errorf("finalized state with root %#x is nil", rt)
		}
		// Set deposit index to the one in the current archived state.
		currIndex = fState.SilaExecutionDepositIndex()

		// When a node pauses for some time and starts again, the deposits to finalize
		// accumulates. We finalize them here before we are ready to receive a block.
		// Otherwise, the first few blocks will be slower to compute as we will
		// hold the lock and be busy finalizing the deposits.
		// The deposit index in the state is always the index of the next deposit
		// to be included (rather than the last one to be processed). This was most likely
		// done as the state cannot represent signed integers.
		actualIndex := int64(currIndex) - 1 // lint:ignore uintcast -- deposit index will not exceed int64 in your lifetime.
		if err = s.cfg.depositCache.InsertFinalizedDeposits(ctx, actualIndex, common.Hash(fState.SilaExecutionData().BlockHash),
			0 /* Setting a zero value as we have no access to block height */); err != nil {
			return err
		}

		// Deposit proofs are only used during state transition and can be safely removed to save space.
		if err = s.cfg.depositCache.PruneProofs(ctx, actualIndex); err != nil {
			return errors.Wrap(err, "could not prune deposit proofs")
		}
	}
	validDepositsCount.Add(float64(currIndex))
	// Only add pending deposits if the container slice length
	// is more than the current index in state.
	if uint64(len(ctrs)) > currIndex {
		for _, c := range ctrs[currIndex:] {
			s.cfg.depositCache.InsertPendingDeposit(ctx, c.Deposit, c.SilaBlockHeight, c.Index, bytesutil.ToBytes32(c.DepositRoot))
		}
	}
	return nil
}

// processBlockHeader adds a newly observed silaexec block to the block cache and
// updates the latest blockHeight, blockHash, and blockTime properties of the service.
func (s *Service) processBlockHeader(header *types.HeaderInfo) {
	defer safelyHandlePanic()
	blockNumberGauge.Set(float64(header.Number.Int64()))
	s.latestSilaExecutionDataLock.Lock()
	s.latestSilaExecutionData.BlockHeight = header.Number.Uint64()
	s.latestSilaExecutionData.BlockHash = header.Hash.Bytes()
	s.latestSilaExecutionData.BlockTime = header.Time
	s.latestSilaExecutionDataLock.Unlock()
	log.WithFields(logrus.Fields{
		"blockNumber": s.latestSilaExecutionData.BlockHeight,
		"blockHash":   hexutil.Encode(s.latestSilaExecutionData.BlockHash),
	}).Debug("Latest silaexec chain event")
}

// batchRequestHeaders requests the block range specified in the arguments. Instead of requesting
// each block in one call, it batches all requests into a single rpc call.
func (s *Service) batchRequestHeaders(startBlock, endBlock uint64) ([]*types.HeaderInfo, error) {
	if startBlock > endBlock {
		return nil, fmt.Errorf("start block height %d cannot be > end block height %d", startBlock, endBlock)
	}
	requestRange := (endBlock - startBlock) + 1
	elems := make([]gethRPC.BatchElem, 0, requestRange)
	headers := make([]*types.HeaderInfo, 0, requestRange)
	for i := startBlock; i <= endBlock; i++ {
		header := &types.HeaderInfo{}
		elems = append(elems, gethRPC.BatchElem{
			Method: "sila_getBlockByNumber",
			Args:   []any{hexutil.EncodeBig(new(big.Int).SetUint64(i)), false},
			Result: header,
			Error:  error(nil),
		})
		headers = append(headers, header)
	}
	ioErr := s.rpcClient.BatchCall(elems)
	if ioErr != nil {
		return nil, ioErr
	}
	for _, e := range elems {
		if e.Error != nil {
			return nil, e.Error
		}
	}
	for _, h := range headers {
		if h != nil {
			if err := s.headerCache.AddHeader(h); err != nil {
				return nil, err
			}
		}
	}
	return headers, nil
}

// safelyHandleHeader will recover and log any panic that occurs from the block
func safelyHandlePanic() {
	if r := recover(); r != nil {
		log.WithFields(logrus.Fields{
			"r": r,
		}).Error("Panicked when handling data from ETH 1.0 Chain! Recovering...")

		debug.PrintStack()
	}
}

func (s *Service) handleSilaExecutionFollowDistance() {
	defer safelyHandlePanic()
	ctx := s.ctx
	if s.depositRequestsStarted {
		return
	}
	// use a 5 minutes timeout for block time, because the max mining time is 278 sec (block 7208027)
	// (analyzed the time of the block from 2018-09-01 to 2019-02-13)
	fiveMinutesTimeout := silaTime.Now().Add(-5 * time.Minute)
	// check that web3 client is syncing
	if time.Unix(int64(s.latestSilaExecutionData.BlockTime), 0).Before(fiveMinutesTimeout) {
		log.Warn("Execution client is not syncing")
	}
	if !s.chainStartData.Chainstarted {
		if err := s.processChainStartFromBlockNum(ctx, big.NewInt(int64(s.latestSilaExecutionData.LastRequestedBlock))); err != nil {
			s.runError = errors.Wrap(err, "processChainStartFromBlockNum")
			log.Error(err)
			return
		}
	}

	// If the last requested block has not changed,
	// we do not request batched logs as this means there are no new
	// logs for the execution service to process. Also it is a potential
	// failure condition as would mean we have not respected the protocol threshold.
	if s.latestSilaExecutionData.LastRequestedBlock == s.latestSilaExecutionData.BlockHeight {
		log.WithField("lastBlockNumber", s.latestSilaExecutionData.LastRequestedBlock).Error("Beacon node is not respecting the follow distance. EL client is syncing.")
		return
	}
	if err := s.requestBatchedHeadersAndLogs(ctx); err != nil {
		s.runError = errors.Wrap(err, "requestBatchedHeadersAndLogs")
		log.Error(err)
		return
	}
	// Reset the Status.
	if s.runError != nil {
		s.runError = nil
	}
}

func (s *Service) initPOWService() {
	// Use a custom logger to only log errors
	logCounter := 0
	errorLogger := func(err error, msg string) {
		if logCounter > logThreshold {
			log.WithError(err).Error(msg)
			logCounter = 0
		}
		logCounter++
	}

	// Run in a select loop to retry in the event of any failures.
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			ctx := s.ctx
			header, err := s.HeaderByNumber(ctx, nil)
			if err != nil {
				err = errors.Wrap(err, "HeaderByNumber")
				s.retryExecutionClientConnection(ctx, err)
				errorLogger(err, "Unable to retrieve latest execution client header")
				continue
			}

			s.latestSilaExecutionDataLock.Lock()
			s.latestSilaExecutionData.BlockHeight = header.Number.Uint64()
			s.latestSilaExecutionData.BlockHash = header.Hash.Bytes()
			s.latestSilaExecutionData.BlockTime = header.Time
			s.latestSilaExecutionDataLock.Unlock()

			if !s.depositRequestsStarted {
				if err := s.processPastLogs(ctx); err != nil {
					err = errors.Wrap(err, "processPastLogs")
					s.retryExecutionClientConnection(ctx, err)
					errorLogger(
						err,
						"Unable to process past sila deposit logs, perhaps your execution client is not fully synced",
					)
					continue
				}
				// Cache silaexec headers from our voting period.
				if err := s.cacheHeadersForSilaExecutionDataVote(ctx); err != nil {
					err = errors.Wrap(err, "cacheHeadersForSilaExecutionDataVote")
					s.retryExecutionClientConnection(ctx, err)
					if errors.Is(err, errBlockTimeTooLate) {
						log.WithError(err).Debug("Unable to cache headers for execution client votes")
					} else {
						errorLogger(err, "Unable to cache headers for execution client votes")
					}
					continue
				}
			}
			// Handle edge case with embedded genesis state by fetching genesis header to determine
			// its height only if the deposit requests have not started yet (Pre Pectra SIP-6110 behavior).
			if s.chainStartData.Chainstarted && s.chainStartData.GenesisBlock == 0 && !s.depositRequestsStarted {
				genHash := common.BytesToHash(s.chainStartData.SilaExecutionData.BlockHash)
				genBlock := s.chainStartData.GenesisBlock
				// In the event our provided chainstart data references a non-existent block hash,
				// we assume the genesis block to be 0.
				if genHash != [32]byte{} {
					genHeader, err := s.HeaderByHash(ctx, genHash)
					if err != nil {
						err = errors.Wrapf(err, "HeaderByHash, hash=%#x", genHash)
						s.retryExecutionClientConnection(ctx, err)
						errorLogger(err, "Unable to retrieve proof-of-stake genesis block data")
						continue
					}
					genBlock = genHeader.Number.Uint64()
				}
				s.chainStartData.GenesisBlock = genBlock
				if err := s.savePowchainData(ctx); err != nil {
					err = errors.Wrap(err, "savePowchainData")
					s.retryExecutionClientConnection(ctx, err)
					errorLogger(err, "Unable to save execution client data")
					continue
				}
			}
			return
		}
	}
}

// run subscribes to all the services for the silaexec chain.
func (s *Service) run(done <-chan struct{}) {
	s.runError = nil

	s.initPOWService()
	// Do not keep storing the finalized state as it is
	// no longer of use.
	s.removeStartupState()

	chainstartTicker := time.NewTicker(logPeriod)
	defer chainstartTicker.Stop()

	// Update graffiti info 4 times per epoch (~96 seconds with 12s slots and 32 slots/epoch)
	graffitiTicker := time.NewTicker(96 * time.Second)
	defer graffitiTicker.Stop()
	// Initial update
	s.updateGraffitiInfo()

	for {
		select {
		case <-done:
			s.isRunning = false
			s.runError = nil
			s.rpcClient.Close()
			s.updateConnectedSilaExecution(false)
			log.Debug("Context closed, exiting goroutine")
			return
		case <-s.silaexecHeadTicker.C:
			head, err := s.HeaderByNumber(s.ctx, nil)
			if err != nil {
				s.pollConnectionStatus(s.ctx)
				log.WithError(err).Debug("Could not fetch latest silaexec header")
				continue
			}
			s.processBlockHeader(head)
			s.handleSilaExecutionFollowDistance()
		case <-chainstartTicker.C:
			if s.chainStartData.Chainstarted {
				chainstartTicker.Stop()
				continue
			}
			s.logTillChainStart(context.Background())
		case <-graffitiTicker.C:
			s.updateGraffitiInfo()
		}
	}
}

// logs the current thresholds required to hit chainstart every minute.
func (s *Service) logTillChainStart(ctx context.Context) {
	if s.chainStartData.Chainstarted {
		return
	}
	_, blockTime, err := s.retrieveBlockHashAndTime(s.ctx, big.NewInt(int64(s.latestSilaExecutionData.LastRequestedBlock)))
	if err != nil {
		log.Error(err)
		return
	}
	valCount, genesisTime := s.currentCountAndTime(ctx, blockTime)
	valNeeded := uint64(0)
	if valCount < params.BeaconConfig().MinGenesisActiveValidatorCount {
		valNeeded = params.BeaconConfig().MinGenesisActiveValidatorCount - valCount
	}
	secondsLeft := uint64(0)
	if genesisTime < params.BeaconConfig().MinGenesisTime {
		secondsLeft = params.BeaconConfig().MinGenesisTime - genesisTime
	}

	fields := logrus.Fields{
		"additionalValidatorsNeeded": valNeeded,
	}
	if secondsLeft > 0 {
		fields["Generating genesis state in"] = time.Duration(secondsLeft) * time.Second
	}

	log.WithFields(fields).Info("Currently waiting for chainstart")
}

// cacheHeadersForSilaExecutionDataVote makes sure that voting for silaExecutionData after startup utilizes cached headers
// instead of making multiple RPC requests to the silaexec endpoint.
func (s *Service) cacheHeadersForSilaExecutionDataVote(ctx context.Context) error {
	// Find the end block to request from.
	end, err := s.followedBlockHeight(ctx)
	if err != nil {
		return errors.Wrap(err, "followedBlockHeight")
	}
	start, err := s.determineEarliestVotingBlock(ctx, end)
	if err != nil {
		return errors.Wrapf(err, "determineEarliestVotingBlock=%d", end)
	}
	return s.cacheBlockHeaders(start, end)
}

// Caches block headers from the desired range.
func (s *Service) cacheBlockHeaders(start, end uint64) error {
	batchSize := s.cfg.silaexecHeaderReqLimit
	for i := start; i < end; i += batchSize {
		startReq := i
		endReq := i + batchSize
		if endReq > 0 {
			// Reduce the end request by one
			// to prevent total batch size from exceeding
			// the allotted limit.
			endReq -= 1
		}
		endReq = min(endReq, end)
		// We call batchRequestHeaders for its header caching side-effect, so we don't need the return value.
		_, err := s.batchRequestHeaders(startReq, endReq)
		if err != nil {
			if clientTimedOutError(err) {
				// Reduce batch size as silaexec node is
				// unable to respond to the request in time.
				batchSize /= 2
				// Always have it greater than 0.
				if batchSize == 0 {
					batchSize += 1
				}

				// Reset request value
				if i > batchSize {
					i -= batchSize
				}
				continue
			}
			return errors.Wrapf(err, "cacheBlockHeaders, start=%d, end=%d", startReq, endReq)
		}
	}
	return nil
}

// Determines the earliest voting block from which to start caching all our previous headers from.
func (s *Service) determineEarliestVotingBlock(ctx context.Context, followBlock uint64) (uint64, error) {
	genesisTime := s.chainStartData.GenesisTime
	currSlot := slots.CurrentSlot(time.Unix(int64(genesisTime), 0)) // lint:ignore uintcast -- Genesis time will never exceed int64 in seconds.

	// In the event genesis has not occurred yet, we just request to go back follow_distance blocks.
	if genesisTime == 0 || currSlot == 0 {
		earliestBlk := uint64(0)
		if followBlock > params.BeaconConfig().SilaExecutionFollowDistance {
			earliestBlk = followBlock - params.BeaconConfig().SilaExecutionFollowDistance
		}
		return earliestBlk, nil
	}
	// This should only come into play in testnets - when the chain hasn't advanced past the follow distance,
	// we don't want to consider any block before the genesis block.
	if s.latestSilaExecutionData.BlockHeight < params.BeaconConfig().SilaExecutionFollowDistance {
		return 0, nil
	}
	votingTime := slots.VotingPeriodStartTime(genesisTime, currSlot)
	followBackDist := 2 * params.BeaconConfig().SecondsPerSilaBlock * params.BeaconConfig().SilaExecutionFollowDistance
	if followBackDist > votingTime {
		return 0, errors.Errorf("invalid genesis time provided. %d > %d", followBackDist, votingTime)
	}
	earliestValidTime := votingTime - followBackDist
	if earliestValidTime < genesisTime {
		return 0, nil
	}
	hdr, err := s.BlockByTimestamp(ctx, earliestValidTime)
	if err != nil {
		return 0, err
	}
	return hdr.Number.Uint64(), nil
}

// initializes our service from the provided silaExecutionData object by initializing all the relevant
// fields and data.
func (s *Service) initializeSilaExecutionData(ctx context.Context, silaexecDataInDB *silapb.SilaExecutionChainData) error {
	// The node has no silaExecutionData persisted on disk, so we exit and instead
	// request from contract logs.
	if silaexecDataInDB == nil {
		return nil
	}
	var err error
	if silaexecDataInDB.DepositSnapshot != nil {
		s.depositTrie, err = depositsnapshot.DepositTreeFromSnapshotProto(silaexecDataInDB.DepositSnapshot)
	} else {
		if err = s.migrateOldDepositTree(silaexecDataInDB); err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	s.chainStartData = silaexecDataInDB.ChainstartData
	if !reflect.ValueOf(silaexecDataInDB.BeaconState).IsZero() {
		s.preGenesisState, err = native.InitializeFromProtoPhase0(silaexecDataInDB.BeaconState)
		if err != nil {
			return errors.Wrap(err, "Could not initialize state trie")
		}
	}
	s.latestSilaExecutionData = silaexecDataInDB.CurrentSilaExecutionData
	ctrs := silaexecDataInDB.DepositContainers
	// Look at previously finalized index, as we are building off a finalized
	// snapshot rather than the full trie.
	lastFinalizedIndex := int64(s.depositTrie.NumOfItems() - 1)
	// Correctly initialize missing deposits into active trie.
	for _, c := range ctrs {
		if c.Index > lastFinalizedIndex {
			depRoot, err := c.Deposit.Data.HashTreeRoot()
			if err != nil {
				return err
			}
			if err := s.depositTrie.Insert(depRoot[:], int(c.Index)); err != nil {
				return err
			}
		}
	}
	numOfItems := s.depositTrie.NumOfItems()
	s.lastReceivedMerkleIndex = int64(numOfItems - 1)
	if err := s.initDepositCaches(ctx, silaexecDataInDB.DepositContainers); err != nil {
		return errors.Wrap(err, "could not initialize caches")
	}
	return nil
}

// Validates that all deposit containers are valid and have their relevant indices
// in order.
func validateDepositContainers(ctrs []*silapb.DepositContainer) bool {
	ctrLen := len(ctrs)
	// Exit for empty containers.
	if ctrLen == 0 {
		return true
	}
	// Sort deposits in ascending order.
	sort.Slice(ctrs, func(i, j int) bool {
		return ctrs[i].Index < ctrs[j].Index
	})
	startIndex := int64(0)
	for _, c := range ctrs {
		if c.Index != startIndex {
			log.Info("Recovering missing deposit containers, node is re-requesting missing deposit data")
			return false
		}
		startIndex++
	}
	return true
}

// Validates the current powchain data is saved and makes sure that any
// embedded genesis state is correctly accounted for.
func (s *Service) validPowchainData(ctx context.Context) (*silapb.SilaExecutionChainData, error) {
	genState, err := s.cfg.beaconDB.GenesisState(ctx)
	if err != nil {
		return nil, err
	}
	silaexecData, err := s.cfg.beaconDB.ExecutionChainData(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve silaexec data")
	}
	if genState == nil || genState.IsNil() {
		return silaexecData, nil
	}
	if s.depositRequestsStarted || silaexecData == nil || !silaexecData.ChainstartData.Chainstarted || !validateDepositContainers(silaexecData.DepositContainers) {
		pbState, err := native.ProtobufBeaconStatePhase0(s.preGenesisState.ToProtoUnsafe())
		if err != nil {
			return nil, err
		}
		s.chainStartData = &silapb.ChainStartData{
			Chainstarted:       true,
			GenesisTime:        uint64(genState.GenesisTime().Unix()),
			GenesisBlock:       0,
			SilaExecutionData:           genState.SilaExecutionData(),
			ChainstartDeposits: make([]*silapb.Deposit, 0),
		}
		silaexecData = &silapb.SilaExecutionChainData{
			CurrentSilaExecutionData:   s.latestSilaExecutionData,
			ChainstartData:    s.chainStartData,
			BeaconState:       pbState,
			DepositContainers: s.cfg.depositCache.AllDepositContainers(ctx),
		}
		trie, ok := s.depositTrie.(*depositsnapshot.DepositTree)
		if !ok {
			return nil, errors.New("deposit trie was not SIP4881 DepositTree")
		}
		silaexecData.DepositSnapshot, err = trie.ToProto()
		if err != nil {
			return nil, err
		}
		if err := s.cfg.beaconDB.SaveExecutionChainData(ctx, silaexecData); err != nil {
			return nil, err
		}
	}
	return silaexecData, nil
}

func dedupEndpoints(endpoints []string) []string {
	selectionMap := make(map[string]bool)
	newEndpoints := make([]string, 0, len(endpoints))
	for _, point := range endpoints {
		if selectionMap[point] {
			continue
		}
		newEndpoints = append(newEndpoints, point)
		selectionMap[point] = true
	}
	return newEndpoints
}

func (s *Service) migrateOldDepositTree(silaexecDataInDB *silapb.SilaExecutionChainData) error {
	oldDepositTrie, err := trie.CreateTrieFromProto(silaexecDataInDB.Trie)
	if err != nil {
		return err
	}
	newDepositTrie := depositsnapshot.NewDepositTree()
	for i, item := range oldDepositTrie.Items() {
		if err = newDepositTrie.Insert(item, i); err != nil {
			return errors.Wrapf(err, "could not insert item at index %d into deposit snapshot tree", i)
		}
	}
	newDepositRoot, err := newDepositTrie.HashTreeRoot()
	if err != nil {
		return err
	}
	depositRoot, err := oldDepositTrie.HashTreeRoot()
	if err != nil {
		return err
	}
	if newDepositRoot != depositRoot {
		return errors.Wrapf(err, "mismatched deposit roots, old %#x != new %#x", depositRoot, newDepositRoot)
	}
	s.depositTrie = newDepositTrie
	return nil
}

func (s *Service) removeStartupState() {
	s.cfg.finalizedStateAtStartup = nil
}

func (s *Service) initDepositRequests() {
	fState := s.cfg.finalizedStateAtStartup
	isNil := fState == nil || fState.IsNil()
	if isNil {
		return
	}
	s.depositRequestsStarted = helpers.DepositRequestsStarted(fState)
}

func newBlobVerifierFromInitializer(ini *verification.Initializer) verification.NewBlobVerifier {
	return func(b blocks.ROBlob, reqs []verification.Requirement) verification.BlobVerifier {
		return ini.NewBlobVerifier(b, reqs)
	}
}

type capabilityCache struct {
	capabilities     map[string]any
	capabilitiesLock sync.RWMutex
}

func (c *capabilityCache) save(cs []string) {
	c.capabilitiesLock.Lock()
	defer c.capabilitiesLock.Unlock()

	if c.capabilities == nil {
		c.capabilities = make(map[string]any)
	}

	for _, capability := range cs {
		c.capabilities[capability] = struct{}{}
	}
}

func (c *capabilityCache) has(capability string) bool {
	c.capabilitiesLock.RLock()
	defer c.capabilitiesLock.RUnlock()

	if c.capabilities == nil {
		return false
	}

	_, ok := c.capabilities[capability]
	return ok
}
