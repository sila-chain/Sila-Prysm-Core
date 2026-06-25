package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/pkg/errors"
)

func (c *beaconApiValidatorClient) submitValidatorRegistrations(ctx context.Context, registrations []*silapb.SignedValidatorRegistrationV1) error {
	const endpoint = "/sila/v1/validator/register_validator"

	jsonRegistration := make([]*structs.SignedValidatorRegistration, len(registrations))

	for index, registration := range registrations {
		jsonRegistration[index] = structs.SignedValidatorRegistrationFromConsensus(registration)
	}

	marshalledJsonRegistration, err := json.Marshal(jsonRegistration)
	if err != nil {
		return errors.Wrap(err, "failed to marshal registration")
	}

	return c.handler.Post(ctx, endpoint, nil, bytes.NewBuffer(marshalledJsonRegistration), nil)
}
