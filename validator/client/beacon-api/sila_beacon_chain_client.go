package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	neturl "net/url"
	"strconv"
	"strings"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/apiutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/rest"
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	validator2 "github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/validator"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/validator/client/iface"
	"github.com/pkg/errors"
)

// NewSilaChainClient returns implementation of iface.SilaChainClient.
func NewSilaChainClient(handler rest.Handler, nodeClient iface.NodeClient) iface.SilaChainClient {
	return silaChainClient{
		handler:    handler,
		nodeClient: nodeClient,
	}
}

type silaChainClient struct {
	handler    rest.Handler
	nodeClient iface.NodeClient
}

func (c silaChainClient) ValidatorCount(ctx context.Context, stateID string, statuses []validator2.Status) ([]iface.ValidatorCount, error) {
	// Check node version for sila beacon node as it is a custom endpoint for sila beacon node.
	nodeVersion, err := c.nodeClient.Version(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node version")
	}

	if !strings.Contains(strings.ToLower(nodeVersion.Version), "sila") {
		return nil, iface.ErrNotSupported
	}

	queryParams := neturl.Values{}
	for _, status := range statuses {
		queryParams.Add("status", status.String())
	}

	queryUrl := apiutil.BuildURL(fmt.Sprintf("/sila/v1/beacon/states/%s/validator_count", stateID), queryParams)

	var validatorCountResponse structs.GetValidatorCountResponse
	if err = c.handler.Get(ctx, queryUrl, &validatorCountResponse); err != nil {
		return nil, err
	}

	if validatorCountResponse.Data == nil {
		return nil, errors.New("validator count data is nil")
	}

	if len(statuses) != 0 && len(statuses) != len(validatorCountResponse.Data) {
		return nil, errors.New("mismatch between validator count data and the number of statuses provided")
	}

	var resp []iface.ValidatorCount
	for _, vc := range validatorCountResponse.Data {
		count, err := strconv.ParseUint(vc.Count, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse validator count %s", vc.Count)
		}

		resp = append(resp, iface.ValidatorCount{
			Status: vc.Status,
			Count:  count,
		})
	}

	return resp, nil
}

func (c silaChainClient) ValidatorPerformance(ctx context.Context, in *silapb.ValidatorPerformanceRequest) (*silapb.ValidatorPerformanceResponse, error) {
	// Check node version for sila beacon node as it is a custom endpoint for sila beacon node.
	nodeVersion, err := c.nodeClient.Version(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get node version")
	}

	if !strings.Contains(strings.ToLower(nodeVersion.Version), "sila") {
		return nil, iface.ErrNotSupported
	}

	request, err := json.Marshal(structs.GetValidatorPerformanceRequest{
		PublicKeys: in.PublicKeys,
		Indices:    in.Indices,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request")
	}
	resp := &structs.GetValidatorPerformanceResponse{}
	if err = c.handler.Post(ctx, "/sila/validators/performance", nil, bytes.NewBuffer(request), resp); err != nil {
		return nil, err
	}

	return &silapb.ValidatorPerformanceResponse{
		CurrentEffectiveBalances:      resp.CurrentEffectiveBalances,
		CorrectlyVotedSource:          resp.CorrectlyVotedSource,
		CorrectlyVotedTarget:          resp.CorrectlyVotedTarget,
		CorrectlyVotedHead:            resp.CorrectlyVotedHead,
		BalancesBeforeEpochTransition: resp.BalancesBeforeEpochTransition,
		BalancesAfterEpochTransition:  resp.BalancesAfterEpochTransition,
		MissingValidators:             resp.MissingValidators,
		PublicKeys:                    resp.PublicKeys,
		InactivityScores:              resp.InactivityScores,
	}, nil
}
