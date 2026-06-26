package silaexec

import (
	"encoding/binary"
	"math/big"
	"testing"
	"time"

	"github.com/sila-chain/Sila"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/cache/depositsnapshot"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed"
	statefeed "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/feed/state"
	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db"
	testDB "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/db/testing"
	mockSila "github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/silaexec/testing"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	contracts "github.com/sila-chain/Sila-Consensus-Core/v7/contracts/deposit"
	"github.com/sila-chain/Sila-Consensus-Core/v7/contracts/deposit/mock"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	"github.com/sila-chain/Sila/common"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

func TestProcessDepositLog_OK(t *testing.T) {
	hook := logTest.NewGlobal()

	testAcc, err := mock.Setup()
	require.NoError(t, err, "Unable to set up simulated backend")

	beaconDB := testDB.SetupDB(t)
	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)

	server, endpoint, err := mockSila.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(t.Context(),
		WithHttpEndpoint(endpoint),
		WithSilaDepositAddress(testAcc.ContractAddr),
		WithDatabase(beaconDB),
		WithDepositCache(depositCache),
	)
	require.NoError(t, err, "unable to setup web3 SILAEXEC.0 chain service")
	web3Service = setDefaultMocks(web3Service)
	web3Service.silaDepositCaller, err = contracts.NewSilaDepositCaller(testAcc.ContractAddr, testAcc.Backend.Client())
	require.NoError(t, err)

	testAcc.Backend.Commit()

	deposits, _, err := util.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)

	_, depositRoots, err := util.DeterministicDepositTrie(len(deposits))
	require.NoError(t, err)
	data := deposits[0].Data

	testAcc.TxOpts.Value = mock.Amount32Eth()
	testAcc.TxOpts.GasLimit = 1000000
	_, err = testAcc.Contract.Deposit(testAcc.TxOpts, data.PublicKey, data.WithdrawalCredentials, data.Signature, depositRoots[0])
	require.NoError(t, err, "Could not deposit to sila deposit")

	testAcc.Backend.Commit()

	query := sila.FilterQuery{
		Addresses: []common.Address{
			web3Service.cfg.silaDepositAddr,
		},
	}

	logs, err := testAcc.Backend.Client().FilterLogs(web3Service.ctx, query)
	require.NoError(t, err, "Unable to retrieve logs")

	if len(logs) == 0 {
		t.Fatal("no logs")
	}

	err = web3Service.ProcessLog(t.Context(), &logs[0])
	require.NoError(t, err)

	require.LogsDoNotContain(t, hook, "Could not unpack log")
	require.LogsDoNotContain(t, hook, "Could not save in trie")
	require.LogsDoNotContain(t, hook, "could not deserialize validator public key")
	require.LogsDoNotContain(t, hook, "could not convert bytes to signature")
	require.LogsDoNotContain(t, hook, "could not sign root for deposit data")
	require.LogsDoNotContain(t, hook, "deposit signature did not verify")
	require.LogsDoNotContain(t, hook, "could not tree hash deposit data")
	require.LogsDoNotContain(t, hook, "deposit merkle branch of deposit root did not verify for root")
	require.LogsContain(t, hook, "Deposit registered from sila deposit")

	hook.Reset()
}

func TestProcessDepositLog_InsertsPendingDeposit(t *testing.T) {
	hook := logTest.NewGlobal()
	testAcc, err := mock.Setup()
	require.NoError(t, err, "Unable to set up simulated backend")
	beaconDB := testDB.SetupDB(t)
	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)
	server, endpoint, err := mockSila.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})

	web3Service, err := NewService(t.Context(),
		WithHttpEndpoint(endpoint),
		WithSilaDepositAddress(testAcc.ContractAddr),
		WithDatabase(beaconDB),
		WithDepositCache(depositCache),
	)
	require.NoError(t, err, "unable to setup web3 SILAEXEC.0 chain service")
	web3Service = setDefaultMocks(web3Service)
	web3Service.silaDepositCaller, err = contracts.NewSilaDepositCaller(testAcc.ContractAddr, testAcc.Backend.Client())
	require.NoError(t, err)

	testAcc.Backend.Commit()

	deposits, _, err := util.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	_, depositRoots, err := util.DeterministicDepositTrie(len(deposits))
	require.NoError(t, err)
	data := deposits[0].Data

	testAcc.TxOpts.Value = mock.Amount32Eth()
	testAcc.TxOpts.GasLimit = 1000000

	_, err = testAcc.Contract.Deposit(testAcc.TxOpts, data.PublicKey, data.WithdrawalCredentials, data.Signature, depositRoots[0])
	require.NoError(t, err, "Could not deposit to sila deposit")

	_, err = testAcc.Contract.Deposit(testAcc.TxOpts, data.PublicKey, data.WithdrawalCredentials, data.Signature, depositRoots[0])
	require.NoError(t, err, "Could not deposit to sila deposit")

	testAcc.Backend.Commit()

	query := sila.FilterQuery{
		Addresses: []common.Address{
			web3Service.cfg.silaDepositAddr,
		},
	}

	logs, err := testAcc.Backend.Client().FilterLogs(web3Service.ctx, query)
	require.NoError(t, err, "Unable to retrieve logs")

	web3Service.chainStartData.Chainstarted = true

	err = web3Service.ProcessDepositLog(t.Context(), &logs[0])
	require.NoError(t, err)
	err = web3Service.ProcessDepositLog(t.Context(), &logs[1])
	require.NoError(t, err)

	pendingDeposits := web3Service.cfg.depositCache.PendingDeposits(t.Context(), nil /*blockNum*/)
	require.Equal(t, 2, len(pendingDeposits), "Unexpected number of deposits")

	hook.Reset()
}

func TestUnpackDepositLogData_OK(t *testing.T) {
	testAcc, err := mock.Setup()
	require.NoError(t, err, "Unable to set up simulated backend")
	beaconDB := testDB.SetupDB(t)
	server, endpoint, err := mockSila.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(t.Context(),
		WithHttpEndpoint(endpoint),
		WithSilaDepositAddress(testAcc.ContractAddr),
		WithDatabase(beaconDB),
	)
	require.NoError(t, err, "unable to setup web3 SILAEXEC.0 chain service")
	web3Service = setDefaultMocks(web3Service)
	web3Service.silaDepositCaller, err = contracts.NewSilaDepositCaller(testAcc.ContractAddr, testAcc.Backend.Client())
	require.NoError(t, err)

	testAcc.Backend.Commit()

	deposits, _, err := util.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	_, depositRoots, err := util.DeterministicDepositTrie(len(deposits))
	require.NoError(t, err)
	data := deposits[0].Data

	testAcc.TxOpts.Value = mock.Amount32Eth()
	testAcc.TxOpts.GasLimit = 1000000
	_, err = testAcc.Contract.Deposit(testAcc.TxOpts, data.PublicKey, data.WithdrawalCredentials, data.Signature, depositRoots[0])
	require.NoError(t, err, "Could not deposit to sila deposit")
	testAcc.Backend.Commit()

	query := sila.FilterQuery{
		Addresses: []common.Address{
			web3Service.cfg.silaDepositAddr,
		},
	}

	logz, err := testAcc.Backend.Client().FilterLogs(web3Service.ctx, query)
	require.NoError(t, err, "Unable to retrieve logs")

	loggedPubkey, withCreds, _, loggedSig, index, err := contracts.UnpackDepositLogData(logz[0].Data)
	require.NoError(t, err, "Unable to unpack logs")

	require.Equal(t, uint64(0), binary.LittleEndian.Uint64(index), "Retrieved merkle tree index is incorrect")
	require.DeepEqual(t, data.PublicKey, loggedPubkey, "Pubkey is not the same as the data that was put in")
	require.DeepEqual(t, data.Signature, loggedSig, "Proof of Possession is not the same as the data that was put in")
	require.DeepEqual(t, data.WithdrawalCredentials, withCreds, "Withdrawal Credentials is not the same as the data that was put in")
}

func TestProcessSilaGenesisLog_8DuplicatePubkeys(t *testing.T) {
	hook := logTest.NewGlobal()
	testAcc, err := mock.Setup()
	require.NoError(t, err, "Unable to set up simulated backend")
	beaconDB := testDB.SetupDB(t)
	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)
	server, endpoint, err := mockSila.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})

	web3Service, err := NewService(t.Context(),
		WithHttpEndpoint(endpoint),
		WithSilaDepositAddress(testAcc.ContractAddr),
		WithDatabase(beaconDB),
		WithDepositCache(depositCache),
	)
	require.NoError(t, err, "unable to setup web3 SILAEXEC.0 chain service")
	web3Service = setDefaultMocks(web3Service)
	web3Service.silaDepositCaller, err = contracts.NewSilaDepositCaller(testAcc.ContractAddr, testAcc.Backend.Client())
	require.NoError(t, err)

	params.SetupTestConfigCleanup(t)
	bConfig := params.MinimalSpecConfig().Copy()
	bConfig.MinGenesisTime = 0
	params.OverrideBeaconConfig(bConfig)

	testAcc.Backend.Commit()
	require.NoError(t, testAcc.Backend.AdjustTime(time.Duration(int64(time.Now().Nanosecond()))))

	deposits, _, err := util.DeterministicDepositsAndKeys(1)
	require.NoError(t, err)
	_, depositRoots, err := util.DeterministicDepositTrie(len(deposits))
	require.NoError(t, err)
	data := deposits[0].Data

	testAcc.TxOpts.Value = mock.Amount32Eth()
	testAcc.TxOpts.GasLimit = 1000000

	// 64 Validators are used as size required for beacon-chain to start. This number
	// is defined in the sila deposit as the number required for the testnet. The actual number
	// is 2**14
	for range depositsReqForChainStart {
		testAcc.TxOpts.Value = mock.Amount32Eth()
		_, err = testAcc.Contract.Deposit(testAcc.TxOpts, data.PublicKey, data.WithdrawalCredentials, data.Signature, depositRoots[0])
		require.NoError(t, err, "Could not deposit to sila deposit")

		testAcc.Backend.Commit()
	}

	query := sila.FilterQuery{
		Addresses: []common.Address{
			web3Service.cfg.silaDepositAddr,
		},
	}

	logs, err := testAcc.Backend.Client().FilterLogs(web3Service.ctx, query)
	require.NoError(t, err, "Unable to retrieve logs")

	for i := range logs {
		err = web3Service.ProcessLog(t.Context(), &logs[i])
		require.NoError(t, err)
	}
	assert.Equal(t, false, web3Service.chainStartData.Chainstarted, "Genesis has been triggered despite being 8 duplicate keys")

	require.LogsDoNotContain(t, hook, "Minimum number of validators reached for beacon-chain to start")
	hook.Reset()
}

func TestProcessSilaGenesisLog(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.GenesisDelay = 0
	params.OverrideBeaconConfig(cfg)
	hook := logTest.NewGlobal()
	testAcc, err := mock.Setup()
	require.NoError(t, err, "Unable to set up simulated backend")
	beaconDB := testDB.SetupDB(t)
	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)

	server, endpoint, err := mockSila.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(t.Context(),
		WithHttpEndpoint(endpoint),
		WithSilaDepositAddress(testAcc.ContractAddr),
		WithDatabase(beaconDB),
		WithDepositCache(depositCache),
	)
	require.NoError(t, err, "unable to setup web3 SILAEXEC.0 chain service")
	web3Service = setDefaultMocks(web3Service)
	web3Service.silaDepositCaller, err = contracts.NewSilaDepositCaller(testAcc.ContractAddr, testAcc.Backend.Client())
	web3Service.rpcClient = &mockSila.RPCClient{Backend: testAcc.Backend}
	require.NoError(t, err)
	params.SetupTestConfigCleanup(t)
	bConfig := params.MinimalSpecConfig().Copy()
	bConfig.MinGenesisTime = 0
	params.OverrideBeaconConfig(bConfig)

	testAcc.Backend.Commit()
	require.NoError(t, testAcc.Backend.AdjustTime(time.Duration(int64(time.Now().Nanosecond()))))

	deposits, _, err := util.DeterministicDepositsAndKeys(uint64(depositsReqForChainStart))
	require.NoError(t, err)
	_, roots, err := util.DeterministicDepositTrie(len(deposits))
	require.NoError(t, err)

	// 64 Validators are used as size required for beacon-chain to start. This number
	// is defined in the sila deposit as the number required for the testnet. The actual number
	// is 2**14
	for i := range depositsReqForChainStart {
		data := deposits[i].Data
		testAcc.TxOpts.Value = mock.Amount32Eth()
		testAcc.TxOpts.GasLimit = 1000000
		_, err = testAcc.Contract.Deposit(testAcc.TxOpts, data.PublicKey, data.WithdrawalCredentials, data.Signature, roots[i])
		require.NoError(t, err, "Could not deposit to sila deposit")

		testAcc.Backend.Commit()
	}

	query := sila.FilterQuery{
		Addresses: []common.Address{
			web3Service.cfg.silaDepositAddr,
		},
	}

	logs, err := testAcc.Backend.Client().FilterLogs(web3Service.ctx, query)
	require.NoError(t, err, "Unable to retrieve logs")
	require.Equal(t, depositsReqForChainStart, len(logs))

	// Set up our subscriber now to listen for the chain started event.
	stateChannel := make(chan *feed.Event, 1)
	stateSub := web3Service.cfg.stateNotifier.StateFeed().Subscribe(stateChannel)
	defer stateSub.Unsubscribe()

	for i := range logs {
		err = web3Service.ProcessLog(t.Context(), &logs[i])
		require.NoError(t, err)
	}

	err = web3Service.ProcessSilaBlock(t.Context(), big.NewInt(int64(logs[len(logs)-1].BlockNumber)))
	require.NoError(t, err)

	cachedDeposits := web3Service.chainStartData.ChainstartDeposits
	require.Equal(t, depositsReqForChainStart, len(cachedDeposits))

	// Receive the chain started event.
	for started := false; !started; {
		event := <-stateChannel
		if event.Type == statefeed.ChainStarted {
			started = true
		}
	}

	require.LogsDoNotContain(t, hook, "Unable to unpack ChainStart log data")
	require.LogsDoNotContain(t, hook, "Receipt root from log doesn't match the root saved in memory")
	require.LogsDoNotContain(t, hook, "Invalid timestamp from log")
	require.LogsContain(t, hook, "Minimum number of validators reached for beacon-chain to start")

	hook.Reset()
}

func TestProcessSilaGenesisLog_CorrectNumOfDeposits(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	hook := logTest.NewGlobal()
	testAcc, err := mock.Setup()
	require.NoError(t, err, "Unable to set up simulated backend")
	kvStore := testDB.SetupDB(t)
	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)
	server, endpoint, err := mockSila.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})

	web3Service, err := NewService(t.Context(),
		WithHttpEndpoint(endpoint),
		WithSilaDepositAddress(testAcc.ContractAddr),
		WithDatabase(kvStore),
		WithDepositCache(depositCache),
	)
	require.NoError(t, err, "unable to setup web3 SILAEXEC.0 chain service")
	web3Service = setDefaultMocks(web3Service)
	web3Service.silaDepositCaller, err = contracts.NewSilaDepositCaller(testAcc.ContractAddr, testAcc.Backend.Client())
	require.NoError(t, err)
	web3Service.rpcClient = &mockSila.RPCClient{Backend: testAcc.Backend}
	web3Service.httpLogger = testAcc.Backend.Client()
	web3Service.latestSilaData.LastRequestedBlock = 0
	block, err := testAcc.Backend.Client().BlockByNumber(t.Context(), nil)
	require.NoError(t, err)
	web3Service.latestSilaData.BlockHeight = block.NumberU64()
	web3Service.latestSilaData.BlockTime = block.Time()
	bConfig := params.MinimalSpecConfig().Copy()
	bConfig.MinGenesisTime = 0
	bConfig.SecondsPerSilaBlock = 1
	params.OverrideBeaconConfig(bConfig)
	nConfig := params.BeaconNetworkConfig()
	nConfig.ContractDeploymentBlock = 0
	params.OverrideBeaconNetworkConfig(nConfig)

	testAcc.Backend.Commit()

	totalNumOfDeposits := depositsReqForChainStart + 30

	deposits, _, err := util.DeterministicDepositsAndKeys(uint64(totalNumOfDeposits))
	require.NoError(t, err)
	_, depositRoots, err := util.DeterministicDepositTrie(len(deposits))
	require.NoError(t, err)
	depositOffset := 5

	// 64 Validators are used as size required for beacon-chain to start. This number
	// is defined in the sila deposit as the number required for the testnet. The actual number
	// is 2**14
	for i := range totalNumOfDeposits {
		data := deposits[i].Data
		testAcc.TxOpts.Value = mock.Amount32Eth()
		testAcc.TxOpts.GasLimit = 1000000
		_, err = testAcc.Contract.Deposit(testAcc.TxOpts, data.PublicKey, data.WithdrawalCredentials, data.Signature, depositRoots[i])
		require.NoError(t, err, "Could not deposit to sila deposit")
		// pack 8 deposits into a block with an offset of
		// 5
		if (i+1)%8 == depositOffset {
			testAcc.Backend.Commit()
		}
	}
	// Forward the chain to account for the follow distance
	for i := uint64(0); i < params.BeaconConfig().SilaExecutionFollowDistance; i++ {
		testAcc.Backend.Commit()
	}
	b, err := testAcc.Backend.Client().BlockByNumber(t.Context(), nil)
	require.NoError(t, err)
	web3Service.latestSilaData.BlockHeight = b.NumberU64()
	web3Service.latestSilaData.BlockTime = b.Time()

	// Set up our subscriber now to listen for the chain started event.
	stateChannel := make(chan *feed.Event, 1)
	stateSub := web3Service.cfg.stateNotifier.StateFeed().Subscribe(stateChannel)
	defer stateSub.Unsubscribe()

	err = web3Service.processPastLogs(t.Context())
	require.NoError(t, err)

	cachedDeposits := web3Service.chainStartData.ChainstartDeposits
	requiredDepsForChainstart := depositsReqForChainStart + depositOffset
	require.Equal(t, requiredDepsForChainstart, len(cachedDeposits), "Did not cache the chain start deposits correctly")

	// Receive the chain started event.
	for started := false; !started; {
		event := <-stateChannel
		if event.Type == statefeed.ChainStarted {
			started = true
		}
	}

	require.LogsDoNotContain(t, hook, "Unable to unpack ChainStart log data")
	require.LogsDoNotContain(t, hook, "Receipt root from log doesn't match the root saved in memory")
	require.LogsDoNotContain(t, hook, "Invalid timestamp from log")
	require.LogsContain(t, hook, "Minimum number of validators reached for beacon-chain to start")

	hook.Reset()
}

func TestProcessLogs_DepositRequestsStarted(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	hook := logTest.NewGlobal()
	testAcc, err := mock.Setup()
	require.NoError(t, err, "Unable to set up simulated backend")
	kvStore := testDB.SetupDB(t)
	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)
	server, endpoint, err := mockSila.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})

	web3Service, err := NewService(t.Context(),
		WithHttpEndpoint(endpoint),
		WithSilaDepositAddress(testAcc.ContractAddr),
		WithDatabase(kvStore),
		WithDepositCache(depositCache),
	)
	require.NoError(t, err, "unable to setup web3 SILAEXEC.0 chain service")
	web3Service = setDefaultMocks(web3Service)
	web3Service.silaDepositCaller, err = contracts.NewSilaDepositCaller(testAcc.ContractAddr, testAcc.Backend.Client())
	require.NoError(t, err)
	web3Service.rpcClient = &mockSila.RPCClient{Backend: testAcc.Backend}
	web3Service.httpLogger = testAcc.Backend.Client()
	web3Service.latestSilaData.LastRequestedBlock = 0
	block, err := testAcc.Backend.Client().BlockByNumber(t.Context(), nil)
	require.NoError(t, err)
	web3Service.latestSilaData.BlockHeight = block.NumberU64()
	web3Service.latestSilaData.BlockTime = block.Time()
	bConfig := params.MinimalSpecConfig().Copy()
	bConfig.MinGenesisTime = 0
	bConfig.SecondsPerSilaBlock = 1
	params.OverrideBeaconConfig(bConfig)
	nConfig := params.BeaconNetworkConfig()
	nConfig.ContractDeploymentBlock = 0
	params.OverrideBeaconNetworkConfig(nConfig)

	testAcc.Backend.Commit()

	totalNumOfDeposits := depositsReqForChainStart + 30

	deposits, _, err := util.DeterministicDepositsAndKeys(uint64(totalNumOfDeposits))
	require.NoError(t, err)
	_, depositRoots, err := util.DeterministicDepositTrie(len(deposits))
	require.NoError(t, err)
	depositOffset := 5

	// 64 Validators are used as size required for beacon-chain to start. This number
	// is defined in the sila deposit as the number required for the testnet. The actual number
	// is 2**14
	for i := range totalNumOfDeposits {
		data := deposits[i].Data
		testAcc.TxOpts.Value = mock.Amount32Eth()
		testAcc.TxOpts.GasLimit = 1000000
		_, err = testAcc.Contract.Deposit(testAcc.TxOpts, data.PublicKey, data.WithdrawalCredentials, data.Signature, depositRoots[i])
		require.NoError(t, err, "Could not deposit to sila deposit")
		// pack 8 deposits into a block with an offset of
		// 5
		if (i+1)%8 == depositOffset {
			testAcc.Backend.Commit()
		}
	}
	// Forward the chain to account for the follow distance
	for i := uint64(0); i < params.BeaconConfig().SilaExecutionFollowDistance; i++ {
		testAcc.Backend.Commit()
	}
	b, err := testAcc.Backend.Client().BlockByNumber(t.Context(), nil)
	require.NoError(t, err)
	web3Service.latestSilaData.BlockHeight = b.NumberU64()
	web3Service.latestSilaData.BlockTime = b.Time()

	// Set up our subscriber now to listen for the chain started event.
	stateChannel := make(chan *feed.Event, 1)
	stateSub := web3Service.cfg.stateNotifier.StateFeed().Subscribe(stateChannel)
	defer stateSub.Unsubscribe()

	web3Service.depositRequestsStarted = true
	web3Service.initPOWService()
	require.NoError(t, err)

	require.Equal(t, int64(-1), web3Service.lastReceivedMerkleIndex, "Processed deposit logs even when requests are active")

	hook.Reset()
}

func TestProcessSilaGenesisLog_LargePeriodOfNoLogs(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	hook := logTest.NewGlobal()
	testAcc, err := mock.Setup()
	require.NoError(t, err, "Unable to set up simulated backend")
	kvStore := testDB.SetupDB(t)
	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)
	server, endpoint, err := mockSila.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})

	web3Service, err := NewService(t.Context(),
		WithHttpEndpoint(endpoint),
		WithSilaDepositAddress(testAcc.ContractAddr),
		WithDatabase(kvStore),
		WithDepositCache(depositCache),
	)
	require.NoError(t, err, "unable to setup web3 SILAEXEC.0 chain service")
	web3Service = setDefaultMocks(web3Service)
	web3Service.silaDepositCaller, err = contracts.NewSilaDepositCaller(testAcc.ContractAddr, testAcc.Backend.Client())
	require.NoError(t, err)
	web3Service.rpcClient = &mockSila.RPCClient{Backend: testAcc.Backend}
	web3Service.httpLogger = testAcc.Backend.Client()
	web3Service.latestSilaData.LastRequestedBlock = 0
	b, err := testAcc.Backend.Client().BlockByNumber(t.Context(), nil)
	require.NoError(t, err)
	web3Service.latestSilaData.BlockHeight = b.NumberU64()
	web3Service.latestSilaData.BlockTime = b.Time()
	bConfig := params.MinimalSpecConfig().Copy()
	bConfig.SecondsPerSilaBlock = 10
	params.OverrideBeaconConfig(bConfig)
	nConfig := params.BeaconNetworkConfig()
	nConfig.ContractDeploymentBlock = 0
	params.OverrideBeaconNetworkConfig(nConfig)

	testAcc.Backend.Commit()

	totalNumOfDeposits := depositsReqForChainStart + 30

	deposits, _, err := util.DeterministicDepositsAndKeys(uint64(totalNumOfDeposits))
	require.NoError(t, err)
	_, depositRoots, err := util.DeterministicDepositTrie(len(deposits))
	require.NoError(t, err)
	depositOffset := 5

	// 64 Validators are used as size required for beacon-chain to start. This number
	// is defined in the sila deposit as the number required for the testnet. The actual number
	// is 2**14
	for i := range totalNumOfDeposits {
		data := deposits[i].Data
		testAcc.TxOpts.Value = mock.Amount32Eth()
		testAcc.TxOpts.GasLimit = 1000000
		_, err = testAcc.Contract.Deposit(testAcc.TxOpts, data.PublicKey, data.WithdrawalCredentials, data.Signature, depositRoots[i])
		require.NoError(t, err, "Could not deposit to sila deposit")
		// pack 8 deposits into a block with an offset of
		// 5
		if (i+1)%8 == depositOffset {
			testAcc.Backend.Commit()
		}
	}
	// Forward the chain to 'mine' blocks without logs
	for range uint64(1500) {
		testAcc.Backend.Commit()
	}
	genesisBlock, err := testAcc.Backend.Client().BlockByNumber(t.Context(), nil)
	require.NoError(t, err)

	wantedGenesisTime := genesisBlock.Time()

	// Forward the chain to account for the follow distance
	for i := uint64(0); i < params.BeaconConfig().SilaExecutionFollowDistance; i++ {
		testAcc.Backend.Commit()
	}
	currBlock, err := testAcc.Backend.Client().BlockByNumber(t.Context(), nil)
	require.NoError(t, err)
	web3Service.latestSilaData.BlockHeight = currBlock.NumberU64()
	web3Service.latestSilaData.BlockTime = currBlock.Time()

	// Set the genesis time 500 blocks ahead of the last
	// deposit log.
	bConfig = params.MinimalSpecConfig().Copy()
	bConfig.MinGenesisTime = wantedGenesisTime - 10
	params.OverrideBeaconConfig(bConfig)

	// Set up our subscriber now to listen for the chain started event.
	stateChannel := make(chan *feed.Event, 1)
	stateSub := web3Service.cfg.stateNotifier.StateFeed().Subscribe(stateChannel)
	defer stateSub.Unsubscribe()

	err = web3Service.processPastLogs(t.Context())
	require.NoError(t, err)

	cachedDeposits := web3Service.chainStartData.ChainstartDeposits
	require.Equal(t, totalNumOfDeposits, len(cachedDeposits), "Did not cache the chain start deposits correctly")

	// Receive the chain started event.
	for started := false; !started; {
		event := <-stateChannel
		if event.Type == statefeed.ChainStarted {
			started = true
		}
	}

	require.LogsDoNotContain(t, hook, "Unable to unpack ChainStart log data")
	require.LogsDoNotContain(t, hook, "Receipt root from log doesn't match the root saved in memory")
	require.LogsDoNotContain(t, hook, "Invalid timestamp from log")
	require.LogsContain(t, hook, "Minimum number of validators reached for beacon-chain to start")

	hook.Reset()
}

func TestCheckForChainstart_NoValidator(t *testing.T) {
	hook := logTest.NewGlobal()
	testAcc, err := mock.Setup()
	require.NoError(t, err, "Unable to set up simulated backend")
	beaconDB := testDB.SetupDB(t)
	s := newPowchainService(t, testAcc, beaconDB)
	s.processChainStartIfReady(t.Context(), [32]byte{}, nil, 0)
	require.LogsDoNotContain(t, hook, "Could not determine active validator count from pre genesis state")
}

func newPowchainService(t *testing.T, silaexecBackend *mock.TestAccount, beaconDB db.Database) *Service {
	depositCache, err := depositsnapshot.New()
	require.NoError(t, err)
	server, endpoint, err := mockSila.SetupRPCServer()
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Stop()
	})
	web3Service, err := NewService(t.Context(),
		WithHttpEndpoint(endpoint),
		WithSilaDepositAddress(silaexecBackend.ContractAddr),
		WithDatabase(beaconDB),
		WithDepositCache(depositCache),
	)
	require.NoError(t, err, "unable to setup web3 SILAEXEC.0 chain service")
	web3Service = setDefaultMocks(web3Service)
	web3Service.silaDepositCaller, err = contracts.NewSilaDepositCaller(silaexecBackend.ContractAddr, silaexecBackend.Backend.Client())
	require.NoError(t, err)

	web3Service.rpcClient = &mockSila.RPCClient{Backend: silaexecBackend.Backend}
	web3Service.httpLogger = &goodLogger{backend: silaexecBackend.Backend}
	params.SetupTestConfigCleanup(t)
	bConfig := params.MinimalSpecConfig().Copy()
	bConfig.MinGenesisTime = 0
	params.OverrideBeaconConfig(bConfig)
	return web3Service
}
