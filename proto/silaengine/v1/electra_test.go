package silaenginev1_test

import (
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silaenginev1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila/common/hexutil"
)

var depositRequestsSSZHex = "0x706b0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000077630000000000000000000000000000000000000000000000000000000000007b00000000000000736967000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c801000000000000706b00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000776300000000000000000000000000000000000000000000000000000000000090010000000000007369670000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002000000000000000"

func TestGetDecodedSilaRequests(t *testing.T) {
	cfg := params.BeaconConfig()
	t.Run("All requests decode successfully", func(t *testing.T) {
		depositRequestBytes, err := hexutil.Decode("0x610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"620000000000000000000000000000000000000000000000000000000000000000" +
			"4059730700000063000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		withdrawalRequestBytes, err := hexutil.Decode("0x6400000000000000000000000000000000000000" +
			"6500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000040597307000000")
		require.NoError(t, err)
		consolidationRequestBytes, err := hexutil.Decode("0x6600000000000000000000000000000000000000" +
			"670000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"680000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		ebe := &silaenginev1.ExecutionBundleElectra{
			SilaRequests: [][]byte{append([]byte{uint8(silaenginev1.DepositRequestType)}, depositRequestBytes...),
				append([]byte{uint8(silaenginev1.WithdrawalRequestType)}, withdrawalRequestBytes...),
				append([]byte{uint8(silaenginev1.ConsolidationRequestType)}, consolidationRequestBytes...)},
		}
		requests, err := ebe.GetDecodedSilaRequests(cfg.ExecutionRequestLimits())
		require.NoError(t, err)
		require.Equal(t, len(requests.Deposits), 1)
		require.Equal(t, len(requests.Withdrawals), 1)
		require.Equal(t, len(requests.Consolidations), 1)
	})
	t.Run("Excluded requests still decode successfully when one request is missing", func(t *testing.T) {
		depositRequestBytes, err := hexutil.Decode("0x610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"620000000000000000000000000000000000000000000000000000000000000000" +
			"4059730700000063000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		consolidationRequestBytes, err := hexutil.Decode("0x6600000000000000000000000000000000000000" +
			"670000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"680000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		ebe := &silaenginev1.ExecutionBundleElectra{
			SilaRequests: [][]byte{append([]byte{uint8(silaenginev1.DepositRequestType)}, depositRequestBytes...), append([]byte{uint8(silaenginev1.ConsolidationRequestType)}, consolidationRequestBytes...)},
		}
		requests, err := ebe.GetDecodedSilaRequests(cfg.ExecutionRequestLimits())
		require.NoError(t, err)
		require.Equal(t, len(requests.Deposits), 1)
		require.Equal(t, len(requests.Withdrawals), 0)
		require.Equal(t, len(requests.Consolidations), 1)
	})
	t.Run("Decode sila requests should fail if ordering is not sorted", func(t *testing.T) {
		depositRequestBytes, err := hexutil.Decode("0x610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"620000000000000000000000000000000000000000000000000000000000000000" +
			"4059730700000063000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		consolidationRequestBytes, err := hexutil.Decode("0x6600000000000000000000000000000000000000" +
			"670000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"680000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		ebe := &silaenginev1.ExecutionBundleElectra{
			SilaRequests: [][]byte{append([]byte{uint8(silaenginev1.ConsolidationRequestType)}, consolidationRequestBytes...), append([]byte{uint8(silaenginev1.DepositRequestType)}, depositRequestBytes...)},
		}
		_, err = ebe.GetDecodedSilaRequests(cfg.ExecutionRequestLimits())
		require.ErrorContains(t, "invalid execution request type order", err)
	})
	t.Run("Requests should error if the request type is shorter than 1 byte", func(t *testing.T) {
		consolidationRequestBytes, err := hexutil.Decode("0x6600000000000000000000000000000000000000" +
			"670000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"680000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		ebe := &silaenginev1.ExecutionBundleElectra{
			SilaRequests: [][]byte{append([]byte{}, []byte{}...), append([]byte{uint8(silaenginev1.ConsolidationRequestType)}, consolidationRequestBytes...)},
		}
		_, err = ebe.GetDecodedSilaRequests(cfg.ExecutionRequestLimits())
		require.ErrorContains(t, "invalid execution request, length less than 1", err)
	})
	t.Run("a duplicate request should fail", func(t *testing.T) {
		withdrawalRequestBytes, err := hexutil.Decode("0x6400000000000000000000000000000000000000" +
			"6500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000040597307000000")
		require.NoError(t, err)
		withdrawalRequestBytes2, err := hexutil.Decode("0x6400000000000000000000000000000000000000" +
			"6500000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000040597307000000")
		require.NoError(t, err)
		ebe := &silaenginev1.ExecutionBundleElectra{
			SilaRequests: [][]byte{append([]byte{uint8(silaenginev1.WithdrawalRequestType)}, withdrawalRequestBytes...), append([]byte{uint8(silaenginev1.WithdrawalRequestType)}, withdrawalRequestBytes2...)},
		}
		_, err = ebe.GetDecodedSilaRequests(cfg.ExecutionRequestLimits())
		require.ErrorContains(t, "requests should be in sorted order and unique", err)
	})
	t.Run("a duplicate withdrawals ( non 0 request type )request should fail", func(t *testing.T) {
		depositRequestBytes, err := hexutil.Decode("0x610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"620000000000000000000000000000000000000000000000000000000000000000" +
			"4059730700000063000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		depositRequestBytes2, err := hexutil.Decode("0x610000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"620000000000000000000000000000000000000000000000000000000000000000" +
			"4059730700000063000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		ebe := &silaenginev1.ExecutionBundleElectra{
			SilaRequests: [][]byte{append([]byte{uint8(silaenginev1.DepositRequestType)}, depositRequestBytes...), append([]byte{uint8(silaenginev1.DepositRequestType)}, depositRequestBytes2...)},
		}
		_, err = ebe.GetDecodedSilaRequests(cfg.ExecutionRequestLimits())
		require.ErrorContains(t, "requests should be in sorted order and unique", err)
	})
	t.Run("If a request type is provided, but the request list is shorter than the ssz of 1 request we error", func(t *testing.T) {
		consolidationRequestBytes, err := hexutil.Decode("0x6600000000000000000000000000000000000000" +
			"670000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
			"680000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.NoError(t, err)
		ebe := &silaenginev1.ExecutionBundleElectra{
			SilaRequests: [][]byte{append([]byte{uint8(silaenginev1.DepositRequestType)}, []byte{}...), append([]byte{uint8(silaenginev1.ConsolidationRequestType)}, consolidationRequestBytes...)},
		}
		_, err = ebe.GetDecodedSilaRequests(cfg.ExecutionRequestLimits())
		require.ErrorContains(t, "invalid deposit requests SSZ size", err)
	})
	t.Run("If deposit requests are over the max allowed per payload then we should error", func(t *testing.T) {
		requests := make([]*silaenginev1.DepositRequest, cfg.MaxDepositRequestsPerPayload+1)
		for i := range requests {
			requests[i] = &silaenginev1.DepositRequest{
				Pubkey:                bytesutil.PadTo([]byte("pk"), 48),
				WithdrawalCredentials: bytesutil.PadTo([]byte("wc"), 32),
				Amount:                123,
				Signature:             bytesutil.PadTo([]byte("sig"), 96),
				Index:                 456,
			}
		}
		by, err := silaenginev1.MarshalItems(requests)
		require.NoError(t, err)
		ebe := &silaenginev1.ExecutionBundleElectra{
			SilaRequests: [][]byte{
				append([]byte{uint8(silaenginev1.DepositRequestType)}, by...),
			},
		}
		_, err = ebe.GetDecodedSilaRequests(cfg.ExecutionRequestLimits())
		require.ErrorContains(t, "invalid deposit requests SSZ size, requests should not be more than the max per payload", err)
	})
	t.Run("If withdrawal requests are over the max allowed per payload then we should error", func(t *testing.T) {
		requests := make([]*silaenginev1.WithdrawalRequest, cfg.MaxWithdrawalRequestsPerPayload+1)
		for i := range requests {
			requests[i] = &silaenginev1.WithdrawalRequest{
				SourceAddress:   bytesutil.PadTo([]byte("sa"), 20),
				ValidatorPubkey: bytesutil.PadTo([]byte("pk"), 48),
				Amount:          55555,
			}
		}
		by, err := silaenginev1.MarshalItems(requests)
		require.NoError(t, err)
		ebe := &silaenginev1.ExecutionBundleElectra{
			SilaRequests: [][]byte{
				append([]byte{uint8(silaenginev1.WithdrawalRequestType)}, by...),
			},
		}
		_, err = ebe.GetDecodedSilaRequests(cfg.ExecutionRequestLimits())
		require.ErrorContains(t, "invalid withdrawal requests SSZ size, requests should not be more than the max per payload", err)
	})
	t.Run("If consolidation requests are over the max allowed per payload then we should error", func(t *testing.T) {
		requests := make([]*silaenginev1.ConsolidationRequest, cfg.MaxConsolidationsRequestsPerPayload+1)
		for i := range requests {
			requests[i] = &silaenginev1.ConsolidationRequest{
				SourceAddress: bytesutil.PadTo([]byte("sa"), 20),
				SourcePubkey:  bytesutil.PadTo([]byte("pk"), 48),
				TargetPubkey:  bytesutil.PadTo([]byte("pk"), 48),
			}
		}
		by, err := silaenginev1.MarshalItems(requests)
		require.NoError(t, err)
		ebe := &silaenginev1.ExecutionBundleElectra{
			SilaRequests: [][]byte{
				append([]byte{uint8(silaenginev1.ConsolidationRequestType)}, by...),
			},
		}
		_, err = ebe.GetDecodedSilaRequests(cfg.ExecutionRequestLimits())
		require.ErrorContains(t, "invalid consolidation requests SSZ size, requests should not be more than the max per payload", err)
	})
}

func TestEncodeSilaRequests(t *testing.T) {
	t.Run("Empty sila requests should return an empty response and not nil", func(t *testing.T) {
		ebe := &silaenginev1.SilaRequests{}
		b, err := silaenginev1.EncodeSilaRequests(ebe)
		require.NoError(t, err)
		require.NotNil(t, b)
		require.Equal(t, len(b), 0)
	})
}

func TestUnmarshalItems_OK(t *testing.T) {
	drb, err := hexutil.Decode(depositRequestsSSZHex)
	require.NoError(t, err)
	exampleRequest := &silaenginev1.DepositRequest{}
	depositRequests, err := silaenginev1.UnmarshalItems(drb, exampleRequest.SizeSSZ(), func() *silaenginev1.DepositRequest { return &silaenginev1.DepositRequest{} })
	require.NoError(t, err)

	exampleRequest1 := &silaenginev1.DepositRequest{
		Pubkey:                bytesutil.PadTo([]byte("pk"), 48),
		WithdrawalCredentials: bytesutil.PadTo([]byte("wc"), 32),
		Amount:                123,
		Signature:             bytesutil.PadTo([]byte("sig"), 96),
		Index:                 456,
	}
	exampleRequest2 := &silaenginev1.DepositRequest{
		Pubkey:                bytesutil.PadTo([]byte("pk"), 48),
		WithdrawalCredentials: bytesutil.PadTo([]byte("wc"), 32),
		Amount:                400,
		Signature:             bytesutil.PadTo([]byte("sig"), 96),
		Index:                 32,
	}
	require.DeepEqual(t, depositRequests, []*silaenginev1.DepositRequest{exampleRequest1, exampleRequest2})
}

func TestMarshalItems_OK(t *testing.T) {
	exampleRequest1 := &silaenginev1.DepositRequest{
		Pubkey:                bytesutil.PadTo([]byte("pk"), 48),
		WithdrawalCredentials: bytesutil.PadTo([]byte("wc"), 32),
		Amount:                123,
		Signature:             bytesutil.PadTo([]byte("sig"), 96),
		Index:                 456,
	}
	exampleRequest2 := &silaenginev1.DepositRequest{
		Pubkey:                bytesutil.PadTo([]byte("pk"), 48),
		WithdrawalCredentials: bytesutil.PadTo([]byte("wc"), 32),
		Amount:                400,
		Signature:             bytesutil.PadTo([]byte("sig"), 96),
		Index:                 32,
	}
	drbs, err := silaenginev1.MarshalItems([]*silaenginev1.DepositRequest{exampleRequest1, exampleRequest2})
	require.NoError(t, err)
	require.DeepEqual(t, depositRequestsSSZHex, hexutil.Encode(drbs))
}

func TestEmptySilaRequestsHashTreeRoot(t *testing.T) {
	want, err := (&silaenginev1.SilaRequests{}).HashTreeRoot()
	require.NoError(t, err)
	got, err := silaenginev1.EmptySilaRequestsHashTreeRoot()
	require.NoError(t, err)
	require.Equal(t, want, got)
}
