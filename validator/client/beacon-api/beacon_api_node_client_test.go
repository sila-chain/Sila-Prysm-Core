package beacon_api

import (
	"errors"
	"net/http"
	"testing"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/testing/assert"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/beacon-api/mock"
	"github.com/sila-chain/Sila/common/hexutil"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetGenesis(t *testing.T) {
	testCases := []struct {
		name                    string
		genesisResponse         *structs.Genesis
		genesisError            error
		silaDepositResponse structs.GetSilaDepositResponse
		silaDepositError    error
		queriesSilaDeposit  bool
		expectedResponse        *silapb.Genesis
		expectedError           string
	}{
		{
			name:          "fails to get genesis",
			genesisError:  errors.New("foo error"),
			expectedError: "failed to get genesis: foo error",
		},
		{
			name: "fails to decode genesis validator root",
			genesisResponse: &structs.Genesis{
				GenesisTime:           "1",
				GenesisValidatorsRoot: "foo",
			},
			expectedError: "failed to decode genesis validator root `foo`",
		},
		{
			name: "fails to parse genesis time",
			genesisResponse: &structs.Genesis{
				GenesisTime:           "foo",
				GenesisValidatorsRoot: hexutil.Encode([]byte{1}),
			},
			expectedError: "failed to parse genesis time `foo`",
		},
		{
			name: "fails to query contract information",
			genesisResponse: &structs.Genesis{
				GenesisTime:           "1",
				GenesisValidatorsRoot: hexutil.Encode([]byte{2}),
			},
			silaDepositError:   errors.New("foo error"),
			queriesSilaDeposit: true,
			expectedError:          "foo error",
		},
		{
			name: "fails to read nil sila deposit data",
			genesisResponse: &structs.Genesis{
				GenesisTime:           "1",
				GenesisValidatorsRoot: hexutil.Encode([]byte{2}),
			},
			queriesSilaDeposit: true,
			silaDepositResponse: structs.GetSilaDepositResponse{
				Data: nil,
			},
			expectedError: "sila deposit data is nil",
		},
		{
			name: "fails to decode sila deposit address",
			genesisResponse: &structs.Genesis{
				GenesisTime:           "1",
				GenesisValidatorsRoot: hexutil.Encode([]byte{2}),
			},
			queriesSilaDeposit: true,
			silaDepositResponse: structs.GetSilaDepositResponse{
				Data: &structs.SilaDepositData{
					Address: "foo",
				},
			},
			expectedError: "failed to decode sila deposit address `foo`",
		},
		{
			name: "successfully retrieves genesis info",
			genesisResponse: &structs.Genesis{
				GenesisTime:           "654812",
				GenesisValidatorsRoot: hexutil.Encode([]byte{2}),
			},
			queriesSilaDeposit: true,
			silaDepositResponse: structs.GetSilaDepositResponse{
				Data: &structs.SilaDepositData{
					Address: hexutil.Encode([]byte{3}),
				},
			},
			expectedResponse: &silapb.Genesis{
				GenesisTime: &timestamppb.Timestamp{
					Seconds: 654812,
				},
				SilaDepositAddress: []byte{3},
				GenesisValidatorsRoot:  []byte{2},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := t.Context()

			genesisProvider := mock.NewMockGenesisProvider(ctrl)
			genesisProvider.EXPECT().Genesis(
				gomock.Any(),
			).Return(
				testCase.genesisResponse,
				testCase.genesisError,
			)

			silaDepositJson := structs.GetSilaDepositResponse{}
			handler := mock.NewMockJsonRestHandler(ctrl)

			if testCase.queriesSilaDeposit {
				handler.EXPECT().Get(
					gomock.Any(),
					"/sila/v1/config/sila_deposit",
					&silaDepositJson,
				).Return(
					testCase.silaDepositError,
				).SetArg(
					2,
					testCase.silaDepositResponse,
				)
			}

			nodeClient := &beaconApiNodeClient{
				genesisProvider: genesisProvider,
				handler:         handler,
			}
			response, err := nodeClient.Genesis(ctx, &emptypb.Empty{})

			if testCase.expectedResponse == nil {
				assert.ErrorContains(t, testCase.expectedError, err)
			} else {
				assert.DeepEqual(t, testCase.expectedResponse, response)
			}
		})
	}
}

func TestGetSyncStatus(t *testing.T) {
	const syncingEndpoint = "/sila/v1/node/syncing"

	testCases := []struct {
		name                 string
		restEndpointResponse structs.SyncStatusResponse
		restEndpointError    error
		expectedResponse     *silapb.SyncStatus
		expectedError        string
	}{
		{
			name:              "fails to query REST endpoint",
			restEndpointError: errors.New("foo error"),
			expectedError:     "foo error",
		},
		{
			name:                 "returns nil syncing data",
			restEndpointResponse: structs.SyncStatusResponse{Data: nil},
			expectedError:        "syncing data is nil",
		},
		{
			name: "returns false syncing status",
			restEndpointResponse: structs.SyncStatusResponse{
				Data: &structs.SyncStatusResponseData{
					IsSyncing: false,
				},
			},
			expectedResponse: &silapb.SyncStatus{
				Syncing: false,
			},
		},
		{
			name: "returns true syncing status",
			restEndpointResponse: structs.SyncStatusResponse{
				Data: &structs.SyncStatusResponseData{
					IsSyncing: true,
				},
			},
			expectedResponse: &silapb.SyncStatus{
				Syncing: true,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := t.Context()

			syncingResponse := structs.SyncStatusResponse{}
			handler := mock.NewMockJsonRestHandler(ctrl)
			handler.EXPECT().Get(
				gomock.Any(),
				syncingEndpoint,
				&syncingResponse,
			).Return(
				testCase.restEndpointError,
			).SetArg(
				2,
				testCase.restEndpointResponse,
			)

			nodeClient := &beaconApiNodeClient{handler: handler}
			syncStatus, err := nodeClient.SyncStatus(ctx, &emptypb.Empty{})

			if testCase.expectedResponse == nil {
				assert.ErrorContains(t, testCase.expectedError, err)
			} else {
				assert.DeepEqual(t, testCase.expectedResponse, syncStatus)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	const versionEndpoint = "/sila/v1/node/version"

	testCases := []struct {
		name                 string
		restEndpointResponse structs.GetVersionResponse
		restEndpointError    error
		expectedResponse     *silapb.Version
		expectedError        string
	}{
		{
			name:              "fails to query REST endpoint",
			restEndpointError: errors.New("foo error"),
			expectedError:     "foo error",
		},
		{
			name:                 "returns nil version data",
			restEndpointResponse: structs.GetVersionResponse{Data: nil},
			expectedError:        "empty version response",
		},
		{
			name: "returns proper version response",
			restEndpointResponse: structs.GetVersionResponse{
				Data: &structs.Version{
					Version: "sila/local",
				},
			},
			expectedResponse: &silapb.Version{
				Version: "sila/local",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := t.Context()

			var versionResponse structs.GetVersionResponse
			handler := mock.NewMockJsonRestHandler(ctrl)
			handler.EXPECT().Get(
				gomock.Any(),
				versionEndpoint,
				&versionResponse,
			).Return(
				testCase.restEndpointError,
			).SetArg(
				2,
				testCase.restEndpointResponse,
			)

			nodeClient := &beaconApiNodeClient{handler: handler}
			version, err := nodeClient.Version(ctx, &emptypb.Empty{})

			if testCase.expectedResponse == nil {
				assert.ErrorContains(t, testCase.expectedError, err)
			} else {
				assert.DeepEqual(t, testCase.expectedResponse, version)
			}
		})
	}
}

func TestIsReady(t *testing.T) {
	const healthEndpoint = "/sila/v1/node/health"

	testCases := []struct {
		name           string
		statusCode     int
		err            error
		expectedResult bool
	}{
		{
			name:           "returns true for 200 OK (fully synced)",
			statusCode:     http.StatusOK,
			expectedResult: true,
		},
		{
			name:           "returns false for 206 Partial Content (syncing)",
			statusCode:     http.StatusPartialContent,
			expectedResult: false,
		},
		{
			name:           "returns false for 503 Service Unavailable",
			statusCode:     http.StatusServiceUnavailable,
			expectedResult: false,
		},
		{
			name:           "returns false for 500 Internal Server Error",
			statusCode:     http.StatusInternalServerError,
			expectedResult: false,
		},
		{
			name:           "returns false on error",
			err:            errors.New("request failed"),
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ctx := t.Context()

			handler := mock.NewMockJsonRestHandler(ctrl)
			handler.EXPECT().GetStatusCode(
				gomock.Any(),
				healthEndpoint,
			).Return(tc.statusCode, tc.err)
			handler.EXPECT().Host().Return("http://localhost:3500").AnyTimes()

			nodeClient := &beaconApiNodeClient{handler: handler}
			result := nodeClient.IsReady(ctx)

			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
