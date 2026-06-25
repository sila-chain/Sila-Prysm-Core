package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/sila-chain/Sila-Consensus-Core/v7/beacon-chain/core/helpers"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/sila-chain/Sila-Consensus-Core/v7/time/slots"
	"github.com/pkg/errors"
)

func (c *beaconApiValidatorClient) proposeAttestation(ctx context.Context, attestation *silapb.Attestation) (*silapb.AttestResponse, error) {
	if err := helpers.ValidateNilAttestation(attestation); err != nil {
		return nil, err
	}
	marshalledAttestation, err := json.Marshal(jsonifyAttestations([]*silapb.Attestation{attestation}))
	if err != nil {
		return nil, err
	}

	headers := map[string]string{"Eth-Consensus-Version": version.String(attestation.Version())}
	err = c.handler.Post(
		ctx,
		"/sila/v2/beacon/pool/attestations",
		headers,
		bytes.NewBuffer(marshalledAttestation),
		nil,
	)
	if err != nil {
		return nil, err
	}

	attestationDataRoot, err := attestation.Data.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to compute attestation data root")
	}

	return &silapb.AttestResponse{AttestationDataRoot: attestationDataRoot[:]}, nil
}

func (c *beaconApiValidatorClient) proposeAttestationElectra(ctx context.Context, attestation *silapb.SingleAttestation) (*silapb.AttestResponse, error) {
	if err := helpers.ValidateNilAttestation(attestation); err != nil {
		return nil, err
	}
	marshalledAttestation, err := json.Marshal(jsonifySingleAttestations([]*silapb.SingleAttestation{attestation}))
	if err != nil {
		return nil, err
	}
	consensusVersion := version.String(slots.ToForkVersion(attestation.Data.Slot))
	headers := map[string]string{"Eth-Consensus-Version": consensusVersion}
	if err = c.handler.Post(
		ctx,
		"/sila/v2/beacon/pool/attestations",
		headers,
		bytes.NewBuffer(marshalledAttestation),
		nil,
	); err != nil {
		return nil, err
	}

	attestationDataRoot, err := attestation.Data.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to compute attestation data root")
	}

	return &silapb.AttestResponse{AttestationDataRoot: attestationDataRoot[:]}, nil
}
