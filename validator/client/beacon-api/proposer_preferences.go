package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/sila-chain/Sila-Consensus-Core/v7/api"
	"github.com/sila-chain/Sila-Consensus-Core/v7/api/server/structs"
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
	"github.com/sila-chain/Sila-Consensus-Core/v7/runtime/version"
	"github.com/pkg/errors"
)

func (c *beaconApiValidatorClient) submitSignedProposerPreferences(ctx context.Context, prefs []*silapb.SignedProposerPreferences) error {
	jsonPrefs := make([]*structs.SignedProposerPreferences, len(prefs))
	for i, p := range prefs {
		if p == nil || p.Message == nil {
			return errors.Errorf("signed proposer preferences at index %d is nil", i)
		}
		jsonPrefs[i] = structs.SignedProposerPreferencesFromConsensus(p)
	}

	body, err := json.Marshal(jsonPrefs)
	if err != nil {
		return errors.Wrap(err, "failed to marshal signed proposer preferences")
	}

	headers := map[string]string{api.VersionHeader: version.String(version.Gloas)}
	return c.handler.Post(ctx, "/sila/v1/validator/proposer_preferences", headers, bytes.NewBuffer(body), nil)
}
