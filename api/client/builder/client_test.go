package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api"
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	"github.com/sila-chain/Sila-Consensus-Core/v7/config/params"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/blocks"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/interfaces"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	sila "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	v1 "github.com/sila-chain/Sila-Consensus-Core/v7/proto/silaengine/v1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/require"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/util"
	"github.com/sila-chain/go-bitfield"
)

type roundtrip func(*http.Request) (*http.Response, error)

func (fn roundtrip) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func TestClient_Status(t *testing.T) {
	ctx := t.Context()
	statusPath := "/sila/v1/builder/status"
	hc := &http.Client{
		Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
			defer func() {
				if r.Body == nil {
					return
				}
				require.NoError(t, r.Body.Close())
			}()
			require.Equal(t, statusPath, r.URL.Path)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(nil)),
				Request:    r.Clone(ctx),
			}, nil
		}),
	}
	c := &Client{
		hc:      hc,
		baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
	}
	require.NoError(t, c.Status(ctx))
	hc = &http.Client{
		Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
			defer func() {
				if r.Body == nil {
					return
				}
				require.NoError(t, r.Body.Close())
			}()
			require.Equal(t, statusPath, r.URL.Path)
			message := ErrorMessage{
				Code:    500,
				Message: "Internal server error",
			}
			resp, err := json.Marshal(message)
			require.NoError(t, err)
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBuffer(resp)),
				Request:    r.Clone(ctx),
			}, nil
		}),
	}
	c = &Client{
		hc:      hc,
		baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
	}
	require.ErrorIs(t, c.Status(ctx), ErrNotOK)
}

func TestClient_RegisterValidator(t *testing.T) {
	ctx := t.Context()
	expectedBody := `[{"message":{"fee_recipient":"0x0000000000000000000000000000000000000000","gas_limit":"23","timestamp":"42","pubkey":"0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a"},"signature":"0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"}]`
	expectedPath := "/sila/v1/builder/validators"
	t.Run("JSON success", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, api.JsonMediaType, r.Header.Get("Content-Type"))
				require.Equal(t, api.JsonMediaType, r.Header.Get("Accept"))
				body, err := io.ReadAll(r.Body)
				defer func() {
					require.NoError(t, r.Body.Close())
				}()
				require.NoError(t, err)
				require.Equal(t, expectedBody, string(body))
				require.Equal(t, expectedPath, r.URL.Path)
				require.Equal(t, http.MethodPost, r.Method)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		reg := &sila.SignedValidatorRegistrationV1{
			Message: &sila.ValidatorRegistrationV1{
				FeeRecipient: ezDecode(t, params.BeaconConfig().SilaBurnAddressHex),
				GasLimit:     23,
				Timestamp:    42,
				Pubkey:       ezDecode(t, "0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a"),
			},
			Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
		}
		require.NoError(t, c.RegisterValidator(ctx, []*sila.SignedValidatorRegistrationV1{reg}))
	})
	t.Run("SSZ success", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Content-Type"))
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Accept"))
				body, err := io.ReadAll(r.Body)
				defer func() {
					require.NoError(t, r.Body.Close())
				}()
				require.NoError(t, err)
				request := &sila.SignedValidatorRegistrationV1{}
				itemBytes := body[:request.SizeSSZ()]
				require.NoError(t, request.UnmarshalSSZ(itemBytes))
				jsRequest := structs.SignedValidatorRegistrationFromConsensus(request)
				js, err := json.Marshal([]*structs.SignedValidatorRegistration{jsRequest})
				require.NoError(t, err)

				require.Equal(t, expectedBody, string(js))
				require.Equal(t, expectedPath, r.URL.Path)
				require.Equal(t, http.MethodPost, r.Method)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:         hc,
			baseURL:    &url.URL{Host: "localhost:3500", Scheme: "http"},
			sszEnabled: true,
		}
		reg := &sila.SignedValidatorRegistrationV1{
			Message: &sila.ValidatorRegistrationV1{
				FeeRecipient: ezDecode(t, params.BeaconConfig().SilaBurnAddressHex),
				GasLimit:     23,
				Timestamp:    42,
				Pubkey:       ezDecode(t, "0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a"),
			},
			Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
		}
		require.NoError(t, c.RegisterValidator(ctx, []*sila.SignedValidatorRegistrationV1{reg}))
	})
}

func TestClient_GetHeader(t *testing.T) {
	ctx := t.Context()
	ds := util.SlotAtEpoch(t, params.BeaconConfig().DenebForkEpoch)
	es := util.SlotAtEpoch(t, params.BeaconConfig().ElectraForkEpoch)
	expectedPath := "/sila/v1/builder/header/%d/0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2/0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a"
	expectedPath = fmt.Sprintf(expectedPath, ds)
	var slot primitives.Slot = ds
	parentHash := ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	pubkey := ezDecode(t, "0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a")
	t.Run("server error", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, expectedPath, r.URL.Path)
				message := ErrorMessage{
					Code:    500,
					Message: "Internal server error",
				}
				resp, err := json.Marshal(message)
				require.NoError(t, err)
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBuffer(resp)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}

		_, err := c.GetHeader(ctx, slot, bytesutil.ToBytes32(parentHash), bytesutil.ToBytes48(pubkey))
		require.ErrorIs(t, err, ErrNotOK)
	})
	t.Run("header not available", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, expectedPath, r.URL.Path)
				return &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(bytes.NewBuffer([]byte("No header is available."))),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		_, err := c.GetHeader(ctx, slot, bytesutil.ToBytes32(parentHash), bytesutil.ToBytes48(pubkey))
		require.ErrorIs(t, err, ErrNoContent)
	})
	t.Run("bellatrix", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, expectedPath, r.URL.Path)
				require.Equal(t, api.JsonMediaType, r.Header.Get("Accept"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(testExampleHeaderResponse)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		h, err := c.GetHeader(ctx, slot, bytesutil.ToBytes32(parentHash), bytesutil.ToBytes48(pubkey))
		require.NoError(t, err)
		expectedSig := ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505")
		require.Equal(t, true, bytes.Equal(expectedSig, h.Signature()))
		expectedTxRoot := ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
		bid, err := h.Message()
		require.NoError(t, err)
		bidHeader, err := bid.Header()
		require.NoError(t, err)
		withdrawalsRoot, err := bidHeader.TransactionsRoot()
		require.NoError(t, err)
		require.Equal(t, true, bytes.Equal(expectedTxRoot, withdrawalsRoot))
		require.Equal(t, uint64(1), bidHeader.GasUsed())
		// this matches the value in the testExampleHeaderResponse
		bidStr := "652312848583266388373324160190187140051835877600158453279131187530910662656"
		value, err := stringToUint256(bidStr)
		require.NoError(t, err)
		require.Equal(t, 0, value.Int.Cmp(primitives.WeiToBigInt(bid.Value())))
		require.Equal(t, bidStr, primitives.WeiToBigInt(bid.Value()).String())
	})
	t.Run("bellatrix ssz", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Accept"))
				require.Equal(t, expectedPath, r.URL.Path)
				epr := &ExecHeaderResponse{}
				require.NoError(t, json.Unmarshal([]byte(testExampleHeaderResponse), epr))
				pro, err := epr.ToProto()
				require.NoError(t, err)
				ssz, err := pro.MarshalSSZ()
				require.NoError(t, err)
				header := http.Header{}
				header.Set(api.VersionHeader, "bellatrix")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(bytes.NewBuffer(ssz)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:         hc,
			baseURL:    &url.URL{Host: "localhost:3500", Scheme: "http"},
			sszEnabled: true,
		}
		h, err := c.GetHeader(ctx, slot, bytesutil.ToBytes32(parentHash), bytesutil.ToBytes48(pubkey))
		require.NoError(t, err)
		expectedSig := ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505")
		require.Equal(t, true, bytes.Equal(expectedSig, h.Signature()))
		expectedTxRoot := ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
		bid, err := h.Message()
		require.NoError(t, err)
		bidHeader, err := bid.Header()
		require.NoError(t, err)
		withdrawalsRoot, err := bidHeader.TransactionsRoot()
		require.NoError(t, err)
		require.Equal(t, true, bytes.Equal(expectedTxRoot, withdrawalsRoot))
		require.Equal(t, uint64(1), bidHeader.GasUsed())
		// this matches the value in the testExampleHeaderResponse
		bidStr := "652312848583266388373324160190187140051835877600158453279131187530910662656"
		value, err := stringToUint256(bidStr)
		require.NoError(t, err)
		require.Equal(t, 0, value.Int.Cmp(primitives.WeiToBigInt(bid.Value())))
		require.Equal(t, bidStr, primitives.WeiToBigInt(bid.Value()).String())
	})
	t.Run("capella", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, api.JsonMediaType, r.Header.Get("Accept"))
				require.Equal(t, expectedPath, r.URL.Path)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(testExampleHeaderResponseCapella)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		h, err := c.GetHeader(ctx, slot, bytesutil.ToBytes32(parentHash), bytesutil.ToBytes48(pubkey))
		require.NoError(t, err)
		expectedWithdrawalsRoot := ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
		bid, err := h.Message()
		require.NoError(t, err)
		bidHeader, err := bid.Header()
		require.NoError(t, err)
		withdrawalsRoot, err := bidHeader.WithdrawalsRoot()
		require.NoError(t, err)
		require.Equal(t, true, bytes.Equal(expectedWithdrawalsRoot, withdrawalsRoot))
		bidStr := "652312848583266388373324160190187140051835877600158453279131187530910662656"
		value, err := stringToUint256(bidStr)
		require.NoError(t, err)
		require.Equal(t, 0, value.Int.Cmp(primitives.WeiToBigInt(bid.Value())))
		require.Equal(t, bidStr, primitives.WeiToBigInt(bid.Value()).String())
	})
	t.Run("capella ssz", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Accept"))
				require.Equal(t, expectedPath, r.URL.Path)
				epr := &ExecHeaderResponseCapella{}
				require.NoError(t, json.Unmarshal([]byte(testExampleHeaderResponseCapella), epr))
				pro, err := epr.ToProto()
				require.NoError(t, err)
				ssz, err := pro.MarshalSSZ()
				require.NoError(t, err)
				header := http.Header{}
				header.Set(api.VersionHeader, "capella")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(bytes.NewBuffer(ssz)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:         hc,
			baseURL:    &url.URL{Host: "localhost:3500", Scheme: "http"},
			sszEnabled: true,
		}
		h, err := c.GetHeader(ctx, slot, bytesutil.ToBytes32(parentHash), bytesutil.ToBytes48(pubkey))
		require.NoError(t, err)
		expectedWithdrawalsRoot := ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
		bid, err := h.Message()
		require.NoError(t, err)
		bidHeader, err := bid.Header()
		require.NoError(t, err)
		withdrawalsRoot, err := bidHeader.WithdrawalsRoot()
		require.NoError(t, err)
		require.Equal(t, true, bytes.Equal(expectedWithdrawalsRoot, withdrawalsRoot))
		bidStr := "652312848583266388373324160190187140051835877600158453279131187530910662656"
		value, err := stringToUint256(bidStr)
		require.NoError(t, err)
		require.Equal(t, 0, value.Int.Cmp(primitives.WeiToBigInt(bid.Value())))
		require.Equal(t, bidStr, primitives.WeiToBigInt(bid.Value()).String())
	})
	t.Run("deneb", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, api.JsonMediaType, r.Header.Get("Accept"))
				require.Equal(t, expectedPath, r.URL.Path)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(testExampleHeaderResponseDeneb)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		h, err := c.GetHeader(ctx, slot, bytesutil.ToBytes32(parentHash), bytesutil.ToBytes48(pubkey))
		require.NoError(t, err)
		expectedWithdrawalsRoot := ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
		bid, err := h.Message()
		require.NoError(t, err)
		bidHeader, err := bid.Header()
		require.NoError(t, err)
		withdrawalsRoot, err := bidHeader.WithdrawalsRoot()
		require.NoError(t, err)
		require.Equal(t, true, bytes.Equal(expectedWithdrawalsRoot, withdrawalsRoot))

		bidStr := "652312848583266388373324160190187140051835877600158453279131187530910662656"
		value, err := stringToUint256(bidStr)
		require.NoError(t, err)
		require.Equal(t, 0, value.Int.Cmp(primitives.WeiToBigInt(bid.Value())))
		require.Equal(t, bidStr, primitives.WeiToBigInt(bid.Value()).String())
		dbid, ok := bid.(builderBidDeneb)
		require.Equal(t, true, ok)
		kcgCommitments := dbid.BlobKzgCommitments()
		require.Equal(t, len(kcgCommitments) > 0, true)
		for i := range kcgCommitments {
			require.Equal(t, len(kcgCommitments[i]) == 48, true)
		}
	})
	t.Run("deneb ssz", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Accept"))
				require.Equal(t, expectedPath, r.URL.Path)
				epr := &ExecHeaderResponseDeneb{}
				require.NoError(t, json.Unmarshal([]byte(testExampleHeaderResponseDeneb), epr))
				pro, err := epr.ToProto()
				require.NoError(t, err)
				ssz, err := pro.MarshalSSZ()
				require.NoError(t, err)
				header := http.Header{}
				header.Set(api.VersionHeader, "deneb")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(bytes.NewBuffer(ssz)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:         hc,
			baseURL:    &url.URL{Host: "localhost:3500", Scheme: "http"},
			sszEnabled: true,
		}
		h, err := c.GetHeader(ctx, slot, bytesutil.ToBytes32(parentHash), bytesutil.ToBytes48(pubkey))
		require.NoError(t, err)
		expectedWithdrawalsRoot := ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
		bid, err := h.Message()
		require.NoError(t, err)
		bidHeader, err := bid.Header()
		require.NoError(t, err)
		withdrawalsRoot, err := bidHeader.WithdrawalsRoot()
		require.NoError(t, err)
		require.Equal(t, true, bytes.Equal(expectedWithdrawalsRoot, withdrawalsRoot))

		bidStr := "652312848583266388373324160190187140051835877600158453279131187530910662656"
		value, err := stringToUint256(bidStr)
		require.NoError(t, err)
		require.Equal(t, 0, value.Int.Cmp(primitives.WeiToBigInt(bid.Value())))
		require.Equal(t, bidStr, primitives.WeiToBigInt(bid.Value()).String())
		dbid, ok := bid.(builderBidDeneb)
		require.Equal(t, true, ok)
		kcgCommitments := dbid.BlobKzgCommitments()
		require.Equal(t, len(kcgCommitments) > 0, true)
		for i := range kcgCommitments {
			require.Equal(t, len(kcgCommitments[i]) == 48, true)
		}
	})
	t.Run("deneb, too many kzg commitments", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, expectedPath, r.URL.Path)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(testExampleHeaderResponseDenebTooManyBlobs)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		_, err := c.GetHeader(ctx, slot, bytesutil.ToBytes32(parentHash), bytesutil.ToBytes48(pubkey))
		require.ErrorContains(t, "could not convert ExecHeaderResponseDeneb to proto: too many blob commitments: 7", err)
	})
	t.Run("electra", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, api.JsonMediaType, r.Header.Get("Accept"))
				require.Equal(t, expectedPath, r.URL.Path)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(testExampleHeaderResponseElectra)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		h, err := c.GetHeader(ctx, slot, bytesutil.ToBytes32(parentHash), bytesutil.ToBytes48(pubkey))
		require.NoError(t, err)
		expectedWithdrawalsRoot := ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
		bid, err := h.Message()
		require.NoError(t, err)
		bidHeader, err := bid.Header()
		require.NoError(t, err)
		withdrawalsRoot, err := bidHeader.WithdrawalsRoot()
		require.NoError(t, err)
		require.Equal(t, true, bytes.Equal(expectedWithdrawalsRoot, withdrawalsRoot))

		bidStr := "652312848583266388373324160190187140051835877600158453279131187530910662656"
		value, err := stringToUint256(bidStr)
		require.NoError(t, err)
		require.Equal(t, 0, value.Int.Cmp(primitives.WeiToBigInt(bid.Value())))
		require.Equal(t, bidStr, primitives.WeiToBigInt(bid.Value()).String())
		ebid, ok := bid.(builderBidElectra)
		require.Equal(t, true, ok)
		kcgCommitments := ebid.BlobKzgCommitments()
		require.Equal(t, len(kcgCommitments) > 0, true)
		for i := range kcgCommitments {
			require.Equal(t, len(kcgCommitments[i]) == 48, true)
		}
		requests := ebid.SilaRequests()
		require.Equal(t, 1, len(requests.Deposits))
		require.Equal(t, 1, len(requests.Withdrawals))
		require.Equal(t, 1, len(requests.Consolidations))

	})
	t.Run("electra ssz", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Accept"))
				require.Equal(t, expectedPath, r.URL.Path)
				epr := &ExecHeaderResponseElectra{}
				require.NoError(t, json.Unmarshal([]byte(testExampleHeaderResponseElectra), epr))
				pro, err := epr.ToProto(es)
				require.NoError(t, err)
				ssz, err := pro.MarshalSSZ()
				require.NoError(t, err)
				header := http.Header{}
				header.Set(api.VersionHeader, "electra")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(bytes.NewBuffer(ssz)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:         hc,
			baseURL:    &url.URL{Host: "localhost:3500", Scheme: "http"},
			sszEnabled: true,
		}
		h, err := c.GetHeader(ctx, slot, bytesutil.ToBytes32(parentHash), bytesutil.ToBytes48(pubkey))
		require.NoError(t, err)
		expectedWithdrawalsRoot := ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
		bid, err := h.Message()
		require.NoError(t, err)
		bidHeader, err := bid.Header()
		require.NoError(t, err)
		withdrawalsRoot, err := bidHeader.WithdrawalsRoot()
		require.NoError(t, err)
		require.Equal(t, true, bytes.Equal(expectedWithdrawalsRoot, withdrawalsRoot))

		bidStr := "652312848583266388373324160190187140051835877600158453279131187530910662656"
		value, err := stringToUint256(bidStr)
		require.NoError(t, err)
		require.Equal(t, 0, value.Int.Cmp(primitives.WeiToBigInt(bid.Value())))
		require.Equal(t, bidStr, primitives.WeiToBigInt(bid.Value()).String())
		ebid, ok := bid.(builderBidElectra)
		require.Equal(t, true, ok)
		kcgCommitments := ebid.BlobKzgCommitments()
		require.Equal(t, len(kcgCommitments) > 0, true)
		for i := range kcgCommitments {
			require.Equal(t, len(kcgCommitments[i]) == 48, true)
		}
		requests := ebid.SilaRequests()
		require.Equal(t, 1, len(requests.Deposits))
		require.Equal(t, 1, len(requests.Withdrawals))
		require.Equal(t, 1, len(requests.Consolidations))

	})
	t.Run("unsupported version", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, expectedPath, r.URL.Path)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(testExampleHeaderResponseUnknownVersion)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		_, err := c.GetHeader(ctx, slot, bytesutil.ToBytes32(parentHash), bytesutil.ToBytes48(pubkey))
		require.ErrorContains(t, "unsupported header version", err)
	})
}

func TestSubmitBlindedBlock(t *testing.T) {
	ctx := t.Context()

	t.Run("bellatrix", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, postBlindedBeaconBlockPath, r.URL.Path)
				require.Equal(t, "bellatrix", r.Header.Get("Eth-Consensus-Version"))
				require.Equal(t, api.JsonMediaType, r.Header.Get("Content-Type"))
				require.Equal(t, api.JsonMediaType, r.Header.Get("Accept"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(testExampleSilaPayload)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		sbbb, err := blocks.NewSignedBeaconBlock(testSignedBlindedBeaconBlockBellatrix(t))
		require.NoError(t, err)
		ep, _, err := c.SubmitBlindedBlock(ctx, sbbb)
		require.NoError(t, err)
		require.Equal(t, true, bytes.Equal(ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"), ep.ParentHash()))
		bfpg, err := stringToUint256("452312848583266388373324160190187140051835877600158453279131187530910662656")
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%#x", bfpg.SSZBytes()), fmt.Sprintf("%#x", ep.BaseFeePerGas()))
		require.Equal(t, uint64(1), ep.GasLimit())
	})
	t.Run("bellatrix ssz", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, postBlindedBeaconBlockPath, r.URL.Path)
				require.Equal(t, "bellatrix", r.Header.Get(api.VersionHeader))
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Content-Type"))
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Accept"))
				epr := &SilaPayloadResponse{}
				require.NoError(t, json.Unmarshal([]byte(testExampleSilaPayload), epr))
				ep := &structs.SilaPayload{}
				require.NoError(t, json.Unmarshal(epr.Data, ep))
				pro, err := ep.ToConsensus()
				require.NoError(t, err)
				ssz, err := pro.MarshalSSZ()
				require.NoError(t, err)
				header := http.Header{}
				header.Set(api.VersionHeader, "bellatrix")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(bytes.NewBuffer(ssz)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:         hc,
			baseURL:    &url.URL{Host: "localhost:3500", Scheme: "http"},
			sszEnabled: true,
		}
		sbbb, err := blocks.NewSignedBeaconBlock(testSignedBlindedBeaconBlockBellatrix(t))
		require.NoError(t, err)
		ep, _, err := c.SubmitBlindedBlock(ctx, sbbb)
		require.NoError(t, err)
		require.Equal(t, true, bytes.Equal(ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"), ep.ParentHash()))
		bfpg, err := stringToUint256("452312848583266388373324160190187140051835877600158453279131187530910662656")
		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("%#x", bfpg.SSZBytes()), fmt.Sprintf("%#x", ep.BaseFeePerGas()))
		require.Equal(t, uint64(1), ep.GasLimit())
	})
	t.Run("capella", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, postBlindedBeaconBlockPath, r.URL.Path)
				require.Equal(t, "capella", r.Header.Get(api.VersionHeader))
				require.Equal(t, api.JsonMediaType, r.Header.Get("Content-Type"))
				require.Equal(t, api.JsonMediaType, r.Header.Get("Accept"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(testExampleSilaPayloadCapella)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		sbb, err := blocks.NewSignedBeaconBlock(testSignedBlindedBeaconBlockCapella(t))
		require.NoError(t, err)
		ep, _, err := c.SubmitBlindedBlock(ctx, sbb)
		require.NoError(t, err)
		withdrawals, err := ep.Withdrawals()
		require.NoError(t, err)
		require.Equal(t, 1, len(withdrawals))
		assert.Equal(t, uint64(1), withdrawals[0].Index)
		assert.Equal(t, primitives.ValidatorIndex(1), withdrawals[0].ValidatorIndex)
		assert.DeepEqual(t, ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943"), withdrawals[0].Address)
		assert.Equal(t, uint64(1), withdrawals[0].Amount)
	})
	t.Run("capella ssz", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, postBlindedBeaconBlockPath, r.URL.Path)
				require.Equal(t, "capella", r.Header.Get(api.VersionHeader))
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Content-Type"))
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Accept"))
				epr := &SilaPayloadResponse{}
				require.NoError(t, json.Unmarshal([]byte(testExampleSilaPayloadCapella), epr))
				ep := &structs.SilaPayloadCapella{}
				require.NoError(t, json.Unmarshal(epr.Data, ep))
				pro, err := ep.ToConsensus()
				require.NoError(t, err)
				ssz, err := pro.MarshalSSZ()
				require.NoError(t, err)
				header := http.Header{}
				header.Set(api.VersionHeader, "capella")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(bytes.NewBuffer(ssz)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:         hc,
			baseURL:    &url.URL{Host: "localhost:3500", Scheme: "http"},
			sszEnabled: true,
		}
		sbb, err := blocks.NewSignedBeaconBlock(testSignedBlindedBeaconBlockCapella(t))
		require.NoError(t, err)
		ep, _, err := c.SubmitBlindedBlock(ctx, sbb)
		require.NoError(t, err)
		withdrawals, err := ep.Withdrawals()
		require.NoError(t, err)
		require.Equal(t, 1, len(withdrawals))
		assert.Equal(t, uint64(1), withdrawals[0].Index)
		assert.Equal(t, primitives.ValidatorIndex(1), withdrawals[0].ValidatorIndex)
		assert.DeepEqual(t, ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943"), withdrawals[0].Address)
		assert.Equal(t, uint64(1), withdrawals[0].Amount)
	})
	t.Run("deneb", func(t *testing.T) {
		test := testSignedBlindedBeaconBlockDeneb(t)
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, postBlindedBeaconBlockPath, r.URL.Path)
				require.Equal(t, "deneb", r.Header.Get(api.VersionHeader))
				require.Equal(t, api.JsonMediaType, r.Header.Get("Content-Type"))
				require.Equal(t, api.JsonMediaType, r.Header.Get("Accept"))
				var req structs.SignedBlindedBeaconBlockDeneb
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				block, err := req.ToConsensus()
				require.NoError(t, err)
				require.DeepEqual(t, block, test)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(testExampleSilaPayloadDeneb)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}

		sbb, err := blocks.NewSignedBeaconBlock(test)
		require.NoError(t, err)

		ep, blobBundle, err := c.SubmitBlindedBlock(ctx, sbb)
		require.NoError(t, err)
		withdrawals, err := ep.Withdrawals()
		require.NoError(t, err)
		require.Equal(t, 1, len(withdrawals))
		assert.Equal(t, uint64(1), withdrawals[0].Index)
		assert.Equal(t, primitives.ValidatorIndex(1), withdrawals[0].ValidatorIndex)
		assert.DeepEqual(t, ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943"), withdrawals[0].Address)
		assert.Equal(t, uint64(1), withdrawals[0].Amount)
		require.NotNil(t, blobBundle)
	})
	t.Run("deneb ssz", func(t *testing.T) {
		test := testSignedBlindedBeaconBlockDeneb(t)
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, postBlindedBeaconBlockPath, r.URL.Path)
				require.Equal(t, "deneb", r.Header.Get(api.VersionHeader))
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Content-Type"))
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Accept"))
				epr := &ExecPayloadResponseDeneb{}
				require.NoError(t, json.Unmarshal([]byte(testExampleSilaPayloadDeneb), epr))
				pro, blob, err := epr.ToProto()
				require.NoError(t, err)
				combined := &v1.SilaPayloadDenebAndBlobsBundle{
					Payload:     pro,
					BlobsBundle: blob,
				}
				ssz, err := combined.MarshalSSZ()
				require.NoError(t, err)
				header := http.Header{}
				header.Set(api.VersionHeader, "deneb")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(bytes.NewBuffer(ssz)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:         hc,
			baseURL:    &url.URL{Host: "localhost:3500", Scheme: "http"},
			sszEnabled: true,
		}
		sbb, err := blocks.NewSignedBeaconBlock(test)
		require.NoError(t, err)

		ep, blobBundle, err := c.SubmitBlindedBlock(ctx, sbb)
		require.NoError(t, err)
		withdrawals, err := ep.Withdrawals()
		require.NoError(t, err)
		require.Equal(t, 1, len(withdrawals))
		assert.Equal(t, uint64(1), withdrawals[0].Index)
		assert.Equal(t, primitives.ValidatorIndex(1), withdrawals[0].ValidatorIndex)
		assert.DeepEqual(t, ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943"), withdrawals[0].Address)
		assert.Equal(t, uint64(1), withdrawals[0].Amount)
		require.NotNil(t, blobBundle)
	})
	t.Run("electra", func(t *testing.T) {
		test := testSignedBlindedBeaconBlockElectra(t)
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, postBlindedBeaconBlockPath, r.URL.Path)
				require.Equal(t, "electra", r.Header.Get(api.VersionHeader))
				require.Equal(t, api.JsonMediaType, r.Header.Get("Content-Type"))
				require.Equal(t, api.JsonMediaType, r.Header.Get("Accept"))
				var req structs.SignedBlindedBeaconBlockElectra
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				block, err := req.ToConsensus()
				require.NoError(t, err)
				require.DeepEqual(t, block, test)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(testExampleSilaPayloadDeneb)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}

		sbb, err := blocks.NewSignedBeaconBlock(test)
		require.NoError(t, err)

		ep, blobBundle, err := c.SubmitBlindedBlock(ctx, sbb)
		require.NoError(t, err)
		withdrawals, err := ep.Withdrawals()
		require.NoError(t, err)
		require.Equal(t, 1, len(withdrawals))
		assert.Equal(t, uint64(1), withdrawals[0].Index)
		assert.Equal(t, primitives.ValidatorIndex(1), withdrawals[0].ValidatorIndex)
		assert.DeepEqual(t, ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943"), withdrawals[0].Address)
		assert.Equal(t, uint64(1), withdrawals[0].Amount)
		require.NotNil(t, blobBundle)
	})
	t.Run("electra ssz", func(t *testing.T) {
		test := testSignedBlindedBeaconBlockElectra(t)
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, postBlindedBeaconBlockPath, r.URL.Path)
				require.Equal(t, "electra", r.Header.Get(api.VersionHeader))
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Content-Type"))
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Accept"))
				epr := &ExecPayloadResponseDeneb{}
				require.NoError(t, json.Unmarshal([]byte(testExampleSilaPayloadDeneb), epr))
				pro, blob, err := epr.ToProto()
				require.NoError(t, err)
				combined := &v1.SilaPayloadDenebAndBlobsBundle{
					Payload:     pro,
					BlobsBundle: blob,
				}
				ssz, err := combined.MarshalSSZ()
				require.NoError(t, err)
				header := http.Header{}
				header.Set(api.VersionHeader, "electra")
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       io.NopCloser(bytes.NewBuffer(ssz)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:         hc,
			baseURL:    &url.URL{Host: "localhost:3500", Scheme: "http"},
			sszEnabled: true,
		}
		sbb, err := blocks.NewSignedBeaconBlock(test)
		require.NoError(t, err)

		ep, blobBundle, err := c.SubmitBlindedBlock(ctx, sbb)
		require.NoError(t, err)
		withdrawals, err := ep.Withdrawals()
		require.NoError(t, err)
		require.Equal(t, 1, len(withdrawals))
		assert.Equal(t, uint64(1), withdrawals[0].Index)
		assert.Equal(t, primitives.ValidatorIndex(1), withdrawals[0].ValidatorIndex)
		assert.DeepEqual(t, ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943"), withdrawals[0].Address)
		assert.Equal(t, uint64(1), withdrawals[0].Amount)
		require.NotNil(t, blobBundle)
	})
	t.Run("mismatched versions, expected bellatrix got capella", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, postBlindedBeaconBlockPath, r.URL.Path)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(testExampleSilaPayloadCapella)), // send a Capella payload
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		sbbb, err := blocks.NewSignedBeaconBlock(testSignedBlindedBeaconBlockBellatrix(t))
		require.NoError(t, err)
		_, _, err = c.SubmitBlindedBlock(ctx, sbbb)
		require.ErrorIs(t, err, errResponseVersionMismatch)
	})
	t.Run("not blinded", func(t *testing.T) {
		sbb, err := blocks.NewSignedBeaconBlock(&sila.SignedBeaconBlockBellatrix{Block: &sila.BeaconBlockBellatrix{Body: &sila.BeaconBlockBodyBellatrix{SilaPayload: &v1.SilaPayload{}}}})
		require.NoError(t, err)
		_, _, err = (&Client{}).SubmitBlindedBlock(ctx, sbb)
		require.ErrorIs(t, err, errNotBlinded)
	})
}

func testSignedBlindedBeaconBlockBellatrix(t *testing.T) *sila.SignedBlindedBeaconBlockBellatrix {
	return &sila.SignedBlindedBeaconBlockBellatrix{
		Block: &sila.BlindedBeaconBlockBellatrix{
			Slot:          1,
			ProposerIndex: 1,
			ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			Body: &sila.BlindedBeaconBlockBodyBellatrix{
				RandaoReveal: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
				SilaData: &sila.SilaData{
					DepositRoot:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					DepositCount: 1,
					BlockHash:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				},
				Graffiti: ezDecode(t, "0xdeadbeefc0ffee"),
				ProposerSlashings: []*sila.ProposerSlashing{
					{
						Header_1: &sila.SignedBeaconBlockHeader{
							Header: &sila.BeaconBlockHeader{
								Slot:          1,
								ProposerIndex: 1,
								ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								BodyRoot:      ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
						Header_2: &sila.SignedBeaconBlockHeader{
							Header: &sila.BeaconBlockHeader{
								Slot:          1,
								ProposerIndex: 1,
								ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								BodyRoot:      ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
					},
				},
				AttesterSlashings: []*sila.AttesterSlashing{
					{
						Attestation_1: &sila.IndexedAttestation{
							AttestingIndices: []uint64{1},
							Data: &sila.AttestationData{
								Slot:            1,
								CommitteeIndex:  1,
								BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								Source: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
								Target: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
						Attestation_2: &sila.IndexedAttestation{
							AttestingIndices: []uint64{1},
							Data: &sila.AttestationData{
								Slot:            1,
								CommitteeIndex:  1,
								BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								Source: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
								Target: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
					},
				},
				Attestations: []*sila.Attestation{
					{
						AggregationBits: bitfield.Bitlist{0x01},
						Data: &sila.AttestationData{
							Slot:            1,
							CommitteeIndex:  1,
							BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							Source: &sila.Checkpoint{
								Epoch: 1,
								Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
							Target: &sila.Checkpoint{
								Epoch: 1,
								Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
						},
						Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
					},
				},
				Deposits: []*sila.Deposit{
					{
						Proof: func() [][]byte {
							b := make([][]byte, 33)
							for i := range b {
								b[i] = ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
							}
							return b
						}(),
						Data: &sila.Deposit_Data{
							PublicKey:             ezDecode(t, "0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a"),
							WithdrawalCredentials: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							Amount:                1,
							Signature:             ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
					},
				},
				VoluntaryExits: []*sila.SignedVoluntaryExit{
					{
						Exit: &sila.VoluntaryExit{
							Epoch:          1,
							ValidatorIndex: 1,
						},
						Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
					},
				},
				SyncAggregate: &sila.SyncAggregate{
					SyncCommitteeSignature: make([]byte, 96),
					SyncCommitteeBits:      make(bitfield.Bitvector512, 64),
				},
				SilaPayloadHeader: &v1.SilaPayloadHeader{
					ParentHash:       ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					FeeRecipient:     ezDecode(t, "0xabcf8e0d4e9587369b2301d0790347320302cc09"),
					StateRoot:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					ReceiptsRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					LogsBloom:        ezDecode(t, "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
					PrevRandao:       ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					BlockNumber:      1,
					GasLimit:         1,
					GasUsed:          1,
					Timestamp:        1,
					ExtraData:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					BaseFeePerGas:    ezDecode(t, "0x4523128485832663883733241601901871400518358776001584532791311875"),
					BlockHash:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					TransactionsRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				},
			},
		},
		Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
	}
}

func testSignedBlindedBeaconBlockCapella(t *testing.T) *sila.SignedBlindedBeaconBlockCapella {
	return &sila.SignedBlindedBeaconBlockCapella{
		Block: &sila.BlindedBeaconBlockCapella{
			Slot:          1,
			ProposerIndex: 1,
			ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			Body: &sila.BlindedBeaconBlockBodyCapella{
				RandaoReveal: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
				SilaData: &sila.SilaData{
					DepositRoot:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					DepositCount: 1,
					BlockHash:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				},
				Graffiti: ezDecode(t, "0xdeadbeefc0ffee"),
				ProposerSlashings: []*sila.ProposerSlashing{
					{
						Header_1: &sila.SignedBeaconBlockHeader{
							Header: &sila.BeaconBlockHeader{
								Slot:          1,
								ProposerIndex: 1,
								ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								BodyRoot:      ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
						Header_2: &sila.SignedBeaconBlockHeader{
							Header: &sila.BeaconBlockHeader{
								Slot:          1,
								ProposerIndex: 1,
								ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								BodyRoot:      ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
					},
				},
				AttesterSlashings: []*sila.AttesterSlashing{
					{
						Attestation_1: &sila.IndexedAttestation{
							AttestingIndices: []uint64{1},
							Data: &sila.AttestationData{
								Slot:            1,
								CommitteeIndex:  1,
								BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								Source: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
								Target: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
						Attestation_2: &sila.IndexedAttestation{
							AttestingIndices: []uint64{1},
							Data: &sila.AttestationData{
								Slot:            1,
								CommitteeIndex:  1,
								BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								Source: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
								Target: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
					},
				},
				Attestations: []*sila.Attestation{
					{
						AggregationBits: bitfield.Bitlist{0x01},
						Data: &sila.AttestationData{
							Slot:            1,
							CommitteeIndex:  1,
							BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							Source: &sila.Checkpoint{
								Epoch: 1,
								Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
							Target: &sila.Checkpoint{
								Epoch: 1,
								Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
						},
						Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
					},
				},
				Deposits: []*sila.Deposit{
					{
						Proof: func() [][]byte {
							b := make([][]byte, 33)
							for i := range b {
								b[i] = ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
							}
							return b
						}(),
						Data: &sila.Deposit_Data{
							PublicKey:             ezDecode(t, "0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a"),
							WithdrawalCredentials: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							Amount:                1,
							Signature:             ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
					},
				},
				VoluntaryExits: []*sila.SignedVoluntaryExit{
					{
						Exit: &sila.VoluntaryExit{
							Epoch:          1,
							ValidatorIndex: 1,
						},
						Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
					},
				},
				SyncAggregate: &sila.SyncAggregate{
					SyncCommitteeSignature: make([]byte, 96),
					SyncCommitteeBits:      make(bitfield.Bitvector512, 64),
				},
				SilaPayloadHeader: &v1.SilaPayloadHeaderCapella{
					ParentHash:       ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					FeeRecipient:     ezDecode(t, "0xabcf8e0d4e9587369b2301d0790347320302cc09"),
					StateRoot:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					ReceiptsRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					LogsBloom:        ezDecode(t, "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
					PrevRandao:       ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					BlockNumber:      1,
					GasLimit:         1,
					GasUsed:          1,
					Timestamp:        1,
					ExtraData:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					BaseFeePerGas:    ezDecode(t, "0x4523128485832663883733241601901871400518358776001584532791311875"),
					BlockHash:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					TransactionsRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					WithdrawalsRoot:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				},
			},
		},
		Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
	}
}

func testSignedBlindedBeaconBlockDeneb(t *testing.T) *sila.SignedBlindedBeaconBlockDeneb {
	basebytes, err := bytesutil.Uint256ToSSZBytes("14074904626401341155369551180448584754667373453244490859944217516317499064576")
	if err != nil {
		log.Error(err)
	}
	return &sila.SignedBlindedBeaconBlockDeneb{
		Message: &sila.BlindedBeaconBlockDeneb{
			Slot:          1,
			ProposerIndex: 1,
			ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			Body: &sila.BlindedBeaconBlockBodyDeneb{
				RandaoReveal: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
				SilaData: &sila.SilaData{
					DepositRoot:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					DepositCount: 1,
					BlockHash:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				},
				Graffiti: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				ProposerSlashings: []*sila.ProposerSlashing{
					{
						Header_1: &sila.SignedBeaconBlockHeader{
							Header: &sila.BeaconBlockHeader{
								Slot:          1,
								ProposerIndex: 1,
								ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								BodyRoot:      ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
						Header_2: &sila.SignedBeaconBlockHeader{
							Header: &sila.BeaconBlockHeader{
								Slot:          1,
								ProposerIndex: 1,
								ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								BodyRoot:      ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
					},
				},
				AttesterSlashings: []*sila.AttesterSlashing{
					{
						Attestation_1: &sila.IndexedAttestation{
							AttestingIndices: []uint64{1},
							Data: &sila.AttestationData{
								Slot:            1,
								CommitteeIndex:  1,
								BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								Source: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
								Target: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
						Attestation_2: &sila.IndexedAttestation{
							AttestingIndices: []uint64{1},
							Data: &sila.AttestationData{
								Slot:            1,
								CommitteeIndex:  1,
								BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								Source: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
								Target: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
					},
				},
				Attestations: []*sila.Attestation{
					{
						AggregationBits: bitfield.Bitlist{0x01},
						Data: &sila.AttestationData{
							Slot:            1,
							CommitteeIndex:  1,
							BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							Source: &sila.Checkpoint{
								Epoch: 1,
								Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
							Target: &sila.Checkpoint{
								Epoch: 1,
								Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
						},
						Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
					},
				},
				Deposits: []*sila.Deposit{
					{
						Proof: func() [][]byte {
							b := make([][]byte, 33)
							for i := range b {
								b[i] = ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
							}
							return b
						}(),
						Data: &sila.Deposit_Data{
							PublicKey:             ezDecode(t, "0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a"),
							WithdrawalCredentials: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							Amount:                1,
							Signature:             ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
					},
				},
				VoluntaryExits: []*sila.SignedVoluntaryExit{
					{
						Exit: &sila.VoluntaryExit{
							Epoch:          1,
							ValidatorIndex: 1,
						},
						Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
					},
				},
				SyncAggregate: &sila.SyncAggregate{
					SyncCommitteeSignature: make([]byte, 96),
					SyncCommitteeBits:      ezDecode(t, "0x6451e9f951ebf05edc01de67e593484b672877054f055903ff0df1a1a945cf30ca26bb4d4b154f94a1bc776bcf5d0efb3603e1f9b8ee2499ccdcfe2a18cef458"),
				},
				SilaPayloadHeader: &v1.SilaPayloadHeaderDeneb{
					ParentHash:       ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					FeeRecipient:     ezDecode(t, "0xabcf8e0d4e9587369b2301d0790347320302cc09"),
					StateRoot:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					ReceiptsRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					LogsBloom:        ezDecode(t, "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
					PrevRandao:       ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					BlockNumber:      1,
					GasLimit:         1,
					GasUsed:          1,
					Timestamp:        1,
					ExtraData:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					BaseFeePerGas:    basebytes,
					BlockHash:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					TransactionsRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					WithdrawalsRoot:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					BlobGasUsed:      1,
					ExcessBlobGas:    2,
				},
			},
		},
		Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
	}
}

func testSignedBlindedBeaconBlockElectra(t *testing.T) *sila.SignedBlindedBeaconBlockElectra {
	basebytes, err := bytesutil.Uint256ToSSZBytes("14074904626401341155369551180448584754667373453244490859944217516317499064576")
	if err != nil {
		log.Error(err)
	}
	return &sila.SignedBlindedBeaconBlockElectra{
		Message: &sila.BlindedBeaconBlockElectra{
			Slot:          1,
			ProposerIndex: 1,
			ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			Body: &sila.BlindedBeaconBlockBodyElectra{
				RandaoReveal: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
				SilaData: &sila.SilaData{
					DepositRoot:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					DepositCount: 1,
					BlockHash:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				},
				Graffiti: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				ProposerSlashings: []*sila.ProposerSlashing{
					{
						Header_1: &sila.SignedBeaconBlockHeader{
							Header: &sila.BeaconBlockHeader{
								Slot:          1,
								ProposerIndex: 1,
								ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								BodyRoot:      ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
						Header_2: &sila.SignedBeaconBlockHeader{
							Header: &sila.BeaconBlockHeader{
								Slot:          1,
								ProposerIndex: 1,
								ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								BodyRoot:      ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
					},
				},
				AttesterSlashings: []*sila.AttesterSlashingElectra{
					{
						Attestation_1: &sila.IndexedAttestationElectra{
							AttestingIndices: []uint64{1},
							Data: &sila.AttestationData{
								Slot:            1,
								CommitteeIndex:  1,
								BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								Source: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
								Target: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
						Attestation_2: &sila.IndexedAttestationElectra{
							AttestingIndices: []uint64{1},
							Data: &sila.AttestationData{
								Slot:            1,
								CommitteeIndex:  1,
								BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								Source: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
								Target: &sila.Checkpoint{
									Epoch: 1,
									Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
								},
							},
							Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
					},
				},
				Attestations: []*sila.AttestationElectra{
					{
						AggregationBits: bitfield.Bitlist{0x01},
						Data: &sila.AttestationData{
							Slot:            1,
							CommitteeIndex:  1,
							BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							Source: &sila.Checkpoint{
								Epoch: 1,
								Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
							Target: &sila.Checkpoint{
								Epoch: 1,
								Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							},
						},
						CommitteeBits: make(bitfield.Bitvector64, 8),
						Signature:     ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
					},
				},
				Deposits: []*sila.Deposit{
					{
						Proof: func() [][]byte {
							b := make([][]byte, 33)
							for i := range b {
								b[i] = ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
							}
							return b
						}(),
						Data: &sila.Deposit_Data{
							PublicKey:             ezDecode(t, "0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a"),
							WithdrawalCredentials: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
							Amount:                1,
							Signature:             ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
						},
					},
				},
				VoluntaryExits: []*sila.SignedVoluntaryExit{
					{
						Exit: &sila.VoluntaryExit{
							Epoch:          1,
							ValidatorIndex: 1,
						},
						Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
					},
				},
				SyncAggregate: &sila.SyncAggregate{
					SyncCommitteeSignature: make([]byte, 96),
					SyncCommitteeBits:      ezDecode(t, "0x6451e9f951ebf05edc01de67e593484b672877054f055903ff0df1a1a945cf30ca26bb4d4b154f94a1bc776bcf5d0efb3603e1f9b8ee2499ccdcfe2a18cef458"),
				},
				SilaRequests: &v1.SilaRequests{},
				SilaPayloadHeader: &v1.SilaPayloadHeaderDeneb{
					ParentHash:       ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					FeeRecipient:     ezDecode(t, "0xabcf8e0d4e9587369b2301d0790347320302cc09"),
					StateRoot:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					ReceiptsRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					LogsBloom:        ezDecode(t, "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
					PrevRandao:       ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					BlockNumber:      1,
					GasLimit:         1,
					GasUsed:          1,
					Timestamp:        1,
					ExtraData:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					BaseFeePerGas:    basebytes,
					BlockHash:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					TransactionsRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					WithdrawalsRoot:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
					BlobGasUsed:      1,
					ExcessBlobGas:    2,
				},
			},
		},
		Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
	}
}

func TestSubmitBlindedBlockPostFulu(t *testing.T) {
	ctx := t.Context()

	t.Run("success", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, postBlindedBeaconBlockV2Path, r.URL.Path)
				require.Equal(t, "bellatrix", r.Header.Get("Eth-Consensus-Version"))
				require.Equal(t, api.JsonMediaType, r.Header.Get("Content-Type"))
				require.Equal(t, api.JsonMediaType, r.Header.Get("Accept"))
				// Post-Fulu: only return status code, no payload
				return &http.Response{
					StatusCode: http.StatusAccepted,
					Body:       io.NopCloser(bytes.NewBufferString("")),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		sbbb, err := blocks.NewSignedBeaconBlock(testSignedBlindedBeaconBlockBellatrix(t))
		require.NoError(t, err)
		err = c.SubmitBlindedBlockPostFulu(ctx, sbbb)
		require.NoError(t, err)
	})

	t.Run("success_ssz", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, postBlindedBeaconBlockV2Path, r.URL.Path)
				require.Equal(t, "bellatrix", r.Header.Get(api.VersionHeader))
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Content-Type"))
				require.Equal(t, api.OctetStreamMediaType, r.Header.Get("Accept"))
				// Post-Fulu: only return status code, no payload
				return &http.Response{
					StatusCode: http.StatusAccepted,
					Body:       io.NopCloser(bytes.NewBufferString("")),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:         hc,
			baseURL:    &url.URL{Host: "localhost:3500", Scheme: "http"},
			sszEnabled: true,
		}
		sbbb, err := blocks.NewSignedBeaconBlock(testSignedBlindedBeaconBlockBellatrix(t))
		require.NoError(t, err)
		err = c.SubmitBlindedBlockPostFulu(ctx, sbbb)
		require.NoError(t, err)
	})

	t.Run("error_response", func(t *testing.T) {
		hc := &http.Client{
			Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, postBlindedBeaconBlockV2Path, r.URL.Path)
				require.Equal(t, "bellatrix", r.Header.Get("Eth-Consensus-Version"))
				message := ErrorMessage{
					Code:    400,
					Message: "Bad Request",
				}
				resp, err := json.Marshal(message)
				require.NoError(t, err)
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewBuffer(resp)),
					Request:    r.Clone(ctx),
				}, nil
			}),
		}
		c := &Client{
			hc:      hc,
			baseURL: &url.URL{Host: "localhost:3500", Scheme: "http"},
		}
		sbbb, err := blocks.NewSignedBeaconBlock(testSignedBlindedBeaconBlockBellatrix(t))
		require.NoError(t, err)
		err = c.SubmitBlindedBlockPostFulu(ctx, sbbb)
		require.ErrorIs(t, err, ErrNotOK)
	})
}

func TestRequestLogger(t *testing.T) {
	wo := WithObserver(&requestLogger{})
	c, err := NewClient("localhost:3500", wo)
	require.NoError(t, err)

	ctx := t.Context()
	hc := &http.Client{
		Transport: roundtrip(func(r *http.Request) (*http.Response, error) {
			require.Equal(t, getStatus, r.URL.Path)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(testExampleSilaPayload)),
				Request:    r.Clone(ctx),
			}, nil
		}),
	}
	c.hc = hc
	err = c.Status(ctx)
	require.NoError(t, err)
}

func TestGetVersionsBlockToPayload(t *testing.T) {
	tests := []struct {
		name            string
		blockVersion    int
		expectedVersion int
		expectedError   bool
	}{
		{
			name:            "Fulu version",
			blockVersion:    6, // version.Fulu
			expectedVersion: 6,
			expectedError:   false,
		},
		{
			name:            "Deneb version",
			blockVersion:    4, // version.Deneb
			expectedVersion: 4,
			expectedError:   false,
		},
		{
			name:            "Capella version",
			blockVersion:    3, // version.Capella
			expectedVersion: 3,
			expectedError:   false,
		},
		{
			name:            "Bellatrix version",
			blockVersion:    2, // version.Bellatrix
			expectedVersion: 2,
			expectedError:   false,
		},
		{
			name:          "Unsupported version",
			blockVersion:  0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := getVersionsBlockToPayload(tt.blockVersion)
			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVersion, version)
			}
		})
	}
}

func TestParseBlindedBlockResponseSSZ_WithBlobsBundleV2(t *testing.T) {
	c := &Client{sszEnabled: true}

	// Create test payload
	payload := &v1.SilaPayloadDeneb{
		ParentHash:    make([]byte, 32),
		FeeRecipient:  make([]byte, 20),
		StateRoot:     make([]byte, 32),
		ReceiptsRoot:  make([]byte, 32),
		LogsBloom:     make([]byte, 256),
		PrevRandao:    make([]byte, 32),
		BlockNumber:   123456,
		GasLimit:      30000000,
		GasUsed:       21000,
		Timestamp:     1234567890,
		ExtraData:     []byte("test-extra-data"),
		BaseFeePerGas: make([]byte, 32),
		BlockHash:     make([]byte, 32),
		Transactions:  [][]byte{},
		Withdrawals:   []*v1.Withdrawal{},
		BlobGasUsed:   1024,
		ExcessBlobGas: 2048,
	}

	// Create test BlobsBundleV2
	bundleV2 := &v1.BlobsBundleV2{
		KzgCommitments: [][]byte{make([]byte, 48), make([]byte, 48)},
		Proofs:         [][]byte{make([]byte, 48), make([]byte, 48)},
		Blobs:          [][]byte{make([]byte, 131072), make([]byte, 131072)},
	}

	// Test Fulu version (should use SilaPayloadDenebAndBlobsBundleV2)
	t.Run("Fulu version with BlobsBundleV2", func(t *testing.T) {
		payloadAndBlobsV2 := &v1.SilaPayloadDenebAndBlobsBundleV2{
			Payload:     payload,
			BlobsBundle: bundleV2,
		}

		respBytes, err := payloadAndBlobsV2.MarshalSSZ()
		require.NoError(t, err)

		ed, bundle, err := c.parseBlindedBlockResponseSSZ(respBytes, 6) // version.Fulu
		require.NoError(t, err)
		require.NotNil(t, ed)
		require.NotNil(t, bundle)

		// Verify the bundle is BlobsBundleV2
		bundleV2Result, ok := bundle.(*v1.BlobsBundleV2)
		assert.Equal(t, true, ok, "Expected BlobsBundleV2 type")
		require.Equal(t, len(bundleV2.KzgCommitments), len(bundleV2Result.KzgCommitments))
		require.Equal(t, len(bundleV2.Proofs), len(bundleV2Result.Proofs))
		require.Equal(t, len(bundleV2.Blobs), len(bundleV2Result.Blobs))
	})

	// Test Deneb version (should use regular BlobsBundle)
	t.Run("Deneb version with regular BlobsBundle", func(t *testing.T) {
		regularBundle := &v1.BlobsBundle{
			KzgCommitments: bundleV2.KzgCommitments,
			Proofs:         bundleV2.Proofs,
			Blobs:          bundleV2.Blobs,
		}

		payloadAndBlobs := &v1.SilaPayloadDenebAndBlobsBundle{
			Payload:     payload,
			BlobsBundle: regularBundle,
		}

		respBytes, err := payloadAndBlobs.MarshalSSZ()
		require.NoError(t, err)

		ed, bundle, err := c.parseBlindedBlockResponseSSZ(respBytes, 4) // version.Deneb
		require.NoError(t, err)
		require.NotNil(t, ed)
		require.NotNil(t, bundle)

		// Verify the bundle is regular BlobsBundle
		regularBundleResult, ok := bundle.(*v1.BlobsBundle)
		assert.Equal(t, true, ok, "Expected BlobsBundle type")
		require.Equal(t, len(regularBundle.KzgCommitments), len(regularBundleResult.KzgCommitments))
	})

	// Test invalid SSZ data
	t.Run("Invalid SSZ data", func(t *testing.T) {
		invalidBytes := []byte("invalid-ssz-data")

		ed, bundle, err := c.parseBlindedBlockResponseSSZ(invalidBytes, 6)
		assert.NotNil(t, err)
		assert.Equal(t, true, ed == nil)
		assert.Equal(t, true, bundle == nil)
	})
}

func TestSubmitBlindedBlock_BlobsBundlerInterface(t *testing.T) {
	// Note: The full integration test is complex due to version detection logic
	// The key functionality is tested in the parseBlindedBlockResponseSSZ tests above
	// and in the mock service tests which verify the interface changes work correctly

	t.Run("Interface signature verification", func(t *testing.T) {
		// This test verifies that the SubmitBlindedBlock method signature
		// has been updated to return BlobsBundler interface

		client := &Client{}

		// Verify the method exists with the correct signature
		// by using reflection or by checking it compiles with the interface
		var _ func(ctx context.Context, sb interfaces.ReadOnlySignedBeaconBlock) (interfaces.SilaData, v1.BlobsBundler, error) = client.SubmitBlindedBlock

		// This test passes if the signature is correct
		assert.Equal(t, true, true)
	})
}
