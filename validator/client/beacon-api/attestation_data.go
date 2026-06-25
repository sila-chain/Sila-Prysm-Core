package beacon_api

import (
	"context"
	"net/url"
	"strconv"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/apiutil"
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	fieldparams "github.com/sila-chain/Sila-Consensus-Core/v7/config/fieldparams"
	"github.com/sila-chain/Sila-Consensus-Core/v7/consensus-types/primitives"
	"github.com/sila-chain/Sila-Consensus-Core/v7/encoding/bytesutil"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

func (c *beaconApiValidatorClient) attestationData(
	ctx context.Context,
	reqSlot primitives.Slot,
	reqCommitteeIndex primitives.CommitteeIndex,
) (*silapb.AttestationData, error) {
	params := url.Values{}
	params.Add("slot", strconv.FormatUint(uint64(reqSlot), 10))
	params.Add("committee_index", strconv.FormatUint(uint64(reqCommitteeIndex), 10))

	query := apiutil.BuildURL("/sila/v1/validator/attestation_data", params)
	produceAttestationDataResponseJson := structs.GetAttestationDataResponse{}

	if err := c.handler.Get(ctx, query, &produceAttestationDataResponseJson); err != nil {
		return nil, err
	}

	if produceAttestationDataResponseJson.Data == nil {
		return nil, errors.New("attestation data is nil")
	}

	attestationData := produceAttestationDataResponseJson.Data
	committeeIndex, err := strconv.ParseUint(attestationData.CommitteeIndex, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse attestation committee index: %s", attestationData.CommitteeIndex)
	}

	beaconBlockRoot, err := bytesutil.DecodeHexWithLength(attestationData.BeaconBlockRoot, fieldparams.RootLength)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode beacon block root: %s", attestationData.BeaconBlockRoot)
	}

	slot, err := strconv.ParseUint(attestationData.Slot, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse attestation slot: %s", attestationData.Slot)
	}

	if attestationData.Source == nil {
		return nil, errors.New("attestation source is nil")
	}

	sourceEpoch, err := strconv.ParseUint(attestationData.Source.Epoch, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse attestation source epoch: %s", attestationData.Source.Epoch)
	}

	sourceRoot, err := bytesutil.DecodeHexWithLength(attestationData.Source.Root, fieldparams.RootLength)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode attestation source root: %s", attestationData.Source.Root)
	}

	if attestationData.Target == nil {
		return nil, errors.New("attestation target is nil")
	}

	targetEpoch, err := strconv.ParseUint(attestationData.Target.Epoch, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse attestation target epoch: %s", attestationData.Target.Epoch)
	}

	targetRoot, err := bytesutil.DecodeHexWithLength(attestationData.Target.Root, fieldparams.RootLength)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode attestation target root: %s", attestationData.Target.Root)
	}

	response := &silapb.AttestationData{
		BeaconBlockRoot: beaconBlockRoot,
		CommitteeIndex:  primitives.CommitteeIndex(committeeIndex),
		Slot:            primitives.Slot(slot),
		Source: &silapb.Checkpoint{
			Epoch: primitives.Epoch(sourceEpoch),
			Root:  sourceRoot,
		},
		Target: &silapb.Checkpoint{
			Epoch: primitives.Epoch(targetEpoch),
			Root:  targetRoot,
		},
	}

	return response, nil
}
