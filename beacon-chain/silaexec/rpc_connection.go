package silaexec

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sila-chain/Sila"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	contracts "github.com/sila-chain/Sila-Consensus-Core/v7/contracts/deposit"
	"github.com/sila-chain/Sila-Consensus-Core/v7/io/logs"
	"github.com/sila-chain/Sila-Consensus-Core/v7/network"
	"github.com/sila-chain/Sila-Consensus-Core/v7/network/authorization"
	"github.com/sila-chain/Sila/common"
	"github.com/sila-chain/Sila/common/hexutil"
	silaTypes "github.com/sila-chain/Sila/core/types"
	silaRPC "github.com/sila-chain/Sila/rpc"
	"github.com/sila-chain/Sila/silaclient"
)

type silaCallArg struct {
	From     *common.Address `json:"from,omitempty"`
	To       *common.Address `json:"to,omitempty"`
	Gas      hexutil.Uint64  `json:"gas,omitempty"`
	GasPrice *hexutil.Big    `json:"gasPrice,omitempty"`
	Value    *hexutil.Big    `json:"value,omitempty"`
	Data     hexutil.Bytes   `json:"data,omitempty"`
}

type silaLogFilterer struct {
	client *silaRPC.Client
}

func (f *silaLogFilterer) FilterLogs(ctx context.Context, q sila.FilterQuery) ([]silaTypes.Log, error) {
	arg := toSilaFilterArg(q)
	var result []silaTypes.Log
	if err := f.client.CallContext(ctx, &result, "sila_getLogs", arg); err != nil {
		return nil, err
	}
	return result, nil
}

func (f *silaLogFilterer) SubscribeFilterLogs(context.Context, sila.FilterQuery, chan<- silaTypes.Log) (sila.Subscription, error) {
	return nil, errors.New("sila log subscriptions are not implemented")
}

func toSilaFilterArg(q sila.FilterQuery) map[string]any {
	arg := map[string]any{}

	if q.BlockHash != nil {
		arg["blockHash"] = *q.BlockHash
	} else {
		if q.FromBlock != nil {
			arg["fromBlock"] = toSilaBlockNumArg(q.FromBlock)
		}
		if q.ToBlock != nil {
			arg["toBlock"] = toSilaBlockNumArg(q.ToBlock)
		}
	}

	switch len(q.Addresses) {
	case 0:
	case 1:
		arg["address"] = q.Addresses[0]
	default:
		arg["address"] = q.Addresses
	}

	if len(q.Topics) > 0 {
		topics := make([]any, len(q.Topics))
		for i, topicSet := range q.Topics {
			switch len(topicSet) {
			case 0:
				topics[i] = nil
			case 1:
				topics[i] = topicSet[0]
			default:
				topics[i] = topicSet
			}
		}
		arg["topics"] = topics
	}

	return arg
}

type silaContractCaller struct {
	client *silaRPC.Client
}

func (c *silaContractCaller) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	var result hexutil.Bytes
	if err := c.client.CallContext(ctx, &result, "sila_getCode", contract, toSilaBlockNumArg(blockNumber)); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *silaContractCaller) CallContract(ctx context.Context, call sila.CallMsg, blockNumber *big.Int) ([]byte, error) {
	from := call.From
	arg := silaCallArg{
		From:     &from,
		To:       call.To,
		Gas:      hexutil.Uint64(call.Gas),
		GasPrice: (*hexutil.Big)(call.GasPrice),
		Value:    (*hexutil.Big)(call.Value),
		Data:     call.Data,
	}
	var result hexutil.Bytes
	if err := c.client.CallContext(ctx, &result, "sila_call", arg, toSilaBlockNumArg(blockNumber)); err != nil {
		return nil, err
	}
	return result, nil
}

func toSilaBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}
func (s *Service) setupSilaClientConnections(ctx context.Context, currEndpoint network.Endpoint) error {
	client, err := s.newRPCClientWithAuth(ctx, currEndpoint)
	if err != nil {
		return errors.Wrap(err, "could not dial Sila node")
	}
	// Attach the clients to the service struct.
	fetcher := silaclient.NewClient(client)
	s.rpcClient = client
	s.httpLogger = &silaLogFilterer{client: client}

	silaDepositCaller, err := contracts.NewSilaDepositCaller(s.cfg.silaDepositAddr, &silaContractCaller{client: client})
	if err != nil {
		client.Close()
		return errors.Wrap(err, "could not initialize sila deposit caller")
	}
	s.silaDepositCaller = silaDepositCaller

	// Ensure we have the correct chain and deposit IDs.
	if err := ensureCorrectSilaChain(ctx, fetcher); err != nil {
		client.Close()
		errStr := err.Error()
		if strings.Contains(errStr, "401 Unauthorized") {
			errStr = "could not verify Sila chain ID as your connection is not authenticated. " +
				"If connecting to your Sila client via HTTP, you will need to set up JWT authentication. " +
				"See our documentation here https://docs.prylabs.network/docs/execution-node/authentication"
		}
		return errors.Wrap(err, errStr)
	}
	s.updateConnectedSilaExecution(true)
	s.runError = nil
	return nil
}

// Every N seconds, defined as a backoffPeriod, attempts to re-establish a Sila client
// connection and if this does not work, we fallback to the next endpoint if defined.
func (s *Service) pollConnectionStatus(ctx context.Context) {
	// Use a custom logger to only log errors
	logCounter := 0
	errorLogger := func(err error, msg string) {
		if logCounter > logThreshold {
			log.WithError(err).Error(msg)
			logCounter = 0
		}
		logCounter++
	}
	ticker := time.NewTicker(backOffPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Debugf("Trying to dial endpoint: %s", logs.MaskCredentialsLogging(s.cfg.currHttpEndpoint.Url))
			currClient := s.rpcClient
			if err := s.setupSilaClientConnections(ctx, s.cfg.currHttpEndpoint); err != nil {
				errorLogger(err, "Could not connect to Sila client endpoint")
				continue
			}
			// Close previous client, if connection was successful.
			if currClient != nil {
				currClient.Close()
			}
			log.WithField("endpoint", logs.MaskCredentialsLogging(s.cfg.currHttpEndpoint.Url)).Info("Connected to new endpoint")

			c, err := s.ExchangeCapabilities(ctx)
			if err != nil {
				errorLogger(err, "Could not exchange capabilities with Sila client")
			}
			s.capabilityCache.save(c)

			return
		case <-s.ctx.Done():
			log.Debug("Received cancelled context,closing existing powchain service")
			return
		}
	}
}

// Forces to retry a Sila client connection.
func (s *Service) retrySilaClientConnection(ctx context.Context, err error) {
	s.runError = errors.Wrap(err, "retrySilaClientConnection")
	s.updateConnectedSilaExecution(false)
	// Back off for a while before redialing.
	time.Sleep(backOffPeriod)
	currClient := s.rpcClient
	if err := s.setupSilaClientConnections(ctx, s.cfg.currHttpEndpoint); err != nil {
		s.runError = errors.Wrap(err, "setupSilaClientConnections")
		return
	}
	// Close previous client, if connection was successful.
	if currClient != nil {
		currClient.Close()
	}
	// Reset run error in the event of a successful connection.
	s.runError = nil
}

// Initializes an RPC connection with authentication headers.
func (s *Service) newRPCClientWithAuth(ctx context.Context, endpoint network.Endpoint) (*silaRPC.Client, error) {
	headers := http.Header{}
	if endpoint.Auth.Method != authorization.None {
		header, err := endpoint.Auth.ToHeaderValue()
		if err != nil {
			return nil, err
		}
		headers.Set("Authorization", header)
	}
	for _, h := range s.cfg.headers {
		if h == "" {
			continue
		}
		keyValue := strings.Split(h, "=")
		if len(keyValue) < 2 {
			log.Warnf("Incorrect HTTP header flag format. Skipping %v", keyValue[0])
			continue
		}
		headers.Set(keyValue[0], strings.Join(keyValue[1:], "="))
	}
	return network.NewExecutionRPCClient(ctx, endpoint, headers)
}

// Checks the chain ID of the Sila client to ensure
// it matches local parameters of what Sila expects.
func ensureCorrectSilaChain(ctx context.Context, client *silaclient.Client) error {
	var chainIDHex string
	if err := client.Client().CallContext(ctx, &chainIDHex, "sila_chainId"); err != nil {
		return err
	}
	cID, ok := new(big.Int).SetString(strings.TrimPrefix(chainIDHex, "0x"), 16)
	if !ok {
		return fmt.Errorf("invalid Sila chain ID %q", chainIDHex)
	}
	wantChainID := params.BeaconConfig().DepositChainID
	if cID.Uint64() != wantChainID {
		return fmt.Errorf("wanted Sila chain ID %d, got %d", wantChainID, cID.Uint64())
	}
	return nil
}
